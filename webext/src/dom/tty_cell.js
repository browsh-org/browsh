import charWidthInTTY from 'string-width';

// A single cell on the TTY grid
export default class {
  // When a character clobbers another character in the grid, we can't use our
  // text show/hide trick to know if the character is visible in the final DOM. So we have
  // to use standard CSS inspection instead. Hopefully this doesn't happen often because
  // it's expensive.
  // TODO: Make comprehensive
  isHighestLayer(challenger_cell) {
    const tenant_styles = this.getStyles();
    const challenger_styles = challenger_cell.getStyles();
    if (
      challenger_styles.visibility === 'hidden' ||
      challenger_styles.display === 'none'
    ) {
      return false;
    }
    return tenant_styles.zIndex > challenger_styles.zIndex;
  }

  // Get or cache the total cascaded calculated styles for an element
  getStyles() {
    let element = this.parent_element;
    if (!element.browsh_calculated_styles) {
      let styles = window.getComputedStyle(element);
      element.browsh_calculated_styles = styles;
    }
    return element.browsh_calculated_styles;
  }

  isTransparent() {
    const is_undefined = this.rune === undefined;
    const is_empty = this.rune === '';
    const is_space = /^\s+$/.test(this.rune);
    const is_not_worth_printing = is_empty || is_space || is_undefined;
    return is_not_worth_printing;
  }

  // Deal with UTF8 characters that take up more than a single cell in the TTY.
  // Eg; 比如说
  // TODO:
  //   1. Do all terminals deal with wide characters the same?
  //   2. Use CSS or JS so that wide characters actually flow in the DOM as 2
  //      monospaced characters. This will allow pages of nothing but wide
  //      characters to render/flow as closely as possible ot how they will appear
  //      in the TTY.
  calculateCharWidthPadding() {
    return charWidthInTTY(this.rune) - 1;
  }
}

