import _ from "lodash";

import utils from "utils";
import CommonMixin from "dom/common_mixin";
import SerialiseMixin from "dom/serialise_mixin";
import TTYCell from "dom/tty_cell";
import TTYGrid from "dom/tty_grid";

// Convert the text on the page into a snapped 2-dimensional grid to be displayed directly
// in the terminal.
export default class extends utils.mixins(CommonMixin, SerialiseMixin) {
  constructor(channel, dimensions, graphics_builder) {
    super();
    this.channel = channel;
    this.dimensions = dimensions;
    this.graphics_builder = graphics_builder;
    this.tty_grid = new TTYGrid(dimensions, graphics_builder);
    this._parse_started_elements = [];
    // A `range` is the DOM's representation of elements and nodes as they are rendered in
    // the DOM. Think of the 'range' that is created when you select/highlight text for
    // copy-pasting, those usually blue-ish rectangles around the selected text are ranges.
    this._range = document.createRange();
  }

  sendFrame() {
    this.buildFormattedText();
    this._sendFrame();
  }

  sendRawText(type) {
    this._raw_mode_type = type;
    // TODO:
    //   The presence of the `getScreenshotWithText()` and `setTimeout()` calls are a hack
    //   that I am unable to understand the reasoning for - unfortunately they came about
    //   by trial and error :( Without them the the returned raw text is largely empty.
    this.graphics_builder.getScreenshotWithText();
    setTimeout(() => {
      this.buildFormattedText();
      this._sendRawText();
    }, 400);
  }

  buildFormattedText() {
    this._updateState();
    this.graphics_builder.getScreenshotWithoutText();
    this.graphics_builder.getScreenshotWithText();
    this._getTextNodes();
    this._positionTextNodes();
  }

  _updateState() {
    this.tty_grid.cells = [];
    this._parse_started_elements = [];
    this._previous_dom_box = {};
    this._convertSubFrameToViewportCoords();
  }

  // This is relatively cheap: around 50ms for a 13,000 word Wikipedia page
  _getTextNodes() {
    this.logPerformance(() => {
      this.__getTextNodes();
    }, "tree walker");
  }

  // This should be around ?? for a largish Wikipedia page of 13,000 words
  _positionTextNodes() {
    this.logPerformance(() => {
      this.__positionTextNodes();
    }, "position text nodes");
  }

  _serialiseFrame() {
    this.logPerformance(() => {
      this.__serialiseFrame();
    }, "serialise text frame");
  }

  // Search through every node in the DOM looking for displayable text.
  __getTextNodes() {
    this._text_nodes = [];
    const walker = document.createTreeWalker(
      document.body,
      NodeFilter.SHOW_TEXT,
      null,
      false
    );
    while (walker.nextNode()) {
      if (this._isRelevantTextNode(walker.currentNode)) {
        this._text_nodes.push(walker.currentNode);
      }
    }
  }

  // Does the node contain text that we want to parse?
  _isRelevantTextNode(node) {
    // Ignore text outside of the sub-frame, therefore outside either the TTY view or
    // outside the larger buffered TTY view.
    const dom_rect = node.parentElement.getBoundingClientRect();
    if (!this._isDOMRectInSubFrame(dom_rect)) {
      return false;
    }

    // Ignore nodes with only whitespace
    if (node.textContent.trim().length === 0) {
      return false;
    }

    return true;
  }

  // In order to decide if a particular DOM rect is inside the current sub frame then we need
  // to compare the sub frame's dimensions to those of the DOM rect. However DOM rects are in
  // viewport-relative coords. In order to save on some CPU cycles, we can just apply the
  // transform to the sub frame.
  _convertSubFrameToViewportCoords() {
    this._viewport_relative_sub_frame = {
      top: this.dimensions.dom.sub.top - window.scrollY,
      bottom:
        this.dimensions.dom.sub.top +
        this.dimensions.dom.sub.height -
        window.scrollY,
      left: this.dimensions.dom.sub.left - window.scrollX,
      right:
        this.dimensions.dom.sub.left +
        this.dimensions.dom.sub.width -
        window.scrollX
    };
  }

