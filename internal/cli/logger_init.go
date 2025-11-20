package cli

import (
	"os"

	"github.com/SuzumiyaAoba/entry/internal/logger"
)

// initLogger initializes the global logger based on flags and environment variables
func initLogger() error {
	// Check if verbose mode is enabled via environment variable
	if os.Getenv("ENTRY_VERBOSE") == "true" {
		verbose = true
	}

	// If not verbose, use nop logger
	if !verbose {
		logger.SetGlobal(logger.NewNopLogger())
		return nil
	}

	// Get default log path
	logPath, err := logger.GetDefaultLogPath()
	if err != nil {
		return err
	}

	// Create logger configuration
	cfg := logger.Config{
		Enabled:    true,
		Level:      logger.DEBUG,
		OutputFile: logPath,
		MaxSize:    10 * 1024 * 1024, // 10MB
	}

	// Create and set global logger
	l, err := logger.New(cfg)
	if err != nil {
		return err
	}

	logger.SetGlobal(l)
	return nil
}
