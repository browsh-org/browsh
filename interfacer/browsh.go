package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"encoding/json"

	"github.com/gorilla/websocket"

	// Termbox seems to be one of the best projects in any language for handling terminal input.
	// It's cross-platform and the maintainer is disciplined about supporting the baseline of escape
	// codes that work across the majority of terminals.
	"github.com/nsf/termbox-go"
)

var (
	logfile           string
	websocketAddresss = flag.String("addr", ":3334", "Web socket service address")
	upgrader          = websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	stdinChannel = make(chan string)
)

func setupLogging() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	logfile = fmt.Sprintf(filepath.Join(dir, "debug.log"))
	if _, err := os.Stat(logfile); err == nil {
		os.Truncate(logfile, 0)
	}
	if err != nil {
		panic(err)
	}
}

func log(msg string) {
	f, oErr := os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if oErr != nil {
		panic(oErr)
	}
	defer f.Close()

	msg = msg + "\n"
	if _, wErr := f.WriteString(msg); wErr != nil {
		panic(wErr)
	}
}

func initialise() {
	setupTermbox()
	setupLogging()
}

func setupTermbox() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)
}

func sendTtySize() {
	x, y := termbox.Size()
	stdinChannel <- fmt.Sprintf("/tty_size,%d,%d", x, y)
}

func readStdin() {
	defer termbox.Close()
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyCtrlQ {
				termbox.Close()
				os.Exit(0)
			}
			log(fmt.Sprintf("EventKey: k: %d, c: %c, mod: %s", ev.Key, ev.Ch, ev.Mod))
			eventMap := map[string]interface{}{
				"key": int(ev.Key),
				"char": string(ev.Ch),
				"mod": int(ev.Mod),
			}
			marshalled, _ := json.Marshal(eventMap)
			stdinChannel <- "/stdin," + string(marshalled)
		case termbox.EventResize:
			sendTtySize()
		case termbox.EventMouse:
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

func socketReader(ws *websocket.Conn) {
	defer ws.Close()
	for {
		_, message, err := ws.ReadMessage()
		parts := strings.Split(string(message), ",")
		command := parts[0]
		if command == "/frame" {
			termbox.SetCursor(0, 0)
			os.Stdout.Write([]byte(strings.Join(parts[1:], ",")))
			termbox.Flush()
		} else {
			log("WEBEXT: " + string(message))
		}
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log("Socket reader detected that the browser closed the websocket")
				triggerSocketWriterClose()
				return
			}
			panic(err)
		}
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

func socketWriter(ws *websocket.Conn) {
	var message string
	defer ws.Close()
	for {
		message = <-stdinChannel
		log(fmt.Sprintf("TTY sending: %s", message))
		if err := ws.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			if err == websocket.ErrCloseSent {
				log("Socket writer detected that the browser closed the websocket")
				return
			}
			panic(err)
		}
	}
}

func socketServer(w http.ResponseWriter, r *http.Request) {
	log("Incoming web request from browser")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	go socketWriter(ws)
	go socketReader(ws)

	sendTtySize()
}

func main() {
	initialise()
	log("Starting Browsh CLI client")
	go readStdin()
	http.HandleFunc("/", socketServer)
	if err := http.ListenAndServe(*websocketAddresss, nil); err != nil {
		panic(err)
	}
	log("Exiting at end of main()")
}
