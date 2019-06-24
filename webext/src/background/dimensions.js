import _ from "lodash";

import utils from "utils";
import CommonMixin from "background/common_mixin";

export default class extends utils.mixins(CommonMixin) {
  constructor() {
    super();
    this.tty = {};
    this.char = {};
    // I *think* this extra height is needed because the browser window height is not the same
    // as the actual viewport height. But then if that is the case, then we'll also have a
    // similar issue with the scroll bars.
    // TODO: Also if this hypothesis is correct, it needs to be applied as an original browser-
    // relative pixel unit, not as a TTY unit. If you look on Google Maps I think you can
    // actually see a little bit of white at the bottom perhaps from where the screen capture
    // goes over the bottom of the viewport.
    this._window_ui_magic_number = 3;
  }

  postConfigSetup(config) {
    this.config = config;
    this._setRawTextTTYSize();
  }

  setCharValues(incoming) {
    if (
      this.char.width != incoming.width ||
      this.char.height != incoming.height
    ) {
      this.log(
        `Requesting browser resize for new char dimensions: ` +
          `${incoming.width}x${incoming.height} (old: ${this.char.width}x${
            this.char.height
          })`
      );
      this.char = _.clone(incoming);
      this.resizeBrowserWindow();
    }
  }

  // The Browsh HTTP Server service doesn't load a TTY, so we need to supply the size.
  // Strictly it shouldn't even be needed if the code was completely refactored. Although
  // it should be worth taking into consideration how the size of the TTY and therefore the
  // resized browser window affects the rendering of a web page, for instance images outside
  // of the viewport can sometimes not be loaded. So is it practical to set the TTY size to
  // the size of the entire DOM?
  _setRawTextTTYSize() {
    this.raw_text_tty_size = {
      width: this.config["http-server"].columns,
      height: this.config["http-server"].rows
    };
  }

  resizeBrowserWindow() {
    if (
      !this.tty.width ||
      !this.char.width ||
      !this.tty.height ||
      !this.char.height
    ) {
      this.log(
        "Not resizing browser window without all of the TTY and character dimensions"
      );
      return;
    }
    // Does this include scrollbars???
    const window_width = parseInt(Math.round(this.tty.width * this.char.width));
    // Leave room for tabs and URL bar
    const tty_dom_height = this.tty.height - 2;
    const window_height = parseInt(
      Math.round(
        (tty_dom_height + this._window_ui_magic_number) * this.char.height
      )
    );
    const current_window = browser.windows.getCurrent();
    current_window.then(
      active_window => {
        this._sendWindowResizeRequest(
          active_window,
          window_width,
          window_height
        );
      },
      error => {
        this.log("Error getting current browser window", error);
      }
    );
  }

  _sendWindowResizeRequest(active_window, width, height) {
    const tag = "Resizing browser window";
    const updating = browser.windows.update(active_window.id, {
      width: width,
      height: height,
      focused: false
    });
    updating.then(
      info => this.log(`${tag} successful (${info.width}x${info.height})`),
      error => this.log(tag + " error: ", error)
    );
  }
}
