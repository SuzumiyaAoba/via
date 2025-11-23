package cli

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helpers", func() {
	var tmpDir string

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
	})

	Describe("isURL", func() {
		It("should return true for valid URLs", func() {
			Expect(isURL("https://example.com")).To(BeTrue())
			Expect(isURL("http://example.com")).To(BeTrue())
			Expect(isURL("ftp://example.com")).To(BeTrue())
			Expect(isURL("file:///path/to/file")).To(BeTrue())
		})

		It("should return false for non-URLs", func() {
			Expect(isURL("test.txt")).To(BeFalse())
			Expect(isURL("/path/to/file")).To(BeFalse())
			Expect(isURL("relative/path")).To(BeFalse())
		})
	})

	Describe("fileExists", func() {
		It("should return true for existing file", func() {
			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())
			Expect(fileExists(testFile)).To(BeTrue())
		})

		It("should return false for non-existing file", func() {
			testFile := filepath.Join(tmpDir, "nonexistent.txt")
			Expect(fileExists(testFile)).To(BeFalse())
		})
	})

	Describe("isFileOrURL", func() {
		It("should return true for existing file", func() {
			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())
			Expect(isFileOrURL(testFile)).To(BeTrue())
		})

		It("should return true for URL", func() {
			Expect(isFileOrURL("https://example.com")).To(BeTrue())
		})

		It("should return false for non-existing file and non-URL", func() {
			Expect(isFileOrURL("nonexistent.txt")).To(BeFalse())
		})
	})

	Describe("executeWithDefault", func() {
		var (
			cfg    *config.Config
			exec   *executor.Executor
			outBuf bytes.Buffer
		)

		BeforeEach(func() {
			cfg = &config.Config{}
			outBuf.Reset()
			exec = executor.NewExecutor(&outBuf, true) // dry-run mode
		})

		It("should use default command when configured", func() {
			cfg.DefaultCommand = "vim {{.File}}"
			err := executeWithDefault(cfg, exec, "test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("vim test.txt"))
		})

		It("should use system opener when no default command", func() {
			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = executeWithDefault(cfg, exec, testFile)
			Expect(err).NotTo(HaveOccurred())
			// System opener command varies by OS, just check it tried to open
			Expect(outBuf.String()).To(ContainSubstring(testFile))
		})
	})

	Describe("executeRule", func() {
		var (
			exec   *executor.Executor
			outBuf bytes.Buffer
		)

		BeforeEach(func() {
			outBuf.Reset()
			exec = executor.NewExecutor(&outBuf, true) // dry-run mode
		})

		It("should execute rule command", func() {
			rule := &config.Rule{
				Command: "cat {{.File}}",
			}
			_, err := executeRule(exec, rule, "test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("cat test.txt"))
		})

		It("should execute with background option", func() {
			rule := &config.Rule{
				Command:    "open {{.File}}",
				Background: true,
			}
			_, err := executeRule(exec, rule, "test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("(background)"))
		})
	})

	Describe("executeRules", func() {
		var (
			exec   *executor.Executor
			outBuf bytes.Buffer
		)

		BeforeEach(func() {
			outBuf.Reset()
			exec = executor.NewExecutor(&outBuf, true) // dry-run mode
		})

		It("should execute multiple rules", func() {
			rules := []*config.Rule{
				{Command: "echo first {{.File}}"},
				{Command: "echo second {{.File}}"},
			}
			err := executeRules(exec, rules, "test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("echo first test.txt"))
			Expect(outBuf.String()).To(ContainSubstring("echo second test.txt"))
		})

		It("should execute single rule", func() {
			rules := []*config.Rule{
				{Command: "cat {{.File}}"},
			}
			err := executeRules(exec, rules, "test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("cat test.txt"))
		})
	})

	Describe("matchRules", func() {
		var cfg *config.Config

		BeforeEach(func() {
			cfg = &config.Config{
				Rules: []config.Rule{
					{
						Extensions: []string{"txt"},
						Command:    "cat {{.File}}",
					},
					{
						Extensions: []string{"md"},
						Command:    "vim {{.File}}",
					},
				},
			}
		})

		It("should match by extension", func() {
			matches, err := matchRules(cfg, "test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(matches).To(HaveLen(1))
			Expect(matches[0].Command).To(Equal("cat {{.File}}"))
		})

		It("should return empty for no match", func() {
			matches, err := matchRules(cfg, "test.pdf")
			Expect(err).NotTo(HaveOccurred())
			Expect(matches).To(BeEmpty())
		})
	})

	Describe("buildOptionLabel", func() {
		It("should use name when available", func() {
			rule := &config.Rule{
				Name:    "My Editor",
				Command: "vim {{.File}}",
			}
			label := buildOptionLabel(rule)
			Expect(label).To(Equal("My Editor"))
		})

		It("should use command when name is empty", func() {
			rule := &config.Rule{
				Command: "cat {{.File}}",
			}
			label := buildOptionLabel(rule)
			Expect(label).To(Equal("Command: cat {{.File}}"))
		})
	})

	Describe("buildInteractiveOptions", func() {
		var cfg *config.Config

		BeforeEach(func() {
			cfg = &config.Config{
				Rules: []config.Rule{
					{
						Name:       "Editor",
						Extensions: []string{"txt"},
						Command:    "vim {{.File}}",
					},
					{
						Extensions: []string{"txt"},
						Command:    "cat {{.File}}",
					},
				},
			}
		})

		It("should build options from matching rules", func() {
			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			options, err := buildInteractiveOptions(cfg, testFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(options).To(HaveLen(3)) // 2 rules + system default
			Expect(options[0].Label).To(Equal("Editor"))
			Expect(options[1].Label).To(ContainSubstring("Command:"))
			Expect(options[2].IsSystem).To(BeTrue())
		})

		It("should include system default for URLs", func() {
			options, err := buildInteractiveOptions(cfg, "https://example.com")
			Expect(err).NotTo(HaveOccurred())
			// URL won't match .txt extension rules
			// But system default should be included
			found := false
			for _, opt := range options {
				if opt.IsSystem {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should not include system default for non-existing file", func() {
			options, err := buildInteractiveOptions(cfg, "nonexistent.xyz")
			Expect(err).NotTo(HaveOccurred())
			found := false
			for _, opt := range options {
				if opt.IsSystem {
					found = true
					break
				}
			}
			Expect(found).To(BeFalse())
		})
	})
})
