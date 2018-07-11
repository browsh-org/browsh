// A single cell on the TTY grid
export default class {
  // When a character clobbers another character in the grid, we can't use our
  // text show/hide trick to know if the character is visible in the final DOM. So we have
  // to use standard CSS inspection instead. Hopefully this doesn't happen often because
  // it's expensive.
  // TODO: Make comprehensive
  isHighestLayer() {
    const found_element = document.elementFromPoint(
      this.dom_coords.x,
      this.dom_coords.y
    );
    return this.parent_element == found_element;
  }
}
