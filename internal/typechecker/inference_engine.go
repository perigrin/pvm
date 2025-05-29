// ABOUTME: Advanced type inference engine for Perl code
// ABOUTME: Implements data flow analysis and contextual type inference

package typechecker

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/typedef"
)

// InferenceEngine performs advanced type inference on Perl code
type InferenceEngine struct {
	// DataFlowAnalyzer performs data flow analysis
	DataFlowAnalyzer *DataFlowAnalyzer

	// ContextAnalyzer handles context-sensitive type inference
	ContextAnalyzer *ContextAnalyzer

	// UsagePatternAnalyzer infers types from usage patterns
	UsagePatternAnalyzer *UsagePatternAnalyzer

	// TypePropagator propagates types through the code
	TypePropagator *TypePropagator

	// TypeHierarchy for type compatibility checks
	TypeHierarchy *typedef.TypeHierarchy

	// SymbolTable for symbol information
	SymbolTable *binder.SymbolTable

	// InferredTypes stores inferred type information
	InferredTypes map[string]*InferredTypeInfo

	// Confidence threshold for type inference (0.0 to 1.0)
	ConfidenceThreshold float64
}

// InferredTypeInfo contains inferred type information
type InferredTypeInfo struct {
	// Variable name or expression
	Name string

	// Inferred type
	Type string

	// Confidence level (0.0 to 1.0)
	Confidence float64

	// Sources of inference
	Sources []InferenceSource

	// Context where the type was inferred
	Context InferenceContext
}

// InferenceSource indicates how a type was inferred
type InferenceSource struct {
	// Type of inference source
	Type InferenceSourceType

	// Location in code
	Location ast.Position

	// Confidence contribution
	Confidence float64

	// Additional information
	Details string
}

// InferenceSourceType categorizes inference sources
type InferenceSourceType int

const (
	// InferenceFromAssignment infers from assignment statements
	InferenceFromAssignment InferenceSourceType = iota
	// InferenceFromUsage infers from how the variable is used
	InferenceFromUsage
	// InferenceFromMethodCall infers from method calls
	InferenceFromMethodCall
	// InferenceFromOperator infers from operators used
	InferenceFromOperator
	// InferenceFromContext infers from Perl context (scalar/list)
	InferenceFromContext
	// InferenceFromPattern infers from common patterns
	InferenceFromPattern
	// InferenceFromDataFlow infers from data flow analysis
	InferenceFromDataFlow
)

// InferenceContext represents the context of type inference
type InferenceContext struct {
	// Function or method name
	Function string

	// Package name
	Package string

	// Perl context (scalar, list, void)
	PerlContext string

	// Control flow context
	ControlFlow string
}

// DataFlowAnalyzer performs data flow analysis
type DataFlowAnalyzer struct {
	// DataFlowGraph represents the flow of data through the program
	DataFlowGraph *DataFlowGraph

	// VariableStates tracks variable states at different points
	VariableStates map[string]*VariableStateTracker
}

// DataFlowGraph represents data flow through the program
type DataFlowGraph struct {
	// Nodes represent program points
	Nodes map[string]*DataFlowNode

	// Edges represent data flow between nodes
	Edges map[string][]*DataFlowEdge
}

// DataFlowNode represents a point in the program
type DataFlowNode struct {
	// ID is the unique identifier
	ID string

	// Type of node (assignment, call, return, etc.)
	Type string

	// AST node reference
	ASTNode ast.Node

	// Variables defined at this point
	Definitions map[string]string

	// Variables used at this point
	Uses map[string]string
}

// DataFlowEdge represents data flow between nodes
type DataFlowEdge struct {
	// Source node
	From string

	// Target node
	To string

	// Type of flow (normal, conditional, loop)
	Type string

	// Condition for conditional flow
	Condition string
}

// VariableStateTracker tracks variable states through the program
type VariableStateTracker struct {
	// Variable name
	Name string

	// States at different program points
	States map[string]*VariableState

	// Type transformations
	Transformations []TypeTransformation
}

