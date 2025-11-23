package cli

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Match command", func() {
	var (
		tmpDir  string
		cfgFile string
		outBuf  bytes.Buffer
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		cfgFile = filepath.Join(tmpDir, "config.yml")
		
		// Create a dummy config
		cfg := &config.Config{
			Version: "1",
			Rules: []config.Rule{
				{
					Name:       "Text Rule",
					Extensions: []string{"txt"},
					Command:    "echo text",
				},
			},
		}
		err := config.SaveConfig(cfgFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		// Set global cfgFile variable in cli package (if accessible)
		// Since we are in same package 'cli', we can access 'cfgFile' var from root.go
		// But wait, root.go defines 'cfgFile' var.
		// We need to set it.
		// However, runMatch uses the global cfgFile variable.
		// We should set it here.
		// But wait, 'cfgFile' in root.go is 'var cfgFile string'.
		// We can set it.
		setCfgFile(cfgFile) // Helper to set the unexported var if needed, or just set it if exported?
		// It is unexported in root.go: var cfgFile string
		// But we are in package cli, so we can access it!
		
		outBuf.Reset()
		matchCmd.SetOut(&outBuf)
		matchCmd.SetErr(&outBuf)
	})

	It("should match existing file", func() {
		// Create a matching file
		f := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(f, []byte("content"), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = runMatch(matchCmd, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("Text Rule"))
	})

	It("should return error if no match", func() {
		// Create a non-matching file
		f := filepath.Join(tmpDir, "test.bin")
		err := os.WriteFile(f, []byte("content"), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = runMatch(matchCmd, f)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no match found"))
	})
})

// Helper to set cfgFile since we are in the same package
func setCfgFile(path string) {
	cfgFile = path
}
