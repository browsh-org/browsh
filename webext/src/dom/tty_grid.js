import utils from "utils";

// The TTY grid
export default class {
  constructor(dimensions, graphics_builder, config) {
    this.dimensions = dimensions;
    this.graphics_builder = graphics_builder;
    this.config = config;
    this._setMiddleOfEm();
  }

  getCell(index) {
    return this.cells[index];
  }

  getCellAt(x, y) {
    return this.cells[y * this.dimensions.frame.width + x];
  }

  addCell(new_cell) {
    new_cell.index = this._calculateIndex(new_cell);
    const is_cell_possibly_obscured = !this._handleCellVisibility(new_cell);
    const is_cell_at_highest_layer = this._isNewCellAtHighestLayer(new_cell);
    if (is_cell_at_highest_layer && !is_cell_possibly_obscured) {
      this.cells[new_cell.index] = new_cell;
    }
  }

  _isNewCellAtHighestLayer(new_cell) {
    let existing_cell = this.cells[new_cell.index];

    return !(
      existing_cell !== undefined && !new_cell.isHighestLayer(existing_cell)
    );
  }

  _handleCellVisibility(new_cell) {
    const colours = this._getColours(new_cell);
    if (!colours) return false;
    if (this._isCharObscured(colours)) return false;
    new_cell.fg_colour = colours[0];
    new_cell.bg_colour = colours[1];
    return true;
  }

  _calculateIndex(cell) {
    return cell.tty_coords.y * this.dimensions.frame.width + cell.tty_coords.x;
  }

  // Get the colours right in the middle of the character's font. Returns both the colour
  // when the text is displayed and when it's hidden.
  _getColours(cell) {
    const offset_x = utils.snap(
      cell.dom_coords.x + this.dimensions.char.width * this._middle_of_em
    );
    const offset_y = utils.snap(
      cell.dom_coords.y + this.dimensions.char.height * this._middle_of_em
    );
    const fg_colour = this.graphics_builder.getUnscaledFGPixelAt(
      offset_x,
      offset_y
    );
    const bg_colour = this.graphics_builder.getUnscaledBGPixelAt(
      offset_x,
      offset_y
    );
    return [fg_colour, bg_colour];
  }

  // This is the value to reach the middle of a uni-glyph font character in order to
  // sample its colour. Obviously it is better to reach for the middle in case there are
  // vagaries of rendering, it increases our chances of actually getting the characters
  // own colour and not some other colour nearby.
  //
  // However during testing, we use very small self-generated pixel arrays which makes
  // the snapped values rather unintuitive. So we just encourage the snapped values to
  // snap lower which just lends itself to more readable test values.
  _setMiddleOfEm() {
    this._middle_of_em = TEST ? 0.49 : 0.5;
  }

  // This is somewhat of a, hopefully elegant, hack. So, imagine that situation where you're
  // browsing a web page and a popup appears; perhaps just a select box, or menu, or worst
  // of all a dreaded full-page overlay. Now, DOM rectangles don't take into account whether
  // they are the uppermost visible element, so we're left in a bit of a pickle. The only JS
  // way to know if an element is visible is to use `Document.elementFromPoint(x, y)`, where
  // you compare the returned element with the element whose visibility you're checking.
  // This is has a number of problems. Firstly, it only checks one coordinate in the element
  // for visibility, which of course isn't going to 100% reliably speak for all the
  // characters in the element. Secondly, even ignoring the first caveat, running
  // `elementFromPoint()` for every character is very expensive, around 25ms for an average
  // DOM. So it's basically a no-go. So instead we take advantage of the fact that we're
  // working with a snapshot of the webpage's pixels. It's pretty good assumption that if
  // you make the text transparent and a pixel's colour doesn't change then that character
  // must be obscured by something.
  //
  // There are of course some potential edge cases with this. What if we get a false
  // positive, where a character is obscured _by another character_? Hopefully in such a
  // case we can work with `z-index` so that characters justifiably overwrite each other in
  // the TTY grid.
  _isCharObscured(colours) {
    if (!this.config.browsh.use_experimental_text_visibility) {
      return false;
    }
    return (
      colours[0][0] === colours[1][0] &&
      colours[0][1] === colours[1][1] &&
      colours[0][2] === colours[1][2]
    );
  }
}
