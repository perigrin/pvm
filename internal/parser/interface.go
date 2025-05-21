// ABOUTME: Interfaces between parser and type checker
// ABOUTME: Defines the shared types and interfaces between components

package parser

import (
	"fmt"
	"os"
	"strings"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/typedef"
)

// PSC Error codes for type checking
const (
	ErrTypeAnnotationMismatch = "810" // Type annotation doesn't match expected type
	ErrTypeInferenceError     = "811" // Failed to infer type
	ErrTypeValidationError    = "812" // Type validation error
	ErrTypeAssignmentError    = "813" // Error in variable assignment
	ErrTypeFunctionError      = "814" // Error in function parameter or return type
	ErrTypeDeclarationError   = "815" // Error in type declaration
	ErrTypeIncompatibleError  = "816" // Incompatible types in expression
)

// TypeCheckResult represents the result of type checking a file
type TypeCheckResult struct {
	// Path is the path to the checked file
	Path string

	// Errors is a list of type errors found during checking
	Errors []TypeCheckError

	// TypeAnnotations is a list of type annotations found in the code
	TypeAnnotations []*TypeAnnotation

	// RefinedTypes maps variable names to their refined types from flow-sensitive analysis
	RefinedTypes map[string]string

	// FlowSensitiveEnabled indicates if flow-sensitive analysis was enabled for this check
	FlowSensitiveEnabled bool
}

// TypeCheck is the main entry point for type checking a file
type TypeCheck struct {
	// Parser is the parser used for parsing Perl code
	Parser Parser

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
}

// NewTypeCheck creates a new TypeCheck instance
func NewTypeCheck() (*TypeCheck, error) {
	// Create a parser
	parser, err := NewParser()
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

	return &TypeCheck{
		Parser:                      parser,
		TypeStore:                   typeStore,
		TypeHierarchy:               hierarchy,
		EnableFlowSensitiveAnalysis: true,       // Enable by default
		SkipFlowChecks:              false,      // Don't skip checks by default
		FlowPatterns:                []string{}, // No additional patterns by default
	}, nil
}

// TypeChecker performs type checking on parsed Perl code
type TypeChecker struct {
	// Hierarchy is the type hierarchy used for checking
	Hierarchy *typedef.TypeHierarchy

	// CurrentModule is the current module being checked
	CurrentModule string

	// ImportedModules tracks imported modules
	ImportedModules map[string]bool

	// TypeAnnotations tracks annotated types
	TypeAnnotations map[string]string

	// VariableTypes maps variable names to their types (from annotations or inference)
	VariableTypes map[string]string

	// FunctionTypes maps function names to their signature information
	FunctionTypes map[string]*FunctionSignature

	// TypeState tracks the current type state for flow-sensitive analysis
	TypeState *TypeState

	// TypeStateStack holds type states for different code paths
	TypeStateStack []*TypeState

	// ValidationPatterns holds recognized validation patterns
	ValidationPatterns []ValidationPattern

	// Debug enables debug mode
	Debug bool
}

// FunctionSignature represents the type signature of a function or method
type FunctionSignature struct {
	// ParameterTypes maps parameter names to their types
	ParameterTypes map[string]string

	// ReturnType is the return type of the function
	ReturnType string

	// IsMethod indicates if this is a method
	IsMethod bool
}

// TypeState represents the types of variables at a specific point in the code
// It is used for flow-sensitive analysis to track how types change based on control flow
type TypeState struct {
	// VariableTypes maps variable names to their types in this state
	VariableTypes map[string]string

	// RefinedTypes maps variable names to their refined types based on control flow
	RefinedTypes map[string]string

	// Conditions tracks the conditions that led to this state
	Conditions []Condition
}

// Condition represents a condition that affects type refinement
type Condition struct {
	// Variable is the variable being checked in the condition
	Variable string

	// Operator is the comparison operator used (==, !=, >, <, etc.)
	Operator string

	// Value is the value being compared against
	Value string

	// Negated indicates if the condition is negated
	Negated bool
}

// ValidationPattern represents a recognized pattern for type validation
type ValidationPattern struct {
	// Name is the name of the pattern (e.g., "defined check", "ref check")
	Name string

	// Pattern is a simplified representation of the pattern
	Pattern string

	// RefinementFunc is the function that refines the type
	RefinementFunc func(variable string, currentType string) string

	// Checker is a function that checks if code matches this pattern
	Checker func(node Node) (string, bool)
}

