package cli

import (
	"fmt"
	os_exec "os/exec"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/logger"
	"github.com/SuzumiyaAoba/entry/internal/matcher"
)

// executeWithDefault executes the filename with either the default command or system default
func executeWithDefault(cfg *config.Config, exec *executor.Executor, filename string) error {
	if cfg.DefaultCommand != "" {
		logger.Debug("Executing with default command: %s", cfg.DefaultCommand)
		return exec.Execute(cfg.DefaultCommand, filename, executor.ExecutionOptions{})
	}
	logger.Debug("Opening with system default")
	return exec.OpenSystem(filename)
}

// executeRule executes a single rule with the appropriate options
// Returns true if the rule was executed (matched), false otherwise
func executeRule(exec *executor.Executor, rule *config.Rule, filename string) (bool, error) {
	logger.Debug("Evaluating rule '%s'", rule.Name)
	
	command := rule.Command

	// Handle Script
	if rule.Script != "" {
		scriptCmd, matched, err := exec.ExecuteScript(rule.Script, filename)
		if err != nil {
			return false, fmt.Errorf("script execution failed: %w", err)
		}
		if !matched {
			logger.Debug("Script returned false/null, skipping rule")
			return false, nil
		}
		if scriptCmd != "" {
			command = scriptCmd
		}
	}

	if command == "" {
		// If script matched but returned true (bool) and no command is defined in rule
		// We can't execute anything.
		logger.Debug("Rule matched but no command to execute")
		return true, nil
	}

	logger.Debug("Executing rule '%s' with command: %s", rule.Name, command)
	opts := executor.ExecutionOptions{
		Background: rule.Background,
		Terminal:   rule.Terminal,
	}
	if err := exec.Execute(command, filename, opts); err != nil {
		return true, err
	}
	return true, nil
}

// executeRules executes all matched rules (with fallthrough support)
func executeRules(exec *executor.Executor, rules []*config.Rule, filename string) error {
	logger.Info("Processing %d matched rules for %s", len(rules), filename)
	for _, rule := range rules {
		executed, err := executeRule(exec, rule, filename)
		if err != nil {
			return err
		}
		if executed {
			if !rule.Fallthrough {
				return nil
			}
		}
	}
	return nil
}

// matchRules matches rules against a filename and returns matched rules
func matchRules(cfg *config.Config, filename string) ([]*config.Rule, error) {
	logger.Debug("Matching rules for file: %s", filename)
	matched, err := matcher.Match(cfg.Rules, filename)
	if err != nil {
		logger.Error("Failed to match rules: %v", err)
		return nil, err
	}
	logger.Debug("Found %d matching rules", len(matched))
	return matched, nil
}

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
