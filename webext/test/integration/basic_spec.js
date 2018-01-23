import helper from 'integration/helper'
import {expect} from 'chai';

describe('Basic', function () {
  this.retries(3);

  it('basic', (done) => {
    helper.getPage('https://www.google.com', (page) => {
      expect(page.title).to.eq('Google');
      expect(page.url).to.eq('https://www.google.com/');
      done();
    });
  });
});
