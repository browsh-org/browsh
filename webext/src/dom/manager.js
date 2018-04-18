import utils from 'utils';

import CommonMixin from 'dom/common_mixin';
import Dimensions from 'dom/dimensions';
import FrameBuilder from 'dom/frame_builder';

// Entrypoint for managing a single tab
export default class extends utils.mixins(CommonMixin) {
  constructor() {
    super();
    this.dimensions = new Dimensions();
    this._setupInit();
  }

  sendFrame() {
    this.frame_builder.makeFrame();
    this._sendTabInfo();
    if (!this._is_first_frame_finished) {
      this.sendMessage('/status,parsing_complete');
    }
    if (this.frame_builder.frame.length > 0) {
      this.sendMessage(`/frame,${JSON.stringify(this.frame_builder.frame)}`);
    } else {
      this.log("Not sending empty frame");
    }
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

  _postCommsInit() {
    this.log('Webextension postCommsInit()');
    this.dimensions.channel = this.channel;
    this.frame_builder = new FrameBuilder(this.channel, this.dimensions);
    this._sendTabInfo();
    this.sendMessage('/status,page_init');
    this._listenForBackgroundMessages();
    this._startWindowEventListeners()
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
        this._buildText();
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
        window.scrollBy(0, -2 * this.dimensions.char.height);
        break;
      case 258: // down arrow
        window.scrollBy(0, 2 * this.dimensions.char.height);
        break;
      case 266: // page up
        window.scrollBy(0, -window.innerHeight);
        break;
      case 267: // page down
        window.scrollBy(0, window.innerHeight);
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
          this.frame_builder.is_graphics_mode = !this.frame_builder.is_graphics_mode;
          this.frame_builder.buildText();
        }
        break;
    }
  }

  _handleMouse(input) {
    switch (input.button) {
      case 256: // scroll up
        window.scrollBy(0, -2);
        break;
      case 512: // scroll down
        window.scrollBy(0, 2);
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
    const [dom_x, dom_y] = this.frame_builder.getDOMCoordsFromMouseCoords(x, y);
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

  _sendTabInfo() {
    const title_object = document.getElementsByTagName("title");
    let info = {
      url: document.location.href,
      title: title_object.length ? title_object[0].innerHTML : ""
    }
    this.sendMessage(`/tab_info,${JSON.stringify(info)}`);
  }
}
