package browsh

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMultiLineTextBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Frame builder tests")
}

var input *inputBox

func toMulti(text string, width int) string {
	input = newInputBox("0")
	input.text = []rune(text)
	input.Width = width
	input.TagName = "TEXTAREA"
	textRunes := input.multiLiner.convert()
	raw := string(textRunes)
	raw = visualiseWhitespace(raw)
	return raw
}

func visualiseWhitespace(text string) string {
	text = strings.Replace(text, " ", "_", -1)
	text = strings.Replace(text, "\n", "\\n\n", -1)
	return text
}

func showWhitespace(textArray []string) string {
	text := strings.Join(textArray, "\n")
	return visualiseWhitespace(text)
}

func compareMultilineText(message, text string, width int, textArray []string) {
	It(message, func() {
		actual := toMulti(text, width)
		expected := showWhitespace(textArray)
		Expect(actual).To(Equal(expected))
	})
}

var _ = Describe("Multiline text", func() {
	compareMultilineText(
		"should wrap basic text",
		"a ab 12 qw 34",
		3,
		[]string{
			"a ",
			"ab ",
			"12 ",
			"qw ",
			"34 ",
		})

	compareMultilineText(
		"should wrap text with a word longer than the width limit",
		"a looooong 12 qw 34",
		3,
		[]string{
			"a ",
			"loo",
			"ooo",
			"ng ",
			"12 ",
			"qw ",
			"34 ",
		})

	compareMultilineText(
		"should wrap text lines with multiple words",
		"some words to make a long sentence with many words on each line",
		20,
		[]string{
			"some words to make ",
			"a long sentence ",
			"with many words on ",
			"each line ",
		})

	Describe("Moving the Y cursor", func() {
		It("should move to a line of greater width", func() {
			toMulti(
				`some words !o make `+
					`a long sent+nce `+
					`with many words on `+
					`each line `, 20)
			input.textCursor = 11
			input.multiLiner.moveYCursorBy(1)
			Expect(input.textCursor).To(Equal(30))
			Expect(input.xCursor).To(Equal(11))
			Expect(input.yCursor).To(Equal(1))
		})

		It("should move to a line of smaller width", func() {
			toMulti(
				`some words to make `+
					`a long sentence `+
					`with many w!rds on `+
					`each line+`, 20)
			input.textCursor = 47
			input.multiLiner.moveYCursorBy(1)
			Expect(input.textCursor).To(Equal(64))
			Expect(input.xCursor).To(Equal(10))
			Expect(input.yCursor).To(Equal(3))
		})
		Describe("In text that has user-added line breaks", func() {
			It("should move to a line of smaller width", func() {
				toMulti(
					`some words to make `+
						"a long \n"+
						`sentence with man! `+
						`words+`, 20)
				input.textCursor = 45
				input.multiLiner.moveYCursorBy(1)
				Expect(input.textCursor).To(Equal(52))
				Expect(input.xCursor).To(Equal(6))
				Expect(input.yCursor).To(Equal(3))
			})
		})
	})
})
