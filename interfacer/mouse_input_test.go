package main

import (
	"github.com/tombh/termbox-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	"os"
)

func TestMouseInput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mouse Input Suite")
}

var _ = Describe("Mouse Input", func() {
	BeforeEach(func() {
		os.Setenv("DESKTOP_WIDTH", "1600")
		os.Setenv("DESKTOP_HEIGHT", "1200")
		os.Setenv("TTY_WIDTH", "90")
		os.Setenv("TTY_HEIGHT", "30")
		setupLogging()
		termbox.Init()
		setupDimensions()
		setCurrentDesktopCoords()
	})

	AfterEach(func() {
		termbox.Close()
	})

	Describe("Mouse position", func() {
		It("Should work in the top left", func() {
			curev.MouseX = 30
			curev.MouseY = 10
			setCurrentDesktopCoords()
			Expect(roundedDesktopX).To(Equal(533))
			Expect(roundedDesktopY).To(Equal(400))
		})
		It("Should work in the middle", func() {
			curev.MouseX = 45
			curev.MouseY = 15
			setCurrentDesktopCoords()
			Expect(roundedDesktopX).To(Equal(800))
			Expect(roundedDesktopY).To(Equal(600))
		})
		It("Should work in the bottom right", func() {
			curev.MouseX = 60
			curev.MouseY = 20
			setCurrentDesktopCoords()
			Expect(roundedDesktopX).To(Equal(1067))
			Expect(roundedDesktopY).To(Equal(800))
		})
	})

	Describe("Zooming", func() {
		BeforeEach(func() {
			curev.MouseX = 30
			curev.MouseY = 10
			setCurrentDesktopCoords()
		})
		It("Should zoom in once", func() {
			Expect(getXGrab()).To(Equal(0))
			Expect(getYGrab()).To(Equal(0))
			Expect(roundedDesktopX).To(Equal(533))
			Expect(roundedDesktopY).To(Equal(400))
			zoom("in")
			setCurrentDesktopCoords()
			Expect(getXGrab()).To(Equal(266))
			Expect(getYGrab()).To(Equal(200))
			Expect(roundedDesktopX).To(Equal(533))
			Expect(roundedDesktopY).To(Equal(400))
		})
		It("Should zoom in then out", func() {
			zoom("in")
			setCurrentDesktopCoords()
			zoom("out")
			// Shouldn't need to do this a second time, but this test helped me
			// figure out a different bug, so I'm leaving it like this for now.
			zoom("out")
			setCurrentDesktopCoords()
			Expect(getXGrab()).To(Equal(0))
			Expect(getYGrab()).To(Equal(0))
			Expect(roundedDesktopX).To(Equal(533))
			Expect(roundedDesktopY).To(Equal(400))
		})
		It("Should zoom near an edge without breaking out", func() {
			curev.MouseX = 0
			curev.MouseY = 0
			setCurrentDesktopCoords()
			zoom("in")
			Expect(getXGrab()).To(Equal(0))
			Expect(getYGrab()).To(Equal(0))
			Expect(roundedDesktopX).To(Equal(0))
			Expect(roundedDesktopY).To(Equal(0))
		})
	})
})