// VariableState represents a variable's state at a program point
type VariableState struct {
	// Possible types at this point
	PossibleTypes []string

	// Definite type if known
	DefiniteType string

	// Nullability
	MayBeNull bool

	// Initialization status
	Initialized bool
}

// TypeTransformation represents a type change
type TypeTransformation struct {
	// From type
	From string

	// To type
	To string

	// Location of transformation
	Location ast.Position

	// Reason for transformation
	Reason string
}

// ContextAnalyzer handles context-sensitive type inference
type ContextAnalyzer struct {
	// ContextStack tracks nested contexts
	ContextStack []PerlContext

	// ContextRules define type inference rules for contexts
	ContextRules map[string]ContextRule
}

// PerlContext represents Perl's context system
type PerlContext struct {
	// Type of context (scalar, list, void)
	Type string

	// Expected type in this context
	ExpectedType string

	// Parent context
	Parent *PerlContext
}

// ContextRule defines how types are inferred in specific contexts
type ContextRule struct {
	// Context pattern
	Pattern string

	// Type inference function
	InferType func(expr ast.Node, context PerlContext) string

	// Confidence for this rule
	Confidence float64
}

// UsagePatternAnalyzer infers types from usage patterns
type UsagePatternAnalyzer struct {
	// Patterns for type inference
	Patterns []UsagePattern

	// PatternCache caches pattern matches
	PatternCache map[string]*PatternMatch
}

// UsagePattern represents a code pattern that implies types
type UsagePattern struct {
	// Name of the pattern
	Name string

	// Pattern matcher
	Matcher func(node ast.Node) bool

	// Type inference from pattern
	InferType func(node ast.Node) (string, float64)

	// Priority for pattern matching
	Priority int
}

// PatternMatch represents a matched pattern
type PatternMatch struct {
	// Pattern that matched
	Pattern *UsagePattern

	// Inferred type
	Type string

	// Confidence level
	Confidence float64

	// Match details
	Details map[string]interface{}
}

// TypePropagator propagates types through the code
type TypePropagator struct {
	// PropagationRules define how types propagate
	PropagationRules []PropagationRule

	// TypeConstraints tracks type constraints
	TypeConstraints map[string][]TypeConstraint

	// SolvedTypes contains solved type variables
	SolvedTypes map[string]string
}

// PropagationRule defines how types propagate
type PropagationRule struct {
	// Name of the rule
	Name string

	// Condition for applying the rule
	Condition func(node ast.Node) bool

	// Propagation function
	Propagate func(node ast.Node, types map[string]string) map[string]string
}

// TypeConstraint represents a constraint on types
type TypeConstraint struct {
	// Variable under constraint
	Variable string

	// Constraint type (subtype, equality, etc.)
	Type string

	// Constraint value
	Value string

	// Source of constraint
	Source ast.Position
}

// NewInferenceEngine creates a new inference engine with symbol table support
func NewInferenceEngine(hierarchy *typedef.TypeHierarchy, symbolTable *binder.SymbolTable) *InferenceEngine {
	engine := &InferenceEngine{
		DataFlowAnalyzer:     NewDataFlowAnalyzer(),
		ContextAnalyzer:      NewContextAnalyzer(),
		UsagePatternAnalyzer: NewUsagePatternAnalyzer(),
		TypePropagator:       NewTypePropagator(),
		TypeHierarchy:        hierarchy,
		SymbolTable:          symbolTable,
		InferredTypes:        make(map[string]*InferredTypeInfo),
		ConfidenceThreshold:  0.7, // 70% confidence threshold
	}

	// Initialize analyzers with rules and patterns
	engine.initializeAnalyzers()

	return engine
}

// NewDataFlowAnalyzer creates a new data flow analyzer
func NewDataFlowAnalyzer() *DataFlowAnalyzer {
	return &DataFlowAnalyzer{
		DataFlowGraph: &DataFlowGraph{
			Nodes: make(map[string]*DataFlowNode),
			Edges: make(map[string][]*DataFlowEdge),
		},
		VariableStates: make(map[string]*VariableStateTracker),
	}
}

