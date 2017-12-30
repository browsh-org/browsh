import _ from 'lodash';

import BaseBuilder from 'base_builder';

// Convert the text on the page into a snapped 2-dimensional grid to be displayed directly
// in the terminal.
export default class TextBuillder extends BaseBuilder {
  constructor() {
    super();
    this.tty_grid = [];
    this.parse_started_elements = [];
  }

  getFormattedText(frame_builder, graphics_builder) {
    this.tty_width = frame_builder.tty_width;
    this.tty_height = frame_builder.ty_height;
    this.char_width = frame_builder.char_width;
    this.char_height = frame_builder.char_height;
    this.pixels_with_text = frame_builder.pixels_with_text;
    this.pixels_without_text = frame_builder.pixels_without_text;
    this.logPerformance(() => {
      // This is relatively cheap: around 50ms for a 13,000 word Wikipedia page.
      this.getTextNodes();
    }, 'Tree Walker');
    this.setViewportDimensions(graphics_builder);
    this.logPerformance(() => {
      // This should be around 125ms for a largish Wikipedia page of 13,000 words.
      this.positionTextNodes();
    }, 'position text nodes');
    this.is_first_frame_finished = true;
    return this.tty_grid;
  }

  setViewportDimensions(graphics_builder) {
    this.viewport = {
      width: graphics_builder.viewport_width,
      height: graphics_builder.viewport_height
    }
  }

  // Search through every node in the DOM looking for displayable text.
  getTextNodes() {
    this.text_nodes = [];
    const walker = document.createTreeWalker(
      document.body,
      NodeFilter.SHOW_TEXT,
      { acceptNode: this.isRelevantTextNode },
      false
    );
    while(walker.nextNode()) this.text_nodes.push(walker.currentNode);
  }

