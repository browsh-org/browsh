import utils from 'utils';

// Handle commands from tabs, like sending a frame or information about
// the current character dimensions.
export default (MixinBase) => class extends MixinBase {
  // TODO: There needs to be some consistency in this message sending protocol.
  //       Eg; always requiring JSON.
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
        this._updateTabInfo(incoming);
        break;
      case '/dimensions':
        incoming = JSON.parse(message.slice(12));
        this.dimensions.setCharValues(incoming.char);
        break;
      case '/status':
        this.updateStatus(parts[1], parts[2]);
        break;
      case '/log':
        this.log(message.slice(5));
        break;
      default:
        this.log('Unknown command from tab to background', message);
    }
  }

  _updateTabInfo(incoming) {
    this.title = incoming.title;
    this.url = incoming.url;
    this.sendStateToTerminal();
  }
};
