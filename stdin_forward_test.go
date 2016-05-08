package main

import (
  "testing"
  "github.com/nsf/termbox-go"
)

func setup() {
  err := termbox.Init()
  if err != nil {
    panic(err)
  }
}

func teardown() {
  termbox.Close()
}

func TestPoint(t *testing.T) {
  setup()
  curev.MouseX = 11
  curev.MouseY = 11
  termWidth = 90
  termWidth = 30
  setCurrentDesktopCoords()
  t.Error(desktopX, desktopY)
  teardown()
}
