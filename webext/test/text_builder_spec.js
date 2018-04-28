import sandbox from 'helper';
import { expect } from 'chai';

import Dimensions from 'dom/dimensions';
import GraphicsBuilder from 'dom/graphics_builder';
import TextBuilder from 'dom/text_builder';
import TTYCell from 'dom/tty_cell';
import text_nodes from 'fixtures/text_nodes';
import {
  with_text,
  without_text,
  scaled
} from 'fixtures/canvas_pixels';

let graphics_builder, text_builder;
let channel = {name: 1};

function setup() {
  let dimensions = new Dimensions();
  graphics_builder = new GraphicsBuilder(channel, dimensions);
  text_builder = new TextBuilder(channel, dimensions, graphics_builder);
  graphics_builder.getScreenshotWithText();
  graphics_builder.getScreenshotWithoutText();
  graphics_builder.getScaledScreenshot();
  sandbox.stub(TTYCell.prototype, 'isHighestLayer').returns(true);
}

describe('Text Builder', () => {
  beforeEach(() => {
    let getPixelsStub = sandbox.stub(GraphicsBuilder.prototype, '_getPixelData');
    getPixelsStub.onCall(0).returns(with_text);
    getPixelsStub.onCall(1).returns(without_text);
    getPixelsStub.onCall(2).returns(scaled);
    setup();
    text_builder._text_nodes = text_nodes;
    text_builder._updateState();
    text_builder._positionTextNodes();
  });

  it('should convert text nodes to a grid of cell objects', () => {
    const grid = text_builder.tty_grid.cells;
    expect(grid[0]).to.deep.equal({
      index: 0,
      rune: 't',
      fg_colour: [255, 255, 255],
      bg_colour: [0, 0, 0],
      parent_element: {
        style: {
          textAlign: "left"
        }
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
    expect(grid[5]).to.equal(undefined);
  });

  it('should ignore spaces on new lines', () => {
    const grid = text_builder.tty_grid.cells;
    expect(grid[3].rune).to.equal('n');
    expect(grid[3].tty_coords.y).to.equal(1);
  });

  it('should serialise a frame', () => {
    text_builder._serialiseFrame();
    expect(text_builder.frame.colours).to.deep.equal([
      255, 255, 255,
      255, 255, 255,
      255, 255, 255,
      255, 255, 255,
      0, 0, 0, 0, 0, 0
    ]);
    expect(text_builder.frame.text).to.deep.equal([
      "t", "e", "s", "n", "", ""
    ]);
  });
});
