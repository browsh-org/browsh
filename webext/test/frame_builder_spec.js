import sandbox from 'helper';
import {expect} from 'chai';

import FrameBuilder from 'frame_builder';
import GraphicsBuilder from 'graphics_builder';
import TextBuilder from 'text_builder';
import canvas_pixels from 'fixtures/canvas_pixels';
import text_grid from 'fixtures/text_grid';

describe('Frame Builder', ()=> {
  let frame_builder;

  beforeEach(()=> {
    sandbox.stub(GraphicsBuilder.prototype, '_getPixelData').returns(canvas_pixels);
    sandbox.stub(TextBuilder.prototype, 'getFormattedText').returns(text_grid);
    frame_builder = new FrameBuilder();
  });

  it('should merge pixels and text into ANSI true colour syntax', ()=> {
    frame_builder.sendFrame(3, 2);
    const frame = frame_builder.frame.replace(/\u001b\[/g, 'ESC');
    expect(frame).to.eq(
      'ESC38;2;0;0;0mESC48;2;111;111;111m▄' +
      'ESC38;2;111;111;111mESC48;2;222;222;222m😐' +
      'ESC38;2;0;0;0mESC48;2;111;111;111m▄\n' +
      'ESC38;2;111;111;111mESC48;2;222;222;222m😄' +
      'ESC38;2;111;111;111mESC48;2;0;0;0m▄' +
      'ESC38;2;111;111;111mESC48;2;222;222;222m😂'
    );
  });
});
