import BaseBuilder from 'dom/base_builder';
import GraphicsBuilder from 'dom/graphics_builder';
import TextBuilder from 'dom/text_builder';

// Takes the graphics and text from the current viewport, combines them, then
// sends it to the background process where the rest of the UI, like tabs,
// address bar, etc will be added.
export default class FrameBuilder extends BaseBuilder{
  constructor() {
    super();
    this.graphics_builder = new GraphicsBuilder();
    this.text_builder = new TextBuilder(this);
    // ID for element we place in the DOM to measure the size of a single monospace
    // character.
    this._measuring_box_id = 'browsh_em_measuring_box';
    this._setupInit();
  }

  sendFrame() {
    this._setupDimensions();
    this._compileFrame();
    this._buildFrame();
    this._sendMessage(`/frame,${this.frame}`);
    this._is_first_frame_finished = true;
  }

  _setupInit() {
    document.addEventListener("DOMContentLoaded", () => {
      this._init();
    }, false);
    // Whilst developing this webextension the auto reload only reloads this code,
    // not the page, so we don't get the `DOMContentLoaded` event to kick everything off.
    if (this._isWindowAlreadyLoaded()) this._init(100);
  }

  _isWindowAlreadyLoaded() {
    return !!this._findMeasuringBox();
  }

  _init(delay = 0) {
    // When the webext devtools auto reloads this code, the background process
    // can sometimes still be loading, in which case we need to wait.
    setTimeout(() => this._registerWithBackground(), delay);
  }

  _registerWithBackground() {
    let sending = browser.runtime.sendMessage('/register');
    sending.then(
      (r) => this._registrationSuccess(r),
      (e) => this._registrationError(e)
    );
  }

  _registrationSuccess(registered) {
    this.channel = browser.runtime.connect({
      // We need to give ourselves a unique channel name, so the background
      // process can identify us amongst other tabs.
      name: registered.id.toString()
    });
    this._postCommsInit();
  }

  _postCommsInit() {
    this._log('Webextension postCommsInit()');
    this._calculateMonospaceDimensions();
    this.graphics_builder.channel = this.channel;
    this.text_builder.channel = this.channel;
    this._requestInitialTTYSize();
    this._listenForBackgroundMessages();
  }

  _listenForBackgroundMessages() {
    this.channel.onMessage.addListener((message) => {
      let input;
      const parts = message.split(',');
      const command = parts[0];
      switch (command) {
        case '/request_frame':
          this.sendFrame();
          break;
        case '/tty_size':
          this.tty_width = parseInt(parts[1]);
          this.tty_height = parseInt(parts[2]);
          this._log(`Tab received TTY size: ${this.tty_width}x${this.tty_height}`);
          break;
        case '/stdin':
          input = JSON.parse(parts.slice(1).join(','));
          this._handleUserInput(input);
          break;
        default:
          this._log('Unknown command sent to tab', message);
      }
    });
  }

  _handleUserInput(input) {
    this._log(input.key);
    switch (input.key) {
      case 65517:
        window.scrollBy(0, -20);
        break;
      case 65516:
        window.scrollBy(0, 20);
       break;
    }
  }

  _registrationError(error) {
    this._log(error);
  }

