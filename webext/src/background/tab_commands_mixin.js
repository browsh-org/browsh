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
          this._sendCurrentFrame();
        }
        break;
      case `/log`:
        this.log(message.slice(5));
        break;
      default:
        this.log('Unknown command from tab to background', message);
    }
  }

  sendTTYSizeToBrowser() {
    this.sendToCurrentTab(`/tty_size,${this.tty_width},${this.tty_height}`);
  }

  _sendCurrentFrame() {
    // TODO: I struggled with unmarshalling a mixed array in Golang so I'm crudely
    // just casting evertything to a string for now.
    this._current_frame = this._current_frame.map((i) => i.toString());
    this.sendToTerminal(`/frame,${JSON.stringify(this._current_frame)}`);
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
    const cell_item_count = 7
    const bottom_line = this.tty_height - 1;
    const start = bottom_line * this.tty_width * cell_item_count;
    for (let i = 0; i < this.tty_width; i++) {
      if (this._page_status[i] !== undefined) {
        cell = utils.ttyPlainCell(this._page_status[i]);
        this._current_frame.splice(start + (i * 7), 7, ...cell)
      }
    }
  }

  _buildTabs() {
    return this.currentTab().info.title.trim();
  }

  _buildTTYRow(text) {
    let char, char_width;
    let row = [];
    for (let index = 0; index < this.tty_width; index++) {
      if (index < text.length) {
        char = text[index];
      } else {
        char = " "
      }
      char_width = charWidthInTTY(char);
      if (char_width === 0) {
        char = " ";
      }
      if (char_width > 1) {
        index += char_width - 1;
      }
      row = row.concat(utils.ttyPlainCell(char));
      for (var padding = 0; padding < char_width - 1; padding++) {
        row = row.concat(utils.ttyPlainCell(" "));
      }
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
