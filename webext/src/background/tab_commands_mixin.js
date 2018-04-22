import charWidthInTTY from 'string-width';

import utils from 'utils';

// Handle commands from tabs, like sending a frame or information about
// the current character dimensions .
export default (MixinBase) => class extends MixinBase {
  handleTabMessage(message) {
    let incoming;
    const parts = message.split(',');
    const command = parts[0];
    switch (command) {
      case '/frame_text':
        this.sendToTerminal(`/frame_text,${message.slice(12)}`);
        break;
      case '/frame_pixels':
        this.sendToTerminal(`/frame_pixels,${message.slice(14)}`);
        break;
      case '/tab_info':
        this.currentTab().info = JSON.parse(utils.rebuildArgsToSingleArg(parts));
        break;
      case '/dimensions':
        incoming = JSON.parse(message.slice(12));
        this._mightResizeWindow(incoming);
        this.dimensions = incoming;
        this._sendFrameSize();
        break;
      case '/status':
        if (this._current_frame) {
          this.updateStatus(parts[1]);
          this.sendState();
        }
        break;
      case `/log`:
        this.log(message.slice(5));
        break;
      default:
        this.log('Unknown command from tab to background', message);
    }
  }

  _mightResizeWindow(incoming) {
    if (this.dimensions.char.width != incoming.char.width ||
        this.dimensions.char.height != incoming.char.height) {
      this.dimensions = incoming;
      this.resizeBrowserWindow();
    }
  }

  _sendFrameSize() {
    this.state['frame_width'] = this.dimensions.frame.width;
    this.state['frame_height'] = this.dimensions.frame.height;
    this.sendState();
  }

  updateStatus(status, message = '') {
    if (typeof this._current_frame === 'undefined') return;
    let status_message;
    switch (status) {
      case 'page_init':
        status_message = `Loading ${this.currentTab().info.url}`;
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
    this.state['page_state'] = status;
    this.state['page_status_message'] = status_message;
    this.sendState();
  }

  _applyUI() {
    const tabs = this._buildTTYRow(this._buildTabs());
    const urlBar = this._buildURLBar();
    this._current_frame = tabs.concat(urlBar).concat(this._current_frame);
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
