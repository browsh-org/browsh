package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/gdamore/tcell"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var simScreen tcell.SimulationScreen
var rootDir = shell("git rev-parse --show-toplevel")
var browserFingerprint = " ← | x | "

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration tests")
}

func StripWhitespace(str string) string {
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
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	return StripWhitespace(string(out))
}

func startWERFirefox() {
	args := []string{
		"run",
		"--firefox=" + rootDir + "/webext/contrib/firefoxheadless.sh",
		"--verbose",
		"--no-reload",
		"--url=http://www.something.com/",
	}
	firefoxProcess := exec.Command(rootDir+"/webext/node_modules/.bin/web-ext", args...)
	firefoxProcess.Dir = rootDir + "/webext/dist/"
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
		if strings.Contains(in.Text(), "JavaScript strict") ||
		   strings.Contains(in.Text(), "D-BUS") ||
		   strings.Contains(in.Text(), "dbus") {
			continue
		}
		log("FF-CONSOLE: " + in.Text())
	}
}

func GetFrameText() string {
	var text string
	cells, _, _ := simScreen.GetContents()
	for _, element := range cells {
		text += string(element.Runes)
	}
	return text
}

var _ = Describe("Integration", func() {
	BeforeEach(func() {
		var count = 0
		simScreen = tcell.NewSimulationScreen("UTF-8")
		go start(simScreen)
		go startWERFirefox()
		simScreen.SetSize(80, 30)
		for {
			if count > 10 {
				break
			}
			time.Sleep(1 * time.Second)
			if (strings.Contains(GetFrameText(), browserFingerprint)) {
				break
			}
			count++
		}
	})

	AfterEach(func() {
		shell(rootDir + "/webext/contrib/firefoxheadless.sh kill")
	})

	Describe("Showing a basic webpage", func() {
		It("have the right text", func() {
			Expect(GetFrameText()).To(ContainSubstring("Something"))
		})
	})
})
