// ABOUTME: Control flow analysis implementation for type inference
// ABOUTME: Tracks type propagation through conditionals, loops, and branches

package inference

import (
	"fmt"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/types"
)

// FlowAnalyzer handles control flow analysis for type inference
type FlowAnalyzer struct {
	// Reference to the inference engine
	engine TypeInferenceEngine

	// Stack of control flow states
	flowStack []*FlowState

	// Current flow state
	currentFlow *FlowState
}

// FlowState represents the type state at a point in control flow
type FlowState struct {
	// Variables and their types at this point
	VariableTypes map[string]*types.TypeInfo

	// Condition that led to this state (for conditional branches)
	Condition ast.Node

	// Branch type (if, elsif, else, loop, etc.)
	BranchType string

	// Parent flow state
	Parent *FlowState

	// Child flow states (for branches)
	Children []*FlowState

	// Confidence modifier based on control flow complexity
	ConfidenceModifier float64
}

// NewFlowAnalyzer creates a new control flow analyzer
func NewFlowAnalyzer(engine TypeInferenceEngine) *FlowAnalyzer {
	return &FlowAnalyzer{
		engine:    engine,
		flowStack: make([]*FlowState, 0),
	}
}

// NewFlowState creates a new flow state
func NewFlowState(branchType string, parent *FlowState) *FlowState {
	state := &FlowState{
		VariableTypes:      make(map[string]*types.TypeInfo),
		BranchType:         branchType,
		Parent:             parent,
		Children:           make([]*FlowState, 0),
		ConfidenceModifier: 1.0,
	}

	// Inherit variable types from parent if available
	if parent != nil {
		for name, typeInfo := range parent.VariableTypes {
			// Copy type info with slightly reduced confidence
			state.VariableTypes[name] = types.NewTypeInfo(
				typeInfo.Type,
				typeInfo.Confidence*0.95, // Reduce confidence through branches
				typeInfo.Source)
		}
	}

	return state
}

// EnterControlFlow enters a new control flow construct (if, loop, etc.)
func (fa *FlowAnalyzer) EnterControlFlow(branchType string, condition ast.Node) *FlowState {
	newState := NewFlowState(branchType, fa.currentFlow)
	newState.Condition = condition

	// Adjust confidence based on control flow complexity
	switch branchType {
	case "if":
		newState.ConfidenceModifier = 0.95
	case "elsif":
		newState.ConfidenceModifier = 0.90
	case "else":
		newState.ConfidenceModifier = 0.85
	case "while", "for", "foreach":
		newState.ConfidenceModifier = 0.80 // Loops can be complex
	case "given", "when":
		newState.ConfidenceModifier = 0.90
	default:
		newState.ConfidenceModifier = 0.90
	}

	// Add to parent's children if parent exists
	if fa.currentFlow != nil {
		fa.currentFlow.Children = append(fa.currentFlow.Children, newState)
	}

	// Push onto stack and set as current
	fa.flowStack = append(fa.flowStack, newState)
	fa.currentFlow = newState

	return newState
}

// ExitControlFlow exits the current control flow construct
func (fa *FlowAnalyzer) ExitControlFlow() *FlowState {
	if len(fa.flowStack) == 0 {
		return nil
	}

	// Pop current state
	exitingState := fa.currentFlow
	fa.flowStack = fa.flowStack[:len(fa.flowStack)-1]

	// Set parent as current
	if len(fa.flowStack) > 0 {
		fa.currentFlow = fa.flowStack[len(fa.flowStack)-1]
	} else {
		fa.currentFlow = nil
	}

	return exitingState
}

// AnalyzeConditionalBranches analyzes type implications of conditional expressions
func (fa *FlowAnalyzer) AnalyzeConditionalBranches(conditionNode ast.Node, inferredAST ast.InferredAST) error {
	if conditionNode == nil {
		return nil
	}

	// Analyze the condition itself for type implications
	switch conditionNode.Type() {
	case "binary_expression":
		return fa.analyzeBinaryCondition(conditionNode, inferredAST)
	case "function_call":
		return fa.analyzeFunctionCallCondition(conditionNode, inferredAST)
	case "variable":
		return fa.analyzeVariableCondition(conditionNode, inferredAST)
	case "negation":
		return fa.analyzeNegationCondition(conditionNode, inferredAST)
	default:
		// For unknown condition types, just analyze children
		for _, child := range conditionNode.Children() {
			if err := fa.AnalyzeConditionalBranches(child, inferredAST); err != nil {
				return err
			}
		}
	}

	return nil
}