// NewTypeChecker creates a new TypeChecker with the given type hierarchy
func NewTypeChecker(hierarchy *typedef.TypeHierarchy, moduleName string) *TypeChecker {
	// Create initial type state
	initialState := &TypeState{
		VariableTypes: make(map[string]string),
		RefinedTypes:  make(map[string]string),
		Conditions:    []Condition{},
	}

	tc := &TypeChecker{
		Hierarchy:          hierarchy,
		CurrentModule:      moduleName,
		ImportedModules:    make(map[string]bool),
		TypeAnnotations:    make(map[string]string),
		VariableTypes:      make(map[string]string),
		FunctionTypes:      make(map[string]*FunctionSignature),
		TypeState:          initialState,
		TypeStateStack:     []*TypeState{},
		ValidationPatterns: []ValidationPattern{},
		Debug:              false,
	}

	// Initialize validation patterns
	tc.initializeValidationPatterns()

	return tc
}

// CheckAST performs type checking on an entire AST
func (tc *TypeChecker) CheckAST(ast *AST) []error {
	var typeErrors []error

	// Extract information about imported modules
	tc.ImportedModules = make(map[string]bool)
	// In a real implementation, this would extract imported modules from the AST

	// First pass: collect all type annotations without validating them yet
	for _, annotation := range ast.TypeAnnotations {
		if err := tc.collectTypeAnnotation(annotation); err != nil {
			typeErrors = append(typeErrors, err)
		}
	}

	// Second pass: validate all type annotations
	for _, annotation := range ast.TypeAnnotations {
		if err := tc.checkTypeAnnotation(annotation); err != nil {
			typeErrors = append(typeErrors, err)
		}
	}

	// Third pass: validate usage of types in code
	assignmentErrors := tc.CheckASTAssignments(ast)
	if len(assignmentErrors) > 0 {
		typeErrors = append(typeErrors, assignmentErrors...)
	}

	// Fourth pass: validate function return types
	// This is a placeholder for future implementation
	// returnErrors := tc.checkASTFunctionReturns(ast)
	// if len(returnErrors) > 0 {
	//     typeErrors = append(typeErrors, returnErrors...)
	// }

	// Finally, perform flow-sensitive type analysis if enabled
	if tc.TypeState != nil {
		flowErrors := tc.performFlowSensitiveAnalysis(ast)
		if len(flowErrors) > 0 {
			typeErrors = append(typeErrors, flowErrors...)
		}
	}

	return typeErrors
}

// initializeValidationPatterns sets up the recognized validation patterns
func (tc *TypeChecker) initializeValidationPatterns() {
	// Add basic validation patterns for common Perl idioms

	// defined() check for Maybe types
	tc.ValidationPatterns = append(tc.ValidationPatterns, ValidationPattern{
		Name:    "defined check",
		Pattern: "defined($var)",
		RefinementFunc: func(varName, currentType string) string {
			// For Maybe[T] types, refine to T
			if strings.HasPrefix(currentType, "Maybe[") {
				baseType, params := ExtractTypeAndParams(currentType)
				if baseType == "Maybe" && len(params) > 0 {
					return params[0]
				}
			}
			return currentType
		},
		Checker: func(node Node) (string, bool) {
			// Check if this is a defined() expression
			if node.Type() != "defined_expression" {
				return "", false
			}

			// Extract the variable name from defined($var)
			nodeText := getNodeText(node)
			if strings.HasPrefix(nodeText, "defined(") && strings.HasSuffix(nodeText, ")") {
				varName := strings.TrimPrefix(nodeText, "defined(")
				varName = strings.TrimSuffix(varName, ")")
				varName = strings.TrimSpace(varName)

				// Only handle variables
				if strings.HasPrefix(varName, "$") {
					return varName, true
				}
			}

			return "", false
		},
	})

	// Add more patterns as needed
}

// getNodeText is a helper function to extract text from a node
func getNodeText(node Node) string {
	// In a full implementation, this would extract the text from the node based on its position
	switch n := node.(type) {
	case *TreeSitterNode:
		return n.Text
	default:
		// For other node types, we would need to access the source code and extract the text
		// based on the node's position
		// This is a simplified implementation
		return ""
	}
}

