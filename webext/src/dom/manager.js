import utils from 'utils';

import CommonMixin from 'dom/common_mixin';
import Dimensions from 'dom/dimensions';
import GraphicsBuilder from 'dom/graphics_builder';
import TextBuilder from 'dom/text_builder';

// Entrypoint for managing a single tab
export default class extends utils.mixins(CommonMixin) {
  constructor() {
    super();
    this.dimensions = new Dimensions();
    this._setupInit();
  }

  _postCommsConstructor() {
    this.dimensions.channel = this.channel;
    this.dimensions.update();
    this.graphics_builder = new GraphicsBuilder(this.channel, this.dimensions);
    this.text_builder = new TextBuilder(
      this.channel,
      this.dimensions,
      this.graphics_builder
    );
    if (!this._is_first_frame_finished) {
      this.sendAllBigFrames();
    }
  }

  sendFrame() {
    this.dimensions.update()
    if (this.dimensions.dom.is_new) {
      this.sendAllBigFrames();
    }
    this.sendSmallPixelFrame();
    this._sendTabInfo();
    if (!this._is_first_frame_finished) {
      this.sendMessage('/status,parsing_complete');
    }
    this._is_first_frame_finished = true;
  }

  sendAllBigFrames() {
    if (!this.dimensions.tty.width) {
      this.log("Not sending big frames without TTY data")
      return
    } else {
      this.log("Sending big frames...")
    }
    this.dimensions.update();
    this.dimensions.setSubFrameDimensions('big');
    this.text_builder.sendFrame();
    this.graphics_builder.sendFrame();
    this.dimensions.frame.x_last_big_frame = this.dimensions.frame.x_scroll;
    this.dimensions.frame.y_last_big_frame = this.dimensions.frame.y_scroll;
  }

  sendSmallPixelFrame() {
    if (!this.dimensions.tty.width) {
      this.log("Not sending small frames without TTY data")
      return
    }
    this.dimensions.update()
    this.dimensions.setSubFrameDimensions('small');
    this.graphics_builder.sendFrame();
  }

  _postCommsInit() {
    this.log('Webextension postCommsInit()');
    this._postCommsConstructor();
    this._sendTabInfo();
    this.sendMessage('/status,page_init');
    this._listenForBackgroundMessages();
    this._startWindowEventListeners()
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
    return !!this.dimensions.findMeasuringBox();
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

  _startWindowEventListeners() {
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
      case '/rebuild_text':
        break;
      case '/scroll_status':
        this.dimensions.frame.x_scroll = parseInt(parts[1]);
        this.dimensions.frame.y_scroll = parseInt(parts[2]);
        this.dimensions.update();
        this._handleScroll();
        this._mightSendBigFrames();
        break;
      case '/tty_size':
        this.dimensions.tty.width = parseInt(parts[1]);
        this.dimensions.tty.height = parseInt(parts[2]);
        this.dimensions.update();
        if (!this._is_first_frame_finished) {
          this.sendAllBigFrames();
        }
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
      case 18: // CTRL+r
        window.location.reload();
        break;
    }
  }

  _handleCharBasedKeys(input) {
    switch (input.char) { default: }
  }

  _handleMouse(input) {
    switch (input.button) {
      case 1: // mousedown
        this._mouseAction('click', input.mouse_x, input.mouse_y);
        this._mouseAction('mousedown', input.mouse_x, input.mouse_y);
        break;
      case 0: //mouseup
        this._mouseAction('mouseup', input.mouse_x, input.mouse_y);
        break;
    }
  }

  _handleScroll() {
    window.scrollTo(
      this.dimensions.frame.x_scroll / this.dimensions.scale_factor.width,
      this.dimensions.frame.y_scroll / this.dimensions.scale_factor.height,
    );
  }

  _mouseAction(type, x, y) {
    const [dom_x, dom_y] = this._getDOMCoordsFromMouseCoords(x, y);
    const element = document.elementFromPoint(
      dom_x - window.scrollX,
      dom_y - window.scrollY
    );
    const event = new MouseEvent(type, {
      bubbles: true,
      cancelable: true,
      pageX: dom_x,
      pageY: dom_y
    });
    element.dispatchEvent(event);
  }

  // The user clicks on a TTY grid which has a significantly lower resolution than the
  // actual browser window. So we scale the coordinates up as if the user clicked on the
  // the central "pixel" of a TTY cell.
  //
  // Furthermore if the TTY click is on a readable character then the click is proxied
  // to the original position of the character before TextBuilder snapped the character into
  // position.
  _getDOMCoordsFromMouseCoords(x, y) {
    let dom_x, dom_y, char, original_position;
    const index = (y * this.dimensions.frame.width) + x;
    if (this.text_builder.tty_grid.cells[index] !== undefined) {
      char = this.text_builder.tty_grid.cells[index].rune;
    } else {
      char = false;
    }
    if (!char || char === '▄') {
      dom_x = (x * this.dimensions.char.width);
      dom_y = (y * this.dimensions.char.height);
    } else {
      // Recall that text can be shifted from its original position in the browser in order
      // to snap it consistently to the TTY grid.
      original_position = this.text_builder.tty_grid.cells[index].dom_coords;
      dom_x = original_position.x;
      dom_y = original_position.y;
    }
    return [
      dom_x + (this.dimensions.char.width / 2),
      dom_y + (this.dimensions.char.height / 2)
    ];
  }

  _sendTabInfo() {
    const title_object = document.getElementsByTagName("title");
    let info = {
      url: document.location.href,
      title: title_object.length ? title_object[0].innerHTML : ""
    }
    this.sendMessage(`/tab_info,${JSON.stringify(info)}`);
  }

  _mightSendBigFrames() {
    const y_diff = this.dimensions.frame.y_last_big_frame - this.dimensions.frame.y_scroll;
    const max_y_scroll_without_new_big_frame = 2 * this.dimensions.tty.height;
    if (Math.abs(y_diff) > max_y_scroll_without_new_big_frame) {
      this.log(
        `Parsing big frames: ` +
        `previous-y: ${this.dimensions.frame.y_last_big_frame}, ` +
        `y-scroll: ${this.dimensions.frame.y_scroll}, ` +
        `diff: ${y_diff}, ` +
        `max-scroll: ${max_y_scroll_without_new_big_frame} `
      )
      this.sendAllBigFrames();
    }
  }
}
