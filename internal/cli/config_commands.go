package cli

import (
	"github.com/spf13/cobra"
)

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigInit()
	},
}

var configCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check configuration validity",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigCheck()
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <index>",
	Short: "Remove a rule by index (1-based)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigRemove(args[0])
	},
}

var configSetDefaultCmd = &cobra.Command{
	Use:   "set-default <command>",
	Short: "Set default command",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSetDefault(args[0])
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an existing rule interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigEdit()
	},
}
