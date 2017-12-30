import BaseBuilder from 'base_builder';
import GraphicsBuilder from 'graphics_builder';
import TextBuilder from 'text_builder';

// Takes the graphics and text from the current viewport, combines them, then
// sends it to the background process where the the rest of the UI, like tabs,
// address bar, etc will be added.
export default class FrameBuilder extends BaseBuilder{
  constructor() {
    super();
    // ID for element we place in the DOM to measure the size of a single monospace
    // character.
    this.measuring_box_id = 'browsh_em_measuring_box';
    this.graphics_builder = new GraphicsBuilder();
    this.text_builder = new TextBuilder();

    document.addEventListener("DOMContentLoaded", () => {
      this.init();
    }, false);

    // Whilst developing this webextension the auto reload only reloads this code,
    // not the page, so we don't get the `DOMContentLoaded` event to kick everything off.
    if (this.isWindowAlreadyLoaded()) this.init(100);
  }

  isWindowAlreadyLoaded() {
    return !!this.findMeasuringBox();
  }

  init(delay = 0) {
    console.log('Browsh init()');
    this.calculateMonospaceDimensions();
    // When the webext devtools auto reloads this code the background process
    // can sometimes still be loading, in which case we need to wait.
    setTimeout(() => this.registerWithBackground(), delay);
  }

  registerWithBackground() {
    let sending = browser.runtime.sendMessage('/register');
    sending.then(
      (r) => this.registrationSuccess(r),
      (e) => this.registrationError(e)
    );
  }

  // The background process tells us when it wants a frame.
  registrationSuccess(registered) {
    this.channel = browser.runtime.connect({
      // We need to give ourselves a unique channel name, so the background
      // process can identify us amongst other tabs.
      name: registered.id.toString()
    });
    this.channel.onMessage.addListener((message) => {
      const parts = message.split(',');
      const command = parts[0];
      const tty_width = parseInt(parts[1]);
      const tty_height = parseInt(parts[2]);
      if (command === '/send_frame') {
        this.sendFrame(tty_width, tty_height);
      }
    });
  }

  registrationError(error) {
    console.error(error);
  }

  // This is critical in order for the terminal to match the browser as closely as possible.
  // Ideally we want the browser's window size to be exactly multiples of the terminal's
  // dimensions. So if the terminal is 80x40 and the font-size is 12px (12x6 pixels), then
  // the window should be 480x480. Also knowing the precise font-size helps the text builder
  // map un-snapped text to the best grid cells - grid cells that represent the terminal's
  // character positions.
  // The reason that we can't just do some basic maths on the CSS `font-size` value we enforce
  // is that there are various factors that can skew the actual font dimensions on the page.
  // For instance, you can't guarantee that a browser is using exactly the same version of
  // a named monospace font. Also different browser families and even different versions of
  // the same browser may have subtle differences in how they render text. Furthermore we can
  // actually get floating point accuracy if we use `Element.getBoundingClientRect()` which
  // further helps as calculations are compounded during our rendering processes.
  calculateMonospaceDimensions() {
    const element = this.getOrCreateMeasuringBox();
    const dom_rect = element.getBoundingClientRect();
    this.char_width = dom_rect.width;
    this.char_height = dom_rect.height;
    this.text_builder.char_width = this.char_width;
    this.text_builder.char_height = this.char_height;
    console.log('char dimensions', this.char_width, this.char_height);
  }

  getOrCreateMeasuringBox() {
    let measuring_box = this.findMeasuringBox();
    if (measuring_box) return measuring_box;
    measuring_box = document.createElement('span');
    measuring_box.id = this.measuring_box_id;
    measuring_box.style.visibility = 'hidden';
    // Back when printing was done by physical stamps, it was convention to measure the
    // font-size using the letter 'M', thus where we get the unit 'em' from. Not that it
    // should make any difference to us, but it's nice to keep a tradition.
    var M = document.createTextNode('M');
    measuring_box.appendChild(M);
    document.body.appendChild(measuring_box);
    return measuring_box;
  }

  findMeasuringBox() {
    return document.getElementById(this.measuring_box_id);
  }

