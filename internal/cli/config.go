package cli

import (
	"fmt"
	"os"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/charmbracelet/huh"
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
	configAddCmd.Flags().String("name", "", "Rule name")
	configAddCmd.Flags().String("regex", "", "Regex pattern to match")
	configAddCmd.Flags().String("mime", "", "MIME type pattern to match")
	configAddCmd.Flags().String("scheme", "", "URL scheme to match")
	configAddCmd.Flags().Bool("terminal", false, "Run in terminal")
	configAddCmd.Flags().Bool("background", false, "Run in background")
	configAddCmd.Flags().Bool("fallthrough", false, "Continue matching other rules")
	configAddCmd.Flags().StringSlice("os", nil, "OS constraints (e.g. darwin, linux)")
	configAddCmd.MarkFlagRequired("cmd")

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configCheckCmd)
	configCmd.AddCommand(configRemoveCmd)
	configCmd.AddCommand(configSetDefaultCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configProfileListCmd)
	configCmd.AddCommand(configProfileCopyCmd)
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

	name, _ := cmd.Flags().GetString("name")
	regex, _ := cmd.Flags().GetString("regex")
	mime, _ := cmd.Flags().GetString("mime")
	scheme, _ := cmd.Flags().GetString("scheme")
	terminal, _ := cmd.Flags().GetBool("terminal")
	background, _ := cmd.Flags().GetBool("background")
	isFallthrough, _ := cmd.Flags().GetBool("fallthrough")
	osList, _ := cmd.Flags().GetStringSlice("os")

	if command == "" {
		return fmt.Errorf("--cmd is required")
	}

	// Validate regex and mime
	if regex != "" {
		if err := config.ValidateRegex(regex); err != nil {
			return fmt.Errorf("invalid regex: %w", err)
		}
	}
	if mime != "" {
		if err := config.ValidateRegex(mime); err != nil {
			return fmt.Errorf("invalid MIME pattern: %w", err)
		}
	}

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		cfg = &config.Config{Version: "1"}
	}

	rule := config.Rule{
		Name:        name,
		Command:     command,
		Regex:       regex,
		Mime:        mime,
		Scheme:      scheme,
		Terminal:    terminal,
		Background:  background,
		Fallthrough: isFallthrough,
		OS:          osList,
	}

	if ext != "" {
		rule.Extensions = splitAndTrim(ext)
	}

	cfg.Rules = append(cfg.Rules, rule)

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return err
	}

	fmt.Println("Rule added successfully")
	return nil
}

func runConfigInit() error {
	configPath, err := config.GetConfigPath(cfgFile)
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	cfg := &config.Config{
		Version: "1",
		DefaultCommand: "vim {{.File}}",
		Rules: []config.Rule{
			{
				Name: "Example Rule",
				Extensions: []string{"txt", "md"},
				Command: "cat {{.File}}",
				Terminal: true,
			},
		},
	}

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}

	fmt.Printf("Created default config at %s\n", configPath)
	return nil
}

func runConfigCheck() error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if err := config.ValidateConfig(cfg); err != nil {
		return err
	}

	fmt.Println("Configuration is valid")
	return nil
}

func runConfigRemove(indexStr string) error {
	var index int
	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		return fmt.Errorf("invalid index: %s", indexStr)
	}

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if index < 1 || index > len(cfg.Rules) {
		return fmt.Errorf("index out of range: %d", index)
	}

	// Remove rule (1-based index)
	cfg.Rules = append(cfg.Rules[:index-1], cfg.Rules[index:]...)

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return err
	}

	fmt.Println("Rule removed successfully")
	return nil
}

func runConfigSetDefault(command string) error {
	// Check if config exists first
	configPath, err := config.GetConfigPath(cfgFile)
	if err != nil {
		return err
	}

	var cfg *config.Config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg = &config.Config{Version: "1"}
	} else {
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			return err
		}
	}

	cfg.DefaultCommand = command
	// Clear alias if present to avoid confusion
	cfg.Default = ""

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return err
	}

	fmt.Println("Default command updated successfully")
	return nil
}

