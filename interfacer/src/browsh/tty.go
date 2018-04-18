package browsh

import (
	"fmt"
	"os"
	"encoding/json"

	"github.com/gdamore/tcell"
	"github.com/go-errors/errors"
)

var (
	xScroll = 0
	yScroll = 0
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
			handleUserKeyPress(ev)
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
				"mouse_x":   int(x + xScroll),
				"mouse_y":   int(y - uiHeight + yScroll),
				"modifiers": int(ev.Modifiers()),
			}
			marshalled, _ := json.Marshal(eventMap)
			sendMessageToWebExtension("/stdin," + string(marshalled))
		}
	}
}

func handleUserKeyPress(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyCtrlQ {
		if !*isUseExistingFirefox {
			quitFirefox()
		}
		Shutdown(errors.New("normal"))
	}
	handleScrolling(ev)
}

func handleScrolling(ev *tcell.EventKey) {
	yScrollOriginal := yScroll
	_, height := screen.Size()
	height -= uiHeight
	if ev.Key() == tcell.KeyUp {
		yScroll -= 2
	}
	if ev.Key() == tcell.KeyDown {
		yScroll += 2
	}
	if ev.Key() == tcell.KeyPgUp {
		yScroll -= height
	}
	if ev.Key() == tcell.KeyPgDn {
		yScroll += height
	}
	limitScroll(height)
	if (yScroll != yScrollOriginal) {
		renderFrame()
	}
}

func limitScroll(height int) {
	maxYScroll := (frameHeight / 2) - height
	if (yScroll > maxYScroll) { yScroll = maxYScroll }
	if (yScroll < 0) { yScroll = 0 }
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

func renderAll() {
	renderUI()
	renderFrame()
}

// Render the tabs and URL bar
// TODO: Temporary function, UI rendering should all be moved into this CLI app
func renderUI() {
	var styling = tcell.StyleDefault
	var character string
	var runeChars []rune
	width, _ := screen.Size()
	index := 0
	for y := 0; y < uiHeight ; y++ {
		for x := 0; x < width; x++ {
			styling = styling.Foreground(getRGBColor(index))
			index += 3
			styling = styling.Background(getRGBColor(index))
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

// Tcell uses a buffer to collect screen updates on, it only actually sends
// ANSI rendering commands to the terminal when we tell it to. And even then it
// will try to minimise rendering commands by only rendering parts of the terminal
// that have changed.
func renderFrame() {
	var styling = tcell.StyleDefault
	var character string
	var runeChars []rune
	width, height := screen.Size()
	uiSize :=  uiHeight * width * 7
	if (len(frame) == uiSize) {
		Log("Not rendering zero-size frame data")
		return
	}
	if (frameWidth == 0 || frameHeight == 0) {
		Log("Not rendering frame with a zero dimension")
		return
	}
	index := 0
	for y := 0; y < height - uiHeight; y++ {
		for x := 0; x < width; x++ {
			index = ((y + yScroll) * frameWidth * 7) + ((x + xScroll) * 7)
			index += uiSize
			if (!checkCell(index, x + xScroll, y + yScroll)) { return }
			styling = styling.Foreground(getRGBColor(index))
			index += 3
			styling = styling.Background(getRGBColor(index))
			index += 3
			character = frame[index]
			runeChars = []rune(character)
			index++
			if (character == "WIDE") {
				continue
			}
			screen.SetCell(x, y + uiHeight, styling, runeChars[0])
		}
	}
	overlayPageStatusMessage(height)
	screen.Show()
}

func overlayPageStatusMessage(height int) {
	message := State["page_status_message"]
	x := 0
	fg := tcell.NewHexColor(int32(0xffffff))
	bg := tcell.NewHexColor(int32(0x000000))
	style := tcell.StyleDefault
	style.Foreground(fg)
	style.Foreground(bg)
	for _, c := range message {
		screen.SetCell(x, height - 1, style, c)
		x++
	}
}

func checkCell(index, x, y int) bool {
	for i := 0; i < 7; i++ {
		if (index + i >= len(frame) || frame[index + i] == "") {
			message := fmt.Sprintf("Blank frame data (size: %d) at %dx%d, index:%d/%d", len(frame), x, y, index, i)
			Log(message)
			Log(fmt.Sprintf("%d", yScroll))
			return false;
		}
	}
	return true;
}

// Given a raw frame from the webextension, find the RGB colour at a given
// 1 dimensional index.
func getRGBColor(index int) tcell.Color {
	rgb := frame[index:index + 3]
	return tcell.NewRGBColor(
			toInt32(rgb[0]),
			toInt32(rgb[1]),
			toInt32(rgb[2]))
}
