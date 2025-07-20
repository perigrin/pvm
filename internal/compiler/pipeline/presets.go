// ABOUTME: Pre-built pipeline configurations for common transformation scenarios
// ABOUTME: Provides ready-to-use pipelines for clean Perl compilation, code formatting, etc.

package pipeline

import (
	"tamarou.com/pvm/internal/types"
)

// PipelinePreset represents a pre-configured transformation pipeline
type PipelinePreset struct {
	Name        string
	Description string
	Pipeline    TransformationPipeline
}

// GetAllPresets returns all available pipeline presets (static ones only)
func GetAllPresets() []PipelinePreset {
	return []PipelinePreset{
		GetCleanPerlPreset(),
		GetTypedPerlPreset(),
		GetCodeFormatterPreset(),
		GetTypedFormatterPreset(),
		GetMinimalPreset(),
	}
}

// GetCleanPerlPreset returns a pipeline for compiling typed Perl to clean Perl
func GetCleanPerlPreset() PipelinePreset {
	pipeline := NewBuilder().
		Add(NewTypeRemovalTransformer()).
		Add(NewWhitespaceNormalizerTransformer()).
		Build()

	return PipelinePreset{
		Name:        "clean_perl",
		Description: "Removes type annotations and normalizes whitespace for clean Perl output",
		Pipeline:    pipeline,
	}
}

// GetTypedPerlPreset returns a pipeline for preserving typed Perl
func GetTypedPerlPreset() PipelinePreset {
	pipeline := NewBuilder().
		Add(NewTypePreservationTransformer()).
		Add(NewWhitespaceNormalizerTransformer()).
		Build()

	return PipelinePreset{
		Name:        "typed_perl",
		Description: "Preserves type annotations and normalizes whitespace for typed Perl output",
		Pipeline:    pipeline,
	}
}

// GetCodeFormatterPreset returns a pipeline for code formatting without type removal
func GetCodeFormatterPreset() PipelinePreset {
	pipeline := NewBuilder().
		Add(NewWhitespaceNormalizerTransformer()).
		Add(NewIndentationNormalizerTransformer(4, false)). // 4 spaces
		Build()

	return PipelinePreset{
		Name:        "formatter",
		Description: "Formats code with consistent whitespace and indentation (preserves types)",
		Pipeline:    pipeline,
	}
}

// GetTypedFormatterPreset returns a pipeline for formatting typed Perl specifically
func GetTypedFormatterPreset() PipelinePreset {
	pipeline := NewBuilder().
		Add(NewWhitespaceNormalizerWithOptions(true)). // Preserve types
		Add(NewIndentationNormalizerTransformer(4, false)).
		Build()

	return PipelinePreset{
		Name:        "typed_formatter",
		Description: "Formats typed Perl code with type preservation and consistent indentation",
		Pipeline:    pipeline,
	}
}

// GetMinimalPreset returns a minimal pipeline for testing
func GetMinimalPreset() PipelinePreset {
	pipeline := NewBuilder().
		Add(NewWhitespaceNormalizerTransformer()).
		Build()

	return PipelinePreset{
		Name:        "minimal",
		Description: "Minimal pipeline with only whitespace normalization",
		Pipeline:    pipeline,
	}
}

// GetTabFormatterPreset returns a pipeline for formatting with tabs
func GetTabFormatterPreset() PipelinePreset {
	pipeline := NewBuilder().
		Add(NewWhitespaceNormalizerTransformer()).
		Add(NewIndentationNormalizerTransformer(1, true)). // Use tabs
		Build()

	return PipelinePreset{
		Name:        "tab_formatter",
		Description: "Formats code using tabs for indentation",
		Pipeline:    pipeline,
	}
}

// GetInferredTypedPerlPreset returns a pipeline for adding type annotations to untyped Perl
func GetInferredTypedPerlPreset(typeInfo map[string]*types.TypeInfo, options TypeInjectionOptions) PipelinePreset {
	pipeline := NewBuilder().
		Add(NewTypeInjectionTransformer(typeInfo, options)).
		Add(NewWhitespaceNormalizerTransformer()).
		Build()

	return PipelinePreset{
		Name:        "inferred_typed_perl",
		Description: "Adds inferred type annotations to untyped Perl code",
		Pipeline:    pipeline,
	}
}

// PipelineBuilder provides a fluent interface for building custom pipelines
type PipelineBuilder struct {
	builder *Builder
}

