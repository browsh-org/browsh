// Create DOM-compatible DOM Rectangles from a simple array of strings
export default class TextNodes {
  constructor() {
    this.offset = 0.0;
    this.char_width = global.dimensions.char.width;
    this.char_height = global.dimensions.char.height;
    this.total_width = global.mock_DOM_template[0].length * this.char_width;
    this.total_height = global.mock_DOM_template.length * this.char_height;
    this.dom_rects = [];
  }

  build() {
    for (let line of global.mock_DOM_text) {
      this.addDomRect(line);
    }
    return [
      {
        textContent: global.mock_DOM_text.join(""),
        parentElement: {
          style: {},
        },
        bounding_box: this.boundingBox(),
        dom_rects: this.dom_rects,
      },
    ];
  }

  boundingBox() {
    return {
      top: this.offset,
      bottom: this.total_height + this.offset,
      left: this.offset,
      right: this.total_width + this.offset,
      width: this.total_width,
      height: this.total_height,
    };
  }

  addDomRect(line) {
    const width = line.length * this.char_width;
    const height = this.char_height;
    const top = this.dom_rects.length * this.char_height + this.offset;
    this.dom_rects.push({
      top: top,
      bottom: top + height,
      left: this.offset,
      right: width + this.offset,
      width: width,
      height: height,
    });
  }
}