  // Does the node contain text that we want to display?
  // TODO: Exclude text outside of viewport
  isRelevantTextNode(node) {
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
  positionTextNodes() {
    let range = document.createRange();
    let bounding_box;
    for (const node of this.text_nodes) {
      range.selectNode(node);
      bounding_box = range.getBoundingClientRect();
      if (this.isOutsideViewport(bounding_box)) continue;
      this.fixJustifiedText(node);
      this.formatTextForTTYGrid(
        this.normaliseWhitespace(node.textContent, node.parentElement),
        range.getClientRects()
      );
    }
  }

  // If even a single pixel is inside the viewport we need to check it
  isOutsideViewport(bounding_box) {
    const is_top_in =
      bounding_box.top >= 0 &&
      bounding_box.top < this.viewport.height;
    const is_bottom_in =
      bounding_box.bottom >= 0 &&
      bounding_box.bottom < this.viewport.height;
    const is_left_in =
      bounding_box.left >= 0 &&
      bounding_box.left < this.viewport.width;
    const is_right_in =
      bounding_box.right >= 0 &&
      bounding_box.right < this.viewport.width;
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
  fixJustifiedText(node) {
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
  normaliseWhitespace(text, parent) {
    text = text.replace(/[\t\n\r ]+/g, " ");
    if (this.isFirstParseInElement(parent)) {
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
  isFirstParseInElement(element) {
    const is_parse_started = _.includes(this.parse_started_elements, element);
    if (is_parse_started) {
      return false
    } else {
      this.parse_started_elements.push(element);
      return true
    }
  }

  // Here is where we actually make use of the rather strict monospaced and fixed font size
  // CSS rules enforced by the webextension. Of course the CSS is never going to be able to
  // perfectly snap characters onto a grid, so we force it here instead. At least we can be
  // fairly certain that every character at least takes up the same space as a TTY cell, it
  // just might not be perfectly aligned. So here we just round down all coordinates to force
  // the snapping.
  formatTextForTTYGrid(text, dom_rects) {
    let col, tty_box, step, character, previous_box;
    let character_index = 0;
    for (const box of dom_rects) {
      if (this.isNewLine(previous_box, box)) {
        character = text.charAt(character_index);
        if (/[\t\n\r ]+/.test(character)) character_index++;
      }
      tty_box = this.convertBoxToTTYUnits(box);
      col = tty_box.col_start;
      for (step = 0; step < tty_box.width; step++) {
        character = text.charAt(character_index);
        this.placeCharacterOnTTYGrid(
          col,
          tty_box.row,
          character
        );
        col++;
        character_index++;
      }
      previous_box = box;
    }
  }

  isNewLine(previous_box, current_box) {
    if (previous_box === undefined) return false;
    return current_box.top > previous_box.top
  }

  convertBoxToTTYUnits(viewport_dom_rect) {
    return {
      col_start: Math.round(viewport_dom_rect.left / this.char_width),
      row: Math.round(viewport_dom_rect.top / this.char_height),
      width: Math.round(viewport_dom_rect.width / this.char_width),
    }
  }

  placeCharacterOnTTYGrid(col, row, character) {
    //let pixel_data_start;
    //let fg = [];
    //let bg = [];
    const index = (row * this.tty_width) + col;
    if (this.isCharOutsideGrid(col, row)) return;
    //if (this.isTextObscured(index)) return;
    // Don't clobber for now. TODO: Use `getComputedStyles()` and save for whole element.
    if (this.tty_grid[index] === undefined) {
      //pixel_data_start = ((row * 2) + 1) * (this.tty_width * 4) + (col * 4);
      //fg[0] = this.pixels_with_text[pixel_data_start + 0];
      //fg[1] = this.pixels_with_text[pixel_data_start + 1];
      //fg[2] = this.pixels_with_text[pixel_data_start + 2];
      //bg[0] = this.pixels_without_text[pixel_data_start + 0];
      //bg[1] = this.pixels_without_text[pixel_data_start + 1];
      //bg[2] = this.pixels_without_text[pixel_data_start + 2];
      this.tty_grid[index] = [
        character,
        [0,0,0],
        [255,255,255]
      ];
    }
  }

  // Theoretically this should only be needed for DOM rectangles that _straddle_ the
  // viewport.
  isCharOutsideGrid(col, row) {
    return col + 1 >= this.tty_width || row + 1 >= this.tty_height;
  }

  // This is somewhat of a, hopefully elegant, hack. So, imagine that situation where you're
  // browsing a web page and a popup appears; perhaps just a select box, or menu, or worst
  // of all a dreaded full-page overlay. Now, DOM rectangles don't take into account whether
  // they are the uppermost visible element, so we're left in a bit of a pickle. The only JS
  // way to know if an element is visible is to use `Document.elementFromPoint(x, y)`, where
  // you compare the returned element with the element whose visibility you're checking.
  // This is has a number of problems. Firstly, it only checks one coordinate in the element
  // for visibility, which of course isn't going to be 100% reliably speak for all the
  // characters in the element. Secondly, even ignoring the first caveat, running
  // `elementFromPoint()` for every character is very expensive, around 25ms for an average
  // DOM. So it's basically a no-go. So instead we take advantage of the fact that we're
  // working with a very scaled down version of the webpage's pixels. In fact not just any
  // scale, but the scale such that every pixel represents a single character. As such it's
  // fairly safe to assume that if you make the text transparent and a pixel's colour
  // doesn't change then that character must be obscured by something.
  // There are of course some potential edge cases with this. What if, as a result of a
  // significant scaling a '.', or a ',' becomes so small that it doesn't actually register
  // any colour change whatsoever? Or what if we get a false positive, where a character is
  // obscured _by another character_? Hopefully in such a case we can work with `z-index`
  // so that characters justifiably overwrite each other in the TTY grid.
  // TODO: compare top and bottom pixel of text row.
  isTextObscured(index) {
    return _.isEqual(
      this.pixels_with_text[index],
      this.pixels_without_text[index]
    );
  }

  // Purely for debugging. Draws a red border around all the DOMClientRect nodes.
  // Based on code from the MDN docs site.
  addClientRectsOverlay(dom_rects, normalised_text) {
    // Don't draw on every frame
    if (this.first_call_finished) return;
    // Absolutely position a div over each client rect so that its border width
    // is the same as the rectangle's width.
    // Note: the overlays will be out of place if the user resizes or zooms.
    for (const rect of dom_rects) {
      let tableRectDiv = document.createElement('div');
      // A DOMClientRect object only contains dimensions, so there's no way to identify it
      // to a node, so let's put its text as an attribute so we can cross-check if needs be.
      tableRectDiv.setAttribute('browsh-text', normalised_text);
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
