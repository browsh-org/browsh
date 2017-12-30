import MockRange from 'mocks/range'

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

global.BUILD_ENV = {};
global.window = {};
global.performance = {
  now: () => {}
}
