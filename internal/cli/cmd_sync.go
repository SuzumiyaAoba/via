package cli

import (
	"fmt"
	"os"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/sync"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var configSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync configuration with Gist",
}

var configSyncInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize sync configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSyncInit(cmd)
	},
}

var configSyncPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push configuration to Gist",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSyncPush(cmd)
	},
}

var configSyncPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull configuration from Gist",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSyncPull(cmd)
	},
}

func init() {
	configSyncCmd.AddCommand(configSyncInitCmd)
	configSyncCmd.AddCommand(configSyncPushCmd)
	configSyncCmd.AddCommand(configSyncPullCmd)
}

func runConfigSyncInit(cmd *cobra.Command) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	var (
		gistID string
		token  string
		create bool
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Create new Gist?").
				Description("If no, you must provide an existing Gist ID").
				Value(&create),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if !create {
		form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Gist ID").
					Value(&gistID),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
	}

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("GitHub Token").
				Description("Required for private Gists or creating new ones").
				EchoMode(huh.EchoModePassword).
				Value(&token),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	client := sync.NewClient(token)

	if create {
		id, err := client.CreateGist(cfg, false) // Default to private
		if err != nil {
			return err
		}
		gistID = id
		fmt.Fprintf(cmd.OutOrStdout(), "Created new Gist: %s\n", gistID)
	}

	if cfg.Sync == nil {
		cfg.Sync = &config.SyncConfig{}
	}
	cfg.Sync.GistID = gistID
	// We don't store token by default for security, but user can add it manually if they want.
	// Or we can store it if they want.
	// For now, let's ask.
	
	var storeToken bool
	confirm := huh.NewConfirm().
		Title("Store token in config?").
		Description("WARNING: This is insecure as the config file is plain text.").
		Value(&storeToken)
	
	if err := confirm.Run(); err != nil {
		return err
	}

	if storeToken {
		cfg.Sync.Token = token
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Token not stored. You will need to provide it via ENTRY_GITHUB_TOKEN env var.")
	}

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Sync initialized successfully")
	return nil
}

func runConfigSyncPush(cmd *cobra.Command) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if cfg.Sync == nil || cfg.Sync.GistID == "" {
		return fmt.Errorf("sync not initialized. Run 'et :config sync init' first")
	}

	token := cfg.Sync.Token
	if token == "" {
		token = os.Getenv("ENTRY_GITHUB_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("token not found. Set ENTRY_GITHUB_TOKEN or store it in config")
	}

	client := sync.NewClient(token)
	if err := client.UpdateGist(cfg.Sync.GistID, cfg); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Configuration pushed to Gist")
	return nil
}

func runConfigSyncPull(cmd *cobra.Command) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if cfg.Sync == nil || cfg.Sync.GistID == "" {
		return fmt.Errorf("sync not initialized. Run 'et :config sync init' first")
	}

	token := cfg.Sync.Token
	if token == "" {
		token = os.Getenv("ENTRY_GITHUB_TOKEN")
	}
	// Token might not be needed for public gists, but usually good to have.
	
	client := sync.NewClient(token)
	newCfg, err := client.GetGist(cfg.Sync.GistID)
	if err != nil {
		return err
	}

	// Preserve local sync settings
	newCfg.Sync = cfg.Sync

	if err := config.SaveConfig(cfgFile, newCfg); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Configuration pulled from Gist")
	return nil
}
