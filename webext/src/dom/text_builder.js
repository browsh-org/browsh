import _ from 'lodash';

import BaseBuilder from 'dom/base_builder';

// Convert the text on the page into a snapped 2-dimensional grid to be displayed directly
// in the terminal.
export default class TextBuillder extends BaseBuilder {
  constructor(frame_builder) {
    super();
    this.graphics_builder = frame_builder.graphics_builder;
    this.frame_builder = frame_builder;
    this._parse_started_elements = [];
  }

  getFormattedText() {
    this._updateState();
    this._getTextNodes();
    this._positionTextNodes();
    this._is_first_frame_finished = true;
    return this.tty_grid;
  }

  _updateState() {
    this.tty_grid = [];
    this.tty_dom_width = this.frame_builder.tty_width;
    // For Tabs and URL bar.
    this.tty_dom_height = this.frame_builder.tty_height - 2;
    this.char_width = this.frame_builder.char_width;
    this.char_height = this.frame_builder.char_height;
    this.pixels_with_text = this.graphics_builder.pixels_with_text;
    this.pixels_without_text = this.graphics_builder.pixels_without_text;
    this._parse_started_elements = [];
  }

  // This is relatively cheap: around 50ms for a 13,000 word Wikipedia page
  _getTextNodes() {
    this._logPerformance(() => {
      this.__getTextNodes();
    }, 'tree walker');
  }

  // This should be around 125ms for a largish Wikipedia page of 13,000 words
  _positionTextNodes() {
    this._logPerformance(() => {
      this.__positionTextNodes();
    }, 'position text nodes');
  }

  // Search through every node in the DOM looking for displayable text.
  __getTextNodes() {
    this.text_nodes = [];
    const walker = document.createTreeWalker(
      document.body,
      NodeFilter.SHOW_TEXT,
      { acceptNode: this._isRelevantTextNode },
      false
    );
    while(walker.nextNode()) this.text_nodes.push(walker.currentNode);
  }

  // Does the node contain text that we want to display?
  _isRelevantTextNode(node) {
    // Ignore nodes with only whitespace
    if (/^\s+$/.test(node.textContent) || node.textContent === '') {
      return NodeFilter.FILTER_REJECT;
    }
    return NodeFilter.FILTER_ACCEPT;
  }

  // This is the key to being able to display formatted text within the strict confines
  // of a TTY. DOM Rectangles are closely related to selection ranges (like when you click
  // and drag the mouse cursor over text). Think of an individual DOM rectangle as a single
  // bar of highlighted selection. So that, for example, a 3 line paragraph will have 3
  // DOM rectangles. Fortunately DOMRect coordinates and dimensions are precisely defined.
  // Although do note that, unlike selection ranges, sub-selections can appear seemingly
  // inside other selections for things like italics or anchor tags.
  __positionTextNodes() {
    let range = document.createRange();
    let bounding_box;
    for (const node of this.text_nodes) {
      range.selectNode(node);
      bounding_box = range.getBoundingClientRect();
      if (this._isBoxOutsideViewport(bounding_box)) continue;
      this._fixJustifiedText(node);
      this._formatTextForTTYGrid(
        this._normaliseWhitespace(node.textContent, node.parentElement),
        range.getClientRects(),
        node.parentElement
      );
    }
  }

  // If even a single pixel is inside the viewport we need to check it
  _isBoxOutsideViewport(bounding_box) {
    const is_top_in =
      bounding_box.top >= 0 &&
      bounding_box.top < this.graphics_builder.viewport.height;
    const is_bottom_in =
      bounding_box.bottom >= 0 &&
      bounding_box.bottom < this.graphics_builder.viewport.height;
    const is_left_in =
      bounding_box.left >= 0 &&
      bounding_box.left < this.graphics_builder.viewport.width;
    const is_right_in =
      bounding_box.right >= 0 &&
      bounding_box.right < this.graphics_builder.viewport.width;
    return !((is_top_in || is_bottom_in) && (is_left_in || is_right_in));
  }

  // Justified text uses the space between words to stretch a line to perfectly fit from
  // end to end. That'd be ok if it only stretched by exact units of monospace width, but
  // it doesn't, which messes with our fragile grid system.
  // TODO:
  //   * It'd be nice to detect right-justified text so we can keep it. Just need to be
  //     careful with things like traversing parents up the DOM, or using `computedStyle()`
  //     because they can be expensive.
  //   * Another approach could be to explore how a global use of `pre` styling renders
  //     pages.
  //   * Also, is it possible and/or faster to do this once in the main style sheet? Or
  //     even by a find-replace on all occurrences of 'justify'?
  //   * Yet another thing, the style change doesn't actually get picked up until the
  //     next frame.
  _fixJustifiedText(node) {
    node.parentElement.style.textAlign = 'left';
  }

