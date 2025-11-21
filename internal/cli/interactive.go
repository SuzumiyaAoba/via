package cli

import (
	"fmt"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/logger"
	"github.com/SuzumiyaAoba/entry/internal/matcher"
	"github.com/charmbracelet/huh"
)

type Option struct {
	Label    string
	Rule     *config.Rule
	IsSystem bool
}

// buildOptionLabel generates a display label for a rule
func buildOptionLabel(rule *config.Rule) string {
	if rule.Name != "" {
		return rule.Name
	}
	return fmt.Sprintf("Command: %s", rule.Command)
}

// buildInteractiveOptions creates a list of options from matched rules and adds system default
func buildInteractiveOptions(cfg *config.Config, filename string) ([]Option, error) {
	logger.Debug("Building interactive options for: %s", filename)
	matches, err := matcher.MatchAll(cfg.Rules, filename)
	if err != nil {
		logger.Error("Error matching rules for interactive mode: %v", err)
		return nil, fmt.Errorf("error matching rules: %w", err)
	}

	logger.Debug("Found %d potential options", len(matches))
	var options []Option
	for _, m := range matches {
		options = append(options, Option{
			Label: buildOptionLabel(m),
			Rule:  m,
		})
	}

	// Add System Default if the file exists or is a URL
	if isFileOrURL(filename) {
		logger.Debug("Adding system default option")
		options = append(options, Option{
			Label:    "System Default",
			IsSystem: true,
		})
	}

	logger.Info("Built %d interactive options", len(options))
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

func handleInteractive(cfg *config.Config, exec *executor.Executor, filename string) error {
	// Build options from matched rules and system default
	options, err := buildInteractiveOptions(cfg, filename)
	if err != nil {
		return err
	}

	if len(options) == 0 {
		return fmt.Errorf("no matching rules found for %s", filename)
	}

	// Show selector and get user's choice
	selected, err := showOptionSelector(options, filename)
	if err != nil {
		return err
	}

	// Execute the selected option
	return executeSelectedOption(cfg, exec, selected, filename)
}
