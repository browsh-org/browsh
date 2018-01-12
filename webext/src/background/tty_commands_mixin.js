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
    switch (command) {
      case '/tty_size':
        this.tty_width = parts[1];
        this.tty_height = parts[2];
        if(this.active_tab_id) this.sendTTYSizeToBrowser();
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
    }
  }

  _handleUICommand(parts) {
    const input = JSON.parse(utils.rebuildArgsToSingleArg(parts));
    if (this.isURLBarFocused) {
      this._handleURLBarInput(input);
      return true;
    }
    switch (input.key) {
      case 12:
        this.isURLBarFocused = true;
        return true;
    }
    return false;
  }

  _handleURLBarInput(input) {
    let url;
    const search_engine = 'https://www.google.com/search?q=';
    let char = input.char;
    switch (input.key) {
      case 12:
        this.isURLBarFocused = false;
        return;
      case 13:
        url = this.urlBarUserContent;
        this.urlBarUserContent = "";
        this.isURLBarFocused = false;
        this.sendToCurrentTab(`/url,${search_engine}${url}`);
        return;
      case 32:
        char = " ";
        break;
      case 127:
        this.urlBarUserContent = this.urlBarUserContent.slice(0, -1);
        return;
    }
    this.urlBarUserContent += char;
  }

  resizeBrowserWindow() {
    if (!this.tty_width || !this.char_width || !this.tty_height || !this.char_height) {
      this.log(
        'Not resizing browser window without all of the TTY and character dimensions'
      );
      return;
    }
    // Does this include scrollbars???
    const window_width = parseInt(Math.round(this.tty_width * this.char_width)) + 4;
    // Leave room for tabs and URL bar. TODO: globally refactor TTY DOM height
    const tty_dom_height = this.tty_height - 2;
    // This is actually line-height
    const window_height = parseInt(Math.round((tty_dom_height) * this.char_height)) + 4;
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
    updating(
      (info) => {
        this.log(tag, info);
      },
      (error) => {
        this.log(tag, error);
      }
    );
  }
}

