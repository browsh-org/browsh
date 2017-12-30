import 'helper';
import sinon from 'sinon';
import {expect} from 'chai';

import FrameBuilder from 'frame_builder';
import GraphicsBuilder from 'graphics_builder';
import TextBuilder from 'text_builder';
import canvas_pixels from 'fixtures/canvas_pixels';
import text_grid from 'fixtures/text_grid';

describe('Frame Builder', ()=> {
  let frame_builder;

  beforeEach(()=> {
    sinon.stub(GraphicsBuilder.prototype, 'getSnapshotWithText').returns(canvas_pixels);
    sinon.stub(GraphicsBuilder.prototype, 'getSnapshotWithoutText').returns(canvas_pixels);
    sinon.stub(TextBuilder.prototype, 'getFormattedText').returns(text_grid);
    frame_builder = new FrameBuilder();
    frame_builder.channel = {
      postMessage: () => {}
    }
  });

  it('should merge pixels and text into ANSI true colour syntax', ()=> {
    frame_builder.sendFrame(3, 2);
    const screen = frame_builder.screen.replace(/\u001b\[/g, 'ESC');
    expect(screen).to.eq(
      'ESC38;2;0;0;0mESC48;2;255;255;255m▄' +
      'ESC38;2;111;111;111mESC48;2;222;222;222mx' +
      'ESC38;2;0;0;0mESC48;2;255;255;255m▄\n' +
      'ESC38;2;111;111;111mESC48;2;222;222;222mx' +
      'ESC38;2;255;255;255mESC48;2;0;0;0m▄' +
      'ESC38;2;111;111;111mESC48;2;222;222;222mx'
    );
  });
});