// NewContextAnalyzer creates a new context analyzer
func NewContextAnalyzer() *ContextAnalyzer {
	return &ContextAnalyzer{
		ContextStack: []PerlContext{},
		ContextRules: make(map[string]ContextRule),
	}
}

// NewUsagePatternAnalyzer creates a new usage pattern analyzer
func NewUsagePatternAnalyzer() *UsagePatternAnalyzer {
	return &UsagePatternAnalyzer{
		Patterns:     []UsagePattern{},
		PatternCache: make(map[string]*PatternMatch),
	}
}

// NewTypePropagator creates a new type propagator
func NewTypePropagator() *TypePropagator {
	return &TypePropagator{
		PropagationRules: []PropagationRule{},
		TypeConstraints:  make(map[string][]TypeConstraint),
		SolvedTypes:      make(map[string]string),
	}
}

// initializeAnalyzers sets up rules and patterns
func (ie *InferenceEngine) initializeAnalyzers() {
	// Initialize context rules
	ie.initializeContextRules()

	// Initialize usage patterns
	ie.initializeUsagePatterns()

	// Initialize propagation rules
	ie.initializePropagationRules()
}

// initializeContextRules sets up context-based inference rules
func (ie *InferenceEngine) initializeContextRules() {
	// Scalar context rule
	ie.ContextAnalyzer.ContextRules["scalar"] = ContextRule{
		Pattern: "scalar",
		InferType: func(expr ast.Node, context PerlContext) string {
			// In scalar context, arrays become counts
			exprType := ie.getNodeType(expr)
			nodeType := expr.Type()
			if strings.HasPrefix(exprType, "Array") || nodeType == "array" {
				return "Int"
			}
			return "Scalar"
		},
		Confidence: 0.8,
	}

	// List context rule
	ie.ContextAnalyzer.ContextRules["list"] = ContextRule{
		Pattern: "list",
		InferType: func(expr ast.Node, context PerlContext) string {
			// In list context, scalars become single-element lists
			exprType := ie.getNodeType(expr)
			nodeType := expr.Type()
			if !strings.Contains(exprType, "Array") && nodeType != "array" {
				return "Array"
			}
			return "Array"
		},
		Confidence: 0.8,
	}

	// Numeric context rule
	ie.ContextAnalyzer.ContextRules["numeric"] = ContextRule{
		Pattern: "numeric",
		InferType: func(expr ast.Node, context PerlContext) string {
			return "Num"
		},
		Confidence: 0.9,
	}

	// String context rule
	ie.ContextAnalyzer.ContextRules["string"] = ContextRule{
		Pattern: "string",
		InferType: func(expr ast.Node, context PerlContext) string {
			return "Str"
		},
		Confidence: 0.9,
	}
}

