package main

import (
	"fmt"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
)

type Option struct {
	Label    string
	Rule     *config.Rule
	IsSystem bool
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
