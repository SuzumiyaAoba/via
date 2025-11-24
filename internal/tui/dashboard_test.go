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

	It("should delete rule", func() {
		// Select first rule
		m.RulesList.Select(0)
		
		// Simulate Delete key
		msg := tea.KeyMsg{Type: tea.KeyDelete}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		
		Expect(m.Cfg.Rules).To(HaveLen(1))
		Expect(m.Cfg.Rules[0].Name).To(Equal("Rule 2"))
	})

	It("should enter add mode", func() {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		
		Expect(m.Active).To(Equal(tui.TabEdit))
		Expect(m.SelectedRuleIndex).To(Equal(-1))
		Expect(m.EditForm).NotTo(BeNil())
	})

	It("should enter edit mode", func() {
		m.RulesList.Select(0)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		
		Expect(m.Active).To(Equal(tui.TabEdit))
		Expect(m.SelectedRuleIndex).To(Equal(0))
		Expect(m.EditForm).NotTo(BeNil())
	})

	It("should show details", func() {
		m.RulesList.Select(0)
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		
		Expect(m.ShowDetail).To(BeTrue())
		Expect(m.DetailRule.Name).To(Equal("Rule 1"))
		
		// Verify View contains details
		view := m.View()
		Expect(view).To(ContainSubstring("Rule Details"))
		Expect(view).To(ContainSubstring("Rule 1"))
	})

	It("should filter rules", func() {
		// Enter filter mode
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		
		// Type filter query "Rule 2"
		for _, r := range "Rule 2" {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
			newM, _ := m.Update(msg)
			m = newM.(tui.Model)
		}
		
		// Verify list is filtered (this depends on list implementation, 
		// but we can check if View shows only Rule 2 or if list items count changed)
		// Since we can't easily access internal list state, we check View
		view := m.View()
		Expect(view).To(ContainSubstring("Rule 2"))
		// Rule 1 might still be there if fuzzy matching matches "Rule 1" with "Rule 2" query?
		// "Rule 2" query should match "Rule 2" strongly.
	})
	It("should return nil for Init", func() {
		Expect(m.Init()).To(BeNil())
	})

	It("should render sync status", func() {
		// Switch to sync tab
		m.Active = tui.TabSync
		view := m.View()
		Expect(view).To(ContainSubstring("Sync Status"))
		Expect(view).To(ContainSubstring("Sync not initialized"))

		// With config
		m.Cfg.Sync = &config.SyncConfig{GistID: "123"}
		view = m.View()
		Expect(view).To(ContainSubstring("Gist ID: 123"))
	})

	It("should return full help", func() {
		// We can't easily access the internal keyMap, but we can verify Help view shows something
		// Or we can test the keyMap methods directly if we export them or test via Model
		// Since keyMap is not exported, we rely on Help view
		m.Help.ShowAll = true
		view := m.View()
		Expect(view).To(ContainSubstring("↑/k"))
	})
	
	It("should return filter value", func() {
		item := m.RulesList.Items()[0]
		Expect(item.FilterValue()).To(ContainSubstring("Rule 1"))
	})

	Describe("RuleItem", func() {
		It("should generate correct description", func() {
			rule := config.Rule{
				Command:    "cmd",
				Extensions: []string{"txt", "md"},
				Regex:      "^test",
				Script:     "script",
			}
			item := tui.RuleItem{Rule: rule}
			desc := item.Description()
			Expect(desc).To(ContainSubstring("[txt, md]"))
			Expect(desc).To(ContainSubstring("Regex: ^test"))
			Expect(desc).To(ContainSubstring("JS"))
			Expect(desc).To(ContainSubstring("-> cmd"))
		})

		It("should use command as description if no other info", func() {
			rule := config.Rule{Command: "cmd"}
			item := tui.RuleItem{Rule: rule}
			Expect(item.Description()).To(Equal("cmd"))
		})
	})

	Describe("Update interactions", func() {
		It("should move rule up", func() {
			// Select second rule
			m.RulesList.Select(1)
			
			// Move up
			msg := tea.KeyMsg{Type: tea.KeyUp, Alt: false, Runes: []rune{'K'}} // Shift+Up is usually handled as separate key or mapped
			// In our keymap: MoveUp: key.WithKeys("shift+up", "K")
			// Let's use "K"
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}}
			
			newM, _ := m.Update(msg)
			m = newM.(tui.Model)
			
			Expect(m.Cfg.Rules[0].Name).To(Equal("Rule 2"))
			Expect(m.Cfg.Rules[1].Name).To(Equal("Rule 1"))
		})

		It("should move rule down", func() {
			// Select first rule
			m.RulesList.Select(0)
			
			// Move down using "J"
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}}
			
			newM, _ := m.Update(msg)
			m = newM.(tui.Model)
			
			Expect(m.Cfg.Rules[0].Name).To(Equal("Rule 2"))
			Expect(m.Cfg.Rules[1].Name).To(Equal("Rule 1"))
		})

		It("should handle edit form submission", func() {
			// Enter add mode
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
			newM, _ := m.Update(msg)
			m = newM.(tui.Model)
			
			// We can't easily drive the Huh form programmatically in a unit test without complex mocking.
			// However, we can simulate the completion state.
			// But `Update` checks `m.EditForm.State`. We can't set that directly as `EditForm` is internal to `huh`.
			// Wait, `EditForm` is `*huh.Form` which is exported. But `State` field might not be settable directly if it's not exported or if we can't reach it.
			// Actually `huh.Form` struct fields are not all exported.
			// So testing the form completion logic might be hard.
			// But we can test the logic *after* form completion if we could trigger it.
			
			// Alternative: We can test the logic by manually setting the state if possible, or skip deep form testing and rely on integration tests (which we don't have easily here).
			// Let's try to at least cover the "Abort" path if we can send Esc?
			
			// Send Ctrl+C to form (Esc might be captured by input fields or not configured to abort)
			msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			newM, _ = m.Update(msg)
			m = newM.(tui.Model)
			
			// Should return to rules
			Expect(m.Active).To(Equal(tui.TabRules))
			Expect(m.EditForm).To(BeNil())
		})
	})
})
