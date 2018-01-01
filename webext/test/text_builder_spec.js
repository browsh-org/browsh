import sandbox from 'helper';
import {expect} from 'chai';

import TextBuilder from 'text_builder';
import GraphicsBuilder from 'graphics_builder';
import text_nodes from 'fixtures/text_nodes';
import {with_text, without_text, scaled} from 'fixtures/canvas_pixels';

let text_builder;

// To save us hand-writing large pixel arrays, let's just have an unrealistically
// small window, it's not a problem, because we'll never actually have to view real
// webpages on it.
window.innerWidth = 3;
window.innerHeight = 4;

const frame = {
  tty_width: 3,
  tty_height: 2,
  char_width: 1,
  char_height: 2
}

function setup() {
  let graphics_builder = new GraphicsBuilder();
  graphics_builder.getSnapshotWithText();
  graphics_builder.getSnapshotWithoutText();
  graphics_builder.getScaledSnapshot(frame.width, frame.height);
  text_builder = new TextBuilder(frame, graphics_builder);
}

describe('Text Builder', ()=> {
  beforeEach(()=> {
    let getPixelsStub = sandbox.stub(GraphicsBuilder.prototype, '_getPixelData');
    getPixelsStub.onCall(0).returns(with_text);
    getPixelsStub.onCall(1).returns(without_text);
    getPixelsStub.onCall(2).returns(scaled);
    setup();
    text_builder.text_nodes = text_nodes;
  });

  it('should convert text nodes to a grid', ()=> {
    text_builder._positionTextNodes();
    expect(text_builder.tty_grid[0]).to.deep.equal(['t', [255, 255, 255], [0, 0, 0]]);
    expect(text_builder.tty_grid[1]).to.deep.equal(['e', [255, 255, 255], [111, 111, 111]]);
    expect(text_builder.tty_grid[2]).to.deep.equal(['s', [255, 255, 255], [0, 0, 0]]);
    expect(text_builder.tty_grid[3]).to.be.undefined;
    expect(text_builder.tty_grid[4]).to.be.undefined;
    expect(text_builder.tty_grid[5]).to.be.undefined;
  });
});
