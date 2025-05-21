// ABOUTME: Interfaces between parser and type checker
// ABOUTME: Defines the shared types and interfaces between components

package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"

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

	// ContextSensitiveFunctions maps function names to their context-dependent return types
	ContextSensitiveFunctions map[string]map[string]string

	// TypeAliases maps alias names to their target types
	TypeAliases map[string]string

	// GenericFunctions maps function names to their generic signature information
	GenericFunctions map[string]*GenericFunctionSignature

	// ModuleTypes maps module names to their exported types
	ModuleTypes map[string]map[string]string

	// HigherKindedTypes maps type names to their higher-kinded definitions
	HigherKindedTypes map[string]*HigherKindedTypeDefinition

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

// GenericFunctionSignature represents a generic function signature
type GenericFunctionSignature struct {
	// TypeParameters lists the generic type parameters
	TypeParameters []string

	// ParameterTypes maps parameter names to their types (may include type parameters)
	ParameterTypes map[string]string

	// ReturnType is the return type (may include type parameters)
	ReturnType string

	// Constraints maps type parameters to their constraint types
	Constraints map[string][]string

	// IsMethod indicates if this is a method
	IsMethod bool
}

// HigherKindedTypeDefinition represents a higher-kinded type definition
type HigherKindedTypeDefinition struct {
	// Name is the name of the higher-kinded type
	Name string

	// TypeConstructors lists the type constructor parameters
	TypeConstructors []string

	// Definition is the type definition body
	Definition string
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
		Hierarchy:                 hierarchy,
		CurrentModule:             moduleName,
		ImportedModules:           make(map[string]bool),
		TypeAnnotations:           make(map[string]string),
		VariableTypes:             make(map[string]string),
		FunctionTypes:             make(map[string]*FunctionSignature),
		TypeState:                 initialState,
		TypeStateStack:            []*TypeState{},
		ValidationPatterns:        []ValidationPattern{},
		ContextSensitiveFunctions: make(map[string]map[string]string),
		TypeAliases:               make(map[string]string),
		GenericFunctions:          make(map[string]*GenericFunctionSignature),
		ModuleTypes:               make(map[string]map[string]string),
		HigherKindedTypes:         make(map[string]*HigherKindedTypeDefinition),
		Debug:                     false,
	}

	// Initialize validation patterns
	tc.initializeValidationPatterns()

	return tc
}

