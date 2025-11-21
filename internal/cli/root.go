package cli

import (
	"fmt"
	"os"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/logger"
	"github.com/spf13/cobra"
)


var (
	cfgFile     string
	dryRun      bool
	interactive bool
	explain     bool
	verbose     bool
	profile     string
)

func init() {
	rootCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/entry/config.yml)")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "print command instead of executing")
	rootCmd.Flags().BoolVarP(&interactive, "select", "s", false, "Interactive selection")
	rootCmd.Flags().BoolVar(&explain, "explain", false, "Show detailed matching information")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().StringVarP(&profile, "profile", "p", "", "Configuration profile to use")
	rootCmd.RegisterFlagCompletionFunc("profile", CompletionProfiles)
	
	// Allow flags after positional arguments to be passed to the command
	rootCmd.Flags().SetInterspersed(false)

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(completionCmd)
}

var rootCmd = &cobra.Command{
	Use:     "et <file>",
	Short:   "Entry is a CLI file association tool",
	Long:    `Entry allows you to execute specific commands based on file extensions or regex patterns matched against a provided file argument.`,
	Version: Version,
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for profile environment variable
		if profile == "" && os.Getenv("ENTRY_PROFILE") != "" {
			profile = os.Getenv("ENTRY_PROFILE")
		}

		// Resolve config file path with profile
		if cfgFile == "" && profile != "" {
			resolvedPath, err := config.GetConfigPathWithProfile("", profile)
			if err != nil {
				return fmt.Errorf("failed to resolve profile config path: %w", err)
			}
			cfgFile = resolvedPath
			logger.Debug("Using profile '%s' with config: %s", profile, cfgFile)
		}

		// Initialize logger
		if err := initLogger(); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		defer logger.GetGlobal().Close()

		logger.Debug("Starting entry with args: %v", args)
		logger.Debug("Flags - verbose: %v, dryRun: %v, interactive: %v, explain: %v, profile: %s", verbose, dryRun, interactive, explain, profile)

		// Args are already parsed by Cobra
		// args[0] is the file/command
		// args[1:] are arguments to the command (if matched) or part of the command

		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		// Initialize Executor
		exec := executor.NewExecutor(cmd.OutOrStdout(), dryRun)

		// Explain mode: show detailed matching information
		// Only valid if we have exactly one argument (the file)
		if explain {
			if len(args) == 1 {
				return handleExplain(cmd, cfg, args[0])
			}
			// If more args, maybe warn? For now, just ignore explain or error?
			// Existing behavior was: if explain && len(commandArgs) == 1
			// Let's stick to that.
		}

		// Handle file execution with single argument
		if len(args) == 1 {
			filename := args[0]
			
			// Interactive mode
			if interactive {
				return handleInteractive(cfg, exec, filename)
			}

			// Normal file execution - if it fails, try as command
			if err := handleFileExecution(cfg, exec, filename); err == nil {
				return nil
			}
			// If file execution failed, fall through to command execution
		}

		// Handle command execution with multiple arguments
		// Or if file execution failed
		return handleCommandExecution(cfg, exec, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}


