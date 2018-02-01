import charWidthInTTY from 'string-width';

import utils from 'utils';

// Handle commands from tabs, like sending a frame or information about
// the current character dimensions .
export default (MixinBase) => class extends MixinBase {
  handleTabMessage(message) {
    const parts = message.split(',');
    const command = parts[0];
    switch (command) {
      case '/frame':
        this._current_frame = JSON.parse(message.slice(7));
        this._applyUI();
        this._sendCurrentFrame();
        break;
      case '/tab_info':
        this.currentTab().info = JSON.parse(utils.rebuildArgsToSingleArg(parts));
        break;
      case '/char_size':
        this.char_width = parts[1];
        this.char_height = parts[2]
        if(this.tty_width && this.tty_height) this.resizeBrowserWindow();
        break;
      case '/request_tty_size':
        this.sendTTYSizeToBrowser();
        break;
      case '/status':
        if (this._current_frame) {
          this.updateStatus(parts[1]);
        }
        break;
      case `/log`:
        this.log(parts[1]);
        break;
      default:
        this.log('Unknown command from tab to background', message);
    }
  }

  sendTTYSizeToBrowser() {
    this.sendToCurrentTab(`/tty_size,${this.tty_width},${this.tty_height}`);
  }

  _sendCurrentFrame() {
    const raw_frame = this._current_frame.join('');
    this.sendToTerminal(`/frame,${raw_frame}`);
  }

  updateStatus(status, message = '') {
    if (typeof this._current_frame === 'undefined') return;
    switch (status) {
      case 'page_init':
        this._page_status = `Loading ${this.currentTab().info.url}`;
        break;
      case 'parsing_complete':
        this._page_status = '';
        break;
      case 'window_unload':
        this._page_status = 'Loading...';
        break;
      default:
        if (message != '') status = message;
        this._page_status = status;
    }
    this._applyStatus();
    this._sendCurrentFrame();
  }

  _applyUI() {
    const tabs = this._buildTTYRow(this._buildTabs());
    const urlBar = this._buildURLBar();
    this._current_frame = tabs.concat(urlBar).concat(this._current_frame);
    this._applyStatus();
  }

  _applyStatus() {
    if (typeof this._page_status === 'undefined') return;
    let cell;
    const start = (this.tty_height - 1) * this.tty_width;
    for (let i = 0; i < this.tty_width; i++) {
      if (this._page_status[i] !== undefined) {
        cell = utils.ttyPixel([255, 255, 255], [0, 0, 0], this._page_status[i]);
        this._current_frame[start + i] = cell;
      }
    }
  }

  _buildTabs() {
    return this.currentTab().info.title.trim();
  }

  _buildTTYRow(text) {
    let char;
    let row = [];
    let index = 0;
    while (index < this.tty_width) {
      if (index < text.length) {
        char = text[index];
      } else {
        char = " "
      }
      if (charWidthInTTY(char) > 0) {
        index += charWidthInTTY(char);
      } else {
        index += 1;
        char = " ";
      }
      row.push(utils.ttyPixel([255, 255, 255], [0, 0, 0], char));
    }
    return row;
  }

  _buildURLBar() {
    let content;
    if (this.isURLBarFocused) {
      content = this.urlBarUserContent;
    } else {
      content = this.currentTab().info.url;
    }
    content = ' ← | x | ' + content;
    return this._buildTTYRow(content);
  }
};