// ExtractTypeAndParams extracts the base type and parameters from a parameterized type
// e.g., "ArrayRef[Int]" -> "ArrayRef", ["Int"]
// This function uses the implementation from typedef package to avoid duplication.
func ExtractTypeAndParams(paramType string) (string, []string) {
	// This implementation should delegate to the typedef package in a real implementation
	// For now, we'll maintain compatibility with the typedef implementation
	idx := strings.Index(paramType, "[")
	if idx < 0 {
		return paramType, nil
	}

	baseType := paramType[:idx]
	paramStr := paramType[idx+1 : len(paramType)-1] // Remove outer brackets

	// Split parameters by comma, handling nested brackets
	var params []string
	bracketCount := 0
	start := 0

	for i, c := range paramStr {
		if c == '[' {
			bracketCount++
		} else if c == ']' {
			bracketCount--
		} else if c == ',' && bracketCount == 0 {
			params = append(params, strings.TrimSpace(paramStr[start:i]))
			start = i + 1
		}
	}

	// Add the last parameter
	if start < len(paramStr) {
		params = append(params, strings.TrimSpace(paramStr[start:]))
	}

	return baseType, params
}

// TypeErrorInfo provides detailed information about a type error
type TypeErrorInfo struct {
	// Message is the error message
	Message string

	// Path is the file path where the error occurred
	Path string

	// Position is the position in the file where the error occurred
	Position Position

	// SourceType is the source type that caused the error (if applicable)
	SourceType string

	// TargetType is the target type that caused the error (if applicable)
	TargetType string

	// Item is the variable, function, or other item involved in the error
	Item string

	// Kind is the kind of type error
	Kind string
}

