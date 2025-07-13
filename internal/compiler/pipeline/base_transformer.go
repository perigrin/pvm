// ABOUTME: Base transformer implementation and utilities for building transformers
// ABOUTME: Provides common functionality and adapter patterns for existing transformation rules

package pipeline

import (
	"fmt"
	"runtime"
	"time"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// BaseTransformer provides common functionality for transformers
type BaseTransformer struct {
	name        string
	description string
}

// NewBaseTransformer creates a new base transformer with the given name and description
func NewBaseTransformer(name, description string) BaseTransformer {
	return BaseTransformer{
		name:        name,
		description: description,
	}
}

// Name returns the transformer name
func (bt BaseTransformer) Name() string {
	return bt.name
}

// Description returns the transformer description
func (bt BaseTransformer) Description() string {
	return bt.description
}

// CanSkip returns false by default - subclasses can override
func (bt BaseTransformer) CanSkip(input *TransformationInput) bool {
	return false
}

// measureMemory captures memory usage before and after an operation
func (bt BaseTransformer) measureMemory(operation func() error) (int64, error) {
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	err := operation()

	runtime.ReadMemStats(&m2)
	memoryDelta := int64(m2.Alloc - m1.Alloc)

	return memoryDelta, err
}

// createOutput creates a TransformationOutput with the given parameters
func (bt BaseTransformer) createOutput(cst *sitter.Node, content []byte, modified bool, metrics TransformationMetrics) *TransformationOutput {
	return &TransformationOutput{
		CST:      cst,
		Content:  content,
		Modified: modified,
		Metrics:  metrics,
	}
}

// NoOpTransformer is a transformer that does nothing (useful for testing)
type NoOpTransformer struct {
	BaseTransformer
}

// NewNoOpTransformer creates a new no-op transformer
func NewNoOpTransformer() Transformer {
	return &NoOpTransformer{
		BaseTransformer: NewBaseTransformer("noop", "Does nothing - useful for testing"),
	}
}

// Transform performs no transformation
func (nt *NoOpTransformer) Transform(input *TransformationInput) (*TransformationOutput, error) {
	metrics := TransformationMetrics{
		NodesProcessed:  1, // Count the root node
		BytesProcessed:  len(input.Content),
		MemoryAllocated: 0,
	}

	return nt.createOutput(input.CST, input.Content, false, metrics), nil
}

// CanSkip always returns true for no-op transformer when optimizations are enabled
func (nt *NoOpTransformer) CanSkip(input *TransformationInput) bool {
	return input.Context.Options.EnableOptimizations
}

// FunctionTransformer wraps a simple function as a transformer
type FunctionTransformer struct {
	BaseTransformer
	transformFunc func(*sitter.Node, []byte) ([]byte, error)
	canSkipFunc   func(*TransformationInput) bool
}

// NewFunctionTransformer creates a transformer from a function
func NewFunctionTransformer(name, description string, transformFunc func(*sitter.Node, []byte) ([]byte, error)) Transformer {
	return &FunctionTransformer{
		BaseTransformer: NewBaseTransformer(name, description),
		transformFunc:   transformFunc,
		canSkipFunc:     func(*TransformationInput) bool { return false },
	}
}

// NewFunctionTransformerWithSkip creates a transformer from a function with custom skip logic
func NewFunctionTransformerWithSkip(name, description string, transformFunc func(*sitter.Node, []byte) ([]byte, error), canSkipFunc func(*TransformationInput) bool) Transformer {
	return &FunctionTransformer{
		BaseTransformer: NewBaseTransformer(name, description),
		transformFunc:   transformFunc,
		canSkipFunc:     canSkipFunc,
	}
}

// Transform executes the wrapped function
func (ft *FunctionTransformer) Transform(input *TransformationInput) (*TransformationOutput, error) {
	startTime := time.Now()

	var transformedContent []byte
	var memoryUsed int64
	var err error

	// Measure memory usage during transformation
	memoryUsed, err = ft.measureMemory(func() error {
		transformedContent, err = ft.transformFunc(input.CST, input.Content)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("function transformer '%s' failed: %w", ft.name, err)
	}

	// Determine if content was modified
	modified := string(transformedContent) != string(input.Content)

	// Count nodes (simple traversal)
	nodeCount := ft.countNodes(input.CST)

	metrics := TransformationMetrics{
		NodesProcessed:  nodeCount,
		BytesProcessed:  len(input.Content),
		MemoryAllocated: memoryUsed,
	}

	duration := time.Since(startTime)
	_ = duration // We track duration at the pipeline level

	return ft.createOutput(input.CST, transformedContent, modified, metrics), nil
}

// CanSkip delegates to the wrapped function
func (ft *FunctionTransformer) CanSkip(input *TransformationInput) bool {
	return ft.canSkipFunc(input)
}

// countNodes counts the total number of nodes in the CST
func (ft *FunctionTransformer) countNodes(node *sitter.Node) int {
	if node == nil {
		return 0
	}

	count := 1
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			count += ft.countNodes(child)
		}
	}

	return count
}

// ConditionalTransformer only executes if a condition is met
type ConditionalTransformer struct {
	BaseTransformer
	condition   func(*TransformationInput) bool
	transformer Transformer
}

// NewConditionalTransformer creates a transformer that only executes if condition is true
func NewConditionalTransformer(name string, condition func(*TransformationInput) bool, transformer Transformer) Transformer {
	return &ConditionalTransformer{
		BaseTransformer: NewBaseTransformer(name, fmt.Sprintf("Conditional: %s", transformer.Description())),
		condition:       condition,
		transformer:     transformer,
	}
}

// Transform executes the wrapped transformer only if condition is met
func (ct *ConditionalTransformer) Transform(input *TransformationInput) (*TransformationOutput, error) {
	if !ct.condition(input) {
		// Condition not met, return input unchanged
		metrics := TransformationMetrics{
			NodesProcessed: 1,
			BytesProcessed: len(input.Content),
		}
		return ct.createOutput(input.CST, input.Content, false, metrics), nil
	}

	return ct.transformer.Transform(input)
}

// CanSkip returns true if condition is not met or wrapped transformer can skip
func (ct *ConditionalTransformer) CanSkip(input *TransformationInput) bool {
	if !ct.condition(input) {
		return true
	}
	return ct.transformer.CanSkip(input)
}
