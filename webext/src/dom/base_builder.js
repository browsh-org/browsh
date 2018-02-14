export default class BaseBuilder {
  _sendMessage(message) {
    this.channel.postMessage(message);
  }

  _snap(number) {
    return parseInt(Math.round(number));
  }

  _log(...messages) {
    this._sendMessage(`/log,${JSON.stringify(messages)}`);
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
