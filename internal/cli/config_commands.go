package cli

import (
	"github.com/spf13/cobra"
)

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigInit(cmd)
	},
}

var configCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check configuration validity",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigCheck(cmd)
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <index>",
	Short: "Remove a rule by index",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigRemove(cmd, args[0])
	},
}

var configSetDefaultCmd = &cobra.Command{
	Use:   "set-default <command>",
	Short: "Set default command",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSetDefault(cmd, args[0])
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigEdit(cmd)
	},
}
