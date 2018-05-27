package test

import (
	"fmt"
	"time"
	"net/http"
	"io/ioutil"

	ginkgo "github.com/onsi/ginkgo"

	"browsh/interfacer/src/browsh"
)

var staticFileServerPort = "4444"
var rootDir = browsh.Shell("git rev-parse --show-toplevel")

func startStaticFileServer() {
	serverMux := http.NewServeMux()
	serverMux.Handle("/", http.FileServer(http.Dir(rootDir + "/interfacer/test/sites")))
	http.ListenAndServe(":" + staticFileServerPort, serverMux)
}

func startBrowsh() {
	browsh.IsTesting = true
	*browsh.IsHTTPServer = true
	browsh.HTTPServerStart()
}

func getPath(path string) string {
	browshServiceBase := "http://localhost:" + *browsh.HTTPServerPort
	staticFileServerBase := "http://localhost:" + staticFileServerPort
	fullBase := browshServiceBase + "/" + staticFileServerBase
	response, err := http.Get(fullBase + path)
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

var _ =	ginkgo.BeforeEach(func() {
	browsh.IsMonochromeMode = false
	browsh.Log("\n---------")
	browsh.Log(ginkgo.CurrentGinkgoTestDescription().FullTestText)
	browsh.Log("---------")
})

var _ = ginkgo.BeforeSuite(func() {
	go startStaticFileServer()
	go startBrowsh()
	time.Sleep(10 * time.Second)
})

var _	= ginkgo.AfterSuite(func() {
	browsh.Shell(rootDir + "/webext/contrib/firefoxheadless.sh kill")
})
