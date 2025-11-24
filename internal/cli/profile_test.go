package cli

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)



var _ = Describe("Profile commands", func() {
	var (
		tmpDir     string
		configFile string
		outBuf     bytes.Buffer
		origHome   string
	)

	BeforeEach(func() {
		resetGlobals()
		tmpDir = GinkgoT().TempDir()
		origHome = os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)

		// Create config dir
		configDir := filepath.Join(tmpDir, ".config", "entry")
		err := os.MkdirAll(configDir, 0755)
		Expect(err).NotTo(HaveOccurred())

		configFile = filepath.Join(configDir, "config.yml")
		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)

		// Mock config path
		cfgFile = configFile
		
		// Create default config
		cfg := &config.Config{Version: "1"}
		err = config.SaveConfig(configFile, cfg)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.Setenv("HOME", origHome)
		cfgFile = ""
	})

	Describe("runConfigProfileList", func() {
		It("should list profiles", func() {
			// Create a profile
			profilesDir := filepath.Join(filepath.Dir(configFile), "profiles")
			err := os.MkdirAll(profilesDir, 0755)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(filepath.Join(profilesDir, "work.yml"), []byte{}, 0644)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "profile-list"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("work"))
		})
	})

	Describe("runConfigProfileCopy", func() {
		It("should copy profile", func() {
			err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "profile-copy", "default", "newprofile"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Profile 'default' copied to 'newprofile'"))

			profilesDir := filepath.Join(filepath.Dir(configFile), "profiles")
			Expect(fileExists(filepath.Join(profilesDir, "newprofile.yml"))).To(BeTrue())
		})

		It("should copy from named profile", func() {
			// Create a source profile
			profilesDir := filepath.Join(filepath.Dir(configFile), "profiles")
			err := os.MkdirAll(profilesDir, 0755)
			Expect(err).NotTo(HaveOccurred())
			
			srcProfile := filepath.Join(profilesDir, "source.yml")
			err = os.WriteFile(srcProfile, []byte("version: '1'"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "profile-copy", "source", "dest"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Profile 'source' copied to 'dest'"))
			Expect(fileExists(filepath.Join(profilesDir, "dest.yml"))).To(BeTrue())
		})

		It("should fail if source profile does not exist", func() {
			err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "profile-copy", "nonexistent", "dest"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("source profile 'nonexistent' does not exist"))
		})

		It("should fail if target profile already exists", func() {
			// Create target profile
			profilesDir := filepath.Join(filepath.Dir(configFile), "profiles")
			err := os.MkdirAll(profilesDir, 0755)
			Expect(err).NotTo(HaveOccurred())
			
			targetProfile := filepath.Join(profilesDir, "existing.yml")
			err = os.WriteFile(targetProfile, []byte{}, 0644)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "profile-copy", "default", "existing"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("target profile 'existing' already exists"))
		})
	})
})
