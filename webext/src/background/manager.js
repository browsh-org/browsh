import _ from 'lodash';
import utils from 'utils';
import CommonMixin from 'background/common_mixin';
import TTYCommandsMixin from 'background/tty_commands_mixin';
import Tab from 'background/tab';
import Dimensions from 'background/dimensions';

// Boots the background process. Mainly involves connecting to the websocket server
// launched by the Browsh CLI client and setting up listeners for new tabs that
// have our webextension content script inside them.
export default class extends utils.mixins(CommonMixin, TTYCommandsMixin) {
  constructor() {
    super();
    this.dimensions = new Dimensions();
    // All of the tabs open in the real browser
    this.tabs = {};
    // The ID of the tab currently opened tab
    this.active_tab_id = null;
    // When the real GUI browser first launches it's sized to the same size as the desktop
    this._is_initial_window_size_pending = true;
    // Used so that reconnections to the terminal don't also attempt to reconnect to the
    // browser DOM.
    this._is_connected_to_browser_dom = false;
    // The time in milliseconds between requesting a new TTY-size pixel frame
    this._small_pixel_frame_rate = 250;
    // Raw text mode is for when Browsh is running as an HTTP server that serves single
    // pages as entire DOMs, in plain text.
    this._is_raw_text_mode = false;
    // A mobile user agent for forcing web pages to use its mobile layout
    this._mobile_user_agent =
      "Mozilla/5.0 (Android 7.0; Mobile; rv:54.0) Gecko/58.0 Firefox/58.0";
    this._is_using_mobile_user_agent = false;
    this._addUserAgentListener();
    // Listen to HTTP requests. This allows us to display some helpful status messages at the
    // bottom of the page, eg; "Loading https://coolwebsite.com..."
    this._addWebRequestListener();
    // The manager is the hub between tabs and the terminal. First we connect to the
    // terminal, as that is the process that would have initially booted the browser and
    // this very code that now runs.
    this._connectToTerminal();
  }

  _connectToTerminal() {
    // This is the websocket server run by the CLI client
    this.terminal = new WebSocket('ws://localhost:3334');
    this.terminal.addEventListener('open', (_event) => {
      this.log("Webextension connected to the terminal's websocket server");
      this.dimensions.terminal = this.terminal;
      this._listenForTerminalMessages();
      this._connectToBrowserDOM();
      this._startFrameRequestLoop();
    });
    this.terminal.addEventListener('close', (_event) => {
      this._reconnectToTerminal();
    });
  }

  // If we've disconnected from the terminal, but we're still running, then that likely
  // means the terminal crashed, so we wait to see if the user restarts the terminal.
  _reconnectToTerminal() {
    try {
      this._connectToTerminal();
    } catch (_e) {
      _.debounce(() => this._reconnectToTerminal(), 50);
    }
  }

  // Mostly listening for forwarded STDIN from the terminal. Therefore, the user
  // pressing the arrow keys, typing, moving the mouse, etc, etc. But we also listen
  // to TTY resize events too.
  _listenForTerminalMessages() {
    this.log('Starting to listen to TTY')
    this.terminal.addEventListener('message', (event) => {
      this.log('Message from terminal: ' + event.data);
      this.handleTerminalMessage(event.data)
    });
  }

  _connectToBrowserDOM() {
    if (!this._is_connected_to_browser_dom) {
      this._initialDOMConnection()
    } else {
      this._reconnectToDOM();
    }
  }

  _initialDOMConnection() {
    this._listenForNewTab();
    this._listenForTabUpdates();
    this._listenForTabChannelOpen();
    this._listenForFocussedTab();
  }

  _reconnectToDOM() {
    this.log("Attempting to resend browser state to terminal...");
    this.currentTab().sendStateToTerminal();
    if (!this._is_raw_text_mode) {
      this.sendToCurrentTab('/rebuild_text');
    }
  }

  // For when a tab's content script, triggered by `onDOMContentLoaded`,
  // phone's home.
  // Curiously `browser.runtime.onMessage` receives the tab's ID, whereas
  // `browser.runtime.onConnect` doesn't. So we have to have 2 tab listeners:
  //   1. to get the tab ID so we can talk to it later with 2.
  //   2. to maintain a long-lived connection to continuously pass messages
  //      back and forth.
  _listenForNewTab() {
    browser.runtime.onMessage.addListener(this._newTabHandler.bind(this));
  }

  // There's what seems to be a bug: tabs can exist and be processed without
  // triggering any `browser.tabs.onUpdated` events. Therefore we need to
  // manually poll :/
  // TODO: Detect deleted tabs to remove the key from `this.tabs[]`
  _listenForTabUpdates() {
    setInterval(() => {
      this._pollAllTabs((native_tab_object) => {
        let tab = this._applyUpdates(native_tab_object);
        tab.ensureConnectionToBackground();
      });
    }, 100);
  }

  _maybeNewTab(tabish_object) {
    const tab_id = parseInt(tabish_object.id);
    if (this.tabs[tab_id] === undefined) {
      let new_tab = new Tab(tabish_object);
      this.tabs[tab_id] = new_tab;
    }
    return this.tabs[tab_id];
  }

