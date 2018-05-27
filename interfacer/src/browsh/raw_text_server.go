package browsh

import (
	"fmt"
	"strings"
	"net/http"
	"net/url"
	"crypto/rand"
	"io"
	"time"
)

// In order to communicate between the incoming HTTP request and the websocket request to the
// real browser to render the webpage, we keep track of requests in a map.
var rawTextRequests = make(map[string]string)

// HTTPServerStart starts the HTTP server is a seperate service from the usual interactive TTY
// app. It accepts normal HTTP requests and uses the path portion of the URL as the entry to the
// Browsh URL bar. It then returns a simple line-broken text version of whatever the browser
// loads. So for example, if you request `curl browsh-http-service.com/http://something.com`,
// it will return:
// `Something                                                                    `
func HTTPServerStart() {
	initialise()
	startFirefox()
	go startWebSocketServer()
	Log("Starting Browsh HTTP server")
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", handleHTTPServerRequest)
	if err := http.ListenAndServe(":" + *HTTPServerPort, &slashFix{serverMux}); err != nil {
		Shutdown(err)
	}
}

func pseudoUUID() (uuid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

type slashFix struct {
	mux http.Handler
}

// The default router from net/http collapses double slashes to a single slash in URL paths.
// This is obviously a problem for putting URLs in the path part of a URL, eg;
// https://domain.com/http://anotherdomain.com
// So here is a little hack that simply escapes the entire path portion to make sure it gets
// through the router unchanged.
func (h *slashFix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/" + url.PathEscape(strings.TrimPrefix(r.URL.Path, "/"))
	h.mux.ServeHTTP(w, r)
}

func handleHTTPServerRequest(w http.ResponseWriter, r *http.Request) {
	urlForBrowsh, _ := url.PathUnescape(strings.TrimPrefix(r.URL.Path, "/"))
	rawTextRequestID := pseudoUUID()
	sendMessageToWebExtension("/raw_text_request," + rawTextRequestID + "," + urlForBrowsh)
	waitForResponse(rawTextRequestID, w)
}

func waitForResponse(rawTextRequestID string, w http.ResponseWriter) {
	var rawTextRequestResponse string
	var ok bool
	for {
		if rawTextRequestResponse, ok = rawTextRequests[rawTextRequestID]; ok {
			io.WriteString(w, rawTextRequestResponse)
			delete(rawTextRequests, rawTextRequestID)
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}
