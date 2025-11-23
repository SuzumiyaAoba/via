package cli

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
)

func TestCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Suite")
}

func resetGlobals() {
	cfgFile = ""
	dryRun = false
	interactive = false
	explain = false
	verbose = false
	profile = ""

	// Reset flags on rootCmd
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := f.Value.Set(f.DefValue); err != nil {
			panic(err)
		}
		f.Changed = false
	})
}
