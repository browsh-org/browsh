import utils from 'utils';

import CommonMixin from 'dom/common_mixin';

// All the various dimensions, sizes, scales, etc
export default class extends utils.mixins(CommonMixin) {
  constructor() {
    super()

    // ID for element we place in the DOM to measure the size of a single monospace
    // character.
    this._measuring_box_id = 'browsh_em_measuring_box';

    if (TEST) {
      this._char_height_magic_number = 0;
    } else {
      // TODO: WTF is this magic number?
      this._char_height_magic_number = 4;
    }

    // This is the region outside the visible area of the TTY that is pre-parsed and
    // sent to the TTY to be buffered to support faster scrolling.
    this._big_sub_frame_factor = 6;

    this.dom = {};
    this.tty = {};
    this.frame = {
      x_scroll: 0,
      y_scroll: 0,
      x_last_big_frame: 0,
      y_last_big_frame: 0
    }
    if (document.body) {
      this.update();
    }
  }

  update() {
    this._calculateCharacterDimensions();
    this._updateDOMDimensions();
    this._calculateScaleFactor();
    this._updateFrameDimensions();
    this._notifyBackground();
  }

  setSubFrameDimensions(size) {
    this._calculateSmallSubFrame();
    if (size === 'big' || size === 'all') {
      this._calculateBigSubFrame();
    }
    if (size === 'raw_text') {
      this._calculateEntireDOMFrames();
    }
    // Only the height needs to be even because of the UTF8 half-block trick. A single
    // TTY cell always contains exactly 2 pseudo pixels.
    this.frame.sub.height = utils.ensureEven(this.frame.sub.height);
  }

  // This is the data that is sent with the JSON payload of every frame to the TTY
  getFrameMeta() {
    return {
      sub_left: utils.snap(this.frame.sub.left),
      sub_top: utils.snap(this.frame.sub.top),
      sub_width: utils.snap(this.frame.sub.width),
      sub_height: utils.snap(this.frame.sub.height),
      total_width: utils.snap(this.frame.width),
      total_height: utils.snap(this.frame.height)
    }
  }

  // This is the sub frame that is the view onto the frame that is visible by the user
  // in the TTY at any given time.
  _calculateSmallSubFrame() {
    this.frame.sub = {
      left: this.frame.x_scroll,
      top: this.frame.y_scroll,
      width: this.tty.width,
      height: this.tty.height * 2
    }

    this._scaleSubFrameToSubDOM()
  }

  // This is the sub frame that is a few factors bigger than what the user can see
  // in the TTY.
  _calculateBigSubFrame() {
    this.frame.sub = {
      left: this.frame.x_scroll - (this._big_sub_frame_factor * this.tty.width),
      top: this.frame.y_scroll - (this._big_sub_frame_factor * this.tty.height),
      width: this.tty.width + (this._big_sub_frame_factor * 2 * this.tty.width),
      height: this.tty.height + (this._big_sub_frame_factor * 2 * this.tty.height),
    }
    this._limitSubFrameDimensions();
    this._scaleSubFrameToSubDOM();
  }

  // The raw text frames requested through the Browsh HTTP server need to be built from the
  // entire DOM, not just a small window onto the DOM.
  _calculateEntireDOMFrames() {
    this.dom.sub = {
      left: 0,
      top: 0,
      width: this.dom.width,
      height: this.dom.height,
    }
    this.frame.sub = {
      left: 0,
      top: 0,
      width: this.dom.sub.width * this.scale_factor.width,
      height: this.dom.sub.height * this.scale_factor.height
    }
  }

  _limitSubFrameDimensions() {
    if (this.frame.sub.left < 0) { this.frame.sub.left = 0 }
    if (this.frame.sub.top < 0) { this.frame.sub.top = 0 }
    if (this.frame.sub.width > this.frame.width) {
      this.frame.sub.width = this.frame.width;
    }
    if (this.frame.sub.height > this.frame.height) {
      this.frame.sub.height = this.frame.height;
    }
  }

  _scaleSubFrameToSubDOM() {
    this.dom.sub = {
      left: this.frame.sub.left / this.scale_factor.width,
      top: this.frame.sub.top / this.scale_factor.height,
      width: this.frame.sub.width / this.scale_factor.width,
      height: this.frame.sub.height / this.scale_factor.height
    }
  }

