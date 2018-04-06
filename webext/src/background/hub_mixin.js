import stripAnsi from 'strip-ansi';

// Here we keep the public functions used to mediate communications between
// the background process, tabs and the terminal.
export default (Base) => class extends Base {
  sendToCurrentTab(message) {
    this.currentTab().channel.postMessage(message);
  }

  sendToTerminal(message) {
    this.terminal.send(message);
  }

  log(...messages) {
    if (messages.length === 1) {
      messages = messages[0];
      messages = stripAnsi(messages);
      messages = messages.replace(/\u001b\[/g, 'ESC');
    }
    this.sendToTerminal(messages);
  }

  currentTab() {
    return this.tabs[this.active_tab_id];
  }
}
