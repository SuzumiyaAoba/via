package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/matcher"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// File and URL detection helpers

// isURL checks if the given string is a valid URL with a scheme
func isURL(filename string) bool {
	u, err := url.Parse(filename)
	return err == nil && u.Scheme != ""
}

// fileExists checks if a file exists at the given path
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// isFileOrURL checks if the filename is either a valid URL or an existing file
func isFileOrURL(filename string) bool {
	return isURL(filename) || fileExists(filename)
}

// Execution helpers

// executeWithDefault executes the filename with either the default command or system default
func executeWithDefault(cfg *config.Config, exec *executor.Executor, filename string) error {
	if cfg.DefaultCommand != "" {
		return exec.Execute(cfg.DefaultCommand, filename, executor.ExecutionOptions{})
	}
	return exec.OpenSystem(filename)
}

// executeRule executes a single rule with the appropriate options
func executeRule(exec *executor.Executor, rule *config.Rule, filename string) error {
	opts := executor.ExecutionOptions{
		Background: rule.Background,
		Terminal:   rule.Terminal,
	}
	return exec.Execute(rule.Command, filename, opts)
}

// executeRules executes all matched rules (with fallthrough support)
func executeRules(exec *executor.Executor, rules []*config.Rule, filename string) error {
	for _, rule := range rules {
		if err := executeRule(exec, rule, filename); err != nil {
			return err
		}
	}
	return nil
}

// matchRules matches rules against a filename and returns matched rules
func matchRules(cfg *config.Config, filename string) ([]*config.Rule, error) {
	return matcher.Match(cfg.Rules, filename)
}

// Option helpers for interactive mode

// buildOptionLabel generates a display label for a rule
func buildOptionLabel(rule *config.Rule) string {
	if rule.Name != "" {
		return rule.Name
	}
	return fmt.Sprintf("Command: %s", rule.Command)
}

// buildInteractiveOptions creates a list of options from matched rules and adds system default
func buildInteractiveOptions(cfg *config.Config, filename string) ([]Option, error) {
	matches, err := matcher.MatchAll(cfg.Rules, filename)
	if err != nil {
		return nil, fmt.Errorf("error matching rules: %w", err)
	}

	var options []Option
	for _, m := range matches {
		options = append(options, Option{
			Label: buildOptionLabel(m),
			Rule:  m,
		})
	}

	// Add System Default if the file exists or is a URL
	if isFileOrURL(filename) {
		options = append(options, Option{
			Label:    "System Default",
			IsSystem: true,
		})
	}

	return options, nil
}

// showOptionSelector displays an interactive selector and returns the selected option
func showOptionSelector(options []Option, filename string) (Option, error) {
	var selected Option

	huhOptions := make([]huh.Option[Option], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt.Label, opt)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[Option]().
				Title("Select action for " + filename).
				Options(huhOptions...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return Option{}, err
	}

	return selected, nil
}

// executeSelectedOption executes the option selected by the user
func executeSelectedOption(cfg *config.Config, exec *executor.Executor, selected Option, filename string) error {
	if selected.IsSystem {
		return executeWithDefault(cfg, exec, filename)
	}
	return executeRule(exec, selected.Rule, filename)
}

// Table creation helpers

// createStyledTable creates a table with standard styling
func createStyledTable(headers []string, rows [][]string) string {
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("14")).
					Bold(true).
					Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		}).
		Headers(headers...)

	for _, row := range rows {
		t.Row(row...)
	}

	return t.Render()
}
