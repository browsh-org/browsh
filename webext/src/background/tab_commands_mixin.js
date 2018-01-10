// Handle commands from tabs, like sending a frame or information about
// the current character dimensions .
export default (MixinBase) => class extends MixinBase {
  handleTabMessage(message) {
    const parts = message.split(',');
    const command = parts[0];
    switch (command) {
      case '/frame':
        // TODO: Add UI, tabs, etc
        this.sendToTerminal(`/frame,${parts.slice(1).join(',')}`);
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
};
