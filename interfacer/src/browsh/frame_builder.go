package browsh

import (
	"encoding/json"
	"unicode"
	"fmt"

	"github.com/gdamore/tcell"
)

// A frame is a single snapshot of the DOM. The TTY is merely a window onto a
// region of this frame.
type frame struct {
	width int
	height int
	xScroll int
	yScroll int
	pixels [][2]tcell.Color
	text [][]rune
	textColours []tcell.Color
	cells []cell
}

type cell struct {
	character []rune
	fgColour tcell.Color
	bgColour tcell.Color
}

type incomingFrameText struct {
	TabID int `json:"id"`
	Width int `json:"width"`
	Height int `json:"height"`
	Text []string `json:"text"`
	Colours []int32 `json:"colours"`
}

// TODO: Can these be sent as binary blobs?
type incomingFramePixels struct {
	TabID int `json:"id"`
	Width int `json:"width"`
	Height int `json:"height"`
	Colours []int32 `json:"colours"`
}

func (f *frame) rowCount() int {
	return f.height / 2
}

func parseJSONFrameText(jsonString string) {
	var incoming incomingFrameText
	jsonBytes := []byte(jsonString)
	if err := json.Unmarshal(jsonBytes, &incoming); err != nil {
		Shutdown(err)
	}
	ensureTabExists(incoming.TabID)
	tabs[incoming.TabID].frame.buildFrameText(incoming)
}

func (f *frame) buildFrameText(incoming incomingFrameText) {
	f.setup(incoming.Width, incoming.Height)
	if (len(f.pixels) == 0) { f.preFillPixels() }
	if (!f.isIncomingFrameTextValid(incoming)) { return }
	CurrentTab = tabs[incoming.TabID]
	f.populateFrameText(incoming)
}

func parseJSONFramePixels(jsonString string) {
	var incoming incomingFramePixels
	jsonBytes := []byte(jsonString)
	if err := json.Unmarshal(jsonBytes, &incoming); err != nil {
		Shutdown(err)
	}
	ensureTabExists(incoming.TabID)
	if (len(tabs[incoming.TabID].frame.text) == 0) { return }
	tabs[incoming.TabID].frame.buildFramePixels(incoming)
}

func (f *frame) buildFramePixels(incoming incomingFramePixels) {
	f.setup(incoming.Width, incoming.Height)
	if (!f.isIncomingFramePixelsValid(incoming)) { return }
	f.populateFramePixels(incoming)
}

func (f *frame) setup(width, height int) {
	f.width = width
	f.height = height
	f.resetCells()
}

func (f *frame) resetCells() {
	f.cells = make([]cell, (f.rowCount()) * f.width)
}

func (f *frame) isIncomingFrameTextValid(incoming incomingFrameText) bool {
	if (len(incoming.Text) < f.width * (f.rowCount())) {
		Log(
			fmt.Sprintf(
				"Not parsing small text frame. Data length: %d, current dimensions: %dx(%d/2)=%d",
				len(incoming.Text),
				f.width,
				f.height,
				f.width * (f.rowCount())))
		return false
	}
	return true
}

func (f *frame) populateFrameText(incoming incomingFrameText) {
	var index, colourIndex int
	f.text = make([][]rune, (f.rowCount()) * f.width)
	f.textColours = make([]tcell.Color, (f.rowCount()) * f.width)
	for y := 0; y < f.rowCount(); y++ {
		for x := 0; x < f.width; x++ {
			index = ((f.width * y) + x)
			colourIndex = index * 3
			f.textColours[index] = tcell.NewRGBColor(
				incoming.Colours[colourIndex + 0],
				incoming.Colours[colourIndex + 1],
				incoming.Colours[colourIndex + 2],
			)
			f.text[index] = []rune(incoming.Text[index])
			f.buildCell(x, y);
		}
	}
}

// This covers the rare situation where a text frame has been sent before any pixel
// data has been populated.
func (f *frame) preFillPixels() {
	f.pixels = make([][2]tcell.Color, f.height * f.width)
	for i := range f.pixels {
		f.pixels[i] = [2]tcell.Color{
			tcell.NewRGBColor(255, 255, 255),
			tcell.NewRGBColor(255, 255, 255),
		}
	}
}

func (f *frame) populateFramePixels(incoming incomingFramePixels) {
	var index, indexFg, indexBg, pixelIndexFg, pixelIndexBg int
	f.resetCells()
	f.pixels = make([][2]tcell.Color, f.height * f.width)
	data := incoming.Colours
	for y := 0; y < f.height; y += 2 {
		for x := 0; x < f.width; x++ {
			index = (f.width * (y / 2)) + x
			indexBg = (f.width * y) + x
			indexFg = (f.width * (y + 1)) + x
			pixelIndexBg = indexBg * 3
			pixelIndexFg = indexFg * 3
			pixels := [2]tcell.Color{
				tcell.NewRGBColor(
					data[pixelIndexBg + 0],
					data[pixelIndexBg + 1],
					data[pixelIndexBg + 2],
				),
				tcell.NewRGBColor(
					data[pixelIndexFg + 0],
					data[pixelIndexFg + 1],
					data[pixelIndexFg + 2],
				),
			}
			f.pixels[index] = pixels
			f.buildCell(x, y / 2);
		}
	}
}

func (f *frame) isIncomingFramePixelsValid(incoming incomingFramePixels) bool {
	if (len(incoming.Colours) != f.width * f.height * 3) {
		Log(
			fmt.Sprintf(
				"Not parsing pixels frame. Data length: %d, current dimensions: %dx%d*3=%d",
				len(incoming.Colours),
				f.width,
				f.height,
				f.width * f.height * 3))
		return false
	}
	return true
}

// This is where we implement the UTF8 half-block trick.
// This a half-block: "▄", notice how it takes up precisely half a text cell. This
// means that we can get 2 pixel colours from it, the top pixel comes from setting
// the background colour and the bottom pixel comes from setting the foreground
// colour, namely the colour of the text.
func (f *frame) buildCell(x int, y int) {
	index := ((f.width * y) + x)
	character, fgColour := f.getCharacterAt(index)
	pixelFg, bgColour := f.getPixelColoursAt(index)
	if (isCharacterTransparent(character)) {
		character = []rune("▄")
		fgColour = pixelFg
	}
	f.addCell(index, fgColour, bgColour, character)
}

func (f *frame) getCharacterAt(index int) ([]rune, tcell.Color) {
	character := f.text[index]
	colour := f.textColours[index]
	return character, colour
}

func (f *frame) getPixelColoursAt(index int) (tcell.Color, tcell.Color) {
	bgColour := f.pixels[index][0]
	fgColour := f.pixels[index][1]
	return fgColour, bgColour
}

func isCharacterTransparent(character []rune) bool {
	return string(character) == "" || unicode.IsSpace(character[0]);
}

func (f *frame) addCell(index int, fgColour, bgColour tcell.Color, character []rune) {
	newCell := cell{
		fgColour: fgColour,
		bgColour: bgColour,
		character: character,
	}
	f.cells[index] = newCell
}

func (f *frame) limitScroll(height int) {
	maxYScroll := f.rowCount() - height
	if (f.yScroll > maxYScroll) { f.yScroll = maxYScroll }
	if (f.yScroll < 0) { f.yScroll = 0 }
}
