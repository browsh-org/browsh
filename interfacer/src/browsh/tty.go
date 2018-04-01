package browsh

import (
	"fmt"
	"os"
	"strconv"
	"encoding/json"

	"github.com/gdamore/tcell"
	"github.com/go-errors/errors"
)

func setupTcell() {
	var err error
	if err = screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	screen.EnableMouse()
	screen.Clear()
}

func sendTtySize() {
	x, y := screen.Size()
	sendMessageToWebExtension(fmt.Sprintf("/tty_size,%d,%d", x, y))
}

// This is basically a proxy that listens to STDIN and forwards all relevant input
// from the user to the webextension. So keyboard, mouse, terminal resizes, etc.
func readStdin() {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlQ {
				if !*isUseExistingFirefox {
					quitFirefox()
				}
				Shutdown(errors.New("normal"))
			}
			eventMap := map[string]interface{}{
				"key":  int(ev.Key()),
				"char": string(ev.Rune()),
				"mod":  int(ev.Modifiers()),
			}
			marshalled, _ := json.Marshal(eventMap)
			sendMessageToWebExtension("/stdin," + string(marshalled))
		case *tcell.EventResize:
			screen.Sync()
			sendTtySize()
		case *tcell.EventMouse:
			x, y := ev.Position()
			button := ev.Buttons()
			eventMap := map[string]interface{}{
				"button":    int(button),
				"mouse_x":   int(x),
				"mouse_y":   int(y),
				"modifiers": int(ev.Modifiers()),
			}
			marshalled, _ := json.Marshal(eventMap)
			sendMessageToWebExtension("/stdin," + string(marshalled))
		}
	}
}

// Write a simple text string to the screen. Not for use in the browser frames
// themselves. If you want anything to appear in the browser that must be done
// through the webextension.
func writeString(x, y int, str string) {
	var defaultColours = tcell.StyleDefault
	rgb := tcell.NewHexColor(int32(0xffffff))
	defaultColours.Foreground(rgb)
	for _, c := range str {
		screen.SetContent(x, y, c, nil, defaultColours)
		x++
	}
	screen.Sync()
}

// Tcell uses a buffer to collect screen updates on, it only actually sends
// ANSI rendering commands to the terminal when we tell it to. And even then it
// will try to minimise rendering commands by only rendering parts of the terminal
// that have changed.
func renderFrame(frame []string) {
	var styling = tcell.StyleDefault
	var character string
	var runeChars []rune
	width, height := screen.Size()
	if (width * height * 7 != len(frame)) {
		Log("Not rendering frame: current frame is not the same size as the screen")
		Log(fmt.Sprintf("screen: %d, frame: %d", width * height * 7, len(frame)))
		return
	}
	index := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			styling = styling.Foreground(getRGBColor(frame, index))
			index += 3
			styling = styling.Background(getRGBColor(frame, index))
			index += 3
			character = frame[index]
			runeChars = []rune(character)
			index++
			if (character == "WIDE") {
				continue
			}
			screen.SetCell(x, y, styling, runeChars[0])
		}
	}
	screen.Show()
}

// Given a raw frame from the webextension, find the RGB colour at a given
// 1 dimensional index.
func getRGBColor(frame []string, index int) tcell.Color {
	rgb := frame[index:index + 3]
	return tcell.NewRGBColor(
			toInt32(rgb[0]),
			toInt32(rgb[1]),
			toInt32(rgb[2]))
}

// Convert a string representation of an integer to an integer
func toInt32(char string) int32 {
	i, err := strconv.ParseInt(char, 10, 32)
	if err != nil {
		Shutdown(err)
	}
	return int32(i)
}
