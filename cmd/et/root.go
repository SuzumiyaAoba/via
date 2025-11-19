package main

import (
	"fmt"
	"net/url"
	"os"
	os_exec "os/exec"
	"strings"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/matcher"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	dryRun  bool
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/entry/config.yml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print command instead of executing")
}

var rootCmd = &cobra.Command{
	Use:   "et <file>",
	Short: "Entry is a CLI file association tool",
	Long:  `Entry allows you to execute specific commands based on file extensions or regex patterns matched against a provided file argument.`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Manual flag parsing
		var commandArgs []string
		for i := 0; i < len(args); i++ {
			arg := args[i]
			if arg == "--version" || arg == "-v" {
				fmt.Println("et v0.1.0")
				return nil
			}
			if arg == "--help" || arg == "-h" {
				return cmd.Help()
			}
			if arg == "--dry-run" {
				dryRun = true
				continue
			}
			if arg == "--config" {
				if i+1 < len(args) {
					cfgFile = args[i+1]
					i++ // Skip value
					continue
				} else {
					return fmt.Errorf("flag needs an argument: --config")
				}
			}
			if strings.HasPrefix(arg, "--config=") {
				cfgFile = strings.TrimPrefix(arg, "--config=")
				continue
			}
			// Stop parsing at first non-flag or "--"
			if arg == "--" {
				commandArgs = args[i+1:]
				break
			}
			if !strings.HasPrefix(arg, "-") {
				commandArgs = args[i:]
				break
			}
			// If we encounter an unknown flag before the command, what do we do?
			// For now, assume it's part of the command if it looks like a flag but we don't know it?
			// But we are in DisableFlagParsing, so we are the parser.
			// If we see "-la" and we haven't seen the command yet, it's ambiguous.
			// But usually `entry -la` is invalid if -la is not an entry flag.
			// However, `entry ls -la` -> `ls` is first non-flag.
			// What if `entry -v` (unknown flag) `ls`?
			// Let's assume all flags before the first positional arg MUST be entry flags.
			// If we hit a flag we don't know, and we haven't found command yet, it's an error OR it belongs to the command?
			// If `entry` is a wrapper, maybe we should be strict about entry flags.
			// But wait, `entry ls -la`. `ls` is not a flag.
			// So we loop until we find a non-flag argument.
			// That argument is the start of the command.
			commandArgs = args[i:]
			break
		}

		if len(commandArgs) == 0 {
			return fmt.Errorf("requires at least 1 argument")
		}

		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		// Initialize Executor
		exec := executor.NewExecutor(cmd.OutOrStdout(), dryRun)

		// 1. Try to match a specific rule if single argument
		if len(commandArgs) == 1 {
			rule, err := matcher.Match(cfg.Rules, commandArgs[0])
			if err != nil {
				return fmt.Errorf("error matching rule: %w", err)
			}
			if rule != nil {
				opts := executor.ExecutionOptions{
					Background: rule.Background,
					Terminal:   rule.Terminal,
				}
				return exec.Execute(rule.Command, commandArgs[0], opts)
			}

			// Check if it is a URL or File
			isURL := false
			if u, err := url.Parse(commandArgs[0]); err == nil && u.Scheme != "" {
				isURL = true
			}
			
			if isURL {
				if cfg.DefaultCommand != "" {
					return exec.Execute(cfg.DefaultCommand, commandArgs[0], executor.ExecutionOptions{})
				}
				return exec.OpenSystem(commandArgs[0])
			}

			if _, err := os.Stat(commandArgs[0]); err == nil {
				if cfg.DefaultCommand != "" {
					return exec.Execute(cfg.DefaultCommand, commandArgs[0], executor.ExecutionOptions{})
				}
				return exec.OpenSystem(commandArgs[0])
			}
		}

		command := commandArgs[0]
		cmdArgs := commandArgs[1:]

		// 2. Check aliases
		if alias, ok := cfg.Aliases[command]; ok {
			command = alias
			return exec.ExecuteCommand(command, cmdArgs)
		}

		// 3. Fallback to command execution
		// Check if command exists in PATH
		if _, err := os_exec.LookPath(command); err != nil {
			// Command not found.
			// If single argument and default command exists, assume it's a new file and use default command.
			if len(commandArgs) == 1 && cfg.DefaultCommand != "" {
				return exec.Execute(cfg.DefaultCommand, commandArgs[0], executor.ExecutionOptions{})
			}
			// If no default command, or multiple args, let ExecuteCommand fail (or return the LookPath error)
			// Actually, ExecuteCommand will try to run it and fail.
			// But we can return the LookPath error directly to be helpful, or proceed to ExecuteCommand to let it fail naturally.
			// Let's proceed to ExecuteCommand so it handles dryRun printing etc?
			// But dryRun prints the command. If we want to print default_command in dryRun, we must decide HERE.
			// So if LookPath fails, we MUST use default_command if applicable.
		}

		return exec.ExecuteCommand(command, cmdArgs)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