  // This is critical in order for the terminal to match the browser as closely as possible.
  // Ideally we want the browser's window size to be exactly multiples of the terminal's
  // dimensions. So if the terminal is 80x40 and the font-size is 12px (12x6 pixels), then
  // the window should be 480x480. Also knowing the precise font-size helps the text builder
  // map un-snapped text to the best grid cells - grid cells that represent the terminal's
  // character positions.
  //
  // The reason that we can't just do some basic maths on the CSS `font-size` value we enforce
  // is that there are various factors that can skew the actual font dimensions on the page.
  // For instance, different browser families and even different versions of the same browser
  // may have subtle differences in how they render text. Furthermore we can actually get
  // floating point accuracy if we use `Element.getBoundingClientRect()` which further helps
  // as calculations are compounded during our rendering processes.
  _calculateCharacterDimensions() {
    const element = this._getOrCreateMeasuringBox();
    const dom_rect = element.getBoundingClientRect();
    this.char = {
      width: dom_rect.width,
      height: dom_rect.height + this._char_height_magic_number
    }
  }

  // Back when printing was done by physical stamps, it was convention to measure the
  // font-size using the letter 'M', thus where we get the unit 'em' from. Not that it
  // should not make any difference to us, but it's nice to keep a tradition.
  _getOrCreateMeasuringBox() {
    let measuring_box = this.findMeasuringBox();
    if (measuring_box) return measuring_box;
    measuring_box = document.createElement('span');
    measuring_box.id = this._measuring_box_id;
    measuring_box.style.visibility = 'hidden';
    var M = document.createTextNode('M');
    measuring_box.appendChild(M);
    document.body.appendChild(measuring_box);
    return measuring_box;
  }

  findMeasuringBox() {
    return document.getElementById(this._measuring_box_id);
  }

  _updateDOMDimensions() {
    const [new_width, new_height] = this._calculateDOMDimensions();
    const is_new = this.dom.width != new_width || this.dom.height != new_height
    this.dom = {
      sub: this.dom.sub,
      width: new_width,
      height: new_height,
      is_new: is_new
    }
  }

  // For discussion on various methods to get total scrollable DOM dimensions, see:
  // https://stackoverflow.com/a/44077777/575773
  _calculateDOMDimensions() {
    let width = document.documentElement.scrollWidth;
    if (window.innerWidth > width) width = window.innerWidth;
    let height = document.documentElement.scrollHeight;
    if (window.innerHeight > height) height = window.innerHeight;
    return [width, height]
  }

  // A frame represents the entire DOM page. Its height usually extends below the window's
  // bottom and occasionally extends beyond the sides too.
  //
  // Note that it treats the height of a single TTY cell as containing 2 pixels. Therefore
  // a TTY of 4x4 will have frame dimensions of 4x8.
  _updateFrameDimensions() {
    let width = this.dom.width * this.scale_factor.width;
    let height = this.dom.height * this.scale_factor.height;
    this.frame.width = utils.snap(width);
    this.frame.height = utils.snap(height);
  }

  // The scale factor is the ratio of the TTY's representation of the DOM to the browser's
  // representation of the DOM. The idea is that the TTY just represents a very low
  // resolution version of the browser - though note that the TTY has the significant
  // benefit of being able to display native fonts (possibly even retina-like high DPI
  // fonts). So Browsh's enforced CSS rules reorient the browser page to render all text
  // at the same monospaced sized - in this sense, theoretically, the TTY and the browser
  // should essentially be facsimilies of each other. However of course the TTY is limited
  // by its cell size in how it renders "pixels", namely pseudo pixels using the UTF8
  // block trick.
  //
  // All of which is to say that the fundamental relationship between the browser's dimensions
  // and the TTY's dimensions is represented by a TTY cell - that which displays a single
  // character. So if we know how many characters fit into the DOM, then we know how many
  // "pixels" the TTY should have.
  _calculateScaleFactor() {
    this.scale_factor = {
      width: 1 / this.char.width,
      // Recall that 2 UTF8 half-black "pixels" can fit into a single TTY cell
      height: 2 / this.char.height
    }
  }

  _notifyBackground() {
    const dimensions = {
      dom: this.dom,
      frame: this.frame,
      char: this.char
    }
    this.sendMessage(`/dimensions,${JSON.stringify(dimensions)}`)
  }
}

