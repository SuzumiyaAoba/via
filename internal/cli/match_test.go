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
		resetGlobals()
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

		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)
	})

	It("should match existing file", func() {
		// Create a matching file
		f := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(f, []byte("content"), 0644)
		Expect(err).NotTo(HaveOccurred())

		// Use RunE directly to avoid cobra Execute re-entrancy issues in tests
		err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":match", f})
		if err != nil {
			GinkgoWriter.Printf("Execute error: %v\nOutput: %s\n", err, outBuf.String())
		}
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("Text Rule"))
	})

	It("should return error if no match", func() {
		// Create a non-matching file
		f := filepath.Join(tmpDir, "test.bin")
		err := os.WriteFile(f, []byte("content"), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":match", f})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no match found"))
	})
})


