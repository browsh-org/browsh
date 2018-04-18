import sandbox from 'helper';
import { expect } from 'chai';

import Dimensions from 'dom/dimensions';
import FrameBuilder from 'dom/frame_builder';
import GraphicsBuilder from 'dom/graphics_builder';
import TTYCell from 'dom/tty_cell';
import text_nodes from 'fixtures/text_nodes';
import {
  with_text,
  without_text,
  scaled
} from 'fixtures/canvas_pixels';


// To save us hand-writing large pixel arrays, let's just have an unrealistically
// small window, it's not a problem, because we'll never actually have to view real
// webpages on it.
window.innerWidth = 3;
window.innerHeight = 4;

let frame_builder;

function setup() {
  frame_builder = new FrameBuilder(undefined, new Dimensions());
  frame_builder.graphics_builder.getScreenshotWithText();
  frame_builder.graphics_builder.getScreenshotWithoutText();
  frame_builder.graphics_builder.getScaledScreenshot();
  sandbox.stub(TTYCell.prototype, 'isHighestLayer').returns(true);
}

describe('Text Builder', () => {
  beforeEach(() => {
    let getPixelsStub = sandbox.stub(GraphicsBuilder.prototype, '_getPixelData');
    getPixelsStub.onCall(0).returns(with_text);
    getPixelsStub.onCall(1).returns(without_text);
    getPixelsStub.onCall(2).returns(scaled);
    setup();
    frame_builder.text_builder._text_nodes = text_nodes;
  });

  it('should convert text nodes to a grid', () => {
    frame_builder.text_builder._updateState();
    frame_builder.text_builder._positionTextNodes();
    const grid = frame_builder.text_builder.tty_grid.cells;
    expect(grid[0]).to.deep.equal({
      index: 0,
      rune: 't',
      fg_colour: [255, 255, 255],
      bg_colour: [0, 0, 0],
      parent_element: {
        style: {}
      },
      tty_coords: {
        x: 0,
        y: 0
      },
      dom_coords: {
        x: 0,
        y: 0
      }
    });
    expect(grid[1]).to.deep.equal({
      index: 1,
      rune: 'e',
      fg_colour: [255, 255, 255],
      bg_colour: [111, 111, 111],
      parent_element: {
        style: {}
      },
      tty_coords: {
        x: 1,
        y: 0
      },
      dom_coords: {
        x: 1,
        y: 0
      }
    });
    expect(grid[2]).to.deep.equal({
      index: 2,
      rune: 's',
      fg_colour: [255, 255, 255],
      bg_colour: [0, 0, 0],
      parent_element: {
        style: {}
      },
      tty_coords: {
        x: 2,
        y: 0
      },
      dom_coords: {
        x: 2,
        y: 0
      }
    });
  });
});
