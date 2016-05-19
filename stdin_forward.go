package main

import (
	"fmt"
	"strings"
	"time"
	"os"
	"os/exec"
	"math"
	"github.com/tombh/termbox-go"
)

// Import the xzoom C code that creates an X window that zooms
// and pans the desktop.
// It's written in C because it borrows from the original xzoom
// binary: http://git.r-36.net/xzoom/
// NB: The following comments are parsed by `go build` ...

// #cgo LDFLAGS: -lXext -lX11 -lXt
// #include "xzoom/xzoom.h"
import "C"

var logfile = "./input.log"
var current string
var curev termbox.Event
var lastMouseButton string
var desktopWidth float32 = 1600
var desktopHeight float32 = 1200
var desktopXFloat float32
var desktopYFloat float32
var roundedDesktopX int
var roundedDesktopY int
// Dimensions of hiptext output
var hipWidth int
var hipHeight int

// For keeping track of the zoom
// TODO: look at the XFCE code to accurately determine the factor. It may
// even be linear.
var zoomFactor float32 = 0.03
var maxZoom float32 = 1000000
var zoomLevel float32
var viewport map[string] float32

func initialise() {
	log("Starting...")
	calculateHipDimensions()
	zoomLevel = 1
	viewport = map[string] float32 {
		"xSize": desktopWidth,
		"ySize": desktopHeight,
		"xOffset": 0,
		"yOffset": 0,
	}
}

// Hiptext needs to render the aspect ratio faithfully. So firstly it tries to fill
// the terminal as much as it can. And secondly it treats a row as representing twice
// as much as a column - thus why there are some multiplications/divisions by 2.
func calculateHipDimensions() {
	_tw, _th := termbox.Size()
	tw := float32(_tw)
	th := float32(_th * 2)
	ratio := desktopWidth / desktopHeight
	bestHeight := min(th, (tw / ratio))
  bestWidth := min(tw, (bestHeight * ratio))
	// Not sure why the +1 and -1 are needed, but they are.
  hipWidth = roundToInt(bestWidth) + 1
  hipHeight = roundToInt(bestHeight / 2) - 1
	log(fmt.Sprintf("Term dimensions: W: %d, H: %d", _tw, _th))
	log(fmt.Sprintf("Hiptext dimensions: W: %d, H: %d", hipWidth, hipHeight))
}

