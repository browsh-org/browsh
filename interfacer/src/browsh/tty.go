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
	screen tcell.Screen
	uiHeight = 2
	// IsMonochromeMode decides whether to render the TTY in full colour or monochrome
	IsMonochromeMode = false
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
		case *tcell.EventResize:
			handleTTYResize()
		case *tcell.EventMouse:
			handleMouseEvent(ev)
		}
	}
}

func handleUserKeyPress(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyCtrlQ:
		quitBrowsh()
	case tcell.KeyCtrlL:
		urlBarFocusToggle()
	}
	if (ev.Rune() == 'm' && ev.Modifiers() == 4) {
		toggleMonochromeMode()
	}
	forwardKeyPress(ev)
	if activeInputBox != nil {
		handleInputBoxInput(ev)
	} else {
		handleScrolling(ev) // TODO: shouldn't you be able to still use mouse scrolling?
	}
}

func quitBrowsh() {
	if !*isUseExistingFirefox {
		quitFirefox()
	}
	Shutdown(errors.New("normal"))
}

func urlBarFocusToggle() {
	if urlInputBox.isActive {
		activeInputBox = nil
		urlInputBox.isActive = false
	} else {
		activeInputBox = &urlInputBox
		urlInputBox.isActive = true
		urlInputBox.text = ""
	}
}

func toggleMonochromeMode() {
	IsMonochromeMode = !IsMonochromeMode
}

func forwardKeyPress(ev *tcell.EventKey) {
	eventMap := map[string]interface{}{
		"key":  int(ev.Key()),
		"char": string(ev.Rune()),
		"mod":  int(ev.Modifiers()),
	}
	marshalled, _ := json.Marshal(eventMap)
	sendMessageToWebExtension("/stdin," + string(marshalled))
}

func handleScrolling(ev *tcell.EventKey) {
	yScrollOriginal := CurrentTab.frame.yScroll
	_, height := screen.Size()
	height -= uiHeight
	if ev.Key() == tcell.KeyUp {
		CurrentTab.frame.yScroll -= 2
	}
	if ev.Key() == tcell.KeyDown {
		CurrentTab.frame.yScroll += 2
	}
	if ev.Key() == tcell.KeyPgUp {
		CurrentTab.frame.yScroll -= height
	}
	if ev.Key() == tcell.KeyPgDn {
		CurrentTab.frame.yScroll += height
	}
	CurrentTab.frame.limitScroll(height)
	if (CurrentTab.frame.yScroll != yScrollOriginal) {
		renderCurrentTabWindow()
	}
}

func handleMouseEvent(ev *tcell.EventMouse) {
	x, y := ev.Position()
	button := ev.Buttons()
	eventMap := map[string]interface{}{
		"button":    int(button),
		"mouse_x":   int(x + CurrentTab.frame.xScroll),
		"mouse_y":   int(y - uiHeight + CurrentTab.frame.yScroll),
		"modifiers": int(ev.Modifiers()),
	}
	marshalled, _ := json.Marshal(eventMap)
	sendMessageToWebExtension("/stdin," + string(marshalled))
}

func handleTTYResize() {
	width, _ := screen.Size()
	// TODO: How does this work with wide UTF8 chars?
	urlInputBox.Width = width - len(urlBarControls)
	screen.Sync()
	sendTtySize()
}

func renderAll() {
	renderUI()
	renderCurrentTabWindow()
}

// Tcell uses a buffer to collect screen updates on, it only actually sends
// ANSI rendering commands to the terminal when we tell it to. And even then it
// will try to minimise rendering commands by only rendering parts of the terminal
// that have changed.
func renderCurrentTabWindow() {
	if (len(CurrentTab.frame.pixels) == 0 || len(CurrentTab.frame.text) == 0) {
		Log("Not rendering frame without complimentary pixels and text:")
		Log(
			fmt.Sprintf(
				"pixels: %d, text: %d",
				len(CurrentTab.frame.pixels), len(CurrentTab.frame.text)))
		return
	}
	var styling = tcell.StyleDefault
	var runeChars []rune
	frame := &CurrentTab.frame
	width, height := screen.Size()
	if (frame.width == 0 || frame.height == 0) {
		Log("Not rendering frame with a zero dimension")
		return
	}
	index := 0
	for y := 0; y < height - uiHeight; y++ {
		for x := 0; x < width; x++ {
			index = ((y + frame.yScroll) * frame.width) + ((x + frame.xScroll))
			if (!checkCell(index, x + frame.xScroll, y + frame.yScroll)) { return }
			runeChars = frame.cells[index].character
			// TODO: do this is in isCharacterTransparent()
			if (len(runeChars) == 0) { continue }
			if IsMonochromeMode {
				styling = styling.Foreground(tcell.ColorWhite)
				styling = styling.Background(tcell.ColorBlack)
				if runeChars[0] == '▄' {
					runeChars[0] = ' '
				}
			} else {
				styling = styling.Foreground(frame.cells[index].fgColour)
				styling = styling.Background(frame.cells[index].bgColour)
			}
			screen.SetCell(x, y + uiHeight, styling, runeChars[0])
		}
	}
	screen.Show()
}

func checkCell(index, x, y int) bool {
	if (index >= len(CurrentTab.frame.cells)) {
		message := fmt.Sprintf(
			"Blank frame data (size: %d) at %dx%d, index: %d",
			len(CurrentTab.frame.cells), x, y, index)
		Log(message)
		return false;
	}
	return true;
}
