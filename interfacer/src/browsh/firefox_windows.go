// +build windows

package browsh

import (
	"fmt"

	"github.com/go-errors/errors"
	"golang.org/x/sys/windows/registry"
)

func getFirefoxPath() string {
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Mozilla\Mozilla Firefox`,
		registry.QUERY_VALUE)
	if err != nil {
		Shutdown(errors.New("Error reading Windows registry: " + fmt.Sprintf("%s", err)))
	}
	defer k.Close()

	versionString, _, err := k.GetStringValue("CurrentVersion")
	if err != nil {
		Shutdown(errors.New("Error reading Windows registry: " + fmt.Sprintf("%s", err)))
	}

	Log("Windows registry Firefox version: " + versionString)

	k, err = registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Mozilla\Mozilla Firefox\`+versionString+`\Main`,
		registry.QUERY_VALUE)
	if err != nil {
		Shutdown(errors.New("Error reading Windows registry: " + fmt.Sprintf("%s", err)))
	}
	path, _, err := k.GetStringValue("PathToExe")
	if err != nil {
		Shutdown(errors.New("Error reading Windows registry: " + fmt.Sprintf("%s", err)))
	}

	return path
}

func ensureFirefoxVersion() {
}