// analyzeBinaryCondition analyzes binary expressions in conditions
func (fa *FlowAnalyzer) analyzeBinaryCondition(node ast.Node, inferredAST ast.InferredAST) error {
	children := node.Children()
	if len(children) < 3 {
		return nil // Invalid binary expression
	}

	left := children[0]
	operator := children[1]
	right := children[2]

	operatorText := operator.Text()

	// Type implications based on operator
	switch operatorText {
	case "==", "!=", "eq", "ne":
		// Equality comparisons imply compatible types
		return fa.propagateEqualityConstraint(left, right, inferredAST)
	case "<", ">", "<=", ">=", "lt", "gt", "le", "ge":
		// Comparison operators imply numeric or string context
		return fa.propagateComparisonConstraint(left, right, operatorText, inferredAST)
	case "=~", "!~":
		// Regex operators imply string context
		return fa.propagateRegexConstraint(left, right, inferredAST)
	case "&&", "||", "and", "or":
		// Logical operators - analyze both branches
		if err := fa.AnalyzeConditionalBranches(left, inferredAST); err != nil {
			return err
		}
		return fa.AnalyzeConditionalBranches(right, inferredAST)
	}

	return nil
}

// analyzeFunctionCallCondition analyzes function calls in conditions
func (fa *FlowAnalyzer) analyzeFunctionCallCondition(node ast.Node, inferredAST ast.InferredAST) error {
	children := node.Children()
	if len(children) == 0 {
		return nil
	}

	functionName := children[0].Text()

	// Type implications for common Perl functions
	switch functionName {
	case "defined":
		// defined($var) implies $var might be undef
		if len(children) > 1 {
			return fa.propagateDefinednessConstraint(children[1], true, inferredAST)
		}
	case "ref":
		// ref($var) implies $var is a reference
		if len(children) > 1 {
			return fa.propagateReferenceConstraint(children[1], inferredAST)
		}
	case "length", "substr", "index":
		// String functions imply string context
		if len(children) > 1 {
			return fa.propagateStringConstraint(children[1], inferredAST)
		}
	case "exists":
		// exists implies hash or array context
		if len(children) > 1 {
			return fa.propagateExistsConstraint(children[1], inferredAST)
		}
	}

	return nil
}

// analyzeVariableCondition analyzes variables used as conditions
func (fa *FlowAnalyzer) analyzeVariableCondition(node ast.Node, inferredAST ast.InferredAST) error {
	// Variables used as conditions are tested for truthiness
	variableName := extractVariableName(node.Text())

	// In Perl, testing a variable for truthiness can tell us about its type
	// For example: if ($var) implies $var is defined and not 0 or ""

	nodeID := fmt.Sprintf("condition_variable_%s_%d_%d",
		variableName, node.Start().Line, node.Start().Column)

	// Store constraint information (this would be used in actual type resolution)
	fa.engine.AddInferenceError(NewInferenceError(nodeID,
		fmt.Sprintf("Variable %s used in condition context", variableName)))

	return nil
}

// analyzeNegationCondition analyzes negated conditions
func (fa *FlowAnalyzer) analyzeNegationCondition(node ast.Node, inferredAST ast.InferredAST) error {
	children := node.Children()
	if len(children) > 0 {
		// Analyze the negated expression
		return fa.AnalyzeConditionalBranches(children[0], inferredAST)
	}
	return nil
}

// Type constraint propagation methods

