// ABOUTME: Tool execution data structures and types
// ABOUTME: Defines core types for global tool execution functionality

package tool

import "time"

// ExecutionMode represents the mode PVX should operate in
type ExecutionMode int

const (
	// ModeScript indicates PVX should execute a script file
	ModeScript ExecutionMode = iota
	// ModeTool indicates PVX should execute a global tool
	ModeTool
	// ModeInline indicates PVX should execute inline code
	ModeInline
	// ModeAmbiguous indicates the execution mode cannot be determined
	ModeAmbiguous
)

// String returns the string representation of ExecutionMode
func (m ExecutionMode) String() string {
	switch m {
	case ModeScript:
		return "script"
	case ModeTool:
		return "tool"
	case ModeInline:
		return "inline"
	case ModeAmbiguous:
		return "ambiguous"
	default:
		return "unknown"
	}
}

// DetectionResult contains the result of execution mode detection
type DetectionResult struct {
	Mode         ExecutionMode
	ToolName     string
	ScriptPath   string
	InlineCode   string
	Arguments    []string
	Confidence   float64
	Reason       string
	Alternatives []string
}

// ToolRequest represents a request to execute a tool
type ToolRequest struct {
	Name      string
	Arguments []string
	Version   string
	Force     bool
}

// ToolInfo contains information about a tool
type ToolInfo struct {
	Name        string
	Module      string
	Version     string
	Description string
	Executable  string
	InstallDate time.Time
	LastUsed    time.Time
	Source      ToolSource
}

// ToolSource indicates where a tool mapping comes from
type ToolSource int

const (
	// SourceBuiltin indicates the tool mapping is built-in
	SourceBuiltin ToolSource = iota
	// SourceConfig indicates the tool mapping comes from configuration
	SourceConfig
	// SourceCPAN indicates the tool mapping was resolved from CPAN
	SourceCPAN
	// SourceCache indicates the tool mapping comes from cache
	SourceCache
)

// String returns the string representation of ToolSource
func (s ToolSource) String() string {
	switch s {
	case SourceBuiltin:
		return "builtin"
	case SourceConfig:
		return "config"
	case SourceCPAN:
		return "cpan"
	case SourceCache:
		return "cache"
	default:
		return "unknown"
	}
}

// ValidationError represents a validation error with detailed information
type ValidationError struct {
	Field   string
	Value   string
	Reason  string
	Suggest []string
}

// Error implements the error interface
func (v *ValidationError) Error() string {
	return v.Reason
}
