package test

import (
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

func TestHTTPServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Server tests")
}

var _ = Describe("HTTP Server", func() {
	It("should return plain text", func() {
		response := getPath("/smorgasbord", "plain")
		Expect(response).To(ContainSubstring("multiple hot       Smörgås"))
		Expect(response).To(ContainSubstring("A special Swedish type of smörgåsbord"))
		Expect(response).ToNot(ContainSubstring("<a href"))
	})

	It("should return HTML text", func() {
		response := getPath("/smorgasbord", "html")
		Expect(response).To(ContainSubstring(
			"<a href=\"/http://localhost:4444/smorgasbord/another.html\">Another page</a>"))
	})

	It("should return the DOM", func() {
		response := getPath("/smorgasbord", "dom")
		Expect(response).To(ContainSubstring(
			"<div class=\"big_middle\">"))
	})

	It("should return a background image", func() {
		response := getPath("/smorgasbord", "html")
		Expect(response).To(ContainSubstring("background-image: url(data:image/jpeg"))
	})

	It("should block specified domains", func() {
		viper.Set(
			"http-server.blocked-domains",
			[]string{"[mail|accounts].google.com", "other"},
		)
		url := getBrowshServiceBase() + "/mail.google.com"
		client := &http.Client{}
		request, _ := http.NewRequest("GET", url, nil)
		response, _ := client.Do(request)
		contents, _ := ioutil.ReadAll(response.Body)
		Expect(string(contents)).To(ContainSubstring("Welcome to the Browsh HTML"))
	})

	It("should block specified user agents", func() {
		viper.Set(
			"http-server.blocked-user-agents",
			[]string{"MJ12bot", "other"},
		)
		url := getBrowshServiceBase() + "/example.com"
		client := &http.Client{}
		request, _ := http.NewRequest("GET", url, nil)
		request.Header.Add("User-Agent", "Blah blah MJ12bot etc")
		response, _ := client.Do(request)
		Expect(response.StatusCode).To(Equal(403))
	})

	It("should allow a blocked user agent to see the home page", func() {
		viper.Set(
			"http-server.blocked-user-agents",
			[]string{"MJ12bot", "other"},
		)
		url := getBrowshServiceBase()
		client := &http.Client{}
		request, _ := http.NewRequest("GET", url, nil)
		request.Header.Add("User-Agent", "Blah blah MJ12bot etc")
		response, _ := client.Do(request)
		Expect(response.StatusCode).To(Equal(200))
	})
})
