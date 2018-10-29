package browsh

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shibukawa/configdir"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	configFilename = "config.toml"

	isDebug   = pflag.Bool("debug", false, "Log to ./debug.log")
	timeLimit = pflag.Int("time-limit", 0, "Kill Browsh after the specified number of seconds")
	_         = pflag.Bool("http-server-mode", false, "Run as an HTTP service")

	_ = pflag.String("startup-url", "https://www.brow.sh", "URL to launch at startup")
	_ = pflag.String("firefox.path", "firefox", "Path to Firefox executable")
	_ = pflag.Bool("firefox.with-gui", false, "Don't use headless Firefox")
	_ = pflag.Bool("firefox.use-existing", false, "Whether Browsh should launch Firefox or not")
	_ = pflag.Bool("monochrome", false, "Start browsh in monochrome mode")
)

func getConfigNamespace() string {
	if IsTesting {
		return "browsh-testing"
	}
	return "browsh"
}

// Gets a cross-platform path to a folder containing Browsh config
func getConfigDir() string {
	marker := "browsh-settings"
	// configdir has no other option but to have a nested folder
	configDirs := configdir.New(getConfigNamespace(), marker)
	folders := configDirs.QueryFolders(configdir.Global)
	// Delete the previously enforced nested folder
	path := strings.Trim(folders[0].Path, marker)
	os.MkdirAll(path, os.ModePerm)
	ensureConfigFile(path)
	return path
}

// Copy the sample config file if the user doesn't already have a config file
func ensureConfigFile(path string) {
	fullPath := filepath.Join(path, configFilename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		file, err := os.Create(fullPath)
		if err != nil {
			Shutdown(err)
		}
		defer file.Close()
		_, err = file.WriteString(configSample)
		if err != nil {
			Shutdown(err)
		}
	}
}

// Gets a cross-platform path to store a Browsh-specific Firefox profile
func getFirefoxProfilePath() string {
	configDirs := configdir.New(getConfigNamespace(), "firefox_profile")
	folders := configDirs.QueryFolders(configdir.Global)
	folders[0].MkdirAll()
	return folders[0].Path
}

func setDefaults() {
	// Temporary experimental configurable keybindings
	viper.SetDefault("tty.keys.next-tab", []string{"\u001c", "28", "2"})

	// Vim commands
	vimCommandsBindings["gg"] = "scrollToTop"
	vimCommandsBindings["G"] = "scrollToBottom"
	vimCommandsBindings["j"] = "scrollDown"
	vimCommandsBindings["k"] = "scrollUp"
	vimCommandsBindings["h"] = "scrollLeft"
	vimCommandsBindings["l"] = "scrollRight"
	vimCommandsBindings["d"] = "scrollHalfPageDown"
	vimCommandsBindings["u"] = "scrollHalfPageUp"
	vimCommandsBindings["e"] = "editURL"
	vimCommandsBindings["ge"] = "editURL"
	vimCommandsBindings["H"] = "historyBack"
	vimCommandsBindings["L"] = "historyForward"
	vimCommandsBindings["J"] = "prevTab"
	vimCommandsBindings["K"] = "nextTab"
	vimCommandsBindings["r"] = "reload"
	vimCommandsBindings["xx"] = "removeTab"
	vimCommandsBindings["X"] = "restoreTab"
	vimCommandsBindings["t"] = "newTab"
	vimCommandsBindings["/"] = "findMode"
	vimCommandsBindings["n"] = "findNext"
	vimCommandsBindings["N"] = "findPrevious"
	vimCommandsBindings["g0"] = "firstTab"
	vimCommandsBindings["g$"] = "lastTab"
	vimCommandsBindings["gu"] = "urlUp"
	vimCommandsBindings["gU"] = "urlRoot"
	vimCommandsBindings["<<"] = "moveTabLeft"
	vimCommandsBindings[">>"] = "moveTabRight"
	vimCommandsBindings["^"] = "previouslyVisitedTab"
	vimCommandsBindings["m"] = "makeMark"
	vimCommandsBindings["'"] = "gotoMark"
	vimCommandsBindings["i"] = "insertMode"
	vimCommandsBindings["yy"] = "copyURL"
	vimCommandsBindings["p"] = "openClipboardURL"
	vimCommandsBindings["P"] = "openClipboardURLInNewTab"
	vimCommandsBindings["gi"] = "focusFirstTextInput"
	vimCommandsBindings["f"] = "openLinkInCurrentTab"
	vimCommandsBindings["F"] = "openLinkInNewTab"
	vimCommandsBindings["yf"] = "copyLinkURL"
	vimCommandsBindings["[["] = "followLinkLabeledPrevious"
	vimCommandsBindings["]]"] = "followLinkLabeledNext"
	vimCommandsBindings["yt"] = "duplicateTab"
	vimCommandsBindings["v"] = "visualMode"
	vimCommandsBindings["?"] = "viewHelp"
}

func loadConfig() {
	dir := getConfigDir()
	fullPath := filepath.Join(dir, configFilename)
	Log("Looking in " + fullPath + " for config.")
	viper.SetConfigType("toml")
	viper.SetConfigName(strings.Trim(configFilename, ".toml"))
	viper.AddConfigPath(dir)
	viper.AddConfigPath(".")
	setDefaults()
	// First load the sample config in case the user hasn't updated any new fields
	err := viper.ReadConfig(bytes.NewBuffer([]byte(configSample)))
	if err != nil {
		panic(fmt.Errorf("Config file error: %s \n", err))
	}
	// Then load the users own config file, overwriting the sample config
	err = viper.MergeInConfig()
	if err != nil {
		panic(fmt.Errorf("Config file error: %s \n", err))
	}
	viper.BindPFlags(pflag.CommandLine)
}
