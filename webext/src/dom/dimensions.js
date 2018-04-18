import utils from 'utils';

import CommonMixin from 'dom/common_mixin';

// All the various dimensions, sizes, scales, etc
export default class extends utils.mixins(CommonMixin) {
  constructor() {
    super()
    // ID for element we place in the DOM to measure the size of a single monospace
    // character.
    this._measuring_box_id = 'browsh_em_measuring_box';
    // TODO: WTF is this magic number?
    if (TEST) {
      this._char_height_magic_number = 2
    } else {
      this._char_height_magic_number = 4
    }
    this.dom = {};
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
      // Even though it is the TTY's responsibility to scroll the DOM, the browser still
      // needs to do scrolling because various events can be triggered by it - think of
      // lazy image loading.
      x_scroll: window.scrollX,
      y_scroll: window.scrollY,
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
    this.frame = {
      width: utils.snap(this.dom.width * this.scale_factor.width),
      height: utils.snap(this.dom.height * this.scale_factor.height)
    }
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

