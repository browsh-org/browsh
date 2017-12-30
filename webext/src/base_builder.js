export default class BaseBuilder {
  logPerformance(work, reference) {
    let start = performance.now();
    work();
    let end = performance.now();
    this.firstFrameLog(reference, end - start);
  }

  // If you're logging large objects and using a high-ish FPS (<1000ms) then you might
  // crash the browser. So use this function instead.
  firstFrameLog(...logs) {
    if (this.is_first_frame_finished) return;
    if (BUILD_ENV === 'development') {
      console.log(logs);
    }
  }
}
