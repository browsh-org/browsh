import utils from "utils";
import { Rect } from "vimium";
import { DomUtils } from "vimium";
import { LocalHints } from "vimium";
import { VimiumNormal } from "vimium";
import { MiscVimium } from "vimium";
MiscVimium();

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
          window.location.href = url;
          break;
        case "/url_up":
          this.urlUp();
          break;
        case "/url_root":
          window.location.href = window.location.origin;
          break;
        case "/history_back":
          history.go(-1);
          break;
        case "/history_forward":
          history.go(1);
          break;
        case "/reload":
          window.location.reload();
          break;
        case "/window_stop":
          window.stop();
          break;
        case "/find_next":
          this.findNext(parts[1]);
          break;
        case "/find_previous":
          window.find(parts[1], false, true, false, false, true, true);
          break;
        case "/get_link_hints":
          this.getLinkHints(false);
          break;
        case "/get_clickable_hints":
          this.getLinkHints(true);
          break;
        case "/focus_first_text_input":
          this.focusFirstTextInput();
          break;
        case "/follow_link_labeled_next":
          this._followLinkLabeledNext();
          break;
        case "/follow_link_labeled_previous":
          this._followLinkLabeledPrevious();
          break;
        default:
          this.log("Unknown command sent to tab", message);
      }
    }

    focusFirstTextInput() {
      VimiumNormal.focusInput(1);
    }

    //adapted vimium code
    followLinkLabeledNext() {
      var nextPatterns = "next,more,newer,>,›,→,»,≫,>>,weiter" || "";
      var nextStrings = nextPatterns.split(",").filter(function(s) {
        return s.trim().length;
      });
      return (
        VimiumNormal.findAndFollowRel("next") ||
        VimiumNormal.findAndFollowLink(nextStrings)
      );
    }

    _followLinkLabeledNext() {
      this.followLinkLabeledNext();
    }

    //adapted vimium code
    followLinkLabeledPrevious() {
      var previousPatterns =
        "prev,previous,back,older,<,‹,←,«,≪,<<,zurück" || "";
      var previousStrings = previousPatterns.split(",").filter(function(s) {
        return s.trim().length;
      });
      return (
        VimiumNormal.findAndFollowRel("prev") ||
        VimiumNormal.findAndFollowLink(previousStrings)
      );
    }

    _followLinkLabeledPrevious() {
      this.followLinkLabeledPrevious();
    }

    // Eg; This goes from www.domain.com/topic/suptopic/ to www.domain.com/topic/
    urlUp() {
      // this is taken from vimium's code
      var url = window.location.href;
      if (url[url.length - 1] === "/") {
        url = url.substring(0, url.length - 1);
      }
      var urlsplit = url.split("/");
      // make sure we haven't hit the base domain yet
      if (urlsplit.length > 3) {
        urlsplit = urlsplit.slice(0, Math.max(3, urlsplit.length - 1));
        window.location.href = urlsplit.join("/");
      }
    }

    getLinkHints(clickable) {
      var hints = LocalHints.getLocalHints(!clickable);
      var rect, bottom, top, left, right, width, height, results, result, href;
      results = [];
      for (let idx in hints) {
        if (!hints[idx].hasOwnProperty("rect")) {
          continue;
        }
        href = hints[idx]["href"];
        rect = hints[idx]["rect"];
        bottom = Math.round(
          ((rect["bottom"] - window.scrollY) *
            this.dimensions.scale_factor.height) /
            2
        );
        top = Math.round(
          ((rect["top"] - window.scrollY) *
            this.dimensions.scale_factor.height) /
            2
        );
        left = Math.round(rect["left"] * this.dimensions.scale_factor.width);
        right = Math.round(rect["right"] * this.dimensions.scale_factor.width);
        result = Rect.create(left, top, right, bottom);
        result.href = href;
        results.push(result);
      }
      this.sendMessage(`/link_hints,${JSON.stringify(results)}`);
    }

    findNext(text) {
      window.find(text, false, false, false, false, true, true);
      //var s = window.getSelection();
      //var oRange = s.getRangeAt(0); //get the text range
      //var oRect = oRange.getBoundingClientRect();
      //window.scrollTo(400, 20000);
      this.dimensions.y_scroll = Math.round(
        window.scrollY * this.dimensions.scale_factor.height
      );
      this.dimensions.x_scroll = Math.round(
        window.scrollX * this.dimensions.scale_factor.width
      );
      this.dimensions.update();
      this._mightSendBigFrames();
    }

    _launch() {
      const mode = this.config.http_server_mode_type;
      if (mode === "raw_text_plain" || mode === "raw_text_html") {
        this._is_raw_text_mode = true;
        this._is_interactive_mode = false;
        this._raw_mode_type = mode;
        this.sendRawText();
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
          var y_hack = false;
          if (input.hasOwnProperty("y_hack")) {
            y_hack = true;
          }
          this._mouseAction(
            "mousemove",
            input.mouse_x,
            input.mouse_y,
            0,
            y_hack
          );
          if (!this._mousedown) {
            this._mouseAction(
              "mousedown",
              input.mouse_x,
              input.mouse_y,
              0,
              y_hack
            );
            setTimeout(() => {
              this.sendSmallTextFrame();
            }, 500);
          }
          this._mousedown = true;
          break;
        case 0:
          var y_hack = false;
          if (input.hasOwnProperty("y_hack")) {
            y_hack = true;
          }
          this._mouseAction(
            "mousemove",
            input.mouse_x,
            input.mouse_y,
            0,
            y_hack
          );
          if (this._mousedown) {
            this._mouseAction("click", input.mouse_x, input.mouse_y, 0, y_hack);
            this._mouseAction(
              "mouseup",
              input.mouse_x,
              input.mouse_y,
              0,
              y_hack
            );
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
      if (el == null) {
        this.log(
          `Not pressing '${key.char}(${key.key})' as there is no active element`
        );
        return;
      }
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

    _mouseAction(type, x, y, button, y_hack = false) {
      let [dom_x, dom_y] = this._getDOMCoordsFromMouseCoords(x, y);
      if (y_hack) {
        const [dom_x2, dom_y2] = this._getDOMCoordsFromMouseCoords(x, y + 1);
        dom_y = (dom_y + dom_y2) / 2;
      }
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
        button,
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
