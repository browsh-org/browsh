import sandbox from 'helper';
import { expect } from 'chai';

import DocumentBuilder from 'dom/document_builder';
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

let document_builder;

function setup() {
  document_builder = new DocumentBuilder();
  document_builder.tty_width = 3
  document_builder.tty_height = 2 + 2
  document_builder.char_width = 1
  document_builder.char_height = 2
  document_builder.getScreenshotWithText();
  document_builder.getScreenshotWithoutText();
  document_builder.getScaledScreenshot();
}

describe('Text Builder', () => {
  beforeEach(() => {
    let getPixelsStub = sandbox.stub(DocumentBuilder.prototype, '_getPixelData');
    getPixelsStub.onCall(0).returns(with_text);
    getPixelsStub.onCall(1).returns(without_text);
    getPixelsStub.onCall(2).returns(scaled);
    setup();
    document_builder._text_nodes = text_nodes;
  });

  it('should convert text nodes to a grid', () => {
    document_builder._updateState();
    document_builder._positionTextNodes();
    const grid = document_builder.tty_grid;
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
