import utils from 'utils';

// Handle commands coming in from the terminal, like; STDIN keystrokes, TTY resize, etc
export default (MixinBase) => class extends MixinBase {
  constructor() {
    super();
  }

  handleTerminalMessage(message) {
    const parts = message.split(',');
    const command = parts[0];
    switch(command) {
      case '/tab_command':
        this.sendToCurrentTab(message.slice(13));
        break;
      case '/tty_size':
        this.tty_width = parseInt(parts[1]);
        this.tty_height = parseInt(parts[2]);
        if (this.currentTab()) {
          this.sendToCurrentTab(`/tty_size,${this.tty_width},${this.tty_height}`)
        }
        this.resizeBrowserWindow();
        break;
      case '/stdin':
        this._handleUICommand(parts);
        this.sendToCurrentTab(message);
        break;
      case '/url_bar':
        // TODO: move to CLI client
        this._handleURLBarInput(parts.slice(1).join(','));
        break;
    }
  }

  _handleUICommand(parts) {
    const input = JSON.parse(utils.rebuildArgsToSingleArg(parts));
    if (input.mod === 4) {
      switch(input.char) {
        case 'P':
          this.screenshotActiveTab();
          break;
      }
    }
    return false;
  }

  _handleURLBarInput(input) {
    const final_url = this._getURLfromUserInput(input);
    this.sendToCurrentTab(`/url,${final_url}`);
  }

  _getURLfromUserInput(input) {
    let url;
    const search_engine = 'https://www.google.com/search?q=';
    // Basically just check to see if there is text either side of a dot
    const is_straddled_dot = RegExp(/^[^\s]+\.[^\s]+/);
    // More comprehensive URL pattern
    const is_url = RegExp(/\/\/\w+(\.\w+)*(:[0-9]+)?\/?(\/[.\w]*)*$/);
    if (is_straddled_dot.test(input) || is_url.test(input)) {
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
    if (!this.tty_width || !this.dimensions.char.width || !this.tty_height || !this.dimensions.char.height) {
      this.log(
        'Not resizing browser window without all of the TTY and character dimensions'
      );
      return;
    }
    // Does this include scrollbars???
    const window_width = parseInt(Math.round(this.tty_width * this.dimensions.char.width));
    // Leave room for tabs and URL bar
    const tty_dom_height = this.tty_height - 2;
    // I don't know why we have to add 4 more lines to the window height?? But without
    // it text doesn't fill the bottom of the TTY.
    const window_height = parseInt(Math.round(
      (tty_dom_height + 4) * this.dimensions.char.height
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
  // because the content script may have crashed, even never loaded.
  screenshotActiveTab() {
    const capturing = browser.tabs.captureVisibleTab({ format: 'jpeg' });
    capturing.then(this.saveScreenshot.bind(this), error => this.log(error));
  }

  saveScreenshot(imageUri) {
    const data = imageUri.replace(/^data:image\/\w+;base64,/, "");
    this.sendToTerminal('/screenshot,' + data);
  }
}

