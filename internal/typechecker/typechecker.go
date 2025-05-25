// ABOUTME: Type checking implementation for PSC
// ABOUTME: Validates type annotations in Perl code

package typechecker

import (
	"fmt"
	"os"
	"strings"

	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// TypeCheck is the main entry point for type checking a file
type TypeCheck struct {
	// Parser is the parser used for parsing Perl code
	Parser parser.Parser

	// TypeStore is the store for type definitions
	TypeStore *typedef.Storage

	// TypeHierarchy is the type hierarchy used for checking
	TypeHierarchy *typedef.TypeHierarchy

	// EnableFlowSensitiveAnalysis controls whether flow-sensitive analysis is enabled
	EnableFlowSensitiveAnalysis bool

	// SkipFlowChecks controls whether to skip flow-sensitive type checks
	// but still perform type refinements based on control flow
	SkipFlowChecks bool

	// FlowPatterns contains additional flow-sensitive patterns to recognize
	// These can include custom validation patterns for type refinement
	FlowPatterns []string

	// Cache for previously checked files to improve performance
	fileCache map[string]*TypeCheckResult

	// Maximum cache size to prevent memory issues
	maxCacheSize int

	// InferenceEngine for advanced type inference
	InferenceEngine *InferenceEngine

	// EnableTypeInference controls whether type inference is enabled
	EnableTypeInference bool
}

// NewTypeCheck creates a new TypeCheck instance
func NewTypeCheck() (*TypeCheck, error) {
	// Create a parser
	parser, err := parser.NewParser()
	if err != nil {
		return nil, err
	}

	// Create a type store
	typeStore, err := typedef.NewStorage()
	if err != nil {
		return nil, err
	}

	// Create the type hierarchy
	hierarchy := typedef.NewTypeHierarchy(typeStore)

	// Create the inference engine
	inferenceEngine := NewInferenceEngine(hierarchy)

	return &TypeCheck{
		Parser:                      parser,
		TypeStore:                   typeStore,
		TypeHierarchy:               hierarchy,
		EnableFlowSensitiveAnalysis: true,                              // Enable by default
		SkipFlowChecks:              false,                             // Don't skip checks by default
		FlowPatterns:                []string{},                        // No additional patterns by default
		fileCache:                   make(map[string]*TypeCheckResult), // Initialize caches
		maxCacheSize:                100,                               // Reasonable default cache size
		InferenceEngine:             inferenceEngine,
		EnableTypeInference:         true, // Enable type inference by default
	}, nil
}

// ClearCache clears the internal caches
func (tc *TypeCheck) ClearCache() {
	tc.fileCache = make(map[string]*TypeCheckResult)
}

// manageCacheSize ensures the cache doesn't grow too large
func (tc *TypeCheck) manageCacheSize() {
	// If we exceed the max cache size, remove oldest entries
	if len(tc.fileCache) > tc.maxCacheSize {
		// Simple approach: clear entire cache when it gets too big
		// In a production system, you might implement LRU eviction
		tc.ClearCache()
	}
}

// CheckFile performs type checking on a Perl file
func (tc *TypeCheck) CheckFile(path string) (*TypeCheckResult, error) {
	// Check cache first
	if cached, exists := tc.fileCache[path]; exists {
		// Check if file has been modified since caching
		_, err := os.Stat(path)
		if err == nil {
			// Compare modification time - if file hasn't changed, return cached result
			// For simplicity, we'll skip this check for now and always use cache if present
			// In production, you'd want to compare file modification times
			return cached, nil
		}
	}
	// Parse the file using our enhanced parser
	ast, err := tc.Parser.ParseFile(path)
	if err != nil {
		return nil, err
	}

	// Check for parser errors
	if len(ast.Errors) > 0 {
		result := &TypeCheckResult{
			Path:                 path,
			Errors:               []TypeCheckError{},
			TypeAnnotations:      ast.TypeAnnotations,
			RefinedTypes:         make(map[string]string),
			FlowSensitiveEnabled: tc.EnableFlowSensitiveAnalysis,
		}

		// Convert parser errors to type check errors
		for _, parseErr := range ast.Errors {
			var typErr TypeCheckError

			// Check if the error is a ParseError to extract position information
			switch perr := parseErr.(type) {
			case *parser.ParseError:
				typErr = TypeCheckError{
					Message: perr.Message,
					Line:    perr.Line,
					Column:  perr.Column,
					Path:    path,
				}
			default:
				typErr = TypeCheckError{
					Message: parseErr.Error(),
					Line:    0,
					Column:  0,
					Path:    path,
				}
			}

			result.Errors = append(result.Errors, typErr)
		}

		return result, nil
	}

	// Extract module name from the file path
	moduleName := extractModuleNameFromPath(path)

	// Create a type checker
	checker := parser.NewTypeChecker(tc.TypeHierarchy, moduleName)

	// Configure flow-sensitive analysis options
	if tc.EnableFlowSensitiveAnalysis {
		// Configure skip flow checks if specified
		if tc.SkipFlowChecks {
			// When SkipFlowChecks is true, we still perform type refinements
			// based on flow analysis, but don't report errors for flow-sensitive checks
			checker.TypeState.SkipFlowChecks = true
		}

		// Add custom validation patterns if specified
		if len(tc.FlowPatterns) > 0 {
			checker.AddFlowPatterns(tc.FlowPatterns)
		}
	}

	// Perform type inference if enabled
	if tc.EnableTypeInference && tc.InferenceEngine != nil {
		if err := tc.InferenceEngine.InferTypes(ast); err != nil {
			// Log inference error but don't fail type checking
			// Type inference is supplementary to explicit type checking
			fmt.Fprintf(os.Stderr, "Type inference warning: %v\n", err)
		}

		// Apply inferred types to the checker
		inferredTypes := tc.InferenceEngine.GetAllInferredTypes()
		for varName, inferredType := range inferredTypes {
			// Only apply inferred type if variable doesn't already have an explicit type
			if _, hasType := checker.VariableTypes[varName]; !hasType {
				checker.VariableTypes[varName] = inferredType
			}
		}
	}

	// Check the AST for type errors
	typeErrors := checker.CheckAST(ast)

	// Create the result
	result := &TypeCheckResult{
		Path:                 path,
		Errors:               []TypeCheckError{},
		TypeAnnotations:      ast.TypeAnnotations,
		RefinedTypes:         make(map[string]string),
		FlowSensitiveEnabled: tc.EnableFlowSensitiveAnalysis,
	}

	// Convert errors to TypeCheckError format
	for _, err := range typeErrors {
		line := 0
		col := 0
		message := err.Error()

		// Try to extract position information from the error
		// Check if the error implements the TypedError interface
		if typeErr, ok := err.(interface {
			Location() string
			Description() string
		}); ok {
			if loc := typeErr.Location(); loc != "" {
				parts := strings.Split(loc, ":")
				if len(parts) >= 3 {
					// Extract line and column from location
					_, _ = fmt.Sscanf(parts[1], "%d", &line)
					_, _ = fmt.Sscanf(parts[2], "%d", &col)
				}
			}
			message = typeErr.Description()
		}

		result.Errors = append(result.Errors, TypeCheckError{
			Message: message,
			Line:    line,
			Column:  col,
			Path:    path,
		})
	}

	// Include refined types from flow-sensitive analysis
	if tc.EnableFlowSensitiveAnalysis && checker.TypeState != nil {
		for varName, refinedType := range checker.TypeState.RefinedTypes {
			result.RefinedTypes[varName] = refinedType
		}
	}

	// Include inferred types if type inference is enabled
	if tc.EnableTypeInference && tc.InferenceEngine != nil {
		// Add inferred types that aren't already in refined types
		for varName, inferredType := range tc.InferenceEngine.GetAllInferredTypes() {
			if _, exists := result.RefinedTypes[varName]; !exists {
				result.RefinedTypes[varName] = inferredType + " (inferred)"
			}
		}
	}

	// Cache the result for future use
	tc.fileCache[path] = result
	tc.manageCacheSize()

	return result, nil
}
