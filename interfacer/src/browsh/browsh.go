package browsh

import (
	"encoding/base64"
	"flag"
	"fmt"
	"strconv"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"path/filepath"
	"strings"
	"unicode"

	// TCell seems to be one of the best projects in any language for handling terminal
	// standards across the major OSs.
	"github.com/gdamore/tcell"

	"github.com/go-errors/errors"
	"github.com/shibukawa/configdir"
)

var (
	webSocketPort        = flag.String("websocket-port", "3334", "Web socket service address")
	firefoxBinary        = flag.String("firefox", "firefox", "Path to Firefox executable")
	isFFGui              = flag.Bool("with-gui", false, "Don't use headless Firefox")
	isUseExistingFirefox = flag.Bool("use-existing-ff", false, "Whether Browsh should launch Firefox or not")
	useFFProfile         = flag.String("ff-profile", "default", "Firefox profile to use")
	isDebug              = flag.Bool("debug", false, "Log to ./debug.log")
	StartupURL           = flag.String("startup-url", "https://google.com", "URL to launch at startup")
	timeLimit            = flag.Int("time-limit", 0, "Kill Browsh after the specified number of seconds")
	// IsHTTPServer needs to be exported for use in tests
	IsHTTPServer         = flag.Bool("http-server", false, "Run as an HTTP service")
	// HTTPServerPort also needs to be exported for use in tests
	HTTPServerPort       = flag.String("http-server-port", "4333", "HTTP server address")
	// IsTesting is used in tests, so it needs to be exported
	IsTesting            = false
	logfile              string
)

func setupLogging() {
	dir, err := os.Getwd()
	if err != nil {
		Shutdown(err)
	}
	logfile = fmt.Sprintf(filepath.Join(dir, "debug.log"))
	if _, err := os.Stat(logfile); err == nil {
		os.Truncate(logfile, 0)
	}
	if err != nil {
		Shutdown(err)
	}
}

// Log for general purpose logging
// TODO: accept generic types
func Log(msg string) {
	if !*isDebug {
		return
	}
	f, oErr := os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if oErr != nil {
		Shutdown(oErr)
	}
	defer f.Close()

	msg = msg + "\n"
	if _, wErr := f.WriteString(msg); wErr != nil {
		Shutdown(wErr)
	}
}

func initialise() {
	if IsTesting {
		*isDebug = true
	}
	setupLogging()
}

// Shutdown tries its best to cleanly shutdown browsh and the associated browser
func Shutdown(err error) {
	exitCode := 0
	if screen != nil {
		screen.Fini()
	}
	if err.Error() != "normal" {
		exitCode = 1
		println(err.Error())
	}
	if *isDebug {
		out := err.(*errors.Error).ErrorStack()
		Log(fmt.Sprintf(out))
	}
	os.Exit(exitCode)
}

func saveScreenshot(base64String string) {
	dec, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		Shutdown(err)
	}
	file, err := ioutil.TempFile(os.TempDir(), "browsh-screenshot")
	if err != nil {
		Shutdown(err)
	}
	if _, err := file.Write(dec); err != nil {
		Shutdown(err)
	}
	if err := file.Sync(); err != nil {
		Shutdown(err)
	}
	fullPath := file.Name() + ".jpg"
	if err := os.Rename(file.Name(), fullPath); err != nil {
		Shutdown(err)
	}
	message := "Screenshot saved to " + fullPath
	sendMessageToWebExtension("/status," + message)
	file.Close()
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

// Shell provides nice and easy shell commands
func Shell(command string) string {
	parts := strings.Fields(command)
	head := parts[0]
	parts = parts[1:len(parts)]
	out, err := exec.Command(head, parts...).CombinedOutput()
	if err != nil {
		fmt.Printf(
			"Browsh tried to run `%s` but failed with: %s", command, string(out))
		Shutdown(err)
	}
	return stripWhitespace(string(out))
}

// TTYStart starts Browsh
func TTYStart(injectedScreen tcell.Screen) {
	screen = injectedScreen
	initialise()
	setupTcell()
	writeString(0, 0, "Starting Browsh, the modern terminal web browser.", tcell.StyleDefault)
	startFirefox()
	Log("Starting Browsh CLI client")
	go readStdin()
	startWebSocketServer()
}

func toInt(char string) int {
	i, err := strconv.ParseInt(char, 10, 16)
	if err != nil {
		Shutdown(err)
	}
	return int(i)
}

func toInt32(char string) int32 {
	i, err := strconv.ParseInt(char, 10, 32)
	if err != nil {
		Shutdown(err)
	}
	return int32(i)
}

func ttyEntry() {
	// Hack to force true colours
	// Follow: https://github.com/gdamore/tcell/pull/183
	if runtime.GOOS != "windows" {
		// On windows this generates a "character set not supported" error. The error comes
		// from tcell.
		os.Setenv("TERM", "xterm-truecolor")
	}
	realScreen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	TTYStart(realScreen)
}

// MainEntry decides between running Browsh as a CLI app or as an HTTP web server
func MainEntry() {
	flag.Parse()
	if (*IsHTTPServer) {
		HTTPServerStart()
	} else {
		ttyEntry()
	}
}
