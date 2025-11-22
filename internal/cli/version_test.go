package cli

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version command", func() {
	var outBuf bytes.Buffer

	BeforeEach(func() {
		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
	})

	It("should print version", func() {
		rootCmd.SetArgs([]string{":version"})
		err := rootCmd.Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("et dev"))
	})
})
