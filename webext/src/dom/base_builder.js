export default class BaseBuilder {
  _sendMessage(message) {
    this.channel.postMessage(message);
  }

  _snap(number) {
    return parseInt(Math.round(number));
  }

  _log(...messages) {
    if (messages.length === 1) {
      messages = messages[0];
    } else {
      messages = JSON.stringify(messages);
    }
    this._sendMessage(`/log,${messages}`);
  }

  _logPerformance(work, reference) {
    let start = performance.now();
    work();
    let end = performance.now();
    this._firstFrameLog(`${reference}: ${end - start}ms`);
  }

  // If you're logging large objects and using a high-ish FPS (<1000ms) then you might
  // crash the browser. So use this function instead.
  _firstFrameLog(...logs) {
    if (this._is_first_frame_finished) return;
    if (DEVELOPMENT) {
      this._log(logs);
    }
  }
}
