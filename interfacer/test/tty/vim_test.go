package test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVim(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration tests")
}

var _ = Describe("Vim tests", func() {
	BeforeEach(func() {
		GotoURL(testSiteURL + "/smorgasbord/")
	})

	It("should navigate to a new page by using a link hint", func() {
		Expect("Another▄page").To(BeInFrameAt(12, 18))
		Keyboard("f")
		Keyboard("a")
		Expect("Another").To(BeInFrameAt(0, 0))
	})

	It("should scroll the page by one line", func() {
		Expect("[ˈsmœrɡɔsˌbuːɖ])▄is▄a").To(BeInFrameAt(12, 11))
		Keyboard("j")
		Expect("type▄of▄Scandinavian▄").To(BeInFrameAt(12, 11))
	})

	Describe("Tabs", func() {
		BeforeEach(func() {
			ensureOnlyOneTab()
		})

		It("should create a new tab", func() {
			Keyboard("t")
			Expect("New Tab").To(BeInFrameAt(21, 0))
		})

		It("should cycle to the next tab", func() {
			Keyboard("t")
			GotoURL(testSiteURL + "/smorgasbord/another.html")
			Keyboard("gt")
			URL := testSiteURL + "/smorgasbord/             "
			Expect(URL).To(BeInFrameAt(0, 1))
		})
	})
})
