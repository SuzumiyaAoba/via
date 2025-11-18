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
default_command: vim
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
})
