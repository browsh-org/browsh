import sinon from 'sinon';

import Dimensions from 'dom/dimensions';
import GraphicsBuilder from 'dom/graphics_builder';
import TextBuilder from 'dom/text_builder';
import TTYCell from 'dom/tty_cell';

import MockRange from 'mocks/range'
import TextNodes from 'fixtures/text_nodes';
import CanvasPixels from 'fixtures/canvas_pixels';

var sandbox = sinon.sandbox.create();
let getPixelsStub;
let channel = {name: 1};

beforeEach(() => {
  sandbox.stub(Dimensions.prototype, '_getOrCreateMeasuringBox').returns(element);
  sandbox.stub(Dimensions.prototype, 'sendMessage').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_hideText').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_showText').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_scaleCanvas').returns(true);
  sandbox.stub(GraphicsBuilder.prototype, '_unScaleCanvas').returns(true);
  sandbox.stub(TextBuilder.prototype, '_getAllInputBoxes').returns([]);
  sandbox.stub(TTYCell.prototype, 'isHighestLayer').returns(true);
  getPixelsStub = sandbox.stub(GraphicsBuilder.prototype, '_getPixelData');
});

afterEach(() => {
  sandbox.restore();
});

global.dimensions = {
  char: {
    width: 1,
    height: 2
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
    scrollWidth: null,
    scrollHeight: null
  },
  location: {
    href: 'https://www.google.com'
  },
  scrollX: 0,
  scrollY: 0,

  innerWidth: null,
  innerHeight: null
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

function _setupMockDOMSize() {
  const width = global.mock_DOM_template[0].length;
  const height = global.mock_DOM_template.length * 2;
  global.document.documentElement.scrollWidth = width;
  global.document.documentElement.scrollHeight = height;
  global.document.innerWidth = width;
  global.document.innerHeight = height;
}

function _setupDimensions() {
  let dimensions = new Dimensions();
  _setupMockDOMSize();
  dimensions.tty.width = global.tty.width;
  dimensions.tty.height = global.tty.height;
  dimensions.frame.x_scroll = global.tty.x_scroll;
  dimensions.frame.y_scroll = global.tty.y_scroll;
  dimensions.update();
  dimensions.setSubFrameDimensions(global.frame_type);
  return dimensions;
}

function _setupGraphicsBuilder(type) {
  let dimensions = _setupDimensions()
  let canvas_pixels = new CanvasPixels(dimensions);
  if (type === 'with_text') {
    getPixelsStub.onCall(0).returns(canvas_pixels.with_text());
    getPixelsStub.onCall(1).returns(canvas_pixels.without_text());
    getPixelsStub.onCall(2).returns(canvas_pixels.scaled());
  } else {
    getPixelsStub.onCall(0).returns(canvas_pixels.scaled());
  }
  let graphics_builder = new GraphicsBuilder(channel, dimensions);
  return graphics_builder;
}

let functions = {
  runTextBuilder: () => {
    let text_nodes = new TextNodes();
    let graphics_builder = _setupGraphicsBuilder('with_text');
    let text_builder = new TextBuilder(
      channel,
      graphics_builder.dimensions,
      graphics_builder
    );
    graphics_builder._getScreenshotWithText();
    graphics_builder._getScreenshotWithoutText();
    graphics_builder.__getScaledScreenshot();
    text_builder._text_nodes = text_nodes.build();
    text_builder._updateState();
    text_builder._positionTextNodes();
    return text_builder;
  },

  runGraphicsBuilder: () => {
    let graphics_builder = _setupGraphicsBuilder();
    graphics_builder.__getScaledScreenshot();
    graphics_builder._serialiseFrame();
    return graphics_builder;
  }
}

export default functions;
