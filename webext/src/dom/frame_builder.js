import utils from 'utils';

import CommonMixin from 'dom/common_mixin';
import GraphicsBuilder from 'dom/graphics_builder';
import TextBuilder from 'dom/text_builder';

// Takes the graphics and text from the current DOM, combines them, then
// sends it to the background process where the rest of the UI, like tabs,
// address bar, etc will be added.
export default class extends utils.mixins(CommonMixin) {
  constructor(channel, dimensions) {
    super();
    this.channel = channel;
    this.is_graphics_mode = true;
    this.dimensions = dimensions;
    this.dimensions.update();
    this.graphics_builder = new GraphicsBuilder(channel, dimensions);
    this.text_builder = new TextBuilder(channel, dimensions, this.graphics_builder);
  }

  makeFrame() {
    this.dimensions.update();
    this._compileFrame();
    this._buildFrame();
    this._is_first_frame_finished = true;
  }

  // The user clicks on a TTY grid which has a significantly lower resolution than the
  // actual browser window. So we scale the coordinates up as if the user clicked on the
  // the central "pixel" of a TTY cell.
  //
  // Furthermore if the TTY click is on a readable character then the click is proxied
  // to the original position of the character before TextBuilder snapped the character into
  // position.
  getDOMCoordsFromMouseCoords(x, y) {
    let dom_x, dom_y, char, original_position;
    const index = (y * this.dimensions.frame.width) + x;
    if (this.text_builder.tty_grid.cells[index] !== undefined) {
      char = this.text_builder.tty_grid.cells[index].rune;
    } else {
      char = false;
    }
    if (!char || char === '▄') {
      dom_x = (x * this.dimensions.char.width);
      dom_y = (y * this.dimensions.char.height);
    } else {
      // Recall that text can be shifted from its original position in the browser in order
      // to snap it consistently to the TTY grid.
      original_position = this.text_builder.tty_grid.cells[index].dom_coords;
      dom_x = original_position.x;
      dom_y = original_position.y;
    }
    return [
      dom_x + (this.dimensions.char.width / 2),
      dom_y + (this.dimensions.char.height / 2)
    ];
  }

  _compileFrame() {
    if (!this._is_first_frame_finished || this.dimensions.dom.is_new) {
      this.buildText();
    }
    this.graphics_builder.getScaledScreenshot();
  }

  // All the processes necessary to build the text for a frame. It can be CPU intensive
  // so we don't want to do it all the time.
  buildText() {
    this.text_builder.fixJustifiedText();
    this.graphics_builder.getScreenshotWithText();
    this.graphics_builder.getScreenshotWithoutText();
    this.text_builder.buildFormattedText();
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
    for (let y = 0; y < this.dimensions.frame.height; y++) {
      for (let x = 0; x < this.dimensions.frame.width; x++) {
        this._buildPixel(x, y);
      }
    }
  }

  // Note how we have to keep track of 2 rows of pixels in order to create 1 row of
  // the terminal.
  _buildPixel(x, y) {
    const colour = this.graphics_builder.getScaledPixelAt(x, y);
    if (this._bg_row.length < this.dimensions.frame.width) {
      this._bg_row.push(colour);
    } else {
      this._fg_row.push(colour);
    }
    if (this._fg_row.length === this.dimensions.frame.width) {
      this._buildTTYRow(parseInt(y / 2));
      this.frame = this.frame.concat(this._row);
      this._bg_row = [];
      this._fg_row = [];
    }
  }

  // This is where we implement the UTF8 half-block trick.
  // This is a half-black: "▄", notice how it takes up precisely half a text cell. This
  // means that we can get 2 pixel colours from it, the top pixel comes from setting
  // the background colour and the bottom pixel comes from setting the foreground
  // colour, namely the colour of the text.
  //
  // However we can't just write these "pixels" to a TTY screen, we must collate 2 rows
  // of "pixels" for every row of the terminal and then we can render them together.
  _buildTTYRow(row_number) {
    this._row = [];
    const width = this.dimensions.frame.width
    for (this._current_col = 0; this._current_col < width; this._current_col++) {
      this._cell = this.text_builder.tty_grid.getCellAt(this._current_col, row_number);
      if (this._cell === undefined || this._cell.isTransparent()) {
        this._addGraphicsBlock();
      } else {
        this._handleCharacter();
      }
    }
  }

  _handleCharacter() {
    const padding_count = this._cell.calculateCharWidthPadding();
    if (padding_count > 0) {
      for (let p = 0; p < padding_count; p++) {
        this._current_column++;
        this._addCharacter(' ');
      }
    } else {
      this._addCharacter();
    }
  }

  _addCharacter(override_rune = false) {
    let char;
    if (override_rune) {
      char = override_rune
    } else {
      char = this._cell.rune
    }
    if (this.is_graphics_mode) {
      this._row = this._row.concat(
        utils.ttyCell(this._cell.fg_colour, this._cell.bg_colour, char)
      );
    } else {
      // TODO: Somehow communicate clickable text
      this._row = this._row.concat(utils.ttyPlainCell(char));
    }
  }

  _addGraphicsBlock() {
    const x = this._current_col;
    if (this.is_graphics_mode) {
      this._row = this._row.concat(utils.ttyCell(this._fg_row[x], this._bg_row[x], '▄'));
    } else {
      this._row = this._row.concat(utils.ttyPlainCell(' '));
    }
  }
}
