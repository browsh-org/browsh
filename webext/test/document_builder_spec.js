import sandbox from 'helper';
import {expect} from 'chai';

import DocumentBuilder from 'dom/document_builder';
import canvas_pixels from 'fixtures/canvas_pixels';
import tty_grid from 'fixtures/tty_grid';

describe('Document Builder', ()=> {
  let document_builder;

  beforeEach(()=> {
    document_builder = new DocumentBuilder();
    document_builder.tty_grid = tty_grid;
    sandbox.stub(DocumentBuilder.prototype, '_getPixelData').returns(canvas_pixels);
    sandbox.stub(DocumentBuilder.prototype, 'buildFormattedText').returns();
  });

  it('should merge pixels and text into a 1D array', ()=> {
    document_builder.tty_width = 3;
    document_builder.tty_height = 2 + 2;
    document_builder.makeFrame();
    const frame = document_builder.frame.join(',');
    expect(frame).to.eq(
      '0,0,0,111,111,111,▄,' +
      '111,111,111,222,222,222,😐,' +
      '0,0,0,111,111,111,▄,' +
      '111,111,111,222,222,222,😄,' +
      '111,111,111,0,0,0,▄,' +
      '111,111,111,222,222,222,😂'
    );
  });
});
