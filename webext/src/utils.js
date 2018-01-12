// Display a single character in true colour
export default {
  ttyPixel: function(fg, bg, character) {
    let fg_code = `\x1b[38;2;${fg[0]};${fg[1]};${fg[2]}m`;
    let bg_code = `\x1b[48;2;${bg[0]};${bg[1]};${bg[2]}m`;
    return `${fg_code}${bg_code}${character}`;
  },

  rebuildArgsToSingleArg: function(args) {
    return args.slice(1).join(',');
  }
}
