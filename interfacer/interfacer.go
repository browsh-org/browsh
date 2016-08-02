package main

import (
	"fmt"
	"github.com/tombh/termbox-go"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Import the xzoom C code that creates an X window that zooms
// and pans the desktop.
// It's written in C because it borrows from the original xzoom
// program: http://git.r-36.net/xzoom/
// NB: The following comments are parsed by `go build` ...

// #cgo LDFLAGS: -lXext -lX11 -lXt
// #include "../xzoom/xzoom.h"
import "C"

var logfile string
var current string
var curev termbox.Event
var lastMouseButton string

var hipWidth int
var hipHeight int
var envDesktopWidth int
var envDesktopHeight int
var desktopWidth float32
var desktopHeight float32
var desktopXFloat float32
var desktopYFloat float32
var roundedDesktopX int
var roundedDesktopY int

// Channels to control the background xzoom go routine
var stopXZoomChannel = make(chan struct{})
var xZoomStoppedChannel = make(chan struct{})

var panNeedsSetup bool
var panStartingX float32
var panStartingY float32

// Keyboard mode is for interacting with the desktop with the keyboard
// instead of the mouse.
var keyboardMode = false
var kbCursorX int
var kbCursorY int
var char string

var debugMode = parseENVVar("DEBUG") == 1

func initialise() {
	setupLogging()
	log("Starting...")
	setupTermbox()
	setupDimensions()
	if !debugMode {
		C.xzoom_init()
		xzoomBackground()
	}
}

func parseENVVar(variable string) int {
	value, err := strconv.Atoi(os.Getenv(variable))
	if err != nil {
		return 0
	}
	return value
}

func setupDimensions() {
	if debugMode {
		hipWidth, hipHeight = termbox.Size()
		envDesktopWidth = 1200
		envDesktopHeight = 900
	} else {
		hipWidth = parseENVVar("TTY_WIDTH")
		hipHeight = parseENVVar("TTY_HEIGHT")
		envDesktopWidth = parseENVVar("DESKTOP_WIDTH")
		envDesktopHeight = parseENVVar("DESKTOP_HEIGHT")
	}
	kbCursorX = int(hipWidth / 2)
	kbCursorY = int(hipHeight / 2)
	C.desktop_width = C.int(envDesktopWidth)
	C.width[C.SRC] = C.desktop_width
	C.width[C.DST] = C.desktop_width
	C.desktop_height = C.int(envDesktopHeight)
	C.height[C.SRC] = C.desktop_height
	C.height[C.DST] = C.desktop_height
	desktopWidth = float32(envDesktopWidth)
	desktopHeight = float32(envDesktopHeight)
	log(fmt.Sprintf("Desktop dimensions: W: %d, H: %d", envDesktopWidth, envDesktopHeight))
	log(fmt.Sprintf("Term dimensions: W: %d, H: %d", hipWidth, hipHeight))
}

func setupTermbox() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)
}

func setupLogging() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	os.Mkdir(filepath.Join(dir, "..", "logs"), os.ModePerm)
	logfile = fmt.Sprintf(filepath.Join(dir, "..", "logs", "input.log"))
	if _, err := os.Stat(logfile); err == nil {
		os.Truncate(logfile, 0)
	}
}

// Render text. This doesn't play very nice with hiptext, it can cause
// random tearing and flickering :( I suspect there are ways to overcome
// this for rendering outside of hiptext's area. But to render over hiptext
// it will probably mean patching hiptext.
func printXY(x, y int, s string, force bool) {
	for _, r := range s {

		// It seems termbox keeps an internal representation of the TTY
		// and won't try updating the cell unless it thinks it has changed.
		// This is only relevant because we're in competition with the
		// rapid screen updates from hiptext. This means that unless we
		// "remove -> flush -> redraw" everything hiptext removes whatever
		// we render.
		if force {
			// 32 is the space character
			termbox.SetCell(x, y, 32, termbox.ColorWhite, termbox.ColorDefault)
			termbox.Flush()
		}

		termbox.SetCell(x, y, r, termbox.ColorWhite, termbox.ColorDefault)
		x++
	}
	termbox.Flush()
}

