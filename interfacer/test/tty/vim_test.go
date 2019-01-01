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

	It("should scroll the page by one line", func() {
		Expect("[ˈsmœrɡɔsˌbuːɖ])▄is▄a").To(BeInFrameAt(12, 11))
		Keyboard("j")
		Expect("type▄of▄Scandinavian▄").To(BeInFrameAt(12, 11))
	})

	Describe("Links", func() {
		BeforeEach(func() {
			GotoURL(testSiteURL + "/smorgasbord/links.html")
		})

		FIt("should navigate to a new page by using a link hint", func() {
			Expect("Links").To(BeInFrameAt(0, 0))
			Keyboard("f")
			Keyboard("a")
			// I've noticed we sometimes haven't loaded the page successfully at this point.
			// Maybe we should have a custom matcher that handles loading webpages?
			// Expect("localhost:3000/smorgasbord/another.html").To(BeActivePage()) or something?
			Expect("Another").To(BeInFrameAt(0, 0))
		})
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
			GotoURL(testSiteURL + "/smorgasbord/")
			Keyboard("t")
			GotoURL(testSiteURL + "/smorgasbord/another.html")
			Keyboard("J")
			URL := testSiteURL + "/smorgasbord/             "
			Expect(URL).To(BeInFrameAt(0, 1))
		})
	})
})