// CheckFile performs type checking on a Perl file
func (tc *TypeCheck) CheckFile(path string) (*TypeCheckResult, error) {
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
			case *ParseError:
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
	checker := NewTypeChecker(tc.TypeHierarchy, moduleName)

	// Configure flow-sensitive analysis options
	checker.Debug = tc.EnableFlowSensitiveAnalysis

	// If enabled, pass additional flow-sensitive analysis options
	if tc.EnableFlowSensitiveAnalysis {
		// Configure to skip flow checks if specified
		// (in a real implementation we would have a field for this in TypeChecker)
		// checker.SkipFlowChecks = tc.SkipFlowChecks

		// Add custom validation patterns if specified
		if len(tc.FlowPatterns) > 0 {
			// In a real implementation, we would parse and add these patterns
			// For now, we'll just log that we received them
			fmt.Printf("INFO: Using %d custom flow patterns\n", len(tc.FlowPatterns))
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

	return result, nil
}

// extractModuleNameFromPath extracts a module name from a file path
func extractModuleNameFromPath(path string) string {
	// This is a simplified implementation that would need to be enhanced in a real system
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}

	filename := parts[len(parts)-1]
	moduleName := strings.TrimSuffix(filename, ".pm")
	moduleName = strings.TrimSuffix(moduleName, ".pl")

	// Handle lib/Module/Name.pm style paths
	if len(parts) >= 3 {
		if parts[len(parts)-3] == "lib" {
			// This might be a module in a standard lib directory
			// Try to construct a module name from the path
			libIndex := -1
			for i, part := range parts {
				if part == "lib" {
					libIndex = i
					break
				}
			}

			if libIndex >= 0 && libIndex < len(parts)-1 {
				// Reconstruct the module name from the path components after "lib"
				moduleComponents := parts[libIndex+1 : len(parts)-1]
				moduleComponents = append(moduleComponents, moduleName)
				moduleName = strings.Join(moduleComponents, "::")
			}
		}
	}

	return moduleName
}

// StripAnnotations removes type annotations from Perl code
func StripAnnotations(path string) (string, error) {
	// Create a parser
	p, err := NewParser()
	if err != nil {
		return "", err
	}

	// Parse the file with our enhanced parser
	ast, err := p.ParseFile(path)
	if err != nil {
		return "", err
	}

	// Read the file content
	content, err := os.ReadFile(path)
	if err != nil {
		return "", errors.NewSystemError("001",
			"Failed to read file", err).
			WithLocation(path)
	}

	// Convert content to string
	originalCode := string(content)

	// Now that we have a more advanced parser, we can implement a more sophisticated
	// annotation stripping process. The approach is to rewrite the code line by line,
	// removing the type annotations where they are found.

	// We need to be careful to avoid changing line numbers, so we'll replace annotations
	// with spaces rather than removing them entirely.

	if len(ast.TypeAnnotations) == 0 {
		// No annotations to strip
		return originalCode, nil
	}

	// Split the code into lines for easier processing
	lines := strings.Split(originalCode, "\n")

	// Map annotations by line for efficient lookup
	annotationsByLine := make(map[int][]*TypeAnnotation)
	for _, ann := range ast.TypeAnnotations {
		lineNum := ann.Pos.Line
		annotationsByLine[lineNum] = append(annotationsByLine[lineNum], ann)
	}

	// Process each line that has annotations
	for lineNum, annotations := range annotationsByLine {
		if lineNum <= 0 || lineNum > len(lines) {
			continue
		}

		line := lines[lineNum-1]

		// Sort annotations in reverse order of column position
		// This ensures we process them from right to left
		// to avoid affecting the positions of subsequent annotations
		for i := 0; i < len(annotations); i++ {
			for j := i + 1; j < len(annotations); j++ {
				if annotations[i].Pos.Column < annotations[j].Pos.Column {
					annotations[i], annotations[j] = annotations[j], annotations[i]
				}
			}
		}

		// Process each annotation on this line
		for _, ann := range annotations {
			switch ann.Kind {
			case VarAnnotation:
				// Handle variable annotations - e.g., "my Type $var" -> "my $var"
				if typePos := strings.Index(line, ann.TypeExpression.String()); typePos >= 0 {
					typeLen := len(ann.TypeExpression.String())
					// Replace the type with spaces to maintain line structure
					spaces := strings.Repeat(" ", typeLen)
					line = line[:typePos] + spaces + line[typePos+typeLen:]
				}

			case SubParamAnnotation, MethodParamAnnotation:
				// Handle parameter annotations - e.g., "Type $param" -> "$param"
				if ann.AnnotatedItem != "" {
					paramPos := strings.Index(line, ann.AnnotatedItem)
					if paramPos > 0 {
						// Look for the type that appears before the parameter
						beforeParam := line[:paramPos]
						if typePos := strings.LastIndex(beforeParam, ann.TypeExpression.String()); typePos >= 0 {
							typeLen := len(ann.TypeExpression.String())
							// Replace the type with spaces to maintain line structure
							spaces := strings.Repeat(" ", typeLen)
							line = line[:typePos] + spaces + line[typePos+typeLen:]
						}
					}
				}

			case SubReturnAnnotation, MethodReturnAnnotation:
				// Handle return type annotations - e.g., "-> Type" -> "-> "
				arrowPos := strings.Index(line, "->")
				if arrowPos >= 0 {
					returnTypePos := arrowPos + 2 // Skip the "->"
					// Find the type after the arrow
					afterArrow := strings.TrimSpace(line[returnTypePos:])
					if strings.HasPrefix(afterArrow, ann.TypeExpression.String()) {
						typeLen := len(ann.TypeExpression.String())
						// Calculate the actual position of the type
						actualTypePos := returnTypePos + strings.Index(line[returnTypePos:], ann.TypeExpression.String())
						// Replace the type with spaces to maintain line structure
						spaces := strings.Repeat(" ", typeLen)
						line = line[:actualTypePos] + spaces + line[actualTypePos+typeLen:]
					}
				}

			case AttrAnnotation:
				// Handle field/attribute annotations - e.g., "field Type $attr" -> "field $attr"
				if typePos := strings.Index(line, ann.TypeExpression.String()); typePos >= 0 {
					typeLen := len(ann.TypeExpression.String())
					// Replace the type with spaces to maintain line structure
					spaces := strings.Repeat(" ", typeLen)
					line = line[:typePos] + spaces + line[typePos+typeLen:]
				}

			case TypeDeclAnnotation:
				// For type declarations, we might want to keep them as-is or
				// replace the entire line with a comment
				// For now, we'll keep them as they don't affect runtime behavior
			}
		}

		// Update the line in the array
		lines[lineNum-1] = line
	}

	// Combine the lines back into a single string
	strippedCode := strings.Join(lines, "\n")

	return strippedCode, nil
}