// initializeUsagePatterns sets up usage-based inference patterns
func (ie *InferenceEngine) initializeUsagePatterns() {
	// Pattern: Variable used as object (method calls) - HIGHEST PRIORITY
	ie.UsagePatternAnalyzer.Patterns = append(ie.UsagePatternAnalyzer.Patterns, UsagePattern{
		Name: "object_method_calls",
		Matcher: func(node ast.Node) bool {
			text := node.Text()
			return strings.Contains(text, "->") && !strings.Contains(text, "=>")
		},
		InferType: func(node ast.Node) (string, float64) {
			// Try to extract class name from method call or constructor
			text := node.Text()
			if strings.Contains(text, "->new") {
				// Look for Class->new pattern
				if idx := strings.Index(text, "->new"); idx > 0 {
					possibleClass := strings.TrimSpace(text[:idx])
					if strings.HasPrefix(possibleClass, "$") {
						return "Object", 0.7
					}
					return possibleClass, 0.9
				}
			}
			return "Object", 0.8
		},
		Priority: 15,
	})

	// Pattern: Variable used in string operations - HIGH PRIORITY
	ie.UsagePatternAnalyzer.Patterns = append(ie.UsagePatternAnalyzer.Patterns, UsagePattern{
		Name: "string_operations",
		Matcher: func(node ast.Node) bool {
			text := node.Text()
			return strings.Contains(text, "=~") || strings.Contains(text, "!~") ||
				strings.Contains(text, "substr") || strings.Contains(text, "length") ||
				strings.Contains(text, "uc") || strings.Contains(text, "lc")
		},
		InferType: func(node ast.Node) (string, float64) {
			return "Str", 0.8
		},
		Priority: 10,
	})

	// Pattern: Variable used with array operations
	ie.UsagePatternAnalyzer.Patterns = append(ie.UsagePatternAnalyzer.Patterns, UsagePattern{
		Name: "array_operations",
		Matcher: func(node ast.Node) bool {
			text := node.Text()
			return strings.Contains(text, "push") || strings.Contains(text, "pop") ||
				strings.Contains(text, "shift") || strings.Contains(text, "unshift") ||
				strings.Contains(text, "@{")
		},
		InferType: func(node ast.Node) (string, float64) {
			return "ArrayRef", 0.9
		},
		Priority: 10,
	})

	// Pattern: Variable used with hash operations
	ie.UsagePatternAnalyzer.Patterns = append(ie.UsagePatternAnalyzer.Patterns, UsagePattern{
		Name: "hash_operations",
		Matcher: func(node ast.Node) bool {
			text := node.Text()
			return strings.Contains(text, "keys") || strings.Contains(text, "values") ||
				strings.Contains(text, "exists") || strings.Contains(text, "%{")
		},
		InferType: func(node ast.Node) (string, float64) {
			return "HashRef", 0.9
		},
		Priority: 10,
	})

	// Pattern: Variable used in file operations
	ie.UsagePatternAnalyzer.Patterns = append(ie.UsagePatternAnalyzer.Patterns, UsagePattern{
		Name: "file_operations",
		Matcher: func(node ast.Node) bool {
			text := node.Text()
			return strings.Contains(text, "open") || strings.Contains(text, "close") ||
				strings.Contains(text, "print") && strings.Contains(text, "FILEHANDLE")
		},
		InferType: func(node ast.Node) (string, float64) {
			return "FileHandle", 0.8
		},
		Priority: 8,
	})

	// Pattern: Variable used in numeric operations - LOWER PRIORITY
	ie.UsagePatternAnalyzer.Patterns = append(ie.UsagePatternAnalyzer.Patterns, UsagePattern{
		Name: "numeric_operations",
		Matcher: func(node ast.Node) bool {
			text := node.Text()
			// Look for numeric operators, but avoid regex operators, method calls, and hash arrows
			if strings.Contains(text, "=~") || strings.Contains(text, "!~") ||
				strings.Contains(text, "->") || strings.Contains(text, "=>") {
				return false
			}
			// Look for numeric operators
			return strings.ContainsAny(text, "+-*/%") ||
				strings.Contains(text, "++") || strings.Contains(text, "--") ||
				strings.Contains(text, "**") || strings.Contains(text, "<=") ||
				strings.Contains(text, ">=") || strings.Contains(text, "<") ||
				strings.Contains(text, ">")
		},
		InferType: func(node ast.Node) (string, float64) {
			// Check if it looks like an integer
			if strings.Contains(node.Text(), ".") {
				return "Num", 0.8
			}
			return "Int", 0.8
		},
		Priority: 5,
	})
}

