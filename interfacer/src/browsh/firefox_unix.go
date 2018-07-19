// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package browsh

import (
	"strings"

	"github.com/go-errors/errors"
)

func getFirefoxPath() string {
	return Shell("which firefox")
}

func ensureFirefoxVersion(path string) {
	output := Shell(path + " --version")
	pieces := strings.Split(output, " ")
	version := pieces[len(pieces)-1]
	if versionOrdinal(version) < versionOrdinal("57") {
		message := "Installed Firefox version " + version + " is too old. " +
			"Firefox 57 or newer is needed."
		Shutdown(errors.New(message))
	}
}