// propagateEqualityConstraint propagates type constraints from equality comparisons
func (fa *FlowAnalyzer) propagateEqualityConstraint(left, right ast.Node, inferredAST ast.InferredAST) error {
	// If one side is a literal, it constrains the other side's type
	if left.Type() == "literal" && right.Type() == "variable" {
		return fa.propagateLiteralToVariable(left, right, inferredAST)
	} else if right.Type() == "literal" && left.Type() == "variable" {
		return fa.propagateLiteralToVariable(right, left, inferredAST)
	}

	// Both variables - they should have compatible types
	// This would be handled by a more sophisticated constraint system
	return nil
}

// propagateComparisonConstraint propagates constraints from comparison operations
func (fa *FlowAnalyzer) propagateComparisonConstraint(left, right ast.Node, operator string, inferredAST ast.InferredAST) error {
	// Numeric comparisons imply numeric context
	if operator == "<" || operator == ">" || operator == "<=" || operator == ">=" {
		if err := fa.propagateNumericConstraint(left, inferredAST); err != nil {
			return err
		}
		return fa.propagateNumericConstraint(right, inferredAST)
	}

	// String comparisons imply string context
	if operator == "lt" || operator == "gt" || operator == "le" || operator == "ge" {
		if err := fa.propagateStringConstraint(left, inferredAST); err != nil {
			return err
		}
		return fa.propagateStringConstraint(right, inferredAST)
	}

	return nil
}

// propagateRegexConstraint propagates string constraints from regex operations
func (fa *FlowAnalyzer) propagateRegexConstraint(left, right ast.Node, inferredAST ast.InferredAST) error {
	// Left side of regex is typically a string
	return fa.propagateStringConstraint(left, inferredAST)
}

// Helper methods for constraint propagation

// propagateLiteralToVariable propagates type from literal to variable
func (fa *FlowAnalyzer) propagateLiteralToVariable(literal, variable ast.Node, inferredAST ast.InferredAST) error {
	literalValue := literal.Text()
	variableName := extractVariableName(variable.Text())

	// Create a literal inferrer to determine the literal's type
	literalInferrer := NewLiteralInferrer()
	literalTypeInfo := literalInferrer.InferLiteralType(literalValue)

	// Create constraint-based type info for the variable
	constraintTypeInfo := types.NewTypeInfo(
		literalTypeInfo.Type,
		literalTypeInfo.Confidence*0.85, // Reduce confidence for constraint propagation
		types.SourceContext)

	// Update variable type in current flow state
	if fa.currentFlow != nil {
		fa.currentFlow.VariableTypes[variableName] = constraintTypeInfo
	}

	return nil
}

// propagateNumericConstraint propagates numeric type constraint
func (fa *FlowAnalyzer) propagateNumericConstraint(node ast.Node, inferredAST ast.InferredAST) error {
	if node.Type() == "variable" {
		variableName := extractVariableName(node.Text())

		// Create numeric constraint
		numericTypeInfo := types.NewTypeInfo(
			types.NewNumType(), // Assuming NumType exists for numeric values
			0.70,               // Medium confidence for context-based inference
			types.SourceContext)

		// Update variable type in current flow state
		if fa.currentFlow != nil {
			fa.currentFlow.VariableTypes[variableName] = numericTypeInfo
		}
	}
	return nil
}

// propagateStringConstraint propagates string type constraint
func (fa *FlowAnalyzer) propagateStringConstraint(node ast.Node, inferredAST ast.InferredAST) error {
	if node.Type() == "variable" {
		variableName := extractVariableName(node.Text())

		// Create string constraint
		stringTypeInfo := types.NewTypeInfo(
			types.NewStrType(),
			0.70,
			types.SourceContext)

		// Update variable type in current flow state
		if fa.currentFlow != nil {
			fa.currentFlow.VariableTypes[variableName] = stringTypeInfo
		}
	}
	return nil
}

// propagateDefinednessConstraint propagates definedness constraints
func (fa *FlowAnalyzer) propagateDefinednessConstraint(node ast.Node, isDefined bool, inferredAST ast.InferredAST) error {
	if node.Type() == "variable" {
		variableName := extractVariableName(node.Text())

		// Create definedness constraint
		// This would integrate with a more sophisticated type system
		// that can represent Maybe types or Union types with Undef

		nodeID := fmt.Sprintf("defined_constraint_%s_%d_%d",
			variableName, node.Start().Line, node.Start().Column)

		fa.engine.AddInferenceError(NewInferenceError(nodeID,
			fmt.Sprintf("Variable %s definedness constraint: %v", variableName, isDefined)))
	}
	return nil
}

