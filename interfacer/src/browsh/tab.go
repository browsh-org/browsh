package browsh

import (
	"encoding/json"
	"fmt"
)

// Tabs is a map of all tab data
var Tabs = make(map[int]*tab)

// CurrentTab is the currently active tab in the TTY browser
var CurrentTab *tab

// Slice of the order in which tabs appear in the tab bar
var tabsOrder []int

// There can be a race condition between the webext sending a tab state update and the
// the tab being deleted, so we need to keep track of all deleted IDs
var tabsDeleted []int

// ID of the tab that was active before the current one
var previouslyVisitedTabID int

// A single tab synced from the browser
type tab struct {
	ID            int    `json:"id"`
	Active        bool   `json:"active"`
	Title         string `json:"title"`
	URI           string `json:"uri"`
	PageState     string `json:"page_state"`
	StatusMessage string `json:"status_message"`
	frame         frame
}

func ResetTabs() {
	Tabs = make(map[int]*tab)
	CurrentTab = nil
	tabsOrder = nil
	tabsDeleted = nil
}

func ensureTabExists(id int) {
	if _, ok := Tabs[id]; !ok {
		newTab(id)
		if isNewEmptyTabActive() {
			removeTab(-1)
		}
	}
}

func isTabPresent(id int) bool {
	_, ok := Tabs[id]
	return ok
}

func newTab(id int) {
	tabsOrder = append(tabsOrder, id)
	Tabs[id] = &tab{
		ID: id,
		frame: frame{
			xScroll: 0,
			yScroll: 0,
		},
	}
}

func restoreTab() {
	sendMessageToWebExtension("/restore_tab")
}

func removeTab(id int) {
	if len(Tabs) == 1 {
		quitBrowsh()
	}
	tabsDeleted = append(tabsDeleted, id)
	sendMessageToWebExtension(fmt.Sprintf("/remove_tab,%d", id))
	nextTab()
	removeTabIDfromTabsOrder(id)
	delete(Tabs, id)
	renderUI()
	renderCurrentTabWindow()
}

// A bit complicated! Just want to remove an integer from a slice whilst retaining
// order :/
func removeTabIDfromTabsOrder(id int) {
	for i := 0; i < len(tabsOrder); i++ {
		if tabsOrder[i] == id {
			tabsOrder = append(tabsOrder[:i], tabsOrder[i+1:]...)
		}
	}
}

func moveTabLeft(id int) {
	// If the tab ID is already completely to the left in the tab order
	// there's nothing to do
	if tabsOrder[0] == id {
		return
	}

	for i, tabID := range tabsOrder {
		if tabID == id {
			tabsOrder[i-1], tabsOrder[i] = tabsOrder[i], tabsOrder[i-1]
			break
		}
	}
}

func moveTabRight(id int) {
	// If the tab ID is already completely to the right in the tab order
	// there's nothing to do
	if tabsOrder[len(tabsOrder)-1] == id {
		return
	}

	for i, tabID := range tabsOrder {
		if tabID == id {
			tabsOrder[i+1], tabsOrder[i] = tabsOrder[i], tabsOrder[i+1]
			break
		}
	}
}

func duplicateTab(id int) {
	sendMessageToWebExtension(fmt.Sprintf("/duplicate_tab,%d", id))
}

// Creating a new tab in the browser without a URI means it won't register with the
// web extension, which means that, come the moment when we actually have a URI for the new
// tab then we can't talk to it to tell it navigate. So we need to only create a real new
// tab when we actually have a URL.
func createNewEmptyTab() {
	createNewEmptyTabWithURI("")
}

func createNewEmptyTabWithURI(URI string) {
	if isNewEmptyTabActive() {
		return
	}
	newTab(-1)
	tab := Tabs[-1]
	tab.Title = "New Tab"
	tab.URI = URI
	tab.Active = true
	CurrentTab = tab
	CurrentTab.frame.resetCells()
	renderUI()
	URLBarFocus(true)
	// Allows for typing directly at the end of URI
	urlInputBox.selectionOff()
	renderCurrentTabWindow()
}

func isNewEmptyTabActive() bool {
	return isTabPresent(-1)
}

func nextTab() {
	for i := 0; i < len(tabsOrder); i++ {
		if tabsOrder[i] == CurrentTab.ID {
			if i+1 == len(tabsOrder) {
				i = 0
			} else {
				i++
			}
			switchToTab(tabsOrder[i])
			break
		}
	}
}

func prevTab() {
	for i := 0; i < len(tabsOrder); i++ {
		if tabsOrder[i] == CurrentTab.ID {
			if i-1 < 0 {
				i = len(tabsOrder) - 1
			} else {
				i--
			}
			switchToTab(tabsOrder[i])
			break
		}
	}
}

func previouslyVisitedTab() {
	if previouslyVisitedTabID == 0 {
		return
	}
	switchToTab(previouslyVisitedTabID)
}

func switchToTab(num int) {
	sendMessageToWebExtension(fmt.Sprintf("/switch_to_tab,%d", num))
	previouslyVisitedTabID = CurrentTab.ID
	CurrentTab = Tabs[num]
	renderUI()
	renderCurrentTabWindow()
}

func isTabPreviouslyDeleted(id int) bool {
	for i := 0; i < len(tabsDeleted); i++ {
		if tabsDeleted[i] == id {
			return true
		}
	}
	return false
}

func parseJSONTabState(jsonString string) {
	var incoming tab
	jsonBytes := []byte(jsonString)
	if err := json.Unmarshal(jsonBytes, &incoming); err != nil {
		Shutdown(err)
	}
	if isTabPreviouslyDeleted(incoming.ID) {
		return
	}
	ensureTabExists(incoming.ID)
	if incoming.Active && !isNewEmptyTabActive() {
		CurrentTab = Tabs[incoming.ID]
	}
	Tabs[incoming.ID].handleStateChange(&incoming)
}

func (t *tab) handleStateChange(incoming *tab) {
	if t.PageState != incoming.PageState {
		// TODO: Take the browser's scroll events as lead
		if incoming.PageState == "page_init" {
			t.frame.yScroll = 0
		}
	}

	// TODO: What's the idiomatic Golang way to do this?
	t.Title = incoming.Title
	t.URI = incoming.URI
	t.PageState = incoming.PageState
	t.StatusMessage = incoming.StatusMessage
}
