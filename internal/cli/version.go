package cli

import (
	"github.com/spf13/cobra"
)

var Version = "dev"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   ":version",
	Short: "Print the version number of entry",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("et %s\n", Version)
	},
}
