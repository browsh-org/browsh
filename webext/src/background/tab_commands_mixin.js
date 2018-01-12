import utils from 'utils';

// Handle commands from tabs, like sending a frame or information about
// the current character dimensions .
export default (MixinBase) => class extends MixinBase {
  handleTabMessage(message) {
    const parts = message.split(',');
    const command = parts[0];
    switch (command) {
      case '/frame':
        this._applyUI(utils.rebuildArgsToSingleArg(parts));
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

  _applyUI(dom_frame) {
    const tabs = this._buildTTYRow(this.currentTab().info.title);
    const urlBar = this._buildURLBar();
    const full_frame = tabs + urlBar + dom_frame;
    this.sendToTerminal(`/frame,${full_frame}`);
  }

  _buildTTYRow(text) {
    let char;
    let row = "";
    for (let index = 0; index < this.tty_width; index++) {
      if (index < text.length) {
        char = text[index];
      } else {
        char = " "
      }
      row += utils.ttyPixel([255, 255, 255], [0, 0, 0], char);
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
    return this._buildTTYRow(content);
  }
};
