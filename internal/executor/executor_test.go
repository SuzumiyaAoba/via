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
			exec := NewExecutor(GinkgoWriter, false)
			err := exec.Execute(commandTmpl, file, ExecutionOptions{})
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("Simple echo", "echo {{.File}}", "test.txt", false),
		Entry("Extended variables", "echo {{.Dir}} {{.Base}} {{.Name}} {{.Ext}}", "test.txt", false),
		Entry("Invalid template", "echo {{.File", "test.txt", true),
		Entry("Command failure", "false", "test.txt", true),
	)

	It("should print command in dry run mode", func() {
		// We can't easily capture stdout here without redirecting it,
		// but we can check that it doesn't error and doesn't run the command (if we could verify that).
		// For now, just check no error.
		exec := NewExecutor(GinkgoWriter, true)
		err := exec.Execute("echo {{.File}}", "test.txt", ExecutionOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should handle background execution", func() {
		exec := NewExecutor(GinkgoWriter, false)
		// Use a command that exits immediately to avoid hanging
		err := exec.Execute("true", "test.txt", ExecutionOptions{Background: true})
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("ExecuteCommand", func() {
	It("should execute raw command", func() {
		exec := NewExecutor(GinkgoWriter, false)
		err := exec.ExecuteCommand("echo", []string{"hello"})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should fail on invalid command", func() {
		exec := NewExecutor(GinkgoWriter, false)
		err := exec.ExecuteCommand("invalid_command_xyz", []string{})
		Expect(err).To(HaveOccurred())
	})

	It("should print raw command in dry run mode", func() {
		exec := NewExecutor(GinkgoWriter, true)
		err := exec.ExecuteCommand("echo", []string{"hello"})
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("ExecuteScript", func() {
	var exec *Executor

	BeforeEach(func() {
		exec = NewExecutor(GinkgoWriter, false)
	})

	It("should return true for matching script", func() {
		cmd, matched, err := exec.ExecuteScript("true", "test.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(matched).To(BeTrue())
		Expect(cmd).To(BeEmpty())
	})

	It("should return false for non-matching script", func() {
		cmd, matched, err := exec.ExecuteScript("false", "test.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(matched).To(BeFalse())
		Expect(cmd).To(BeEmpty())
	})

	It("should return command string", func() {
		cmd, matched, err := exec.ExecuteScript("'echo matched'", "test.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(matched).To(BeTrue())
		Expect(cmd).To(Equal("echo matched"))
	})

	It("should access file variables", func() {
		script := `file == 'test.txt' && ext == '.txt'`
		_, matched, err := exec.ExecuteScript(script, "test.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(matched).To(BeTrue())
	})

	It("should return error on script failure", func() {
		_, _, err := exec.ExecuteScript("throw 'error'", "test.txt")
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("OpenSystem", func() {
	It("should print command in dry run mode", func() {
		exec := NewExecutor(GinkgoWriter, true)
		err := exec.OpenSystem("test.txt")
		Expect(err).NotTo(HaveOccurred())
	})
})
