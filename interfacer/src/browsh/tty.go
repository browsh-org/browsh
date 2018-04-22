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
	maxYScroll := (frame.height / 2) - height
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
	return
	var styling = tcell.StyleDefault
	var runeChars []rune
	width, _ := screen.Size()
	index := 0
	for y := 0; y < uiHeight ; y++ {
		for x := 0; x < width; x++ {
			styling = styling.Foreground(frame.cells[index].fgColour)
			styling = styling.Background(frame.cells[index].bgColour)
			runeChars = frame.cells[index].character
			index++
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
	if (len(frame.pixels) == 0 || len(frame.text) == 0) { return }
	var styling = tcell.StyleDefault
	var runeChars []rune
	width, height := screen.Size()
	if (frame.width == 0 || frame.height == 0) {
		Log("Not rendering frame with a zero dimension")
		return
	}
	index := 0
	for y := 0; y < height - uiHeight; y++ {
		for x := 0; x < width; x++ {
			index = ((y + yScroll) * frame.width) + ((x + xScroll))
			if (!checkCell(index, x + xScroll, y + yScroll)) { return }
			styling = styling.Foreground(frame.cells[index].fgColour)
			styling = styling.Background(frame.cells[index].bgColour)
			runeChars = frame.cells[index].character
			if (len(runeChars) == 0) { continue } // TODO: shouldn't need this
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
	if (index >= len(frame.cells)) {
		message := fmt.Sprintf(
			"Blank frame data (size: %d) at %dx%d, index: %d",
			len(frame.cells), x, y, index)
		Log(message)
		return false;
	}
	return true;
}
