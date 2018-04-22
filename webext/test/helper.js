import sinon from 'sinon';

import Dimensions from 'dom/dimensions';
import GraphicsBuilder from 'dom/graphics_builder';
import MockRange from 'mocks/range'

var sandbox = sinon.sandbox.create();

beforeEach(() => {
  sandbox.stub(Dimensions.prototype, '_getOrCreateMeasuringBox').returns(element);
  sandbox.stub(Dimensions.prototype, 'sendMessage').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_hideText').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_showText').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_scaleCanvas').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_unScaleCanvas').returns(true);
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
    scrollHeight: 4
  },
  location: {
    href: 'https://www.google.com'
  },
  scrollX: 0,
  scrollY: 0,

  // To save us hand-writing large pixel arrays, let's just have an unrealistically
  // small window, it's not a problem, because we'll never actually have to view real
  // webpages on it.
  innerWidth: 3,
  innerHeight: 4
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
