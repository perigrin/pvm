// ABOUTME: Core pipeline implementation for CST transformations
// ABOUTME: Provides composable, sequential processing of tree-sitter CST nodes

package pipeline

import (
	"fmt"
	"runtime"
	"time"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Pipeline implements TransformationPipeline
type Pipeline struct {
	transformers []Transformer
	options      PipelineOptions
}

// NewPipeline creates a new transformation pipeline with default options
func NewPipeline() TransformationPipeline {
	return &Pipeline{
		transformers: make([]Transformer, 0),
		options:      DefaultPipelineOptions(),
	}
}

// NewPipelineWithOptions creates a new transformation pipeline with specified options
func NewPipelineWithOptions(options PipelineOptions) TransformationPipeline {
	return &Pipeline{
		transformers: make([]Transformer, 0),
		options:      options,
	}
}

// AddTransformer adds a transformer to the pipeline
func (p *Pipeline) AddTransformer(transformer Transformer) TransformationPipeline {
	if len(p.transformers) >= p.options.MaxTransformers {
		// Return a new pipeline to maintain immutability
		return &Pipeline{
			transformers: p.transformers,
			options:      p.options,
		}
	}

	// Create a new pipeline with the added transformer
	newTransformers := make([]Transformer, len(p.transformers)+1)
	copy(newTransformers, p.transformers)
	newTransformers[len(p.transformers)] = transformer

	return &Pipeline{
		transformers: newTransformers,
		options:      p.options,
	}
}

// WithOptions creates a new pipeline with modified options
func (p *Pipeline) WithOptions(options PipelineOptions) TransformationPipeline {
	return &Pipeline{
		transformers: p.transformers,
		options:      options,
	}
}

// GetTransformers returns the list of transformers in execution order
func (p *Pipeline) GetTransformers() []Transformer {
	// Return a copy to prevent external modification
	result := make([]Transformer, len(p.transformers))
	copy(result, p.transformers)
	return result
}

// Execute runs the pipeline on the given CST and content
func (p *Pipeline) Execute(cst *sitter.Node, content []byte) (*TransformationResult, error) {
	if cst == nil {
		return nil, fmt.Errorf("CST node cannot be nil")
	}
	if content == nil {
		return nil, fmt.Errorf("content cannot be nil")
	}

	startTime := time.Now()

	// Initialize transformation context
	context := &TransformationContext{
		Options: p.options,
		State:   make(map[string]interface{}),
	}

	// Initialize metrics
	pipelineMetrics := PipelineMetrics{
		TransformerDurations: make(map[string]int64),
		TotalMetrics:         TransformationMetrics{},
	}

	// Track transformation steps
	transformationSteps := make([]TransformationStep, 0, len(p.transformers))

	// Current state being passed through the pipeline
	currentCST := cst
	currentContent := content

	// Execute each transformer in sequence
	for i, transformer := range p.transformers {
		stepStartTime := time.Now()

		// Create input for this transformer
		input := &TransformationInput{
			CST:     currentCST,
			Content: currentContent,
			Context: context,
			Index:   i,
		}

		// Check if transformer can be skipped
		if p.options.EnableOptimizations && transformer.CanSkip(input) {
			pipelineMetrics.SkippedTransformers++

			if p.options.Debug {
				step := TransformationStep{
					Name:        transformer.Name(),
					Description: transformer.Description(),
					Modified:    false,
					Metrics:     TransformationMetrics{},
					Duration:    0,
				}
				transformationSteps = append(transformationSteps, step)
			}
			continue
		}

		// Execute the transformer
		output, err := transformer.Transform(input)
		if err != nil {
			return nil, fmt.Errorf("transformer '%s' failed: %w", transformer.Name(), err)
		}

		// Validate output
		if output == nil {
			return nil, fmt.Errorf("transformer '%s' returned nil output", transformer.Name())
		}

		// Update current state
		if output.Modified {
			currentCST = output.CST
			currentContent = output.Content
		}

		// Record metrics
		stepDuration := time.Since(stepStartTime).Nanoseconds()
		pipelineMetrics.TransformerDurations[transformer.Name()] = stepDuration
		pipelineMetrics.TotalMetrics.NodesProcessed += output.Metrics.NodesProcessed
		pipelineMetrics.TotalMetrics.BytesProcessed += output.Metrics.BytesProcessed
		pipelineMetrics.TotalMetrics.MemoryAllocated += output.Metrics.MemoryAllocated
		pipelineMetrics.TotalMetrics.CacheHits += output.Metrics.CacheHits
		pipelineMetrics.TotalMetrics.CacheMisses += output.Metrics.CacheMisses

		// Record transformation step
		step := TransformationStep{
			Name:        transformer.Name(),
			Description: transformer.Description(),
			Modified:    output.Modified,
			Metrics:     output.Metrics,
			Duration:    stepDuration,
		}
		transformationSteps = append(transformationSteps, step)
	}

	// Calculate total duration
	pipelineMetrics.TotalDuration = time.Since(startTime).Nanoseconds()

	// Create final result
	result := &TransformationResult{
		Content:         string(currentContent),
		CST:             currentCST,
		Metrics:         pipelineMetrics,
		Transformations: transformationSteps,
	}

	return result, nil
}

// Builder provides a fluent interface for building pipelines
type Builder struct {
	pipeline *Pipeline
}

// NewBuilder creates a new pipeline builder
func NewBuilder() *Builder {
	return &Builder{
		pipeline: &Pipeline{
			transformers: make([]Transformer, 0),
			options:      DefaultPipelineOptions(),
		},
	}
}

// Add adds a transformer to the pipeline being built
func (b *Builder) Add(transformer Transformer) *Builder {
	updated := b.pipeline.AddTransformer(transformer)
	//nolint:go-critic
	b.pipeline = updated.(*Pipeline)
	return b
}

// WithOptions sets the options for the pipeline being built
func (b *Builder) WithOptions(options PipelineOptions) *Builder {
	updated := b.pipeline.WithOptions(options)
	//nolint:go-critic
	b.pipeline = updated.(*Pipeline)
	return b
}

// Build returns the constructed pipeline
func (b *Builder) Build() TransformationPipeline {
	return b.pipeline
}

// GetMemoryUsage returns current memory usage statistics
func GetMemoryUsage() TransformationMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return TransformationMetrics{
		MemoryAllocated: int64(m.Alloc),
	}
}