// CheckAST performs type checking on an entire AST
func (tc *TypeChecker) CheckAST(ast *AST) []error {
	var typeErrors []error

	// Extract information about imported modules
	tc.extractImports(ast)

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
	returnErrors := tc.checkASTFunctionReturns(ast)
	if len(returnErrors) > 0 {
		typeErrors = append(typeErrors, returnErrors...)
	}

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
			if node.Type() != "defined_expression" && !strings.Contains(node.Type(), "function_call") {
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

	// Add ref() check for reftype refinement
	tc.ValidationPatterns = append(tc.ValidationPatterns, ValidationPattern{
		Name:    "ref check",
		Pattern: "ref($var) eq 'TYPE'",
		RefinementFunc: func(varName, currentType string) string {
			// When checking ref type, we can refine Ref to a specific reference type
			if currentType == "Ref" || currentType == "Any" {
				// The specific type would be determined based on the ref type string
				// For now, we're just demonstrating with ArrayRef
				return "ArrayRef"
			}
			return currentType
		},
		Checker: func(node Node) (string, bool) {
			// This is a simplified check that would need to be enhanced in a real implementation
			nodeText := getNodeText(node)

			// Look for patterns like 'ref($var) eq "ARRAY"'
			if strings.Contains(nodeText, "ref(") &&
				(strings.Contains(nodeText, "'ARRAY'") || strings.Contains(nodeText, "\"ARRAY\"")) {
				// Extract the variable from ref($var)
				start := strings.Index(nodeText, "ref(") + 4
				end := strings.Index(nodeText[start:], ")")
				if end > 0 {
					varName := nodeText[start : start+end]
					varName = strings.TrimSpace(varName)
					if strings.HasPrefix(varName, "$") {
						return varName, true
					}
				}
			}

			return "", false
		},
	})

	// Add more patterns as needed
}

// AddFlowPatterns adds custom flow patterns for validation
func (tc *TypeChecker) AddFlowPatterns(patterns []string) {
	// This would be implemented to add custom validation patterns
	// For now, it's a placeholder for future implementation
}

// getNodeText is a helper function to extract text from a node
func getNodeText(node Node) string {
	// Use the Text() method from the Node interface
	return node.Text()
}

// ExtractTypeAndParams extracts the base type and parameters from a parameterized type
// e.g., "ArrayRef[Int]" -> "ArrayRef", ["Int"]
func ExtractTypeAndParams(paramType string) (string, []string) {
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
		switch {
		case c == '[':
			bracketCount++
		case c == ']':
			bracketCount--
		case c == ',' && bracketCount == 0:
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

// TypeCheckError provides detailed information about a type error
type TypeCheckError struct {
	// Message is the error message
	Message string

	// Path is the file path where the error occurred
	Path string

	// Line is the line number where the error occurred
	Line int

	// Column is the column number where the error occurred
	Column int
}

// Error implements the error interface
func (e TypeCheckError) Error() string {
	if e.Path != "" && e.Line > 0 {
		return fmt.Sprintf("%s:%d:%d: %s", e.Path, e.Line, e.Column, e.Message)
	}
	return e.Message
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
	if tc.EnableFlowSensitiveAnalysis {
		// Configure skip flow checks if specified
		// TODO: Fix SkipFlowChecks usage

		// Add custom validation patterns if specified
		if len(tc.FlowPatterns) > 0 {
			checker.AddFlowPatterns(tc.FlowPatterns)
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
	// For now, use a purely regex-based approach since tree-sitter
	// has trouble with complex parameterized types in certain constructs
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	result := regexCleanupTypeAnnotations(string(content))

	return result, nil
}

// GeneratePerlFromAST generates Perl code from an AST with optional type annotations
func GeneratePerlFromAST(ast *AST, includeTypes bool) (string, error) {
	return GenerateFromAST(ast, includeTypes)
}

// GenerateTypedPerl generates Perl code with type annotations from an AST
func GenerateTypedPerl(ast *AST) (string, error) {
	return GenerateFromAST(ast, true)
}

// GenerateCleanPerl generates Perl code without type annotations from an AST
func GenerateCleanPerl(ast *AST) (string, error) {
	return GenerateFromAST(ast, false)
}

// AddTypesToAST creates a type injector for adding type annotations to an AST
func AddTypesToAST(ast *AST) *TypeInjector {
	return NewTypeInjector(ast)
}

// RoundTripParse parses a file and returns both the AST and the regenerated source
func RoundTripParse(path string, includeTypes bool) (*AST, string, error) {
	// Parse the file
	parser, err := NewParser()
	if err != nil {
		return nil, "", err
	}

	ast, err := parser.ParseFile(path)
	if err != nil {
		return nil, "", err
	}

	// Generate source from AST
	generatedSource, err := GenerateFromAST(ast, includeTypes)
	if err != nil {
		return nil, "", err
	}

	return ast, generatedSource, nil
}

// ConvertUntypedToTyped takes untyped Perl and adds type annotations based on a type mapping
func ConvertUntypedToTyped(path string, typeMapping map[string]string) (string, error) {
	// Parse the untyped file
	parser, err := NewParser()
	if err != nil {
		return "", err
	}

	ast, err := parser.ParseFile(path)
	if err != nil {
		return "", err
	}

	// Create type injector and apply type mapping
	injector := NewTypeInjector(ast)
	err = injector.ApplyTypeMapping(typeMapping)
	if err != nil {
		return "", err
	}

	// Generate typed source
	return GenerateFromAST(ast, true)
}

// ValidateRoundTrip validates that a file can be round-tripped without data loss
func ValidateRoundTrip(path string) error {
	// Parse original file
	parser, err := NewParser()
	if err != nil {
		return err
	}

	originalAST, err := parser.ParseFile(path)
	if err != nil {
		return err
	}

	// Generate source from AST
	regeneratedSource, err := GenerateFromAST(originalAST, true)
	if err != nil {
		return err
	}

	// Parse the regenerated source
	regeneratedAST, err := parser.ParseString(regeneratedSource)
	if err != nil {
		return fmt.Errorf("failed to parse regenerated source: %v", err)
	}

	// Basic validation - check type annotation count
	if len(originalAST.TypeAnnotations) != len(regeneratedAST.TypeAnnotations) {
		return fmt.Errorf("type annotation count mismatch: original=%d, regenerated=%d",
			len(originalAST.TypeAnnotations), len(regeneratedAST.TypeAnnotations))
	}

	return nil
}

// regexCleanupTypeAnnotations removes type annotations using regex patterns
// This is a fallback for complex parameterized types that tree-sitter might miss
func regexCleanupTypeAnnotations(code string) string {
	// Process line by line for better control
	lines := strings.Split(code, "\n")
	for i, line := range lines {

		// Handle variable declarations
		// Pattern: my Type $var or my Complex[Type[Nested]] $var
		varPattern := regexp.MustCompile(`\b(my|our|state)\s+[A-Z][^$]+\s+(\$[a-zA-Z_][a-zA-Z0-9_]*)`)
		if varPattern.MatchString(line) {
			line = varPattern.ReplaceAllString(line, `$1 $2`)
		}

		// Handle function parameters
		// Pattern: sub name(Type $param) or sub name(Complex[Type] $param)
		funcPattern := regexp.MustCompile(`\bsub\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(([^)]*)\)`)
		if funcPattern.MatchString(line) {
			line = funcPattern.ReplaceAllStringFunc(line, func(match string) string {
				parts := funcPattern.FindStringSubmatch(match)
				if len(parts) != 3 {
					return match
				}

				funcName := parts[1]
				params := parts[2]

				// Clean parameters
				paramPattern := regexp.MustCompile(`[A-Z][^$]*\s+(\$[a-zA-Z_][a-zA-Z0-9_]*)`)
				cleanParams := paramPattern.ReplaceAllString(params, `$1`)

				return fmt.Sprintf("sub %s(%s)", funcName, cleanParams)
			})
		}

		// Handle for loops
		// Pattern: for my Type $var (@array)
		forPattern := regexp.MustCompile(`\bfor\s+my\s+[A-Z][^$]+\s+(\$[a-zA-Z_][a-zA-Z0-9_]*\s+\([^)]+\))`)
		if forPattern.MatchString(line) {
			line = forPattern.ReplaceAllString(line, `for my $1`)
		}

		// Clean up any remaining return type annotations
		// Pattern: -> Type or -> Complex[Type]
		returnTypePattern := regexp.MustCompile(`\s*->\s*[A-Z][a-zA-Z_:]*(?:\[[^\]]*\])*`)
		line = returnTypePattern.ReplaceAllString(line, "")

		lines[i] = line
	}

	return strings.Join(lines, "\n")
}
