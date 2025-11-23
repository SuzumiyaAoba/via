package cli

import (
	"bytes"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/history"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)



var _ = Describe("History command", func() {
	var (
		tmpDir      string
		historyFile string
		outBuf      bytes.Buffer
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		historyFile = filepath.Join(tmpDir, "history.json")
		history.SetHistoryPath(historyFile)
		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)
	})

	AfterEach(func() {
		history.SetHistoryPath("")
	})

	It("should show empty message if no history", func() {
		err := runHistory(historyCmd)
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("No history available"))
	})

	It("should clear history", func() {
		// Add some history first
		err := history.AddEntry("cmd1", "rule1")
		Expect(err).NotTo(HaveOccurred())

		// Run clear command
		// We need to execute the clear subcommand logic.
		// Since historyClearCmd is a subcommand, we can run its RunE
		err = historyClearCmd.RunE(historyClearCmd, []string{})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("History cleared"))

		// Verify history is empty
		entries, err := history.LoadHistory()
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(BeEmpty())
	})
})
