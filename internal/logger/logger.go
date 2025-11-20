package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents the log level
type Level int

const (
	// DEBUG level for detailed debugging information
	DEBUG Level = iota
	// INFO level for general informational messages
	INFO
	// WARN level for warning messages
	WARN
	// ERROR level for error messages
	ERROR
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

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
	config     Config
	stderrLog  *log.Logger
	fileLog    *log.Logger
	file       *os.File
	mu         sync.Mutex
	currentSize int64
}

// New creates a new logger with the given configuration
func New(config Config) (Logger, error) {
	l := &entryLogger{
		config:    config,
		stderrLog: log.New(os.Stderr, "", 0),
	}

	if !config.Enabled {
		return l, nil
	}

	// Setup file logging if output file is specified
	if config.OutputFile != "" {
		if err := l.setupFileLogging(); err != nil {
			return nil, fmt.Errorf("failed to setup file logging: %w", err)
		}
	}

	return l, nil
}

// setupFileLogging initializes file logging
func (l *entryLogger) setupFileLogging() error {
	// Ensure directory exists
	dir := filepath.Dir(l.config.OutputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Check if rotation is needed
	if l.config.MaxSize > 0 {
		if info, err := os.Stat(l.config.OutputFile); err == nil {
			if info.Size() >= l.config.MaxSize {
				if err := l.rotateLogFile(); err != nil {
					return err
				}
			}
		}
	}

	// Open log file
	file, err := os.OpenFile(l.config.OutputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file
	l.fileLog = log.New(file, "", 0)

	// Get current file size
	if info, err := file.Stat(); err == nil {
		l.currentSize = info.Size()
	}

	return nil
}

// rotateLogFile rotates the log file
func (l *entryLogger) rotateLogFile() error {
	// Close current file if open
	if l.file != nil {
		l.file.Close()
	}

	// Rename current log file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	rotatedPath := fmt.Sprintf("%s.%s", l.config.OutputFile, timestamp)

	if err := os.Rename(l.config.OutputFile, rotatedPath); err != nil {
		// If file doesn't exist, that's okay
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to rotate log file: %w", err)
		}
	}

	return nil
}

// log is the internal logging method
func (l *entryLogger) log(level Level, format string, args ...interface{}) {
	if !l.config.Enabled {
		return
	}

	if level < l.config.Level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] %s: %s\n", timestamp, level.String(), message)

	// Write to stderr
	l.stderrLog.Print(logLine)

	// Write to file if configured
	if l.fileLog != nil {
		l.fileLog.Print(logLine)
		l.currentSize += int64(len(logLine))

		// Check if rotation is needed
		if l.config.MaxSize > 0 && l.currentSize >= l.config.MaxSize {
			if err := l.rotateLogFile(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
			} else {
				// Re-setup file logging after rotation
				if err := l.setupFileLogging(); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to re-setup file logging: %v\n", err)
				}
			}
		}
	}
}

// Debug logs a debug message
func (l *entryLogger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *entryLogger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *entryLogger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *entryLogger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Close closes the logger and any open file handles
func (l *entryLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
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

// GetDefaultLogPath returns the default log file path
func GetDefaultLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}
	return filepath.Join(home, ".config", "entry", "logs", "entry.log"), nil
}

// MultiWriter creates a writer that writes to multiple writers
type MultiWriter struct {
	writers []io.Writer
}

// NewMultiWriter creates a new MultiWriter
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

// Write writes to all writers
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return len(p), nil
}