// initializePropagationRules sets up type propagation rules
func (ie *InferenceEngine) initializePropagationRules() {
	// Assignment propagation
	ie.TypePropagator.PropagationRules = append(ie.TypePropagator.PropagationRules, PropagationRule{
		Name: "assignment",
		Condition: func(node ast.Node) bool {
			return node.Type() == "assignment_expression" || strings.Contains(node.Text(), "=")
		},
		Propagate: func(node ast.Node, types map[string]string) map[string]string {
			// Propagate type from right to left in assignment
			result := make(map[string]string)
			for k, v := range types {
				result[k] = v
			}
			// Implementation would extract LHS and RHS and propagate type
			return result
		},
	})

	// Function parameter propagation
	ie.TypePropagator.PropagationRules = append(ie.TypePropagator.PropagationRules, PropagationRule{
		Name: "function_params",
		Condition: func(node ast.Node) bool {
			return node.Type() == "function_call" || node.Type() == "method_call"
		},
		Propagate: func(node ast.Node, types map[string]string) map[string]string {
			// Propagate types to function parameters
			result := make(map[string]string)
			for k, v := range types {
				result[k] = v
			}
			// Implementation would match arguments to parameters
			return result
		},
	})

	// Return type propagation
	ie.TypePropagator.PropagationRules = append(ie.TypePropagator.PropagationRules, PropagationRule{
		Name: "return_type",
		Condition: func(node ast.Node) bool {
			return node.Type() == "return_statement" || strings.HasPrefix(node.Text(), "return")
		},
		Propagate: func(node ast.Node, types map[string]string) map[string]string {
			// Propagate return expression type to function return type
			result := make(map[string]string)
			for k, v := range types {
				result[k] = v
			}
			// Implementation would track function return types
			return result
		},
	})
}

// InferTypes performs type inference on an AST
func (ie *InferenceEngine) InferTypes(ast *ast.AST) error {
	// Build data flow graph
	if err := ie.buildDataFlowGraph(ast); err != nil {
		return fmt.Errorf("failed to build data flow graph: %w", err)
	}

	// Analyze usage patterns
	ie.analyzeUsagePatterns(ast.Root)

	// Perform data flow analysis
	ie.performDataFlowAnalysis()

	// Apply context-sensitive inference
	ie.applyContextAnalysis(ast.Root)

	// Propagate types
	ie.propagateTypes()

	// Resolve type constraints
	ie.resolveTypeConstraints()

	return nil
}

// buildDataFlowGraph builds the data flow graph from AST
func (ie *InferenceEngine) buildDataFlowGraph(ast *ast.AST) error {
	// Implementation would traverse AST and build data flow graph
	// For now, return nil
	return nil
}

// analyzeUsagePatterns analyzes how variables are used
func (ie *InferenceEngine) analyzeUsagePatterns(node ast.Node) {
	// Check all patterns against this node
	for _, pattern := range ie.UsagePatternAnalyzer.Patterns {
		if pattern.Matcher(node) {
			varType, confidence := pattern.InferType(node)
			ie.recordInference(node.Text(), varType, confidence, InferenceFromPattern, node.Start())
		}
	}

	// Recursively analyze children
	for _, child := range node.Children() {
		ie.analyzeUsagePatterns(child)
	}
}

// performDataFlowAnalysis performs data flow analysis
func (ie *InferenceEngine) performDataFlowAnalysis() {
	// Implementation would analyze data flow graph
	// Track variable states through the program
}

// applyContextAnalysis applies context-sensitive type inference
func (ie *InferenceEngine) applyContextAnalysis(node ast.Node) {
	// Determine current context
	context := ie.determineContext(node)

	// Apply context rules
	for _, rule := range ie.ContextAnalyzer.ContextRules {
		if rule.Pattern == context.Type {
			inferredType := rule.InferType(node, context)
			ie.recordInference(node.Text(), inferredType, rule.Confidence, InferenceFromContext, node.Start())
		}
	}

	// Recursively analyze children
	for _, child := range node.Children() {
		ie.applyContextAnalysis(child)
	}
}

// propagateTypes propagates types through the program
func (ie *InferenceEngine) propagateTypes() {
	// Apply propagation rules
	for range ie.TypePropagator.PropagationRules {
		// Implementation would apply rules to propagate types
	}
}

// resolveTypeConstraints resolves type constraints to determine final types
func (ie *InferenceEngine) resolveTypeConstraints() {
	// Implementation would solve constraint system
	// For now, use simple resolution
	for varName, constraints := range ie.TypePropagator.TypeConstraints {
		if len(constraints) > 0 {
			// Take the most specific type from constraints
			ie.TypePropagator.SolvedTypes[varName] = constraints[0].Value
		}
	}
}

