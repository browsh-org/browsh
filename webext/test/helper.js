import sinon from 'sinon';

import DocumentBuilder from 'dom/document_builder';
import MockRange from 'mocks/range'

var sandbox = sinon.sandbox.create();

beforeEach(() => {
  sandbox.stub(DocumentBuilder.prototype, '_hideText').returns(true);
  sandbox.stub(DocumentBuilder.prototype, '_showText').returns(true);
  sandbox.stub(DocumentBuilder.prototype, '_scaleCanvas').returns(true);
  sandbox.stub(DocumentBuilder.prototype, '_unScaleCanvas').returns(true);
  sandbox.stub(DocumentBuilder.prototype, 'sendMessage').returns(true);
});

afterEach(() => {
  sandbox.restore();
});

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
  location: {
    href: 'https://www.google.com'
  }
};

global.DEVELOPMENT = false;
global.PRODUCTION = false;
global.TEST = true;
global.window = global.document;
global.performance = {
  now: () => {}
}

export default sandbox;
