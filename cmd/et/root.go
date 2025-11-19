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
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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
			
			// Define styles
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("12")).
				Padding(0, 2).
				MarginTop(1).
				MarginBottom(1)
			
			sectionTitleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("14")).
				Padding(0, 1).
				MarginTop(1)
			
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Width(12)
			
			valueStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("15"))
			
			matchStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)
			
			skipStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Faint(true)
			
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("9"))
			
			// Title
			fmt.Fprintln(cmd.OutOrStdout(), titleStyle.Render("═══ EXPLAIN MODE ═══"))
			fmt.Fprintln(cmd.OutOrStdout(), "")
			
			// Check if file/URL exists
			u, err := url.Parse(filename)
			isURL := err == nil && u.Scheme != ""
			
			// File Information Section
			fmt.Fprintln(cmd.OutOrStdout(), sectionTitleStyle.Render("FILE INFORMATION"))
			fmt.Fprintln(cmd.OutOrStdout(), "")
			
			fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Path:")+" "+valueStyle.Render(filename))
			
			if isURL {
				fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Type:")+" "+valueStyle.Render("URL"))
				fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Scheme:")+" "+valueStyle.Render(u.Scheme))
				if u.Path != "" {
					ext := filepath.Ext(u.Path)
					if ext != "" {
						fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Extension:")+" "+valueStyle.Render(strings.TrimPrefix(ext, ".")))
					}
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Type:")+" "+valueStyle.Render("File"))
				ext := filepath.Ext(filename)
				if ext != "" {
					fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Extension:")+" "+valueStyle.Render(strings.TrimPrefix(ext, ".")))
				}
				
				// Check MIME type
				if _, err := os.Stat(filename); err == nil {
					mtype, err := mimetype.DetectFile(filename)
					if err == nil {
						fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  MIME Type:")+" "+valueStyle.Render(mtype.String()))
					}
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Status:")+" "+errorStyle.Render("Does not exist"))
				}
			}
			
			// Rule Evaluation Section
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), sectionTitleStyle.Render("RULE EVALUATION"))
			fmt.Fprintln(cmd.OutOrStdout(), "")
			
			// Prepare table data
			type ruleResult struct {
				num        string
				name       string
				conditions []string
				result     string
				matched    bool
			}
			
			var results []ruleResult
			matched := false
			
			for i, rule := range cfg.Rules {
				result := ruleResult{
					num:  fmt.Sprintf("%d", i+1),
					name: rule.Name,
				}
				
				if result.name == "" {
					result.name = "-"
				}
				
				// Check each condition
				ruleMatched := false
				var matchReasons []string
				
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
						result.conditions = append(result.conditions, "✓ OS: "+runtime.GOOS)
						matchReasons = append(matchReasons, "OS")
					} else {
						result.conditions = append(result.conditions, "✗ OS mismatch")
						result.result = errorStyle.Render("SKIP")
						results = append(results, result)
						continue
					}
				}
				
				// Scheme check
				if rule.Scheme != "" {
					if isURL && strings.ToLower(u.Scheme) == strings.ToLower(rule.Scheme) {
						result.conditions = append(result.conditions, "✓ Scheme: "+rule.Scheme)
						matchReasons = append(matchReasons, "Scheme")
						ruleMatched = true
					} else {
						result.conditions = append(result.conditions, "✗ Scheme: "+rule.Scheme)
						result.result = errorStyle.Render("SKIP")
						results = append(results, result)
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
					
					extMatched := false
					for _, ruleExt := range rule.Extensions {
						if strings.ToLower(ruleExt) == pathExt {
							result.conditions = append(result.conditions, "✓ Ext: ."+pathExt)
							matchReasons = append(matchReasons, "Extension")
							ruleMatched = true
							extMatched = true
							break
						}
					}
					if !extMatched {
						result.conditions = append(result.conditions, skipStyle.Render("○ Ext: "+fmt.Sprintf("%v", rule.Extensions)))
					}
				}
				
				// Regex check
				if !ruleMatched && rule.Regex != "" {
					regexMatched, _ := regexp.MatchString(rule.Regex, filename)
					if regexMatched {
						result.conditions = append(result.conditions, "✓ Regex")
						matchReasons = append(matchReasons, "Regex")
						ruleMatched = true
					} else {
						result.conditions = append(result.conditions, skipStyle.Render("○ Regex: "+rule.Regex))
					}
				}
				
				// MIME check
				if !ruleMatched && rule.Mime != "" && !isURL {
					if _, err := os.Stat(filename); err == nil {
						mtype, err := mimetype.DetectFile(filename)
						if err == nil {
							mimeMatched, _ := regexp.MatchString(rule.Mime, mtype.String())
							if mimeMatched {
								result.conditions = append(result.conditions, "✓ MIME")
								matchReasons = append(matchReasons, "MIME")
								ruleMatched = true
							} else {
								result.conditions = append(result.conditions, skipStyle.Render("○ MIME: "+rule.Mime))
							}
						}
					}
				}
				
				if ruleMatched {
					result.result = matchStyle.Render("[MATCH]")
					if rule.Fallthrough {
						result.result += skipStyle.Render(" →")
					}
					result.matched = true
					matched = true
					results = append(results, result)
					if !rule.Fallthrough {
						break
					}
				} else {
					result.result = skipStyle.Render("—")
					results = append(results, result)
				}
			}
			
			// Create table
			t := table.New().
				Border(lipgloss.RoundedBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
				StyleFunc(func(row, col int) lipgloss.Style {
					if row == 0 {
						return lipgloss.NewStyle().
							Foreground(lipgloss.Color("14")).
							Bold(true).
							Padding(0, 1)
					}
					return lipgloss.NewStyle().Padding(0, 1)
				}).
				Headers("#", "Rule Name", "Conditions", "Result")
			
			for _, r := range results {
				condStr := strings.Join(r.conditions, "\n")
				t.Row(r.num, r.name, condStr, r.result)
			}
			
			fmt.Fprintln(cmd.OutOrStdout(), t.Render())
			
			// Result Section
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), sectionTitleStyle.Render("RESULT"))
			fmt.Fprintln(cmd.OutOrStdout(), "")
			
			if !matched {
				fmt.Fprintln(cmd.OutOrStdout(), "  "+skipStyle.Render("No rules matched."))
				fmt.Fprintln(cmd.OutOrStdout(), "")
				if cfg.DefaultCommand != "" {
					fmt.Fprintln(cmd.OutOrStdout(), "  "+labelStyle.Render("Action:")+" "+valueStyle.Render("Execute default command"))
					fmt.Fprintln(cmd.OutOrStdout(), "  "+labelStyle.Render("Command:")+" "+valueStyle.Render(cfg.DefaultCommand))
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "  "+labelStyle.Render("Action:")+" "+valueStyle.Render("Use system default application"))
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "  "+matchStyle.Render("[OK] Rule matched successfully"))
			}
			
			fmt.Fprintln(cmd.OutOrStdout(), "")
			
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
