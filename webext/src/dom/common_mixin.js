export default (MixinBase) => class extends MixinBase {
  sendMessage(message) {
    this.channel.postMessage(message);
  }

  snap(number) {
    return parseInt(Math.round(number));
  }

  log(...messages) {
    this.sendMessage(`/log,${JSON.stringify(messages)}`);
  }

  logPerformance(work, reference) {
    let start = performance.now();
    work();
    let end = performance.now();
    this.firstFrameLog(`${reference}: ${end - start}ms`);
  }

  // If you're logging large objects and using a high-ish FPS (<1000ms) then you might
  // crash the browser. So use this function instead.
  firstFrameLog(...logs) {
    if (this._is_first_frame_finished) return;
    if (DEVELOPMENT) {
      this.log(logs);
    }
  }
}
