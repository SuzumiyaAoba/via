package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEntry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Entry CLI Suite")
}

var _ = Describe("Entry CLI", func() {
	var (
		tmpDir     string
		configFile string
		outBuf     bytes.Buffer
		errBuf     bytes.Buffer
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		configFile = filepath.Join(tmpDir, "config.yml")
		outBuf.Reset()
		errBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&errBuf)
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
})
