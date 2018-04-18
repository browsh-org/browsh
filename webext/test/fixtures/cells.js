import TTYCell from "dom/tty_cell";

let base = build(
  [
    '',   '😐', '',
    '😄', '',   '😂'
  ],
  [111, 111, 111],
  [222, 222, 222]
);

export default base;

function build(text, fg_colour, bg_colour) {
  let cell;
  let grid = [];
  for(const character of text) {
    cell = new TTYCell();
    cell.rune = character;
    cell.fg_colour = fg_colour;
    cell.bg_colour = bg_colour;
    grid.push(cell);
  }
  return grid;
}
