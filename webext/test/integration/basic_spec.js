import helper from 'integration/helper'
import {expect} from 'chai';

describe('Basic', function () {
  this.retries(3);
  it('basic', (done) => {
    helper.getPage('http://www.something.com/', (page) => {
      expect(page.title).to.eq('Something.');
      expect(page.url).to.eq('http://www.something.com/');
      done();
    });
  });
});
