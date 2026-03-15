// ABOUTME: Progress tracking types and interfaces for CLI operations
// ABOUTME: Defines standardized progress reporting across all PVM components

package progress

import (
	"time"
)

// Tracker provides progress tracking for single operations
type Tracker interface {
	// Start begins tracking an operation
	Start(operation string, total int)

	// Update reports progress on the current operation
	Update(current int, message string)

	// Finish completes the operation with final result
	Finish(result *Result)

	// SetTotal updates the total count (useful for dynamic operations)
	SetTotal(total int)

	// SetMessage updates the current status message
	SetMessage(message string)

	// IsRunning returns true if tracking is currently active
	IsRunning() bool

	// GetProgress returns the current progress information
	GetProgress() *Status
}

// ParallelTracker provides progress tracking for parallel operations
type ParallelTracker interface {
	// StartParallel begins tracking multiple parallel operations
	StartParallel(operations []string)

	// UpdateOperation updates the status of a specific operation
	UpdateOperation(id string, status OperationStatus, message string)

	// FinishOperation completes a specific operation
	FinishOperation(id string, result *Result)

	// FinishAll completes all parallel operations
	FinishAll(results []*Result)

	// GetOperationStatus returns status for a specific operation
	GetOperationStatus(id string) *OperationStatus

	// GetOverallProgress returns aggregate progress information
	GetOverallProgress() *ParallelStatus
}

// Reporter provides callback-based progress reporting
type Reporter interface {
	// Subscribe adds a progress callback
	Subscribe(callback Callback)

	// Unsubscribe removes a progress callback
	Unsubscribe(callback Callback)

	// Notify sends a progress update to all subscribers
	Notify(status *Status)
}

// Status represents the current progress status
type Status struct {
	// Operation is the current operation name
	Operation string `json:"operation"`

	// Current is the current progress count
	Current int `json:"current"`

	// Total is the total expected count
	Total int `json:"total"`

	// Message is the current status message
	Message string `json:"message"`

	// Percentage is the completion percentage (0-100)
	Percentage float64 `json:"percentage"`

	// StartTime is when the operation started
	StartTime time.Time `json:"start_time"`

	// ElapsedTime is the time elapsed since start
	ElapsedTime time.Duration `json:"elapsed_time"`

	// EstimatedRemaining is the estimated time remaining
	EstimatedRemaining time.Duration `json:"estimated_remaining,omitempty"`

	// Stage represents the current operation stage
	Stage string `json:"stage,omitempty"`

	// SubOperation represents any sub-operation being performed
	SubOperation string `json:"sub_operation,omitempty"`

	// lastProgressChange tracks when progress last changed (for bounds checking)
	lastProgressChange time.Time `json:"-"`
}

// ParallelStatus represents the status of parallel operations
type ParallelStatus struct {
	// Operations maps operation IDs to their status
	Operations map[string]*OperationStatus `json:"operations"`

	// TotalOperations is the total number of operations
	TotalOperations int `json:"total_operations"`

	// CompletedOperations is the number of completed operations
	CompletedOperations int `json:"completed_operations"`

	// FailedOperations is the number of failed operations
	FailedOperations int `json:"failed_operations"`

	// RunningOperations is the number of currently running operations
	RunningOperations int `json:"running_operations"`

	// OverallPercentage is the overall completion percentage
	OverallPercentage float64 `json:"overall_percentage"`

	// StartTime is when parallel operations started
	StartTime time.Time `json:"start_time"`

	// ElapsedTime is the total time elapsed
	ElapsedTime time.Duration `json:"elapsed_time"`

	// EstimatedRemaining is the estimated time remaining
	EstimatedRemaining time.Duration `json:"estimated_remaining,omitempty"`
}

// OperationStatus represents the status of a single operation in parallel execution
type OperationStatus struct {
	// ID is the operation identifier
	ID string `json:"id"`

	// Name is the operation name
	Name string `json:"name"`

	// Status is the current operation status
	Status OpStatus `json:"status"`

	// Message is the current status message
	Message string `json:"message"`

	// Progress is the operation progress (0-100)
	Progress float64 `json:"progress"`

	// StartTime is when the operation started
	StartTime time.Time `json:"start_time,omitempty"`

	// EndTime is when the operation completed
	EndTime time.Time `json:"end_time,omitempty"`

	// Duration is the total operation duration
	Duration time.Duration `json:"duration,omitempty"`

	// Error contains error information if the operation failed
	Error error `json:"error,omitempty"`

	// Result contains the operation result
	Result *Result `json:"result,omitempty"`
}

// OpStatus represents the status of an operation
type OpStatus int

const (
	// StatusPending indicates the operation is pending
	StatusPending OpStatus = iota

	// StatusRunning indicates the operation is in progress
	StatusRunning

	// StatusCompleted indicates the operation completed successfully
	StatusCompleted

	// StatusFailed indicates the operation failed
	StatusFailed

	// StatusCancelled indicates the operation was cancelled
	StatusCancelled

	// StatusSkipped indicates the operation was skipped
	StatusSkipped
)

