import sandbox from 'helper';
import {expect} from 'chai';

import Dimensions from 'dom/dimensions';
import FrameBuilder from 'dom/frame_builder';
import GraphicsBuilder from 'dom/graphics_builder';
import TextBuilder from 'dom/text_builder';
import canvas_pixels from 'fixtures/canvas_pixels';
import cells from 'fixtures/cells';

describe('Frame Builder', ()=> {
  let frame_builder;

  beforeEach(()=> {
    global.document.documentElement.scrollHeight = 4;
    frame_builder = new FrameBuilder(undefined, new Dimensions());
    frame_builder.text_builder.tty_grid.cells = cells;
    sandbox.stub(GraphicsBuilder.prototype, '_getPixelData').returns(canvas_pixels);
    sandbox.stub(TextBuilder.prototype, 'buildFormattedText').returns();
    sandbox.stub(TextBuilder.prototype, 'fixJustifiedText').returns();
  });

  it('should merge pixels and text into a 1D array', () => {
    frame_builder.makeFrame();
    const frame = frame_builder.frame.join(',');
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
