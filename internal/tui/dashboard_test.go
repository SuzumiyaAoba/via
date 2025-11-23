package tui_test

import (
	"testing"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TUI Suite")
}

var _ = Describe("Dashboard", func() {
	var (
		cfg *config.Config
		m   tui.Model
		err error
	)

	BeforeEach(func() {
		cfg = &config.Config{
			Rules: []config.Rule{
				{Name: "Rule 1", Command: "cmd1"},
				{Name: "Rule 2", Command: "cmd2"},
			},
		}
		m, err = tui.NewModel(cfg, "config.yml")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should initialize correctly", func() {
		Expect(m.Cfg).To(Equal(cfg))
		Expect(m.Active).To(Equal(tui.TabRules))
		Expect(m.RulesList.Items()).To(HaveLen(2))
	})

	It("should switch tabs", func() {
		// Simulate Tab key
		msg := tea.KeyMsg{Type: tea.KeyTab}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		Expect(m.Active).To(Equal(tui.TabHistory))

		newM, _ = m.Update(msg)
		m = newM.(tui.Model)
		Expect(m.Active).To(Equal(tui.TabSync))

		newM, _ = m.Update(msg)
		m = newM.(tui.Model)
		Expect(m.Active).To(Equal(tui.TabRules))
	})

	It("should handle window size", func() {
		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		Expect(m.Width).To(Equal(100))
		Expect(m.Height).To(Equal(50))
	})
})
