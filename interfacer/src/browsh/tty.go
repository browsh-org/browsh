package browsh

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/go-errors/errors"
	"github.com/spf13/viper"
)

var (
	screen   tcell.Screen
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
	IsMonochromeMode = viper.GetBool("monochrome")
	screen.EnableMouse()
	screen.Clear()
}

func sendTtySize() {
	width, height := screen.Size()
	urlInputBox.Width = width
	sendMessageToWebExtension(fmt.Sprintf("/tty_size,%d,%d", width, height))
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
	if CurrentTab == nil {
		if ev.Key() == tcell.KeyCtrlQ {
			quitBrowsh()
		}
		return
	}
	switch ev.Key() {
	case tcell.KeyCtrlQ:
		quitBrowsh()
	case tcell.KeyCtrlL:
		urlBarFocusToggle()
	case tcell.KeyCtrlT:
		createNewEmptyTab()
	case tcell.KeyCtrlU:
		if !isNewEmptyTabActive() {
			sendMessageToWebExtension("/new_tab,view-source:" + CurrentTab.URI)
		}
	case tcell.KeyCtrlW:
		removeTab(CurrentTab.ID)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if activeInputBox == nil {
			sendMessageToWebExtension("/tab_command,/history_back")
		}
	}
	if ev.Rune() == 'm' && ev.Modifiers() == 4 {
		toggleMonochromeMode()
	}
	if ev.Key() == 279 && ev.Modifiers() == 0 {
		// F1 key
		openHelpTab()
	}
	if isKey("tty.keys.next-tab", ev) {
		nextTab()
	}
	if !urlInputBox.isActive {
		forwardKeyPress(ev)
	}
	if activeInputBox != nil {
		handleInputBoxInput(ev)
	} else {
		handleScrolling(ev) // TODO: shouldn't you be able to still use mouse scrolling?
	}
}

func isKey(userKey string, ev *tcell.EventKey) bool {
	key := viper.GetStringSlice(userKey)
	runeMatch := []rune(key[0])[0] == ev.Rune()
	intKey, _ := strconv.Atoi(key[1])
	keyCodeMatch := intKey == int(ev.Key())
	modifierKey, _ := strconv.Atoi(key[2])
	modifierMatch := modifierKey == int(ev.Modifiers())
	return runeMatch && keyCodeMatch && modifierMatch
}

func quitBrowsh() {
	if !viper.GetBool("firefox.use-existing") {
		quitFirefox()
	}
	Shutdown(errors.New("normal"))
}

func toggleMonochromeMode() {
	IsMonochromeMode = !IsMonochromeMode
}

func openHelpTab() {
	sendMessageToWebExtension("/new_tab,https://www.brow.sh/docs/introduction/")
}

func forwardKeyPress(ev *tcell.EventKey) {
	if isMultiLineEnter(ev) {
		return
	}
	eventMap := map[string]interface{}{
		"key":  int(ev.Key()),
		"char": string(ev.Rune()),
		"mod":  int(ev.Modifiers()),
	}
	marshalled, _ := json.Marshal(eventMap)
	sendMessageToWebExtension("/stdin," + string(marshalled))
}

// Allow user to use ENTER key without triggering submission on multiline input
// boxes.
func isMultiLineEnter(ev *tcell.EventKey) bool {
	if activeInputBox == nil {
		return false
	}
	return activeInputBox.isMultiLine() && ev.Key() == 13 && ev.Modifiers() != 4
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
	sendMessageToWebExtension(
		fmt.Sprintf(
			"/tab_command,/scroll_status,%d,%d",
			CurrentTab.frame.xScroll,
			CurrentTab.frame.yScroll*2))
	if CurrentTab.frame.yScroll != yScrollOriginal {
		renderCurrentTabWindow()
	}
}

