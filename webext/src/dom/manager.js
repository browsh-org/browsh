import _ from "lodash";

import utils from "utils";

import CommonMixin from "dom/common_mixin";
import CommandsMixin from "dom/commands_mixin";
import Dimensions from "dom/dimensions";
import GraphicsBuilder from "dom/graphics_builder";
import TextBuilder from "dom/text_builder";

// Entrypoint for managing a single tab
export default class extends utils.mixins(CommonMixin, CommandsMixin) {
  constructor() {
    super();
    this.dimensions = new Dimensions();
    // Whether the DOM has loaded
    this.is_dom_loaded = false;
    // Whether the page has finished "spinning"
    this.is_page_finished_loading = false;
    // For Browsh used via the interactive CLI ap
    this._is_interactive_mode = false;
    // For Browsh used via the HTTP server
    this._is_raw_mode = false;
    this._setupInit();
  }

  _postSetupConstructor() {
    this._injectCustomCSS();
    this.dimensions.channel = this.channel;
    this.graphics_builder = new GraphicsBuilder(
      this.channel,
      this.dimensions,
      this.config
    );
    this.text_builder = new TextBuilder(
      this.channel,
      this.dimensions,
      this.graphics_builder,
      this.config
    );
  }

  _willHideText() {
    if (this.is_dom_loaded && this.graphics_builder) {
      this.graphics_builder.hideText();
    } else {
      setTimeout(this._willHideText.bind(this), 1);
    }
  }

  sendFrame() {
    this.dimensions.update();
    if (this.dimensions.dom.is_new) {
      this.sendAllBigFrames();
    }
    this.sendSmallPixelFrame();
    this._sendTabInfo();
    if (!this._is_first_frame_finished) {
      this.sendMessage("/status,parsing_complete");
    }
    this._is_first_frame_finished = true;
  }

  sendAllBigFrames() {
    if (!this._is_interactive_mode) {
      return;
    }
    if (!this.dimensions.tty.width) {
      this.log("Not sending big frames without TTY data");
      return;
    } else {
      this.log("Sending big frames...");
    }
    this.dimensions.update();
    this.dimensions.setSubFrameDimensions("big");
    this.text_builder.sendFrame();
    this.graphics_builder.sendFrame();
    this.dimensions.frame.x_last_big_frame = this.dimensions.frame.x_scroll;
    this.dimensions.frame.y_last_big_frame = this.dimensions.frame.y_scroll;
  }

  sendRawText() {
    if (this.is_page_finished_loading) {
      this.dimensions.update();
      this.dimensions.setSubFrameDimensions("raw_text");
      this.text_builder.sendRawText(this._raw_mode_type);
    } else {
      setTimeout(this.sendRawText.bind(this), 1);
    }
  }

  sendSmallPixelFrame() {
    if (!this._is_interactive_mode) {
      return;
    }
    if (!this.dimensions.tty.width) {
      this.log("Not sending small frames without TTY data");
      return;
    }
    this.dimensions.update();
    this.dimensions.setSubFrameDimensions("small");
    this.graphics_builder.sendFrame();
  }

  sendSmallTextFrame() {
    if (!this._is_interactive_mode) {
      return;
    }
    if (!this.dimensions.tty.width) {
      this.log("Not sending small frames without TTY data");
      return;
    }
    this.dimensions.update();
    this.dimensions.setSubFrameDimensions("small");
    this.text_builder.sendFrame();
  }

  _postCommsInit() {
    this.log("Webextension postCommsInit()");
    this._sendTabInfo();
    this.sendMessage("/status,page_init");
    this._listenForBackgroundMessages();
    this._startWindowEventListeners();
  }

  // Fire up the TTY interactive mode. It doesn't need to wait for any particular
  // DOM stage as it's good to just get something in front of the user as soon
  // as possible.
  _setupInteractiveMode() {
    this._setupDebouncedFunctions();
    this._startMutationObserver();
    this.sendAllBigFrames();
    // TODO:
    //   Disabling CSS transitions is not easy, many pages won't even render
    //   if they're disabled. Eg; Google's login process.
    //   What if we could get a post-transition hook?
    setTimeout(() => {
      this.sendAllBigFrames();
    }, 500);
  }

  _setupDebouncedFunctions() {
    this._debouncedSmallTextFrame = _.debounce(this.sendSmallTextFrame, 100, {
      leading: true
    });
  }

  _setupInit() {
    if (this._isWindowAlreadyLoaded()) {
      this._init(100);
    } else {
      this._init();
    }
  }

  _isWindowAlreadyLoaded() {
    if (document.body === undefined) {
      return false;
    }
    return !!this.dimensions.findMeasuringBox();
  }

  _init(delay = 0) {
    // When the webext devtools auto reloads this code, the background process
    // can sometimes still be loading, in which case we need to wait.
    setTimeout(() => this._registerWithBackground(), delay);
  }

  _registerWithBackground() {
    let sending = browser.runtime.sendMessage("/register");
    sending.then(
      r => this._registrationSuccess(r),
      e => this._registrationError(e)
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
    window.addEventListener("DOMContentLoaded", () => {
      this.is_dom_loaded = true;
      this.log("DOM LOADED");
      this._fixStickyElements();
      this._willHideText();
    });
    window.addEventListener("load", () => {
      this.is_page_finished_loading = true;
      this.config.page_load_duration = Date.now() - this.config.start_time;
      this.log("PAGE LOADED");
    });
    window.addEventListener("unload", () => {
      this.sendMessage("/status,window_unload");
    });
    window.addEventListener("error", error => {
      this.logError(error);
    });
  }

  _startMutationObserver() {
    let target = document.querySelector("body");
    let observer = new MutationObserver(mutations => {
      mutations.forEach(mutation => {
        this.log("!!MUTATION!!", mutation);
        this._debouncedSmallTextFrame();
      });
    });
    observer.observe(target, {
      subtree: true,
      characterData: true,
      childList: true
    });
  }

  _listenForBackgroundMessages() {
    this.channel.onMessage.addListener(message => {
      try {
        this._handleBackgroundMessage(message);
      } catch (error) {
        this.logError(error);
      }
    });
  }

  // Sticky elements are, for example, those headers that follow you down the page as you
  // scroll. They are annoying even in desktop browsers, however because of the lower frame
  // rate of Browsh, sticky elements stutter down the page, so it's even more annoying. Not
  // to mention the screen real estate that sticky elements take up, which is even more
  // noticeable on a small TTY screen like Browsh's.
  //
  // Note that this uses `getComputedStyle()`, which can be expensive, there should only
  // be 1 that parses that entire tree during page load. So if there's reason to use more
  // CSS tricks like this, then the call should be refactored.
  _fixStickyElements() {
    let position;
    let i,
      elements = document.querySelectorAll("body *");
    for (i = 0; i < elements.length; i++) {
      position = getComputedStyle(elements[i]).position;
      if (position === "fixed" || position === "sticky") {
        elements[i].style.setProperty("position", "absolute", "important");
      }
    }
  }

  _injectCustomCSS() {
    var node = document.createElement("style");
    node.innerHTML = this.config.browsh.custom_css;
    if (document.body) {
      document.body.appendChild(node);
    }
  }
}
