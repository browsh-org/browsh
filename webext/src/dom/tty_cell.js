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
}

