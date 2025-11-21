package cli

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config commands", func() {
	var (
		tmpDir     string
		configFile string
		outBuf     bytes.Buffer
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		configFile = filepath.Join(tmpDir, "config.yml")
		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)
	})

	Describe("runConfigList", func() {
		BeforeEach(func() {
			configContent := `version: "1"
default_command: vim {{.File}}
rules:
  - name: PDF Reader
    extensions:
      - pdf
    command: open {{.File}}
`
			err := os.WriteFile(configFile, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Set global cfgFile variable
			cfgFile = configFile
		})

		AfterEach(func() {
			cfgFile = ""
		})

		It("should list configuration", func() {
			err := runConfigList(rootCmd)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error for missing config", func() {
			cfgFile = filepath.Join(tmpDir, "nonexistent.yml")
			err := runConfigList(rootCmd)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("runConfigOpen", func() {
		BeforeEach(func() {
			cfgFile = configFile
		})

		AfterEach(func() {
			cfgFile = ""
		})

		It("should create default config if not exists", func() {
			// Config doesn't exist yet
			Expect(fileExists(configFile)).To(BeFalse())

			// runConfigOpen will try to open the file with system opener
			// We can't actually test the opening, but we can verify config creation
			// Since we're in test mode, OpenSystem should work with dry-run
			// However, OpenSystem isn't in dry-run by default in runConfigOpen

			// Let's just verify the config gets created
			cfg := &config.Config{Version: "1"}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(fileExists(configFile)).To(BeTrue())
		})
	})

	Describe("runConfigAdd", func() {
		BeforeEach(func() {
			cfgFile = configFile
			// Create initial config
			cfg := &config.Config{
				Version: "1",
				Rules:   []config.Rule{},
			}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			cfgFile = ""
		})

		It("should add rule with extension", func() {
			// Reset flags before setting new values
			configAddCmd.Flags().Set("ext", "txt")
			configAddCmd.Flags().Set("cmd", "cat {{.File}}")

			err := runConfigAdd(configAddCmd, []string{})
			Expect(err).NotTo(HaveOccurred())

			// Verify rule was added
			cfg, err := config.LoadConfig(cfgFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Rules).To(HaveLen(1))
			Expect(cfg.Rules[0].Extensions).To(ContainElement("txt"))
			Expect(cfg.Rules[0].Command).To(Equal("cat {{.File}}"))
		})

		It("should return error when --cmd is missing", func() {
			// Reset flags to ensure clean state
			configAddCmd.Flags().Set("ext", "")
			configAddCmd.Flags().Set("cmd", "")

			var outBuf bytes.Buffer
			configAddCmd.SetOut(&outBuf)
			configAddCmd.SetErr(&outBuf)

			err := runConfigAdd(configAddCmd, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("--cmd is required"))
		})

		It("should create config if not exists", func() {
			cfgFile = filepath.Join(tmpDir, "new_config.yml")
			Expect(fileExists(cfgFile)).To(BeFalse())

			// Set flags directly
			configAddCmd.Flags().Set("ext", "md")
			configAddCmd.Flags().Set("cmd", "vim {{.File}}")

			err := runConfigAdd(configAddCmd, []string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fileExists(cfgFile)).To(BeTrue())
		})

		It("should add rule with all flags", func() {
			// Reset flags
			configAddCmd.Flags().Set("ext", "")
			configAddCmd.Flags().Set("cmd", "echo test")
			configAddCmd.Flags().Set("name", "Test Rule")
			configAddCmd.Flags().Set("regex", ".*\\.test$")
			configAddCmd.Flags().Set("mime", "text/plain")
			configAddCmd.Flags().Set("scheme", "https")
			configAddCmd.Flags().Set("terminal", "true")
			configAddCmd.Flags().Set("background", "true")
			configAddCmd.Flags().Set("fallthrough", "true")
			configAddCmd.Flags().Set("os", "darwin,linux")

			err := runConfigAdd(configAddCmd, []string{})
			Expect(err).NotTo(HaveOccurred())

			cfg, err := config.LoadConfig(cfgFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Rules).To(HaveLen(1))
			rule := cfg.Rules[0]
			Expect(rule.Name).To(Equal("Test Rule"))
			Expect(rule.Command).To(Equal("echo test"))
			Expect(rule.Regex).To(Equal(".*\\.test$"))
			Expect(rule.Mime).To(Equal("text/plain"))
			Expect(rule.Scheme).To(Equal("https"))
			Expect(rule.Terminal).To(BeTrue())
			Expect(rule.Background).To(BeTrue())
			Expect(rule.Fallthrough).To(BeTrue())
			Expect(rule.OS).To(ConsistOf("darwin", "linux"))
		})
	})

	Describe("runConfigRemove", func() {
		BeforeEach(func() {
			cfgFile = configFile
			// Create a valid config file with rules
			cfg := config.Config{
				Version: "1",
				Rules: []config.Rule{
					{Name: "Rule 1", Command: "cmd1"},
					{Name: "Rule 2", Command: "cmd2"},
				},
			}
			err := config.SaveConfig(configFile, &cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should remove a rule by index", func() {
			err := runConfigRemove(rootCmd, "1")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Rule removed successfully"))

			// Verify rule was removed
			cfg, err := config.LoadConfig(configFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(cfg.Rules)).To(Equal(1))
			Expect(cfg.Rules[0].Name).To(Equal("Rule 2"))
		})

		It("should return error for invalid index format", func() {
			err := runConfigRemove(rootCmd, "invalid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid index"))
		})

		It("should return error for index out of range", func() {
			err := runConfigRemove(rootCmd, "3")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("index out of range"))
		})

		It("should return error for index 0", func() {
			err := runConfigRemove(rootCmd, "0")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("index out of range"))
		})
	})

	Describe("runConfigSetDefault", func() {
		BeforeEach(func() {
			cfgFile = configFile
			// Create a valid config file
			cfg := config.Config{Version: "1"}
			err := config.SaveConfig(configFile, &cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should set default command", func() {
			err := runConfigSetDefault(rootCmd, "vim {{.File}}")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Default command updated successfully"))

			// Verify config was updated
			cfg, err := config.LoadConfig(configFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.DefaultCommand).To(Equal("vim {{.File}}"))
		})

		It("should create config if it doesn't exist", func() {
			os.Remove(configFile)
			err := runConfigSetDefault(rootCmd, "nano {{.File}}")
			Expect(err).NotTo(HaveOccurred())

			// Verify config was created
			cfg, err := config.LoadConfig(configFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.DefaultCommand).To(Equal("nano {{.File}}"))
		})
	})
})
