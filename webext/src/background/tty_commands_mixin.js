import utils from "utils";

// Handle commands coming in from the terminal, like; STDIN keystrokes, TTY resize, etc
export default MixinBase =>
  class extends MixinBase {
    handleTerminalMessage(message) {
      const parts = message.split(",");
      const command = parts[0];
      switch (command) {
        case "/config":
          this._loadConfig(message.slice(8));
          break;
        case "/tab_command":
          this.sendToCurrentTab(message.slice(13));
          break;
        case "/tty_size":
          this._updateTTYSize(parts[1], parts[2]);
          break;
        case "/stdin":
          this._handleUICommand(parts);
          this.sendToCurrentTab(message);
          break;
        case "/url_bar":
          this._handleURLBarInput(parts.slice(1).join(","));
          break;
        case "/new_tab":
          this.createNewTab(parts.slice(1).join(","));
          break;
        case "/switch_to_tab":
          this.switchToTab(parts.slice(1).join(","));
          break;
        case "/duplicate_tab":
          this.duplicateTab(parts.slice(1).join(","));
          break;
        case "/remove_tab":
          this.removeTab(parts.slice(1).join(","));
          break;
        case "/restore_tab":
          this.restoreTab();
          break;
        case "/raw_text_request":
          this._rawTextRequest(parts[1], parts[2], parts.slice(3).join(","));
          break;
      }
    }

    _loadConfig(json_string) {
      this.log(json_string);
      this.config = JSON.parse(json_string);
      this.config.browsh_version = browser.runtime.getManifest().version;
      if (this.currentTab()) {
        this.currentTab().sendGlobalConfig(this.config);
      }
      this.dimensions.postConfigSetup(this.config);
      this._setupRawTextMode();
    }

    _setupRawTextMode() {
      if (!this.config["http-server-mode"]) {
        return;
      }
      this._is_raw_text_mode = true;
      this._updateTTYSize(
        this.dimensions.raw_text_tty_size.width,
        this.dimensions.raw_text_tty_size.height
      );
    }

    _updateTTYSize(width, height) {
      this.dimensions.tty.width = parseInt(width);
      this.dimensions.tty.height = parseInt(height);
      if (this.currentTab()) {
        this.sendToCurrentTab(
          `/tty_size,${this.dimensions.tty.width},${this.dimensions.tty.height}`
        );
      }
      this.log(
        `Requesting browser resize for new TTY dimensions: ` +
          `${width}x${height}`
      );
      this.dimensions.resizeBrowserWindow();
    }

    _handleUICommand(parts) {
      const input = JSON.parse(utils.rebuildArgsToSingleArg(parts));
      // CTRL mappings
      /*
      if (input.mod === 2) {
        switch (input.char) {
          default:
        }
      }
      */
      // ALT mappings
      if (input.mod === 4) {
        switch (input.char) {
          case "p":
            this.screenshotActiveTab();
            break;
          case "u":
            this.toggleUserAgent();
            break;
        }
      }
      return false;
    }

    _handleURLBarInput(input) {
      const final_url = this._getURLfromUserInput(input);
      this.gotoURL(final_url);
    }

    // TODO: move to CLI client
    _getURLfromUserInput(input) {
      let url;
      const search_engine = this.config.default_search_engine_base;
      // Basically just check to see if there is text either side of a dot
      const is_straddled_dot = RegExp(/^[^\s]+\.[^\s]+/);
      // More comprehensive URL pattern
      const is_url = RegExp(/\/\/\w+(\.\w+)*(:[0-9]+)?\/?(\/[.\w]*)*$/);
      if (is_straddled_dot.test(input) || is_url.test(input)) {
        url = input;
        if (!url.startsWith("http")) {
          url = "http://" + url;
        }
      } else {
        url = `${search_engine}${input}`;
      }
      this.urlBarUserContent = url;
      return url;
    }

    createNewTab(url, callback) {
      const final_url = this._getURLfromUserInput(url);
      let creating = browser.tabs.create({
        url: final_url
      });
      creating.then(
        tab => {
          if (callback) {
            callback(tab);
          }
          this.log(`New tab created: ${tab}`);
        },
        error => {
          this.log(`Error creating new tab: ${error}`);
        }
      );
    }

    gotoURL(url) {
      let updating = browser.tabs.update(parseInt(this.currentTab().id), {
        url: url
      });
      updating.then(
        tab => {
          this.log(`Tab ${tab.id} loaded: ${url}`);
        },
        error => {
          this.log(`Error loading: ${url} \nError: ${error}`);
        }
      );
    }

    switchToTab(id) {
      let updating = browser.tabs.update(parseInt(id), {
        active: true
      });
      updating.then(
        tab => {
          this.log(`Switched to tab: ${tab.id}`);
        },
        error => {
          this.log(`Error switching to tab: ${error}`);
        }
      );
    }

    removeTab(id) {
      this.tabs[id].remove();
      this.tabs[id] = null;
    }

    duplicateTab(id) {
      browser.tabs.duplicate(parseInt(id));
    }

    restoreTab() {
      var sessionsInfo = browser.sessions.getRecentlyClosed({ maxResults: 1 });
      sessionsInfo.then(this._restoreTab);
    }

    _restoreTab(sessionsInfo) {
      var mySessionInfo = sessionsInfo[0];
      if (mySessionInfo.tab) {
        browser.sessions.restore(mySessionInfo.tab.sessionId);
      } else {
        browser.sessions.restore(mySessionInfo.window.sessionId);
      }
    }

    // We use the `browser` object here rather than going into the actual content script
    // because the content script may have crashed, even never loaded.
    screenshotActiveTab() {
      const capturing = browser.tabs.captureVisibleTab({
        format: "jpeg"
      });
      capturing.then(this._saveScreenshot.bind(this), error => this.log(error));
    }

    _saveScreenshot(imageUri) {
      const data = imageUri.replace(/^data:image\/\w+;base64,/, "");
      this.sendToTerminal("/screenshot," + data);
    }

    _rawTextRequest(request_id, mode, url) {
      this.createNewTab(url, native_tab => {
        this._acknowledgeNewTab({
          id: native_tab.id,
          request_id: request_id,
          raw_text_mode_type: mode.toLowerCase(),
          start_time: Date.now()
        });
        // Sometimes tabs fail to load for whatever reason. Make sure they get
        // removed to save RAM in long-lived Browsh HTTP servers
        setTimeout(() => {
          if (this.tabs[native_tab.id]) {
            this.removeTab(native_tab.id);
          }
        }, 60000);
      });
    }

    toggleUserAgent() {
      let message;
      this._is_using_mobile_user_agent = !this._is_using_mobile_user_agent;
      message = this._is_using_mobile_user_agent
        ? "Mobile user agent active"
        : "Desktop user agent active";
      this.currentTab().updateStatus("info", message);
    }

    _addUserAgentListener() {
      browser.webRequest.onBeforeSendHeaders.addListener(
        e => {
          if (this._is_using_mobile_user_agent) {
            e.requestHeaders.forEach(header => {
              if (header.name.toLowerCase() == "user-agent") {
                header.value = this.config.mobile_user_agent;
              }
            });
            return {
              requestHeaders: e.requestHeaders
            };
          }
        },
        {
          urls: ["*://*/*"]
        },
        ["blocking", "requestHeaders"]
      );
    }
  };
