package cli

import (
	"bytes"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/tui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dashboard command", func() {
	var (
		tmpDir  string
		cfgFile string
		outBuf  bytes.Buffer
	)

	BeforeEach(func() {
		resetGlobals()
		tmpDir = GinkgoT().TempDir()
		cfgFile = filepath.Join(tmpDir, "config.yml")
		
		cfg := &config.Config{Version: "1"}
		err := config.SaveConfig(cfgFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)
	})

	It("should initialize TUI model", func() {
		cfg := &config.Config{
			Version: "1",
			Rules: []config.Rule{
				{Name: "Rule 1", Command: "cmd1"},
			},
		}

		model, err := tui.NewModel(cfg, cfgFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(model).NotTo(BeNil())
		Expect(model.Cfg).To(Equal(cfg))
	})

	It("should verify dashboard command exists", func() {
		cmd, _, err := rootCmd.Find([]string{":dashboard"})
		Expect(err).NotTo(HaveOccurred())
		Expect(cmd.Name()).To(Equal(":dashboard"))
		Expect(cmd.RunE).NotTo(BeNil())
	})
})
