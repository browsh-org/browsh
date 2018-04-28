import utils from 'utils';

import CommonMixin from 'dom/common_mixin';

// Converts an instance of the visible DOM into an array of pixel values.
// Note that it does this both with and without the text visible in order
// to aid in a clean separation of the graphics and text in the final frame
// rendered in the terminal.
export default class extends utils.mixins(CommonMixin) {
  constructor(channel, dimensions) {
    super();
    this.channel = channel;
    this.dimensions = dimensions;
    this._off_screen_canvas = document.createElement('canvas');
    this._ctx = this._off_screen_canvas.getContext('2d');
  }

  sendFrame() {
    this.getScaledScreenshot();
    this._serialiseFrame();
    this.frame.width = this.dimensions.frame.width;
    this.frame.height = this.dimensions.frame.height;
    if (this.frame.colours.length > 0) {
      this.sendMessage(`/frame_pixels,${JSON.stringify(this.frame)}`);
    } else {
      this.log("Not sending empty pixels frame");
    }
  }

  // With full-block single-glyph font on
  getUnscaledFGPixelAt(x, y) {
    const pixel_data_start = parseInt(
      (y * this.dimensions.dom.width * 4) + (x * 4)
    );
    let fg_rgb = this.pixels_with_text.slice(
      pixel_data_start, pixel_data_start + 3
    );
    return [fg_rgb[0], fg_rgb[1], fg_rgb[2]];
  }

  // Without any text showing at all
  getUnscaledBGPixelAt(x, y) {
    const pixel_data_start = parseInt(
      (y * this.dimensions.dom.width * 4) + (x * 4)
    );
    let bg_rgb = this.pixels_without_text.slice(
      pixel_data_start, pixel_data_start + 3
    );
    return [bg_rgb[0], bg_rgb[1], bg_rgb[2]];
  }

  // Scaled to so the size where each pixel is the same size as a TTY cell
  getScaledPixelAt(x, y) {
    const pixel_data_start = (y * this.dimensions.frame.width * 4) + (x * 4);
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
    this.dimensions.update()
    return this._getPixelData();
  }

  // Scale the screenshot so that 1 pixel approximates half a TTY cell.
  _scaleCanvas() {
    this._is_scaled = true;
    this._hideText();
    this._ctx.save();
    this._ctx.scale(
      this.dimensions.scale_factor.width,
      this.dimensions.scale_factor.height
    );
  }

  _unScaleCanvas() {
    this._ctx.restore();
    this._showText();
    this._is_scaled = false;
  }

  _updateCanvasSize() {
    if (this._is_scaled) return;
    this._off_screen_canvas.width = this.dimensions.dom.width;
    this._off_screen_canvas.height = this.dimensions.dom.height;
  }

  // Get an array of RGB values.
  // This is Firefox-only. Chrome has a nicer MediaStream for this.
  _getPixelData() {
    let width, height;
    let background_colour = 'rgb(255,255,255)';
    if (this._is_scaled) {
      width = this.dimensions.frame.width;
      height = this.dimensions.frame.height;
    } else {
      width = this.dimensions.dom.width;
      height = this.dimensions.dom.height;
    }
    this._updateCanvasSize();
    this._ctx.drawWindow(
      window, 0, 0,
      this.dimensions.dom.width,
      this.dimensions.dom.height,
      background_colour
    );
    return this._ctx.getImageData(0, 0, width, height).data;
  }

  _serialiseFrame() {
    this.frame = {
      id: parseInt(this.channel.name),
      colours: []
    };
    const height = this.dimensions.frame.height;
    const width = this.dimensions.frame.width;
    for (let y = 0; y < height; y++) {
      for (let x = 0; x < width; x++) {
        // TODO: Explore sending as binary data
        this.getScaledPixelAt(x, y).map((c) => this.frame.colours.push(c));
      }
    }
  }
}
