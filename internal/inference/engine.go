// ABOUTME: Main type inference engine interface and basic implementation
// ABOUTME: Provides literal type inference and variable propagation with confidence scoring

package inference

import (
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/types"
)

// TypeInferenceEngine defines the interface for type inference
type TypeInferenceEngine interface {
	// InferTypes performs type inference on an AST and returns an enhanced AST with type information
	InferTypes(inputAST *ast.AST) (ast.InferredAST, error)

	// CalculateConfidence calculates confidence score for a type inference
	CalculateConfidence(source types.TypeSource, context map[string]interface{}) float64

	// GetInferenceErrors returns any errors collected during inference
	GetInferenceErrors() []InferenceError

	// AddInferenceError adds an error to the collection
	AddInferenceError(err InferenceError)

	// ClearErrors clears all collected errors
	ClearErrors()
}

// basicInferenceEngine implements TypeInferenceEngine with simple literal-based inference
type basicInferenceEngine struct {
	// Collected errors during inference
	errors []InferenceError

	// Configuration options
	options InferenceOptions
}

// InferenceOptions holds configuration for the inference engine
type InferenceOptions struct {
	// EnableFlowAnalysis enables control flow analysis
	EnableFlowAnalysis bool

	// MinConfidenceThreshold sets minimum confidence for type annotations
	MinConfidenceThreshold float64

	// EnableVariablePropagation enables variable type propagation
	EnableVariablePropagation bool
}

// NewTypeInferenceEngine creates a new basic inference engine
func NewTypeInferenceEngine() TypeInferenceEngine {
	return &basicInferenceEngine{
		errors: make([]InferenceError, 0),
		options: InferenceOptions{
			EnableFlowAnalysis:        false, // Start simple
			MinConfidenceThreshold:    0.5,
			EnableVariablePropagation: true,
		},
	}
}

// NewTypeInferenceEngineWithOptions creates an engine with custom options
func NewTypeInferenceEngineWithOptions(options InferenceOptions) TypeInferenceEngine {
	return &basicInferenceEngine{
		errors:  make([]InferenceError, 0),
		options: options,
	}
}

// InferTypes performs type inference on the given AST
func (bie *basicInferenceEngine) InferTypes(inputAST *ast.AST) (ast.InferredAST, error) {
	// Clear previous errors
	bie.ClearErrors()

	// Create enhanced AST
	inferredAST := ast.NewInferredAST(inputAST)

	// Create traverser for systematic analysis
	traverser := NewASTTraverser(bie)

	// Perform AST traversal and type inference
	if err := traverser.TraverseAndInfer(inputAST, inferredAST); err != nil {
		return inferredAST, err
	}
	
	return inferredAST, nil
}

// CalculateConfidence calculates confidence score based on inference source
func (bie *basicInferenceEngine) CalculateConfidence(source types.TypeSource, context map[string]interface{}) float64 {
	// Base confidence scores by source type
	baseConfidences := map[types.TypeSource]float64{
		types.SourceLiteral:   0.95, // Very high confidence for literals
		types.SourceVariable:  0.85, // High confidence for variable declarations
		types.SourceParameter: 0.70, // Medium-high confidence for parameters
		types.SourceReturn:    0.75, // Medium-high confidence for returns
		types.SourceContext:   0.60, // Medium confidence for context inference
		types.SourceExternal:  0.90, // High confidence for external type info
	}

	baseConfidence, exists := baseConfidences[source]
	if !exists {
		return 0.50 // Default medium confidence
	}

	// TODO: Adjust confidence based on context
	// For example:
	// - Reduce confidence if there are conflicting type hints
	// - Increase confidence if multiple sources agree
	// - Consider code complexity and nesting levels

	return baseConfidence
}

// GetInferenceErrors returns collected errors
func (bie *basicInferenceEngine) GetInferenceErrors() []InferenceError {
	// Return a copy to prevent external modification
	result := make([]InferenceError, len(bie.errors))
	copy(result, bie.errors)
	return result
}

// AddInferenceError adds an error to the collection
func (bie *basicInferenceEngine) AddInferenceError(err InferenceError) {
	bie.errors = append(bie.errors, err)
}

// ClearErrors clears all collected errors
func (bie *basicInferenceEngine) ClearErrors() {
	bie.errors = make([]InferenceError, 0)
}

// InferenceError represents an error during type inference
type InferenceError struct {
	// NodeID identifies the AST node where the error occurred
	NodeID string

	// Message describes the error
	Message string

	// Source indicates what caused the error
	Source string

	// Confidence indicates how certain we are about this error
	Confidence float64
}

// NewInferenceError creates a new inference error
func NewInferenceError(nodeID, message string) InferenceError {
	return InferenceError{
		NodeID:     nodeID,
		Message:    message,
		Source:     "inference",
		Confidence: 1.0,
	}
}

// Error implements the error interface
func (ie InferenceError) Error() string {
	return ie.Message
}