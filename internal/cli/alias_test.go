package cli

import (
	"bytes"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Alias command", func() {
	var (
		tmpDir  string
		cfgFile string
		outBuf  bytes.Buffer
	)

	BeforeEach(func() {
		resetGlobals()
		tmpDir = GinkgoT().TempDir()
		cfgFile = filepath.Join(tmpDir, "config.yml")
		
		cfg := &config.Config{Version: "1"}
		err := config.SaveConfig(cfgFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)
	})

	It("should add alias", func() {
		err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "alias", "add", "ll", "ls -la"})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("Alias 'll' added"))

		// Verify config
		cfg, err := config.LoadConfig(cfgFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Aliases["ll"]).To(Equal("ls -la"))
	})

	It("should remove alias", func() {
		// Add alias first
		cfg := &config.Config{
			Version: "1",
			Aliases: map[string]string{"ll": "ls -la"},
		}
		err := config.SaveConfig(cfgFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "alias", "remove", "ll"})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("Alias 'll' removed"))

		// Verify config
		cfg, err = config.LoadConfig(cfgFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Aliases).NotTo(HaveKey("ll"))
	})

	It("should fail if alias does not exist", func() {
		err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "alias", "remove", "nonexistent"})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("alias 'nonexistent' not found"))
	})

	It("should fail if aliases map is nil", func() {
		// Setup config with nil aliases
		cfg := &config.Config{Version: "1"}
		err := config.SaveConfig(cfgFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "alias", "remove", "foo"})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("alias 'foo' not found"))
	})

	It("should list aliases", func() {
		// Add aliases
		cfg := &config.Config{
			Version: "1",
			Aliases: map[string]string{
				"ll": "ls -la",
				"gs": "git status",
			},
		}
		err := config.SaveConfig(cfgFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "alias", "list"})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("Aliases:"))
		Expect(outBuf.String()).To(ContainSubstring("ll: ls -la"))
		Expect(outBuf.String()).To(ContainSubstring("gs: git status"))
	})

	It("should show message when no aliases exist", func() {
		err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "alias", "list"})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("No aliases defined"))
	})
})
