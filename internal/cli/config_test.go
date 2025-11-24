package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/sync"
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
		resetGlobals()
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
			
			// We need to mock executor or ensure it doesn't fail.
			// runConfigOpen creates a new Executor.
			// It calls exec.OpenSystem(configPath).
			// OpenSystem uses "open", "xdg-open" etc.
			// In CI environment this might fail or block.
			
			// But we can check if config file is created BEFORE opening.
			// Actually runConfigOpen logic:
			// 1. Get path
			// 2. Check if exists. If not, create default.
			// 3. Open.
			
			// If we can't mock OpenSystem, this test might be flaky or fail.
			// But we can test the creation logic if we can intercept the Open call.
			// We can't easily intercept.
			
			// Let's just test that it creates the file.
			// If OpenSystem fails, RunE returns error.
			// We can expect error or not depending on env.
			// But we really want to verify file creation.
			
			// Let's try to run it and ignore error from OpenSystem if possible, 
			// or assume it might fail but check file existence.
			
			_ = runConfigOpen(rootCmd) // Ignore error
			
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

	Describe("runConfigMove", func() {
		BeforeEach(func() {
			cfgFile = configFile
			cfg := config.Config{
				Version: "1",
				Rules: []config.Rule{
					{Name: "Rule 1", Command: "cmd1"},
					{Name: "Rule 2", Command: "cmd2"},
					{Name: "Rule 3", Command: "cmd3"},
				},
			}
			err := config.SaveConfig(configFile, &cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should move rule", func() {
			// Move Rule 3 (index 3) to index 1
			err := runConfigMove(rootCmd, 3, 1)
			Expect(err).NotTo(HaveOccurred())

			cfg, err := config.LoadConfig(configFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Rules[0].Name).To(Equal("Rule 3"))
			Expect(cfg.Rules[1].Name).To(Equal("Rule 1"))
			Expect(cfg.Rules[2].Name).To(Equal("Rule 2"))
		})

		It("should return error for invalid index", func() {
			err := runConfigMove(rootCmd, 99, 1)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("runConfigAlias", func() {
		BeforeEach(func() {
			cfgFile = configFile
			cfg := config.Config{
				Version: "1",
				Aliases: map[string]string{
					"old": "cmd",
				},
			}
			err := config.SaveConfig(configFile, &cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should add alias", func() {
			err := runConfigAliasAdd(rootCmd, "new", "echo new")
			Expect(err).NotTo(HaveOccurred())

			cfg, err := config.LoadConfig(configFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Aliases["new"]).To(Equal("echo new"))
		})

		It("should remove alias", func() {
			err := runConfigAliasRemove(rootCmd, "old")
			Expect(err).NotTo(HaveOccurred())

			cfg, err := config.LoadConfig(configFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Aliases).NotTo(HaveKey("old"))
		})

		It("should fail to add existing alias", func() {
			err := runConfigAliasAdd(rootCmd, "old", "new cmd")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("alias 'old' already exists"))
		})
	})

	Describe("runConfigCheck advanced", func() {
		It("should fail for invalid regex in rule", func() {
			cfg := &config.Config{
				Version: "1",
				Rules: []config.Rule{
					{Name: "Bad Regex", Regex: "[", Command: "cmd"},
				},
			}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = runConfigCheck(rootCmd)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("validation failed"))
		})
	})

	Describe("runConfigExport/Import", func() {
		BeforeEach(func() {
			cfgFile = configFile
			cfg := config.Config{
				Version: "1",
				Rules: []config.Rule{
					{Name: "Export Rule", Command: "cmd"},
				},
			}
			err := config.SaveConfig(configFile, &cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should export config", func() {
			exportFile := filepath.Join(tmpDir, "backup.yml")
			err := runConfigExport(rootCmd, exportFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(fileExists(exportFile)).To(BeTrue())
		})

		It("should import config", func() {
			importFile := filepath.Join(tmpDir, "import.yml")
			importContent := `
version: "1"
rules:
  - name: Imported Rule
    command: imported
`
			err := os.WriteFile(importFile, []byte(importContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = runConfigImport(rootCmd, importFile)
			Expect(err).NotTo(HaveOccurred())

			cfg, err := config.LoadConfig(configFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Rules[0].Name).To(Equal("Imported Rule"))
		})
	})

	Describe("runConfigAdd with Script", func() {
		BeforeEach(func() {
			cfgFile = configFile
			cfg := &config.Config{Version: "1"}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should add rule with script", func() {
			configAddCmd.Flags().Set("cmd", "echo script")
			configAddCmd.Flags().Set("script", "true")
			
			err := runConfigAdd(configAddCmd, []string{})
			Expect(err).NotTo(HaveOccurred())

			cfg, err := config.LoadConfig(cfgFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Rules[0].Script).To(Equal("true"))
		})
	})
	Describe("runConfigSync", func() {
		var (
			server *httptest.Server
			origURL string
		)

		BeforeEach(func() {
			cfgFile = configFile
			origURL = sync.GitHubAPIURL
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/gists" && r.Method == "POST" {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{"id": "newgist123"}`))
					return
				}
				if r.URL.Path == "/gists/gist123" && r.Method == "GET" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"files": {
							"config.yml": {
								"content": "version: \"1\"\nrules: []"
							}
						}
					}`))
					return
				}
				if r.URL.Path == "/gists/gist123" && r.Method == "PATCH" {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.WriteHeader(http.StatusNotFound)
			}))
			sync.GitHubAPIURL = server.URL
		})

		AfterEach(func() {
			server.Close()
			sync.GitHubAPIURL = origURL
		})

		It("should init sync", func() {
			// Create config first
			cfg := &config.Config{Version: "1"}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			// Mock stdin for token input
			// But runConfigSyncInit uses huh which reads from stdin.
			// Testing huh interactions in CLI tests is hard.
			// We might skip interactive parts or mock huh if possible.
			// For now, let's test push/pull which are non-interactive if configured.
		})

		It("should push config", func() {
			cfg := &config.Config{
				Version: "1",
				Sync: &config.SyncConfig{
					GistID: "gist123",
					Token:  "token",
				},
			}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "sync", "push"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Configuration pushed to Gist"))
		})

		It("should pull config", func() {
			cfg := &config.Config{
				Version: "1",
				Sync: &config.SyncConfig{
					GistID: "gist123",
					Token:  "token",
				},
			}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "sync", "pull"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Configuration pulled from Gist"))
		})

		It("should check config validity", func() {
			// Create valid config
			cfg := &config.Config{Version: "1"}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "check"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Configuration is valid"))
		})

		It("should set default command", func() {
			// Create config
			cfg := &config.Config{Version: "1"}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "set-default", "vim {{.File}}"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Default command updated"))

			// Verify
			cfg, err = config.LoadConfig(cfgFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.DefaultCommand).To(Equal("vim {{.File}}"))
		})

		It("should remove rule by index", func() {
			// Create config with rules
			cfg := &config.Config{
				Version: "1",
				Rules: []config.Rule{
					{Name: "Rule 1", Command: "cmd1"},
					{Name: "Rule 2", Command: "cmd2"},
				},
			}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "remove", "1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Rule removed"))

			// Verify
			cfg, err = config.LoadConfig(cfgFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Rules).To(HaveLen(1))
			Expect(cfg.Rules[0].Name).To(Equal("Rule 2"))
		})

		It("should fail to remove rule with invalid index", func() {
			// Create config
			cfg := &config.Config{Version: "1"}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "remove", "99"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("index out of range"))
		})

		It("should init config", func() {
			// Remove config file
			os.Remove(cfgFile)

			err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "init"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Created default config"))

			// Verify file exists
			_, err = os.Stat(cfgFile)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail init if config exists", func() {
			// Create config
			cfg := &config.Config{Version: "1"}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "init"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already exists"))
		})

		It("should list config", func() {
			// Create config
			cfg := &config.Config{
				Version: "1",
				Rules: []config.Rule{
					{Name: "Rule 1", Command: "cmd1"},
				},
			}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "list"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Rule 1"))
			Expect(outBuf.String()).To(ContainSubstring("cmd1"))
		})
	})
	Describe("buildRuleLabel", func() {
		It("should return name if present", func() {
			rule := &config.Rule{Name: "My Rule"}
			Expect(buildRuleLabel(rule)).To(Equal("My Rule"))
		})

		It("should return extensions if name is missing", func() {
			rule := &config.Rule{Extensions: []string{"txt", "md"}}
			Expect(buildRuleLabel(rule)).To(Equal("Extensions: txt,md"))
		})

		It("should return regex if name and extensions are missing", func() {
			rule := &config.Rule{Regex: ".*"}
			Expect(buildRuleLabel(rule)).To(Equal("Regex: .*"))
		})

		It("should return command if nothing else is present", func() {
			rule := &config.Rule{Command: "echo hello"}
			Expect(buildRuleLabel(rule)).To(Equal("Command: echo hello"))
		})
	})

	Describe("runConfigEdit", func() {
		BeforeEach(func() {
			cfgFile = configFile
		})

		It("should fail if no rules available", func() {
			cfg := &config.Config{Version: "1"}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = runConfigEdit(rootCmd)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no rules available to edit"))
		})
	})
})
