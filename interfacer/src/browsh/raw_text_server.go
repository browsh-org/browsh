package browsh

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/spf13/viper"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/middleware/stdlib"
	"github.com/ulule/limiter/drivers/store/memory"
)

// In order to communicate between the incoming HTTP request and the websocket request to the
// real browser to render the webpage, we keep track of requests in a map.
var rawTextRequests = newRequestsMap()

type threadSafeRequestsMap struct {
	sync.RWMutex
	internal map[string]string
}

func newRequestsMap() *threadSafeRequestsMap {
	return &threadSafeRequestsMap{
		internal: make(map[string]string),
	}
}

func (m *threadSafeRequestsMap) load(key string) (value string, ok bool) {
	m.RLock()
	result, ok := m.internal[key]
	m.RUnlock()
	return result, ok
}

func (m *threadSafeRequestsMap) store(key string, value string) {
	m.Lock()
	m.internal[key] = value
	m.Unlock()
}

func (m *threadSafeRequestsMap) remove(key string) {
	m.Lock()
	delete(m.internal, key)
	m.Unlock()
}

type rawTextResponse struct {
	PageloadDuration int    `json:"page_load_duration"`
	ParsingDuration  int    `json:"parsing_duration"`
	Text             string `json:"body"`
}

// HTTPServerStart starts the HTTP server is a seperate service from the usual interactive TTY
// app. It accepts normal HTTP requests and uses the path portion of the URL as the entry to the
// Browsh URL bar. It then returns a simple line-broken text version of whatever the browser
// loads. So for example, if you request `curl browsh-http-service.com/http://something.com`,
// it will return:
// `Something                                                                    `
func HTTPServerStart() {
	startFirefox()
	go startWebSocketServer()
	Log("Starting Browsh HTTP server")
	bind := viper.GetString("http-server.bind")
	port := viper.GetString("http-server.port")
	serverMux := http.NewServeMux()
	uncompressed := http.HandlerFunc(handleHTTPServerRequest)
	limiterMiddleware := setupRateLimiter()
	serverMux.Handle("/", limiterMiddleware.Handler(gziphandler.GzipHandler(uncompressed)))
	if err := http.ListenAndServe(bind+":"+port, &slashFix{serverMux}); err != nil {
		Shutdown(err)
	}
}

func HTTPServerStop() {
	quitFirefox()
}

func setupRateLimiter() *stdlib.Middleware {
	rate, err := limiter.NewRateFromFormatted(viper.GetString("http-server.rate-limit"))
	if err != nil {
		Shutdown(err)
	}
	// TODO: Centralise store amongst instances with Redis
	store := memory.NewStore()
	middleware := stdlib.NewMiddleware(limiter.New(store, rate), stdlib.WithForwardHeader(true))
	return middleware
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
	r.URL.Path = "/" + url.PathEscape(strings.TrimPrefix(r.URL.RequestURI(), "/"))
	h.mux.ServeHTTP(w, r)
}

