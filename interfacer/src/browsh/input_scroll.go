package browsh

import "unicode/utf8"

func (i *inputBox) xScrollBy(magnitude int) {
	if !i.isMultiLine() {
		i.handleSingleLineScroll(magnitude)
	}
	i.limitScroll()
}

func (i *inputBox) yScrollBy(magnitude int) {
	if i.isMultiLine() {
		i.yScroll += magnitude
	}
	i.limitScroll()
}

func (i *inputBox) handleSingleLineScroll(magnitude int) {
	detectionTextWidth := utf8.RuneCountInString(i.text)
	detectionBoxWidth := i.Width
	if magnitude < 0 {
		detectionTextWidth++
		detectionBoxWidth -= 2
	}
	isOverflowing := detectionTextWidth >= i.Width
	if isOverflowing {
		if i.isCursorAtEdgeOfBox(detectionBoxWidth) || !i.isBestFit() {
			i.xScroll += magnitude
		}
	}
}

func (i *inputBox) isCursorAtEdgeOfBox(detectionBoxWidth int) bool {
	isCursorAtStartOfBox := i.textCursor - i.xScroll < 0
	isCursorAtEndOfBox := i.textCursor - i.xScroll >= detectionBoxWidth
	return isCursorAtStartOfBox || isCursorAtEndOfBox
}

func (i *inputBox) isBestFit() bool {
	lengthOfVisibleText := utf8.RuneCountInString(i.text) - i.xScroll
	return lengthOfVisibleText >= i.Width
}

// Note that distinct methods are used for single line and multiline overflow, so their
// respective limit checks never encroach on each other.
func (i *inputBox) limitScroll() {
	if (i.xScroll < 0) {
		i.xScroll = 0
	}
	if (i.xScroll > utf8.RuneCountInString(i.text)) {
		i.xScroll = utf8.RuneCountInString(i.text)
	}
	if i.isMultiLine() {
		if (i.yScroll < 0) {
			i.yScroll = 0
		}
		if (i.yScroll > i.lineCount() - 1) {
			i.yScroll = (i.lineCount() - 1) - i.Height
		}
	}
}
