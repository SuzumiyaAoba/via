package cli

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
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
})
