import utils from 'utils';

// Handle commands coming in from the terminal, like; STDIN keystrokes, TTY resize, etc
export default (MixinBase) => class extends MixinBase {
  constructor() {
    super();
    this.isURLBarFocused = false;
    this.urlBarUserContent = "";
  }

  handleTerminalMessage(message) {
    const parts = message.split(',');
    const command = parts[0];
    switch(command) {
      case '/tty_size':
        this.tty_width = parts[1];
        this.tty_height = parts[2];
        if (this.char_width && this.char_height){
          this.resizeBrowserWindow();
        }
        break;
      case '/stdin':
        if (!this._handleUICommand(parts)) {
          this.sendToCurrentTab(message);
        }
        // Trigger a faster feedback response
        // TODO: cancel the current FPS iteration when using this
        this.sendToCurrentTab('/request_frame');
        break;
      case '/status':
        this.updateStatus('', parts.slice(1).join(','));
        break;
    }
  }

  _handleUICommand(parts) {
    const input = JSON.parse(utils.rebuildArgsToSingleArg(parts));
    if (this.isURLBarFocused) {
      this._handleURLBarInput(input);
      return true;
    }
    switch(input.key) {
      case 12: // CTRL+L
        this.isURLBarFocused = true;
        this.urlBarUserContent = "";
        return true;
    }
    if (input.mod === 1) {
      switch(input.char) {
        case 'P':
          this.screenshotActiveTab();
          break;
      }
    }
    if (input.key === 65512 && input.mouse_y === 1) {
      const x = input.mouse_x;
      switch(true) {
        case x > 0 && x < 3:
          this.sendToCurrentTab('/location_back');
          break;
        case x > 3 && x < 6:
          this.sendToCurrentTab('/window_stop');
          break;
        default:
          this.isURLBarFocused = true;
      }
      return true;
    }
    return false;
  }

  _handleURLBarInput(input) {
    let char = input.char;
    switch (input.key) {
      case 12: // CTRL+L
        this.isURLBarFocused = false;
        return;
      case 13: // enter
        this.sendToCurrentTab(`/url,${this._getURLfromUserInput()}`);
        this.isURLBarFocused = false;
        return;
      case 32: // spacebar
        char = " ";
        break;
      case 127: // backspace
        this.urlBarUserContent = this.urlBarUserContent.slice(0, -1);
        return;
    }
    this.urlBarUserContent += char;
  }

  _getURLfromUserInput() {
    let url;
    const search_engine = 'https://www.google.com/search?q=';
    let input = this.urlBarUserContent;
    // Basically just check to see if there is text either side of a dot
    const is_url = RegExp(/^[^\s]+\.[^\s]+/);
    if (is_url.test(input)) {
      url = input;
      if (!url.startsWith('http')) {
        url = 'http://' + url;
      }
    } else {
      url = `${search_engine}${input}`;
    }
    this.urlBarUserContent = url;
    return url;
  }

  resizeBrowserWindow() {
    if (!this.tty_width || !this.char_width || !this.tty_height || !this.char_height) {
      this.log(
        'Not resizing browser window without all of the TTY and character dimensions'
      );
      return;
    }
    // Does this include scrollbars???
    const window_width = parseInt(Math.round(this.tty_width * this.char_width));
    // Leave room for tabs and URL bar. TODO: globally refactor TTY DOM height
    const tty_dom_height = this.tty_height - 2;
    // I don't know why we have to add 4 more lines to the window height?? But without
    // it text doesn't fill the bottom of the TTY.
    const window_height = parseInt(Math.round(
      (tty_dom_height + 4) * this.char_height
    ));
    const current_window = browser.windows.getCurrent();
    current_window.then(
      (active_window) => {
        this._sendWindowResizeRequest(active_window, window_width, window_height);
      },
      (error) => {
        this.log('Error getting current browser window', error);
      }
    );
  }

  _sendWindowResizeRequest(active_window, width, height) {
    const tag = 'Resizing browser window';
    this.log(tag, active_window, width, height);
    const updating = browser.windows.update(
      active_window.id,
      {
        width: width,
        height: height,
        focused: false
      }
    );
    updating.then(
      (info) => {
        this.log(tag, info);
      },
      (error) => {
        this.log(tag, error);
      }
    );
  }

  // We use the `browser` object here rather than going into the actual content script
  // because the content script may have crashed even never loaded.
  screenshotActiveTab() {
    const capturing = browser.tabs.captureVisibleTab({ format: 'jpeg' });
    capturing.then(this.saveScreenshot.bind(this), error => this.log(error));
  }

  saveScreenshot(imageUri) {
    const data = imageUri.replace(/^data:image\/\w+;base64,/, "");
    this.sendToTerminal('/screenshot,' + data);
  }
}

