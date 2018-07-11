import stripAnsi from "strip-ansi";

// Here we keep the public functions used to mediate communications between
// the background process, tabs and the terminal.
export default MixinBase =>
  class extends MixinBase {
    sendToCurrentTab(message) {
      if (this.currentTab().channel === undefined) {
        this.log(`Attempting to send "${message}" to tab without a channel`);
      } else {
        this.currentTab().channel.postMessage(message);
      }
    }

    sendToTerminal(message) {
      if (this.terminal === undefined) {
        return;
      }
      if (this.terminal.readyState === 1) {
        this.terminal.send(message);
      }
    }

    log(...messages) {
      if (messages === undefined) {
        messages = "undefined";
      }
      if (messages.length === 1) {
        messages = messages[0].toString();
        messages = stripAnsi(messages);
        messages = messages.replace(/\u001b\[/g, "ESC");
      }
      this.sendToTerminal(messages);
    }

    currentTab() {
      return this.tabs[this.active_tab_id];
    }
  };
