package browsh

import (
	"strings"
	"fmt"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

func sendMessageToWebExtension(message string) {
	if (!isConnectedToWebExtension) {
		Log("Webextension not connected. Message not sent: " + message)
		return
	}
	stdinChannel <- message
}

// Listen to all messages coming from the webextension
func webSocketReader(ws *websocket.Conn) {
	defer ws.Close()
	for {
		_, message, err := ws.ReadMessage()
		handleWebextensionCommand(message)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				Log("Socket reader detected that the browser closed the websocket")
				triggerSocketWriterClose()
				return
			}
			Shutdown(err)
		}
	}
}

func handleWebextensionCommand(message []byte) {
	parts := strings.Split(string(message), ",")
	command := parts[0]
	switch command {
	case "/frame_text":
		frame.parseJSONFrameText(strings.Join(parts[1:], ","))
		renderUI()
		renderFrame()
	case "/frame_pixels":
		frame.parseJSONFramePixels(strings.Join(parts[1:], ","))
		renderUI()
		renderFrame()
	case "/state":
		oldState := map[string]string{}
		for k,v := range State{
      oldState[k] = v
    }
		parseJSONState(strings.Join(parts[1:], ","))
		handleStateChange(oldState)
	case "/screenshot":
		saveScreenshot(parts[1])
	default:
		Log("WEBEXT: " + string(message))
	}
}

func parseJSONState(jsonString string) {
	jsonBytes := []byte(jsonString)
	if err := json.Unmarshal(jsonBytes, &State); err != nil {
		Shutdown(err)
	}
}

func handleStateChange(oldState map[string]string) {
	if (State["page_state"] != oldState["page_state"]) {
		Log("State change: page_state=" + State["page_state"])
		if (State["page_state"] == "page_init") {
			yScroll = 0
		}
	}
	if (State["frame_width"] != "" && State["frame_height"] != "") {
		frame.width = toInt(State["frame_width"])
		frame.height = toInt(State["frame_height"])
	}
}

// When the socket reader attempts to read from a closed websocket it quickly and
// simply closes its associated Go routine. However the socket writer won't
// automatically notice until it actually needs to send something. So we force that
// by sending this NOOP text.
// TODO: There's a potential race condition because new connections share the same
//       Go channel. So we need to setup a new channel for every connection.
func triggerSocketWriterClose() {
	stdinChannel <- "BROWSH CLIENT FORCING CLOSE OF WEBSOCKET WRITER"
}

// Send a message to the webextension
func webSocketWriter(ws *websocket.Conn) {
	var message string
	defer ws.Close()
	for {
		message = <-stdinChannel
		Log(fmt.Sprintf("TTY sending: %s", message))
		if err := ws.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			if err == websocket.ErrCloseSent {
				Log("Socket writer detected that the browser closed the websocket")
				return
			}
			Shutdown(err)
		}
	}
}

func webSocketServer(w http.ResponseWriter, r *http.Request) {
	Log("Incoming web request from browser")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Shutdown(err)
	}

	isConnectedToWebExtension = true

	go webSocketWriter(ws)
	go webSocketReader(ws)

	sendTtySize()
}
