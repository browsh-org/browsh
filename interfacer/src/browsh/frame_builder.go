package browsh

import (
	"encoding/json"
	"unicode"
	"fmt"

	"github.com/gdamore/tcell"
)

// Frame is a single frame for the entire DOM. The TTY is merely a window onto a
// region of this frame.
type Frame struct {
	width int
	height int
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

// Text frames received from the webextension are 1 dimensional arrays of strings.
// They are made up of a repeating pattern of 4 items:
// ["RED", "GREEN", "BLUE", "CHARACTER" ...]
func (f *Frame) parseJSONFrameText(jsonString string) {
	if (len(f.pixels) == 0) { f.preFillPixelsWithBlack() }
	var index, textIndex int
	var frame []string
	jsonBytes := []byte(jsonString)
	if err := json.Unmarshal(jsonBytes, &frame); err != nil {
		Shutdown(err)
	}
	if (len(frame) < f.width * (f.height / 2 ) * 4) {
		Log(
			fmt.Sprintf(
				"Not parsing small text frame. Data length: %d, current dimensions: %dx(%d/2)*4=%d",
				len(frame),
				f.width,
				f.height,
				f.width * (f.height / 2 ) * 4))
		return
	}
	f.cells = make([]cell, (f.height / 2) * f.width)
	f.text = make([][]rune, (f.height / 2) * f.width)
	f.textColours = make([]tcell.Color, (f.height / 2) * f.width)
	for y := 0; y < f.height / 2; y++ {
		for x := 0; x < f.width; x++ {
			index = ((f.width * y) + x)
			textIndex = index * 4
			f.textColours[index] = tcell.NewRGBColor(
				toInt32(frame[textIndex + 0]),
				toInt32(frame[textIndex + 1]),
				toInt32(frame[textIndex + 2]),
			)
			f.text[index] = []rune(frame[textIndex + 3])
			f.buildCell(x, y);
		}
	}
}

// This covers the rare situation where a text frame has been sent before any pixel
// data has been populated.
func (f *Frame) preFillPixelsWithBlack() {
	f.pixels = make([][2]tcell.Color, f.height * f.width)
	for i := range f.pixels {
		f.pixels[i] = [2]tcell.Color{
			tcell.NewRGBColor(0, 0, 0),
			tcell.NewRGBColor(0, 0, 0),
		}
	}
}

// Pixel frames received from the webextension are 1 dimensional arrays of strings.
// They are made up of a repeating pattern of 6 items:
// ["FG RED", "FG GREEN", "FG BLUE", "BG RED", "BG GREEN", "BG BLUE" ...]
// TODO: Can these be sent as binary blobs?
func (f *Frame) parseJSONFramePixels(jsonString string) {
	if (len(f.text) == 0) { return }
	var index, indexFg, indexBg, pixelIndexFg, pixelIndexBg int
	var frame []string
	jsonBytes := []byte(jsonString)
	if err := json.Unmarshal(jsonBytes, &frame); err != nil {
		Shutdown(err)
	}
	if (len(frame) != f.width * f.height * 3) {
		Log(
			fmt.Sprintf(
				"Not parsing pixels frame. Data length: %d, current dimensions: %dx%d*3=%d",
				len(frame),
				f.width,
				f.height,
				f.width * f.height * 3))
		return
	}
	f.cells = make([]cell, (f.height / 2) * f.width)
	f.pixels = make([][2]tcell.Color, f.height * f.width)
	for y := 0; y < f.height; y += 2 {
		for x := 0; x < f.width; x++ {
			index = (f.width * (y / 2)) + x
			indexBg = (f.width * y) + x
			indexFg = (f.width * (y + 1)) + x
			pixelIndexBg = indexBg * 3
			pixelIndexFg = indexFg * 3
			pixels := [2]tcell.Color{
				tcell.NewRGBColor(
					toInt32(frame[pixelIndexBg + 0]),
					toInt32(frame[pixelIndexBg + 1]),
					toInt32(frame[pixelIndexBg + 2]),
				),
				tcell.NewRGBColor(
					toInt32(frame[pixelIndexFg + 0]),
					toInt32(frame[pixelIndexFg + 1]),
					toInt32(frame[pixelIndexFg + 2]),
				),
			}
			f.pixels[index] = pixels
			f.buildCell(x, y / 2);
		}
	}
}

// This is where we implement the UTF8 half-block trick.
// This a half-block: "▄", notice how it takes up precisely half a text cell. This
// means that we can get 2 pixel colours from it, the top pixel comes from setting
// the background colour and the bottom pixel comes from setting the foreground
// colour, namely the colour of the text.
func (f *Frame) buildCell(x int, y int) {
	index := ((f.width * y) + x)
	if (index >= len(f.pixels)) { return } // TODO: There must be a better way
	character, fgColour := f.getCharacterAt(index)
	pixelFg, bgColour := f.getPixelColoursAt(index)
	if (isCharacterTransparent(character)) {
		character = []rune("▄")
		fgColour = pixelFg
	}
	f.addCell(index, fgColour, bgColour, character)
}

func (f *Frame) getCharacterAt(index int) ([]rune, tcell.Color) {
	character := f.text[index]
	colour := f.textColours[index]
	return character, colour
}

func (f *Frame) getPixelColoursAt(index int) (tcell.Color, tcell.Color) {
	bgColour := f.pixels[index][0]
	fgColour := f.pixels[index][1]
	return fgColour, bgColour
}

func isCharacterTransparent(character []rune) bool {
	return string(character) == "" || unicode.IsSpace(character[0]);
}

func (f *Frame) addCell(index int, fgColour, bgColour tcell.Color, character []rune) {
	newCell := cell{
		fgColour: fgColour,
		bgColour: bgColour,
		character: character,
	}
	f.cells[index] = newCell
}
