// ABOUTME: Type definitions and interfaces for PVM UI framework
// ABOUTME: Provides structured interfaces for Fang-powered CLI output

package ui

import (
	"io"
)

// OutputLevel represents the level of output messages
type OutputLevel int

const (
	LevelDebug OutputLevel = iota
	LevelInfo
	LevelSuccess
	LevelWarning
	LevelError
)

// String returns the string representation of the output level
func (l OutputLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelSuccess:
		return "SUCCESS"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// UIContext provides context for UI operations
type UIContext struct {
	Writer      io.Writer
	ErrorWriter io.Writer // Separate writer for error messages
	ColorMode   ColorMode
	Quiet       bool
	Verbose     bool
	Interactive bool
	RawMarkdown bool // Use plain text markdown instead of styled rendering
}

// ColorMode represents different color output modes
type ColorMode int

const (
	ColorAuto ColorMode = iota
	ColorAlways
	ColorNever
)

// OutputRenderer interface defines methods for rendering output
type OutputRenderer interface {
	// Basic output methods
	Info(message string, args ...interface{})
	Success(message string, args ...interface{})
	Warning(message string, args ...interface{})
	Error(message string, args ...interface{})
	Debug(message string, args ...interface{})

	// Formatted output methods
	Printf(format string, args ...interface{})
	Println(args ...interface{})

	// Structured output methods
	Table(headers []string, rows [][]string)
	List(items []string)
	KeyValue(pairs map[string]string)

	// Status and progress methods
	Status(message string)
	Progress(current, total int, message string)
}

// TableOptions configures table rendering
type TableOptions struct {
	Headers     []string
	Rows        [][]string
	Title       string
	ShowBorders bool
	Compact     bool
}

// ListOptions configures list rendering
type ListOptions struct {
	Items      []string
	Title      string
	Numbered   bool
	BulletChar string
}

// ProgressOptions configures progress display
type ProgressOptions struct {
	Current int
	Total   int
	Message string
	Width   int
}
