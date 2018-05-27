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
	It("should return text", func() {
		response := getPath("/smorgasbord")
		Expect(response).To(ContainSubstring("multiple hot       Smörgås"))
		Expect(response).To(ContainSubstring("A special Swedish type of smörgåsbord"))
	})
})