func min(a float32, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func log(msg string) {
	msg = msg + "\n"
	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(msg); err != nil {
		panic(err)
	}
}

// Issue an xdotool command to simulate mouse and keyboard input
func xdotool(args ...string) {
	log(strings.Join(args, " "))
	if args[0] == "noop" {
		return
	}
	if err := exec.Command("xdotool", args...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func roundToInt(value32 float32) int {
	var rounded float64
	value := float64(value32)
	if value < 0 {
		rounded = math.Ceil(value - 0.5)
	}
	rounded = math.Floor(value + 0.5)
	return int(rounded)
}

// Whether the current input event includes a depressed CTRL key.
func ctrlPressed() bool {
	return curev.Mod&termbox.ModCtrl != 0
}

// Convert Termbox symbols to xdotool arguments
func mouseButtonStr(k termbox.Key) []string {
	switch k {
	case termbox.MouseLeft:
		lastMouseButton = "1"
		return []string{"mousedown", lastMouseButton}
	case termbox.MouseMiddle:
		lastMouseButton = "2"
		return []string{"mousedown", lastMouseButton}
	case termbox.MouseRight:
		lastMouseButton = "3"
		return []string{"mousedown", lastMouseButton}
	case termbox.MouseRelease:
		return []string{"mouseup", lastMouseButton}
	case termbox.MouseWheelUp:
		if ctrlPressed() {
			C.magx++
			C.magy++
			return []string{"noop"}
		}
		return []string{"click", "4"}
	case termbox.MouseWheelDown:
		if ctrlPressed() {
			C.magx--
			C.magy--
			return []string{"noop"}
		}
		return []string{"click", "5"}
	}
	return []string{""}
}

// Auxillary data. Whether the mouse was moving or a mod key like CTRL
// is being pressed at the same time.
func modStr(m termbox.Modifier) string {
	var out []string
	if m&termbox.ModAlt != 0 {
		out = append(out, "Alt")
	}
	if m&termbox.ModMotion != 0 {
		out = append(out, "Motion")
	}
	// Depends on this PR: https://github.com/nsf/termbox-go/pull/126
	if m&termbox.ModCtrl != 0 {
		out = append(out, "Ctrl")
	}

	return strings.Join(out, " ")
}

func mouseEvent() {
	log(
		fmt.Sprintf(
			"EventMouse: x: %d, y: %d, b: %s, mod: %s",
			curev.MouseX, curev.MouseY, mouseButtonStr(curev.Key), modStr(curev.Mod)))

	// CTRL allows the user to drag the mouse to pan and zoom the desktop.
	if !ctrlPressed() {
		xdotool("keyup", "alt")
	}

	// Always move the mouse first. This is because we're not constantly updating the mouse position,
	// *unless* a drag event is happening. This saves bandwidth. Also, mouse movement isn't supported
	// on all terminals.
	setCurrentDesktopCoords()
	xdotool("mousemove", fmt.Sprintf("%d", roundedDesktopX), fmt.Sprintf("%d", roundedDesktopY))

	// Send a button press to X. Note that the "Motion" modifier is sent when the user is doing
	// a drag event and thus mouse reporting will be constantly streamed.
	if !strings.Contains(modStr(curev.Mod), "Motion") {
		xdotool(mouseButtonStr(curev.Key)...)
	}
}

// Convert terminal coords into desktop coords
func setCurrentDesktopCoords() {
	hipWidthFloat := float32(hipWidth)
	hipHeightFloat := float32(hipHeight)
	eventX := float32(curev.MouseX)
	eventY := float32(curev.MouseY)
	desktopXFloat = (eventX * (viewport["xSize"] / hipWidthFloat)) + viewport["xOffset"]
	desktopYFloat = (eventY * (viewport["ySize"] / hipHeightFloat)) + viewport["yOffset"]
	log(
		fmt.Sprintf(
			"setCurrentDesktopCoords: tw: %d, th: %d, dx: %d, dy: %d",
			hipHeightFloat, hipWidthFloat, desktopXFloat, desktopYFloat))
	roundedDesktopX = roundToInt(desktopXFloat)
	roundedDesktopY = roundToInt(desktopYFloat)
}

// XFCE doesn't provide the current zoom, so *we* need to keep track of it.
// For every zoom level, the terminal coords will be mapped differently onto the X desktop.
// TODO: support custom desktop sizes.
func trackZoom(direction string) {
	xdotool("keydown", "alt")

	if direction == "in" {
		if zoomLevel <= maxZoom {
			zoomLevel += zoomFactor
		} else {
			return
		}
	} else {
		if zoomLevel >= 1 {
			zoomLevel -= zoomFactor
		} else {
			return
		}
	}
	// Use the existing viewport to get the current coords
	setCurrentDesktopCoords()

	// The actual zoom
	viewport["xSize"] = desktopWidth / zoomLevel
	viewport["ySize"] = desktopHeight / zoomLevel
	viewport["xOffset"] = desktopXFloat - (viewport["xSize"] / 2)
	viewport["yOffset"] = desktopYFloat - (viewport["ySize"] / 2)

	keepViewportInDesktop()

	log(fmt.Sprintf("zoom: %s", zoomLevel))
	log(fmt.Sprintf("viewport: %s", viewport))
}

// When zooming near the edges of the desktop it is possible that the viewport's edges overlap
// the desktop's edges. So just limit the possible movement of the viewport.
func keepViewportInDesktop() {
	xLeft   := viewport["xOffset"]
	xRight  := viewport["xOffset"] + viewport["xSize"]
	yTop    := viewport["yOffset"]
	yBottom := viewport["yOffset"] + viewport["ySize"]

	if xLeft < 0 {
		viewport["xOffset"] = 0
	}
	if xRight > desktopWidth {
		viewport["xOffset"] = desktopWidth - viewport["xSize"]
	}
	if yTop < 0 {
		viewport["yOffset"] = 0
	}
	if yBottom > desktopHeight {
		viewport["yOffset"] = desktopHeight - viewport["ySize"]
	}
}

// Convert a keyboard event into an xdotool command
func keyEvent() {
	// I've no idea why this gets picked up by the terminal, or what it refers to. But whatever, we don't
	// want it passed onto X.
	if fmt.Sprintf("%s", curev.Ch) == "%!s(int32=0)" {
		return
	}

	log(fmt.Sprintf("EventKey: k: %d, c: %c, mod: %s", curev.Key, curev.Ch, modStr(curev.Mod)))

	char := fmt.Sprintf("%c", curev.Ch)
	if char == " " {
		char = "space"
	}
	xdotool("key", char)
}

func parseInput() {
	switch curev.Type {
	case termbox.EventKey:
		keyEvent()
	case termbox.EventMouse:
		mouseEvent()
	case termbox.EventNone:
		log("EventNone")
	}
}

// a channel to tell it to stop
var stopchan = make(chan struct{})
// a channel to signal that it's stopped
var stoppedchan = make(chan struct{})

func xzoomBackground(){
	go func(){ // work in background
	  // close the stoppedchan when this func
	  // exits
	  defer close(stoppedchan)
	  // TODO: do setup work
	  defer func(){
	    // TODO: do teardown work
	  }()
	  for {
	    select {
	      default:
	        C.xzoom_loop()
					time.Sleep(40 * time.Millisecond) // 25fps
	      case <-stopchan:
	        // stop
	        return
	    }
	  }
	}()
}

func main() {
	C.xzoom_init()
	xzoomBackground()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputMouse)
	initialise()
	parseInput()

	data := make([]byte, 0, 64)
mainloop:
	for {
		if cap(data)-len(data) < 32 {
			newdata := make([]byte, len(data), len(data)+32)
			copy(newdata, data)
			data = newdata
		}
		beg := len(data)
		d := data[beg : beg+32]
		switch ev := termbox.PollRawEvent(d); ev.Type {
		case termbox.EventRaw:
			data = data[:beg+ev.N]
			current = fmt.Sprintf("%q", data)
			// TODO: think of a different way to exit, 'q' will be needed for actual text input.
			if current == `"q"` {
				close(stopchan)  // tell it to stop
				<-stoppedchan    // wait for it to have stopped
				break mainloop
			}

			for {
				ev := termbox.ParseEvent(data)
				if ev.N == 0 {
					break
				}
				curev = ev
				copy(data, data[curev.N:])
				data = data[:len(data)-curev.N]
			}
		case termbox.EventError:
			panic(ev.Err)
		}
		parseInput()
	}
}