func runConfigEdit() error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if len(cfg.Rules) == 0 {
		return fmt.Errorf("no rules available to edit")
	}

	// Build rule options for selection
	type ruleOption struct {
		Label string
		Index int
	}

	var ruleOptions []huh.Option[ruleOption]
	for i, rule := range cfg.Rules {
		label := fmt.Sprintf("%d. %s", i+1, buildRuleLabel(&rule))
		ruleOptions = append(ruleOptions, huh.NewOption(label, ruleOption{Label: label, Index: i}))
	}

	var selected ruleOption
	selectForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[ruleOption]().
				Title("Select a rule to edit").
				Options(ruleOptions...).
				Value(&selected),
		),
	)

	if err := selectForm.Run(); err != nil {
		return err
	}

	// Get the selected rule
	rule := &cfg.Rules[selected.Index]

	// Create edit form
	var (
		name       = rule.Name
		extensions = joinStrings(rule.Extensions)
		regex      = rule.Regex
		mime       = rule.Mime
		scheme     = rule.Scheme
		command    = rule.Command
		background = rule.Background
		terminal   = rule.Terminal
	)

	editForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Description("Display name for this rule").
				Value(&name),

			huh.NewInput().
				Title("Extensions").
				Description("Comma-separated list of file extensions (e.g., pdf,doc)").
				Value(&extensions),

			huh.NewInput().
				Title("Regex").
				Description("Regular expression pattern to match filenames").
				Value(&regex),

			huh.NewInput().
				Title("MIME Type").
				Description("MIME type pattern (regex) to match").
				Value(&mime),

			huh.NewInput().
				Title("Scheme").
				Description("URL scheme to match (e.g., https, ftp)").
				Value(&scheme),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Command").
				Description("Command to execute (use {{.File}}, {{.Dir}}, etc.)").
				Value(&command).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("command is required")
					}
					return nil
				}),

			huh.NewConfirm().
				Title("Background").
				Description("Run command in background?").
				Value(&background),

			huh.NewConfirm().
				Title("Terminal").
				Description("Require terminal for execution?").
				Value(&terminal),
		),
	)

	if err := editForm.Run(); err != nil {
		return err
	}

	// Validate regex and mime patterns if provided
	if regex != "" {
		if err := config.ValidateRegex(regex); err != nil {
			return fmt.Errorf("invalid regex: %w", err)
		}
	}
	if mime != "" {
		if err := config.ValidateRegex(mime); err != nil {
			return fmt.Errorf("invalid MIME pattern: %w", err)
		}
	}

	// Update the rule
	rule.Name = name
	rule.Extensions = splitAndTrim(extensions)
	rule.Regex = regex
	rule.Mime = mime
	rule.Scheme = scheme
	rule.Command = command
	rule.Background = background
	rule.Terminal = terminal

	// Save config
	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return err
	}

	fmt.Println("Rule updated successfully")
	return nil
}

// buildRuleLabel creates a display label for a rule
func buildRuleLabel(rule *config.Rule) string {
	if rule.Name != "" {
		return rule.Name
	}
	if len(rule.Extensions) > 0 {
		return fmt.Sprintf("Extensions: %s", joinStrings(rule.Extensions))
	}
	if rule.Regex != "" {
		return fmt.Sprintf("Regex: %s", rule.Regex)
	}
	return fmt.Sprintf("Command: %s", rule.Command)
}

// joinStrings joins a slice of strings with commas
func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	return fmt.Sprintf("%s", strs[0]) + func() string {
		if len(strs) == 1 {
			return ""
		}
		result := ""
		for _, s := range strs[1:] {
			result += "," + s
		}
		return result
	}()
}

// splitAndTrim splits a comma-separated string and trims whitespace
func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := []string{}
	for _, part := range splitString(s, ",") {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	current := ""
	for _, c := range s {
		if string(c) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" || len(result) > 0 {
		result = append(result, current)
	}
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
