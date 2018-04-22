import sandbox from 'helper';
import { expect } from 'chai';

import Dimensions from 'dom/dimensions';
import GraphicsBuilder from 'dom/graphics_builder';
import { scaled } from 'fixtures/canvas_pixels';

let graphics_builder;

function setup() {
  let dimensions = new Dimensions();
  graphics_builder = new GraphicsBuilder(undefined, dimensions);
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
    expect(graphics_builder.frame.length).to.equal(36);
    expect(graphics_builder.frame[0]).to.equal('111');
    expect(graphics_builder.frame[3]).to.equal('0');
    expect(graphics_builder.frame[32]).to.equal('111');
    expect(graphics_builder.frame[35]).to.equal('0');
  });
});
