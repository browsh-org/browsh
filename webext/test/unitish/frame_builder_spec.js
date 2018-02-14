import sandbox from 'unitish/helper';
import {expect} from 'chai';

import FrameBuilder from 'dom/frame_builder';
import GraphicsBuilder from 'dom/graphics_builder';
import TextBuilder from 'dom/text_builder';
import canvas_pixels from 'unitish/fixtures/canvas_pixels';
import text_grid from 'unitish/fixtures/text_grid';

describe('Frame Builder', ()=> {
  let frame_builder;

  beforeEach(()=> {
    sandbox.stub(GraphicsBuilder.prototype, '_getPixelData').returns(canvas_pixels);
    sandbox.stub(TextBuilder.prototype, 'getFormattedText').returns(text_grid);
    frame_builder = new FrameBuilder();
  });

  it('should merge pixels and text into a 1D array', ()=> {
    frame_builder.tty_width = 3;
    frame_builder.tty_height = 2 + 2;
    frame_builder.sendFrame();
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