// recordInference records an inferred type
func (ie *InferenceEngine) recordInference(name, inferredType string, confidence float64, source InferenceSourceType, location ast.Position) {
	if existing, exists := ie.InferredTypes[name]; exists {
		// Update confidence if this inference is stronger
		if confidence > existing.Confidence {
			existing.Type = inferredType
			existing.Confidence = confidence
		}
		existing.Sources = append(existing.Sources, InferenceSource{
			Type:       source,
			Location:   location,
			Confidence: confidence,
		})
	} else {
		ie.InferredTypes[name] = &InferredTypeInfo{
			Name:       name,
			Type:       inferredType,
			Confidence: confidence,
			Sources: []InferenceSource{
				{
					Type:       source,
					Location:   location,
					Confidence: confidence,
				},
			},
		}
	}
}

// determineContext determines the Perl context for a node
func (ie *InferenceEngine) determineContext(node ast.Node) PerlContext {
	// Simple context determination based on parent nodes
	nodeType := node.Type()

	// Check for scalar context indicators
	if strings.Contains(nodeType, "scalar") || strings.Contains(node.Text(), "scalar") {
		return PerlContext{Type: "scalar"}
	}

	// Check for list context indicators
	if strings.Contains(nodeType, "list") || strings.Contains(nodeType, "array") {
		return PerlContext{Type: "list"}
	}

	// Check for numeric context
	if strings.ContainsAny(node.Text(), "+-*/%<>=") {
		return PerlContext{Type: "numeric"}
	}

	// Default to scalar context
	return PerlContext{Type: "scalar"}
}

// getNodeType gets the current type of a node (if known)
func (ie *InferenceEngine) getNodeType(node ast.Node) string {
	if info, exists := ie.InferredTypes[node.Text()]; exists {
		return info.Type
	}
	return "Any"
}

// GetInferredType returns the inferred type for a variable
func (ie *InferenceEngine) GetInferredType(varName string) (string, float64) {
	if info, exists := ie.InferredTypes[varName]; exists {
		if info.Confidence >= ie.ConfidenceThreshold {
			return info.Type, info.Confidence
		}
	}
	return "Any", 0.0
}

// GetAllInferredTypes returns all inferred types above confidence threshold
func (ie *InferenceEngine) GetAllInferredTypes() map[string]string {
	result := make(map[string]string)
	for name, info := range ie.InferredTypes {
		if info.Confidence >= ie.ConfidenceThreshold {
			result[name] = info.Type
		}
	}
	return result
}

// GetInferenceReport generates a detailed inference report
func (ie *InferenceEngine) GetInferenceReport() string {
	var report strings.Builder

	report.WriteString("Type Inference Report\n")
	report.WriteString("====================\n\n")

	for name, info := range ie.InferredTypes {
		report.WriteString(fmt.Sprintf("Variable: %s\n", name))
		report.WriteString(fmt.Sprintf("  Inferred Type: %s (confidence: %.2f)\n", info.Type, info.Confidence))
		report.WriteString("  Sources:\n")
		for _, source := range info.Sources {
			report.WriteString(fmt.Sprintf("    - %s at %d:%d (confidence: %.2f)\n",
				getSourceTypeName(source.Type), source.Location.Line, source.Location.Column, source.Confidence))
		}
		report.WriteString("\n")
	}

	return report.String()
}

// getSourceTypeName returns a readable name for inference source type
func getSourceTypeName(sourceType InferenceSourceType) string {
	switch sourceType {
	case InferenceFromAssignment:
		return "Assignment"
	case InferenceFromUsage:
		return "Usage"
	case InferenceFromMethodCall:
		return "Method Call"
	case InferenceFromOperator:
		return "Operator"
	case InferenceFromContext:
		return "Context"
	case InferenceFromPattern:
		return "Pattern"
	case InferenceFromDataFlow:
		return "Data Flow"
	default:
		return "Unknown"
	}
}
