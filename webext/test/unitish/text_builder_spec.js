import sandbox from 'unitish/helper';
import {
  expect
} from 'chai';

import FrameBuilder from 'dom/frame_builder';
import TextBuilder from 'dom/text_builder';
import GraphicsBuilder from 'dom/graphics_builder';
import text_nodes from 'unitish/fixtures/text_nodes';
import {
  with_text,
  without_text,
  scaled
} from 'unitish/fixtures/canvas_pixels';

let text_builder;

// To save us hand-writing large pixel arrays, let's just have an unrealistically
// small window, it's not a problem, because we'll never actually have to view real
// webpages on it.
window.innerWidth = 3;
window.innerHeight = 4;

function setup() {
  let frame_builder = new FrameBuilder();
  frame_builder.tty_width = 3
  frame_builder.tty_height = 2 + 2
  frame_builder.char_width = 1
  frame_builder.char_height = 2
  frame_builder.graphics_builder.getSnapshotWithText();
  frame_builder.graphics_builder.getSnapshotWithoutText();
  frame_builder.graphics_builder.getScaledSnapshot();
  text_builder = new TextBuilder(frame_builder);
}

describe('Text Builder', () => {
  beforeEach(() => {
    let getPixelsStub = sandbox.stub(GraphicsBuilder.prototype, '_getPixelData');
    getPixelsStub.onCall(0).returns(with_text);
    getPixelsStub.onCall(1).returns(without_text);
    getPixelsStub.onCall(2).returns(scaled);
    setup();
    text_builder.text_nodes = text_nodes;
  });

  it('should convert text nodes to a grid', () => {
    text_builder._updateState();
    text_builder._positionTextNodes();
    const grid = text_builder.tty_grid;
    expect(grid[0]).to.deep.equal([
      't', [255, 255, 255],
      [0, 0, 0],
      {
        "style": {
          "textAlign": "left"
        }
      }, {
        "x": 0,
        "y": 0
      }
    ]);
    expect(grid[1]).to.deep.equal([
      'e', [255, 255, 255],
      [111, 111, 111], {
        "style": {
          "textAlign": "left"
        }
      }, {
        "x": 1,
        "y": 0
      }
    ]);
    expect(grid[2]).to.deep.equal([
      's', [255, 255, 255],
      [0, 0, 0], {
        "style": {
          "textAlign": "left"
        }
      }, {
        "x": 2,
        "y": 0
      }
    ]);
    expect(grid[3]).to.be.undefined;
    expect(grid[4]).to.be.undefined;
    expect(grid[5]).to.be.undefined;
  });
});
