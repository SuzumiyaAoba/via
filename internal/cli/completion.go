package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(et completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ et completion bash > /etc/bash_completion.d/et
  # macOS:
  $ et completion bash > /usr/local/etc/bash_completion.d/et

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ et completion zsh > "${fpath[1]}/_et"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ et completion fish | source

  # To load completions for each session, execute once:
  $ et completion fish > ~/.config/fish/completions/et.fish

PowerShell:

  PS> et completion powershell | Out-String | Invoke-Expression

  # To load completions for each session, execute once:
  PS> et completion powershell > et.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

// CompletionProfiles provides shell completion for profile names
func CompletionProfiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	profilesDir := filepath.Join(home, ".config", "entry", "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveError
	}

	var profiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yml") {
			profileName := strings.TrimSuffix(entry.Name(), ".yml")
			if strings.HasPrefix(profileName, toComplete) {
				profiles = append(profiles, profileName)
			}
		}
	}

	return profiles, cobra.ShellCompDirectiveNoFileComp
}
