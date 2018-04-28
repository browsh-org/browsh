import sandbox from 'helper';
import { expect } from 'chai';

import Dimensions from 'dom/dimensions';
import GraphicsBuilder from 'dom/graphics_builder';
import { scaled } from 'fixtures/canvas_pixels';

let graphics_builder;

function setup() {
  let dimensions = new Dimensions();
  graphics_builder = new GraphicsBuilder({name: "1"}, dimensions);
  graphics_builder.getScaledScreenshot();
}

describe('Graphics Builder', () => {
  beforeEach(() => {
    let getPixelsStub = sandbox.stub(GraphicsBuilder.prototype, '_getPixelData');
    getPixelsStub.onCall(0).returns(scaled);
    setup();
  });

  it('should serialise a scaled frame', () => {
    graphics_builder._serialiseFrame();
    const colours = graphics_builder.frame.colours
    expect(colours.length).to.equal(36);
    expect(colours[0]).to.equal(111);
    expect(colours[3]).to.equal(0);
    expect(colours[32]).to.equal(111);
    expect(colours[35]).to.equal(0);
  });
});
