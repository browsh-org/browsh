package browsh

import (
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	// TCell seems to be one of the best projects in any language for handling terminal
	// standards across the major OSs.
	"github.com/gdamore/tcell"

	"github.com/go-errors/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	logo = `
////  ////
 / /   / /
 //    //
 //    //    ,,,,,,,,
 ////////  ..,,,,,,,,,
 //    //  .., ,,, .,.
 ////////  .., ,,,,,..
 ////////  ..,,,,,,,,,
 ////////    ...........
 //////////
 ****///////////////////
   ********///////////////
     ***********************`
	// IsTesting is used in tests, so it needs to be exported
	IsTesting        = false
	IsHTTPServerMode = false
	logfile          string
	_                = pflag.Bool("version", false, "Output current Browsh version")
)

func setupLogging() {
	out := io.Discard
	if *isDebug {
		dir, err := os.Getwd()
		if err != nil {
			Shutdown(err)
		}
		logfile = fmt.Sprintf(filepath.Join(dir, "debug.log"))
		if out, err = os.OpenFile(logfile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644); err != nil {
			Shutdown(err)
		}
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(out, nil)))
}

// Initialise browsh
func Initialise() {
	if IsTesting {
		*isDebug = true
	}
	setupLogging()
	loadConfig()
}

// Shutdown tries its best to cleanly shutdown browsh and the associated browser
func Shutdown(err error) {
	msg := "shutting down"
	var e *errors.Error
	if errors.As(err, &e) {
		slog.Error(msg, "errorStack", e.ErrorStack())
	} else {
		slog.Error(msg, "error", err)
	}
	if screen != nil {
		screen.Fini()
	}
	exitCode := 0
	if !errors.Is(err, errNormalExit) {
		exitCode = 1
	}
	os.Exit(exitCode)
}

func Log(message string) {
}

func saveScreenshot(base64String string) {
	dec, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		Shutdown(err)
	}
	file, err := os.CreateTemp("", "browsh-screenshot")
	if err != nil {
		Shutdown(err)
	}
	defer file.Close()
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
}

// Shell provides nice and easy shell commands
func Shell(command string) string {
	parts := strings.Fields(command)
	head := parts[0]
	parts = parts[1:]
	out, err := exec.Command(head, parts...).CombinedOutput()
	if err != nil {
		err := fmt.Errorf(
			"Browsh tried to run `%s` but failed with: %s, err: %w",
			command,
			string(out),
			err,
		)
		Shutdown(err)
	}
	return strings.TrimSpace(string(out))
}

// TTYStart starts Browsh
func TTYStart(injectedScreen tcell.Screen) {
	screen = injectedScreen
	setupTcell()
	writeString(1, 0, logo, tcell.StyleDefault)
	writeString(
		0,
		15,
		"Starting Browsh v"+browshVersion+", the modern text-based web browser.",
		tcell.StyleDefault,
	)
	StartFirefox()
	slog.Info("Starting Browsh CLI client")
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
	pflag.Parse()
	// validURL contains array of valid user inputted links.
	var validURL []string
	if pflag.NArg() != 0 {
		for i := 0; i < len(pflag.Args()); i++ {
			u, _ := url.ParseRequestURI(pflag.Args()[i])
			if u != nil {
				validURL = append(validURL, pflag.Args()[i])
			}
		}
	}
	viper.SetDefault("validURL", validURL)
	Initialise()

	// Print version if asked and exit
	if viper.GetBool("version") || viper.GetBool("v") {
		println(browshVersion)
		os.Exit(0)
	}

	// Print name if asked and exit
	if viper.GetBool("name") || viper.GetBool("n") {
		println("Browsh")
		os.Exit(0)
	}

	// Decide whether to run in http-server-mode or CLI app
	if viper.GetBool("http-server-mode") {
		HTTPServerStart()
	} else {
		ttyEntry()
	}
}
