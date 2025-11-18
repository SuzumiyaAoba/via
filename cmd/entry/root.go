package main

import (
	"fmt"
	"os"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/matcher"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	dryRun  bool
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/entry/config.yml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print command instead of executing")
}

var rootCmd = &cobra.Command{
	Use:   "entry <file>",
	Short: "Entry is a CLI file association tool",
	Long:  `Entry allows you to execute specific commands based on file extensions or regex patterns matched against a provided file argument.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		// 1. Try to match a specific rule if single argument
		if len(args) == 1 {
			rule, err := matcher.Match(cfg.Rules, "", args[0])
			if err != nil {
				return fmt.Errorf("error matching rule: %w", err)
			}
			if rule != nil {
				return executor.Execute(rule.Command, args[0], dryRun)
			}
		}

		command := args[0]
		cmdArgs := args[1:]

		// 2. Check aliases
		if alias, ok := cfg.Aliases[command]; ok {
			command = alias
			return executor.ExecuteCommand(command, cmdArgs, dryRun)
		}

		// 3. If single argument and default command exists, use it
		if len(args) == 1 && cfg.DefaultCommand != "" {
			return executor.Execute(cfg.DefaultCommand, args[0], dryRun)
		}

		// 4. Fallback to command execution
		return executor.ExecuteCommand(command, cmdArgs, dryRun)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