func handleMouseEvent(ev *tcell.EventMouse) {
	if CurrentTab == nil {
		return
	}
	x, y := ev.Position()
	xInFrame := x + CurrentTab.frame.xScroll
	yInFrame := y - uiHeight + CurrentTab.frame.yScroll
	button := ev.Buttons()
	if button == tcell.WheelUp || button == tcell.WheelDown {
		handleMouseScroll(button)
	}
	if button == 1 {
		CurrentTab.frame.maybeFocusInputBox(xInFrame, yInFrame)
	}
	eventMap := map[string]interface{}{
		"button":    int(button),
		"mouse_x":   int(xInFrame),
		"mouse_y":   int(yInFrame),
		"modifiers": int(ev.Modifiers()),
	}
	marshalled, _ := json.Marshal(eventMap)
	sendMessageToWebExtension("/stdin," + string(marshalled))
}

func handleMouseScroll(scrollType tcell.ButtonMask) {
	yScrollOriginal := CurrentTab.frame.yScroll
	_, height := screen.Size()
	height -= uiHeight
	if scrollType == tcell.WheelUp {
		CurrentTab.frame.yScroll -= 1
	} else if scrollType == tcell.WheelDown {
		CurrentTab.frame.yScroll += 1
	}
	CurrentTab.frame.limitScroll(height)
	sendMessageToWebExtension(
		fmt.Sprintf(
			"/tab_command,/scroll_status,%d,%d",
			CurrentTab.frame.xScroll,
			CurrentTab.frame.yScroll*2))
	if CurrentTab.frame.yScroll != yScrollOriginal {
		renderCurrentTabWindow()
	}
}

func handleTTYResize() {
	width, _ := screen.Size()
	urlInputBox.Width = width
	screen.Sync()
	sendTtySize()
}

// Tcell uses a buffer to collect screen updates on, it only actually sends
// ANSI rendering commands to the terminal when we tell it to. And even then it
// will try to minimise rendering commands by only rendering parts of the terminal
// that have changed.
func renderCurrentTabWindow() {
	var currentCell cell
	var styling = tcell.StyleDefault
	var runeChars []rune
	width, height := screen.Size()
	if CurrentTab == nil || CurrentTab.frame.cells == nil {
		return
	}
	CurrentTab.frame.overlayInputBoxContent()
	for y := 0; y < height-uiHeight; y++ {
		for x := 0; x < width; x++ {
			currentCell = getCell(x, y)
			runeChars = currentCell.character
			// TODO: do this is in isCharacterTransparent()
			if len(runeChars) == 0 {
				continue
			}
			if IsMonochromeMode {
				styling = styling.Foreground(tcell.ColorWhite)
				styling = styling.Background(tcell.ColorBlack)
				if runeChars[0] == '▄' {
					runeChars[0] = ' '
				}
			} else {
				styling = styling.Foreground(currentCell.fgColour)
				styling = styling.Background(currentCell.bgColour)
			}
			screen.SetCell(x, y+uiHeight, styling, runeChars[0])
		}
	}
	if activeInputBox != nil {
		activeInputBox.renderCursor()
	}
	overlayPageStatusMessage()
	overlayCallToSupport()
	screen.Show()
}

func getCell(x, y int) cell {
	var currentCell cell
	var ok bool
	frame := &CurrentTab.frame
	index := ((y + frame.yScroll) * frame.totalWidth) + (x + frame.xScroll)
	if currentCell, ok = frame.cells.load(index); !ok {
		fgColour, bgColour := getHatchedCellColours(x)
		currentCell = cell{
			fgColour:  fgColour,
			bgColour:  bgColour,
			character: []rune("▄"),
		}
	}
	return currentCell
}

func getHatchedCellColours(x int) (tcell.Color, tcell.Color) {
	var bgColour, fgColour tcell.Color
	if x%2 == 0 {
		bgColour = tcell.NewHexColor(0xa9a9a9)
		fgColour = tcell.NewHexColor(0x797979)
	} else {
		bgColour = tcell.NewHexColor(0x797979)
		fgColour = tcell.NewHexColor(0xa9a9a9)
	}
	return fgColour, bgColour
}
