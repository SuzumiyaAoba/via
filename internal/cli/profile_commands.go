package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/spf13/cobra"
)

var configProfileListCmd = &cobra.Command{
	Use:   "profile-list",
	Short: "List available configuration profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigProfileList()
	},
}

var configProfileCopyCmd = &cobra.Command{
	Use:   "profile-copy <from> <to>",
	Short: "Copy a configuration profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigProfileCopy(args[0], args[1])
	},
}

func runConfigProfileList() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home dir: %w", err)
	}

	profilesDir := filepath.Join(home, ".config", "entry", "profiles")
	
	// Check if profiles directory exists
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		fmt.Println("No profiles available")
		return nil
	}

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return fmt.Errorf("failed to read profiles directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No profiles available")
		return nil
	}

	fmt.Println("Available profiles:")
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yml") {
			profileName := strings.TrimSuffix(entry.Name(), ".yml")
			fmt.Printf("  - %s\n", profileName)
		}
	}

	return nil
}

func runConfigProfileCopy(from, to string) error {
	var sourcePath string
	var err error

	// Determine source path
	if from == "default" {
		sourcePath, err = config.GetConfigPath("")
		if err != nil {
			return err
		}
	} else {
		sourcePath, err = config.GetConfigPathWithProfile("", from)
		if err != nil {
			return err
		}
	}

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source profile '%s' does not exist", from)
	}

	// Load source config
	var cfg *config.Config
	if from == "default" {
		cfg, err = config.LoadConfig("")
	} else {
		cfg, err = config.LoadConfig(sourcePath)
	}
	if err != nil {
		return fmt.Errorf("failed to load source profile: %w", err)
	}

	// Determine target path
	targetPath, err := config.GetConfigPathWithProfile("", to)
	if err != nil {
		return err
	}

	// Check if target already exists
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("target profile '%s' already exists", to)
	}

	// Save to target path
	if err := config.SaveConfig(targetPath, cfg); err != nil {
		return fmt.Errorf("failed to copy profile: %w", err)
	}

	fmt.Printf("Profile '%s' copied to '%s'\n", from, to)
	return nil
}
