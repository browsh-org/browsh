// Handle commands coming in from the terminal, like; STDIN keystrokes, TTY resize, etc
export default (MixinBase) => class extends MixinBase {
  handleTerminalMessage(message) {
    const parts = message.split(',');
    const command = parts[0];
    if (command === '/tty_size') {
      this.tty_width = parts[1];
      this.tty_height = parts[2];
      if (this.char_width && this.char_height) this.resizeBrowserWindow();
    } else {
      this.sendToCurrentTab(message);
    }
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
    // This is actually line-height
    const window_height = parseInt(Math.round(this.tty_height * this.char_height)) + 4;
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
        focussed: false
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

