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
	input.text = text
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

var _ = Describe("Multiline text", func() {
	It("should wrap basic text", func() {
		actual := toMulti("a ab 12 qw 34", 3)
		expected := showWhitespace([]string{
			"a ",
			"ab ",
			"12 ",
			"qw ",
			"34 ",
		})
		Expect(actual).To(Equal(expected))
	})

	It("should wrap text with a word longer than the width limit", func() {
		actual := toMulti("a looooong 12 qw 34", 3)
		expected := showWhitespace([]string{
			"a ",
			"loo",
			"ooo",
			"ng ",
			"12 ",
			"qw ",
			"34 ",
		})
		Expect(actual).To(Equal(expected))
	})

	It("should wrap text lines with multiple words", func() {
		actual := toMulti("some words to make a long sentence with many words on each line", 20)
		expected := showWhitespace([]string{
			"some words to make ",
			"a long sentence ",
			"with many words on ",
			"each line ",
		})
		Expect(actual).To(Equal(expected))
	})

	Describe("Moving the Y cursor", func() {
		It("should move to a line of greater width", func() {
			toMulti(
				`some words !o make ` +
				`a long sent+nce ` +
				`with many words on ` +
				`each line `, 20)
			input.textCursor = 11
			input.multiLiner.moveYCursorBy(1)
			Expect(input.textCursor).To(Equal(30))
			Expect(input.xCursor).To(Equal(11))
			Expect(input.yCursor).To(Equal(1))
		})

		It("should move to a line of smaller width", func() {
			toMulti(
				`some words to make ` +
				`a long sentence ` +
				`with many w!rds on ` +
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
					`some words to make ` +
					"a long \n" +
					`sentence with man! ` +
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

