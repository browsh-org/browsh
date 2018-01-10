import sinon from 'sinon';

import GraphicsBuilder from 'dom/graphics_builder';
import FrameBuilder from 'dom/frame_builder';
import MockRange from 'mocks/range'

var sandbox = sinon.sandbox.create();

beforeEach(() => {
  sandbox.stub(GraphicsBuilder.prototype, '_hideText').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_showText').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_scaleCanvas').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_unScaleCanvas').returns(true);
  sandbox.stub(FrameBuilder.prototype, '_sendMessage').returns(true);
});

afterEach(() => {
  sandbox.restore();
});

global.document = {
  addEventListener: () => {},
  getElementById: () => {},
  createRange: () => {
    return new MockRange()
  },
  createElement: () => {
    return {
      getContext: () => {}
    }
  }
};

global.DEVELOPMENT = false;
global.PRODUCTION = false;
global.TEST = true;
global.window = {};
global.performance = {
  now: () => {}
}

export default sandbox;
