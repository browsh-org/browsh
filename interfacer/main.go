package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	// Termbox seems to be one of the best projects in any language for handling terminal input.
	// It"s cross-platform and the maintainer is disciplined about supporting the baseline of escape
	// codes that work across the majority of terminals.
	"github.com/nsf/termbox-go"

	"github.com/gorilla/websocket"
	"github.com/shibukawa/configdir"

)

var (
	logfile              string
	webSocketAddresss    = flag.String("port", ":3334", "Web socket service address")
	firefoxBinary        = flag.String("firefox", "firefox", "Path to Firefox executable")
	isFFGui              = flag.Bool("with-gui", false, "Don't use headless Firefox")
	isUseExistingFirefox = flag.Bool("use-existing-ff", false, "Whether Browsh should launch Firefox or not")
	useFFProfile         = flag.String("ff-profile", "default", "Firefox profile to use")
	isDebug              = flag.Bool("debug", false, "Log to ./debug.log")
	upgrader             = websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	stdinChannel   = make(chan string)
	marionette     net.Conn
	ffCommandCount = 0
	defaultFFPrefs = map[string]string{
		"browser.startup.homepage":                "'about:blank'",
		"startup.homepage_welcome_url":            "'about:blank'",
		"startup.homepage_welcome_url.additional": "''",
		"devtools.errorconsole.enabled":           "true",
		"devtools.chrome.enabled":                 "true",

		// Send Browser Console (different from Devtools console) output to
		// STDOUT.
		"browser.dom.window.dump.enabled": "true",

		// From:
		// http://hg.mozilla.org/mozilla-central/file/1dd81c324ac7/build/automation.py.in//l388
		// Make url-classifier updates so rare that they won"t affect tests.
		"urlclassifier.updateinterval": "172800",
		// Point the url-classifier to a nonexistent local URL for fast failures.
		"browser.safebrowsing.provider.0.gethashURL": "'http://localhost/safebrowsing-dummy/gethash'",
		"browser.safebrowsing.provider.0.keyURL":     "'http://localhost/safebrowsing-dummy/newkey'",
		"browser.safebrowsing.provider.0.updateURL":  "'http://localhost/safebrowsing-dummy/update'",

		// Disable self repair/SHIELD
		"browser.selfsupport.url": "'https://localhost/selfrepair'",
		// Disable Reader Mode UI tour
		"browser.reader.detectedFirstArticle": "true",

		// Set the policy firstURL to an empty string to prevent
		// the privacy info page to be opened on every "web-ext run".
		// (See #1114 for rationale)
		"datareporting.policy.firstRunURL": "''",
	}
)

func setupLogging() {
	dir, err := os.Getwd()
	if err != nil {
		shutdown(err.Error())
	}
	logfile = fmt.Sprintf(filepath.Join(dir, "debug.log"))
	if _, err := os.Stat(logfile); err == nil {
		os.Truncate(logfile, 0)
	}
	if err != nil {
		shutdown(err.Error())
	}
}

func log(msg string) {
	if !*isDebug {
		return
	}
	f, oErr := os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if oErr != nil {
		shutdown(oErr.Error())
	}
	defer f.Close()

	msg = msg + "\n"
	if _, wErr := f.WriteString(msg); wErr != nil {
		shutdown(wErr.Error())
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
		shutdown(err.Error())
	}
	termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)
}

func shutdown(message string) {
	exitCode := 0
	if message != "normal" {
		exitCode = 1
	}
	println(message)
	log("Shutting down with: " + message)
	termbox.Close()
	os.Exit(exitCode)
}

func sendTtySize() {
	x, y := termbox.Size()
	sendMessageToWebExtension(fmt.Sprintf("/tty_size,%d,%d", x, y))
}

func readStdin() {
	defer termbox.Close()
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyCtrlQ {
				if !*isUseExistingFirefox {
					sendFirefoxCommand("quitApplication", map[string]interface{}{})
				}
				shutdown("normal")
			}
			log(fmt.Sprintf("EventKey: k: %d, c: %c, mod: %s", ev.Key, ev.Ch, ev.Mod))
			eventMap := map[string]interface{}{
				"key":  int(ev.Key),
				"char": string(ev.Ch),
				"mod":  int(ev.Mod),
			}
			marshalled, _ := json.Marshal(eventMap)
			sendMessageToWebExtension("/stdin," + string(marshalled))
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
			sendMessageToWebExtension("/stdin," + string(marshalled))
		case termbox.EventError:
			shutdown(ev.Err.Error())
		}
	}
}

func sendMessageToWebExtension(message string) {
	stdinChannel <- message
}

func webSocketReader(ws *websocket.Conn) {
	defer ws.Close()
	for {
		_, message, err := ws.ReadMessage()
		handleWebextensionCommand(message)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log("Socket reader detected that the browser closed the websocket")
				triggerSocketWriterClose()
				return
			}
			shutdown(err.Error())
		}
	}
}

func handleWebextensionCommand(message []byte) {
	parts := strings.Split(string(message), ",")
	command := parts[0]
	switch command {
	case "/frame":
		renderFrame(strings.Join(parts[1:], ","))
	case "/screenshot":
		saveScreenshot(parts[1])
	default:
		log("WEBEXT: " + string(message))
	}
}

func renderFrame(frame string) {
	termbox.SetCursor(0, 0)
	os.Stdout.Write([]byte(frame))
	termbox.HideCursor()
	termbox.Flush()
}

