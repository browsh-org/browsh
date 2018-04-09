import charWidthInTTY from 'string-width';

import utils from 'utils';
import CommonMixin from 'dom/common_mixin';
import GraphicsBuilderMixin from 'dom/graphics_builder_mixin';
import TextBuilderMixin from 'dom/text_builder_mixin';

// Takes the graphics and text from the current DOM, combines them, then
// sends it to the background process where the rest of the UI, like tabs,
// address bar, etc will be added.
export default class extends utils.mixins(CommonMixin, GraphicsBuilderMixin, TextBuilderMixin) {
  constructor(channel) {
    super();
    this.channel = channel;
    this.is_graphics_mode = true;
  }

  makeFrame() {
    this._setupDimensions();
    this._compileFrame();
    this._buildFrame();
  }

  _compileFrame() {
    this.getScreenshotWithText();
    this.getScreenshotWithoutText();
    this.getScaledScreenshot();
    this.buildFormattedText();
  }

  _buildFrame() {
    this.logPerformance(() => {
      this.__buildFrame();
    }, 'build frame');
  }

  __buildFrame() {
    this.frame = [];
    this._bg_row = [];
    this._fg_row = [];
    for (let y = 0; y < this.frame_height; y++) {
      for (let x = 0; x < this.frame_width; x++) {
        this._buildPixel(x, y);
      }
    }
  }

  _setupDimensions() {
    if (!this.tty_width || !this.tty_height) {
      throw new Error("DocumentBuilder doesn't have the TTY dimensions");
    }
    // A frame is 'taller' than the TTY because of the special UTF8 half-block
    this.frame_width = this.tty_width;
    // trick. Also we need to reserve 2 lines at the top for the tabs and URL bar.
    this.frame_height = (this.tty_height - 2) * 2;
  }

  // Note how we have to keep track of 2 rows of pixels in order to create 1 row of
  // the terminal.
  _buildPixel(x, y) {
    const colour = this.getScaledPixelAt(x, y);
    if (this._bg_row.length < this.frame_width) {
      this._bg_row.push(colour);
    } else {
      this._fg_row.push(colour);
    }
    if (this._fg_row.length === this.frame_width) {
      this._buildTtyRow(this._bg_row, this._fg_row, y);
      this.frame = this.frame.concat(this._row);
      this._bg_row = [];
      this._fg_row = [];
    }
  }

  // This is where we implement the UTF8 half-block trick.
  // This is a half-black: ▄ Notice how it takes up precisely half a text cell. This
  // means that we can get 2 pixel colours from it, the top pixel comes from setting
  // the background colour and the bottom pixel comes from setting the foreground
  // colour, namely the colour of the text.
  // However we can't just write random pixels to a TTY screen, we must collate 2 rows
  // of native pixels for every row of the terminal.
  _buildTtyRow(bg_row, fg_row, y) {
    let tty_index, padding, char;
    this._row = [];
    const tty_row = parseInt(y / 2);
    for (let x = 0; x < this.frame_width; x++) {
      tty_index = (tty_row * this.frame_width) + x;
      if (this._doesCellHaveACharacter(tty_index)) {
        this._addCharacter(tty_index);
        char = this.tty_grid[tty_index][0]
        padding = this._calculateCharWidthPadding(char);
        for (let p = 0; p < padding; p++) {
          x++;
          this._addCharacter(tty_index, ' ');
        }
      } else {
        this._addGraphicsBlock(x, fg_row, bg_row);
      }
    }
  }

  _addCharacter(tty_index, padding = false) {
    const cell = this.tty_grid[tty_index];
    let char = padding ? padding : cell[0];
    const fg = cell[1];
    const bg = cell[2];
    if (this.is_graphics_mode) {
      this._row = this._row.concat(utils.ttyCell(fg, bg, char));
    } else {
      // TODO: Somehow communicate clickable text
      this._row = this._row.concat(utils.ttyPlainCell(char));
    }
  }

  _addGraphicsBlock(x, fg_row, bg_row) {
    if (this.is_graphics_mode) {
      this._row = this._row.concat(utils.ttyCell(fg_row[x], bg_row[x], '▄'));
    } else {
      this._row = this._row.concat(utils.ttyPlainCell(' '));
    }
  }

  // Deal with UTF8 characters that take up more than a single cell in the TTY.
  // TODO:
  //   1. Do all terminals deal with wide characters the same?
  //   2. Use CSS or JS so that wide characters actually flow in the DOM as 2
  //      monospaced characters. This will allow pages of nothing but wide
  //      characters to properly display.
  _calculateCharWidthPadding(char) {
    return charWidthInTTY(char) - 1;
  }

  // We need to know this because we want all empty cells to be 'transparent'
  _doesCellHaveACharacter(index) {
    if (this.tty_grid[index] === undefined) return false;
    const char = this.tty_grid[index][0];
    const is_undefined = char === undefined;
    const is_empty = char === '';
    const is_space = /^\s+$/.test(char);
    const is_not_worth_printing = is_empty || is_space || is_undefined;
    return !is_not_worth_printing;
  }
}
