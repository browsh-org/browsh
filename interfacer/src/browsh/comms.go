package browsh

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	stdinChannel              = make(chan string)
	isConnectedToWebExtension = false
)

type incomingRawText struct {
	RequestID string `json:"request_id"`
	RawJSON   string `json:"json"`
}

func startWebSocketServer() {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", webSocketServer)
	port := viper.GetString("browsh.websocket-port")
	if err := http.ListenAndServe(":"+port, serverMux); err != nil {
		Shutdown(err)
	}
}

func sendMessageToWebExtension(message string) {
	if !isConnectedToWebExtension {
		Log("Webextension not connected. Message not sent: " + message)
		return
	}
	stdinChannel <- message
}

// Listen to all messages coming from the webextension
// TODO: It seems this *also* receives sent to the webextention!?
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
	if viper.GetBool("http-server-mode") {
		handleRawFrameTextCommands(parts)
		return
	}
	switch command {
	case "/frame_text":
		parseJSONFrameText(strings.Join(parts[1:], ","))
		renderCurrentTabWindow()
	case "/frame_pixels":
		parseJSONFramePixels(strings.Join(parts[1:], ","))
		renderCurrentTabWindow()
	case "/tab_state":
		parseJSONTabState(strings.Join(parts[1:], ","))
		if CurrentTab != nil {
			renderUI()
		}
	case "/screenshot":
		saveScreenshot(parts[1])
	default:
		Log("WEBEXT: " + string(message))
	}
}

func handleRawFrameTextCommands(parts []string) {
	var incoming incomingRawText
	command := parts[0]
	if command == "/raw_text" {
		jsonBytes := []byte(strings.Join(parts[1:], ","))
		if err := json.Unmarshal(jsonBytes, &incoming); err != nil {
			Shutdown(err)
		}
		if incoming.RequestID != "" {
			Log("Raw text for " + incoming.RequestID)
			rawTextRequests.store(incoming.RequestID, incoming.RawJSON)
		} else {
			Log("Raw text but no associated request ID")
		}
	} else {
		Log("WEBEXT: " + strings.Join(parts[0:], ","))
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
	sendConfigToWebExtension()
	if !viper.GetBool("http-server-mode") {
		sendTtySize()
	}
	// For some reason, using Firefox's CLI arg `--url https://google.com` doesn't consistently
	// work. So we do it here instead.
	validURI := viper.GetStringSlice("validURI")
	if len(validURI) == 0 {
		sendMessageToWebExtension("/new_tab," + viper.GetString("startup-url"))
	} else {
		for i := 0; i < len(validURI); i++ {
			sendMessageToWebExtension("/new_tab," + validURI[i])
		}
	}
}

func sendConfigToWebExtension() {
	configJSON, _ := json.Marshal(viper.AllSettings())
	sendMessageToWebExtension("/config," + string(configJSON))
}
