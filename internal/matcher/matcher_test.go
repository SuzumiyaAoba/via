package matcher_test

import (
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
				Extensions: []string{"go"},
				Command:    "echo go",
			},
		}
	})

	DescribeTable("matching files against rules",
		func(filename string, wantCmd string, wantErr bool) {
			got, err := Match(rules, filename)
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
		Entry("Match extension go", "main.go", "echo go", false),
		Entry("No match", "image.png", "", false),
		Entry("Case insensitive extension", "FILE.TXT", "echo text", false),
	)
})
