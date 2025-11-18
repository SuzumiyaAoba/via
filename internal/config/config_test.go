package config_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/SuzumiyaAoba/entry/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("LoadConfig", func() {
	var (
		tmpDir     string
		configPath string
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		configPath = filepath.Join(tmpDir, "config.yml")
	})

	Context("when config file exists", func() {
		BeforeEach(func() {
			configContent := `
version: "1"
aliases:
  grep: rg
rules:
  - extensions: [txt]
    command: "echo text"
  - regex: ".*\\.log$"
    command: "echo log"
`
			err := os.WriteFile(configPath, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should load config successfully", func() {
			cfg, err := LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
			Expect(cfg.Version).To(Equal("1"))
			Expect(cfg.Rules).To(HaveLen(2))
			Expect(cfg.Aliases).To(HaveKeyWithValue("grep", "rg"))

			foundRegex := false
			for _, rule := range cfg.Rules {
				if rule.Regex == ".*\\.log$" {
					foundRegex = true
					break
				}
			}
			Expect(foundRegex).To(BeTrue(), "LoadConfig() did not load regex rule")
		})
	})

	Context("when config file does not exist", func() {
		It("should return an error", func() {
			_, err := LoadConfig(filepath.Join(tmpDir, "nonexistent.yml"))
			Expect(err).To(HaveOccurred())
		})
	})
})
