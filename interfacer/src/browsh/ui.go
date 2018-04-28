package browsh

import (
	"unicode/utf8"
)

// Render tabs, URL bar, status messages, etc
func renderUI() {
	renderTabs()
	renderURLBar()
	overlayPageStatusMessage()
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
	content := " ← | x | " + CurrentTab.URI
	writeString(0, 1, content)
	fillLineToEnd(utf8.RuneCountInString(content) - 1, 0)
}

func overlayPageStatusMessage() {
	_, height := screen.Size()
	writeString(0, height - 1, CurrentTab.StatusMessage)
}

