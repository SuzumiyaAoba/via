package executor_test

import (
	"testing"

	. "github.com/SuzumiyaAoba/entry/internal/executor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestExecutor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Executor Suite")
}

var _ = Describe("Execute", func() {
	DescribeTable("executing commands",
		func(commandTmpl string, file string, wantErr bool) {
			err := Execute(commandTmpl, file, false)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("Simple echo", "echo {{.File}}", "test.txt", false),
		Entry("Invalid template", "echo {{.File", "test.txt", true),
		Entry("Command failure", "false", "test.txt", true),
	)

	It("should print command in dry run mode", func() {
		// We can't easily capture stdout here without redirecting it,
		// but we can check that it doesn't error and doesn't run the command (if we could verify that).
		// For now, just check no error.
		err := Execute("echo {{.File}}", "test.txt", true)
		Expect(err).NotTo(HaveOccurred())
	})
})
