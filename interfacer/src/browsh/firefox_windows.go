// +build windows

package browsh

import (
	"fmt"
	"strings"

	"github.com/go-errors/errors"
	"golang.org/x/sys/windows/registry"
)

const ERROR_RD_WIN_REG_MSG = "Error reading Windows registry: "

func getFirefoxPath() string {
	versionString := getWindowsFirefoxVersionString()

	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Mozilla\Mozilla Firefox\`+versionString+`\Main`,
		registry.QUERY_VALUE)
	if err != nil {
		Shutdown(errors.New(ERROR_RD_WIN_REG_MSG + fmt.Sprintf("%s", err)))
	}
	path, _, err := k.GetStringValue("PathToExe")
	if err != nil {
		Shutdown(errors.New(ERROR_RD_WIN_REG_MSG + fmt.Sprintf("%s", err)))
	}

	return path
}

func getWindowsFirefoxVersionString() string {
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Mozilla\Mozilla Firefox`,
		registry.QUERY_VALUE)
	if err != nil {
		Shutdown(errors.New(ERROR_RD_WIN_REG_MSG + fmt.Sprintf("%s", err)))
	}
	defer k.Close()

	versionString, _, err := k.GetStringValue("CurrentVersion")
	if err != nil {
		Shutdown(errors.New(ERROR_RD_WIN_REG_MSG + fmt.Sprintf("%s", err)))
	}

	Log("Windows registry Firefox version: " + versionString)

	return versionString
}

func ensureFirefoxVersion(path string) {
	versionString := getWindowsFirefoxVersionString()
	pieces := strings.Split(versionString, " ")
	version := pieces[0]
	if versionOrdinal(version) < versionOrdinal("57") {
		message := "Installed Firefox version " + version + " is too old. " +
			"Firefox 57 or newer is needed."
		Shutdown(errors.New(message))
	}
}
