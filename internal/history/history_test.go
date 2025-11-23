package history_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SuzumiyaAoba/entry/internal/history"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHistory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "History Suite")
}

var _ = Describe("History", func() {
	var (
		tmpDir      string
		historyFile string
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		historyFile = filepath.Join(tmpDir, "history.json")
		history.SetHistoryPath(historyFile)
	})

	AfterEach(func() {
		history.SetHistoryPath("")
	})

	It("should add and load entries", func() {
		err := history.AddEntry("cmd1", "rule1")
		Expect(err).NotTo(HaveOccurred())

		entries, err := history.LoadHistory()
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		Expect(entries[0].Command).To(Equal("cmd1"))
		Expect(entries[0].RuleName).To(Equal("rule1"))
	})

	It("should clear history", func() {
		err := history.AddEntry("cmd1", "rule1")
		Expect(err).NotTo(HaveOccurred())

		err = history.ClearHistory()
		Expect(err).NotTo(HaveOccurred())

		entries, err := history.LoadHistory()
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(BeEmpty())
	})

	It("should handle max history size", func() {
		for i := 0; i < history.MaxHistorySize+10; i++ {
			err := history.AddEntry("cmd", "rule")
			Expect(err).NotTo(HaveOccurred())
		}

		entries, err := history.LoadHistory()
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(history.MaxHistorySize))
	})

	It("should create directory if not exists", func() {
		// historyFile is already in a temp dir, but let's make a nested one
		nestedFile := filepath.Join(tmpDir, "nested", "history.json")
		history.SetHistoryPath(nestedFile)

		err := history.AddEntry("cmd", "rule")
		Expect(err).NotTo(HaveOccurred())

		_, err = os.Stat(nestedFile)
		Expect(err).NotTo(HaveOccurred())
	})
})