  // The need for this wasn't immediately obvious to me. The fact is that the DOM stores
  // text nodes _as they are written in the HTML doc_. Therefore, if you've written some
  // nicely indented HTML, then the text node will actually contain those as something like
  //   `\n      text starts here`
  // It's just that the way CSS works most of the time means that whitespace is collapsed
  // so viewers never notice.
  //
  // TODO:
  //   The normalisation here of course destroys the formatting of `white-space: pre`
  //   styling, like code snippets for example. So hopefully we can detect the node's
  //   `white-space` setting and skip this function if necessary?
  _normaliseWhitespace(text, parent) {
    text = text.replace(/[\t\n\r ]+/g, " ");
    if (this._isFirstParseInElement(parent)) {
      if (text.charAt(0) === " ") text = text.substring(1, text.length);
    }
    if (text.charAt(text.length - 1) === " ") text = text.substring(0, text.length - 1);
    return text;
  }

  // An element may contain many text nodes. For example a `<p>` element may contain a
  // starting text node followed by a `<a>` tag, finishing with another plain text node. We
  // only want to remove leading whitespace from the text at the _beginning_ of a line.
  // Usually we can do this just by checking if a DOM rectangle's position is further down
  // the page than the previous one - but of course there is nothing to compare the first
  // DOM rectangle to. What's more, DOM rects are grouped per _text node_, NOT per element
  // and we are not guaranteed to iterate through elements in the order that text flows.
  // Therefore we need to make the assumption that plain text nodes flow within their shared
  // parent element. There is a possible caveat here for elements starting with another
  // element (like a link), where that sub-element contains leading whitespace.
  _isFirstParseInElement(element) {
    const is_parse_started = _.includes(this._parse_started_elements, element);
    if (is_parse_started) {
      return false
    } else {
      this._parse_started_elements.push(element);
      return true
    }
  }

  // Here is where we actually make use of the rather strict monospaced and fixed font size
  // CSS rules enforced by the webextension. Of course the CSS is never going to be able to
  // perfectly snap characters onto a grid, so we force it here instead. At least we can be
  // fairly certain that every character at least takes up the same space as a TTY cell, it
  // just might not be perfectly aligned. So here we just round down all coordinates to force
  // the snapping.
  // Use `this.addClientRectsOverlay(dom_rects, text);` to see DOM rectangle outlines in a
  // real browser.
  _formatTextForTTYGrid(text, dom_rects, parent_element) {
    let col, tty_box, step, character, previous_box, origin;
    this.char_width_debt = 0
    let character_index = 0;
    for (const box of dom_rects) {
      if (this._isBoxOutsideViewport(box)) return;
      if (this._isNewLine(previous_box, box)) {
        character = text.charAt(character_index);
        if (/[\t\n\r ]+/.test(character)) character_index++;
      }
      tty_box = this._convertBoxToTTYUnits(box);
      col = tty_box.col_start;
      origin = {
        x: parseInt(Math.round(box.left)),
        y: parseInt(Math.round(box.top))
      }
      for (step = 0; step < tty_box.width; step++) {
        character = text.charAt(character_index);
        this._placeCharacterOnTTYGrid(col, tty_box.row, origin, character, parent_element);
        origin.x = origin.x + this.char_width;
        character_index++;
        col++
      }
      previous_box = box;
    }
  }

  // Is the current DOM rectangle further down the page than the previous?
  _isNewLine(previous_box, current_box) {
    if (previous_box === undefined) return false;
    return current_box.top > previous_box.top
  }

  // Round and snap a DOM rectangle as if it were placed in the terminal
  _convertBoxToTTYUnits(viewport_dom_rect) {
    return {
      col_start: this._snap(viewport_dom_rect.left / this.char_width),
      row: this._snap(viewport_dom_rect.top / this.char_height),
      width: this._snap(viewport_dom_rect.width / this.char_width),
    }
  }

  _placeCharacterOnTTYGrid(col, row, original_position, character, parent_element) {
    const index = (row * this.tty_dom_width) + col;
    if (this._isExistingCharacter(index)) {
      if (!this._isHighestLayer(index, parent_element)) return;
    }
    if (this._isCharOutsideGrid(col, row)) return;
    const colours = this._getCharacterColours(original_position);
    if (!colours) return;
    if (this._isCharObscured(colours)) return;
    this.tty_grid[index] = [
      character,
      ...colours,
      parent_element,
      _.clone(original_position)
    ];
  }

