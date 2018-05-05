package browsh

import (
	"sync"

	"github.com/gdamore/tcell"
)

// A cell represents an individual TTY cell. An entire representation of the browser
// DOM is stored in a local in-memory "frame". The TTY can then quickly render a region
// of this frame for fast scrolling.
type cell struct {
	character []rune
	fgColour tcell.Color
	bgColour tcell.Color
}

// Both updating a frame and scrolling a frame can happen at the same time, so we need
// to use mutexes.
type threadSafeCellsMap struct {
	sync.RWMutex
	internal map[int]cell
}

func newCellsMap() *threadSafeCellsMap {
	return &threadSafeCellsMap{
		internal: make(map[int]cell),
	}
}

func (m *threadSafeCellsMap) load(key int) (value cell, ok bool) {
	m.RLock()
	result, ok := m.internal[key]
	m.RUnlock()
	return result, ok
}

func (m *threadSafeCellsMap) store(key int, value cell) {
	m.Lock()
	m.internal[key] = value
	m.Unlock()
}
