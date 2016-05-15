package main

import (
  "testing"
  "github.com/nsf/termbox-go"
  "github.com/stretchr/testify/assert"
)

func setup() {
  err := termbox.Init()
  if err != nil {
    panic(err)
  }
  initialise()
  hipWidth = 90
  hipHeight = 30
  setCurrentDesktopCoords()
}

func teardown() {
  termbox.Close()
}

func TestPointTL(t *testing.T) {
  assert := assert.New(t)
  curev.MouseX = 30
  curev.MouseY = 10
  setup()

  assert.Equal(533, roundedDesktopX, "Mapped X coord")
  assert.Equal(400, roundedDesktopY, "Mapped Y coord")

  teardown()
}

func TestPointMiddle(t *testing.T) {
  assert := assert.New(t)
  curev.MouseX = 45
  curev.MouseY = 15
  setup()

  assert.Equal(800, roundedDesktopX, "Mapped X coord")
  assert.Equal(600, roundedDesktopY, "Mapped Y coord")

  teardown()
}

func TestPointBR(t *testing.T) {
  assert := assert.New(t)
  curev.MouseX = 60
  curev.MouseY = 20
  setup()

  assert.Equal(1067, roundedDesktopX, "Mapped X coord")
  assert.Equal(800, roundedDesktopY, "Mapped Y coord")

  teardown()
}


func TestZoomIn(t *testing.T) {
  assert := assert.New(t)
  curev.MouseX = 30
  curev.MouseY = 10
  setup()

  trackZoom("in")
  setCurrentDesktopCoords()

  assert.Equal(444, roundedDesktopX, "Mapped X coord")
  assert.Equal(333, roundedDesktopY, "Mapped Y coord")

  teardown()
}

func TestZoomInMiddle(t *testing.T) {
  assert := assert.New(t)
  curev.MouseX = 45
  curev.MouseY = 15
  setup()

  trackZoom("in")
  setCurrentDesktopCoords()

  assert.Equal(800, roundedDesktopX, "Mapped X coord")
  assert.Equal(600, roundedDesktopY, "Mapped Y coord")

  teardown()
}

func TestZoomInOut(t *testing.T) {
  assert := assert.New(t)
  curev.MouseX = 45
  curev.MouseY = 15
  setup()

  trackZoom("in")
  trackZoom("out")
  setCurrentDesktopCoords()

  assert.Equal(800, roundedDesktopX, "Mapped X coord")
  assert.Equal(600, roundedDesktopY, "Mapped Y coord")

  teardown()
}

func TestZoomInEdgeTL(t *testing.T) {
  assert := assert.New(t)
  curev.MouseX = 0
  curev.MouseY = 0
  setup()

  trackZoom("in")
  setCurrentDesktopCoords()

  assert.Equal(0, roundedDesktopX, "Mapped X coord")
  assert.Equal(0, roundedDesktopY, "Mapped Y coord")

  teardown()
}

func TestZoomInEdgeBR(t *testing.T) {
  assert := assert.New(t)
  curev.MouseX = hipWidth
  curev.MouseY = hipHeight
  setup()

  trackZoom("in")
  setCurrentDesktopCoords()

  assert.Equal(int(desktopWidth), roundedDesktopX, "Mapped X coord")
  assert.Equal(int(desktopHeight), roundedDesktopY, "Mapped Y coord")

  teardown()
}
