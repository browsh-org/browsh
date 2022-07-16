import utils from "utils";

// Handle commands from tabs, like sending a frame or information about
// the current character dimensions.
export default MixinBase =>
  class extends MixinBase {
    // TODO: There needs to be some consistency in this message sending protocol.
    //       Eg; always requiring JSON.
    handleTabMessage(message) {
      let incoming;
      const parts = message.split(",");
      const command = parts[0];
      switch (command) {
        case "/frame_text":
          this.sendToTerminal(`/frame_text,${message.slice(12)}`);
          break;
        case "/frame_pixels":
          this.sendToTerminal(`/frame_pixels,${message.slice(14)}`);
          break;
        case "/tab_info":
          incoming = JSON.parse(utils.rebuildArgsToSingleArg(parts));
          this._updateTabInfo(incoming);
          break;
        case "/dimensions":
          incoming = JSON.parse(message.slice(12));
          this.dimensions.setCharValues(incoming.char);
          break;
        case "/status":
          this.updateStatus(parts[1], parts[2]);
          break;
        case "/log":
          this.log(message.slice(5));
          break;
        case "/raw_text":
          incoming = JSON.parse(utils.rebuildArgsToSingleArg(parts));
          this._rawTextRequest(incoming);
          break;
        default:
          this.log("Unknown command from tab to background", message);
      }
    }

    _updateTabInfo(incoming) {
      this.title = incoming.title;
      this.url = incoming.url;
      this.sendStateToTerminal();
    }

    _rawTextRequest(incoming) {
      // I think the only reason that a tab would send a raw text payload is the
      // automatic startup URL loading, which should now be disabled for HTTP Server
      // mode.
      if (this.request_id) {
        let payload = {
          json: JSON.stringify(incoming),
          request_id: this.request_id
        };
        this.sendToTerminal(`/raw_text,${JSON.stringify(payload)}`);
      }
      this._tabCount(count => {
        if (count > 1) {
          this.remove();
        }
      });
    }

    _tabCount(callback) {
      this._getAllTabs(windowInfoArray => {
        callback(windowInfoArray[0].tabs.length);
      });
    }

    _getAllTabs(callback) {
      var getting = browser.windows.getAll({
        populate: true,
        windowTypes: ["normal"]
      });
      getting.then(
        windowInfoArray => callback(windowInfoArray),
        () => this.log("Error getting all tabs in Tab class")
      );
    }
  };
