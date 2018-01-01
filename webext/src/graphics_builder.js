import BaseBuilder from 'base_builder';

// Converts an instance of the viewport into a an array of pixel values.
// Note, that it does this both with and without the text visible in order
// to aid in a clean separation of the graphics and text in the final frame
// rendered in the terminal.
export default class GraphicsBuilder extends BaseBuilder {
  constructor() {
    super();
    this.off_screen_canvas = document.createElement('canvas');
    this.ctx = this.off_screen_canvas.getContext('2d');
    this._updateCurrentViewportDimensions();
  }

  getPixelsAt(x, y) {
    const pixel_data_start = parseInt(
      (y * (this.viewport.width * 4)) + (x * 4)
    );
    let fg_colour = this.pixels_with_text.slice(
      pixel_data_start, pixel_data_start + 3
    );
    let bg_colour = this.pixels_without_text.slice(
      pixel_data_start, pixel_data_start + 3
    );
    return [fg_colour, bg_colour];
  }

  getScaledPixelAt(x, y) {
    const pixel_data_start = (y * this.frame_width * 4) + (x * 4);
    return this.scaled_pixels.slice(pixel_data_start, pixel_data_start + 3);
  }

  getSnapshotWithText() {
    this._logPerformance(() => {
      this._getSnapshotWithText();
    }, 'get snapshot with text');
  }

  getSnapshotWithoutText() {
    this._logPerformance(() => {
      this._getSnapshotWithoutText();
    }, 'get snapshot without text');
  }

  getScaledSnapshot(frame_width, frame_height) {
    this._logPerformance(() => {
      this._getScaledSnapshot(frame_width, frame_height);
    }, 'get scaled snapshot');
  }

  _getSnapshotWithoutText() {
    this._hideText();
    this.pixels_without_text = this._getSnapshot();
    this._showText();
    return this.pixels_without_text;
  }

  _getSnapshotWithText() {
    this.pixels_with_text = this._getSnapshot();
    return this.pixels_with_text;
  }

  _getScaledSnapshot(frame_width, frame_height) {
    this.frame_width = frame_width;
    this.frame_height = frame_height;
    this._scaleCanvas();
    this.scaled_pixels = this._getSnapshot();
    this._unScaleCanvas();
    this._is_first_frame_finished = true;
    return this.scaled_pixels;
  }

  _hideText() {
    this.styles = document.createElement("style");
    document.head.appendChild(this.styles);
    this.styles.sheet.insertRule(
      'html * {' +
      '  color: transparent !important;' +
      // Note the disabling of transition effects here. Some websites have a fancy fade
      // animation when changing colours, which we don't have time for in taking a snapshot.
      // However, a drawback here is that, when we remove this style the transition actually
      // kicks in - not that the terminal sees it because, by the nature of this style change
      // here, we only ever capture the screen when text is invisible. However, I wonder if
      // triggering color transitions for every frame might add some unnecessary load? What
      // about permanently disabling color transitions in the global stylesheet?
      '  transition: color 0s !important;' +
      '}'
    );
  }

  _showText() {
    this.styles.parentNode.removeChild(this.styles);
  }

  _getSnapshot() {
    this._updateCurrentViewportDimensions()
    let pixel_data = this._getPixelData();
    return pixel_data;
  }

  // Deal with page scrolling and other viewport changes.
  // Perhaps the window has been resized to better accommodate text-sizing, or to try
  // to trigger some mobile responsive CSS.
  _updateCurrentViewportDimensions() {
    this.viewport = {
      x_scroll: window.scrollX,
      y_scroll: window.scrollY,
      width: window.innerWidth,
      height: window.innerHeight
    }
    if (!this.is_scaled) {
      // Resize our canvas to match the viewport. I guess this makes for efficient
      // use of memory?
      this.off_screen_canvas.width = this.viewport.width;
      this.off_screen_canvas.height = this.viewport.height;
    }
  }

  // Scale the screenshot so that 1 pixel approximates half a TTY cell.
  _scaleCanvas() {
    this.is_scaled = true;
    const scale_x = this.frame_width / this.viewport.width;
    const scale_y = this.frame_height / this.viewport.height;
    this._hideText();
    this.ctx.save();
    this.ctx.scale(scale_x, scale_y);
  }

  _unScaleCanvas() {
    this.ctx.restore();
    this._showText();
    this.is_scaled = false;
  }

  // Get an array of RGB values.
  // This is Firefox-only. Chrome has a nicer MediaStream for this.
  _getPixelData() {
    let width, height;
    let background_colour = 'rgb(255,255,255)';
    if (this.is_scaled) {
      width = this.frame_width;
      height = this.frame_height;
    } else {
      width = this.viewport.width;
      height = this.viewport.height;
    }
    this.ctx.drawWindow(
      window,
      this.viewport.x_scroll,
      this.viewport.y_scroll,
      this.viewport.width,
      this.viewport.height,
      background_colour
    );
    return this.ctx.getImageData(0, 0, width, height).data;
  }
}