  // The background process can't send the TTY size as soon as it gets it because maybe
  // the a tab doesn't exist yet. So we request it ourselves - because we'd have to be
  // ready in order to request.
  _requestInitialTTYSize() {
    this._sendMessage('/request_tty_size');
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
  _calculateMonospaceDimensions() {
    const element = this._getOrCreateMeasuringBox();
    const dom_rect = element.getBoundingClientRect();
    this.char_width = dom_rect.width;
    this.char_height = dom_rect.height + 2;
    this.text_builder.char_width = this.char_width;
    this.text_builder.char_height = this.char_height;
    this._sendMessage(`/char_size,${this.char_width},${this.char_height}`);
    this._log(`Tab char dimensions: ${this.char_width}x${this.char_height}`);
  }

  // Back when printing was done by physical stamps, it was convention to measure the
  // font-size using the letter 'M', thus where we get the unit 'em' from. Not that it
  // should make any difference to us, but it's nice to keep a tradition.
  _getOrCreateMeasuringBox() {
    let measuring_box = this._findMeasuringBox();
    if (measuring_box) return measuring_box;
    measuring_box = document.createElement('span');
    measuring_box.id = this._measuring_box_id;
    measuring_box.style.visibility = 'hidden';
    var M = document.createTextNode('M');
    measuring_box.appendChild(M);
    document.body.appendChild(measuring_box);
    return measuring_box;
  }

  _findMeasuringBox() {
    return document.getElementById(this._measuring_box_id);
  }

  _setupDimensions() {
    if (!this.tty_width || !this.tty_height) {
      throw new Error("Frame Builder doesn't have the TTY dimensions");
    }
    this.frame_width = this.tty_width;
    // A frame is 'taller' than the TTY because of the special UTF8 half-block
    // trick.
    this.frame_height = this.tty_height * 2;
  }

  _compileFrame() {
    this.graphics_builder.getSnapshotWithText();
    this.graphics_builder.getSnapshotWithoutText();
    this.graphics_builder.getScaledSnapshot(
      this.frame_width,
      this.frame_height
    );
    this.formatted_text = this.text_builder.getFormattedText();
  }

  _buildFrame() {
    this._logPerformance(() => {
      this.__buildFrame();
    }, 'build frame');
  }

  __buildFrame() {
    this.frame = "";
    this._bg_row = [];
    this._fg_row = [];
    for (let y = 0; y < this.frame_height; y++) {
      for (let x = 0; x < this.frame_width; x++) {
        this._buildPixel(x, y);
      }
    }
  }

  // Note how we have to keep track of 2 rows of pixels in order to create 1 row of
  // the terminal.
  _buildPixel(x, y) {
    const colour = this.graphics_builder.getScaledPixelAt(x, y);
    if (this._bg_row.length < this.frame_width) {
      this._bg_row.push(colour);
    } else {
      this._fg_row.push(colour);
    }
    if (this._fg_row.length === this.frame_width) {
      this.frame += this._buildTtyRow(this._bg_row, this._fg_row, y);
      this._bg_row = [];
      this._fg_row = [];
    }
  }

  // This is where we implement the UTF8 half-block trick.
  // This is a half-black: ▄ Notice how it takes up precisely half a text cell. This
  // means that we can get 2 pixel colours from it, the top pixel comes from setting
  // the background colour and the bottom pixel comes from setting the foreground
  // colour, namely the colour of the text.
  // However we can't just write random pixels to a TTY screen, we must collate 2 rows
  // of native pixels for every row of the terminal.
  _buildTtyRow(bg_row, fg_row, y) {
    let tty_index, char;
    let row = "";
    const tty_row = parseInt(y / 2);
    for (let x = 0; x < this.frame_width; x++) {
      tty_index = (tty_row * this.frame_width) + x;
      if (this._doesCellHaveACharacter(tty_index)) {
        char = this.formatted_text[tty_index];
        row += this._ttyPixel(char[1], char[2], char[0]);
      } else {
        row += this._ttyPixel(fg_row[x], bg_row[x], '▄');
      }
    }
    if (tty_row + 1 < this.tty_height) {
      row += "\n";
    }
    return row;
  }

  // We need to know this because we want all empty cells to be 'transparent'
  _doesCellHaveACharacter(index) {
    if (this.formatted_text[index] === undefined) return false;
    const char = this.formatted_text[index][0];
    const is_undefined = char === undefined;
    const is_empty = char === '';
    const is_space = /^\s+$/.test(char);
    const is_not_worth_printing = is_empty || is_space || is_undefined;
    return !is_not_worth_printing;
  }

  // Display a single character in true colour
  _ttyPixel(fg, bg, character) {
    let fg_code = `\x1b[38;2;${fg[0]};${fg[1]};${fg[2]}m`;
    let bg_code = `\x1b[48;2;${bg[0]};${bg[1]};${bg[2]}m`;
    return `${fg_code}${bg_code}${character}`;
  }
}
