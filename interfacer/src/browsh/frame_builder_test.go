package browsh

import (
	"testing"
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFrameBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Frame builder tests")
}

var testFrame Frame

var frameJSONText, _ = json.Marshal([]string{
	"77", "77", "77", "A",
	"101", "101", "101", "b",
	"102", "102", "102", "c",
	"103", "103", "103", "",
})
var frameText = string(frameJSONText)

var frameJSONPixels, _ = json.Marshal([]string{
	"254", "254", "254", "111", "111", "111",
	"1", "1", "1", "2", "2", "2",
	"3", "3", "3", "4", "4", "4",
	"123", "123", "123", "200", "200", "200",
})
var framePixels = string(frameJSONPixels)


var _ = Describe("Frame struct", func() {
	BeforeEach(func() {
		testFrame = Frame{width: 2, height: 4}
		testFrame.parseJSONFrameText(frameText)
	})

	It("should parse JSON frame text", func() {
		Expect(testFrame.cells[0].character[0]).To(Equal('A'))
		Expect(testFrame.cells[1].character[0]).To(Equal('b'))
		Expect(testFrame.cells[2].character[0]).To(Equal('c'))
		Expect(testFrame.cells[3].character[0]).To(Equal('▄'))
	})

	It("should parse JSON pixels (for text-less cells)", func() {
		var r, g, b int32
		testFrame.parseJSONFramePixels(framePixels)
		r, g, b = testFrame.cells[3].fgColour.RGB()
		Expect([3]int32{r, g, b}).To(Equal([3]int32{200, 200, 200}))
		r, g, b = testFrame.cells[3].bgColour.RGB()
		Expect([3]int32{r, g, b}).To(Equal([3]int32{4, 4, 4}))
	})

	It("should parse JSON pixels (using text for foreground)", func() {
		var r, g, b int32
		testFrame.parseJSONFramePixels(framePixels)
		r, g, b = testFrame.cells[0].fgColour.RGB()
		Expect([3]int32{r, g, b}).To(Equal([3]int32{77, 77, 77}))
		r, g, b = testFrame.cells[0].bgColour.RGB()
		Expect([3]int32{r, g, b}).To(Equal([3]int32{254, 254, 254}))
	})
})

