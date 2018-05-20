package browsh

import (
	"unicode/utf8"
	"encoding/json"

	"github.com/gdamore/tcell"
)

var (
	activeInputBox *inputBox
)

// A box into which you can enter text. Generally will be forwarded to a standard
// HTML input box in the real browser.
//
// Note that tcell alreay has some ready-made code in its 'views' concept for
// dealing with input areas. However, at the time of writing it wasn't well documented,
// so it was unclear how easy it would be to integrate the requirements of Browsh's
// input boxes - namely overlaying them onto the existing graphics and having them
// scroll in sync.
type inputBox struct {
	ID string `json:"id"`
	X int `json:"x"`
	Y int `json:"y"`
	Width int `json:"width"`
	Height int `json:"height"`
	FgColour [3]int32 `json:"colour"`
	bgColour [3]int32
	isActive bool
	text string
	xCursor int
	yCursor int
}

// This is used only for the URL input box
func (i *inputBox) renderURLBox() {
	bgRGB := tcell.ColorDefault
	fgRGB := tcell.NewRGBColor(i.FgColour[0], i.FgColour[1], i.FgColour[2])
	style := tcell.StyleDefault
	style = style.Foreground(fgRGB).Background(bgRGB)
	x := i.X
	y := i.Y
	for _, c := range i.text + " " {
		if (x - i.X == i.xCursor && y - i.Y == i.yCursor) {
			style = style.Reverse(true)
		} else {
			style = style.Reverse(false)
		}
		screen.SetContent(x, i.Y, c, nil, style)
		x++
	}
	screen.Show()
}

// This is used for all input boxes in the frame
func (i *inputBox) setCells() {
	if i == nil { return }
	var (
		index int
		inputBoxCell, existingCell cell
		cellFGColour, cellBGColour tcell.Color
		ok bool
	)
	x := i.X
	y := i.Y
	cellFGColour = tcell.NewRGBColor(i.FgColour[0], i.FgColour[1], i.FgColour[2])
	for _, c := range i.text + " " {
		y = i.Y
		index = (y * CurrentTab.frame.totalWidth) + x
		if existingCell, ok = CurrentTab.frame.cells.load(index); ok {
			cellBGColour = existingCell.bgColour
		} else {
			continue
		}
		inputBoxCell = cell{
			character: []rune{c},
			fgColour: cellFGColour,
			bgColour: cellBGColour,
		}
		CurrentTab.frame.cells.store(index, inputBoxCell)
		x++
	}
}

func (i *inputBox) setCursor() {
	if urlInputBox.isActive { return }
	x := (i.X + i.xCursor) - CurrentTab.frame.xScroll
	y := (i.Y + i.yCursor) - CurrentTab.frame.yScroll + uiHeight
	mainRune, combiningRunes, style, _ := screen.GetContent(x, y)
	style = style.Reverse(true)
	screen.SetContent(x, y, mainRune, combiningRunes, style)
}

func (i *inputBox) cursorLeft() {
	activeInputBox.xCursor--
	i.limitCursor()
}

func (i *inputBox) cursorRight() {
	activeInputBox.xCursor++
	i.limitCursor()
}

func (i *inputBox) cursorUp() {
	activeInputBox.yCursor--
	i.limitCursor()
}

func (i *inputBox) cursorDown() {
	activeInputBox.yCursor++
	i.limitCursor()
}

func (i *inputBox) cursorBackspace() {
	if (utf8.RuneCountInString(activeInputBox.text) == 0) { return }
	start := activeInputBox.text[:activeInputBox.xCursor - 1]
	end := activeInputBox.text[activeInputBox.xCursor:]
	activeInputBox.text = start + end
	activeInputBox.xCursor--
	i.limitCursor()
}

func (i *inputBox) cursorInsertRune(theRune rune) {
	character := string(theRune)
	start := activeInputBox.text[:activeInputBox.xCursor]
	end := activeInputBox.text[activeInputBox.xCursor:]
	activeInputBox.text = start + character + end
	activeInputBox.xCursor++
	i.limitCursor()
}

func (i *inputBox) putCursorAtEnd() {
	i.xCursor = utf8.RuneCountInString(urlInputBox.text)
}

func (i *inputBox) limitCursor() {
	if (i.xCursor < 0) {
		i.xCursor = 0
	}
	if (i.xCursor > utf8.RuneCountInString(i.text)) {
		i.xCursor = utf8.RuneCountInString(i.text)
	}
	if (i.yCursor < 0) {
		i.yCursor = 0
	}
	if (i.yCursor > i.Height - 1) {
		i.yCursor = i.Height - 1
	}
}

func (i *inputBox) sendInputBoxToBrowser() {
	inputBoxMap := map[string]interface{}{
		"id": i.ID,
		"text": i.text,
	}
	marshalled, _ := json.Marshal(inputBoxMap)
	sendMessageToWebExtension("/tab_command,/input_box," + string(marshalled))
}

func handleInputBoxInput(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyLeft:
		activeInputBox.cursorLeft()
	case tcell.KeyRight:
		activeInputBox.cursorRight()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		activeInputBox.cursorBackspace()
	case tcell.KeyEnter:
		if urlInputBox.isActive {
			sendMessageToWebExtension("/url_bar," + activeInputBox.text)
			urlBarFocus(false)
		}
	case tcell.KeyRune:
		activeInputBox.cursorInsertRune(ev.Rune())
		if !urlInputBox.isActive {
			activeInputBox.sendInputBoxToBrowser()
		}
	}
	if urlInputBox.isActive {
		renderURLBar()
	} else {
		renderCurrentTabWindow()
	}
}