  _handleTabUpdate(_tab_id, changes, native_tab_object) {
    this.log(`Tab ${native_tab_object.id} detected chages: ${JSON.stringify(changes)}`);
    let tab = this.tabs[native_tab_object.id];
    tab.native_last_change = changes
    tab.ensureConnectionToBackground();
  }

  // Note that although this callback signifies that the tab now exists, it is not fully
  // booted and functional until it has opened a communication channel. It can't do that
  // until it knows its internally represented ID.
  _newTabHandler(_request, sender, sendResponse) {
    this.log(`Tab ${sender.tab.id} (${sender.tab.title}) registered with background process`);
    // Send the tab back to itself, such that it can be enlightened unto its own nature
    sendResponse(sender.tab);
    this._acknowledgeNewTab(sender.tab);
  }

  _acknowledgeNewTab(native_tab_object) {
    let tab = this._applyUpdates(native_tab_object);
    tab._is_raw_text_mode = this._is_raw_text_mode;
    tab.postDOMLoadInit(this.terminal, this.dimensions);
  }

  _applyUpdates(tabish_object) {
    let tab = this._maybeNewTab({id: tabish_object.id});
    ['id', 'title', 'url', 'active', 'request_id'].map(key => {
      if (tabish_object.hasOwnProperty(key)){
        tab[key] = tabish_object[key]
      }
    });
    if (tabish_object.active) {
      this.active_tab_id = tab.id;
    }
    return tab;
  }

  // This is the main communication channel for all back and forth messages to tabs
  _listenForTabChannelOpen() {
    browser.runtime.onConnect.addListener(this._tabChannelOpenHandler.bind(this));
  }

  _tabChannelOpenHandler(channel) {
    this.log(`Tab ${channel.name} connected for communication with background process`);
    let tab = this.tabs[parseInt(channel.name)];
    tab.postConnectionInit(channel);
    this._is_connected_to_browser_dom = true;
  }

  _listenForFocussedTab() {
    browser.tabs.onActivated.addListener(this._focussedTabHandler.bind(this));
  }

  _focussedTabHandler(tab) {
    this.log(`Tab ${tab.id} received new focus`);
    this.active_tab_id = tab.id
  }

  _getTabsOnSuccess(windowInfoArray, callback) {
    for (let windowInfo of windowInfoArray) {
      windowInfo.tabs.map((tab) => {
        callback(tab);
      });
    }
  }

  _getTabsOnError(error) {
    this.log(`Error: ${error}`);
  }

  _pollAllTabs(callback) {
    var getting = browser.windows.getAll({
      populate: true,
      windowTypes: ["normal"]
    });
    getting.then(
      (windowInfoArray) => this._getTabsOnSuccess(windowInfoArray, callback),
      () => this._getTabsOnError(callback)
    );
  }

  // The browser window can only be resized once we have both the character dimensions from
  // the browser tab _and the TTY dimensions from the terminal. There's probably a more
  // efficient way of triggering this initial window resize, than just waiting for the data
  // on every frame tick.
  _initialWindowResize() {
    if (!this._is_initial_window_size_pending) return;
    this.dimensions.resizeBrowserWindow();
    this._is_initial_window_size_pending = false;
  }

  // Instead of having each tab manage its own frame rate, just keep this single, centralised
  // heartbeat in the background process that switches automatically to the current active
  // tab.
  //
  // Note that by "frame rate" here we justs mean the rate at which a TTY-sized frame of
  // graphics pixles are sent. Larger frames are sent in response to scroll events and
  // TTY-sized text frames are sent in response to DOM mutation events.
  _startFrameRequestLoop() {
    this.log('BACKGROUND: Frame loop starting')
    setInterval(() => {
      if (this._is_initial_window_size_pending) this._initialWindowResize();
      if (this._isAbleToRequestFrame()) {
        this.sendToCurrentTab('/request_frame');
      }
    }, this._small_pixel_frame_rate);
  }

  _isAbleToRequestFrame() {
    if (this._is_raw_text_mode) { return false }
    if (!this.dimensions.tty.width || !this.dimensions.tty.height) {
      this.log("Not sending frame to TTY without TTY size")
      return false;
    }
    if (!this.tabs.hasOwnProperty(this.active_tab_id)) {
      this.log("No active tab, so not requesting a frame");
      return false;
    }
    if (this.currentTab().channel === undefined) {
      this.log(
        `Active tab ${this.active_tab_id} does not have a channel, so not requesting a frame`
      );
      return false;
    }
    return true;
  }

  _addWebRequestListener() {
    browser.webRequest.onBeforeRequest.addListener(
      (e) => {
        let message;
        if (e.type == 'main_frame') {
          message = `Loading ${e.url}`;
          if (this.currentTab() !== undefined) {
            this.currentTab().updateStatus('info', message);
          }
        }
      },
      {urls: ['*://*/*']},
      ["blocking"]
    );
  }
}
