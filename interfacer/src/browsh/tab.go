package browsh

import (
	"encoding/json"
)

var tabs = make(map[int]*tab)
// CurrentTab is the currently active tab in the TTY browser
var CurrentTab *tab

// A single tab synced from the browser
type tab struct {
	ID int `json:"id"`
	Title string `json:"title"`
	URI string `json:"uri"`
	PageState string `json:"page_state"`
	StatusMessage string `json:"status_message"`
	frame frame
}

func ensureTabExists(id int) {
	if _, ok := tabs[id]; !ok {
		tabs[id] = &tab{ID: id}
	}
}

func parseJSONTabState(jsonString string) {
	var incoming tab
	jsonBytes := []byte(jsonString)
	if err := json.Unmarshal(jsonBytes, &incoming); err != nil {
		Shutdown(err)
	}
	ensureTabExists(incoming.ID)
	tabs[incoming.ID].handleStateChange(incoming)
}


func (t *tab) handleStateChange(incoming tab) {
	if (t.PageState != incoming.PageState) {
		// TODO: Take the browser's scroll events as lead
		if (incoming.PageState == "page_init") {
			t.frame.yScroll = 0
		}
	}

	// TODO: What's the idiomatic Golang way to do this?
	t.Title = incoming.Title
	t.URI = incoming.URI
	t.PageState = incoming.PageState
	t.StatusMessage = incoming.StatusMessage

	renderUI()
}
