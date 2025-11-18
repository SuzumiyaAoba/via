package matcher_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/SuzumiyaAoba/entry/internal/config"
	. "github.com/SuzumiyaAoba/entry/internal/matcher"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Matcher Suite")
}

var _ = Describe("Match", func() {
	var rules []config.Rule

	BeforeEach(func() {
		// Create dummy files for testing
		_ = os.WriteFile("file.txt", []byte("text"), 0644)
		_ = os.WriteFile("README.md", []byte("markdown"), 0644)
		_ = os.WriteFile("app.log", []byte("log"), 0644)
		_ = os.WriteFile("main.go", []byte("go"), 0644)
		_ = os.WriteFile("script.sh", []byte("#!/bin/sh"), 0755)
		_ = os.WriteFile("script.bat", []byte("echo off"), 0644)
		_ = os.WriteFile("unknown.dat", []byte("data"), 0644)
		_ = os.WriteFile("FILE.TXT", []byte("TEXT"), 0644)
		// Create a fake PNG file signature
		// PNG signature: 89 50 4E 47 0D 0A 1A 0A
		_ = os.WriteFile("image.png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, 0644)

		rules = []config.Rule{
			{
				Extensions: []string{"txt", "md"},
				Command:    "echo text",
			},
			{
				Regex:   ".*\\.log$",
				Command: "echo log",
			},
			{
				Mime:    "image/.*",
				Command: "echo image",
			},
			{
				Extensions: []string{"go"},
				Command:    "echo go",
			},
			{
				Extensions: []string{"sh"},
				OS:         []string{runtime.GOOS},
				Command:    "echo sh",
			},
			{
				Extensions: []string{"bat"},
				OS:         []string{"otheros"},
				Command:    "echo bat",
			},
		}
	})

	AfterEach(func() {
		os.Remove("file.txt")
		os.Remove("README.md")
		os.Remove("app.log")
		os.Remove("main.go")
		os.Remove("script.sh")
		os.Remove("script.bat")
		os.Remove("unknown.dat")
		os.Remove("FILE.TXT")
		os.Remove("image.png")
	})

	DescribeTable("matching files against rules",
		func(filename string, wantCmd string, wantErr bool) {
			got, err := Match(rules, "echo default", filename)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}

			if wantCmd != "" {
				Expect(got).NotTo(BeNil())
				Expect(got.Command).To(Equal(wantCmd))
			} else {
				Expect(got).To(BeNil())
			}
		},
		Entry("Match extension txt", "file.txt", "echo text", false),
		Entry("Match extension md", "README.md", "echo text", false),
		Entry("Match regex log", "app.log", "echo log", false),
		Entry("Match mime image", "image.png", "echo image", false),
		Entry("Match extension go", "main.go", "echo go", false),
		Entry("Match OS specific", "script.sh", "echo sh", false),
		Entry("No match OS mismatch", "script.bat", "echo default", false),
		Entry("No match", "unknown.dat", "echo default", false),
		Entry("Case insensitive extension", "FILE.TXT", "echo text", false),
	)
})
