package cli

import (
	"fmt"
	"os"

	"github.com/SuzumiyaAoba/via/internal/config"
	"github.com/SuzumiyaAoba/via/internal/executor"
	"github.com/SuzumiyaAoba/via/internal/logger"
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/via/config.yml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print command instead of executing")
	rootCmd.Flags().BoolVarP(&interactive, "select", "s", false, "Interactive selection")
	rootCmd.Flags().BoolVar(&explain, "explain", false, "Show detailed matching information")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "Configuration profile to use")
	rootCmd.RegisterFlagCompletionFunc("profile", CompletionProfiles)
	
	// Allow flags after positional arguments to be passed to the command
	rootCmd.Flags().SetInterspersed(false)

	// Define custom help command with colon prefix
	helpCmd := &cobra.Command{
		Use:   ":help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any command in the application.
Simply type vv :help [path to command] for full details.`,
		Run: func(c *cobra.Command, args []string) {
			cmd, _, e := c.Root().Find(args)
			if cmd == nil || e != nil {
				c.Printf("Unknown help topic %#q\n", args)
				c.Root().Usage()
			} else {
				cmd.InitDefaultHelpFlag() // make sure help flag is init
				cmd.Help()
			}
		},
	}
	rootCmd.SetHelpCommand(helpCmd)

	// Register subcommands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(matchCmd)
}

var rootCmd = &cobra.Command{
	Use:     "vv <file>",
	Short:   "Via is a CLI file association tool",
	Long:    `Via allows you to execute specific commands based on file extensions or regex patterns matched against a provided file argument.`,
	Version: Version,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	DisableFlagParsing: true,
	Args: cobra.ArbitraryArgs,
	// PersistentPreRunE is called for subcommands
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// If DisableFlagParsing is true (root command), we skip this because
		// we handle it manually in RunE.
		if cmd.DisableFlagParsing {
			return nil
		}
		return initialize(cmd, args)
	},
	RunE: runRoot,
}

// initialize handles common setup like logging and config loading
func initialize(cmd *cobra.Command, args []string) error {
	// Check for profile environment variable
	if profile == "" && os.Getenv("VIA_PROFILE") != "" {
		profile = os.Getenv("VIA_PROFILE")
	}

	// Resolve config file path with profile
	if cfgFile == "" && profile != "" {
		resolvedPath, err := config.GetConfigPathWithProfile("", profile)
		if err != nil {
			return fmt.Errorf("failed to resolve profile config path: %w", err)
		}
		cfgFile = resolvedPath
	}

	// Initialize logger
	if err := initLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	
	// We defer logger closing in main, not here, or we let the process exit handle it
	// defer logger.GetGlobal().Close()

	logger.Debug("Initialized with flags - verbose: %v, dryRun: %v, profile: %s, config: %s", verbose, dryRun, profile, cfgFile)
	return nil
}

func runRoot(cmd *cobra.Command, args []string) error {
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
		fmt.Fprintf(cmd.OutOrStdout(), "vv version %s\n", Version)
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
				
				// Ensure persistent flags are parsed for subcommand
				subCmd.ParseFlags(subArgs)

				subCmd.SetArgs(subArgs)
				return subCmd.Execute()
			}
		}
	}

	args = remainingArgs

	if len(args) < 1 {
		return cmd.Help()
	}

	// Check for system commands (prefixed with :)
	// This ensures 'vv :config' runs the config command
	// 'vv config' will fall through to file/alias execution
	if len(args) > 0 {
		subCmd, subArgs, err := cmd.Find(args)
		if err == nil && subCmd != cmd {
			// Found a subcommand!
			// Detach subcommand from parent to prevent infinite recursion
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
	
	// Manual initialization for root command since PersistentPreRunE is skipped
	if err := initialize(cmd, args); err != nil {
		return err
	}
	defer logger.GetGlobal().Close()

	logger.Debug("Starting via execution with args: %v", args)

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


