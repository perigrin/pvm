// ABOUTME: Logging framework for the PVM Ecosystem
// ABOUTME: Provides consistent logging across all components

package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Log levels
const (
	LevelDebug = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

// Level names for display
var levelNames = map[int]string{
	LevelDebug:   "DEBUG",
	LevelInfo:    "INFO",
	LevelWarning: "WARNING",
	LevelError:   "ERROR",
	LevelFatal:   "FATAL",
}

// Logger is a simple logging interface
type Logger struct {
	level      int
	output     io.Writer
	component  string
	timeFormat string
	mu         sync.Mutex
	quiet      bool
}

// Global logger instance
var (
	globalLogger *Logger
	globalMu     sync.Mutex
)

// init initializes the global logger
func init() {
	globalLogger = &Logger{
		level:      LevelInfo,
		output:     os.Stderr,
		timeFormat: time.RFC3339,
	}
}

// NewLogger creates a new logger with the specified level and output
func NewLogger(level int, output io.Writer, component string) *Logger {
	return &Logger{
		level:      level,
		output:     output,
		component:  component,
		timeFormat: time.RFC3339,
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
}

// SetComponent sets the component name
func (l *Logger) SetComponent(component string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.component = component
}

// SetTimeFormat sets the time format string
func (l *Logger) SetTimeFormat(format string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timeFormat = format
}

// SetQuiet enables or disables quiet mode (no output)
func (l *Logger) SetQuiet(quiet bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.quiet = quiet
}

// logf logs a message at the specified level
func (l *Logger) logf(level int, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.quiet || level < l.level {
		return
	}

	// Format the message
	msg := fmt.Sprintf(format, args...)

	// Format the log line
	timestamp := time.Now().Format(l.timeFormat)
	levelName := levelNames[level]

	var line string
	if l.component != "" {
		line = fmt.Sprintf("%s [%s] [%s] %s\n", timestamp, levelName, l.component, msg)
	} else {
		line = fmt.Sprintf("%s [%s] %s\n", timestamp, levelName, msg)
	}

	// Write to output
	_, _ = fmt.Fprint(l.output, line)

	// Exit if fatal
	if level == LevelFatal {
		os.Exit(1)
	}
}

// Debugf logs a debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logf(LevelDebug, format, args...)
}

// Infof logs an info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf(LevelInfo, format, args...)
}

// Warningf logs a warning message
func (l *Logger) Warningf(format string, args ...interface{}) {
	l.logf(LevelWarning, format, args...)
}

// Errorf logs an error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf(LevelError, format, args...)
}

// Fatalf logs a fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logf(LevelFatal, format, args...)
}

// Global logger functions

// SetGlobalLevel sets the global logger level
func SetGlobalLevel(level int) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger.SetLevel(level)
}

// SetGlobalOutput sets the global logger output
func SetGlobalOutput(output io.Writer) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger.SetOutput(output)
}

// SetGlobalComponent sets the global logger component
func SetGlobalComponent(component string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger.SetComponent(component)
}

// SetGlobalQuiet enables or disables quiet mode for the global logger
func SetGlobalQuiet(quiet bool) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger.SetQuiet(quiet)
}

// Debugf logs a debug message to the global logger
func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// Infof logs an info message to the global logger
func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Warningf logs a warning message to the global logger
func Warningf(format string, args ...interface{}) {
	globalLogger.Warningf(format, args...)
}

// Errorf logs an error message to the global logger
func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// Fatalf logs a fatal message to the global logger and exits
func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatalf(format, args...)
}

// ParseLevel parses a level string to a level constant
func ParseLevel(level string) (int, error) {
	level = strings.ToUpper(level)
	
	switch level {
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARNING", "WARN":
		return LevelWarning, nil
	case "ERROR":
		return LevelError, nil
	case "FATAL":
		return LevelFatal, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level: %s", level)
	}
}