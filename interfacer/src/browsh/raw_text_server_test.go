package browsh

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRawTextServer(t *testing.T) {
	RegisterFailHandler(Fail)
}

var _ = Describe("Raw text server", func() {
	Describe("De-recursing URLs", func() {
		It("should not do anything to normal URLs", func() {
			testURL := "https://google.com/path?q=hey"
			url, _ := deRecurseURL(testURL)
			Expect(url).To(Equal(testURL))
		})
		It("should de-recurse a single level", func() {
			testURL := "https://html.brow.sh/word"
			url, _ := deRecurseURL(testURL)
			Expect(url).To(Equal("word"))
		})
		It("should de-recurse a multi level recurse without a URL ending", func() {
			testURL := "https://html.brow.sh/https://html.brow.sh"
			url, _ := deRecurseURL(testURL)
			Expect(url).To(Equal(""))
		})
		It("should de-recurse a multi level recurse with a URL ending", func() {
			google := "https://google.com/path?q=hey"
			testURL := "https://html.brow.sh/https://html.brow.sh/" + google
			url, _ := deRecurseURL(testURL)
			Expect(url).To(Equal(google))
		})
	})
})
