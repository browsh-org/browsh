package browsh

func (i *inputBox) renderCursor() {
	if !i.isActive {
		return
	}
	if i.isSelection() {
		i.renderSelectionCursor()
	} else {
		i.renderSingleCursor()
	}
}

func (i *inputBox) isSelection() bool {
	return i.selectionStart > 0 || i.selectionEnd > 0
}

func (i *inputBox) renderSingleCursor() {
	x, y := i.getCoordsOfCursor()
	reverseCellColour(x, y)
}

func (i *inputBox) renderSelectionCursor() {
	var x, y int
	textLength := len(i.text)
	for index := 0; index < textLength; index++ {
		x, y = i.getCoordsOfIndex(index)
		if x >= i.selectionStart && x < i.selectionEnd {
			reverseCellColour(x, y)
		}
	}
}

func (i *inputBox) getCoordsOfCursor() (int, int) {
	var index int
	if i.isMultiLine() {
		index = i.xCursor
	} else {
		index = i.textCursor
	}
	return i.getCoordsOfIndex(index)
}

func (i *inputBox) getCoordsOfIndex(index int) (int, int) {
	xFrameOffset := CurrentTab.frame.xScroll
	yFrameOffset := CurrentTab.frame.yScroll - uiHeight
	if urlInputBox.isActive {
		xFrameOffset = 0
		yFrameOffset = 0
	}
	x := (i.X + index) - i.xScroll - xFrameOffset
	y := (i.Y + i.yCursor) - i.yScroll - yFrameOffset
	return x, y
}

func (i *inputBox) cursorLeft() {
	i.xCursor--
	i.textCursor--
	i.updateAllCursors()
}

func (i *inputBox) cursorRight() {
	i.xCursor++
	i.textCursor++
	i.updateAllCursors()
}

func (i *inputBox) cursorUp() {
	i.multiLiner.moveYCursorBy(-1)
	i.updateAllCursors()
}

func (i *inputBox) cursorDown() {
	i.multiLiner.moveYCursorBy(1)
	i.updateAllCursors()
}

func (i *inputBox) cursorBackspace() {
	if len(i.text) == 0 {
		return
	}
	if i.textCursor == 0 {
		return
	}
	start := i.text[:i.textCursor-1]
	end := i.text[i.textCursor:]
	i.text = append(start, end...)
	i.cursorLeft()
	i.sendInputBoxToBrowser()
}

func (i *inputBox) cursorInsertRune(theRune rune) {
	start := i.text[:i.textCursor]
	end := i.text[i.textCursor:]
	endWithRune := append([]rune{theRune}, end...)
	i.text = append(start, endWithRune...)
	i.cursorRight()
	i.sendInputBoxToBrowser()
}

func (i *inputBox) isCursorOverRightEdge() bool {
	return i.textCursor-i.xScroll >= i.Width
}

func (i *inputBox) isCursorOverLeftEdge() bool {
	return i.textCursor-i.xScroll <= -1
}

func (i *inputBox) isCursorOverTopEdge() bool {
	return i.yCursor-i.yScroll <= -1
}

func (i *inputBox) isCursorOverBottomEdge() bool {
	return i.yCursor-i.yScroll > i.Height
}

func (i *inputBox) putCursorAtEnd() {
	i.textCursor = len(urlInputBox.text)
	// TODO: Do for multiline
}

func (i *inputBox) updateAllCursors() {
	i.updateXYCursors()
	if i.isCursorOverLeftEdge() || !i.isBestFit() {
		i.xScrollBy(-1)
	}
	if i.isCursorOverTopEdge() {
		i.yScrollBy(-1)
	}
	if i.isCursorOverRightEdge() {
		i.xScrollBy(1)
	}
	if i.isCursorOverBottomEdge() {
		i.yScrollBy(1)
	}
	i.limitTextCursor()
	i.updateXYCursors()
}

func (i *inputBox) limitTextCursor() {
	if i.textCursor < 0 {
		i.textCursor = 0
	}
	if i.textCursor > len(i.text) {
		i.textCursor = len(i.text)
	}
}

func (i *inputBox) updateXYCursors() {
	if !i.isMultiLine() {
		return
	}
	i.multiLiner.updateCursor()
	i.renderCursor()
}
