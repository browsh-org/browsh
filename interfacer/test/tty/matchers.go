package test

import (
	"fmt"
	"time"

	gomegaTypes "github.com/onsi/gomega/types"
)

// BeInFrameAt is a custom matcher that looks for the expected text at the given
// coordinates.
func BeInFrameAt(x, y int) gomegaTypes.GomegaMatcher {
	return &textInFrameMatcher{
		x: x,
		y: y,
		found: "",
	}
}

type textInFrameMatcher struct {
  x int
	y int
	found string
}

func (matcher *textInFrameMatcher) Match(actual interface{}) (success bool, err error) {
	text, _ := actual.(string)
	start := time.Now()
	for time.Since(start) < perTestTimeout {
		matcher.found = GetText(matcher.x, matcher.y, runeCount(text))
		if matcher.found == text { return true, nil }
		time.Sleep(100 * time.Millisecond)
	}
	return false, fmt.Errorf("Timeout. Expected\n\t%#v\nto be in the Browsh frame, but found\n\t%#v", text, matcher.found)
}

func (matcher *textInFrameMatcher) FailureMessage(text interface{}) (message string) {
  return fmt.Sprintf("Expected\n\t%#v\nto equal\n\t%#v", text, matcher.found)
}

func (matcher *textInFrameMatcher) NegatedFailureMessage(text interface{}) (message string) {
  return fmt.Sprintf("Expected\n\t%#v\nnot to equal of\n\t%#v", text, matcher.found)
}

