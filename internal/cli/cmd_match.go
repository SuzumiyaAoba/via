package cli

import (
	"fmt"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/matcher"
	"github.com/spf13/cobra"
)

var matchCmd = &cobra.Command{
	Use:   ":match <file>",
	Short: "Check if a file matches any rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMatch(cmd, args[0])
	},
}

func runMatch(cmd *cobra.Command, filename string) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	matches, err := matcher.Match(cfg.Rules, filename)
	if err != nil {
		return err
	}

	if len(matches) > 0 {
		rule := matches[0]
		name := rule.Name
		if name == "" {
			name = rule.Command
		}
		fmt.Fprintln(cmd.OutOrStdout(), name)
		return nil
	}

	// No match found
	return fmt.Errorf("no match found")
}
