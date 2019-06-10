package browsh

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/go-errors/errors"
	"github.com/spf13/viper"
)

var (
	marionette     net.Conn
	ffCommandCount = 0
	defaultFFPrefs = map[string]string{
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

func startHeadlessFirefox() {
	Log("Starting Firefox in headless mode")
	checkIfFirefoxIsAlreadyRunning()
	firefoxPath := ensureFirefoxBinary()
	ensureFirefoxVersion(firefoxPath)
	args := []string{"--marionette"}
	if !viper.GetBool("firefox.with-gui") {
		args = append(args, "--headless")
	}
	profile := viper.GetString("firefox.profile")
	if profile != "browsh-default" {
		Log("Using profile: " + profile)
		args = append(args, "-P", profile)
	} else {
		profilePath := getFirefoxProfilePath()
		Log("Using default profile at: " + profilePath)
		args = append(args, "--profile", profilePath)
	}
	firefoxProcess := exec.Command(firefoxPath, args...)
	defer firefoxProcess.Process.Kill()
	stdout, err := firefoxProcess.StdoutPipe()
	if err != nil {
		Shutdown(err)
	}
	if err := firefoxProcess.Start(); err != nil {
		Shutdown(err)
	}
	in := bufio.NewScanner(stdout)
	for in.Scan() {
		Log("FF-CONSOLE: " + in.Text())
	}
}

func checkIfFirefoxIsAlreadyRunning() {
	if runtime.GOOS == "windows" {
		return
	}
	processes := Shell("ps aux")
	r, _ := regexp.Compile("firefox.*--headless")
	if r.MatchString(processes) {
		Shutdown(errors.New("A headless Firefox is already running"))
	}
}

func ensureFirefoxBinary() string {
	path := viper.GetString("firefox.path")
	if path == "firefox" {
		switch runtime.GOOS {
		case "windows":
			path = getFirefoxPath()
		case "darwin":
			path = "/Applications/Firefox.app/Contents/MacOS/firefox"
		default:
			path = getFirefoxPath()
		}
	}
	Log("Using Firefox at: " + path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		Shutdown(errors.New("Firefox binary not found: " + path))
	}
	return path
}

// Taken from https://stackoverflow.com/a/18411978/575773
func versionOrdinal(version string) string {
	// ISO/IEC 14651:2011
	const maxByte = 1<<8 - 1
	vo := make([]byte, 0, len(version)+8)
	j := -1
	for i := 0; i < len(version); i++ {
		b := version[i]
		if '0' > b || b > '9' {
			vo = append(vo, b)
			j = -1
			continue
		}
		if j == -1 {
			vo = append(vo, 0x00)
			j = len(vo) - 1
		}
		if vo[j] == 1 && vo[j+1] == '0' {
			vo[j+1] = b
			continue
		}
		if vo[j]+1 > maxByte {
			panic("VersionOrdinal: invalid version")
		}
		vo = append(vo, b)
		vo[j]++
	}
	return string(vo)
}

// Start Firefox via the `web-ext` CLI tool. This is for development and testing,
// because I haven't been able to recreate the way `web-ext` injects an unsigned
// extension.
func startWERFirefox() {
	Log("Attempting to start headless Firefox with `web-ext`")
	if IsConnectedToWebExtension {
		Shutdown(errors.New("There appears to already be an existing Web Extension connection"))
	}
	checkIfFirefoxIsAlreadyRunning()
	var rootDir = Shell("git rev-parse --show-toplevel")
	args := []string{
		"run",
		"--firefox=" + rootDir + "/webext/contrib/firefoxheadless.sh",
		"--verbose",
		"--no-reload",
	}
	firefoxProcess := exec.Command(rootDir+"/webext/node_modules/.bin/web-ext", args...)
	firefoxProcess.Dir = rootDir + "/webext/dist/"
	stdout, err := firefoxProcess.StdoutPipe()
	if err != nil {
		Shutdown(err)
	}
	if err := firefoxProcess.Start(); err != nil {
		Shutdown(err)
	}
	in := bufio.NewScanner(stdout)
	for in.Scan() {
		if strings.Contains(in.Text(), "Connected to the remote Firefox debugger") {
		}
		if strings.Contains(in.Text(), "JavaScript strict") ||
			strings.Contains(in.Text(), "D-BUS") ||
			strings.Contains(in.Text(), "dbus") {
			continue
		}
		Log("FF-CONSOLE: " + in.Text())
	}
	Log("WER Firefox unexpectedly closed")
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
	var (
		err  error
		conn net.Conn
	)
	connected := false
	Log("Attempting to connect to Firefox Marionette")
	start := time.Now()
	for time.Since(start) < 30*time.Second {
		conn, err = net.Dial("tcp", "127.0.0.1:2828")
		if err != nil {
			if !strings.Contains(err.Error(), "refused") {
				Shutdown(err)
			} else {
				time.Sleep(10 * time.Millisecond)
				continue
			}
		} else {
			connected = true
			break
		}
	}
	if !connected {
		Shutdown(errors.New("Failed to connect to Firefox's Marionette within 30 seconds"))
	}
	marionette = conn
	readMarionette()
	sendFirefoxCommand("WebDriver:NewSession", map[string]interface{}{})
}

// Install the Browsh extension that was bundled with `go-bindata` under
// `webextension.go`.
func installWebextension() {
	data, err := Asset("/browsh.xpi")
	if err != nil {
		Shutdown(err)
	}
	file, err := ioutil.TempFile(os.TempDir(), "browsh-webext-addon")
	defer os.Remove(file.Name())
	ioutil.WriteFile(file.Name(), []byte(data), 0644)
	args := map[string]interface{}{"path": file.Name()}
	sendFirefoxCommand("Addon:Install", args)
}

// Set a Firefox preference as you would in `about:config`
// `value` needs to be supplied with quotes if it's to be used as a JS string
func setFFPreference(key string, value string) {
	sendFirefoxCommand("Marionette:SetContext", map[string]interface{}{"value": "chrome"})
	script := fmt.Sprintf(`
		Components.utils.import("resource://gre/modules/Preferences.jsm");
		prefs = new Preferences({defaultBranch: false});
		prefs.set("%s", %s);`, key, value)
	args := map[string]interface{}{"script": script}
	sendFirefoxCommand("WebDriver:ExecuteScript", args)
	sendFirefoxCommand("Marionette:SetContext", map[string]interface{}{"value": "content"})
}

// Consume output from Marionette, we don't do anything with it. It"s just
// useful to have it in the logs.
func readMarionette() {
	buffer := make([]byte, 4096)
	count, err := marionette.Read(buffer)
	if err != nil {
		Shutdown(err)
	}
	Log("FF-MRNT: " + string(buffer[:count]))
}

func sendFirefoxCommand(command string, args map[string]interface{}) {
	Log("Sending `" + command + "` to Firefox Marionette")
	fullCommand := []interface{}{0, ffCommandCount, command, args}
	marshalled, _ := json.Marshal(fullCommand)
	message := fmt.Sprintf("%d:%s", len(marshalled), marshalled)
	fmt.Fprintf(marionette, message)
	ffCommandCount++
	readMarionette()
}

func setDefaultPreferences() {
	for key, value := range defaultFFPrefs {
		setFFPreference(key, value)
	}
}

func beginTimeLimit() {
	warningLength := 10
	warningLimit := time.Duration(*timeLimit - warningLength)
	time.Sleep(warningLimit * time.Second)
	message := fmt.Sprintf("Browsh will close in %d seconds...", warningLength)
	sendMessageToWebExtension("/status," + message)
	time.Sleep(time.Duration(warningLength) * time.Second)
	quitBrowsh()
}

// Note that everything executed in and from this function is not covered by the integration
// tests, because it uses the officially signed webextension, of which there can be only one.
// We can't bump the version and create a new signed webextension for every commit.
func setupFirefox() {
	go startHeadlessFirefox()
	if *timeLimit > 0 {
		go beginTimeLimit()
	}
	firefoxMarionette()
	setDefaultPreferences()
	installWebextension()
}

func StartFirefox() {
	if !viper.GetBool("firefox.use-existing") {
		writeString(0, 16, "Waiting for Firefox to connect...", tcell.StyleDefault)
		if IsTesting {
			writeString(0, 17, "TEST MODE", tcell.StyleDefault)
			go startWERFirefox()
		} else {
			setupFirefox()
		}
	} else {
		writeString(0, 16, "Waiting for a user-initiated Firefox instance to connect...", tcell.StyleDefault)
	}
}

func quitFirefox() {
	sendFirefoxCommand("Marionette:Quit", map[string]interface{}{})
}