  _isDOMRectInSubFrame(dom_rect) {
    const isBottomIn =
      dom_rect.bottom >= this._viewport_relative_sub_frame.top &&
      dom_rect.bottom <= this._viewport_relative_sub_frame.bottom;
    const isTopIn =
      dom_rect.top >= this._viewport_relative_sub_frame.top &&
      dom_rect.top <= this._viewport_relative_sub_frame.bottom;
    const isLeftIn =
      dom_rect.left >= this._viewport_relative_sub_frame.left &&
      dom_rect.left <= this._viewport_relative_sub_frame.right;
    const isRightIn =
      dom_rect.right >= this._viewport_relative_sub_frame.left &&
      dom_rect.right <= this._viewport_relative_sub_frame.right;
    return (isBottomIn || isTopIn) && (isLeftIn || isRightIn);
  }

  __positionTextNodes() {
    for (const node of this._text_nodes) {
      this._node = node;
      this._text = node.textContent;
      this._formatText();
      this._character_index = 0;
      this._positionSingleTextNode();
    }
  }

  _formatText() {
    this._normaliseWhitespace();
    this._fixJustifiedText();
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
  //     next frame. Thus why the loop is independent of the `positionTextNodes()` loop.
  _fixJustifiedText() {
    this._node.parentElement.style.textAlign = "left";
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
  _normaliseWhitespace() {
    // Unify all whitespace to a single space character
    this._text = this._text.replace(/[\t\n\r ]+/g, " ");
    if (this._isFirstParseInElement()) {
      // Remove whitespace at the beginning
      if (this._text.charAt(0) === " ") {
        this._text = this._text.substring(1, this._text.length);
      }
      // Remove whitespace at the end
      if (this._text.charAt(this._text.length - 1) === " ") {
        this._text = this._text.substring(0, this._text.length - 1);
      }
    }
  }

  // Knowing if a text node is the first within its parent element helps to decide
  // whether to remove its leading whitespace or not.
  //
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
  _isFirstParseInElement() {
    let element = this._node.parentElement;
    const is_parse_started = _.includes(this._parse_started_elements, element);
    if (is_parse_started) {
      return false;
    } else {
      this._parse_started_elements.push(element);
      return true;
    }
  }

  // Here is where we actually make use of the rather strict monospaced and fixed font size
  // CSS rules enforced by the webextension. Of course the CSS is never going to be able to
  // perfectly snap characters onto a grid, so we force it here instead. At least we can be
  // fairly certain that every character at least takes up the same space as a TTY cell, it
  // just might not be perfectly aligned. So here we just round down all coordinates to force
  // the snapping.
  //
  // Use `this.addClientRectsOverlay(dom_rects, text);` to see DOM rectangle outlines in a
  // real browser.
  _positionSingleTextNode() {
    this._dom_box = {};
    for (const dom_box of this._getNodeDOMBoxes()) {
      if (!this._isDOMRectInSubFrame(dom_box)) {
        continue;
      }
      this._dom_box.top = dom_box.top;
      this._dom_box.left = dom_box.left;
      this._dom_box.width = dom_box.width;
      this._handleSingleDOMBox();
      this._previous_dom_box = _.clone(this._dom_box);
    }
  }

  // This is the key to being able to display formatted text within the strict confines
  // of a TTY. DOM Rectangles are closely related to selection ranges (like when you click
  // and drag the mouse cursor over text). Think of an individual DOM rectangle as a single
  // bar of highlighted selection. So that, for example, a 3 line paragraph will have 3
  // DOM rectangles. Fortunately DOMRect coordinates and dimensions are precisely defined.
  // Although do note that, unlike selection ranges, sub-selections can appear seemingly
  // inside other selections for things like italics or anchor tags.
  _getNodeDOMBoxes() {
    this._range.selectNode(this._node);
    return this._range.getClientRects();
  }

  // A single box is always a valid rectangle. Therefore a single box will, for example,
  // never straddle 2 lines as there is no guarantee that a valid rectangle can be formed.
  // We can use this to our advantage by stepping through coordinates of a box to get the
  // exact position of every single individual character. We just have to understand and
  // follow exactly how the DOM flows text - easier said than done.
  _handleSingleDOMBox() {
    this._prepareToParseDOMBox();
    for (let step = 0; step < this._tty_box.width; step++) {
      this._handleSingleCharacter();
      this._stepToNextCharacter();
    }
  }

  _prepareToParseDOMBox() {
    this._dom_box = this._convertDOMRectToAbsoluteCoords(this._dom_box);
    this._createSyncedTTYBox();
    this._createTrackers();
    this._setCurrentCharacter();
    this._ignoreUnrenderedWhitespace();
  }

  // Note that it's possible for this._text to straddle many DOM boxes
  _setCurrentCharacter() {
    this._current_character = this._text.charAt(this._character_index);
  }

  // Everything hinges on these 2 trackers being in sync. The DOM tracker is defined by
  // actual pixel coordinates and we move horizontally, from left to right, each step
  // being the width of a single character. The TTY tracker moves in the same way except
  // each step is a new single cell within the TTY.
  _createTrackers() {
    this._dom_tracker = {
      x: this._dom_box.left,
      y: this._dom_box.top
    };
    this._tty_tracker = {
      x: this._tty_box.col_start,
      y: this._tty_box.row
    };
  }

  _handleSingleCharacter() {
    let cell = new TTYCell();
    cell.rune = this._current_character;
    cell.tty_coords = _.clone(this._tty_tracker);
    cell.dom_coords = _.clone(this._dom_tracker);
    cell.parent_element = this._node.parentElement;
    this.tty_grid.addCell(cell);
  }

  _stepToNextCharacter(tracked = true) {
    this._character_index++;
    this._setCurrentCharacter();
    if (tracked) {
      this._dom_tracker.x += this.dimensions.char.width;
      this._tty_tracker.x++;
    }
  }

  // There is a careful tracking between the currently parsed character of `this._text`
  // and the position of the current 'cell' space within `this._dom_box`. So we must be precise
  // in how we synchronise them. This requires following the DOM's method for wrapping text.
  // Recall how the DOM will split a line at a space character boundry. That space character
  // is then in fact never rendered - its existence is never registered within the dimensions
  // of a DOM rectangle's box (`this._dom_box`).
  _ignoreUnrenderedWhitespace() {
    if (this._isNewLine()) {
      if (this._current_character.trim().length == 0)
        this._stepToNextCharacter(false);
    }
  }

  // Is the current DOM rectangle further down the page than the previous?
  _isNewLine() {
    if (Object.keys(this._previous_dom_box).length === 0) return false;
    return this._dom_box.top > this._previous_dom_box.top;
  }

  // The DOM returns box coordinates relative to the viewport. As we are rendering the
  // entire DOM as a single frame, then we need the coords to be relative to the top-left
  // of the DOM itself.
  _convertDOMRectToAbsoluteCoords(dom_rect) {
    return {
      top: dom_rect.top + window.scrollY,
      bottom: dom_rect.bottom + window.scrollY,
      left: dom_rect.left + window.scrollX,
      right: dom_rect.right + window.scrollX,
      height: dom_rect.height,
      width: dom_rect.width
    };
  }

  // Round and snap a DOM rectangle as if it were placed in the TTY frame
  _createSyncedTTYBox() {
    this._tty_box = {
      col_start: utils.snap(
        this._dom_box.left * this.dimensions.scale_factor.width
      ),
      row: utils.snap(
        (this._dom_box.top * this.dimensions.scale_factor.height) / 2
      ),
      width: utils.snap(
        this._dom_box.width * this.dimensions.scale_factor.width
      )
    };
  }

  // Purely for debugging.
  //
  // Draws a red border around all the DOMClientRect nodes.
  // Based on code from the MDN docs site.
  _addClientRectsOverlay(dom_rects, normalised_text) {
    // Don't draw on every frame
    if (this.is_first_frame_finished) return;
    // Absolutely position a div over each client rect so that its border width
    // is the same as the rectangle's width.
    // Note: the overlays will be out of place if the user resizes or zooms.
    for (const rect of dom_rects) {
      let tableRectDiv = document.createElement("div");
      // A DOMClientRect object only contains dimensions, so there's no way to identify it
      // to a node, so let's put its text as an attribute so we can cross-check if needs be.
      tableRectDiv.setAttribute("browsh-text", normalised_text);
      let tty_row = parseInt(
        Math.round(rect.top / this.dimemnsions.char.height)
      );
      tableRectDiv.setAttribute("tty_row", tty_row);
      tableRectDiv.style.position = "absolute";
      tableRectDiv.style.border = "1px solid red";
      tableRectDiv.style.margin = tableRectDiv.style.padding = "0";
      tableRectDiv.style.top = rect.top + "px";
      tableRectDiv.style.left = rect.left + "px";
      // We want rect.width to be the border width, so content width is 2px less.
      tableRectDiv.style.width = rect.width - 2 + "px";
      tableRectDiv.style.height = rect.height - 2 + "px";
      document.body.appendChild(tableRectDiv);
    }
  }
}
