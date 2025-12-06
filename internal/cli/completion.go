package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   ":completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(vv :completion bash)
 
   # To load completions for each session, execute once:
   # Linux:
   $ vv :completion bash > /etc/bash_completion.d/vv
   # macOS:
   $ vv :completion bash > /usr/local/etc/bash_completion.d/vv

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
   $ vv :completion zsh > "${fpath[1]}/_vv"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ vv :completion fish | source
 
   # To load completions for each session, execute once:
   $ vv :completion fish > ~/.config/fish/completions/vv.fish

PowerShell:

  PS> vv :completion powershell | Out-String | Invoke-Expression
 
   # To load completions for each session, execute once:
   PS> vv :completion powershell > vv.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		}
	},
}

// CompletionProfiles provides shell completion for profile names
func CompletionProfiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	profilesDir := filepath.Join(home, ".config", "via", "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveError
	}

	profiles := lo.FilterMap(entries, func(entry os.DirEntry, _ int) (string, bool) {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			return "", false
		}
		profileName := strings.TrimSuffix(entry.Name(), ".yml")
		if !strings.HasPrefix(profileName, toComplete) {
			return "", false
		}
		return profileName, true
	})

	return profiles, cobra.ShellCompDirectiveNoFileComp
}