// propagateReferenceConstraint propagates reference type constraints
func (fa *FlowAnalyzer) propagateReferenceConstraint(node ast.Node, inferredAST ast.InferredAST) error {
	if node.Type() == "variable" {
		variableName := extractVariableName(node.Text())

		// Create reference constraint - would be a union of reference types
		refTypeInfo := types.NewTypeInfo(
			types.NewRefType(), // Assuming RefType exists for references
			0.75,
			types.SourceContext)

		// Update variable type in current flow state
		if fa.currentFlow != nil {
			fa.currentFlow.VariableTypes[variableName] = refTypeInfo
		}
	}
	return nil
}

// propagateExistsConstraint propagates constraints from exists() checks
func (fa *FlowAnalyzer) propagateExistsConstraint(node ast.Node, inferredAST ast.InferredAST) error {
	// exists() implies hash or array access
	// This would need more sophisticated analysis of the expression
	// For now, just record that an exists check occurred

	nodeID := fmt.Sprintf("exists_constraint_%d_%d",
		node.Start().Line, node.Start().Column)

	fa.engine.AddInferenceError(NewInferenceError(nodeID,
		"exists() constraint detected"))

	return nil
}

// GetCurrentFlowState returns the current flow state
func (fa *FlowAnalyzer) GetCurrentFlowState() *FlowState {
	return fa.currentFlow
}

// GetVariableTypeInFlow gets a variable's type considering current flow state
func (fa *FlowAnalyzer) GetVariableTypeInFlow(variableName string) *types.TypeInfo {
	if fa.currentFlow == nil {
		return nil
	}

	if typeInfo, exists := fa.currentFlow.VariableTypes[variableName]; exists {
		return typeInfo
	}

	return nil
}

// MergeFlowStates merges type information from multiple flow branches
func (fa *FlowAnalyzer) MergeFlowStates(states []*FlowState) *FlowState {
	if len(states) == 0 {
		return nil
	}

	if len(states) == 1 {
		return states[0]
	}

	// Create merged state
	merged := NewFlowState("merged", nil)

	// Collect all variable names across all states
	allVariables := make(map[string]bool)
	for _, state := range states {
		for varName := range state.VariableTypes {
			allVariables[varName] = true
		}
	}

	// For each variable, merge type information across branches
	for varName := range allVariables {
		mergedType := fa.mergeVariableTypes(varName, states)
		if mergedType != nil {
			merged.VariableTypes[varName] = mergedType
		}
	}

	// Reduce confidence due to branching complexity
	merged.ConfidenceModifier = 0.80

	return merged
}

// mergeVariableTypes merges type information for a single variable across flow states
func (fa *FlowAnalyzer) mergeVariableTypes(varName string, states []*FlowState) *types.TypeInfo {
	var typeInfos []*types.TypeInfo

	// Collect type information from all states
	for _, state := range states {
		if typeInfo, exists := state.VariableTypes[varName]; exists {
			typeInfos = append(typeInfos, typeInfo)
		}
	}

	if len(typeInfos) == 0 {
		return nil
	}

	if len(typeInfos) == 1 {
		// Only one type - use it but reduce confidence
		return types.NewTypeInfo(
			typeInfos[0].Type,
			typeInfos[0].Confidence*0.90,
			typeInfos[0].Source)
	}

	// Multiple types - need to create union or find common type
	// For now, just use the most confident type with reduced confidence
	bestType := typeInfos[0]
	for _, typeInfo := range typeInfos[1:] {
		if typeInfo.Confidence > bestType.Confidence {
			bestType = typeInfo
		}
	}

	// Significantly reduce confidence for merged types
	return types.NewTypeInfo(
		bestType.Type,
		bestType.Confidence*0.70,
		types.SourceContext)
}
