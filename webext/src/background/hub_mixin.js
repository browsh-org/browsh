import stripAnsi from 'strip-ansi';

// Here we keep the public funcntions used to mediate communications between
// the background process, tabs and the terminal.
export default (Base) => class extends Base {
  sendToCurrentTab(message) {
    this.currentTab().channel.postMessage(message);
  }

  sendToTerminal(message) {
    this.terminal.send(message);
  }

  log(...message) {
    if (message.length === 1) {
      message = message[0];
      message = stripAnsi(message);
      message = message.replace(/\u001b\[/g, 'ESC');
    }
    this.sendToTerminal(message);
  }

  currentTab() {
    return this.tabs[this.active_tab_id];
  }
}
