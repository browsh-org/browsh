import utils from 'utils';

import CommonMixin from 'background/common_mixin';
import TabCommandsMixin from 'background/tab_commands_mixin';

export default class extends utils.mixins(CommonMixin, TabCommandsMixin) {
  constructor() {
    super();
    // Keep track of automatic reloads to problematic tabs
    this._tab_reloads = 0;
    // The maximum amount of times to try to recover a tab that won't connect
    this._max_number_of_tab_recovery_reloads = 3;
    // Type of raw text mode; HTML or plain
    this.raw_text_mode_type = '';
  }

  postDOMLoadInit(terminal, dimensions) {
    this.terminal = terminal;
    this.dimensions = dimensions;
    this._closeUnwantedStartupTabs();
  }

  postConnectionInit(channel) {
    this.channel = channel;
    this._sendTTYDimensions();
    this._listenForMessages();
    this._calculateMode();
  }

  _calculateMode() {
    let mode;
    if (!this._is_raw_text_mode) {
      mode = 'interactive';
    } else {
      mode = this.raw_text_mode_type;
    }
    this.channel.postMessage(`/mode,${mode}`);
  }

  isConnected() {
    return this.channel !== undefined;
  }

  reload() {
    const reloading = browser.tabs.reload(this.id);
    reloading.then(
      tab => this.log(`Tab ${tab.id} reloaded.`),
      error => this.log(error)
    );
  }

  remove() {
    const removing = browser.tabs.remove(this.id);
    removing.then(
      () => this.log(`Tab ${this.id} removed.`),
      error => this.log(error)
    );
  }

  updateStatus(status, message = '') {
    let status_message;
    switch (status) {
      case 'page_init':
        status_message = `Loading ${this.url}`;
        break;
      case 'parsing_complete':
        status_message = '';
        break;
      case 'window_unload':
        status_message = 'Loading...';
        break;
      default:
        if (message != '') status_message = message;
    }
    this.page_state = status;
    this.status_message = status_message;
    this.sendStateToTerminal();
  }

  getStateObject() {
    return {
      id: this.id,
      active: this.active,
      removed: this.removed,
      title: this.title,
      uri: this.url,
      page_state: this.page_state,
      status_message: this.status_message
    };
  }

  sendStateToTerminal() {
    this.sendToTerminal(`/tab_state,${JSON.stringify(this.getStateObject())}`);
  }

  // For various reasons a tab's content script doesn't always load. Currently
  // the known reasons are;
  //   1. Pages without content, such as direct links to images.
  //   2. Native pages such as `about:config`.
  //   3. Unknown buggy behaviour such as on Travis :/
  // So here we attempt some workarounds.
  ensureConnectionToBackground() {
    let native_status;
    if (!this._isItOKToRetryReload()) {
      return;
    }
    if (this.native_last_change) {
      native_status = this.native_last_change.status;
    }
    if (native_status === 'complete' && !this._isConnected()) {
      this.log(
        `Automatically reloading tab ${this.id} that has loaded but not connected ` +
        'to the webextension'
      );
      this.reload();
      this._reload_count++;
    }
  }

  setMode(mode) {
    this.raw_text_mode_type = mode;
    // Send it here, in case there is a race condition with the postCommsInit() not knowing
    // the mode.
    if (this._is_raw_text_mode) {
      this.channel.postMessage(`/mode,${mode}`);
    }
  }

  _listenForMessages() {
    this.channel.onMessage.addListener(this.handleTabMessage.bind(this));
  }

  _sendTTYDimensions() {
    this.channel.postMessage(
      `/tty_size,${this.dimensions.tty.width},${this.dimensions.tty.height}`
    );
  }

  _isItOKToRetryReload() {
    return this._reload_count <= this._max_number_of_tab_recovery_reloads;
  }

  // On the very first startup of Firefox on a new profile it loads a tab disclaiming
  // its data collection to a third-party. Sometimes this tab loads first, sometimes
  // it loads second. Especially for testing we always need to load the tab we requested
  // first. So let's just close that tab.
  // TODO: Only do this for a testing ENV?
  _closeUnwantedStartupTabs() {
    if (this.title === undefined) { return false }
    if (
      this.title.includes('Firefox by default shares data to:') ||
      this.title.includes('Firefox Privacy Notice')
    ) {
      this.log("Removing Firefox startup page")
      this.remove();
      return true;
    }
    return false;
  }
}