// NewPipelineBuilder creates a new pipeline builder with default options
func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		builder: NewBuilder(),
	}
}

// WithTypeRemoval adds type removal to the pipeline
func (pb *PipelineBuilder) WithTypeRemoval() *PipelineBuilder {
	pb.builder.Add(NewTypeRemovalTransformer())
	return pb
}

// WithTypePreservation adds type preservation to the pipeline
func (pb *PipelineBuilder) WithTypePreservation() *PipelineBuilder {
	pb.builder.Add(NewTypePreservationTransformer())
	return pb
}

// WithWhitespaceNormalization adds whitespace normalization to the pipeline
func (pb *PipelineBuilder) WithWhitespaceNormalization() *PipelineBuilder {
	pb.builder.Add(NewWhitespaceNormalizerTransformer())
	return pb
}

// WithSpaceIndentation adds space-based indentation normalization to the pipeline
func (pb *PipelineBuilder) WithSpaceIndentation(size int) *PipelineBuilder {
	pb.builder.Add(NewIndentationNormalizerTransformer(size, false))
	return pb
}

// WithTabIndentation adds tab-based indentation normalization to the pipeline
func (pb *PipelineBuilder) WithTabIndentation() *PipelineBuilder {
	pb.builder.Add(NewIndentationNormalizerTransformer(1, true))
	return pb
}

// WithTypeInjection adds type injection to the pipeline
func (pb *PipelineBuilder) WithTypeInjection(typeInfo map[string]*types.TypeInfo, options TypeInjectionOptions) *PipelineBuilder {
	pb.builder.Add(NewTypeInjectionTransformer(typeInfo, options))
	return pb
}

// WithCustomTransformer adds a custom transformer to the pipeline
func (pb *PipelineBuilder) WithCustomTransformer(transformer Transformer) *PipelineBuilder {
	pb.builder.Add(transformer)
	return pb
}

// WithOptions sets the pipeline options
func (pb *PipelineBuilder) WithOptions(options PipelineOptions) *PipelineBuilder {
	pb.builder.WithOptions(options)
	return pb
}

// Build returns the constructed pipeline
func (pb *PipelineBuilder) Build() TransformationPipeline {
	return pb.builder.Build()
}

// Common pipeline configurations as functions for convenience

// CreateCleanPerlPipeline creates a pipeline for clean Perl compilation
func CreateCleanPerlPipeline() TransformationPipeline {
	return GetCleanPerlPreset().Pipeline
}

// CreateFormatterPipeline creates a pipeline for code formatting
func CreateFormatterPipeline() TransformationPipeline {
	return GetCodeFormatterPreset().Pipeline
}

// CreateTypedFormatterPipeline creates a pipeline for typed Perl formatting
func CreateTypedFormatterPipeline() TransformationPipeline {
	return GetTypedFormatterPreset().Pipeline
}

// CreateCustomPipeline creates a custom pipeline using the builder pattern
func CreateCustomPipeline(builderFunc func(*PipelineBuilder) *PipelineBuilder) TransformationPipeline {
	builder := NewPipelineBuilder()
	return builderFunc(builder).Build()
}

// CreateOptimizedPipeline creates a pipeline with optimization enabled
func CreateOptimizedPipeline(basePreset string) TransformationPipeline {
	var basePipeline TransformationPipeline

	switch basePreset {
	case "clean_perl":
		basePipeline = GetCleanPerlPreset().Pipeline
	case "formatter":
		basePipeline = GetCodeFormatterPreset().Pipeline
	case "typed_formatter":
		basePipeline = GetTypedFormatterPreset().Pipeline
	default:
		basePipeline = GetMinimalPreset().Pipeline
	}

	options := DefaultPipelineOptions()
	options.EnableOptimizations = true
	options.Debug = false

	return basePipeline.WithOptions(options)
}

// CreateDebugPipeline creates a pipeline with debug enabled
func CreateDebugPipeline(basePreset string) TransformationPipeline {
	var basePipeline TransformationPipeline

	switch basePreset {
	case "clean_perl":
		basePipeline = GetCleanPerlPreset().Pipeline
	case "formatter":
		basePipeline = GetCodeFormatterPreset().Pipeline
	case "typed_formatter":
		basePipeline = GetTypedFormatterPreset().Pipeline
	default:
		basePipeline = GetMinimalPreset().Pipeline
	}

	options := DefaultPipelineOptions()
	options.Debug = true
	options.EnableOptimizations = false

	return basePipeline.WithOptions(options)
}
