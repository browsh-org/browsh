package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	// Termbox seems to be one of the best projects in any language for handling terminal input.
	// It's cross-platform and the maintainer is disciplined about supporting the baseline of escape
	// codes that work across the majority of terminals.
	"github.com/nsf/termbox-go"

	"github.com/gorilla/websocket"
)

var (
	logfile              string
	webSocketAddresss    = flag.String("port", ":3334", "Web socket service address")
	firefoxBinary        = flag.String("firefox", "firefox", "Path to Firefox executable")
	isFFGui              = flag.Bool("with-gui", false, "Don't use headless Firefox")
	isUseExistingFirefox = flag.Bool("use-existing-ff", false, "Use an existing Firefox process")
	upgrader             = websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	stdinChannel   = make(chan string)
	marionette     net.Conn
	ffCommandCount = 0
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
	flag.Parse()
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
				if (!*isUseExistingFirefox) {
					sendFirefoxCommand("quitApplication", map[string]interface{}{})
				}
				termbox.Close()
				os.Exit(0)
			}
			log(fmt.Sprintf("EventKey: k: %d, c: %c, mod: %s", ev.Key, ev.Ch, ev.Mod))
			eventMap := map[string]interface{}{
				"key":  int(ev.Key),
				"char": string(ev.Ch),
				"mod":  int(ev.Mod),
			}
			marshalled, _ := json.Marshal(eventMap)
			stdinChannel <- "/stdin," + string(marshalled)
		case termbox.EventResize:
			// Need to flush STDOUT before getting the new TTY size because there
			// can be a discrepancy between the "internal buffer" size and the
			// actual size.
			termbox.Flush()
			sendTtySize()
		case termbox.EventMouse:
			log(fmt.Sprintf("Mouse: k: %d, x: %d, y: %d, mod: %s", ev.Key, ev.MouseX, ev.MouseY, ev.Mod))
			eventMap := map[string]interface{}{
				"key":     int(ev.Key),
				"mouse_x": int(ev.MouseX),
				"mouse_y": int(ev.MouseY),
				"mod":     int(ev.Mod),
			}
			marshalled, _ := json.Marshal(eventMap)
			stdinChannel <- "/stdin," + string(marshalled)
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

func webSocketReader(ws *websocket.Conn) {
	defer ws.Close()
	for {
		_, message, err := ws.ReadMessage()
		parts := strings.Split(string(message), ",")
		command := parts[0]
		if command == "/frame" {
			termbox.SetCursor(0, 0)
			os.Stdout.Write([]byte(parts[1]))
			termbox.HideCursor()
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

func webSocketWriter(ws *websocket.Conn) {
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

func webSocketServer(w http.ResponseWriter, r *http.Request) {
	log("Incoming web request from browser")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	go webSocketWriter(ws)
	go webSocketReader(ws)

	sendTtySize()
}

func startHeadlessFirefox() {
	println("Starting...")
	log("Starting Firefox in headless mode")
	args := []string{"--marionette", "--new-instance", "-P", "browsh2"}
	if !*isFFGui {
		args = append(args, "--headless")
	}
	firefoxProcess := exec.Command(*firefoxBinary, args...)
	defer firefoxProcess.Process.Kill()
	_, err := firefoxProcess.Output()
	if err != nil {
		panic(err)
	}
}

func firefoxMarionette() {
	log("Attempting to connect to Firefox Marionette")
	conn, err := net.Dial("tcp", "127.0.0.1:2828")
	marionette = conn
	readMarionette()
	if err != nil {
		panic(err)
	}
	sendFirefoxCommand("newSession", map[string]interface{}{})
}

func installWebextension() {
	data, err := Asset("webext/dist/web-ext-artifacts/browsh.xpi")
	if err != nil {
		panic(err)
	}
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(file.Name())
	ioutil.WriteFile(file.Name(), []byte(data), 0644)
	args := map[string]interface{}{ "path": file.Name() }
	sendFirefoxCommand("addon:install", args)
}

func readMarionette() {
	buffer := make([]byte, 4096)
	count, err := marionette.Read(buffer)
	if err != nil {
		if err != io.EOF {
			log(fmt.Sprintf("FF-MRNT: read error: %s", err))
		}
	}
	log("FF-MRNT: " + string(buffer[:count]))
}

func sendFirefoxCommand(command string, args map[string]interface{}) {
	log("Sending `" + command + "` to Firefox Marionette")
	fullCommand := []interface{}{0, ffCommandCount, command, args}
	marshalled, _ := json.Marshal(fullCommand)
	message := fmt.Sprintf("%d:%s", len(marshalled), marshalled)
	fmt.Fprintf(marionette, message)
	ffCommandCount++
	readMarionette()
}

func loadHomePage() {
	// Wait for the CLI websocket server to start listening
	time.Sleep(200 * time.Millisecond)
	args := map[string]interface{}{
		"url": "https://google.com",
	}
	sendFirefoxCommand("get", args)
}

func setupFirefox() {
	go startHeadlessFirefox()
	time.Sleep(2 * time.Second)
	firefoxMarionette()
	installWebextension()
	go loadHomePage()
}

func main() {
	initialise()
	if !*isUseExistingFirefox {
		setupFirefox()
	} else {
		println("Waiting for a Firefox instance to connect...")
	}
	log("Starting Browsh CLI client")
	go readStdin()
	http.HandleFunc("/", webSocketServer)
	if err := http.ListenAndServe(*webSocketAddresss, nil); err != nil {
		panic(err)
	}
	log("Exiting at end of main()")
}
