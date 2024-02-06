//go:build windows

package browsh

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-errors/errors"
	"github.com/spf13/viper"
	"golang.org/x/sys/windows/registry"
)

func getFirefoxPath() string {
	versionString := getWindowsFirefoxVersionString()
	flavor := getFirefoxFlavor()

	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`Software\Mozilla\`+flavor+`\`+versionString+`\Main`,
		registry.QUERY_VALUE)
	if err != nil {
		Shutdown(fmt.Errorf("Error reading Windows registry: %w", err))
	}
	defer k.Close()

	path, _, err := k.GetStringValue("PathToExe")
	if err != nil {
		Shutdown(fmt.Errorf("Error reading Windows registry: %w", err))
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
		Shutdown(fmt.Errorf("Error reading Windows registry: %w", err))
	}
	defer k.Close()

	versionString, _, err := k.GetStringValue("CurrentVersion")
	if err != nil {
		Shutdown(fmt.Errorf("Error reading Windows registry: %w", err))
	}

	slog.Info("Windows registry Firefox", "version", versionString)

	return versionString
}

func getFirefoxFlavor() string {
	flavor := "null"
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
	if viper.GetBool("firefox.ignore-version") {
		return
	}
	versionString := getWindowsFirefoxVersionString()
	pieces := strings.Split(versionString, " ")
	version := pieces[0]
	if versionOrdinal(version) < versionOrdinal("57") {
		message := "Installed Firefox version " + version + " is too old. " +
			"Firefox 57 or newer is needed."
		Shutdown(errors.New(message))
	}
}
