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
	TagName string `json:"tag_name"`
	Type string `json:"type"`
	FgColour [3]int32 `json:"colour"`
	bgColour [3]int32
	isActive bool
	text string
	textLines []string
	xCursor int
	yCursor int
	xScroll int
	yScroll int
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
	i.setCursor()
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
	for _, c := range i.textToDisplay() {
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

// Different methods are used for containing and displaying overflowed text depending on the
// size of the input box.
func (i *inputBox) isMultiLine() bool {
	return i.TagName == "textarea"
}

func (i *inputBox) textToDisplay() []rune {
	if i.isMultiLine() {
		return i.textToDisplayForMultiLine()
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

func (i *inputBox) textToDisplayForMultiLine() []rune {
	return []rune{'!'}
}

func (i *inputBox) setCursor() {
	xFrameOffset := CurrentTab.frame.xScroll
	yFrameOffset := CurrentTab.frame.yScroll - uiHeight
	if urlInputBox.isActive {
		xFrameOffset = 0
		yFrameOffset = 0
	}
	x := (i.X + i.xCursor) - i.xScroll - xFrameOffset
	y := (i.Y + i.yCursor) - i.yScroll - yFrameOffset
	mainRune, combiningRunes, style, _ := screen.GetContent(x, y)
	style = style.Reverse(true)
	screen.SetContent(x, y, mainRune, combiningRunes, style)
}

func (i *inputBox) cursorLeft() {
	i.xCursor--
	if (i.xCursor - i.xScroll == -1) { i.xScrollBy(-1) }
	i.limitCursor()
}

func (i *inputBox) cursorRight() {
	i.xCursor++
	i.limitCursor()
	if (i.xCursor - i.xScroll == i.Width) { i.xScrollBy(1) }
}

func (i *inputBox) cursorUp() {
	i.yCursor--
	i.limitCursor()
	if (i.yCursor - i.yScroll == 0) { i.yScrollBy(-1) }
}

func (i *inputBox) cursorDown() {
	i.yCursor++
	i.limitCursor()
	if (i.yCursor - i.yScroll >= i.Height) { i.yScrollBy(1) }
}

func (i *inputBox) cursorBackspace() {
	if (utf8.RuneCountInString(i.text) == 0) { return }
	if (i.xCursor == 0) { return }
	start := i.text[:i.xCursor - 1]
	end := i.text[i.xCursor:]
	i.text = start + end
	i.xCursor--
	i.limitCursor()
	i.xScrollBy(-1)
	i.sendInputBoxToBrowser()
}

func (i *inputBox) cursorInsertRune(theRune rune) {
	character := string(theRune)
	start := i.text[:i.xCursor]
	end := i.text[i.xCursor:]
	i.text = start + character + end
	i.xCursor++
	i.limitCursor()
	i.xScrollBy(1)
	i.sendInputBoxToBrowser()
}

func (i *inputBox) xScrollBy(magnitude int) {
	detectionTextWidth := utf8.RuneCountInString(i.text)
	detectionBoxWidth := i.Width
	if !i.isMultiLine() {
		if magnitude < 0 {
			detectionTextWidth++
			detectionBoxWidth -= 2
		}
		isOverflowing := detectionTextWidth >= i.Width
		isCursorAtStartOfBox := i.xCursor - i.xScroll < 0
		isCursorAtEndOfBox := i.xCursor - i.xScroll == detectionBoxWidth
		isCursorAtEdgeOfBox := isCursorAtStartOfBox || isCursorAtEndOfBox
		if isOverflowing && isCursorAtEdgeOfBox {
			i.xScroll += magnitude
		}
	} else {
		i.yScroll += magnitude
	}
	i.limitScroll()
}

func (i *inputBox) yScrollBy(magnitude int) {
	if i.isMultiLine() {
		i.yScroll += magnitude
	}
	i.limitScroll()
}

func (i *inputBox) putCursorAtEnd() {
	i.xCursor = utf8.RuneCountInString(urlInputBox.text)
}

// Not that distinct methods are used for single line and multiline overflow, so their
// respective limit checks never encroach on each other.
func (i *inputBox) limitScroll() {
	if (i.xScroll < 0) {
		i.xScroll = 0
	}
	if (i.xScroll > utf8.RuneCountInString(i.text)) {
		i.xScroll = utf8.RuneCountInString(i.text)
	}
	if (i.yScroll < 0) {
		i.yScroll = 0
	}
	if (i.yScroll > i.textLineCount()) {
		i.yScroll = i.textLineCount()
	}
}

func (i *inputBox) textLineCount() int {
	return len(i.textLines)
}

func (i *inputBox) limitCursor() {
	var upperXLimit int
	if (i.xCursor < 0) {
		i.xCursor = 0
	}
	if (i.isMultiLine()) {
		upperXLimit = i.Width
	} else {
		upperXLimit = utf8.RuneCountInString(i.text)
	}
	if (i.xCursor > upperXLimit) {
		i.xCursor = upperXLimit
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
			if isNewEmptyTabActive() {
				sendMessageToWebExtension("/new_tab," + activeInputBox.text)
			} else {
				sendMessageToWebExtension("/url_bar," + activeInputBox.text)
			}
			urlBarFocus(false)
		}
	case tcell.KeyRune:
		activeInputBox.cursorInsertRune(ev.Rune())
	}
	if urlInputBox.isActive {
		renderURLBar()
	} else {
		renderCurrentTabWindow()
	}
}

