package browsh

import (
	"github.com/gdamore/tcell"
)

var (
	activeInputBox *inputBox
)

// A box into which you can enter text. Generally will be forwarded to a standard
// HTML input box in the real browser.
type inputBox struct {
	ID int `json:"id"`
	X int `json:"x"`
	Y int `json:"y"`
	Width int `json:"width"`
	Height int `json:"height"`
	isActive bool
	text string
}

func handleInputBoxInput(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyLeft:
	case tcell.KeyRight:
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		length := len(activeInputBox.text)
		if (length >= 1) {
			activeInputBox.text = activeInputBox.text[:length - 1]
		}
	case tcell.KeyEnter:
		sendMessageToWebExtension("/url_bar," + activeInputBox.text)
		urlBarFocusToggle()
	}
	character := string(ev.Rune())
	if ev.Key() == tcell.KeyRune {
		activeInputBox.text += character
	}
	renderURLBar()
}

