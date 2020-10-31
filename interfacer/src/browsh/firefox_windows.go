// +build windows

package browsh

import (
	"fmt"
	"strings"

	"github.com/go-errors/errors"
	"golang.org/x/sys/windows/registry"
)

func getFirefoxPath() string {
	versionString := getWindowsFirefoxVersionString()
	flavor := getFirefoxFlavor()

	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`Software\Mozilla\`+flavor+` `+versionString+`\bin`,
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

func getWindowsFirefoxVersionString() string {
	flavor := getFirefoxFlavor()

	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`Software\Mozilla\`+flavor,
		registry.QUERY_VALUE)
	if err != nil {
		Shutdown(errors.New("Error reading Windows registry: " + fmt.Sprintf("%s", err)))
	}
	defer k.Close()

	versionString, _, err := k.GetStringValue("")
	if err != nil {
		Shutdown(errors.New("Error reading Windows registry: " + fmt.Sprintf("%s", err)))
	}

	Log("Windows registry Firefox version: " + versionString)

	return versionString
}

func getFirefoxFlavor() string {
	var flavor = "null"
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`Software\Mozilla\Mozilla Firefox`,
		registry.QUERY_VALUE)

	if err == nil {
		flavor = "Mozilla Firefox"
	}
	defer k.Close()

	if flavor == "null" {
		k, err := registry.OpenKey(
			registry.LOCAL_MACHINE,
			`Software\Mozilla\Firefox Developer Edition`,
			registry.QUERY_VALUE)

		if err == nil {
			flavor = "Firefox Developer Edition"
		}
		defer k.Close()
	}

	if flavor == "null" {
		k, err := registry.OpenKey(
			registry.LOCAL_MACHINE,
			`Software\Mozilla\Nightly`,
			registry.QUERY_VALUE)

		if err == nil {
			flavor = "Nightly"
		}
		defer k.Close()
	}

	if flavor == "null" {
		Shutdown(errors.New("Could not find Firefox on your registry"))
	}
	return flavor
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