func saveScreenshot(base64String string) {
	dec, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		shutdown(err.Error())
	}
	file, err := ioutil.TempFile(os.TempDir(), "browsh-screenshot")
	if err != nil {
		shutdown(err.Error())
	}
	if _, err := file.Write(dec); err != nil {
		shutdown(err.Error())
	}
	if err := file.Sync(); err != nil {
		shutdown(err.Error())
	}
	fullPath := file.Name() + ".jpg"
	if err := os.Rename(file.Name(), fullPath); err != nil {
		shutdown(err.Error())
	}
	message := "Screenshot saved to " + fullPath
	sendMessageToWebExtension("/status," + message)
	file.Close()
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
			shutdown(err.Error())
		}
		log(fmt.Sprintf("TTY sent: %s", message))
	}
}

func webSocketServer(w http.ResponseWriter, r *http.Request) {
	log("Incoming web request from browser")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		shutdown(err.Error())
	}

	go webSocketWriter(ws)
	go webSocketReader(ws)

	sendTtySize()
}

// Gets a cross-platform path to store Browsh config
func getConfigFolder() string {
	configDirs := configdir.New("browsh", "firefox_profile")
	folders := configDirs.QueryFolders(configdir.Global)
	folders[0].MkdirAll()
	return folders[0].Path
}

func startHeadlessFirefox() {
	println("Starting...")
	log("Starting Firefox in headless mode")
	args := []string{"--marionette"}
	if !*isFFGui {
		args = append(args, "--headless")
	}
	if *useFFProfile != "default" {
		log("Using profile: " + *useFFProfile)
		args = append(args, "-P", *useFFProfile)
	} else {
		profilePath := getConfigFolder()
		log("Using default profile at: " + profilePath)
		args = append(args, "--profile", profilePath)
	}
	firefoxProcess := exec.Command(*firefoxBinary, args...)
	defer firefoxProcess.Process.Kill()
	stdout, err := firefoxProcess.StdoutPipe()
	if err != nil {
		shutdown(err.Error())
	}
	if err := firefoxProcess.Start(); err != nil {
		shutdown(err.Error())
	}
	in := bufio.NewScanner(stdout)
	for in.Scan() {
		log("FF-CONSOLE: " + in.Text()) // write each line to your log, or anything you need
	}
}

// Connect to Firefox's Marionette service.
// RANT: Firefox's remote control tools are so confusing. There seem to be 2
// services that come with your Firefox binary; Marionette and the Remote
// Debugger. The latter you would expect to follow the widely supported
// Chrome standard, but no, it's merely on the roadmap. There is very little
// documentation on either. I have the impression, but I'm not sure why, that
// the Remote Debugger is better, seemingly more API methods, and as mentioned
// is on the roadmap to follow the Chrome standard.
// I've used Marionette here, simply because it was easier to reverse engineer
// from the Python Marionette package.
func firefoxMarionette() {
	log("Attempting to connect to Firefox Marionette")
	conn, err := net.Dial("tcp", "127.0.0.1:2828")
	if err != nil {
		shutdown(err.Error())
	}
	marionette = conn
	readMarionette()
	sendFirefoxCommand("newSession", map[string]interface{}{})
}

// Install the Browsh extension that was bundled with `go-bindata` under
// `webextension.go`.
func installWebextension() {
	data, err := Asset("webext/dist/web-ext-artifacts/browsh.xpi")
	if err != nil {
		shutdown(err.Error())
	}
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(file.Name())
	ioutil.WriteFile(file.Name(), []byte(data), 0644)
	args := map[string]interface{}{"path": file.Name()}
	sendFirefoxCommand("addon:install", args)
}

// Set a Firefox preference as you would in `about:config`
// `value` needs to be supplied with quotes if it's to be used as a JS string
func setFFPreference(key string, value string) {
	sendFirefoxCommand("setContext", map[string]interface{}{"value": "chrome"})
	script := fmt.Sprintf(`
		Components.utils.import("resource://gre/modules/Preferences.jsm");
		prefs = new Preferences({defaultBranch: false});
		prefs.set("%s", %s);`, key, value)
	args := map[string]interface{}{"script": script}
	sendFirefoxCommand("executeScript", args)
	sendFirefoxCommand("setContext", map[string]interface{}{"value": "content"})
}

// Consume output from Marionette, we don't do anything with it. It"s just
// useful to have it in the logs.
func readMarionette() {
	buffer := make([]byte, 4096)
	count, err := marionette.Read(buffer)
	if err != nil {
		shutdown(err.Error())
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

func setDefaultPreferences() {
	for key, value := range defaultFFPrefs {
		setFFPreference(key, value)
	}
}

// Note that everything executed in and from this function is not covered by the integration
// tests, because it uses the officially signed webextension, of which there can be only one.
// We can't bump the version and create a new signed webextension for every commit.
func setupFirefox() {
	go startHeadlessFirefox()
	// TODO: Do something better than just waiting
	time.Sleep(3 * time.Second)
	firefoxMarionette()
	setDefaultPreferences()
	installWebextension()
	go loadHomePage()
}

func main() {
	initialise()
	if !*isUseExistingFirefox {
		println("Starting Browsh...")
		setupFirefox()
	} else {
		println("Waiting for a Firefox instance to connect...")
	}
	log("Starting Browsh CLI client")
	go readStdin()
	http.HandleFunc("/", webSocketServer)
	if err := http.ListenAndServe(*webSocketAddresss, nil); err != nil {
		shutdown(err.Error())
	}
	log("Exiting at end of main()")
}
