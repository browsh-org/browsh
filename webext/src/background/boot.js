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
      this._connectToBrowser();
    });
    this._listenForTerminalMessages();
  }

  // Mostly listening for forwarded STDIN from the terminal. Therefore, the user
  // pressing the arrow keys, typing, moving the mouse, etc, etc. But we also listen
  // to TTY resize events too.
  _listenForTerminalMessages() {
    this.terminal.addEventListener('message', (event) => {
      this.log('Message from terminal', event.data);
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
  //   1. to get the tab ID so we can talk to it later.
  //   2. to maintain a long-lived connection to continuously pass messages
  //      back and forth.
  _listenForNewTab() {
    browser.runtime.onMessage.addListener(this._newTabHandler.bind(this));
  }

  _newTabHandler(_request, sender, sendResponse) {
    this.log(`Tab ${sender.tab.id} registered with background process`);
    // Send the tab back to itself, such that it can be enlightened unto its own nature
    sendResponse(sender.tab);
    if (sender.tab.active) this.active_tab_id = sender.tab.id;
  }

  // This is the main communication channel for all back and forth messages to tabs
  _listenForTabChannelOpen() {
    browser.runtime.onConnect.addListener(this._tabChannelOpenHandler.bind(this));
  }

  _tabChannelOpenHandler(channel) {
    this.log(`Tab ${channel.name} connected for communication with background process`);
    this.tabs[channel.name] = channel;
    channel.onMessage.addListener(this.handleTabMessage.bind(this));
  }

  _listenForFocussedTab() {
    browser.tabs.onActivated.addListener(this._focussedTabHandler.bind(this));
  }

  _focussedTabHandler(tab) {
    this.log(tab);
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

  // Instead of having each tab manage its own frame rate, just keep a single, centralised
  // heartbeat in the background process that switches automatically to the current active
  // tab.
  _startFrameRequestLoop() {
    setInterval(() => {
      if (!this.tty_width || !this.tty_height) {
        this.log("Not sending frame to TTY without TTY size")
        return;
      }
      if (this._is_intial_window_pending) this._initialWindowResize();
      if (!this.tabs.hasOwnProperty(this.active_tab_id)) return;
      this.sendToCurrentTab('/request_frame');
    }, 300);
  }
}
