export default {
  mixins: function(...mixins) {
    return mixins.reduce((base, mixin) => {
      return mixin(base);
    }, class {});
  },

  ttyCell: function(
    fg_colour = [255, 255, 255],
    bg_colour = [0, 0, 0],
    character
  ) {
    let cell = fg_colour.concat(bg_colour);
    cell.push(character);
    return cell;
  },

  ttyPlainCell: function(character) {
    return this.ttyCell(null, null, character);
  },

  snap: function(number) {
    return parseInt(Math.round(number));
  },

  ensureEven: function(number) {
    number = this.snap(number);
    if (number % 2) {
      number++;
    }
    return number;
  },

  rebuildArgsToSingleArg: function(args) {
    return args.slice(1).join(",");
  },

  uuidv4: function() {
    return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function(c) {
      var r = (Math.random() * 16) | 0,
        v = c == "x" ? r : (r & 0x3) | 0x8;
      return v.toString(16);
    });
  }
};
