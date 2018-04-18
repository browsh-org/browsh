export default {
  mixins: function (...mixins) {
    return mixins.reduce((base, mixin) => {
      return mixin(base);
    }, class {});
  },

  ttyCell: function (fg_colour, bg_colour, character) {
    fg_colour = fg_colour || [255, 255, 255];
    bg_colour = bg_colour || [0, 0, 0];
    let cell = fg_colour.concat(bg_colour);
    cell.push(character);
    return cell;
  },

  ttyPlainCell: function (character) {
    return this.ttyCell(null, null, character);
  },

  snap: function (number) {
    return parseInt(Math.round(number));
  },

  rebuildArgsToSingleArg: function (args) {
    return args.slice(1).join(',');
  }
}
