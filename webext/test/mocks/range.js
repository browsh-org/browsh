export default class MockRange {
  selectNode(node) {
    this.node = node;
  }
  getBoundingClientRect() {
    return this.node.bounding_box;
  }
  getClientRects() {
    return this.node.dom_rects;
  }
}
