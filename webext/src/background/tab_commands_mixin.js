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
        incoming = JSON.parse(utils.rebuildArgsToSingleArg(parts));
        this.currentTab().title = incoming.title
        this.currentTab().url = incoming.url
        this.sendState();
        break;
      case '/dimensions':
        incoming = JSON.parse(message.slice(12));
        this._mightResizeWindow(incoming);
        this.dimensions = incoming;
        break;
      case '/status':
        this.updateStatus(parts[1]);
        this.sendState();
        break;
      case '/log':
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

  updateStatus(status, message = '') {
    let status_message;
    switch (status) {
      case 'page_init':
        status_message = `Loading ${this.currentTab().url}`;
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
    this.state['status_message'] = status_message;
    this.sendState();
  }
};
