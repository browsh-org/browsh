package main

import (
	"flag"
	"fmt"
	"os"
	"net/http"
  "path/filepath"

	"github.com/gorilla/websocket"

	// Termbox seems to be one of the best projects in any language for handling terminal input.
	// It's cross-platform and the maintainer is disciplined about supporting the baseline of escape
	// codes that work across the majority of terminals.
  "github.com/nsf/termbox-go"
)


var (
	logfile string
	websocketAddresss = flag.String("addr", ":3334", "Web socket service address")
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
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
	setupLogging();
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
	log(fmt.Sprintf("%d,%d", x, y))
	stdinChannel <- fmt.Sprintf("/tty_size,%d,%d", x, y)
}


func readStdin() {
	var event string;
	defer termbox.Close()
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyCtrlQ {
				termbox.Close()
				os.Exit(0)
			}
			event = fmt.Sprintf("EventKey: k: %d, c: %c, mod: %s", ev.Key, ev.Ch, ev.Mod)
			stdinChannel <- event
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
		termbox.SetCursor(0, 0)
		os.Stdout.Write([]byte(message))
		termbox.Flush()
		if err != nil {
			panic(err)
		}
	}
}

func socketWriter(ws *websocket.Conn) {
	var message string
	defer ws.Close()
	for {
		message = <- stdinChannel
		log(fmt.Sprintf("sending ... %s", message))
		if err := ws.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			panic(err)
		}
	}
}

func socketServer(w http.ResponseWriter, r *http.Request) {
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
	go readStdin()
	http.HandleFunc("/", socketServer)
	if err := http.ListenAndServe(*websocketAddresss, nil); err != nil {
		panic(err)
	}
}
