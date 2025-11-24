package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLogger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger Suite")
}

var _ = Describe("Logger", func() {
	var (
		tmpDir  string
		logFile string
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		logFile = filepath.Join(tmpDir, "test.log")
	})

	Describe("DefaultConfig", func() {
		It("should return default configuration", func() {
			cfg := DefaultConfig()
			Expect(cfg.Enabled).To(BeFalse())
			Expect(cfg.Level).To(Equal(INFO))
			Expect(cfg.OutputFile).To(BeEmpty())
			Expect(cfg.MaxSize).To(Equal(int64(10 * 1024 * 1024)))
		})
	})

	Describe("New", func() {
		It("should create a disabled logger", func() {
			cfg := Config{Enabled: false}
			logger, err := New(cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger).NotTo(BeNil())

			// Should not panic when logging
			logger.Debug("test")
			logger.Info("test")
			logger.Warn("test")
			logger.Error("test")
		})

		It("should create an enabled logger without file output", func() {
			cfg := Config{
				Enabled: true,
				Level:   DEBUG,
			}
			logger, err := New(cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger).NotTo(BeNil())
			defer logger.Close()
		})

		It("should create an enabled logger with file output", func() {
			cfg := Config{
				Enabled:    true,
				Level:      DEBUG,
				OutputFile: logFile,
			}
			logger, err := New(cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger).NotTo(BeNil())
			defer logger.Close()

			logger.Info("test message")

			// Verify file was created and contains the message
			// Charm log output might be slightly different, check for substring
			data, err := os.ReadFile(logFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("test message"))
		})

		It("should create log directory if it doesn't exist", func() {
			nestedLogFile := filepath.Join(tmpDir, "nested", "dir", "test.log")
			cfg := Config{
				Enabled:    true,
				Level:      INFO,
				OutputFile: nestedLogFile,
			}
			logger, err := New(cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger).NotTo(BeNil())
			defer logger.Close()

			// Verify directory was created
			dir := filepath.Dir(nestedLogFile)
			info, err := os.Stat(dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
		})
	})

	Describe("Log levels", func() {
		var logger Logger

		BeforeEach(func() {
			cfg := Config{
				Enabled:    true,
				Level:      INFO,
				OutputFile: logFile,
			}
			var err error
			logger, err = New(cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if logger != nil {
				logger.Close()
			}
		})

		It("should only log messages at or above the configured level", func() {
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			logger.Error("error message")

			// Allow some time for async writing if any (though standard logger is sync)
			time.Sleep(10 * time.Millisecond)

			data, err := os.ReadFile(logFile)
			Expect(err).NotTo(HaveOccurred())
			content := string(data)

			// DEBUG should not be logged (below INFO level)
			Expect(content).NotTo(ContainSubstring("debug message"))
			// INFO, WARN, ERROR should be logged
			Expect(content).To(ContainSubstring("info message"))
			Expect(content).To(ContainSubstring("warn message"))
			Expect(content).To(ContainSubstring("error message"))
		})
	})

	Describe("Log rotation", func() {
		It("should initialize with rotation config", func() {
			cfg := Config{
				Enabled:    true,
				Level:      INFO,
				OutputFile: logFile,
				MaxSize:    1, // 1MB
			}
			logger, err := New(cfg)
			Expect(err).NotTo(HaveOccurred())
			defer logger.Close()

			logger.Info("test message")
			
			// Just verify file exists, trusting lumberjack for actual rotation
			_, err = os.Stat(logFile)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("NopLogger", func() {
		It("should do nothing", func() {
			logger := NewNopLogger()
			Expect(logger).NotTo(BeNil())

			// Should not panic
			logger.Debug("test")
			logger.Info("test")
			logger.Warn("test")
			logger.Error("test")
			err := logger.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Global logger", func() {
		It("should set and get global logger", func() {
			cfg := Config{
				Enabled: true,
				Level:   DEBUG,
			}
			logger, err := New(cfg)
			Expect(err).NotTo(HaveOccurred())
			defer logger.Close()

			SetGlobal(logger)
			global := GetGlobal()
			Expect(global).To(Equal(logger))
		})

		It("should use global logger for convenience functions", func() {
			cfg := Config{
				Enabled:    true,
				Level:      DEBUG,
				OutputFile: logFile,
			}
			logger, err := New(cfg)
			Expect(err).NotTo(HaveOccurred())
			defer logger.Close()

			SetGlobal(logger)

			Debug("debug")
			Info("info")
			Warn("warn")
			Error("error")

			data, err := os.ReadFile(logFile)
			Expect(err).NotTo(HaveOccurred())
			content := string(data)

			Expect(content).To(ContainSubstring("debug"))
			Expect(content).To(ContainSubstring("info"))
			Expect(content).To(ContainSubstring("warn"))
			Expect(content).To(ContainSubstring("error"))
		})
	})

	Describe("GetDefaultLogPath", func() {
		It("should return a valid path", func() {
			path, err := GetDefaultLogPath()
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(ContainSubstring(".config/entry/logs/entry.log"))
		})

		It("should return error if home dir fails", func() {
			origUserHomeDir := UserHomeDir
			UserHomeDir = func() (string, error) {
				return "", fmt.Errorf("mock error")
			}
			defer func() { UserHomeDir = origUserHomeDir }()

			_, err := GetDefaultLogPath()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Level.String", func() {
		It("should return correct string representations", func() {
			Expect(DEBUG.String()).To(Equal("debug"))
			Expect(INFO.String()).To(Equal("info"))
			Expect(WARN.String()).To(Equal("warn"))
			Expect(ERROR.String()).To(Equal("error"))
		})
	})
})
