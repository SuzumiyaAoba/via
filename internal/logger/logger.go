package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Level represents the log level
type Level = log.Level

const (
	// DEBUG level for detailed debugging information
	DEBUG = log.DebugLevel
	// INFO level for general informational messages
	INFO = log.InfoLevel
	// WARN level for warning messages
	WARN = log.WarnLevel
	// ERROR level for error messages
	ERROR = log.ErrorLevel
)

// Logger is the interface for logging
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Close() error
}

// Config holds the logger configuration
type Config struct {
	// Enabled determines if logging is enabled
	Enabled bool
	// Level is the minimum log level to output
	Level Level
	// OutputFile is the path to the log file (empty for stderr only)
	OutputFile string
	// MaxSize is the maximum size in bytes before rotation (0 for no rotation)
	MaxSize int64
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() Config {
	return Config{
		Enabled:    false,
		Level:      INFO,
		OutputFile: "",
		MaxSize:    10 * 1024 * 1024, // 10MB
	}
}

// entryLogger is the concrete implementation of Logger
type entryLogger struct {
	logger *log.Logger
	closer io.Closer
}

// New creates a new logger with the given configuration
func New(config Config) (Logger, error) {
	if !config.Enabled {
		return NewNopLogger(), nil
	}

	var output io.Writer = os.Stderr
	var closer io.Closer

	if config.OutputFile != "" {
		// Ensure directory exists
		dir := filepath.Dir(config.OutputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Use lumberjack for log rotation
		lumber := &lumberjack.Logger{
			Filename:   config.OutputFile,
			MaxSize:    int(config.MaxSize / 1024 / 1024), // MB
			MaxBackups: 3,
			MaxAge:     28,   // days
			Compress:   true, // disabled by default
		}
		output = io.MultiWriter(os.Stderr, lumber)
		closer = lumber
	}

	l := log.New(output)
	l.SetLevel(config.Level)
	l.SetReportTimestamp(true)
	l.SetReportCaller(false)

	return &entryLogger{
		logger: l,
		closer: closer,
	}, nil
}

func (l *entryLogger) Debug(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *entryLogger) Info(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *entryLogger) Warn(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *entryLogger) Error(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *entryLogger) Close() error {
	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}

// NopLogger is a logger that does nothing
type nopLogger struct{}

// NewNopLogger returns a logger that does nothing
func NewNopLogger() Logger {
	return &nopLogger{}
}

func (n *nopLogger) Debug(format string, args ...interface{}) {}
func (n *nopLogger) Info(format string, args ...interface{})  {}
func (n *nopLogger) Warn(format string, args ...interface{})  {}
func (n *nopLogger) Error(format string, args ...interface{}) {}
func (n *nopLogger) Close() error                              { return nil }

// Global logger instance
var (
	global Logger = NewNopLogger()
	mu     sync.RWMutex
)

// SetGlobal sets the global logger instance
func SetGlobal(l Logger) {
	mu.Lock()
	defer mu.Unlock()
	global = l
}

// GetGlobal returns the global logger instance
func GetGlobal() Logger {
	mu.RLock()
	defer mu.RUnlock()
	return global
}

// Convenience functions for global logger
func Debug(format string, args ...interface{}) {
	GetGlobal().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetGlobal().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetGlobal().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetGlobal().Error(format, args...)
}

// UserHomeDir is a variable to allow mocking in tests
var UserHomeDir = os.UserHomeDir

// GetDefaultLogPath returns the default log file path
func GetDefaultLogPath() (string, error) {
	home, err := UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}
	return filepath.Join(home, ".config", "entry", "logs", "entry.log"), nil
}
