import 'helper';
import _ from 'lodash';
import {expect} from 'chai';

import TextBuilder from 'text_builder';
import text_nodes from 'fixtures/text_nodes';

const defaults = {
  tty_width: 10,
  tty_height: 5,
  char_width: 10,
  char_height: 20,
  viewport: {
    height: 100,
    width: 100
  },
  text_nodes: text_nodes
}

describe('Text Builder', ()=> {
  let text_builder;

  beforeEach(()=> {
    text_builder = new TextBuilder();
    _.merge(text_builder, defaults);
  });

  it('should convert text nodes to a grid', ()=> {
    text_builder.positionTextNodes();
    expect(text_builder.tty_grid[0]).to.be.undefined;
    expect(text_builder.tty_grid[11]).to.deep.equal(['t', [0, 0, 0], [255, 255, 255]]);
    expect(text_builder.tty_grid[17][0]).to.eq('g');
    expect(text_builder.tty_grid[21][0]).to.eq('n');
    expect(text_builder.tty_grid[25][0]).to.eq('s');
  });
});
