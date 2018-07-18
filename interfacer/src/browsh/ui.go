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
		screen.SetContent(x, y, c, nil, style)
		x++
	}
	screen.Show()
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
		urlBarFocus(false)
	} else {
		urlBarFocus(true)
	}
}

func urlBarFocus(on bool) {
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

func overlayPageStatusMessage() {
	_, height := screen.Size()
	writeString(0, height-1, CurrentTab.StatusMessage, tcell.StyleDefault)
}

func reverseCellColour(x, y int) {
	mainRune, combiningRunes, style, _ := screen.GetContent(x, y)
	style = style.Reverse(true)
	screen.SetContent(x, y, mainRune, combiningRunes, style)
}
