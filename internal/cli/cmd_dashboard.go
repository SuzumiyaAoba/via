package cli

import (
	"fmt"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   ":dashboard",
	Short: "Open TUI dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDashboard(cmd)
	},
}

func runDashboard(cmd *cobra.Command) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		// If config doesn't exist, start with empty config
		cfg = &config.Config{Version: "1"}
	}

	model, err := tui.NewModel(cfg)
	if err != nil {
		return err
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running dashboard: %w", err)
	}

	return nil
}