func handleHTTPServerRequest(w http.ResponseWriter, r *http.Request) {
	var message string
	var isErrored bool
	var start = time.Now().Format(time.RFC3339)
	urlForBrowsh, _ := url.PathUnescape(strings.TrimPrefix(r.URL.Path, "/"))
	urlForBrowsh, isErrored = deRecurseURL(urlForBrowsh)
	if isErrored {
		message = "Invalid URL"
		io.WriteString(w, message)
		return
	}
	if isProductionHTTP(r) {
		http.Redirect(w, r, "https://"+r.Host+"/"+urlForBrowsh, 301)
		return
	}
	if urlForBrowsh == "favicon.ico" {
		http.Redirect(w, r, "https://www.brow.sh/assets/favicon-16x16.png", 301)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=600")
	if isDisallowedDomain(urlForBrowsh) {
		http.Redirect(w, r, "/", 301)
		return
	}
	if isDisallowedUserAgent(r.Header.Get("User-Agent")) {
		if urlForBrowsh != "" {
			http.Redirect(w, r, "/", 403)
			return
		}
	}
	Log(r.Header.Get("User-Agent"))
	if isKubeReadinessProbe(r.Header.Get("User-Agent")) {
		io.WriteString(w, "healthy")
		return
	}
	if strings.TrimSpace(urlForBrowsh) == "" {
		if strings.Contains(r.Host, "text.") {
			message = "Welcome to the Browsh plain text client.\n" +
				"You can use it by appending URLs like this;\n" +
				"https://text.brow.sh/https://www.brow.sh"
			io.WriteString(w, message)
			return
		}
		urlForBrowsh = "https://www.brow.sh/html-service-welcome"
	}
	if urlForBrowsh == "robots.txt" {
		message = "User-agent: *\nAllow: /$\nDisallow: /\n"
		io.WriteString(w, message)
		return
	}
	rawTextRequestID := pseudoUUID()
	rawTextRequests.store(rawTextRequestID+"-start", start)
	mode := getRawTextMode(r)
	sendMessageToWebExtension(
		"/raw_text_request," + rawTextRequestID + "," +
			mode + "," +
			urlForBrowsh)
	waitForResponse(rawTextRequestID, w)
}

// Prevent https://html.brow.sh/html.brow.sh/... being recursive
func deRecurseURL(urlForBrowsh string) (string, bool) {
	nestedURL, err := url.Parse(urlForBrowsh)
	if err != nil {
		return urlForBrowsh, false
	}
	if nestedURL.Host != "html.brow.sh" && nestedURL.Host != "text.brow.sh" {
		return urlForBrowsh, false
	}
	return deRecurseURL(strings.TrimPrefix(nestedURL.RequestURI(), "/"))
}

func isDisallowedDomain(urlForBrowsh string) bool {
	for _, domainish := range viper.GetStringSlice("http-server.blocked-domains") {
		r, _ := regexp.Compile(domainish)
		if r.MatchString(urlForBrowsh) {
			return true
		}
	}
	return false
}

func isDisallowedUserAgent(userAgent string) bool {
	for _, agentish := range viper.GetStringSlice("http-server.blocked-user-agents") {
		r, _ := regexp.Compile(agentish)
		if r.MatchString(userAgent) {
			return true
		}
	}
	return false
}

func isKubeReadinessProbe(userAgent string) bool {
	r, _ := regexp.Compile("GoogleHC")
	if r.MatchString(userAgent) {
		return true
	}
	return false
}

func isProductionHTTP(r *http.Request) bool {
	if strings.Contains(r.Host, "brow.sh") {
		return r.Header.Get("X-Forwarded-Proto") == "http"
	}
	return false
}

// 'PLAIN' mode returns raw text without any HTML whatsoever.
// 'HTML' mode returns some basic HTML tags for things like anchor links.
func getRawTextMode(r *http.Request) string {
	var mode = "HTML"
	if strings.Contains(r.Host, "text.") {
		mode = "PLAIN"
	}
	if r.Header.Get("X-Browsh-Raw-Mode") == "PLAIN" {
		mode = "PLAIN"
	}
	return mode
}

func waitForResponse(rawTextRequestID string, w http.ResponseWriter) {
	var rawTextRequestResponse string
	var jsonResponse rawTextResponse
	var totalTime, pageLoad, parsing string
	var ok bool
	for {
		if rawTextRequestResponse, ok = rawTextRequests.load(rawTextRequestID); ok {
			jsonResponse = unpackResponse(rawTextRequestResponse)
			requestStart, _ := rawTextRequests.load(rawTextRequestID + "-start")
			totalTime = getTotalTiming(requestStart)
			pageLoad = fmt.Sprintf("%d", jsonResponse.PageloadDuration)
			parsing = fmt.Sprintf("%d", jsonResponse.ParsingDuration)
			w.Header().Set("X-Browsh-Duration-Total", totalTime)
			w.Header().Set("X-Browsh-Duration-Pageload", pageLoad)
			w.Header().Set("X-Browsh-Duration-Parsing", parsing)
			io.WriteString(w, jsonResponse.Text)
			rawTextRequests.remove(rawTextRequestID)
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func unpackResponse(jsonString string) rawTextResponse {
	var response rawTextResponse
	jsonBytes := []byte(jsonString)
	if err := json.Unmarshal(jsonBytes, &response); err != nil {
	}
	return response
}

func getTotalTiming(startString string) string {
	start, _ := time.Parse(time.RFC3339, startString)
	elapsed := time.Since(start) / time.Millisecond
	return fmt.Sprintf("%d", elapsed)
}
