// ABOUTME: Type definitions for language service pooling and analysis
// ABOUTME: Defines result types and context structures for efficient LS operations

package ls

import "time"

// AnalysisResult represents the result of a language analysis operation
type AnalysisResult struct {
	URI       string
	Errors    []error
	StartTime time.Time
	EndTime   time.Time
}

// CompletionResult represents the result of a completion analysis
type CompletionResult struct {
	URI      string
	Position Position
	Items    []string
}

// DefinitionResult represents the result of a definition lookup
type DefinitionResult struct {
	URI       string
	Position  Position
	Locations []Location
}

// DiagnosticResult represents the result of diagnostic analysis
type DiagnosticResult struct {
	URI         string
	Diagnostics []string
}

// HoverResult represents the result of a hover request
type HoverResult struct {
	URI      string
	Position Position
	Content  string
}

// AnalysisContext represents context for an analysis operation
type AnalysisContext struct {
	URI       string
	StartTime time.Time
	Metadata  map[string]interface{}
}

// RequestContext represents context for a language service request
type RequestContext struct {
	RequestID string
	Method    string
	StartTime time.Time
	Metadata  map[string]interface{}
}
