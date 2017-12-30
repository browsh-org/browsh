let base = build(
  [
    '',  'x', '',
    'x', '',  'x'
  ],
  [111, 111, 111],
  [222, 222, 222]
);

export default base;

function build(text, fg_colour, bg_colour) {
  let grid = [];
  for(const character of text) {
    grid.push([character, fg_colour, bg_colour]);
  }
  return grid;
}
