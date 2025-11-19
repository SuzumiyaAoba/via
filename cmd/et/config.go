package main

import (
	"fmt"
	"os"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage the entry configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigList()
	},
}

var configOpenCmd = &cobra.Command{
	Use:   "open",
	Short: "Open configuration file in editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigOpen()
	},
}

var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigAdd(cmd, args)
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configOpenCmd)
	configCmd.AddCommand(configAddCmd)

	configAddCmd.Flags().String("ext", "", "Extension to match (comma separated)")
	configAddCmd.Flags().String("cmd", "", "Command to execute")
	configAddCmd.MarkFlagRequired("cmd")
}

func handleConfigCommand(args []string) error {
	// We need to manually dispatch or use configCmd with custom args
	// Since configCmd is attached to rootCmd, executing it might be tricky if we don't want full root execution.
	// But we can just use it for Help and dispatch manually, OR try to use it for execution.
	
	// Let's try to use SetArgs and Execute? 
	// But configCmd.Execute() calls the root command's Execute usually? No, it calls the command's Execute.
	// However, configCmd is a child.
	
	// Simplest way: Use manual dispatch but use Cobra for 'add' flag parsing if possible, 
	// or just keep manual dispatch and use configCmd ONLY for help registration in root.
	
	// User wants "help" to show config.
	// So configCmd MUST be added to rootCmd.
	
	// For execution:
	if len(args) == 0 {
		return configCmd.Help()
	}

	// Check for help flag in args to show config help
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return configCmd.Help()
		}
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "list":
		return runConfigList()
	case "open":
		return runConfigOpen()
	case "add":
		// Use Cobra's flag parsing for add?
		// We can use configAddCmd.ParseFlags(subargs)
		if err := configAddCmd.ParseFlags(subargs); err != nil {
			return err
		}
		// And then call RunE
		return runConfigAdd(configAddCmd, configAddCmd.Flags().Args())
	default:
		return fmt.Errorf("unknown config subcommand: %s", subcmd)
	}
}

func runConfigList() error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func runConfigOpen() error {
	configPath, err := config.GetConfigPath(cfgFile)
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := &config.Config{Version: "1"}
		if err := config.SaveConfig(cfgFile, cfg); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
		fmt.Printf("Created default config at %s\n", configPath)
	}

	exec := executor.NewExecutor(os.Stdout, false)
	return exec.OpenSystem(configPath)
}

func runConfigAdd(cmd *cobra.Command, args []string) error {
	ext, _ := cmd.Flags().GetString("ext")
	command, _ := cmd.Flags().GetString("cmd")

	if command == "" {
		return fmt.Errorf("--cmd is required")
	}

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		cfg = &config.Config{Version: "1"}
	}

	rule := config.Rule{
		Command: command,
	}

	if ext != "" {
		rule.Extensions = []string{ext}
	}

	cfg.Rules = append(cfg.Rules, rule)

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return err
	}

	fmt.Println("Rule added successfully")
	return nil
}
