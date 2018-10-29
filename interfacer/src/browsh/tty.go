package browsh

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/go-errors/errors"
	"github.com/spf13/viper"
)

type VimMode int

const (
	NormalMode VimMode = iota + 1
	InsertMode
	FindMode
	LinkMode
	LinkModeNewTab
	LinkModeCopy
	WaitMode
	VisualMode
	CaretMode
	MakeMarkMode
	GotoMarkMode
)

type mark struct {
	tabID   int
	URI     string
	xScroll int
	yScroll int
}

type Coordinate struct {
	X, Y int
}

type HintRect struct {
	Bottom int    `json:"bottom"`
	Top    int    `json:"top"`
	Left   int    `json:"left"`
	Right  int    `json:"right"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Href   string `json:"href"`
}

var (
	screen   tcell.Screen
	uiHeight = 2
	// IsMonochromeMode decides whether to render the TTY in full colour or monochrome
	IsMonochromeMode = false
	// vimMode
	vimMode             = NormalMode
	vimCommandsBindings = make(map[string]string)
	keyEvents           = make([]*tcell.EventKey, 11)
	//runeTime            = make(map[rune]time.Time)
	//lastRune            rune
	waitModeStartTime time.Time
	findText          string
	// marks
	globalMarkMap = make(map[rune]*mark)
	localMarkMap  = make(map[int]map[rune]*mark)
	// position coordinate for caret mode
	caretPos Coordinate
	// for link modes
	linkText                 string
	linkHintRects            []HintRect
	linkHintKeys             = "asdfwerxcv"
	linkHints                []string
	linkHintsToRects         = make(map[string]*HintRect)
	linkModeWithHints        = true
	linkHintWriteStringCalls *func()
)

func init() {
	setupLinkHints()
}

func setupLinkHints() {
	lowerAlpha := "abcdefghijklmnopqrstuvwxyz"
	missingAlpha := lowerAlpha

	// Use linkHintKeys first to generate link hints
	for i := 0; i < len(linkHintKeys); i++ {
		for j := 0; j < len(linkHintKeys); j++ {
			linkHints = append(linkHints, string(linkHintKeys[i])+string(linkHintKeys[j]))
		}
		missingAlpha = strings.Replace(missingAlpha, string(linkHintKeys[i]), "", -1)
	}

	// missingAlpha contains all keys that aren't in linkHintKeys
	// we use this to generate the last link hint key combinations,
	// so this will only be used when we run out of linkHintKeys based
	// link hint key combinations.
	for i := 0; i < len(missingAlpha); i++ {
		for j := 0; j < len(lowerAlpha); j++ {
			linkHints = append(linkHints, string(missingAlpha[i])+string(lowerAlpha[j]))
		}
	}
}

func setupTcell() {
	var err error
	if err = screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	screen.EnableMouse()
	screen.Clear()
}

func sendTtySize() {
	width, height := screen.Size()
	urlInputBox.Width = width
	sendMessageToWebExtension(fmt.Sprintf("/tty_size,%d,%d", width, height))
}

// This is basically a proxy that listens to STDIN and forwards all relevant input
// from the user to the webextension. So keyboard, mouse, terminal resizes, etc.
func readStdin() {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			handleUserKeyPress(ev)
		case *tcell.EventResize:
			handleTTYResize()
		case *tcell.EventMouse:
			handleMouseEvent(ev)
		}
	}
}

func handleUserKeyPress(ev *tcell.EventKey) {
	if CurrentTab == nil {
		if ev.Key() == tcell.KeyCtrlQ {
			quitBrowsh()
		}
		return
	}
	switch ev.Key() {
	case tcell.KeyCtrlQ:
		quitBrowsh()
	case tcell.KeyCtrlL:
		urlBarFocusToggle()
	case tcell.KeyCtrlT:
		createNewEmptyTab()
	case tcell.KeyCtrlW:
		removeTab(CurrentTab.ID)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if activeInputBox == nil {
			sendMessageToWebExtension("/tab_command,/history_back")
		}
	}
	if ev.Rune() == 'm' && ev.Modifiers() == 4 {
		toggleMonochromeMode()
	}
	if ev.Key() == 279 && ev.Modifiers() == 0 {
		// F1 key
		openHelpTab()
	}
	if isKey("tty.keys.next-tab", ev) {
		nextTab()
	}
	if !urlInputBox.isActive {
		forwardKeyPress(ev)
	}
	if activeInputBox != nil {
		handleInputBoxInput(ev)
	} else {
		handleVimControl(ev)
		handleScrolling(ev) // TODO: shouldn't you be able to still use mouse scrolling?
	}
}

func isKey(userKey string, ev *tcell.EventKey) bool {
	key := viper.GetStringSlice(userKey)
	runeMatch := []rune(key[0])[0] == ev.Rune()
	intKey, _ := strconv.Atoi(key[1])
	keyCodeMatch := intKey == int(ev.Key())
	modifierKey, _ := strconv.Atoi(key[2])
	modifierMatch := modifierKey == int(ev.Modifiers())
	return runeMatch && keyCodeMatch && modifierMatch
}

func quitBrowsh() {
	if !viper.GetBool("firefox.use-existing") {
		quitFirefox()
	}
	Shutdown(errors.New("normal"))
}

func toggleMonochromeMode() {
	IsMonochromeMode = !IsMonochromeMode
}

func openHelpTab() {
	sendMessageToWebExtension("/new_tab,https://www.brow.sh/docs/introduction/")
}

func forwardKeyPress(ev *tcell.EventKey) {
	if isMultiLineEnter(ev) {
		return
	}
	eventMap := map[string]interface{}{
		"key":  int(ev.Key()),
		"char": string(ev.Rune()),
		"mod":  int(ev.Modifiers()),
	}
	marshalled, _ := json.Marshal(eventMap)
	sendMessageToWebExtension("/stdin," + string(marshalled))
}

// Allow user to use ENTER key without triggering submission on multiline input
// boxes.
func isMultiLineEnter(ev *tcell.EventKey) bool {
	if activeInputBox == nil {
		return false
	}
	return activeInputBox.isMultiLine() && ev.Key() == 13 && ev.Modifiers() != 4
}

// Moves the caret in CaretMode.
// testFunc is a function that tests for the reaching of the boundaries of the given axis.
// The axis of motion is decided by giving a reference to caretPos.X or caretPos.Y as valRef.
// The step size and direction is given by the value of step.
func moveVimCaret(testFunc func() bool, valRef *int, step int) {
	var prevCell, nextCell, nextNextCell cell
	var r rune
	hasNextNextCell := false

	for testFunc() {
		prevCell = getCell(caretPos.X, caretPos.Y-uiHeight)
		*valRef += step
		nextCell = getCell(caretPos.X, caretPos.Y-uiHeight)

		if testFunc() {
			*valRef += step
			nextNextCell = getCell(caretPos.X, caretPos.Y-uiHeight)
			*valRef -= step
			hasNextNextCell = true
		} else {
			hasNextNextCell = false
		}

		r = nextCell.character[0]
		// Check if the next cell is different in any way
		if !reflect.DeepEqual(prevCell, nextCell) {
			if hasNextNextCell {
				// This condition should apply to the spaces between words and the like
				// Checking with unicode.isSpace() didn't give correct results for some reason
				// TODO: find out what that reason is and improve this
				if !unicode.IsLetter(r) && unicode.IsLetter(nextNextCell.character[0]) {
					continue
				}
				// If the upcoming cell is deeply equal we can continue to go forward
				if reflect.DeepEqual(nextCell, nextNextCell) {
					continue
				}
			}
			// This cell is different and other conditions for continuing don't apply
			// therefore we stop going forward.
			break
		}
	}
}

// TODO: This fails if the tab with mark.tabID doesn't exist anymore it should recreate said tab, then go to the mark's URL and position
func gotoMark(mark *mark) {
	if CurrentTab.ID != mark.tabID {
		ensureTabExists(mark.tabID)
		switchToTab(mark.tabID)
	}
	if CurrentTab.URI != mark.URI {
		sendMessageToWebExtension("/tab_command,/url," + mark.URI)
		//sleep?
	}
	doScrollAbsolute(mark.xScroll, mark.yScroll)
}

// Make a mark at the current position in the current tab
func makeMark() *mark {
	return &mark{CurrentTab.ID, CurrentTab.URI, CurrentTab.frame.xScroll, CurrentTab.frame.yScroll}
}

func generateLeftClickYHack(x, y int, yHack bool) {
	newMouseEvent := tcell.NewEventMouse(x, y+uiHeight, tcell.Button1, 0)
	handleMouseEventYHack(newMouseEvent, yHack)
	time.Sleep(time.Millisecond * 100)
	newMouseEvent = tcell.NewEventMouse(x, y+uiHeight, 0, 0)
	handleMouseEventYHack(newMouseEvent, yHack)
}

func generateLeftClick(x, y int) {
	generateLeftClickYHack(x, y, false)
}

// TODO: This isn't working for opening new tabs.
func generateMiddleClick(x, y int) {
	newMouseEvent := tcell.NewEventMouse(x, y+uiHeight, tcell.Button2, 0)
	handleMouseEvent(newMouseEvent)
	time.Sleep(time.Millisecond * 100)
	newMouseEvent = tcell.NewEventMouse(x, y+uiHeight, 0, 0)
	handleMouseEvent(newMouseEvent)
}

func goIntoWaitMode() {
	vimMode = WaitMode
	waitModeStartTime = time.Now()
}

func updateLinkHintDisplay() {
	linkHintsToRects = make(map[string]*HintRect)
	lh := len(linkHintRects)
	var ht string
	// List of closures
	var fc []*func()

	for i, r := range linkHintRects {
		// When the number of link hints is small enough
		// using just one key for individual link hints suffices.
		// Otherwise use the prepared link hint key combinations.
		if lh <= len(linkHintKeys) {
			ht = string(linkHintKeys[i])
		} else {
			ht = linkHints[i]
		}
		// Add the key combination ht to the linkHintsToRects map.
		// When the user presses it, we can easily lookup the
		// link hint properties associated with it.
		linkHintsToRects[ht] = &linkHintRects[i]

		// When the first key got hit,
		// shorten the link hints accordingly
		offsetLeft := 0
		if strings.HasPrefix(ht, linkText) {
			ht = ht[len(linkText):len(ht)]
			offsetLeft = len(linkText)
		}

		// Make copies of parameter values
		rLeftCopy, rTopCopy, htCopy := r.Left, r.Top, ht

		// Link hints are in upper case in new tab mode
		if vimMode == LinkModeNewTab {
			htCopy = strings.ToUpper(htCopy)
		}

		// Create closure
		f := func() {
			writeString(rLeftCopy+offsetLeft, rTopCopy+uiHeight, htCopy, tcell.StyleDefault)
		}
		fc = append(fc, &f)
	}
	// Create closure that calls the other closures
	ff := func() {
		for _, f := range fc {
			(*f)()
		}
	}
	linkHintWriteStringCalls = &ff
}

func eraseLinkHints() {
	linkText = ""
	linkHintWriteStringCalls = nil
	linkHintsToRects = make(map[string]*HintRect)
	linkHintRects = nil
}

func isNormalModeKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyESC {
		return true
	}
	return false
}

func handleVimControl(ev *tcell.EventKey) {
	var lastRune rune
	command := ""

	if len(keyEvents) > 0 && keyEvents[0] != nil {
		lastRune = keyEvents[len(keyEvents)-1].Rune()
	}

	keyEvents = append(keyEvents, ev)
	if len(keyEvents) > 10 {
		keyEvents = keyEvents[1:]
	}

	keyCombination := string(lastRune) + string(ev.Rune())

	switch vimMode {
	case WaitMode:
		if time.Since(waitModeStartTime) < time.Millisecond*1000 {
			return
		}
		vimMode = NormalMode
		fallthrough
	case NormalMode:
		command = vimCommandsBindings[keyCombination]
		if len(command) == 0 {
			keyCombination := string(ev.Rune())
			command = vimCommandsBindings[keyCombination]
		}
	case InsertMode:
		if isNormalModeKey(ev) {
			command = "normalMode"
		}
	case VisualMode:
		if isNormalModeKey(ev) {
			command = "normalMode"
		} else {
			if ev.Rune() == 'c' {
				command = "caretMode"
			}
			if ev.Rune() == 'o' {
				//swap cursor position begin->end or end->begin
			}
			if ev.Rune() == 'y' {
				//clipboard
			}
		}
	case CaretMode:
		if isNormalModeKey(ev) {
			command = "normalMode"
		} else {
			switch ev.Key() {
			case tcell.KeyEnter:
				generateLeftClick(caretPos.X, caretPos.Y-uiHeight)
			}
			switch ev.Rune() {
			case 'v':
				command = "visualMode"
			case 'h':
				moveVimCaret(func() bool { return caretPos.X > 0 }, &caretPos.X, -1)
			case 'l':
				width, _ := screen.Size()
				moveVimCaret(func() bool { return caretPos.X < width }, &caretPos.X, 1)
			case 'k':
				_, height := screen.Size()
				moveVimCaret(func() bool { return caretPos.Y >= uiHeight }, &caretPos.Y, -1)
				if caretPos.Y < uiHeight {
					command = "scrollHalfPageUp"
					if CurrentTab.frame.yScroll == 0 {
						caretPos.Y = uiHeight
					} else {
						caretPos.Y += (height - uiHeight) / 2
					}
				}
			case 'j':
				_, height := screen.Size()
				moveVimCaret(func() bool { return caretPos.Y <= height-uiHeight }, &caretPos.Y, 1)
				if caretPos.Y > height-uiHeight {
					command = "scrollHalfPageDown"
					caretPos.Y -= (height - uiHeight) / 2
				}
			}
		}
	case MakeMarkMode:
		if unicode.IsLower(ev.Rune()) {
			if localMarkMap[CurrentTab.ID] == nil {
				localMarkMap[CurrentTab.ID] = make(map[rune]*mark)
			}
			localMarkMap[CurrentTab.ID][ev.Rune()] = makeMark()
		} else if unicode.IsUpper(ev.Rune()) {
			globalMarkMap[ev.Rune()] = makeMark()
		}

		command = "normalMode"
	case GotoMarkMode:
		if mark, ok := globalMarkMap[ev.Rune()]; ok {
			gotoMark(mark)
		} else if m, ok := localMarkMap[CurrentTab.ID]; unicode.IsLower(ev.Rune()) && ok {
			if mark, ok := m[ev.Rune()]; ok {
				gotoMark(mark)
			}
		}

		command = "normalMode"
	case FindMode:
		if isNormalModeKey(ev) {
			command = "normalMode"
			findText = ""
		} else {
			if ev.Key() == tcell.KeyEnter {
				vimMode = NormalMode
				command = "findText"
				break
			}
			if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				if len(findText) > 0 {
					findText = findText[:len(findText)-1]
				}
			} else {
				findText += string(ev.Rune())
			}
		}
	case LinkMode, LinkModeNewTab, LinkModeCopy:
		if isNormalModeKey(ev) {
			command = "normalMode"
			eraseLinkHints()
		} else {
			linkText += string(ev.Rune())
			updateLinkHintDisplay()
			if linkModeWithHints {
				if r, ok := linkHintsToRects[linkText]; ok {
					if r != nil {
						switch vimMode {
						case LinkMode:
							if (*r).Height == 2 {
								generateLeftClickYHack((*r).Left, (*r).Top, true)
							} else {
								generateLeftClick((*r).Left, (*r).Top)
							}
						case LinkModeNewTab:
							//generateMiddleClick(r.Left, r.Top)
							sendMessageToWebExtension("/new_tab," + r.Href)
						case LinkModeCopy:
							clipboard.WriteAll(r.Href)
						}
						goIntoWaitMode()
						eraseLinkHints()
					}
				}
			} else {
				coords := findAndHighlightTextOnScreen(linkText)
				if len(coords) == 1 {
					goIntoWaitMode()

					if vimMode == LinkModeNewTab {
						generateMiddleClick(coords[0].X, coords[0].Y)
					} else {
						generateLeftClick(coords[0].X, coords[0].Y)
					}
					linkText = ""
					return
				} else if len(coords) == 0 {
					vimMode = NormalMode
					linkText = ""
					return
				}
			}
		}
	}

	if len(command) > 0 {
		executeVimCommand(command)
	}
}

func executeVimCommand(command string) {
	switch command {
	case "urlUp":
		sendMessageToWebExtension("/tab_command,/url_up")
	case "urlRoot":
		sendMessageToWebExtension("/tab_command,/url_root")
	case "scrollToTop":
		doScroll(0, -CurrentTab.frame.domRowCount())
	case "scrollToBottom":
		doScroll(0, CurrentTab.frame.domRowCount())
	case "scrollUp":
		doScroll(0, -1)
	case "scrollDown":
		doScroll(0, 1)
	case "scrollLeft":
		doScroll(-1, 0)
	case "scrollRight":
		doScroll(1, 0)
	case "editURL":
		urlBarFocusToggle()
	case "firstTab":
		switchToTab(tabsOrder[0])
	case "lastTab":
		switchToTab(tabsOrder[len(tabsOrder)-1])
	case "scrollHalfPageDown":
		_, height := screen.Size()
		doScroll(0, (height-uiHeight)/2)
	case "scrollHalfPageUp":
		_, height := screen.Size()
		doScroll(0, -((height - uiHeight) / 2))
	case "historyBack":
		sendMessageToWebExtension("/tab_command,/history_back")
	case "historyForward":
		sendMessageToWebExtension("/tab_command,/history_forward")
	case "reload":
		sendMessageToWebExtension("/tab_command,/reload")
	case "prevTab":
		prevTab()
	case "nextTab":
		nextTab()
	case "previouslyVisitedTab":
		previouslyVisitedTab()
	case "newTab":
		createNewEmptyTab()
	case "removeTab":
		removeTab(CurrentTab.ID)
	case "restoreTab":
		restoreTab()
	case "duplicateTab":
		duplicateTab(CurrentTab.ID)
	case "moveTabLeft":
		moveTabLeft(CurrentTab.ID)
	case "moveTabRight":
		moveTabRight(CurrentTab.ID)
	case "copyURL":
		clipboard.WriteAll(CurrentTab.URI)
	case "openClipboardURL":
		URI, _ := clipboard.ReadAll()
		sendMessageToWebExtension("/tab_command,/url," + URI)
	case "openClipboardURLInNewTab":
		URI, _ := clipboard.ReadAll()
		sendMessageToWebExtension("/new_tab," + URI)
	case "focusFirstTextInput":
		sendMessageToWebExtension("/tab_command,/focus_first_text_input")
	case "followLinkLabeledNext":
		sendMessageToWebExtension("/tab_command,/follow_link_labeled_next")
	case "followLinkLabeledPrevious":
		sendMessageToWebExtension("/tab_command,/follow_link_labeled_previous")
	case "viewHelp":
		sendMessageToWebExtension("/new_tab,https://www.brow.sh/docs/keybindings/")
	case "openLinkInCurrentTab":
		vimMode = LinkMode
		sendMessageToWebExtension("/tab_command,/get_clickable_hints")
		eraseLinkHints()
	case "openLinkInNewTab":
		vimMode = LinkModeNewTab
		sendMessageToWebExtension("/tab_command,/get_link_hints")
		eraseLinkHints()
	case "copyLinkURL":
		vimMode = LinkModeCopy
		sendMessageToWebExtension("/tab_command,/get_link_hints")
		eraseLinkHints()
	case "findText":
		fallthrough
	case "findNext":
		sendMessageToWebExtension("/tab_command,/find_next," + findText)
	case "findPrevious":
		sendMessageToWebExtension("/tab_command,/find_previous," + findText)
	case "makeMark":
		vimMode = MakeMarkMode
	case "gotoMark":
		vimMode = GotoMarkMode
	case "insertMode":
		vimMode = InsertMode
	case "findMode":
		vimMode = FindMode
	case "normalMode":
		vimMode = NormalMode
	case "visualMode":
		vimMode = VisualMode
	case "caretMode":
		vimMode = CaretMode
		width, height := screen.Size()
		caretPos.X, caretPos.Y = width/2, height/2
	}
}

func doScroll(relX int, relY int) {
	doScrollAbsolute(CurrentTab.frame.xScroll+relX, CurrentTab.frame.yScroll+relY)
}

func doScrollAbsolute(absX int, absY int) {
	yScrollOriginal := CurrentTab.frame.yScroll
	_, height := screen.Size()
	height -= uiHeight

	CurrentTab.frame.yScroll = absY
	CurrentTab.frame.xScroll = absX

	CurrentTab.frame.limitScroll(height)
	sendMessageToWebExtension(
		fmt.Sprintf("/tab_command,/scroll_status,%d,%d",
			CurrentTab.frame.xScroll, CurrentTab.frame.yScroll*2))
	if CurrentTab.frame.yScroll != yScrollOriginal {
		renderCurrentTabWindow()
	}
}

func handleScrolling(ev *tcell.EventKey) {
	_, height := screen.Size()
	height -= uiHeight
	if ev.Key() == tcell.KeyUp {
		doScroll(0, -2)
	}
	if ev.Key() == tcell.KeyDown {
		doScroll(0, 2)
	}
	if ev.Key() == tcell.KeyPgUp {
		doScroll(0, -height)
	}
	if ev.Key() == tcell.KeyPgDn {
		doScroll(0, height)
	}
}

func handleMouseEventYHack(ev *tcell.EventMouse, yHack bool) {
	if CurrentTab == nil {
		return
	}
	x, y := ev.Position()
	xInFrame := x + CurrentTab.frame.xScroll
	yInFrame := y - uiHeight + CurrentTab.frame.yScroll
	button := ev.Buttons()
	if button == tcell.WheelUp || button == tcell.WheelDown {
		handleMouseScroll(button)
	}
	if button == 1 {
		CurrentTab.frame.maybeFocusInputBox(xInFrame, yInFrame)
	}
	eventMap := map[string]interface{}{
		"button":    int(button),
		"mouse_x":   int(xInFrame),
		"mouse_y":   int(yInFrame),
		"modifiers": int(ev.Modifiers()),
	}
	if yHack {
		eventMap["y_hack"] = true
	}
	marshalled, _ := json.Marshal(eventMap)
	sendMessageToWebExtension("/stdin," + string(marshalled))
}

func handleMouseEvent(ev *tcell.EventMouse) {
	handleMouseEventYHack(ev, false)
}

func handleMouseScroll(scrollType tcell.ButtonMask) {
	_, height := screen.Size()
	height -= uiHeight
	if scrollType == tcell.WheelUp {
		doScroll(0, -1)
	} else if scrollType == tcell.WheelDown {
		doScroll(0, 1)
	}
}

func handleTTYResize() {
	width, _ := screen.Size()
	urlInputBox.Width = width
	screen.Sync()
	sendTtySize()
}

// Tcell uses a buffer to collect screen updates on, it only actually sends
// ANSI rendering commands to the terminal when we tell it to. And even then it
// will try to minimise rendering commands by only rendering parts of the terminal
// that have changed.
func renderCurrentTabWindow() {
	var currentCell cell
	var styling = tcell.StyleDefault
	var runeChars []rune
	width, height := screen.Size()
	if CurrentTab == nil || CurrentTab.frame.cells == nil {
		return
	}
	CurrentTab.frame.overlayInputBoxContent()
	for y := 0; y < height-uiHeight; y++ {
		for x := 0; x < width; x++ {
			currentCell = getCell(x, y)
			runeChars = currentCell.character
			// TODO: do this is in isCharacterTransparent()
			if len(runeChars) == 0 {
				continue
			}
			if IsMonochromeMode {
				styling = styling.Foreground(tcell.ColorWhite)
				styling = styling.Background(tcell.ColorBlack)
				if runeChars[0] == '▄' {
					runeChars[0] = ' '
				}
			} else {
				styling = styling.Foreground(currentCell.fgColour)
				styling = styling.Background(currentCell.bgColour)
			}
			screen.SetCell(x, y+uiHeight, styling, runeChars[0])
		}
	}
	if activeInputBox != nil {
		activeInputBox.renderCursor()
	}
	overlayPageStatusMessage()
	overlayVimMode()
	overlayCallToSupport()
	screen.Show()
}

func searchScreenForText(text string) []Coordinate {
	var offsets = make([]Coordinate, 0)
	var splitString []string
	var r rune
	var s string
	width, height := screen.Size()
	screenText := ""
	index := 0

	for y := 0; y < height-uiHeight; y++ {
		screenText = ""
		for x := 0; x < width; x++ {
			r = getCell(x, y).character[0]
			s = string(r)
			if len(s) == 0 || len(s) > 1 {
				screenText += " "
			} else {
				screenText += string(getCell(x, y).character[0])
			}
		}
		index = 0
		splitString = strings.Split(strings.ToLower(screenText), strings.ToLower(text))
		for _, s := range splitString {
			if index+len(s) >= width {
				break
			}

			offsets = append(offsets, Coordinate{index + len(s), y})
			index += len(s) + len(text)
		}
	}
	return offsets
}

func findAndHighlightTextOnScreen(text string) []Coordinate {
	var x, y int
	var styling = tcell.StyleDefault

	offsets := searchScreenForText(text)
	for _, offset := range offsets {
		y = offset.Y
		x = offset.X
		for z := 0; z < len(text); z++ {
			screen.SetContent(x+z, y+uiHeight, rune(text[z]), nil, styling)
		}
	}
	screen.Show()
	return offsets
}

// Parse incoming link hints
func parseJSONLinkHints(jsonString string) {
	jsonBytes := []byte(jsonString)
	if err := json.Unmarshal(jsonBytes, &linkHintRects); err != nil {
		Shutdown(err)
	}

	// Optimize link hint positions
	for i := 0; i < len(linkHintRects); i++ {
		r := &linkHintRects[i]

		// For links that are more than one line high
		// we want to position the link hint in the vertical middle
		if r.Height > 2 {
			if r.Height%2 == 0 {
				r.Top += r.Height / 2
			} else {
				r.Top += r.Height/2 - 1
			}
		}

		// For links that are more one character long we try to move
		// the link hint two characters to the right, if possible.
		if r.Width > 1 {
			o := r.Left
			r.Left += r.Width/2 - 1
			if r.Left > o+2 {
				r.Left = o + 2
			}
		}
	}

	Log("Received parseJSONLinkHint")
	// This is where the display of actual link hints is prepared
	updateLinkHintDisplay()
}

func getCell(x, y int) cell {
	var currentCell cell
	var ok bool
	frame := &CurrentTab.frame
	index := ((y + frame.yScroll) * frame.totalWidth) + (x + frame.xScroll)
	if currentCell, ok = frame.cells.load(index); !ok {
		fgColour, bgColour := getHatchedCellColours(x)
		currentCell = cell{
			fgColour:  fgColour,
			bgColour:  bgColour,
			character: []rune("▄"),
		}
	}
	return currentCell
}

func getHatchedCellColours(x int) (tcell.Color, tcell.Color) {
	var bgColour, fgColour tcell.Color
	if x%2 == 0 {
		bgColour = tcell.NewHexColor(0xa9a9a9)
		fgColour = tcell.NewHexColor(0x797979)
	} else {
		bgColour = tcell.NewHexColor(0x797979)
		fgColour = tcell.NewHexColor(0xa9a9a9)
	}
	return fgColour, bgColour
}