  // Don't clobber - for now at least.
  // TODO: Use `getComputedStyles()` and save for the whole parent element.
  _isExistingCharacter(index) {
    return !!this.tty_grid[index];
  }

  // When a character clobbers another character in the grid, we can't use our
  // text show/hide trick to know if the character is visible in the final DOM. So we have
  // to use standard CSS inspection instead. Hopefully this doesn't happen often because
  // it's expensive.
  // TODO: Make comprehensive
  _isHighestLayer(index_of_tenant, parent_element_of_challenger) {
    const tenant_styles = this._getStyles(this.tty_grid[index_of_tenant][3]);
    const challenger_styles = this._getStyles(parent_element_of_challenger);
    if (
      challenger_styles.visibility === 'hidden' ||
      challenger_styles.display === 'none'
    ) {
      return false;
    }
    return tenant_styles.zIndex < challenger_styles.zIndex;
  }

  // Get or cache the total cascaded calculated styles for an element
  _getStyles(element) {
    if (!element.browsh_calculated_styles) {
      let styles = window.getComputedStyle(element);
      element.browsh_calculated_styles = styles;
    }
    return element.browsh_calculated_styles;
  }

  // Get the colours right in the middle of the character's font. Returns both the colour
  // when the text is displayed and when it's hidden.
  _getCharacterColours(original_position) {
    // Don't use a full half, just because it means that we can use very small mock pixel
    // arrays during testing - rounding to the top-left saves having to write and extra
    // column and row.
    const half = 0.449;
    const offset_x = this._snap(original_position.x + (this.char_width * half));
    const offset_y = this._snap(original_position.y + (this.char_height * half));
    if (this._isCharCentreOutsideViewport(offset_x, offset_y)) return false;
    return this.graphics_builder.getPixelsAt(offset_x, offset_y);
  }

  // Check if the char is in the viewport again because of x increments, y potentially
  // being rounded up and of course the offset to make sure the sample is within the
  // unicode block.
  _isCharCentreOutsideViewport(x, y) {
    if (
      x >= this.graphics_builder.viewport.width ||
      x < 0 ||
      y >= this.graphics_builder.viewport.height ||
      y < 0
    ) return false;
  }

  // Theoretically this should only be needed for DOM rectangles that _straddle_ the
  // viewport.
  _isCharOutsideGrid(col, row) {
    return col >= this.tty_dom_width || row >= this.tty_dom_height;
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
  // working with a snapshot of the the webpage's pixels. It's pretty good assumption that if
  // you make the text transparent and a pixel's colour doesn't change then that character
  // must be obscured by something.
  // There are of course some potential edge cases with this. What if we get a false
  // positive, where a character is obscured _by another character_? Hopefully in such a
  // case we can work with `z-index` so that characters justifiably overwrite each other in
  // the TTY grid.
  _isCharObscured(colours) {
    return colours[0][0] === colours[1][0] &&
           colours[0][1] === colours[1][1] &&
           colours[0][2] === colours[1][2];
  }

  // Purely for debugging. Draws a red border around all the DOMClientRect nodes.
  // Based on code from the MDN docs site.
  _addClientRectsOverlay(dom_rects, normalised_text) {
    // Don't draw on every frame
    if (this.is_first_frame_finished) return;
    // Absolutely position a div over each client rect so that its border width
    // is the same as the rectangle's width.
    // Note: the overlays will be out of place if the user resizes or zooms.
    for (const rect of dom_rects) {
      let tableRectDiv = document.createElement('div');
      // A DOMClientRect object only contains dimensions, so there's no way to identify it
      // to a node, so let's put its text as an attribute so we can cross-check if needs be.
      tableRectDiv.setAttribute('browsh-text', normalised_text);
      let tty_row = parseInt(Math.round(rect.top / this.char_height));
      tableRectDiv.setAttribute('tty_row', tty_row);
      tableRectDiv.style.position = 'absolute';
      tableRectDiv.style.border = '1px solid red';
      let scrollTop = document.documentElement.scrollTop || document.body.scrollTop;
      let scrollLeft = document.documentElement.scrollLeft || document.body.scrollLeft;
      tableRectDiv.style.margin = tableRectDiv.style.padding = '0';
      tableRectDiv.style.top = (rect.top + scrollTop) + 'px';
      tableRectDiv.style.left = (rect.left + scrollLeft) + 'px';
      // We want rect.width to be the border width, so content width is 2px less.
      tableRectDiv.style.width = (rect.width - 2) + 'px';
      tableRectDiv.style.height = (rect.height - 2) + 'px';
      document.body.appendChild(tableRectDiv);
    }
  }
}
