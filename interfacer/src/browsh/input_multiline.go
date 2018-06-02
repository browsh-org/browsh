package browsh

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type multiLine struct {
	inputBox *inputBox
	index int
	finalText []string
	previousCharacter string
	currentCharacter string
	currentWordish string
	currentLine string
	userAddedLines []int
}

func (m *multiLine) convert() []rune {
	var aRune rune
	m.reset()
	for m.index, aRune = range m.inputBox.text + " " {
		m.previousCharacter = m.currentCharacter
		m.currentCharacter = string(aRune)
		if m.isWordishReady() {
			m.addWordish()
		}
		if m.isInsideWord() {
			// TODO: This sometimes causes a panic :/
			m.currentWordish += m.currentCharacter
		} else {
			m.addWhitespace()
		}
		if m.isFinalCharacter() { m.finish() }
	}
	finalText := []rune(strings.Join(m.finalText, "\n"))
	return finalText
}

func (m *multiLine) reset() {
	m.finalText = nil
	m.previousCharacter = ""
	m.currentCharacter = ""
	m.currentWordish = ""
	m.currentLine = ""
	m.userAddedLines = nil
}

func (m *multiLine) isInsideWord() bool {
	return !m.isCurrentCharacterWhitespace()
}

func (m *multiLine) isPreviousCharacterWhitespace() bool {
	// TODO: Not certain returning `true` for emptiness is best
	if m.previousCharacter == "" { return true }
	runes := []rune(m.previousCharacter)
	if len(runes) == 0 { return true }
	return unicode.IsSpace(runes[0])
}

func (m *multiLine) isCurrentCharacterWhitespace() bool {
	if (len([]rune(m.currentCharacter)) == 0) { return false }
	return unicode.IsSpace([]rune(m.currentCharacter)[0])
}

func (m *multiLine) isWordishReady() bool {
	return m.isNaturalWordEnding() || m.isProjectedLineFull()
}

func (m *multiLine) isNaturalWordEnding() bool {
	return !m.isPreviousCharacterWhitespace() && m.isCurrentCharacterWhitespace()
}

func (m *multiLine) isForcedWordEnding() bool {
	return m.isCurrentWordishFillingLine() && m.isProjectedLineFull()
}

func (m *multiLine) isCurrentWordishFillingLine() bool {
	return m.currentWordishLength() == m.inputBox.Width
}

func (m *multiLine) currentWordishLength() int {
	return utf8.RuneCountInString(m.currentWordish)
}

func (m *multiLine) currentLineLength() int {
	return utf8.RuneCountInString(m.currentLine)
}

func (m *multiLine) isProjectedLineFull() bool {
	return m.currentLineLength() + m.currentWordishLength() >= m.inputBox.Width
}

func (m *multiLine) addWordish() {
	if m.isProjectedLineFull() {
		if m.isForcedWordEnding() {
			m.addLineWithTruncatedWordish()
		} else {
			m.addLineButWrapWord()
		}
	} else {
		m.appendWordToLine()
	}
}

func (m *multiLine) addLineWithTruncatedWordish() {
	m.currentLine += m.currentWordish
	m.currentWordish = ""
	m.addLine()
}

func (m *multiLine) addLineButWrapWord() {
	m.addLine()
	if m.isNaturalWordEnding() {
		m.appendWordToLine()
	}
}

func (m *multiLine) noteUserAddedLineIndex() {
	m.userAddedLines = append(m.userAddedLines, m.lineCount() - 1)
}

func (m *multiLine) appendWordToLine() {
	m.currentLine += m.currentWordish
	m.currentWordish = ""
}

func (m *multiLine) addLine() {
	m.finalText = append(m.finalText, m.currentLine)
	m.currentLine = ""
}

func (m *multiLine) addWhitespace() {
	if m.isNaturalLineBreak() {
		m.addLine()
		m.noteUserAddedLineIndex()
	} else {
		m.currentLine += string(m.currentCharacter)
	}
}

func (m *multiLine) isNaturalLineBreak() bool {
	return isLineBreak(m.currentCharacter)
}

func (m *multiLine) isFinalCharacter() bool {
	return m.index + 1 == utf8.RuneCountInString(m.inputBox.text) + 1
}

func (m *multiLine) lineCount() int {
	return len(m.finalText)
}

func (m *multiLine) finish() {
	m.finalText = append(m.finalText, m.currentLine)
}

func (m *multiLine) updateCursor() {
	xCursor := 0
	yCursor := 0
	index := 0
	m.convert()
	for lineIndex, line := range m.finalText {
		for range line + " " {
			if index == m.inputBox.textCursor {
				m.inputBox.xCursor = xCursor
				m.inputBox.yCursor = yCursor
			}
			xCursor++
			index++
		}
		if !m.isUserAddedLine(lineIndex) { index-- }
		xCursor = 0
		yCursor++
	}
}

func (m *multiLine) moveYCursorBy(magnitude int) {
	if !m.inputBox.isMultiLine() { return }
	m.convert()
	m.updateCursor()
	lastLineIndex := m.lineCount() - 1
	m.inputBox.yCursor += magnitude
	if m.inputBox.yCursor < 0 { m.inputBox.yCursor = 0 }
	if m.inputBox.yCursor > lastLineIndex { m.inputBox.yCursor = lastLineIndex }
	targetLineLength := utf8.RuneCountInString(m.finalText[m.inputBox.yCursor])
	if m.inputBox.xCursor > targetLineLength - 1 {
		m.inputBox.xCursor = targetLineLength
		if !m.isUserAddedLine(m.inputBox.yCursor) { m.inputBox.xCursor-- }
		if m.inputBox.xCursor < 0 { m.inputBox.xCursor = 0 }
	}
	m.convertXYCursorToTextCursor()
}

func (m *multiLine) convertXYCursorToTextCursor() {
	newTextCursor := 0
	for i := 0; i < m.inputBox.yCursor; i++ {
		newTextCursor += utf8.RuneCountInString(m.finalText[i])
		if m.isUserAddedLine(i) {
			newTextCursor++
		}
	}
	newTextCursor += m.inputBox.xCursor
	m.inputBox.textCursor = newTextCursor
	m.updateCursor()
}

func (m *multiLine) isUserAddedLine(index int) bool {
	for i := 0; i < len(m.userAddedLines); i++ {
		if m.userAddedLines[i] == index {
			return true
		}
	}
	return false
}
