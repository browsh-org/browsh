import utils from 'utils';
import CommonMixin from 'dom/common_mixin';
import DocumentBuilder from 'dom/document_builder';

// Entrypoint for managing a single tab
export default class extends utils.mixins(CommonMixin) {
  constructor() {
    super();
    // ID for element we place in the DOM to measure the size of a single monospace
    // character.
    this._measuring_box_id = 'browsh_em_measuring_box';
    this._setupInit();
  }

  sendFrame() {
    this.document_builder.makeFrame();
    this._sendTabInfo();
    if (!this._is_first_frame_finished) {
      this.sendMessage('/status,parsing_complete');
    }
    this.sendMessage(`/frame,${JSON.stringify(this.document_builder.frame)}`);
    this._is_first_frame_finished = true;
  }

  _setupInit() {
    // TODO: Can we not just boot up as soon as we detect the background script?
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

  _registrationError(error) {
    this.log(error);
  }

  _postCommsInit() {
    this.log('Webextension postCommsInit()');
    this.document_builder = new DocumentBuilder(this.channel)
    this._sendTabInfo();
    this.sendMessage('/status,page_init');
    this._calculateMonospaceDimensions();
    this._requestInitialTTYSize();
    this._listenForBackgroundMessages();
    window.addEventListener("unload", () => {
      this.sendMessage('/status,window_unload')
    });
    window.addEventListener('error', (event) => {
      this.log("TAB JS: " + event)
    });
  }

  _listenForBackgroundMessages() {
    this.channel.onMessage.addListener((message) => {
      try {
        this._handleBackgroundMessage(message);
      }
      catch(error) {
        this.log(`'${error.name}' ${error.message}`);
        this.log(`@${error.fileName}:${error.lineNumber}`);
        this.log(error.stack);
      }
    });
  }

  _handleBackgroundMessage(message) {
    let input, url;
    const parts = message.split(',');
    const command = parts[0];
    switch (command) {
      case '/request_frame':
        this.sendFrame();
        break;
      case '/tty_size':
        this.document_builder.tty_width = parseInt(parts[1]);
        this.document_builder.tty_height = parseInt(parts[2]);
        this.log(
          `Tab received TTY size: ` +
          `${this.document_builder.tty_width}x${this.document_builder.tty_height}`
        );
        break;
      case '/stdin':
        input = JSON.parse(utils.rebuildArgsToSingleArg(parts));
        this._handleUserInput(input);
        break;
      case '/url':
        url = utils.rebuildArgsToSingleArg(parts);
        document.location.href = url;
        break;
      case '/location_back':
        history.go(-1);
        break;
      case '/window_stop':
        window.stop();
        break;
      default:
        this.log('Unknown command sent to tab', message);
    }
  }

  _handleUserInput(input) {
    this._handleSpecialKeys(input);
    this._handleCharBasedKeys(input);
    this._handleMouse(input);
  }

  _handleSpecialKeys(input) {
    switch (input.key) {
      case 257: // up arow
        window.scrollBy(0, -2 * this.document_builder.char_height);
        break;
      case 258: // down arrow
        window.scrollBy(0, 2 * this.document_builder.char_height);
        break;
      case 266: // page up
        window.scrollBy(0, -this.document_builder.tty_height * this.document_builder.char_height);
        break;
      case 267: // page down
        window.scrollBy(0, this.document_builder.tty_height * this.document_builder.char_height);
        break;
      case 18: // CTRL+r
        window.location.reload();
        break;
    }
  }

  _handleCharBasedKeys(input) {
    switch (input.char) {
      case 'M':
        if (input.mod === 4) {
          this.document_builder.is_graphics_mode = !this.document_builder.is_graphics_mode;
        }
        break;
    }
  }

  _handleMouse(input) {
    switch (input.button) {
      case 256: // scroll up
        window.scrollBy(0, -20);
        break;
      case 512: // scroll down
        window.scrollBy(0, 20);
        break;
      case 1: // mousedown
        this._mouseAction('click', input.mouse_x, input.mouse_y);
        this._mouseAction('mousedown', input.mouse_x, input.mouse_y);
        break;
      case 0: //mouseup
        this._mouseAction('mouseup', input.mouse_x, input.mouse_y);
        break;
    }
  }

  _mouseAction(type, x, y) {
    const [dom_x, dom_y] = this._getDOMCoordsFromMouseCoords(x, y);
    const element = document.elementFromPoint(dom_x, dom_y);
    const event = new MouseEvent(type, {
      bubbles: true,
      cancelable: true,
      clientX: dom_x,
      clientY: dom_y
    });
    element.dispatchEvent(event);
  }

  // The user clicks on a TTY grid which has a significantly lower resolution than the
  // actual browser window. So we scale the coordinates up as if the user clicked on the
  // the central "pixel" of a TTY cell.
  _getDOMCoordsFromMouseCoords(x, y) {
    let dom_x, dom_y, char, original_position;
    y = y - 2; // Because of the UI header bar
    const index = (y * this.tty_width) + x;
    if (this.document_builder.tty_grid[index] !== undefined) {
      char = this.document_builder.tty_grid[index][0];
    } else {
      char = false;
    }
    if (!char || char === '▄') {
      dom_x = (x * this.document_builder.char_width);
      dom_y = (y * this.document_builder.char_height);
    } else {
      // Recall that text can be shifted from its original position in the browser in order
      // to snap it consistently to the TTY grid.
      original_position = this.document_builder.tty_grid[index][4];
      dom_x = original_position.x;
      dom_y = original_position.y;
    }
    return [
      dom_x + (this.document_builder.char_width / 2),
      dom_y + (this.document_builder.char_height / 2)
    ];
  }

  // The background process can't send the TTY size as soon as it gets it because maybe
  // the a tab doesn't exist yet. So we request it ourselves - because we'd have to be
  // ready in order to request.
  _requestInitialTTYSize() {
    this.sendMessage('/request_tty_size');
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
    this.document_builder.char_width = dom_rect.width;
    this.document_builder.char_height = dom_rect.height + 2; // TODO: WTF is this magic number?
    this.sendMessage(
      `/char_size,` +
      `${this.document_builder.char_width},` +
      `${this.document_builder.char_height}`
    );
    this.log(
      `Tab char dimensions: ` +
      `${this.document_builder.char_width}x${this.document_builder.char_height}`
    );
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

  _sendTabInfo() {
    const title_object = document.getElementsByTagName("title");
    let info = {
      url: document.location.href,
      title: title_object.length ? title_object[0].innerHTML : ""
    }
    this.sendMessage(`/tab_info,${JSON.stringify(info)}`);
  }
}
