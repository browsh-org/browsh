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

func startBrowsh() {
	browsh.IsTesting = true
	browsh.Initialise()
	viper.Set("http-server-mode", true)
	browsh.HTTPServerStart()
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

var _ = ginkgo.BeforeEach(func() {
	browsh.IsMonochromeMode = false
	browsh.Log("\n---------")
	browsh.Log(ginkgo.CurrentGinkgoTestDescription().FullTestText)
	browsh.Log("---------")
})

var _ = ginkgo.BeforeSuite(func() {
	go startStaticFileServer()
	go startBrowsh()
	time.Sleep(5 * time.Second)
	// Allow the browser to sort its sizing out, because sometimes the first test catches the
	// browser before it's completed its resizing.
	getPath("/smorgasbord", "plain")
})

var _ = ginkgo.AfterSuite(func() {
	browsh.Shell(rootDir + "/webext/contrib/firefoxheadless.sh kill")
})
