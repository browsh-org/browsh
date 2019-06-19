package test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	ginkgo "github.com/onsi/ginkgo"

	"browsh/interfacer/src/browsh"
	"github.com/spf13/viper"
)

var staticFileServerPort = "4444"
var rootDir = browsh.Shell("git rev-parse --show-toplevel")

func startStaticFileServer() {
	serverMux := http.NewServeMux()
	serverMux.Handle("/", http.FileServer(http.Dir(rootDir+"/interfacer/test/sites")))
	http.ListenAndServe(":"+staticFileServerPort, serverMux)
}

func initBrowsh() {
	browsh.IsTesting = true
	browsh.Initialise()
	viper.Set("http-server-mode", true)
}

func waitUntilConnectedToWebExtension(maxTime time.Duration) {
	start := time.Now()
	for time.Since(start) < maxTime {
		if browsh.IsConnectedToWebExtension {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	panic("Didn't connect to webextension in time")
}

func getBrowshServiceBase() string {
	return "http://localhost:" + viper.GetString("http-server.port")
}

func getPath(path string, mode string) string {
	browshServiceBase := getBrowshServiceBase()
	staticFileServerBase := "http://localhost:" + staticFileServerPort
	fullBase := browshServiceBase + "/" + staticFileServerBase
	client := &http.Client{}
	request, err := http.NewRequest("GET", fullBase+path, nil)
	if mode == "plain" {
		request.Header.Add("X-Browsh-Raw-Mode", "PLAIN")
	}
	if mode == "dom" {
		request.Header.Add("X-Browsh-Raw-Mode", "DOM")
	}
	response, err := client.Do(request)
	if err != nil {
		panic(fmt.Sprintf("%s", err))
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			panic(fmt.Sprintf("%s", err))
		}
		return string(contents)
	}
}

func stopFirefox() {
	browsh.IsConnectedToWebExtension = false
	browsh.Shell(rootDir + "/webext/contrib/firefoxheadless.sh kill")
	time.Sleep(500 * time.Millisecond)
}

var _ = ginkgo.BeforeEach(func() {
	stopFirefox()
	browsh.ResetTabs()
	browsh.StartFirefox()
	waitUntilConnectedToWebExtension(15 * time.Second)
	browsh.IsMonochromeMode = false
	browsh.Log("\n---------")
	browsh.Log(ginkgo.CurrentGinkgoTestDescription().FullTestText)
	browsh.Log("---------")
})

var _ = ginkgo.BeforeSuite(func() {
	initBrowsh()
	stopFirefox()
	go startStaticFileServer()
	go browsh.HTTPServerStart()
	time.Sleep(1 * time.Second)
})

var _ = ginkgo.AfterSuite(func() {
	stopFirefox()
})