// String returns a string representation of the operation status
func (s OpStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusCancelled:
		return "cancelled"
	case StatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// Result represents the result of an operation
type Result struct {
	// Operation is the type of operation performed
	Operation string `json:"operation"`

	// Target is the module or target of the operation
	Target string `json:"target"`

	// Success indicates if the operation was successful
	Success bool `json:"success"`

	// Duration is the time taken for the operation
	Duration time.Duration `json:"duration"`

	// Message provides additional information
	Message string `json:"message,omitempty"`

	// Error contains error information if unsuccessful
	Error error `json:"error,omitempty"`

	// Metadata contains operation-specific metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Warnings contains any warnings generated during the operation
	Warnings []string `json:"warnings,omitempty"`

	// Stats contains operation statistics
	Stats *Stats `json:"stats,omitempty"`
}

// Stats contains operation statistics
type Stats struct {
	// ItemsProcessed is the number of items processed
	ItemsProcessed int `json:"items_processed"`

	// ItemsSuccessful is the number of successful items
	ItemsSuccessful int `json:"items_successful"`

	// ItemsFailed is the number of failed items
	ItemsFailed int `json:"items_failed"`

	// ItemsSkipped is the number of skipped items
	ItemsSkipped int `json:"items_skipped"`

	// BytesProcessed is the number of bytes processed
	BytesProcessed int64 `json:"bytes_processed,omitempty"`

	// AverageTime is the average processing time per item
	AverageTime time.Duration `json:"average_time,omitempty"`

	// ThroughputPerSecond is the processing throughput
	ThroughputPerSecond float64 `json:"throughput_per_second,omitempty"`
}

// Callback is called to report operation progress
type Callback func(status *Status)

// ParallelCallback is called to report parallel operation progress
type ParallelCallback func(status *ParallelStatus)

// DisplayOptions contains options for progress display
type DisplayOptions struct {
	// ShowPercentage shows completion percentage
	ShowPercentage bool

	// ShowETA shows estimated time of arrival
	ShowETA bool

	// ShowThroughput shows processing throughput
	ShowThroughput bool

	// ShowElapsed shows elapsed time
	ShowElapsed bool

	// RefreshRate is the display refresh rate
	RefreshRate time.Duration

	// Width is the progress bar width
	Width int

	// Style is the progress bar style
	Style DisplayStyle

	// Color enables colored output
	Color bool

	// Quiet suppresses most output
	Quiet bool

	// Verbose shows detailed progress information
	Verbose bool
}

// DisplayStyle represents different progress display styles
type DisplayStyle int

const (
	// StyleBar shows a traditional progress bar
	StyleBar DisplayStyle = iota

	// StyleSpinner shows a spinner animation
	StyleSpinner

	// StyleDots shows dots progress
	StyleDots

	// StyleLines shows line-based progress
	StyleLines

	// StyleJSON outputs progress as JSON
	StyleJSON

	// StyleSilent suppresses all visual progress
	StyleSilent
)

// String returns a string representation of the display style
func (s DisplayStyle) String() string {
	switch s {
	case StyleBar:
		return "bar"
	case StyleSpinner:
		return "spinner"
	case StyleDots:
		return "dots"
	case StyleLines:
		return "lines"
	case StyleJSON:
		return "json"
	case StyleSilent:
		return "silent"
	default:
		return "unknown"
	}
}

// TaskInfo represents information about a task for progress tracking
type TaskInfo struct {
	// ID is the unique task identifier
	ID string `json:"id"`

	// Name is the human-readable task name
	Name string `json:"name"`

	// Description provides additional task details
	Description string `json:"description,omitempty"`

	// Category groups related tasks
	Category string `json:"category,omitempty"`

	// Priority indicates task priority
	Priority int `json:"priority,omitempty"`

	// EstimatedDuration is the expected task duration
	EstimatedDuration time.Duration `json:"estimated_duration,omitempty"`

	// Dependencies lists tasks that must complete before this one
	Dependencies []string `json:"dependencies,omitempty"`

	// Metadata contains task-specific metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SummaryInfo represents a summary of operation results
type SummaryInfo struct {
	// TotalOperations is the total number of operations
	TotalOperations int `json:"total_operations"`

	// SuccessfulOperations is the number of successful operations
	SuccessfulOperations int `json:"successful_operations"`

	// FailedOperations is the number of failed operations
	FailedOperations int `json:"failed_operations"`

	// SkippedOperations is the number of skipped operations
	SkippedOperations int `json:"skipped_operations"`

	// TotalDuration is the total time for all operations
	TotalDuration time.Duration `json:"total_duration"`

	// AverageDuration is the average operation duration
	AverageDuration time.Duration `json:"average_duration"`

	// Errors lists all errors encountered
	Errors []string `json:"errors,omitempty"`

	// Warnings lists all warnings generated
	Warnings []string `json:"warnings,omitempty"`

	// Stats contains aggregate statistics
	Stats *Stats `json:"stats,omitempty"`
}
