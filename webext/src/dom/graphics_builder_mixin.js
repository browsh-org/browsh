// Converts an instance of the visible DOM into an array of pixel values.
// Note that it does this both with and without the text visible in order
// to aid in a clean separation of the graphics and text in the final frame
// rendered in the terminal.
export default (MixinBase) => class extends MixinBase {
  constructor() {
    super();
    this._off_screen_canvas = document.createElement('canvas');
    this._ctx = this._off_screen_canvas.getContext('2d');
    this._updateCurrentViewportDimensions();
  }

  getPixelsAt(x, y) {
    const pixel_data_start = parseInt(
      (y * (this.viewport.width * 4)) + (x * 4)
    );
    let fg_rgb = this.pixels_with_text.slice(
      pixel_data_start, pixel_data_start + 3
    );
    let bg_rgb = this.pixels_without_text.slice(
      pixel_data_start, pixel_data_start + 3
    );
    return [
      [fg_rgb[0], fg_rgb[1], fg_rgb[2]],
      [bg_rgb[0], bg_rgb[1], bg_rgb[2]]
    ]
  }

  getScaledPixelAt(x, y) {
    const pixel_data_start = (y * this.frame_width * 4) + (x * 4);
    const rgb = this.scaled_pixels.slice(pixel_data_start, pixel_data_start + 3);
    return [rgb[0], rgb[1], rgb[2]];
  }

  getScreenshotWithText() {
    this.logPerformance(() => {
      this._getScreenshotWithText();
    }, 'get screenshot with text');
  }

  getScreenshotWithoutText() {
    this.logPerformance(() => {
      this._getScreenshotWithoutText();
    }, 'get screenshot without text');
  }

  getScaledScreenshot() {
    this.logPerformance(() => {
      this._getScaledScreenshot();
    }, 'get scaled screenshot');
  }

  _getScreenshotWithoutText() {
    this._hideText();
    this.pixels_without_text = this._getScreenshot();
    this._showText();
    return this.pixels_without_text;
  }

  _getScreenshotWithText() {
    this.pixels_with_text = this._getScreenshot();
    return this.pixels_with_text;
  }

  _getScaledScreenshot() {
    this._scaleCanvas();
    this.scaled_pixels = this._getScreenshot();
    this._unScaleCanvas();
    this._is_first_frame_finished = true;
    return this.scaled_pixels;
  }

  _hideText() {
    this._styles = document.createElement("style");
    document.head.appendChild(this._styles);
    this._styles.sheet.insertRule(
      'html * {' +
      '  color: transparent !important;' +
      // Note the disabling of transition effects here. Some websites have a fancy fade
      // animation when changing colours, which we don't have time for in taking a screenshot.
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
    this._styles.parentNode.removeChild(this._styles);
  }

  _getScreenshot() {
    this._updateCurrentViewportDimensions()
    return this._getPixelData();
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
    if (!this._is_scaled) {
      // Resize our canvas to match the viewport. I guess this makes for efficient
      // use of memory?
      this._off_screen_canvas.width = this.viewport.width;
      this._off_screen_canvas.height = this.viewport.height;
    }
  }

  // Scale the screenshot so that 1 pixel approximates half a TTY cell.
  _scaleCanvas() {
    this._is_scaled = true;
    const scale_x = this.frame_width / this.viewport.width;
    const scale_y = this.frame_height / this.viewport.height;
    this._hideText();
    this._ctx.save();
    this._ctx.scale(scale_x, scale_y);
  }

  _unScaleCanvas() {
    this._ctx.restore();
    this._showText();
    this._is_scaled = false;
  }

  // Get an array of RGB values.
  // This is Firefox-only. Chrome has a nicer MediaStream for this.
  _getPixelData() {
    let width, height;
    let background_colour = 'rgb(255,255,255)';
    if (this._is_scaled) {
      width = this.frame_width;
      height = this.frame_height;
    } else {
      width = this.viewport.width;
      height = this.viewport.height;
    }
    this._ctx.drawWindow(
      window,
      this.viewport.x_scroll,
      this.viewport.y_scroll,
      this.viewport.width,
      this.viewport.height,
      background_colour
    );
    return this._ctx.getImageData(0, 0, width, height).data;
  }
}