func log(msg string) {
	f, oErr := os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if oErr != nil {
		panic(oErr)
	}
	defer f.Close()

	msg = msg + "\n"
	if _, wErr := f.WriteString(msg); wErr != nil {
		panic(wErr)
	}

	if debugMode {
		printXY(0, hipHeight - 1, msg, true)
	}
}

func min(a float32, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func getXGrab() int {
	return int(C.xgrab)
}
func getYGrab() int {
	return int(C.ygrab)
}

// Issue an xdotool command to simulate mouse and keyboard input
func xdotool(args ...string) {
	log(strings.Join(args, " "))
	if debugMode {
		return
	}
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
// Waiting for this PR: https://github.com/nsf/termbox-go/pull/126
func ctrlPressed() bool {
	return curev.Mod&termbox.ModCtrl != 0
}

func altPressed() bool {
	return curev.Mod&termbox.ModAlt != 0
}

// Whether the mouse is moving
func mouseMotion() bool {
	return curev.Mod&termbox.ModMotion != 0
}

// Convert Termbox symbols to xdotool arguments
func mouseButtonStr() []string {
	switch curev.Key {
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
			zoom("in")
			return []string{"noop"}
		}
		return []string{"click", "4"}
	case termbox.MouseWheelDown:
		if ctrlPressed() {
			zoom("out")
			return []string{"noop"}
		}
		return []string{"click", "5"}
	}
	return []string{""}
}

func zoom(direction string) {
	oldZoom := C.magnification

	// The actual zoom
	if direction == "in" {
		C.magnification++
	} else {
		if C.magnification > 1 {
			C.magnification--
		}
	}
	C.width[C.SRC] = (C.desktop_width + C.magnification - 1) / C.magnification
	C.height[C.SRC] = (C.desktop_height + C.magnification - 1) / C.magnification

	moveViewportForZoom(oldZoom)
	keepViewportInDesktop()
}

// Move the viewport so that the mouse is still over the same part of
// the desktop.
func moveViewportForZoom(oldZoom C.int) {
	factor := float32(oldZoom) / float32(C.magnification)
	magnifiedRelativeX := factor * (desktopXFloat - float32(C.xgrab))
	magnifiedRelativeY := factor * (desktopYFloat - float32(C.ygrab))
	C.xgrab = C.int(desktopXFloat - magnifiedRelativeX)
	C.ygrab = C.int(desktopYFloat - magnifiedRelativeY)
}

func keepViewportInDesktop() {
	manageViewportSize()
	manageViewportPosition()
}

func manageViewportSize() {
	if C.width[C.SRC] < 1 {
		C.width[C.SRC] = 1
	}
	if C.width[C.SRC] > C.desktop_width {
		C.width[C.SRC] = C.desktop_width
	}
	if C.height[C.SRC] < 1 {
		C.height[C.SRC] = 1
	}
	if C.height[C.SRC] > C.desktop_height {
		C.height[C.SRC] = C.desktop_height
	}
}

func manageViewportPosition() {
	if C.xgrab > (C.desktop_width - C.width[C.SRC]) {
		C.xgrab = C.desktop_width - C.width[C.SRC]
	}
	if C.xgrab < 0 {
		C.xgrab = 0
	}
	if C.ygrab > (C.desktop_height - C.height[C.SRC]) {
		C.ygrab = C.desktop_height - C.height[C.SRC]
	}
	if C.ygrab < 0 {
		C.ygrab = 0
	}
}

// Auxillary data. Whether the mouse was moving or a mod key like CTRL
// is being pressed at the same time.
func modStr(m termbox.Modifier) string {
	var out []string
	if mouseMotion() {
		out = append(out, "Motion")
	}
	if ctrlPressed() {
		out = append(out, "Ctrl")
	}
	return strings.Join(out, " ")
}

func isPanning() bool {
	mousePanning := ctrlPressed() && mouseMotion() && lastMouseButton == "1"
	kbPanning := (char == "U" || char == "K" || char == "N" || char == "H") && keyboardMode
	return mousePanning || kbPanning
}

func mouseEvent() {
	// Figure out where the mouse is on the actual real desktop.
	// Note that the zomming and panning code effectively keep the mouse in the exact same position relative
	// to the desktop, so mousemove *should* have no effect.
	setCurrentDesktopCoords()

	// Always move the mouse first so that button presses are correct. This is because we're not constantly
	// updating the mouse position, *unless* a drag event is happening. This saves bandwidth. Also, mouse
	// movement isn't supported on all terminals.
	xdotool("mousemove", fmt.Sprintf("%d", roundedDesktopX), fmt.Sprintf("%d", roundedDesktopY))

	if isPanning() {
		pan()
	} else {
		panNeedsSetup = true
		// Pressing of CTRL indicates that the user is panning or zooming, so there is no need to send
		// button presses.
		// TODO: What about CTRL+leftbutton to open new tab!?
		if !keyboardMode && !ctrlPressed() {
			xdotool(mouseButtonStr()...)
		}
	}
}

func pan() {
	if panNeedsSetup {
		panStartingX = desktopXFloat
		panStartingY = desktopYFloat
		panNeedsSetup = false
	}
	C.xgrab = C.int(float32(C.xgrab) + panStartingX - desktopXFloat)
	C.ygrab = C.int(float32(C.ygrab) + panStartingY - desktopYFloat)
	keepViewportInDesktop()
}

// Convert terminal coords into desktop coords
func setCurrentDesktopCoords() {
	hipWidthFloat := float32(hipWidth)
	hipHeightFloat := float32(hipHeight)
	eventX := float32(getCursorX())
	eventY := float32(getCursorY())
	width := float32(C.width[C.SRC])
	height := float32(C.height[C.SRC])
	xOffset := float32(C.xgrab)
	yOffset := float32(C.ygrab)
	desktopXFloat = (eventX * (width / hipWidthFloat)) + xOffset
	desktopYFloat = (eventY * (height / hipHeightFloat)) + yOffset
	roundedDesktopX = roundToInt(desktopXFloat)
	roundedDesktopY = roundToInt(desktopYFloat)
	log(
		fmt.Sprintf(
			"setCurrentDesktopCoords: tw: %d, th: %d, dx: %d, dy: %d, mag: %d",
			hipHeightFloat, hipWidthFloat, desktopXFloat, desktopYFloat, C.magnification))
}

func getCursorX() int {
	if keyboardMode {
		return kbCursorX
	}
	return curev.MouseX
}

func getCursorY() int {
	if keyboardMode {
		return kbCursorY
	}
	return curev.MouseY
}

// Convert a keyboard event into an xdotool command
// See: http://wiki.linuxquestions.org/wiki/List_of_Keysyms_Recognised_by_Xmodmap
func keyEvent() {
	var command string
	log(fmt.Sprintf("EventKey: k: %d, c: %c, mod: %s", curev.Key, curev.Ch, modStr(curev.Mod)))

	key := getSpecialKeyPress()

	if curev.Key == 0 {
		key = fmt.Sprintf("%c", curev.Ch)
		char = key
		command = "type"
	} else {
		command = "key"
	}

	// What is this? It always appears when the program starts :/
	badkey := fmt.Sprintf("%s", curev.Ch) == "%!s(int32=0)" && curev.Key == 0

	if key == "" || badkey {
		log(fmt.Sprintf("No key found for keycode: %d"))
		return
	}

	if (curev.Key == termbox.KeyCtrlM) && altPressed() {
		keyboardMode = !keyboardMode
		printXY(0, hipHeight, "     ", true)
	}

	if keyboardMode {
		mouseEvent()
		handleKeyboardMode(key)
		renderCursor()
		printXY(0, hipHeight, "KB ON ", true)
		return
	}

	xdotool(command, key)
}

func handleKeyboardMode(key string) {
	switch key {
	case "u", "U":
		kbCursorY--
		if kbCursorY < 0 {
			kbCursorY = 0
		}
	case "n", "N":
		kbCursorY++
		if kbCursorY > hipHeight {
			kbCursorY = hipHeight
		}
	case "h", "H":
		kbCursorX--
		if kbCursorX < 0 {
			kbCursorX = 0
		}
	case "k", "K":
		kbCursorX++
		if kbCursorX > hipWidth {
			kbCursorX = hipWidth
		}
	case "ctrl+u":
		zoom("in")
	case "ctrl+n":
		zoom("out")
	case "j":
		xdotool("click", "1")
	case "r":
		xdotool("click", "2")
	case "t":
		xdotool("click", "3")
	}

}

func renderCursor() {
	printXY(kbCursorX, kbCursorY, "+", false)
}

func getSpecialKeyPress() string {
	var key string
	switch curev.Key {
	case termbox.KeyEnter:
		key = "Return"
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		key = "BackSpace"
	case termbox.KeySpace:
		key = "space"
	case termbox.KeyF1:
		key = "F1"
	case termbox.KeyF2:
		key = "F2"
	case termbox.KeyF3:
		key = "F3"
	case termbox.KeyF4:
		key = "F4"
	case termbox.KeyF5:
		key = "F5"
	case termbox.KeyF6:
		key = "F6"
	case termbox.KeyF7:
		key = "F7"
	case termbox.KeyF8:
		key = "F8"
	case termbox.KeyF9:
		key = "F9"
	case termbox.KeyF10:
		key = "F10"
	case termbox.KeyF11:
		key = "F11"
	case termbox.KeyF12:
		key = "F12"
	case termbox.KeyInsert:
		key = "Insert"
	case termbox.KeyDelete:
		key = "Delete"
	case termbox.KeyHome:
		key = "Home"
	case termbox.KeyEnd:
		key = "End"
	case termbox.KeyPgup:
		key = "Prior"
	case termbox.KeyPgdn:
		key = "Next"
	case termbox.KeyArrowUp:
		key = "Up"
	case termbox.KeyArrowDown:
		key = "Down"
	case termbox.KeyArrowLeft:
		key = "Left"
	case termbox.KeyArrowRight:
		key = "Right"
	case termbox.KeyCtrlU:
		key = "ctrl+u"
	case termbox.KeyCtrlL:
		key = "ctrl+l"
	case termbox.KeyCtrlN:
		key = "ctrl+n"
	}
	return key
}

func parseInput() {
	switch curev.Type {
	case termbox.EventKey:
		keyEvent()
	case termbox.EventMouse:
		log(
			fmt.Sprintf(
				"EventMouse: x: %d, y: %d, b: %s, mod: %s",
				getCursorX(), getCursorY(), mouseButtonStr(), modStr(curev.Mod)))
		mouseEvent()
	case termbox.EventNone:
		log("EventNone")
	}
}

// Run the xzoom window in a background go routine
func xzoomBackground() {
	go func() {
		defer close(xZoomStoppedChannel)
		for {
			select {
			default:
				C.do_iteration()
				time.Sleep(40 * time.Millisecond) // 25fps
			case <-stopXZoomChannel:
				// Gracefully close the xzoom go routine
				return
			}
		}
	}()
}

func needToExit() bool {
	// CTRL+ALT+Q
	if (curev.Key == termbox.KeyCtrlQ) && altPressed() {
		return true
	}
	return false
}

func teardown() {
	termbox.Close()
	if !debugMode {
		close(stopXZoomChannel)
		<-xZoomStoppedChannel
	}
}

// I'm afraid I don't understand most of what this does :/
// TODO: if anyone can shed some light on this. Add some comments, refactor it...
func mainLoop() {
	data := make([]byte, 0, 64)
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
			for {
				ev := termbox.ParseEvent(data)
				if ev.N == 0 {
					break
				}
				curev = ev
				if needToExit() {
					log("Exit requested by user")
					return
				}
				copy(data, data[curev.N:])
				data = data[:len(data)-curev.N]
			}
		case termbox.EventError:
			panic(ev.Err)
		}
		parseInput()
	}

}

func main() {
	initialise()
	defer func() {
		teardown()
	}()
	mainLoop()
}
