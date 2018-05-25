package browsh

import (
	"testing"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFrameBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Frame builder tests")
}

var testFrame *frame

func testGetCell(index int) cell {
	result, _ := Tabs[1].frame.cells.load(index)
	return result
}

func testGetCellChar(index int) string {
	result, _ := Tabs[1].frame.cells.load(index)
	return string(result.character[0])
}

func debugCells() {
	fmt.Printf("\n")
	for i := 0; i < 20; i++ {
		if result, ok := Tabs[1].frame.cells.load(i); ok {
			fmt.Printf("%d:%s ", i, string(result.character[0]))
		}
	}
}

var _ = Describe("Frame struct", func() {
	BeforeEach(func() {
		newTab(1)
	})

	Describe("No Offset", func() {
		var frameJSONText = `{
			"meta": {
				"id": 1,
				"sub_left": 0,
				"sub_top": 0,
				"sub_width": 2,
				"sub_height": 4,
				"total_width": 2,
				"total_height": 8
			},
			"text": ["A", "b", "c", ""],
			"colours": [
				77, 77, 77,
				101, 101, 101,
				102, 102, 102,
				103, 103, 103
			]
		}`

		var frameJSONPixels = `{
			"meta": {
				"id": 1,
				"sub_left": 0,
				"sub_top": 0,
				"sub_width": 2,
				"sub_height": 4,
				"total_width": 2,
				"total_height": 8
			},
			"colours": [
				254, 254, 254, 111, 111, 111,
				1, 1, 1, 2, 2, 2,
				3, 3, 3, 4, 4, 4,
				123, 123, 123, 200, 200, 200
			]
		}`

		BeforeEach(func() {
			parseJSONFrameText(frameJSONText)
		})

		It("should parse JSON frame text", func() {
			Expect(testGetCell(0).character[0]).To(Equal('A'))
			Expect(testGetCell(1).character[0]).To(Equal('b'))
			Expect(testGetCell(2).character[0]).To(Equal('c'))
			Expect(testGetCell(3).character[0]).To(Equal('▄'))
		})

		It("should parse JSON pixels (for text-less cells)", func() {
			var r, g, b int32
			parseJSONFramePixels(frameJSONPixels)
			r, g, b = testGetCell(3).fgColour.RGB()
			Expect([3]int32{r, g, b}).To(Equal([3]int32{200, 200, 200}))
			r, g, b = testGetCell(3).bgColour.RGB()
			Expect([3]int32{r, g, b}).To(Equal([3]int32{4, 4, 4}))
		})

		It("should parse JSON pixels (using text for foreground)", func() {
			var r, g, b int32
			parseJSONFramePixels(frameJSONPixels)
			r, g, b = testGetCell(0).fgColour.RGB()
			Expect([3]int32{r, g, b}).To(Equal([3]int32{77, 77, 77}))
			r, g, b = testGetCell(0).bgColour.RGB()
			Expect([3]int32{r, g, b}).To(Equal([3]int32{254, 254, 254}))
		})
	})

	Describe("With Offset", func() {
		var subFrameJSONText = `{
			"meta": {
				"id": 1,
				"sub_left": 1,
				"sub_top": 2,
				"sub_width": 2,
				"sub_height": 4,
				"total_width": 3,
				"total_height": 8
			},
			"text": ["A", "b", "c", ""],
			"colours": [
				77, 77, 77,
				101, 101, 101,
				102, 102, 102,
				103, 103, 103
			]
		}`

		var subFrameJSONPixels = `{
			"meta": {
				"id": 1,
				"sub_left": 1,
				"sub_top": 2,
				"sub_width": 2,
				"sub_height": 4,
				"total_width": 3,
				"total_height": 8
			},
			"colours": [
				254, 254, 254, 111, 111, 111,
				1, 1, 1, 2, 2, 2,
				3, 3, 3, 4, 4, 4,
				123, 123, 123, 200, 200, 200
			]
		}`

		BeforeEach(func() {
			parseJSONFrameText(subFrameJSONText)
		})

		It("should parse text for an offset sub-frame", func() {
			Expect(testGetCell(4).character[0]).To(Equal('A'))
			Expect(testGetCell(5).character[0]).To(Equal('b'))
			Expect(testGetCell(7).character[0]).To(Equal('c'))
			Expect(testGetCell(8).character[0]).To(Equal('▄'))
		})

		It("should parse offset JSON pixels (for text-less cells)", func() {
			var r, g, b int32
			parseJSONFramePixels(subFrameJSONPixels)
			r, g, b = testGetCell(8).fgColour.RGB()
			Expect([3]int32{r, g, b}).To(Equal([3]int32{200, 200, 200}))
			r, g, b = testGetCell(8).bgColour.RGB()
			Expect([3]int32{r, g, b}).To(Equal([3]int32{4, 4, 4}))
		})

		It("should parse offset JSON pixels (using text for foreground)", func() {
			var r, g, b int32
			parseJSONFramePixels(subFrameJSONPixels)
			r, g, b = testGetCell(4).fgColour.RGB()
			Expect([3]int32{r, g, b}).To(Equal([3]int32{77, 77, 77}))
			r, g, b = testGetCell(4).bgColour.RGB()
			Expect([3]int32{r, g, b}).To(Equal([3]int32{254, 254, 254}))
		})

		Describe("Partially overwriting previous cells", func() {
			var overSubFrameJSONText = `{
				"meta": {
					"id": 1,
					"sub_left": 1,
					"sub_top": 4,
					"sub_width": 2,
					"sub_height": 4,
					"total_width": 3,
					"total_height": 8
				},
				"text": ["D", "", "f", ""],
				"colours": [
					78, 78, 78,
					111, 111, 111,
					112, 112, 112,
					113, 113, 113
				]
			}`

			var overSubFrameJSONPixels = `{
				"meta": {
					"id": 1,
					"sub_left": 1,
					"sub_top": 4,
					"sub_width": 2,
					"sub_height": 4,
					"total_width": 3,
					"total_height": 8
				},
				"colours": [
					154, 154, 154, 211, 211, 211,
					11, 11, 11, 12, 12, 12,
					13, 13, 13, 14, 14, 14,
					223, 223, 223, 100, 100, 100
				]
			}`

			It("should partially overwrite text", func() {
				parseJSONFrameText(overSubFrameJSONText)

				// Pre-existing cells
				Expect(testGetCellChar(4)).To(Equal("A"))
				Expect(testGetCellChar(5)).To(Equal("b"))

				// Overwritten cells
				Expect(testGetCellChar(7)).To(Equal("D"))
				Expect(testGetCellChar(8)).To(Equal("▄"))
				Expect(testGetCellChar(10)).To(Equal("f"))
				Expect(testGetCellChar(11)).To(Equal("▄"))
			})

			It("should overwrite colours in text-less cells", func() {
				var r, g, b int32
				parseJSONFramePixels(subFrameJSONPixels)
				parseJSONFrameText(overSubFrameJSONText)
				parseJSONFramePixels(overSubFrameJSONPixels)

				overwrittenCell := 8
				r, g, b = testGetCell(overwrittenCell).fgColour.RGB()
				Expect([3]int32{r, g, b}).To(Equal([3]int32{12, 12, 12}))
				r, g, b = testGetCell(overwrittenCell).bgColour.RGB()
				Expect([3]int32{r, g, b}).To(Equal([3]int32{211, 211, 211}))
			})

			It("should partially overwrite text colours", func() {
				var r, g, b int32
				parseJSONFramePixels(subFrameJSONPixels)
				parseJSONFrameText(overSubFrameJSONText)
				parseJSONFramePixels(overSubFrameJSONPixels)

				preExistingCell := 4
				r, g, b = testGetCell(preExistingCell).fgColour.RGB()
				Expect([3]int32{r, g, b}).To(Equal([3]int32{77, 77, 77}))
				r, g, b = testGetCell(preExistingCell).bgColour.RGB()
				Expect([3]int32{r, g, b}).To(Equal([3]int32{254, 254, 254}))

				overwrittenCell := 7
				r, g, b = testGetCell(overwrittenCell).fgColour.RGB()
				Expect([3]int32{r, g, b}).To(Equal([3]int32{78, 78, 78}))
				r, g, b = testGetCell(overwrittenCell).bgColour.RGB()
				Expect([3]int32{r, g, b}).To(Equal([3]int32{154, 154, 154}))
			})
		})
	})
})

