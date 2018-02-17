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
	"strconv"
	"time"
	"math/rand"
	"unicode"

	// TCell seems to be one of the best projects in any language for handling terminal
	// standards across the major OSs.
	"github.com/gdamore/tcell"

	"github.com/go-errors/errors"
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
	startupURL           = flag.String("startup-url", "https://google.com", "URL to launch at startup")
	timeLimit            = flag.Int("time-limit", 0, "Kill Browsh after the specified number of seconds")
	upgrader             = websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	stdinChannel   = make(chan string)
	marionette     net.Conn
	ffCommandCount = 0
	isConnectedToWebExtension = false
	screen tcell.Screen
	defaultFFPrefs = map[string]string{
		"browser.startup.homepage":                "'https://www.google.com'",
		"startup.homepage_welcome_url":            "'https://www.google.com'",
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
		shutdown(err)
	}
	logfile = fmt.Sprintf(filepath.Join(dir, "debug.log"))
	if _, err := os.Stat(logfile); err == nil {
		os.Truncate(logfile, 0)
	}
	if err != nil {
		shutdown(err)
	}
}

func log(msg string) {
	if !*isDebug {
		return
	}
	f, oErr := os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if oErr != nil {
		shutdown(oErr)
	}
	defer f.Close()

	msg = msg + "\n"
	if _, wErr := f.WriteString(msg); wErr != nil {
		shutdown(wErr)
	}
}

// Write a simple text string to the screen. Not for use in the browser frames
// themselves. If you want anything to appear in the browser then must be done
// through the webextension.
func writeString(x, y int, str string) {
	var defaultColours = tcell.StyleDefault
	rgb := tcell.NewHexColor(int32(rand.Int() & 0xffffff))
	defaultColours.Foreground(rgb)
	defaultColours.Background(tcell.ColorRed)
	for _, c := range str {
		screen.SetContent(x, y, c, nil, defaultColours)
		x++
	}
}

func initialise() {
	flag.Parse()
	setupTcell()
	setupLogging()
}

func setupTcell() {
	var err error
	if err = screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	screen.EnableMouse()
	screen.Clear()
}

func shutdown(err error) {
	exitCode := 0
	screen.Fini()
	if err.Error() != "normal" {
		exitCode = 1
		println(err.Error())
	}
	out := err.(*errors.Error).ErrorStack()
	log(fmt.Sprintf(out))
	os.Exit(exitCode)
}

func sendTtySize() {
	x, y := screen.Size()
	sendMessageToWebExtension(fmt.Sprintf("/tty_size,%d,%d", x, y))
}

func readStdin() {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlQ {
				if !*isUseExistingFirefox {
					quitFirefox()
				}
				shutdown(errors.New("normal"))
			}
			eventMap := map[string]interface{}{
				"key":  int(ev.Key()),
				"char": string(ev.Rune()),
				"mod":  int(ev.Modifiers()),
			}
			marshalled, _ := json.Marshal(eventMap)
			sendMessageToWebExtension("/stdin," + string(marshalled))
		case *tcell.EventResize:
			screen.Sync()
			sendTtySize()
		case *tcell.EventMouse:
			x, y := ev.Position()
			button := ev.Buttons()
			eventMap := map[string]interface{}{
				"button":    int(button),
				"mouse_x":   int(x),
				"mouse_y":   int(y),
				"modifiers": int(ev.Modifiers()),
			}
			marshalled, _ := json.Marshal(eventMap)
			sendMessageToWebExtension("/stdin," + string(marshalled))
		}
	}
}

func sendMessageToWebExtension(message string) {
	if (!isConnectedToWebExtension) {
		log("Webextension not connected. Message not sent: " + message)
		return
	}
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
			shutdown(err)
		}
	}
}

func handleWebextensionCommand(message []byte) {
	parts := strings.Split(string(message), ",")
	command := parts[0]
	switch command {
	case "/frame":
		frame := parseJSONframe(strings.Join(parts[1:], ","))
		renderFrame(frame)
	case "/screenshot":
		saveScreenshot(parts[1])
	default:
		log("WEBEXT: " + string(message))
	}
}

// Frames received from the webextension are 1 dimensional arrays of strings.
// They are made up of a repeating pattern of 7 items:
// ["FG RED", "FG GREEN", "FG BLUE", "BG RED", "BG GREEN", "BG BLUE", "CHARACTER" ...]
func parseJSONframe(jsonString string) []string {
	var frame []string
	jsonBytes := []byte(jsonString)
	if err := json.Unmarshal(jsonBytes, &frame); err != nil {
		shutdown(err)
	}
	return frame
}

