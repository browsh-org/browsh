package main

import (
	"strings"
	"testing"
	"time"
	"strconv"
	"fmt"

	"github.com/gdamore/tcell"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"browsh"
)

var simScreen tcell.SimulationScreen
var startupWait = 10
var browserFingerprint = " ← | x | "
var rootDir = browsh.Shell("git rev-parse --show-toplevel")

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration tests")
}

func GetFrameText() string {
	var text string
	cells, _, _ := simScreen.GetContents()
	for _, element := range cells {
		text += string(element.Runes)
	}
	fmt.Println(text)
	return text
}

var _ = Describe("Integration", func() {
	BeforeEach(func() {
		var count = 0
		simScreen = tcell.NewSimulationScreen("UTF-8")
		go browsh.Start(simScreen)
		for {
			if count > startupWait {
				var message = "Couldn't find browsh " +
					"startup signature within " +
					strconv.Itoa(startupWait) +
					" seconds"
				panic(message)
			}
			time.Sleep(time.Second)
			if (strings.Contains(GetFrameText(), browserFingerprint)) {
				break
			}
			count++
		}
	})

	AfterEach(func() {
		browsh.Shell(rootDir + "/webext/contrib/firefoxheadless.sh kill")
	})

	Describe("Showing a basic webpage", func() {
		It("have the right text", func() {
			Expect(GetFrameText()).To(ContainSubstring("Something"))
		})
	})
})
