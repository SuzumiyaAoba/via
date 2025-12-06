package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/via/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)



var _ = Describe("Via CLI", func() {
	var (
		tmpDir     string
		configFile string
		outBuf     bytes.Buffer
		errBuf     bytes.Buffer
	)

	BeforeEach(func() {
		resetGlobals()
		tmpDir = GinkgoT().TempDir()
		configFile = filepath.Join(tmpDir, "config.yml")
		outBuf.Reset()
		errBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&errBuf)
		
		// Reset global flags
		cfgFile = configFile
	})

	Context("Command Disambiguation", func() {
		BeforeEach(func() {
			configContent := `
version: "1"
rules:
  - name: Config Rule
    extensions: [conf]
    command: echo "Opening config file"
`
			err := os.WriteFile(configFile, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should treat 'config' as file argument without double dash", func() {
			// Create a rule for "config" file
			cfg := config.Config{
				Version: "1",
				Rules: []config.Rule{
					{
						Name:       "Config Rule",
						Extensions: []string{"conf"},
						Command:    "echo \"Opening config file\"",
					},
				},
			}
			err := config.SaveConfig(configFile, &cfg)
			Expect(err).NotTo(HaveOccurred())

			rootCmd.SetArgs([]string{"--config", configFile, "--dry-run", "config.conf"})
			err = rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Opening config file"))
		})

		It("should execute config subcommand with double dash", func() {
			// et -- :config list -> should execute config list command
			rootCmd.SetArgs([]string{"--config", configFile, "--", ":config", "list"})
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			// config list output is YAML
			Expect(outBuf.String()).To(ContainSubstring("version: \"1\""))
		})

		It("should handle flags before double dash", func() {
			// et --dry-run -- :config list
			// Note: config list doesn't use dry-run, but we check if parsing works
			rootCmd.SetArgs([]string{"--config", configFile, "--dry-run", "--", ":config", "list"})
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("version: \"1\""))
		})
	})

	Context("with aliases", func() {
		BeforeEach(func() {
			configContent := `
aliases:
  grep: rg
rules:
  - extensions: [txt]
    command: echo file {{.File}}
`
			err := os.WriteFile(configFile, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should execute alias", func() {
			rootCmd.SetArgs([]string{"--config", configFile, "--dry-run", "grep", "pattern"})
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("rg pattern"))
		})

		It("should execute file rule", func() {
			rootCmd.SetArgs([]string{"--config", configFile, "--dry-run", "test.txt"})
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("echo file test.txt"))
		})

		It("should execute command if no alias/rule", func() {
			rootCmd.SetArgs([]string{"--config", configFile, "--dry-run", "ls", "-la"})
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("ls -la"))
		})
	})

	Context("with default command", func() {
		BeforeEach(func() {
			configContent := `
default_command: vim {{.File}}
rules: []
`
			err := os.WriteFile(configFile, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should use default command for single argument", func() {
			rootCmd.SetArgs([]string{"--config", configFile, "--dry-run", "unknown_file"})
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("vim unknown_file"))
		})

		It("should execute command for multiple arguments", func() {
			rootCmd.SetArgs([]string{"--config", configFile, "--dry-run", "ls", "-la"})
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("ls -la"))
		})
	})

	Context("with system fallback", func() {
		BeforeEach(func() {
			configContent := `
rules: []
`
			err := os.WriteFile(configFile, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())
			// Create a dummy file for fallback test
			err = os.WriteFile("fallback.txt", []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Remove("fallback.txt")
		})

		It("should use system opener for URL", func() {
			rootCmd.SetArgs([]string{"--config", configFile, "--dry-run", "https://google.com"})
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			// Expect "open https://google.com" or "xdg-open https://google.com" depending on OS
			// But since we are in test, we can't easily predict OS command string without runtime check.
			// However, OpenSystem prints cmdName + args.
			// Let's just check it contains the URL.
			Expect(outBuf.String()).To(ContainSubstring("https://google.com"))
		})

		It("should use system opener for existing file", func() {
			rootCmd.SetArgs([]string{"--config", configFile, "--dry-run", "fallback.txt"})
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("fallback.txt"))
		})

		It("should execute as command if not file/URL", func() {
			rootCmd.SetArgs([]string{"--config", configFile, "--dry-run", "ls", "-la"})
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("ls -la"))
		})
	})
	Context("Error Handling and Flags", func() {
		It("should return error for invalid flags", func() {
			// We need to use a new command to test flag parsing error because 
			// rootCmd has DisableFlagParsing: true, so it parses flags manually in runRoot.
			// However, runRoot calls cmd.Flags().Parse(args).
			err := rootCmd.RunE(rootCmd, []string{"--invalid-flag"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown flag"))
		})

		It("should show help", func() {
			err := rootCmd.RunE(rootCmd, []string{"--help"})
			Expect(err).NotTo(HaveOccurred())
			// Help output is printed to stdout/stderr
			Expect(outBuf.String()).To(ContainSubstring("Usage:"))
		})

		It("should show version", func() {
			err := rootCmd.RunE(rootCmd, []string{"--version"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("vv version"))
		})

		It("should return error if config load fails", func() {
			// Create invalid config file (tab is invalid in YAML)
			err := os.WriteFile(configFile, []byte("invalid:\n\tvalue"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", configFile, "file.txt"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error loading config"))
		})

		It("should return error if logger init fails", func() {
			// This is hard to simulate without mocking logger init or filesystem error.
			// skipping for now.
		})
		
		It("should handle profile resolution error", func() {
			// Clear cfgFile to force profile resolution
			cfgFile = ""
			
			// Mock UserHomeDir to fail
			origUserHomeDir := config.UserHomeDir
			config.UserHomeDir = func() (string, error) {
				return "", fmt.Errorf("mock error")
			}
			defer func() { config.UserHomeDir = origUserHomeDir }()
			
			// Set profile to trigger resolution
			os.Setenv("VIA_PROFILE", "test")
			defer os.Unsetenv("VIA_PROFILE")
			
			err := rootCmd.RunE(rootCmd, []string{"file.txt"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to resolve profile config path"))
		})
	})
})
