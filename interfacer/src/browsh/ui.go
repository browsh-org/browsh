package browsh

import (
	"github.com/gdamore/tcell"
	"github.com/spf13/viper"
)

var (
	urlInputBox = inputBox{
		X:        0,
		Y:        1,
		Height:   1,
		text:     nil,
		FgColour: [3]int32{255, 255, 255},
		bgColour: [3]int32{-1, -1, -1},
	}
)

// Render tabs, URL bar, status messages, etc
func renderUI() {
	renderTabs()
	renderURLBar()
}

// Write a simple text string to the screen.
// Not for use in the browser frames themselves. If you want anything to appear in
// the browser that must be done through the webextension.
func writeString(x, y int, str string, style tcell.Style) {
	xOriginal := x
	if viper.GetBool("http-server-mode") {
		Log(str)
		return
	}
	for _, c := range str {
		if string(c) == "\n" {
			y++
			x = xOriginal
			continue
		}
		screen.SetCell(x, y, style, c)
		x++
	}
}

func fillLineToEnd(x, y int) {
	width, _ := screen.Size()
	for i := x; i < width-1; i++ {
		writeString(i, y, " ", tcell.StyleDefault)
	}
}

func renderTabs() {
	var tab *tab
	var style tcell.Style
	count := 0
	xPosition := 0
	tabTitleLength := 20
	for _, tabID := range tabsOrder {
		tab = Tabs[tabID]
		tabTitle := []rune(tab.Title)
		tabTitleContent := string(tabTitle[0:tabTitleLength])
		style = tcell.StyleDefault
		if CurrentTab.ID == tabID {
			style = tcell.StyleDefault.Reverse(true)
		}
		writeString(xPosition, 0, tabTitleContent, style)
		style = tcell.StyleDefault.Reverse(false)
		count++
		xPosition = count * (tabTitleLength + 1)
		writeString(xPosition-1, 0, "|", style)
	}
	fillLineToEnd(xPosition, 0)
}

func renderURLBar() {
	var content []rune
	if urlInputBox.isActive {
		writeString(0, 1, string(content), tcell.StyleDefault)
		content = append(urlInputBox.text, ' ')
		urlInputBox.renderURLBox()
	} else {
		content = []rune(CurrentTab.URI)
		writeString(0, 1, string(content), tcell.StyleDefault)
	}
	fillLineToEnd(len(content), 1)
}

func urlBarFocusToggle() {
	if urlInputBox.isActive {
		URLBarFocus(false)
	} else {
		URLBarFocus(true)
	}
}

// Set the focus of the URL bar. Also used in tests to ensure the URL bar is in fact focussed as
// toggling doesn't guarantee that you will gain focus.
func URLBarFocus(on bool) {
	if !on {
		activeInputBox = nil
		urlInputBox.isActive = false
		urlInputBox.selectionOff()
	} else {
		activeInputBox = &urlInputBox
		urlInputBox.isActive = true
		urlInputBox.xScroll = 0
		urlInputBox.text = []rune(CurrentTab.URI)
		urlInputBox.putCursorAtEnd()
		urlInputBox.selectAll()
	}
}

func overlayVimMode() {
	_, height := screen.Size()
	switch currentVimMode {
	case insertMode:
		writeString(0, height-1, "ins", tcell.StyleDefault)
	case insertModeHard:
		writeString(0, height-1, "INS", tcell.StyleDefault)
	case linkMode:
		writeString(0, height-1, "lnk", tcell.StyleDefault)
	case linkModeNewTab:
		writeString(0, height-1, "LNK", tcell.StyleDefault)
	case linkModeMultipleNewTab:
		writeString(0, height-1, "*LNK", tcell.StyleDefault)
	case linkModeCopy:
		writeString(0, height-1, "cp", tcell.StyleDefault)
	case visualMode:
		writeString(0, height-1, "vis", tcell.StyleDefault)
	case caretMode:
		writeString(0, height-1, "car", tcell.StyleDefault)
		writeString(caretPos.X, caretPos.Y, "#", tcell.StyleDefault)
	case findMode:
		writeString(0, height-1, "/"+findText, tcell.StyleDefault)
	case markModeMake:
		writeString(0, height-1, "mark", tcell.StyleDefault)
	case markModeGoto:
		writeString(0, height-1, "goto", tcell.StyleDefault)
	}

	switch currentVimMode {
	case linkMode, linkModeNewTab, linkModeMultipleNewTab, linkModeCopy:
		if !linkModeWithHints {
			findAndHighlightTextOnScreen(linkText)
		}

		if linkHintWriteStringCalls != nil {
			(*linkHintWriteStringCalls)()
		}
	}
}

func overlayPageStatusMessage() {
	_, height := screen.Size()
	writeString(0, height-1, CurrentTab.StatusMessage, tcell.StyleDefault)
}

func overlayCallToSupport() {
	var right int
	var message string
	if viper.GetString("browsh_supporter") == "I have shown my support for Browsh" {
		return
	}
	width, height := screen.Size()
	message = " Unsupported version"
	right = width - len(message)
	writeString(right, height-2, message, tcell.StyleDefault)
	message = "  See brow.sh/donate"
	right = width - len(message)
	writeString(right, height-1, message, tcell.StyleDefault)
}

func reverseCellColour(x, y int) {
	mainRune, combiningRunes, style, _ := screen.GetContent(x, y)
	style = style.Reverse(true)
	screen.SetContent(x, y, mainRune, combiningRunes, style)
}
