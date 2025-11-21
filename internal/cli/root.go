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

	// Subcommands are added in runRoot to allow manual dispatch and prevent
	// Cobra from hijacking execution when arguments match subcommand names
	// but are intended as file arguments (e.g. "et config").
	// rootCmd.AddCommand(configCmd)
	// rootCmd.AddCommand(completionCmd)
}

var rootCmd = &cobra.Command{
	Use:     "et <file>",
	Short:   "Entry is a CLI file association tool",
	Long:    `Entry allows you to execute specific commands based on file extensions or regex patterns matched against a provided file argument.`,
	Version: Version,
	DisableFlagParsing: true,
	Args: cobra.ArbitraryArgs,
	RunE: runRoot,
}

func runRoot(cmd *cobra.Command, args []string) error {
	// Temporarily add subcommands to allow Find to work and Help to show them
	cmd.AddCommand(configCmd)
	cmd.AddCommand(completionCmd)
	defer cmd.RemoveCommand(configCmd)
	defer cmd.RemoveCommand(completionCmd)

	// Manually parse flags
	// Note: We must use cmd.Flags().Parse() directly because cmd.ParseFlags()
	// returns early if DisableFlagParsing is true.
	if err := cmd.Flags().Parse(args); err != nil {
		return err
	}

	// Handle help flag manually since parsing is disabled
	if helpVal, _ := cmd.Flags().GetBool("help"); helpVal {
		return cmd.Help()
	}

	// Handle version flag manually
	if versionVal, _ := cmd.Flags().GetBool("version"); versionVal {
		fmt.Printf("et version %s\n", Version)
		return nil
	}

	remainingArgs := cmd.Flags().Args()
	dashLen := cmd.Flags().ArgsLenAtDash()

	// Check if the command was invoked with `--` to separate flags from subcommand
	if dashLen != -1 {
		if len(remainingArgs) > 0 {
			subCmd, subArgs, err := cmd.Find(remainingArgs)
			if err == nil && subCmd != cmd {
				// Detach subcommand from parent to prevent infinite recursion
				// since subCmd.Execute() calls Root().Execute() which would loop
				parent := subCmd.Parent()
				if parent != nil {
					parent.RemoveCommand(subCmd)
					defer parent.AddCommand(subCmd)
				}

				// Inherit output from root command
				subCmd.SetOut(cmd.OutOrStdout())
				subCmd.SetErr(cmd.ErrOrStderr())

				subCmd.SetArgs(subArgs)
				return subCmd.Execute()
			}
		}
	}

	args = remainingArgs

	if len(args) < 1 {
		return cmd.Help()
	}

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

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Initialize Executor
	exec := executor.NewExecutor(cmd.OutOrStdout(), dryRun)

	// Explain mode: show detailed matching information
	if explain {
		if len(args) == 1 {
			return handleExplain(cmd, cfg, args[0])
		}
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
	}

	// Handle command execution with multiple arguments
	// Or if file execution failed
	return handleCommandExecution(cfg, exec, args)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}


