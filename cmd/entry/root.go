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
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]

		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		rule, err := matcher.Match(cfg.Rules, file)
		if err != nil {
			return fmt.Errorf("error matching rule: %w", err)
		}

		if rule == nil {
			fmt.Printf("No matching rule found for %s\n", file)
			return nil
		}

		if err := executor.Execute(rule.Command, file, dryRun); err != nil {
			return fmt.Errorf("error executing command: %w", err)
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
