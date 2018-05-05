import helper from 'helper';
import { expect } from 'chai';

describe('Graphics Builder', () => {
  let graphics_builder;

  describe('Non-offsetted frames', () => {
    beforeEach(() => {
      global.mock_DOM_template = [
        "    ",
        "    "
      ];
      global.frame_type = 'small';
      global.tty = {
        width: 4,
        height: 2,
        x_scroll: 0,
        y_scroll: 0
      }
      graphics_builder = helper.runGraphicsBuilder();
    });

    it('should serialise a scaled frame', () => {
      const colours = graphics_builder.frame.colours
      expect(colours.length).to.equal(48);
      expect(colours[0]).to.equal(0);
      expect(colours[2]).to.equal(1);
      expect(colours[46]).to.equal(0);
      expect(colours[47]).to.equal(16);
    });

    it("should populate the frame's meta", () => {
      const meta = graphics_builder.frame.meta
      expect(meta).to.deep.equal({
        sub_left: 0,
        sub_top: 0,
        sub_width: 4,
        sub_height: 4,
        total_width: 4,
        total_height: 4,
        id: 1
      });
    });
  });

  describe('Offset frames', () => {
    beforeEach(() => {
      global.tty = {
        width: 2,
        height: 2,
        x_scroll: 2,
        y_scroll: 1
      }
      global.frame_type = 'small';
      global.mock_DOM_template = [
        "    ",
        "    ",
        "    ",
        "    "
      ];
      graphics_builder = helper.runGraphicsBuilder();
    });

    it('should serialise a scaled frame', () => {
      const colours = graphics_builder.frame.colours
      expect(colours.length).to.equal(24);
      expect(colours[0]).to.equal(0);
      expect(colours[2]).to.equal(1);
      expect(colours[22]).to.equal(0);
      expect(colours[23]).to.equal(8);
    });

    it("should populate the frame's meta", () => {
      const meta = graphics_builder.frame.meta
      expect(meta).to.deep.equal({
        sub_left: 2,
        sub_top: 1,
        sub_width: 2,
        sub_height: 4,
        total_width: 4,
        total_height: 8,
        id: 1
      });
    });
  });
});
