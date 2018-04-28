package browsh

import (
	"unicode/utf8"

	"github.com/gdamore/tcell"
)

var (
	urlBarControls = " ← | x | "
	urlInputBox = inputBox{
		X: len(urlBarControls),
		Y: 1,
		Height: 1,
		text: "",
	}
)
// Render tabs, URL bar, status messages, etc
func renderUI() {
	renderTabs()
	renderURLBar()
	overlayPageStatusMessage()
}

// Write a simple text string to the screen.
// Not for use in the browser frames themselves. If you want anything to appear in
// the browser that must be done through the webextension.
func writeString(x, y int, str string) {
	var defaultColours = tcell.StyleDefault
	rgb := tcell.NewHexColor(int32(0xffffff))
	defaultColours.Foreground(rgb)
	for _, c := range str {
		screen.SetContent(x, y, c, nil, defaultColours)
		x++
	}
	screen.Show()
}

func fillLineToEnd(x, y int) {
	width, _ := screen.Size()
	for i := x; i < width - 1; i++ {
		writeString(i, y, " ")
	}
}

func renderTabs() {
	count := 0
	xPosition := 0
	tabTitleLength := 15
	for _, tab := range tabs {
		if (tab.frame.text == nil) { continue } // TODO: this shouldn't be needed
		tabTitle := []rune(tab.Title)
		tabTitleContent := string(tabTitle[0:tabTitleLength]) + " |x "
		writeString(xPosition, 0, tabTitleContent)
		count++
		xPosition = (count * tabTitleLength) + 4
	}
	fillLineToEnd(xPosition, 0)
}

func renderURLBar() {
	content := urlBarControls + getCurrentURLBarContents()
	writeString(0, 1, content)
	fillLineToEnd(utf8.RuneCountInString(content), 1)
}

func overlayPageStatusMessage() {
	_, height := screen.Size()
	writeString(0, height - 1, CurrentTab.StatusMessage)
}

func getCurrentURLBarContents() string {
	var contents string
	if urlInputBox.isActive {
		contents = urlInputBox.text
	} else {
		contents = CurrentTab.URI
	}
	return contents
}
