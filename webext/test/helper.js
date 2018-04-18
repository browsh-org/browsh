import sinon from 'sinon';

import Dimensions from 'dom/dimensions';
import FrameBuilder from 'dom/frame_builder';
import GraphicsBuilder from 'dom/graphics_builder';
import MockRange from 'mocks/range'

var sandbox = sinon.sandbox.create();

beforeEach(() => {
  sandbox.stub(Dimensions.prototype, '_getOrCreateMeasuringBox').returns(element);
  sandbox.stub(GraphicsBuilder.prototype, '_hideText').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_showText').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_scaleCanvas').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_unScaleCanvas').returns(true);
  sandbox.stub(FrameBuilder.prototype, 'sendMessage').returns(true);
});

afterEach(() => {
  sandbox.restore();
});

global.dimensions = {
  char: {
    width: 1,
    height: 2 - 2
  }
}

global.document = {
  addEventListener: () => {},
  getElementById: () => {},
  getElementsByTagName: () => {
    return [{
      innerHTML: 'Google'
    }]
  },
  createRange: () => {
    return new MockRange()
  },
  createElement: () => {
    return {
      getContext: () => {}
    }
  },
  documentElement: {
    scrollWidth: 3,
    scrollHeight: 8
  },
  location: {
    href: 'https://www.google.com'
  },
  scrollX: 0,
  scrollY: 0
};

global.DEVELOPMENT = false;
global.PRODUCTION = false;
global.TEST = true;
global.window = global.document;
global.performance = {
  now: () => {}
}

let element = {
  getBoundingClientRect: () => {
    return {
      width: global.dimensions.char.width,
      height: global.dimensions.char.height
    }
  }
}

export default sandbox;
