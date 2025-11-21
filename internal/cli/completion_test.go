package cli

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func TestCompletion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Completion Suite")
}

var _ = Describe("Completion", func() {
	var (
		tmpHome string
		origHome string
	)

	BeforeEach(func() {
		var err error
		tmpHome, err = os.MkdirTemp("", "entry-test-home")
		Expect(err).NotTo(HaveOccurred())
		
		origHome = os.Getenv("HOME")
		os.Setenv("HOME", tmpHome)
	})

	AfterEach(func() {
		os.Setenv("HOME", origHome)
		os.RemoveAll(tmpHome)
	})

	Describe("CompletionProfiles", func() {
		It("should return empty list when profiles directory does not exist", func() {
			profiles, directive := CompletionProfiles(nil, nil, "")
			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoFileComp))
			Expect(profiles).To(BeEmpty())
		})

		It("should return empty list when profiles directory is empty", func() {
			err := os.MkdirAll(filepath.Join(tmpHome, ".config", "entry", "profiles"), 0755)
			Expect(err).NotTo(HaveOccurred())

			profiles, directive := CompletionProfiles(nil, nil, "")
			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoFileComp))
			Expect(profiles).To(BeEmpty())
		})

		It("should list available profiles", func() {
			profilesDir := filepath.Join(tmpHome, ".config", "entry", "profiles")
			err := os.MkdirAll(profilesDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			// Create some dummy profiles
			err = os.WriteFile(filepath.Join(profilesDir, "work.yml"), []byte{}, 0644)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(filepath.Join(profilesDir, "home.yml"), []byte{}, 0644)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(filepath.Join(profilesDir, "readme.txt"), []byte{}, 0644) // Should be ignored
			Expect(err).NotTo(HaveOccurred())

			profiles, directive := CompletionProfiles(nil, nil, "")
			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoFileComp))
			Expect(profiles).To(ConsistOf("work", "home"))
		})

		It("should filter profiles by prefix", func() {
			profilesDir := filepath.Join(tmpHome, ".config", "entry", "profiles")
			err := os.MkdirAll(profilesDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(profilesDir, "dev.yml"), []byte{}, 0644)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(filepath.Join(profilesDir, "demo.yml"), []byte{}, 0644)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(filepath.Join(profilesDir, "prod.yml"), []byte{}, 0644)
			Expect(err).NotTo(HaveOccurred())

			profiles, directive := CompletionProfiles(nil, nil, "de")
			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoFileComp))
			Expect(profiles).To(ConsistOf("dev", "demo"))
		})
	})
})
