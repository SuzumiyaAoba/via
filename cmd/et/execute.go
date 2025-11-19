package main

import (
	"fmt"
	os_exec "os/exec"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
)

func handleFileExecution(cfg *config.Config, exec *executor.Executor, filename string) error {
	// Try to match rules
	rules, err := matchRules(cfg, filename)
	if err != nil {
		return fmt.Errorf("error matching rule: %w", err)
	}

	// Execute matched rules (with fallthrough support)
	if len(rules) > 0 {
		return executeRules(exec, rules, filename)
	}

	// Use default command or system opener for files/URLs
	if isFileOrURL(filename) {
		return executeWithDefault(cfg, exec, filename)
	}

	// File not found - caller should handle this as a command
	return fmt.Errorf("file not found and no matching rule")
}

func handleCommandExecution(cfg *config.Config, exec *executor.Executor, commandArgs []string) error {
	command := commandArgs[0]
	cmdArgs := commandArgs[1:]

	// Check aliases
	if alias, ok := cfg.Aliases[command]; ok {
		command = alias
		return exec.ExecuteCommand(command, cmdArgs)
	}

	// Fallback to command execution
	// Check if command exists in PATH
	if _, err := os_exec.LookPath(command); err != nil {
		// Command not found.
		// If single argument and default command exists, assume it's a new file and use default command.
		if len(commandArgs) == 1 && cfg.DefaultCommand != "" {
			return exec.Execute(cfg.DefaultCommand, commandArgs[0], executor.ExecutionOptions{})
		}
	}

	return exec.ExecuteCommand(command, cmdArgs)
}
