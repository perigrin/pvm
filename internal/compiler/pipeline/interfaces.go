// ABOUTME: Core interfaces and types for the transformation pipeline system
// ABOUTME: Defines the contract for composable CST transformations

package pipeline

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// TransformationPipeline represents a composable sequence of CST transformations
type TransformationPipeline interface {
	// AddTransformer adds a transformer to the pipeline
	AddTransformer(transformer Transformer) TransformationPipeline
	// Execute runs the pipeline on the given CST and content
	Execute(cst *sitter.Node, content []byte) (*TransformationResult, error)
	// WithOptions creates a new pipeline with modified options
	WithOptions(options PipelineOptions) TransformationPipeline
	// GetTransformers returns the list of transformers in execution order
	GetTransformers() []Transformer
}

// Transformer represents a single CST transformation step
type Transformer interface {
	// Name returns a human-readable name for this transformer
	Name() string
	// Transform performs the transformation on the input
	Transform(input *TransformationInput) (*TransformationOutput, error)
	// CanSkip returns true if this transformer can be skipped for the given input
	CanSkip(input *TransformationInput) bool
	// Description returns a description of what this transformer does
	Description() string
}

// TransformationInput contains all data needed for a transformation
type TransformationInput struct {
	// CST is the current state of the CST being transformed
	CST *sitter.Node
	// Content is the source code content corresponding to the CST
	Content []byte
	// Context provides additional transformation context
	Context *TransformationContext
	// Index is the position of this transformer in the pipeline
	Index int
}

// TransformationOutput contains the result of a transformation
type TransformationOutput struct {
	// CST is the transformed CST (may be the same as input if no changes)
	CST *sitter.Node
	// Content is the updated source code content
	Content []byte
	// Modified indicates whether this transformer made any changes
	Modified bool
	// Metrics contains performance and statistics data
	Metrics TransformationMetrics
}

// TransformationResult contains the final result of pipeline execution
type TransformationResult struct {
	// Content is the final transformed source code
	Content string
	// CST is the final transformed CST
	CST *sitter.Node
	// Metrics contains aggregated metrics from all transformers
	Metrics PipelineMetrics
	// Transformations contains details about each transformation step
	Transformations []TransformationStep
}

// TransformationStep represents the result of a single transformer execution
type TransformationStep struct {
	// Name is the name of the transformer that executed
	Name string
	// Description is the description of what the transformer does
	Description string
	// Modified indicates whether this step made changes
	Modified bool
	// Metrics contains performance data for this step
	Metrics TransformationMetrics
	// Duration is how long this step took to execute
	Duration int64 // nanoseconds
}

// TransformationContext provides context and configuration for transformations
type TransformationContext struct {
	// Options contains pipeline-wide options
	Options PipelineOptions
	// Target specifies the compilation target (clean_perl, typed_perl, etc.)
	Target string
	// SourceFilename is the name of the file being transformed (for error reporting)
	SourceFilename string
	// State allows transformers to share state within a pipeline execution
	State map[string]interface{}
}

// PipelineOptions contains configuration options for the pipeline
type PipelineOptions struct {
	// PreserveComments controls whether comments are preserved during transformation
	PreserveComments bool
	// PreserveWhitespace controls whether original whitespace is preserved
	PreserveWhitespace bool
	// EnableOptimizations allows transformers to skip work when possible
	EnableOptimizations bool
	// Debug enables debug output and additional metrics collection
	Debug bool
	// MaxTransformers limits the number of transformers that can be added
	MaxTransformers int
}

// TransformationMetrics contains performance and statistics data
type TransformationMetrics struct {
	// NodesProcessed is the number of CST nodes processed
	NodesProcessed int
	// BytesProcessed is the number of content bytes processed
	BytesProcessed int
	// MemoryAllocated is the amount of memory allocated during transformation
	MemoryAllocated int64
	// CacheHits is the number of cache hits (if caching is enabled)
	CacheHits int
	// CacheMisses is the number of cache misses (if caching is enabled)
	CacheMisses int
}

// PipelineMetrics contains aggregated metrics from all transformers
type PipelineMetrics struct {
	// TotalDuration is the total time taken by the entire pipeline
	TotalDuration int64 // nanoseconds
	// TransformerDurations contains the duration of each transformer
	TransformerDurations map[string]int64
	// TotalMetrics is the sum of all transformer metrics
	TotalMetrics TransformationMetrics
	// SkippedTransformers is the number of transformers that were skipped
	SkippedTransformers int
}

// DefaultPipelineOptions returns sensible default options for a pipeline
func DefaultPipelineOptions() PipelineOptions {
	return PipelineOptions{
		PreserveComments:    true,
		PreserveWhitespace:  true,
		EnableOptimizations: true,
		Debug:               false,
		MaxTransformers:     50,
	}
}
