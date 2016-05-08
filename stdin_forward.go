package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"strings"
	"os"
	"os/exec"
)

var current string
var curev termbox.Event
var lastMouseButton string
var desktopX int
var desktopY int
var termWidth, termHeight = termbox.Size()

// For keeping track of the zoom
var zoomLevel float32 = 1
var viewport = map[string] float32 {
	"xSize": 1600,
	"ySize": 1200,
	"xOffset": 0,
	"yOffset": 0,
}

func log(msg string) {
	msg = msg + "\n"
	f, err := os.OpenFile("stdin.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
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
	if err := exec.Command("xdotool", args...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
		if curev.Mod&termbox.ModCtrl != 0 {
			trackZoom("out")
		}
		return []string{"click", "4"}
	case termbox.MouseWheelDown:
		if curev.Mod&termbox.ModCtrl != 0 {
			trackZoom("in")
		}
		return []string{"click", "5"}
	}
	return []string{""}
}

// Auxillary data, whether the mouse was moving or a mod key like CTRL
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
	if curev.Mod&termbox.ModCtrl != 0 {
		xdotool("keydown", "alt")
	} else {
		xdotool("keyup", "alt")
	}


	// Always move the mouse first. This is because we're not constantly updating the mouse position,
	// *unless* a drag event is happening. This saves bandwidth and also mouse movement isn't supported
	// on all terminals.
	setCurrentDesktopCoords()
	xdotool("mousemove", fmt.Sprintf("%d", desktopX), fmt.Sprintf("%d", desktopY))

	// Send a button press to X. Note that the "Motion" modifier is sent when the user is doing
	// a drag event and thus mouse reporting will be constantly streamed.
	if !strings.Contains(modStr(curev.Mod), "Motion") {
		xdotool(mouseButtonStr(curev.Key)...)
	}
}

// Convert terminal coords into desktop coords
func setCurrentDesktopCoords() {
	termWidthFloat := float32(termWidth)
	termHeightFloat := float32(termHeight)
	eventX := float32(curev.MouseX)
	eventY := float32(curev.MouseY)
	x := (eventX * (viewport["xSize"] / termWidthFloat)) + viewport["xOffset"]
	y := (eventY * (viewport["ySize"] / termHeightFloat)) + viewport["yOffset"]
	desktopX = int(x)
	desktopY = int(y)
}

// XFCE doesn't provide the current zoom, so we need to keep track of it.
// For every zoom level the terminal coords will be mapped differently onto the X desktop.
// TODO: support custom desktop sizes.
func trackZoom(direction string) {
	if direction == "in" {
		zoomLevel += 0.2
	} else {
		zoomLevel -= 0.2
	}
	setCurrentDesktopCoords()
	desktopXFloat := float32(desktopX)
	desktopYFloat := float32(desktopY)
	viewport["xOffset"] = desktopXFloat - (viewport["xSize"] / 2)
	viewport["yOffset"] = desktopYFloat - (viewport["ySize"] / 2)
	viewport["xSize"] = 1600 / zoomLevel
	viewport["ySize"] = 1200 / zoomLevel
	log(fmt.Sprintf("viewport: %s", viewport))
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

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputMouse)
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
			if current == `"q"` {
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
