package browsh

import (
	"unicode/utf8"
	"encoding/json"

	"github.com/gdamore/tcell"
)

var activeInputBox *inputBox

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
	TagName string `json:"tag_name"`
	Type string `json:"type"`
	FgColour [3]int32 `json:"colour"`
	bgColour [3]int32
	isActive bool
	multiLiner multiLine
	text string
	xCursor int
	yCursor int
	textCursor int
	xScroll int
	yScroll int
	selectionStart int
	selectionEnd int
}

func newInputBox(id string) *inputBox {
	newInputBox := &inputBox{
		ID: id,
	}
	// TODO: Circular reference, what's the proper Golang way to do this?
	newInputBox.multiLiner.inputBox = newInputBox
	return newInputBox
}

// This is used only for the URL input box
func (i *inputBox) renderURLBox() {
	bgRGB := tcell.ColorDefault
	fgRGB := tcell.NewRGBColor(i.FgColour[0], i.FgColour[1], i.FgColour[2])
	style := tcell.StyleDefault
	style = style.Foreground(fgRGB).Background(bgRGB)
	x := i.X
	for _, c := range i.textToDisplay() {
		screen.SetContent(x, i.Y, c, nil, style)
		x++
	}
	i.renderCursor()
	screen.Show()
}

// This is used for all input boxes in the frame
func (i *inputBox) setCells() {
	if i == nil { return }
	i.resetCells()
	x := i.X
	y := i.Y
	lineCount := 0
	for index, c := range i.textToDisplay() {
		if i.isMultiLine() && lineCount < i.yScroll {
			if isLineBreak(string(c)) { lineCount++ }
			continue
		}
		if i.Type == "password" && index != utf8.RuneCountInString(i.text) {
			c = '●'
		}
		i.addCharacterToFrame(x, y, c)
		x++
		if i.isMultiLine() && isLineBreak(string(c)) {
			x = i.X
			y++
			lineCount++
			if lineCount - i.yScroll > i.Height { break }
		}
	}
	screen.Show()
}

func (i *inputBox) resetCells() {
	for y := i.Y; y < i.Height; y++ {
		for x := i.X; x < i.Width; x++ {
			i.addCharacterToFrame(x, y, ' ')
		}
	}
}

func (i *inputBox) addCharacterToFrame(x int, y int, c rune) {
	var (
		index int
		inputBoxCell, existingCell cell
		cellFGColour, cellBGColour tcell.Color
		ok bool
	)
	cellFGColour = tcell.NewRGBColor(i.FgColour[0], i.FgColour[1], i.FgColour[2])
	index = (y * CurrentTab.frame.totalWidth) + x
	if existingCell, ok = CurrentTab.frame.cells.load(index); ok {
		cellBGColour = existingCell.bgColour
	} else {
		return
	}
	inputBoxCell = cell{
		character: []rune{c},
		fgColour: cellFGColour,
		bgColour: cellBGColour,
	}
	CurrentTab.frame.cells.store(index, inputBoxCell)
}

// Different methods are used for containing and displaying overflowed text depending on the
// size of the input box.
func (i *inputBox) isMultiLine() bool {
	if urlInputBox.isActive { return false }
	return i.TagName == "TEXTAREA" || i.Type == "textbox"
}

func (i *inputBox) textToDisplay() []rune {
	if i.isMultiLine() {
		return i.multiLiner.convert()
	}
	return i.textToDisplayForSingleLine()
}

func (i *inputBox) textToDisplayForSingleLine() []rune {
	var textToDisplay string
	index := 0
	for _, c := range i.text + " " {
		if (index >= i.xScroll) {
			textToDisplay += string(c)
		}
		if utf8.RuneCountInString(textToDisplay) >= i.Width { break }
    index++
	}
	return []rune(textToDisplay)
}

func (i *inputBox) lineCount() int {
	return len(i.multiLiner.finalText)
}

func isLineBreak(character string) bool {
	return character == "\n" || character == "\r"
}

func (i *inputBox) sendInputBoxToBrowser() {
	inputBoxMap := map[string]interface{}{
		"id": i.ID,
		"text": i.text,
	}
	marshalled, _ := json.Marshal(inputBoxMap)
	sendMessageToWebExtension("/tab_command,/input_box," + string(marshalled))
}

func (i *inputBox) handleEnterKey() {
	if urlInputBox.isActive {
		if isNewEmptyTabActive() {
			sendMessageToWebExtension("/new_tab," + i.text)
		} else {
			sendMessageToWebExtension("/url_bar," + i.text)
		}
		urlBarFocus(false)
	}
	if i.isMultiLine() {
		i.cursorInsertRune([]rune("\n")[0])
	}
}

func (i *inputBox) selectionOff() {
	i.selectionStart = 0
	i.selectionEnd = 0
}

func (i *inputBox) selectAll() {
	urlInputBox.selectionStart = 0
	urlInputBox.selectionEnd = utf8.RuneCountInString(urlInputBox.text)
}

func (i *inputBox) removeSelectedText() {
	if (i.selectionEnd - i.selectionStart <= 0) { return }
	start := i.text[:i.selectionStart]
	end := i.text[i.selectionEnd:]
	i.text = start + end
	i.textCursor = i.selectionStart
	i.updateXYCursors()
	activeInputBox.selectionOff()
}

func handleInputBoxInput(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyLeft:
		activeInputBox.selectionOff()
		activeInputBox.cursorLeft()
	case tcell.KeyRight:
		activeInputBox.selectionOff()
		activeInputBox.cursorRight()
	case tcell.KeyDown:
		activeInputBox.selectionOff()
		activeInputBox.cursorDown()
	case tcell.KeyUp:
		activeInputBox.selectionOff()
		activeInputBox.cursorUp()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		activeInputBox.removeSelectedText()
		activeInputBox.cursorBackspace()
	case tcell.KeyEnter:
		activeInputBox.removeSelectedText()
		activeInputBox.handleEnterKey()
	case tcell.KeyRune:
		activeInputBox.removeSelectedText()
		activeInputBox.cursorInsertRune(ev.Rune())
	}
	if urlInputBox.isActive {
		renderURLBar()
	} else {
		renderCurrentTabWindow()
	}
}