  sendFrame(tty_width, tty_height) {
    this.setupDimensions(tty_width, tty_height);
    this.compileFrame();
    this.logPerformance(() => {
      this.buildFrame();
    }, 'build frame');
    this.channel.postMessage(this.screen);
    this.is_first_frame_finished = true;
  }

  setupDimensions(tty_width, tty_height) {
    this.tty_width = tty_width;
    this.tty_height = tty_height;
    this.frame_width = tty_width;
    // A frame is 'taller' than the TTY because of the special UTF8 half-block
    // trick.
    this.frame_height = tty_height * 2;
  }

  compileFrame() {
    this.logPerformance(() => {
      this.pixels_with_text = this.graphics_builder.getSnapshotWithText(
        this.frame_width,
        this.frame_height
      );
    }, 'get snapshot with text');
    this.logPerformance(() => {
      this.pixels_without_text = this.graphics_builder.getSnapshotWithoutText(
      this.frame_width,
      this.frame_height
    );
    }, 'get snapshot without text');
    this.formatted_text = this.text_builder.getFormattedText(
      this,
      this.graphics_builder
    );
  }

  buildFrame() {
    this.screen = "";
    this.bg_row = [];
    this.fg_row = [];
    for (let y = 0; y < this.frame_height; y++) {
      for (let x = 0; x < this.frame_width; x++) {
        this.buildPixel(x, y);
      }
    }
  }

  // Note how we have to keep track of 2 rows of pixels in order to create 1 row of
  // the terminal.
  buildPixel(x, y) {
    let pixel_data_start, r, g, b;
    pixel_data_start = y * (this.frame_width * 4) + (x * 4);
    r = this.pixels_without_text[pixel_data_start + 0];
    g = this.pixels_without_text[pixel_data_start + 1];
    b = this.pixels_without_text[pixel_data_start + 2];
    if (this.bg_row.length < this.frame_width) {
      this.bg_row.push([r, g, b]);
    } else {
      this.fg_row.push([r, g, b]);
    }
    if (this.fg_row.length === this.frame_width) {
      this.screen += this.buildTtyRow(this.bg_row, this.fg_row, y);
      this.bg_row = [];
      this.fg_row = [];
    }
  }

  // This is where we implement the UTF8 half-block trick.
  // This is a half-black: ▄ Notice how it takes up precisely half a text cell. This
  // means that we can get 2 pixel colours from it, the top pixel comes from setting
  // the background colour and the bottom pixel comes from setting the foreground
  // colour, namely the colour of the text.
  // However we can't just write random pixels to a TTY screen, we must collate 2 rows
  // of native pixels for every row of the terminal.
  buildTtyRow(bg_row, fg_row, y) {
    let tty_index, char;
    let row = "";
    const tty_row = parseInt(y / 2);
    for (let x = 0; x < this.frame_width; x++) {
      tty_index = (tty_row * this.frame_width) + x;
      if (this.doesCellHaveACharacter(tty_index)) {
        char = this.formatted_text[tty_index];
        row += this.ttyPixel(char[1], char[2], char[0]);
      } else {
        row += this.ttyPixel(fg_row[x], bg_row[x], '▄');
      }
    }
    if (tty_row + 1 < this.tty_height) {
      row += "\n";
    }
    return row;
  }

  // We need to know this because we want all empty cells to be 'transparent'
  doesCellHaveACharacter(index) {
    if (this.formatted_text[index] === undefined) return false;
    const char = this.formatted_text[index][0];
    const is_undefined = char === undefined;
    const is_empty = char === '';
    const is_space = /^\s+$/.test(char);
    const is_not_worth_printing = is_empty || is_space || is_undefined;
    return !is_not_worth_printing;
  }

  // Display a single character in true colour
  ttyPixel(fg, bg, character) {
    let fg_code = `\x1b[38;2;${fg[0]};${fg[1]};${fg[2]}m`;
    let bg_code = `\x1b[48;2;${bg[0]};${bg[1]};${bg[2]}m`;
    return `${fg_code}${bg_code}${character}`;
  }
}