// Tcell uses a buffer to collect screen updates on, it only actually sends
// ANSI rendering commands to the terminal when we tell it to. And even then it
// will try to minimise rendering commands by only rendering parts of the terminal
// that have changed.
func renderFrame(frame []string) {
	var styling = tcell.StyleDefault
	var character string
	var runeChars []rune
	width, height := screen.Size()
	if (width * height * 7 != len(frame)) {
		log("Not rendering frame: current frame is not the same size as the screen")
		log(fmt.Sprintf("screen: %d, frame: %d", width * height * 7, len(frame)))
		return
	}
	index := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			styling = styling.Foreground(getRGBColor(frame, index))
			index += 3
			styling = styling.Background(getRGBColor(frame, index))
			index += 3
			character = frame[index]
			runeChars = []rune(character)
			index++
			if (character == "WIDE") {
				continue
			}
			screen.SetCell(x, y, styling, runeChars[0])
		}
	}
	screen.Show()
}

func getRGBColor(frame []string, index int) tcell.Color {
	rgb := frame[index:index + 3]
	return tcell.NewRGBColor(
			toInt32(rgb[0]),
			toInt32(rgb[1]),
			toInt32(rgb[2]))
}

func toInt32(char string) int32 {
	i, err := strconv.ParseInt(char, 10, 32)
	if err != nil {
		shutdown(err)
	}
	return int32(i)
}

func saveScreenshot(base64String string) {
	dec, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		shutdown(err)
	}
	file, err := ioutil.TempFile(os.TempDir(), "browsh-screenshot")
	if err != nil {
		shutdown(err)
	}
	if _, err := file.Write(dec); err != nil {
		shutdown(err)
	}
	if err := file.Sync(); err != nil {
		shutdown(err)
	}
	fullPath := file.Name() + ".jpg"
	if err := os.Rename(file.Name(), fullPath); err != nil {
		shutdown(err)
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
			shutdown(err)
		}
	}
}

func webSocketServer(w http.ResponseWriter, r *http.Request) {
	log("Incoming web request from browser")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		shutdown(err)
	}

	isConnectedToWebExtension = true

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

func stripWhitespace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func shell(command string) string {
	parts := strings.Fields(command)
	head := parts[0]
	parts = parts[1:len(parts)]
	out, err := exec.Command(head, parts...).Output()
	if err != nil {
		return "firefox not found"
	}
	return stripWhitespace(string(out))
}

func startHeadlessFirefox() {
	log("Starting Firefox in headless mode")
	firefoxPath := shell("which " + *firefoxBinary)
	if _, err := os.Stat(firefoxPath); os.IsNotExist(err) {
		shutdown(errors.New("Firefox command not found: " + *firefoxBinary))
	}
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
		shutdown(err)
	}
	if err := firefoxProcess.Start(); err != nil {
		shutdown(err)
	}
	in := bufio.NewScanner(stdout)
	for in.Scan() {
		log("FF-CONSOLE: " + in.Text())
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
		shutdown(err)
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
		shutdown(err)
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
		shutdown(err)
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
		"url": *startupURL,
	}
	sendFirefoxCommand("get", args)
}

func setDefaultPreferences() {
	for key, value := range defaultFFPrefs {
		setFFPreference(key, value)
	}
}

func beginTimeLimit() {
	warningLength := 10
	warningLimit := time.Duration(*timeLimit - warningLength);
	time.Sleep(warningLimit * time.Second)
	message := fmt.Sprintf("Browsh will close in %d seconds...", warningLength)
	sendMessageToWebExtension("/status," + message)
	time.Sleep(time.Duration(warningLength) * time.Second)
	quitFirefox()
	shutdown(errors.New("normal"))
}

// Note that everything executed in and from this function is not covered by the integration
// tests, because it uses the officially signed webextension, of which there can be only one.
// We can't bump the version and create a new signed webextension for every commit.
func setupFirefox() {
	go startHeadlessFirefox()
	if (*timeLimit > 0) {
		go beginTimeLimit()
	}
	// TODO: Do something better than just waiting
	time.Sleep(3 * time.Second)
	firefoxMarionette()
	setDefaultPreferences()
	installWebextension()
	go loadHomePage()
}

func quitFirefox() {
	sendFirefoxCommand("quitApplication", map[string]interface{}{})
}

func start(injectedScreen tcell.Screen) {
	screen = injectedScreen
	initialise()
	if !*isUseExistingFirefox {
		writeString(0, 0, "Starting Browsh, the modern terminal web browser...")
		setupFirefox()
	} else {
		writeString(0, 0, "Waiting for a Firefox instance to connect...")
	}
	log("Starting Browsh CLI client")
	go readStdin()
	http.HandleFunc("/", webSocketServer)
	if err := http.ListenAndServe(*webSocketAddresss, nil); err != nil {
		shutdown(err)
	}
	log("Exiting at end of main()")
}

func ttyStart() {
	// Hack to force true colours
	// Follow: https://github.com/gdamore/tcell/pull/183
	os.Setenv("TERM", "xterm-truecolor")

	realScreen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	start(realScreen)
}
