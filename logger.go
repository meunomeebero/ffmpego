package ffmpego

import (
	"io"
	"log"
)

// Logger interface for logging operations
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Section(name string)
	Step(format string, args ...interface{})
	Success(format string, args ...interface{})
}

// DefaultLogger implements the Logger interface using Go's standard log package
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger(out io.Writer) *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(out, "", log.LstdFlags),
	}
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	l.logger.Printf("[DEBUG] "+format, args...)
}

// Info logs an info message
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	l.logger.Printf("[INFO] "+format, args...)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(format string, args ...interface{}) {
	l.logger.Printf("[WARN] "+format, args...)
}

// Error logs an error message
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+format, args...)
}

// Section logs a section header
func (l *DefaultLogger) Section(name string) {
	l.logger.Printf("[SECTION] === %s ===", name)
}

// Step logs a step in a process
func (l *DefaultLogger) Step(format string, args ...interface{}) {
	l.logger.Printf("[STEP] "+format, args...)
}

// Success logs a success message
func (l *DefaultLogger) Success(format string, args ...interface{}) {
	l.logger.Printf("[SUCCESS] "+format, args...)
}
