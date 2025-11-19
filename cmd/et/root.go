package main

import (
	"fmt"
	"net/url"
	"os"
	os_exec "os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/matcher"
	"github.com/charmbracelet/huh"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	dryRun  bool
)

func init() {
	rootCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/entry/config.yml)")
	rootCmd.Flags().Bool("dry-run", false, "print command instead of executing")
	rootCmd.Flags().BoolP("select", "s", false, "Interactive selection")
	rootCmd.Flags().Bool("explain", false, "Show detailed matching information")

	rootCmd.AddCommand(configCmd)
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
			} else if arg == "--config" {
				if i+1 < len(args) {
					cfgFile = args[i+1]
					i++ // Skip next arg
					continue
				} else {
					return fmt.Errorf("flag needs an argument: --config")
				}
			} else if strings.HasPrefix(arg, "--config=") {
				cfgFile = strings.TrimPrefix(arg, "--config=")
				continue
			} else if arg == "--version" || arg == "-v" {
				fmt.Println("et v0.1.0") // Assuming version is hardcoded or defined elsewhere
				return nil
			} else if arg == "--help" || arg == "-h" {
				return cmd.Help()
			} else if arg == "--select" || arg == "-s" || arg == "--explain" {
				// These flags are handled later, skip them here
				continue
			} else if arg == "--" { // Stop parsing at first non-flag or "--"
				commandArgs = args[i+1:]
				break
			} else if !strings.HasPrefix(arg, "-") {
				commandArgs = args[i:]
				break
			}
			// If we reach here, it's an unknown flag or a flag that doesn't take a value.
			// For now, assume it's part of the command if it looks like a flag but we don't know it.
			// This logic needs careful consideration if strict flag parsing is desired.
			// For this change, we'll assume any remaining flags are part of the command if no non-flag arg is found.
			// The original logic was:
			// commandArgs = args[i:]
			// break
			// This means if an unknown flag is encountered, it and everything after it becomes commandArgs.
			// Let's keep that behavior for now.
			commandArgs = args[i:]
			break
		}

		if len(commandArgs) == 0 {
			return fmt.Errorf("requires at least 1 argument")
		}

		// Check for built-in commands (manual dispatch)
		if commandArgs[0] == "config" {
			return handleConfigCommand(commandArgs[1:])
		}

		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		// Initialize Executor
		exec := executor.NewExecutor(cmd.OutOrStdout(), dryRun)

		// Check for --select flag
		interactive := false
		explain := false
		for _, arg := range args {
			if arg == "--select" {
				interactive = true
			}
			if arg == "--explain" {
				explain = true
			}
		}

		// Explain mode: show detailed matching information
		if explain && len(commandArgs) == 1 {
			filename := commandArgs[0]
			fmt.Fprintf(cmd.OutOrStdout(), "=== Explain Mode for: %s ===\n\n", filename)
			
			// Check if file/URL exists
			u, err := url.Parse(filename)
			isURL := err == nil && u.Scheme != ""
			
			if isURL {
				fmt.Fprintf(cmd.OutOrStdout(), "Type: URL\n")
				fmt.Fprintf(cmd.OutOrStdout(), "Scheme: %s\n", u.Scheme)
				if u.Path != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "Path: %s\n", u.Path)
					ext := filepath.Ext(u.Path)
					if ext != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "Extension: %s\n", strings.TrimPrefix(ext, "."))
					}
				}
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Type: File\n")
				ext := filepath.Ext(filename)
				if ext != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "Extension: %s\n", strings.TrimPrefix(ext, "."))
				}
				
				// Check MIME type
				if _, err := os.Stat(filename); err == nil {
					mtype, err := mimetype.DetectFile(filename)
					if err == nil {
						fmt.Fprintf(cmd.OutOrStdout(), "MIME Type: %s\n", mtype.String())
					}
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "File Status: Does not exist\n")
				}
			}
			
			fmt.Fprintf(cmd.OutOrStdout(), "\n=== Rule Evaluation ===\n\n")
			
			// Evaluate each rule
			matched := false
			for i, rule := range cfg.Rules {
				fmt.Fprintf(cmd.OutOrStdout(), "Rule #%d:\n", i+1)
				if rule.Name != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  Name: %s\n", rule.Name)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  Command: %s\n", rule.Command)
				
				// Check each condition
				ruleMatched := false
				reasons := []string{}
				
				// OS check
				if len(rule.OS) > 0 {
					osMatch := false
					for _, osName := range rule.OS {
						if strings.ToLower(osName) == runtime.GOOS {
							osMatch = true
							break
						}
					}
					if osMatch {
						reasons = append(reasons, fmt.Sprintf("OS matches (%s)", runtime.GOOS))
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "  ❌ OS mismatch (requires: %v, current: %s)\n", rule.OS, runtime.GOOS)
						fmt.Fprintf(cmd.OutOrStdout(), "\n")
						continue
					}
				}
				
				// Scheme check
				if rule.Scheme != "" {
					if isURL && strings.ToLower(u.Scheme) == strings.ToLower(rule.Scheme) {
						reasons = append(reasons, fmt.Sprintf("Scheme matches (%s)", rule.Scheme))
						ruleMatched = true
					} else if rule.Scheme != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "  ❌ Scheme mismatch (requires: %s)\n", rule.Scheme)
						fmt.Fprintf(cmd.OutOrStdout(), "\n")
						continue
					}
				}
				
				// Extension check
				if !ruleMatched && len(rule.Extensions) > 0 {
					var pathExt string
					if isURL {
						pathExt = filepath.Ext(u.Path)
					} else {
						pathExt = filepath.Ext(filename)
					}
					pathExt = strings.ToLower(strings.TrimPrefix(pathExt, "."))
					
					for _, ruleExt := range rule.Extensions {
						if strings.ToLower(ruleExt) == pathExt {
							reasons = append(reasons, fmt.Sprintf("Extension matches (.%s)", pathExt))
							ruleMatched = true
							break
						}
					}
					if !ruleMatched {
						fmt.Fprintf(cmd.OutOrStdout(), "  Extensions: %v (no match)\n", rule.Extensions)
					}
				}
				
				// Regex check
				if !ruleMatched && rule.Regex != "" {
					regexMatched, _ := regexp.MatchString(rule.Regex, filename)
					if regexMatched {
						reasons = append(reasons, fmt.Sprintf("Regex matches (%s)", rule.Regex))
						ruleMatched = true
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "  Regex: %s (no match)\n", rule.Regex)
					}
				}
				
				// MIME check
				if !ruleMatched && rule.Mime != "" && !isURL {
					if _, err := os.Stat(filename); err == nil {
						mtype, err := mimetype.DetectFile(filename)
						if err == nil {
							mimeMatched, _ := regexp.MatchString(rule.Mime, mtype.String())
							if mimeMatched {
								reasons = append(reasons, fmt.Sprintf("MIME matches (%s)", rule.Mime))
								ruleMatched = true
							} else {
								fmt.Fprintf(cmd.OutOrStdout(), "  MIME: %s (no match)\n", rule.Mime)
							}
						}
					}
				}
				
				if ruleMatched {
					fmt.Fprintf(cmd.OutOrStdout(), "  ✅ MATCHED: %s\n", strings.Join(reasons, ", "))
					if rule.Fallthrough {
						fmt.Fprintf(cmd.OutOrStdout(), "  Fallthrough: true (will continue to next rule)\n")
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "  Fallthrough: false (will stop here)\n")
					}
					matched = true
					if !rule.Fallthrough {
						fmt.Fprintf(cmd.OutOrStdout(), "\n")
						break
					}
				}
				
				fmt.Fprintf(cmd.OutOrStdout(), "\n")
			}
			
			if !matched {
				fmt.Fprintf(cmd.OutOrStdout(), "No rules matched.\n\n")
				if cfg.DefaultCommand != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "=== Default Command ===\n")
					fmt.Fprintf(cmd.OutOrStdout(), "Command: %s\n", cfg.DefaultCommand)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "=== System Default ===\n")
					fmt.Fprintf(cmd.OutOrStdout(), "Will use system default application.\n")
				}
			}
			
			return nil
		}

		// 1. Try to match rules
		if len(commandArgs) == 1 {
			filename := commandArgs[0]
			
			// If interactive, use MatchAll
			if interactive {
				matches, err := matcher.MatchAll(cfg.Rules, filename)
				if err != nil {
					return fmt.Errorf("error matching rules: %w", err)
				}

				// Add System Default option
				// We need to represent options.
				type Option struct {
					Label string
					Rule  *config.Rule
					IsSystem bool
				}
				var options []Option

				for _, m := range matches {
					label := m.Name
					if label == "" {
						label = fmt.Sprintf("Command: %s", m.Command)
					}
					options = append(options, Option{Label: label, Rule: m})
				}

				// Add System Default if applicable
				// Check if it is a URL or File to decide if system default is valid option
				isURL := false
				if u, err := url.Parse(filename); err == nil && u.Scheme != "" {
					isURL = true
				}
				if isURL || fileExists(filename) {
					options = append(options, Option{Label: "System Default", IsSystem: true})
				}

				if len(options) == 0 {
					return fmt.Errorf("no matching rules found for %s", filename)
				}

				// If only one option and it's system default, just run it?
				// Or if user asked for --select, always show?
				// Let's always show if --select is present.

				var selected Option
				
				// Use huh for selection
				// We need to map options to huh options
				var huhOptions []huh.Option[Option]
				for _, opt := range options {
					huhOptions = append(huhOptions, huh.NewOption(opt.Label, opt))
				}

				form := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[Option]().
							Title("Select action for " + filename).
							Options(huhOptions...).
							Value(&selected),
					),
				)

				if err := form.Run(); err != nil {
					return err
				}

				if selected.IsSystem {
					if cfg.DefaultCommand != "" {
						return exec.Execute(cfg.DefaultCommand, filename, executor.ExecutionOptions{})
					}
					return exec.OpenSystem(filename)
				} else {
					opts := executor.ExecutionOptions{
						Background: selected.Rule.Background,
						Terminal:   selected.Rule.Terminal,
					}
					return exec.Execute(selected.Rule.Command, filename, opts)
				}
			}

			// Non-interactive normal flow
			rules, err := matcher.Match(cfg.Rules, filename)
			if err != nil {
				return fmt.Errorf("error matching rule: %w", err)
			}
			if len(rules) > 0 {
				// Execute all matched rules (fallthrough support)
				for _, rule := range rules {
					opts := executor.ExecutionOptions{
						Background: rule.Background,
						Terminal:   rule.Terminal,
					}
					if err := exec.Execute(rule.Command, filename, opts); err != nil {
						return err
					}
				}
				return nil
			}

			// Check if it is a URL or File
			isURL := false
			if u, err := url.Parse(filename); err == nil && u.Scheme != "" {
				isURL = true
			}
			
			if isURL {
				if cfg.DefaultCommand != "" {
					return exec.Execute(cfg.DefaultCommand, filename, executor.ExecutionOptions{})
				}
				return exec.OpenSystem(filename)
			}

			if _, err := os.Stat(filename); err == nil {
				if cfg.DefaultCommand != "" {
					return exec.Execute(cfg.DefaultCommand, filename, executor.ExecutionOptions{})
				}
				return exec.OpenSystem(filename)
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

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}


func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
