import mixins from 'mixin_factory';
import HubMixin from 'background/hub_mixin';
import TTYCommandsMixin from 'background/tty_commands_mixin';
import TabCommandsMixin from 'background/tab_commands_mixin';

// Boots the background process. Mainly involves connecting to the websocket server
// launched by the Browsh CLI client and setting up listeners for new tabs that
// have our webextension content script inside them.
export default class extends mixins(HubMixin, TTYCommandsMixin, TabCommandsMixin) {
  constructor() {
    super();
    // Keep track of connections to active tabs
    this.tabs = {};
    // The ID of the tab currently opened in the browser
    this.active_tab_id = null;
    this._connectToTerminal();
  }

  _connectToTerminal() {
    // This is the websocket server run by the CLI client
    this.terminal = new WebSocket('ws://localhost:3334');
    this.terminal.addEventListener('open', (_event) => {
      this.log("Webextension connected to the terminal's websocket server");
      this._listenForTerminalMessages();
      this._connectToBrowser();
    });
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

  _connectToBrowser() {
    this._listenForNewTab();
    this._listenForTabChannelOpen();
    this._listenForFocussedTab();
    this._startFrameRequestLoop();
  }

  // Curiously `browser.runtime.onMessage` receives the tab's ID, whereas
  // `browser.runtime.onConnect` doesn't. So we have to have 2 tab listeners:
  //   1. to get the tab ID so we can talk to it later with 2.
  //   2. to maintain a long-lived connection to continuously pass messages
  //      back and forth.
  _listenForNewTab() {
    browser.runtime.onMessage.addListener(this._newTabHandler.bind(this));
  }

  getTabsOnSuccess(windowInfoArray) {
    for (let windowInfo of windowInfoArray) {
      this.log('BACKGROUND: Current tab count: ' + windowInfo.tabs.length);
      windowInfo.tabs.map((tab) => {
        this.log(
          `ID:${tab.id}(${tab.active ? '!' : 'x'}) ${tab.title} - ${tab.url}`
        )
        this.log(JSON.stringify(tab));
      });
    }
  }

  getTabsOnError(error) {
    this.log(`Error: ${error}`);
  }

  logTabs() {
    var getting = browser.windows.getAll({
      populate: true,
      windowTypes: ["normal"]
    });
    getting.then(this.getTabsOnSuccess.bind(this), this.getTabsOnError.bind(this));
  }

  // On the very first startup of Firefox on a new profile it loads a tab disclaiming
  // its data collection to a third-party. Sometimes this tab loads first, sometimes
  // it loads second. Especially for testing we always need to load the tab we requested
  // first. So let's just close that tab.
  // TODO: Only do this for a testing ENV?
  checkForMozillaCliqzTab(tab) {
    if (tab.title.includes('Firefox by default shares data to:')) {
      this.log("Removing the Mozilla Cliqz disclaimer startup tab")
      const removing = browser.tabs.remove(tab.id);
      removing.then(() => {}, (error) => this.log(error));
      return true;
    }
    return false;
  }

  _newTabHandler(_request, sender, sendResponse) {
    this.logTabs()
    this.log(`Tab ${sender.tab.id} (${sender.tab.title}) registered with background process`);
    if (this.checkForMozillaCliqzTab(sender.tab)) return;
    // Send the tab back to itself, such that it can be enlightened unto its own nature
    sendResponse(sender.tab);
    //if (sender.tab.active) this.active_tab_id = sender.tab.id;
    this.active_tab_id = sender.tab.id;
  }

  // This is the main communication channel for all back and forth messages to tabs
  _listenForTabChannelOpen() {
    browser.runtime.onConnect.addListener(this._tabChannelOpenHandler.bind(this));
  }

  _tabChannelOpenHandler(channel) {
    // TODO: Can we not just assume that channel.name is the same as this.active_tab_id?
    this.log(`Tab ${channel.name} connected for communication with background process`);
    this.tabs[channel.name] = {
      channel: channel
    };
    channel.onMessage.addListener(this.handleTabMessage.bind(this));
  }

  _listenForFocussedTab() {
    browser.tabs.onActivated.addListener(this._focussedTabHandler.bind(this));
  }

  _focussedTabHandler(tab) {
    this.active_tab_id = tab.id
  }

  // The browser window can only be resized once we have both the character dimensions from
  // the browser tab _and the TTY dimensions from the terminal. There's probably a more
  // efficient way of triggering this initial window resize, than just waiting for the data
  // on every frame tick.
  _initialWindowResize() {
    if (!this._is_intial_window_pending) return;
    if(this.char_width && this.char_height && this.tty_width && this.tty_height) {
      this.resizeBrowserWindow();
      this._is_intial_window_pending = false;
    }
  }

  // Instead of having each tab manage its own frame rate, just keep this single, centralised
  // heartbeat in the background process that switches automatically to the current active
  // tab.
  _startFrameRequestLoop() {
    this.log('BACKGROUND: Frame loop starting')
    setInterval(() => {
      if (!this.tty_width || !this.tty_height) {
        this.log("Not sending frame to TTY without TTY size")
        return;
      }
      if (this._is_intial_window_pending) this._initialWindowResize();
      if (!this.tabs.hasOwnProperty(this.active_tab_id)) {
        this.log("No active tab, so not requesting a frame");
        return;
      }
      this.sendToCurrentTab('/request_frame');
    }, 1000);
  }
}
