package test

import (
	"testing"

	"github.com/gdamore/tcell"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration tests")
}

var _ = Describe("Showing a basic webpage", func() {
	BeforeEach(func() {
		GotoURL(testSiteURL + "/smorgasbord/")
  })

	Describe("Browser UI", func() {
		It("should have the page title and current URL", func() {
			Expect("Smörgåsbord").To(BeInFrameAt(0, 0))
			URL := testSiteURL + "/smorgasbord/"
			Expect(URL).To(BeInFrameAt(0, 1))
		})

		Describe("Interaction", func() {
			It("should navigate to a new page by using the URL bar", func() {
				SpecialKey(tcell.KeyCtrlL)
				Keyboard(testSiteURL + "/smorgasbord/another.html")
				SpecialKey(tcell.KeyEnter)
				Expect("Another").To(BeInFrameAt(0, 0))
			})

			It("should navigate to a new page by clicking a link", func() {
				Expect("Another▄page").To(BeInFrameAt(12, 18))
				mouseClick(12, 18)
				Expect("Another").To(BeInFrameAt(0, 0))
			})

			It("should scroll the page by one line", func() {
				SpecialKey(tcell.KeyDown)
				Expect("meal,▄originating▄in▄").To(BeInFrameAt(12, 11))
			})

			It("should scroll the page by one page", func() {
				SpecialKey(tcell.KeyPgDn)
				Expect("continuing▄with▄a▄variety▄of▄fish").To(BeInFrameAt(12, 13))
			})

			Describe("Text Input", func() {
				Describe("Single line", func() {
					BeforeEach(func() {
						SpecialKey(tcell.KeyDown)
						SpecialKey(tcell.KeyDown)
						simScreen.InjectMouse(12, 16, tcell.Button1, tcell.ModNone)
					})

					It("should have basic cursor movement", func() {
						Keyboard("|||")
						SpecialKey(tcell.KeyLeft)
						Keyboard("2")
						SpecialKey(tcell.KeyLeft)
						SpecialKey(tcell.KeyLeft)
						Keyboard("1")
						Expect("|1|2|").To(BeInFrameAt(12, 16))
					})

					It("should scroll single line boxes on overflow", func() {
						Keyboard("12345678901234567890")
						Expect("45678901234567890").To(BeInFrameAt(12, 16))
					})

					It("should scroll overflowed boxes to the left and right", func() {
						Keyboard("12345678901234567890")
						for i := 0; i < 19; i++ {
							SpecialKey(tcell.KeyLeft)
						}
						Expect("23456789012345678").To(BeInFrameAt(12, 16))
						for i := 0; i < 19; i++ {
							SpecialKey(tcell.KeyRight)
						}
						Expect("45678901234567890").To(BeInFrameAt(12, 16))
					})

					It("should submit text into an input box", func() {
						Expect("Unsubmitted").To(BeInFrameAt(12, 20))
						Keyboard("Reverse Me!")
						SpecialKey(tcell.KeyEnter)
						Expect("!eM▄esreveR").To(BeInFrameAt(12, 20))
					})
				})

				Describe("Multi line", func() {
					BeforeEach(func() {
						GotoURL(testSiteURL + "/smorgasbord/textarea.html")
						mouseClick(2, 3)
					})

					It("should enter multiple lines of text", func() {
						Keyboard(`So here is a lot of text that will hopefully split across lines`)
						Expect("So here is a lot of").To(BeInFrameAt(1, 3))
						Expect("text that will").To(BeInFrameAt(1, 4))
						Expect("hopefully split across").To(BeInFrameAt(1, 5))
						Expect("lines").To(BeInFrameAt(1, 6))
					})

					It("should scroll multiple lines of text", func() {
						Keyboard(`So here is a lot of text that will hopefully split across lines`)
						SpecialKey(tcell.KeyEnter)
						Keyboard(`And here is even more filler, it's endless!`)
						Expect("filler, it's endless!").To(BeInFrameAt(1, 6))
						for i := 1; i <= 6; i++ {
							SpecialKey(tcell.KeyUp)
						}
						Expect("lines").To(BeInFrameAt(1, 6))
					})
				})
			})

			Describe("Tabs", func() {
				BeforeEach(func() {
					SpecialKey(tcell.KeyCtrlT)
				})

				AfterEach(func() {
					ensureOnlyOneTab()
				})

				It("should create a new tab", func() {
					Expect("New Tab").To(BeInFrameAt(21, 0))
				})

				It("should be able to goto a new URL", func() {
					Keyboard(testSiteURL + "/smorgasbord/another.html")
					SpecialKey(tcell.KeyEnter)
					Expect("Another").To(BeInFrameAt(21, 0))
				})

				It("should cycle to the next tab", func() {
					Expect("                   ").To(BeInFrameAt(0, 1))
					SpecialKey(tcell.KeyCtrlL)
					GotoURL(testSiteURL + "/smorgasbord/another.html")
					simScreen.InjectKey(9, 0, tcell.ModNone)
					URL := testSiteURL + "/smorgasbord/             "
					Expect(URL).To(BeInFrameAt(0, 1))
				})
			})
		})
	})

	Describe("Rendering", func() {
		It("should reset page scroll to zero on page load", func() {
			SpecialKey(tcell.KeyPgDn)
			Expect("continuing▄with▄a▄variety▄of▄fish").To(BeInFrameAt(12, 13))
			GotoURL(testSiteURL + "/smorgasbord/another.html")
			Expect("Another▄webpage").To(BeInFrameAt(1, 3))
		})

		It("should render dynamic content", func() {
			var greens, pinks int
			var colours [10][3]int32
			for i := 0; i < 10; i++ {
				colours[i] = GetFgColour(39, 3)
				waitForNextFrame()
			}
			for i := 0; i < 10; i++ {
				if colours[i] == [3]int32{0, 255, 255} { greens++ }
				if colours[i] == [3]int32{255, 0, 255} { pinks++ }
			}
			Expect(greens).To(BeNumerically(">=", 1))
			Expect(pinks).To(BeNumerically(">=", 1))
		})

		It("should switch to monochrome mode", func() {
			simScreen.InjectKey(tcell.KeyRune, 'm', tcell.ModAlt)
			waitForNextFrame()
			Expect([3]int32{0, 0, 0}).To(Equal(GetBgColour(0, 2)))
			Expect([3]int32{255, 255, 255}).To(Equal(GetFgColour(12, 11)))
		})

		Describe("Text positioning", func() {
			It("should position the left/right-aligned coloumns", func() {
				Expect("Smörgåsbord▄(Swedish:").To(BeInFrameAt(12, 10))
				Expect("The▄Swedish▄word").To(BeInFrameAt(42, 10))
			})
		})
	})
})
