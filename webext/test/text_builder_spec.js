import {
  expect
} from 'chai';
import helper from 'helper'

let text_builder, grid;

describe('Text Builder', () => {
  beforeEach(() => {
    global.mock_DOM_template = [
      '                ',
      '                ',
      '                ',
      '                ',
      '                ',
      '       !!!      ',
      '       !!!      ',
    ];

    // We can't simulate anything that uses groups of spaces, as TextBuilder collapses all spaces
    // to a single space in order to sync with how the DOM renders monospaced text.
    //
    // TODO: That being said, I can surely imagine that multiple spaces within a single DOM rect
    // would not be collapsed by the DOM, so maybe that's something to take into account for
    // the TextBuilder code?
    global.mock_DOM_text = [
      'Testing nodes. ',
      'Max 15 chars ',
      'wide. ',
      'Diff kinds of ',
      'Whitespace. ',
      'Also we need to ',
      'test subframes.'
    ]

    global.tty = {
      width: 5,
      height: 3,
      x_scroll: 0,
      y_scroll: 0
    }
    global.frame_type = 'small';
    text_builder = helper.runTextBuilder();
    grid = text_builder.tty_grid.cells;
  });

  it('should convert text nodes to a grid of cell objects', () => {
    expect(grid.length).to.equal(37);
    expect(grid[0]).to.deep.equal({
      index: 0,
      rune: 'T',
      fg_colour: [6, 0, 0],
      bg_colour: [0, 0, 6],
      parent_element: {
        style: {
          textAlign: "left"
        }
      },
      tty_coords: {
        x: 0,
        y: 0
      },
      dom_coords: {
        x: 0,
        y: 0
      }
    });
    expect(grid[5]).to.equal(undefined);
    expect(grid[16]).to.deep.equal({
      index: 16,
      rune: 'M',
      fg_colour: [16, 0, 0],
      bg_colour: [0, 0, 16],
      parent_element: {
        style: {
          textAlign: "left"
        }
      },
      tty_coords: {
        x: 0,
        y: 1
      },
      dom_coords: {
        x: 0,
        y: 2
      }
    });
    expect(grid[36]).to.deep.equal({
      index: 36,
      rune: '.',
      fg_colour: [30, 0, 0],
      bg_colour: [0, 0, 30],
      parent_element: {
        style: {
          textAlign: "left"
        }
      },
      tty_coords: {
        x: 4,
        y: 2
      },
      dom_coords: {
        x: 4,
        y: 4
      }
    });
    expect(grid[37]).to.equal(undefined);
  });

  it('should not detect the colour of whitespace characters', () => {
    expect(grid[19].rune).to.equal(' ');
    expect(grid[19].fg_colour).to.deep.equal([0, 19, 0]);
  });

  it('should serialise a frame', () => {
    text_builder._serialiseFrame();
    expect(text_builder.frame.meta).to.deep.equal({
      sub_left: 0,
      sub_top: 0,
      sub_width: 5,
      sub_height: 6,
      total_width: 16,
      total_height: 14,
      id: 1
    });
    expect(text_builder.frame.text).to.deep.equal([
      'T', 'e', 's', 't', 'i', 'M', 'a', 'x', ' ', '1', 'w', 'i', 'd', 'e', '.'
    ]);
    expect(text_builder.frame.colours).to.deep.equal([
      6, 0, 0, 7, 0, 0, 8, 0, 0, 9, 0, 0, 10, 0, 0,
      16, 0, 0, 17, 0, 0, 18, 0, 0, 0, 19, 0, 20, 0, 0,
      26, 0, 0, 27, 0, 0, 28, 0, 0, 29, 0, 0, 30, 0, 0
    ]);
  });
});
