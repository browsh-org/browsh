import utils from "utils";

export default MixinBase =>
  class extends MixinBase {
    _handleBackgroundMessage(message) {
      let input, url, config;
      const parts = message.split(",");
      const command = parts[0];
      switch (command) {
        case "/config":
          config = JSON.parse(utils.rebuildArgsToSingleArg(parts));
          this._loadConfig(config);
          break;
        case "/request_frame":
          this.sendFrame();
          break;
        case "/rebuild_text":
          if (this._is_interactive_mode) {
            this.sendAllBigFrames();
          }
          break;
        case "/scroll_status":
          this._handleScroll(parts[1], parts[2]);
          break;
        case "/tty_size":
          this._handleTTYSize(parts[1], parts[2]);
          break;
        case "/stdin":
          input = JSON.parse(utils.rebuildArgsToSingleArg(parts));
          this._handleUserInput(input);
          break;
        case "/input_box":
          input = JSON.parse(utils.rebuildArgsToSingleArg(parts));
          this._handleInputBoxContent(input);
          break;
        case "/url":
          url = utils.rebuildArgsToSingleArg(parts);
          document.location.href = url;
          break;
        case "/history_back":
          history.go(-1);
          break;
        case "/window_stop":
          window.stop();
          break;
        default:
          this.log("Unknown command sent to tab", message);
      }
    }

    _launch() {
      const mode = this.config.http_server_mode_type;
      if (mode === "raw_text_plain" || mode === "raw_text_html") {
        this._is_raw_text_mode = true;
        this._is_interactive_mode = false;
        this._raw_mode_type = mode;
        this.willSendRawText();
      }
      if (mode === "interactive") {
        this._is_raw_text_mode = false;
        this._is_interactive_mode = true;
        this._setupInteractiveMode();
      }
    }

    _loadConfig(config) {
      this.config = config;
      this._postSetupConstructor();
      this._launch();
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
      switch (input.char) {
        default:
          this._triggerKeyPress(input);
      }
    }

    _handleInputBoxContent(input) {
      let input_box = document.querySelectorAll(
        `[data-browsh-id="${input.id}"]`
      )[0];
      if (input_box) {
        input_box.focus();
        if (input_box.getAttribute("role") == "textbox") {
          input_box.innerHTML = input.text;
        } else {
          input_box.value = input.text;
        }
      }
    }

    // TODO: Dragndrop doesn't seem to work :/
    _handleMouse(input) {
      switch (input.button) {
        case 1:
          this._mouseAction("mousemove", input.mouse_x, input.mouse_y);
          if (!this._mousedown) {
            this._mouseAction("mousedown", input.mouse_x, input.mouse_y);
            setTimeout(() => {
              this.sendSmallTextFrame();
            }, 500);
          }
          this._mousedown = true;
          break;
        case 0:
          this._mouseAction("mousemove", input.mouse_x, input.mouse_y);
          if (this._mousedown) {
            this._mouseAction("click", input.mouse_x, input.mouse_y);
            this._mouseAction("mouseup", input.mouse_x, input.mouse_y);
          }
          this._mousedown = false;
          break;
      }
    }

    _handleTTYSize(x, y) {
      if (!this._is_first_frame_finished) {
        this.dimensions.tty.width = parseInt(x);
        this.dimensions.tty.height = parseInt(y);
        this.dimensions.update();
        this.sendAllBigFrames();
      }
    }

    _handleScroll(x, y) {
      this.dimensions.frame.x_scroll = parseInt(x);
      this.dimensions.frame.y_scroll = parseInt(y);
      this.dimensions.update();
      window.scrollTo(
        this.dimensions.frame.x_scroll / this.dimensions.scale_factor.width,
        this.dimensions.frame.y_scroll / this.dimensions.scale_factor.height
      );
      this._mightSendBigFrames();
    }

    _triggerKeyPress(key) {
      let el = document.activeElement;
      const key_object = {
        key: key.char,
        keyCode: key.key
      };
      let event_press = new KeyboardEvent("keypress", key_object);
      let event_down = new KeyboardEvent("keydown", key_object);
      let event_up = new KeyboardEvent("keyup", key_object);
      // Generally sending down/up serves more use cases. But default input forms
      // don't listen for down/up to make the form submit. So this makes the assumption
      // that it's okay to send ENTER twice to an input box without any serious side
      // effects.
      if (key.key === 13 && el.tagName === "INPUT") {
        el.dispatchEvent(event_press);
      } else {
        el.dispatchEvent(event_down);
        el.dispatchEvent(event_up);
      }
    }

    _mouseAction(type, x, y) {
      const [dom_x, dom_y] = this._getDOMCoordsFromMouseCoords(x, y);
      const element = document.elementFromPoint(
        dom_x - window.scrollX,
        dom_y - window.scrollY
      );
      element.focus();
      var clickEvent = document.createEvent("MouseEvents");
      clickEvent.initMouseEvent(
        type,
        true,
        true,
        window,
        0,
        0,
        0,
        dom_x,
        dom_y,
        false,
        false,
        false,
        false,
        0,
        null
      );
      element.dispatchEvent(clickEvent);
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
      const index = y * this.dimensions.frame.width + x;
      if (this.text_builder.tty_grid.cells[index] !== undefined) {
        char = this.text_builder.tty_grid.cells[index].rune;
      } else {
        char = false;
      }
      if (!char || char === "▄") {
        dom_x = x * this.dimensions.char.width;
        dom_y = y * this.dimensions.char.height;
      } else {
        // Recall that text can be shifted from its original position in the browser in order
        // to snap it consistently to the TTY grid.
        original_position = this.text_builder.tty_grid.cells[index].dom_coords;
        dom_x = original_position.x;
        dom_y = original_position.y;
      }
      return [
        dom_x + this.dimensions.char.width / 2,
        dom_y + this.dimensions.char.height / 2
      ];
    }

    _sendTabInfo() {
      const title_object = document.getElementsByTagName("title");
      let info = {
        url: document.location.href,
        title: title_object.length ? title_object[0].innerHTML : ""
      };
      this.sendMessage(`/tab_info,${JSON.stringify(info)}`);
    }

    _mightSendBigFrames() {
      if (this._is_raw_text_mode) {
        return;
      }
      const y_diff =
        this.dimensions.frame.y_last_big_frame - this.dimensions.frame.y_scroll;
      const max_y_scroll_without_new_big_frame =
        (this.dimensions._big_sub_frame_factor - 1) *
        this.dimensions.tty.height;
      if (Math.abs(y_diff) > max_y_scroll_without_new_big_frame) {
        this.log(
          `Parsing big frames: ` +
            `previous-y: ${this.dimensions.frame.y_last_big_frame}, ` +
            `y-scroll: ${this.dimensions.frame.y_scroll}, ` +
            `diff: ${y_diff}, ` +
            `max-scroll: ${max_y_scroll_without_new_big_frame} `
        );
        this.sendAllBigFrames();
      }
    }
  };
