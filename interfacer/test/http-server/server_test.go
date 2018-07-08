package test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	It("should return a background image", func() {
		response := getPath("/smorgasbord", "html")
		Expect(response).To(ContainSubstring("background-image: url(data:image/jpeg"))
	})
})
