// ABOUTME: Flow-sensitive type analysis for the typechecker
// ABOUTME: Handles control flow and type refinement based on conditions

package typechecker

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"tamarou.com/pvm/internal/ast"
)

// ControlFlowGraph represents the control flow graph of a program
type ControlFlowGraph struct {
	// Nodes represents the basic blocks in the control flow graph
	Nodes []*BasicBlock

	// Edges represents the control flow transitions between blocks
	Edges []*FlowEdge

	// Entry is the entry point of the control flow graph
	Entry *BasicBlock

	// Exit is the exit point of the control flow graph
	Exit *BasicBlock
}

// BasicBlock represents a straight-line sequence of statements
type BasicBlock struct {
	// ID is the unique identifier for this block
	ID int

	// Statements contains the AST nodes in this block
	Statements []ast.Node

	// Predecessors are the blocks that can flow into this block
	Predecessors []*BasicBlock

	// Successors are the blocks that this block can flow into
	Successors []*BasicBlock

	// TypeState is the type state at the beginning of this block
	TypeState *TypeState

	// ExitTypeState is the type state at the end of this block
	ExitTypeState *TypeState
}

// FlowEdge represents a control flow edge between basic blocks
type FlowEdge struct {
	// From is the source block
	From *BasicBlock

	// To is the destination block
	To *BasicBlock

	// Condition is the condition that must be true for this edge to be taken
	Condition *Condition

	// EdgeType describes the type of control flow edge
	EdgeType FlowEdgeType
}

// FlowEdgeType represents the type of control flow edge
type FlowEdgeType int

const (
	// UnconditionalEdge represents an unconditional flow
	UnconditionalEdge FlowEdgeType = iota

	// ConditionalTrueEdge represents the true branch of a conditional
	ConditionalTrueEdge

	// ConditionalFalseEdge represents the false branch of a conditional
	ConditionalFalseEdge

	// LoopBackEdge represents a loop back edge
	LoopBackEdge

	// BreakEdge represents a break/last statement
	BreakEdge

	// ContinueEdge represents a continue/next statement
	ContinueEdge

	// ExceptionEdge represents an exception flow
	ExceptionEdge
)

// FlowAnalyzer performs flow-sensitive type analysis
type FlowAnalyzer struct {
	// TypeChecker is the parent type checker
	TypeChecker *TypeChecker

	// CFG is the control flow graph being analyzed
	CFG *ControlFlowGraph

	// WorkList contains blocks that need to be processed
	WorkList []*BasicBlock

	// ProcessedBlocks tracks which blocks have been processed
	ProcessedBlocks map[int]bool

	// MaxIterations prevents infinite loops in analysis
	MaxIterations int

	// BuiltinTypes provides type signatures for Perl built-in functions
	BuiltinTypes *BuiltinTypeRegistry

	// OperatorTypes provides type signatures for Perl binary operators
	OperatorTypes *OperatorTypeRegistry

	// SourceCode contains the original source code for analysis
	SourceCode string

	// processVariableDeclarationHook for testing data flow (optional hook)
	processVariableDeclarationHook func(ast.Node)
}

// NewFlowAnalyzer creates a new flow analyzer
func NewFlowAnalyzer(tc *TypeChecker) *FlowAnalyzer {
	// Initialize built-in type registry
	builtinTypes, err := NewBuiltinTypeRegistry()
	if err != nil {
		// Log error but don't fail - fall back to hardcoded types
		// In production, you might want to handle this differently
		builtinTypes = nil
	}

	// Initialize operator type registry
	operatorTypes, err := NewOperatorTypeRegistry()
	if err != nil {
		// Log error but don't fail - fall back to hardcoded operator logic
		operatorTypes = nil
	}

	return &FlowAnalyzer{
		TypeChecker:     tc,
		ProcessedBlocks: make(map[int]bool),
		MaxIterations:   100, // Reasonable default to prevent infinite loops
		BuiltinTypes:    builtinTypes,
		OperatorTypes:   operatorTypes,
	}
}

// performFlowSensitiveAnalysis performs flow-sensitive type analysis
func (tc *TypeChecker) performFlowSensitiveAnalysis(ast *ast.AST) []error {
	var errors []error

	// Skip if flow analysis is disabled
	if tc.TypeState == nil {
		return errors
	}

	// Create flow analyzer
	analyzer := NewFlowAnalyzer(tc)

	// Build control flow graph
	cfg, err := analyzer.buildControlFlowGraph(ast)
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	// Perform data flow analysis
	flowErrors := analyzer.analyzeDataFlow(cfg)
	if len(flowErrors) > 0 {
		errors = append(errors, flowErrors...)
	}

	return errors
}

// buildControlFlowGraph constructs a control flow graph from the AST
func (fa *FlowAnalyzer) buildControlFlowGraph(ast *ast.AST) (*ControlFlowGraph, error) {
	if ast.Root == nil {
		return nil, fmt.Errorf("AST has no root node")
	}

	cfg := &ControlFlowGraph{
		Nodes: []*BasicBlock{},
		Edges: []*FlowEdge{},
	}

	// Create entry block with type state populated from symbol table
	initialTypeState := fa.createInitialTypeState()
	entry := &BasicBlock{
		ID:            0,
		Statements:    nil,
		Predecessors:  []*BasicBlock{},
		Successors:    []*BasicBlock{},
		TypeState:     initialTypeState,
		ExitTypeState: fa.copyTypeState(initialTypeState),
	}
	cfg.Entry = entry
	cfg.Nodes = append(cfg.Nodes, entry)

	// Build CFG by traversing the AST
	currentBlock := entry
	blockID := 1

	for _, child := range ast.Root.Children() {
		newBlock, newID, err := fa.processNode(child, currentBlock, blockID, cfg)
		if err != nil {
			return nil, err
		}
		currentBlock = newBlock
		blockID = newID
	}

	// Create exit block if needed
	if currentBlock != nil {
		exit := &BasicBlock{
			ID:            blockID,
			Statements:    nil,
			Predecessors:  []*BasicBlock{currentBlock},
			Successors:    []*BasicBlock{},
			TypeState:     fa.copyTypeState(currentBlock.ExitTypeState),
			ExitTypeState: fa.copyTypeState(currentBlock.ExitTypeState),
		}
		cfg.Exit = exit
		cfg.Nodes = append(cfg.Nodes, exit)
		currentBlock.Successors = append(currentBlock.Successors, exit)
		cfg.Edges = append(cfg.Edges, &FlowEdge{
			From:     currentBlock,
			To:       exit,
			EdgeType: UnconditionalEdge,
		})
	}

	fa.CFG = cfg
	return cfg, nil
}

// createInitialTypeState creates the initial type state for flow analysis
// by populating it with variable types from the TypeChecker's symbol table
func (fa *FlowAnalyzer) createInitialTypeState() *TypeState {
	state := &TypeState{
		VariableTypes: make(map[string]string),
		RefinedTypes:  make(map[string]string),
		Conditions:    []Condition{},
	}

	// Populate from TypeChecker's symbol table information
	if fa.TypeChecker != nil {
		// Copy all variable types from TypeChecker
		for varName, varType := range fa.TypeChecker.VariableTypes {
			state.VariableTypes[varName] = varType
		}

		// Also include type annotations that might not be in VariableTypes
		for varName, varType := range fa.TypeChecker.TypeAnnotations {
			if _, exists := state.VariableTypes[varName]; !exists {
				state.VariableTypes[varName] = varType
			}
		}

		// Preserve any existing flow-sensitive state if it exists
		if fa.TypeChecker.TypeState != nil {
			// Copy refined types from existing flow state
			for varName, refinedType := range fa.TypeChecker.TypeState.RefinedTypes {
				state.RefinedTypes[varName] = refinedType
			}
			// Copy conditions from existing flow state
			state.Conditions = append(state.Conditions, fa.TypeChecker.TypeState.Conditions...)
		}
	}

	return state
}

// processNode processes a single AST node and updates the control flow graph
func (fa *FlowAnalyzer) processNode(node ast.Node, currentBlock *BasicBlock, blockID int, cfg *ControlFlowGraph) (*BasicBlock, int, error) {
	if node == nil {
		return currentBlock, blockID, nil
	}

	// Handle different node types
	switch node.Type() {
	case "if_statement", "unless_statement", "if_stmt":
		return fa.processConditional(node, currentBlock, blockID, cfg)
	case "while_statement", "until_statement", "for_statement", "foreach_statement", "for_stmt", "while_stmt":
		return fa.processLoop(node, currentBlock, blockID, cfg)
	case "given_statement":
		return fa.processGivenWhen(node, currentBlock, blockID, cfg)
	case "sub_decl", "method_decl":
		return fa.processSubroutine(node, currentBlock, blockID, cfg)
	default:
		// Regular statement - add to current block
		currentBlock.Statements = append(currentBlock.Statements, node)
		// Update type state based on this statement
		fa.updateTypeStateForStatement(node, currentBlock)
		return currentBlock, blockID, nil
	}
}

// processSubroutine handles subroutine declarations by processing their body statements
func (fa *FlowAnalyzer) processSubroutine(node ast.Node, currentBlock *BasicBlock, blockID int, cfg *ControlFlowGraph) (*BasicBlock, int, error) {
	// For subroutines, we need to process the statements in the body
	// First, find the body (typically a block_stmt child)
	for _, child := range node.Children() {
		if child.Type() == "block_stmt" {
			// Process each statement in the block
			for _, stmt := range child.Children() {
				if stmt.Type() == "token" {
					// Skip tokens like { and }
					continue
				}
				newBlock, newID, err := fa.processNode(stmt, currentBlock, blockID, cfg)
				if err != nil {
					return nil, 0, err
				}
				currentBlock = newBlock
				blockID = newID
			}
			break
		}
	}
	return currentBlock, blockID, nil
}

// processConditional handles if/unless statements
func (fa *FlowAnalyzer) processConditional(node ast.Node, currentBlock *BasicBlock, blockID int, cfg *ControlFlowGraph) (*BasicBlock, int, error) {
	// Create condition evaluation block
	condBlock := &BasicBlock{
		ID:            blockID,
		Statements:    []ast.Node{node},
		Predecessors:  []*BasicBlock{currentBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(currentBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(currentBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, condBlock)
	blockID++

	// Connect current block to condition block
	currentBlock.Successors = append(currentBlock.Successors, condBlock)
	cfg.Edges = append(cfg.Edges, &FlowEdge{
		From:     currentBlock,
		To:       condBlock,
		EdgeType: UnconditionalEdge,
	})

	// Create true branch block
	trueBlock := &BasicBlock{
		ID:            blockID,
		Statements:    nil,
		Predecessors:  []*BasicBlock{condBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(condBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(condBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, trueBlock)
	blockID++

	// Apply type refinement for true branch
	fa.refineTypeForCondition(node, trueBlock, true)

	// Create false branch block
	falseBlock := &BasicBlock{
		ID:            blockID,
		Statements:    nil,
		Predecessors:  []*BasicBlock{condBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(condBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(condBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, falseBlock)
	blockID++

	// Apply type refinement for false branch
	fa.refineTypeForCondition(node, falseBlock, false)

	// Connect condition to branches
	condBlock.Successors = append(condBlock.Successors, trueBlock, falseBlock)
	cfg.Edges = append(cfg.Edges,
		&FlowEdge{
			From:     condBlock,
			To:       trueBlock,
			EdgeType: ConditionalTrueEdge,
		},
		&FlowEdge{
			From:     condBlock,
			To:       falseBlock,
			EdgeType: ConditionalFalseEdge,
		},
	)

	// Create merge block for after conditional
	mergeBlock := &BasicBlock{
		ID:            blockID,
		Statements:    nil,
		Predecessors:  []*BasicBlock{trueBlock, falseBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.mergeTypeStates(trueBlock.ExitTypeState, falseBlock.ExitTypeState),
		ExitTypeState: fa.mergeTypeStates(trueBlock.ExitTypeState, falseBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, mergeBlock)
	blockID++

	// Connect branches to merge
	trueBlock.Successors = append(trueBlock.Successors, mergeBlock)
	falseBlock.Successors = append(falseBlock.Successors, mergeBlock)
	cfg.Edges = append(cfg.Edges,
		&FlowEdge{
			From:     trueBlock,
			To:       mergeBlock,
			EdgeType: UnconditionalEdge,
		},
		&FlowEdge{
			From:     falseBlock,
			To:       mergeBlock,
			EdgeType: UnconditionalEdge,
		},
	)

	return mergeBlock, blockID, nil
}

// processLoop handles loop statements
func (fa *FlowAnalyzer) processLoop(node ast.Node, currentBlock *BasicBlock, blockID int, cfg *ControlFlowGraph) (*BasicBlock, int, error) {
	// Create loop header block
	headerBlock := &BasicBlock{
		ID:            blockID,
		Statements:    []ast.Node{node},
		Predecessors:  []*BasicBlock{currentBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(currentBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(currentBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, headerBlock)
	blockID++

	// Connect current to header
	currentBlock.Successors = append(currentBlock.Successors, headerBlock)
	cfg.Edges = append(cfg.Edges, &FlowEdge{
		From:     currentBlock,
		To:       headerBlock,
		EdgeType: UnconditionalEdge,
	})

	// Create loop body block
	bodyBlock := &BasicBlock{
		ID:            blockID,
		Statements:    nil,
		Predecessors:  []*BasicBlock{headerBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(headerBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(headerBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, bodyBlock)
	blockID++

	// Create loop exit block
	exitBlock := &BasicBlock{
		ID:            blockID,
		Statements:    nil,
		Predecessors:  []*BasicBlock{headerBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(headerBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(headerBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, exitBlock)
	blockID++

	// Connect header to body and exit
	headerBlock.Successors = append(headerBlock.Successors, bodyBlock, exitBlock)
	cfg.Edges = append(cfg.Edges,
		&FlowEdge{
			From:     headerBlock,
			To:       bodyBlock,
			EdgeType: ConditionalTrueEdge,
		},
		&FlowEdge{
			From:     headerBlock,
			To:       exitBlock,
			EdgeType: ConditionalFalseEdge,
		},
	)

	// Connect body back to header (loop back edge)
	bodyBlock.Successors = append(bodyBlock.Successors, headerBlock)
	headerBlock.Predecessors = append(headerBlock.Predecessors, bodyBlock)
	cfg.Edges = append(cfg.Edges, &FlowEdge{
		From:     bodyBlock,
		To:       headerBlock,
		EdgeType: LoopBackEdge,
	})

	return exitBlock, blockID, nil
}

// processGivenWhen handles given/when statements
func (fa *FlowAnalyzer) processGivenWhen(node ast.Node, currentBlock *BasicBlock, blockID int, cfg *ControlFlowGraph) (*BasicBlock, int, error) {
	// For now, treat given/when like a simple conditional
	// In a full implementation, we'd handle multiple when clauses
	return fa.processConditional(node, currentBlock, blockID, cfg)
}

// analyzeDataFlow performs data flow analysis on the control flow graph
func (fa *FlowAnalyzer) analyzeDataFlow(cfg *ControlFlowGraph) []error {
	var errors []error

	// Initialize worklist with entry block
	fa.WorkList = []*BasicBlock{cfg.Entry}
	iterations := 0

	// Process blocks until fixpoint is reached
	for len(fa.WorkList) > 0 && iterations < fa.MaxIterations {
		block := fa.WorkList[0]
		fa.WorkList = fa.WorkList[1:]

		// Skip if already processed in this iteration
		if fa.ProcessedBlocks[block.ID] {
			continue
		}

		// Process the block
		blockErrors := fa.processBlock(block)
		if len(blockErrors) > 0 {
			errors = append(errors, blockErrors...)
		}

		// Mark as processed
		fa.ProcessedBlocks[block.ID] = true

		// Add successors to worklist if their input changed
		for _, successor := range block.Successors {
			if fa.shouldReprocess(successor) {
				fa.WorkList = append(fa.WorkList, successor)
				delete(fa.ProcessedBlocks, successor.ID)
			}
		}

		iterations++
	}

	if iterations >= fa.MaxIterations {
		errors = append(errors, fmt.Errorf("flow analysis reached maximum iterations, possible infinite loop"))
	}

	// After flow analysis completes, perform safety analysis (if enabled)
	if fa.TypeChecker.SafetyAnalysisEnabled {
		safetyErrors := fa.performSafetyAnalysis(cfg)
		errors = append(errors, safetyErrors...)
	}

	return errors
}

// processBlock processes a single basic block for data flow analysis
func (fa *FlowAnalyzer) processBlock(block *BasicBlock) []error {
	var errors []error

	// Merge input from predecessors
	if len(block.Predecessors) > 1 {
		var predecessorStates []*TypeState
		for _, pred := range block.Predecessors {
			predecessorStates = append(predecessorStates, pred.ExitTypeState)
		}
		block.TypeState = fa.mergeMultipleTypeStates(predecessorStates)
	}

	// Process statements in the block
	currentState := fa.copyTypeState(block.TypeState)
	fmt.Printf("DEBUG: Block %d has %d statements\n", block.ID, len(block.Statements))
	for i, stmt := range block.Statements {
		fmt.Printf("DEBUG: Statement %d type: %s\n", i, stmt.Type())
		stmtErrors := fa.processStatement(stmt, currentState)
		if len(stmtErrors) > 0 {
			errors = append(errors, stmtErrors...)
		}
	}

	// Save the updated state as the block's exit state
	block.ExitTypeState = currentState

	return errors
}

// processStatement processes a single statement and updates the type state
func (fa *FlowAnalyzer) processStatement(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// DEBUG: Log what statement types we're processing
	fmt.Printf("DEBUG: Processing statement type: %s\n", stmt.Type())

	// Handle different statement types
	switch stmt.Type() {
	case "variable_declaration", "var_decl":
		errors = append(errors, fa.processVariableDeclaration(stmt, state)...)
	case "assignment", "assignment_expression", "assignment_expr":
		fmt.Printf("DEBUG: Processing assignment statement\n")
		errors = append(errors, fa.processAssignment(stmt, state)...)
	case "function_call":
		errors = append(errors, fa.processFunctionCall(stmt, state)...)
	case "expression_stmt", "expression_statement":
		errors = append(errors, fa.processExpressionStatement(stmt, state)...)
	case "return_stmt":
		errors = append(errors, fa.processReturnStatement(stmt, state)...)
	case "literal":
		// Handle literal expressions that might contain assignments/function calls
		if literalExpr, ok := stmt.(*ast.LiteralExpr); ok {
			fmt.Printf("DEBUG: Processing literal: %s\n", literalExpr.Value)
			if strings.Contains(literalExpr.Value, "=") && (strings.Contains(literalExpr.Value, "ref(") || strings.Contains(literalExpr.Value, "eq")) {
				fmt.Printf("DEBUG: Found literal with assignment: %s\n", literalExpr.Value)
				errors = append(errors, fa.processLiteralAssignment(literalExpr, state)...)
			}
		}
	default:
		// For unknown statement types, we don't change the type state
		// In a full implementation, we'd handle more statement types
		fmt.Printf("DEBUG: Unhandled statement type: %s\n", stmt.Type())
	}

	return errors
}

// processLiteralAssignment processes literal expressions that contain assignment patterns
func (fa *FlowAnalyzer) processLiteralAssignment(literal *ast.LiteralExpr, state *TypeState) []error {
	var errors []error

	// Initialize type maps if needed
	if state.VariableTypes == nil {
		state.VariableTypes = make(map[string]string)
	}

	// Simple pattern matching for assignments in literals
	text := literal.Value

	// Pattern: my $var = func(...);
	if strings.Contains(text, "my $") && strings.Contains(text, "=") {
		// Extract variable name and function call
		parts := strings.Split(text, "=")
		if len(parts) == 2 {
			leftPart := strings.TrimSpace(parts[0])
			rightPart := strings.TrimSpace(parts[1])

			// Extract variable name from "my $varname"
			if strings.HasPrefix(leftPart, "my $") {
				varName := strings.TrimPrefix(leftPart, "my $")
				varName = strings.TrimSpace(varName)

				// Infer type from right-hand side
				var inferredType string

				if strings.HasPrefix(rightPart, "ref(") {
					inferredType = "Str"
				} else if strings.Contains(rightPart, " eq ") {
					inferredType = "Bool"
				} else if strings.HasPrefix(rightPart, "length(") {
					inferredType = "Int"
				} else if strings.HasPrefix(rightPart, "defined(") {
					inferredType = "Bool"
				} else {
					inferredType = "Any"
				}

				fmt.Printf("DEBUG: Literal assignment: %s = %s (type: %s)\n", varName, rightPart, inferredType)
				state.VariableTypes[varName] = inferredType
			}
		}
	}

	return errors
}

// processExpressionStatement processes expression statements that might contain variable declarations or assignments
func (fa *FlowAnalyzer) processExpressionStatement(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Expression statements may contain various types of expressions
	// Look for the actual expression inside the expression_stmt node
	for _, child := range stmt.Children() {
		if child.Type() == "token" {
			// Skip tokens like semicolons
			continue
		}

		// Process based on the child type
		switch child.Type() {
		case "variable_declaration", "var_decl":
			errors = append(errors, fa.processVariableDeclaration(child, state)...)
		case "assignment", "assignment_expr":
			errors = append(errors, fa.processAssignment(child, state)...)
		case "call_expr":
			errors = append(errors, fa.processFunctionCall(child, state)...)
		case "literal":
			// DEBUG: Check what's in the literal
			if literalExpr, ok := child.(*ast.LiteralExpr); ok {
				// For now, try to parse variable declarations from literals
				if strings.Contains(literalExpr.Value, "my $") {
					// This might be a variable declaration that wasn't properly parsed
					errors = append(errors, fa.processLiteralAsVariableDeclaration(literalExpr, state)...)
				} else {
					// Check for assignment expressions in literals
					// TODO: Implement processLiteralAsAssignment
				}
			}
		default:
			// Try to process as a generic statement
			childErrors := fa.processStatement(child, state)
			errors = append(errors, childErrors...)
		}
	}

	return errors
}

// processReturnStatement processes return statements and their expressions
func (fa *FlowAnalyzer) processReturnStatement(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Process the return expression to check for variable usage
	for _, child := range stmt.Children() {
		if child.Type() == "token" {
			// Skip tokens like semicolons
			continue
		}

		// Check for variable references in return expression
		switch child.Type() {
		case "variable", "var_expr":
			if varExpr, ok := child.(*ast.VariableExpr); ok {
				varName := varExpr.Name
				if varName != "" && !strings.HasPrefix(varName, "$") {
					// Ensure variable name doesn't have sigil already
					if _, exists := state.VariableTypes[varName]; !exists {
						errors = append(errors, fmt.Errorf("Variable $%s may be uninitialized in return statement", varName))
					}
				}
			}
		case "literal":
			// Literals in return statements are generally safe
			continue
		default:
			// For other expression types, recursively check for unsafe patterns
			// TODO: Implement comprehensive expression safety checking
			// For now, skip complex expressions in return statements
		}
	}

	return errors
}

// processLiteralAsVariableDeclaration processes literals that contain variable declarations
func (fa *FlowAnalyzer) processLiteralAsVariableDeclaration(literal *ast.LiteralExpr, state *TypeState) []error {
	var errors []error

	// Initialize type maps if needed
	if state.VariableTypes == nil {
		state.VariableTypes = make(map[string]string)
	}

	// Parse patterns like "my $var = func()" from the literal value
	// This is a simplified parser for the literal content
	content := strings.TrimSpace(literal.Value)

	// Handle "my $var = function_call" pattern
	if strings.HasPrefix(content, "my $") {
		// Extract variable name and function call
		parts := strings.Split(content, "=")
		if len(parts) == 2 {
			varPart := strings.TrimSpace(parts[0])
			funcPart := strings.TrimSpace(parts[1])

			// Extract variable name from "my $varname"
			varMatch := strings.TrimPrefix(varPart, "my $")
			varName := strings.TrimSpace(varMatch)

			// Try to infer type from the function call
			if strings.Contains(funcPart, "(") {
				// Extract function name (e.g., "ref($input)" -> "ref")
				funcName := strings.Split(funcPart, "(")[0]
				funcName = strings.TrimSpace(funcName)

				// Create a dummy call expression for type inference
				inferredType := fa.inferBuiltinFunctionType(funcName, nil)
				if inferredType == "" {
					// Fallback to hardcoded inference
					switch funcName {
					case "ref":
						inferredType = "Str"
					case "defined":
						inferredType = "Bool"
					case "keys":
						inferredType = "ArrayRef[Str]"
					default:
						inferredType = "Any"
					}
				}

				// Store the variable type
				state.VariableTypes[varName] = inferredType
			}
		}
	}

	return errors
}

// Helper methods for type state management

// copyTypeState creates a deep copy of a type state
func (fa *FlowAnalyzer) copyTypeState(state *TypeState) *TypeState {
	if state == nil {
		return &TypeState{
			VariableTypes:  make(map[string]string),
			RefinedTypes:   make(map[string]string),
			Conditions:     []Condition{},
			FieldAccess:    make(map[string]map[string]bool),
			ExceptionTypes: make(map[string]bool),
		}
	}

	newState := &TypeState{
		VariableTypes:  make(map[string]string),
		RefinedTypes:   make(map[string]string),
		Conditions:     make([]Condition, len(state.Conditions)),
		FieldAccess:    make(map[string]map[string]bool),
		ExceptionTypes: make(map[string]bool),
		SkipFlowChecks: state.SkipFlowChecks,
	}

	// Copy variable types
	for k, v := range state.VariableTypes {
		newState.VariableTypes[k] = v
	}

	// Copy refined types
	for k, v := range state.RefinedTypes {
		newState.RefinedTypes[k] = v
	}

	// Copy field access tracking (deep copy of nested map)
	for varName, fields := range state.FieldAccess {
		newState.FieldAccess[varName] = make(map[string]bool)
		for fieldName, accessed := range fields {
			newState.FieldAccess[varName][fieldName] = accessed
		}
	}

	// Copy exception types
	for excType, present := range state.ExceptionTypes {
		newState.ExceptionTypes[excType] = present
	}

	// Copy conditions
	copy(newState.Conditions, state.Conditions)

	return newState
}

// mergeTypeStates merges two type states (for control flow join points)
func (fa *FlowAnalyzer) mergeTypeStates(state1, state2 *TypeState) *TypeState {
	if state1 == nil {
		return fa.copyTypeState(state2)
	}
	if state2 == nil {
		return fa.copyTypeState(state1)
	}

	merged := &TypeState{
		VariableTypes:  make(map[string]string),
		RefinedTypes:   make(map[string]string),
		Conditions:     []Condition{},
		FieldAccess:    make(map[string]map[string]bool),
		ExceptionTypes: make(map[string]bool),
	}

	// Merge variable types - use union types if they differ
	for varName := range state1.VariableTypes {
		type1 := state1.VariableTypes[varName]
		type2, exists := state2.VariableTypes[varName]
		if exists {
			if type1 == type2 {
				merged.VariableTypes[varName] = type1
			} else {
				// Create union type
				merged.VariableTypes[varName] = fa.createUnionType(type1, type2)
			}
		} else {
			// Variable only exists in state1, include it
			merged.VariableTypes[varName] = type1
		}
	}

	// Add variables that only exist in state2
	for varName, varType := range state2.VariableTypes {
		if _, exists := state1.VariableTypes[varName]; !exists {
			merged.VariableTypes[varName] = varType
		}
	}

	// Merge field access tracking - union of all accessed fields
	for varName, fields1 := range state1.FieldAccess {
		merged.FieldAccess[varName] = make(map[string]bool)
		// Copy fields from state1
		for fieldName, accessed := range fields1 {
			merged.FieldAccess[varName][fieldName] = accessed
		}
		// Add fields from state2 if they exist
		if fields2, exists := state2.FieldAccess[varName]; exists {
			for fieldName, accessed := range fields2 {
				merged.FieldAccess[varName][fieldName] = accessed
			}
		}
	}

	// Add field access from state2 for variables not in state1
	for varName, fields2 := range state2.FieldAccess {
		if _, exists := merged.FieldAccess[varName]; !exists {
			merged.FieldAccess[varName] = make(map[string]bool)
			for fieldName, accessed := range fields2 {
				merged.FieldAccess[varName][fieldName] = accessed
			}
		}
	}

	// Merge exception types - union of all exception types from both states
	for excType, present := range state1.ExceptionTypes {
		if present {
			merged.ExceptionTypes[excType] = true
		}
	}
	for excType, present := range state2.ExceptionTypes {
		if present {
			merged.ExceptionTypes[excType] = true
		}
	}

	return merged
}

// mergeMultipleTypeStates merges multiple type states
func (fa *FlowAnalyzer) mergeMultipleTypeStates(states []*TypeState) *TypeState {
	if len(states) == 0 {
		return &TypeState{
			VariableTypes:  make(map[string]string),
			RefinedTypes:   make(map[string]string),
			Conditions:     []Condition{},
			FieldAccess:    make(map[string]map[string]bool),
			ExceptionTypes: make(map[string]bool),
		}
	}
	if len(states) == 1 {
		return fa.copyTypeState(states[0])
	}

	result := fa.copyTypeState(states[0])
	for i := 1; i < len(states); i++ {
		result = fa.mergeTypeStates(result, states[i])
	}
	return result
}

// createUnionType creates a union type from two types
func (fa *FlowAnalyzer) createUnionType(type1, type2 string) string {
	// Simple union type creation - in a full implementation,
	// this would use the type hierarchy to create proper union types
	if type1 == type2 {
		return type1
	}
	return type1 + "|" + type2
}

// shouldReprocess determines if a block should be reprocessed
func (fa *FlowAnalyzer) shouldReprocess(block *BasicBlock) bool {
	// Simple heuristic - in a full implementation, we'd compare
	// the input type state to see if it changed
	return !fa.ProcessedBlocks[block.ID]
}

// Type refinement methods

// refineTypeForCondition refines types based on conditional expressions
func (fa *FlowAnalyzer) refineTypeForCondition(condNode ast.Node, block *BasicBlock, positive bool) {
	// Extract condition information and apply type refinement
	// This is a simplified implementation - a full version would
	// parse the condition and extract the variable and operation

	// For now, we'll implement basic 'defined' checking
	if fa.isDefinedCheck(condNode) {
		varName := fa.extractVariableFromCondition(condNode)
		if varName != "" {
			fa.TypeChecker.refineTypeAfterCondition(varName, "defined", positive)
			// Update the block's type state
			if currentType, exists := block.TypeState.VariableTypes[varName]; exists {
				if positive {
					// After defined($var), exclude Undef
					refinedType := fa.TypeChecker.excludeTypeFromUnion(currentType, "Undef")
					block.TypeState.RefinedTypes[varName] = refinedType
				} else {
					// After !defined($var), variable must be Undef
					block.TypeState.RefinedTypes[varName] = "Undef"
				}
			}
		}
	}
}

// updateTypeStateForStatement updates type state based on a statement
func (fa *FlowAnalyzer) updateTypeStateForStatement(stmt ast.Node, block *BasicBlock) {
	// Update type state based on the statement
	// This would analyze assignments, declarations, etc.
	// For now, this is a placeholder
}

// Helper methods for condition analysis

// isDefinedCheck checks if a condition is a 'defined' check
func (fa *FlowAnalyzer) isDefinedCheck(node ast.Node) bool {
	// Simple heuristic - check if the node text contains 'defined'
	return strings.Contains(node.Text(), "defined")
}

// extractVariableFromCondition extracts the variable name from a condition
func (fa *FlowAnalyzer) extractVariableFromCondition(node ast.Node) string {
	// Simple extraction - in a full implementation, this would
	// parse the condition expression properly
	text := node.Text()
	if strings.Contains(text, "defined(") {
		// Extract variable name from defined($var)
		start := strings.Index(text, "defined(") + 8
		end := strings.Index(text[start:], ")")
		if end > 0 {
			varText := strings.TrimSpace(text[start : start+end])
			// Remove $ prefix if present
			if strings.HasPrefix(varText, "$") {
				return varText[1:]
			}
			return varText
		}
	}
	return ""
}

// Statement processing methods

// processVariableDeclaration processes variable declarations
func (fa *FlowAnalyzer) processVariableDeclaration(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Call test hook if present
	if fa.processVariableDeclarationHook != nil {
		fa.processVariableDeclarationHook(stmt)
	}

	// DEBUG: Log what we're processing
	fmt.Printf("DEBUG: processVariableDeclaration called\n")
	fmt.Printf("DEBUG: Statement type: %s, text: %q\n", stmt.Type(), stmt.Text())

	// Cast to VarDecl node - check by node type first
	if stmt.Type() != "var_decl" {
		return errors // Not a variable declaration, skip
	}

	varDecl, ok := stmt.(*ast.VarDecl)
	if !ok {
		return errors // Type assertion failed, skip
	}

	// DEBUG: Print detailed info about the variable declaration
	fmt.Printf("DEBUG: VarDecl.Initializer: %v\n", varDecl.Initializer)
	if varDecl.Initializer != nil {
		fmt.Printf("DEBUG: Initializer type: %s, text: %q\n", varDecl.Initializer.Type(), varDecl.Initializer.Text())
	} else {
		fmt.Printf("DEBUG: Initializer is nil\n")
	}

	// Initialize type maps if needed
	if state.VariableTypes == nil {
		state.VariableTypes = make(map[string]string)
	}

	// Process each variable in the declaration
	variables := varDecl.Variables()
	fmt.Printf("DEBUG: Found %d variables in declaration\n", len(variables))

	for _, variable := range variables {
		if variable == nil {
			fmt.Printf("DEBUG: Skipping nil variable\n")
			continue
		}

		varName := variable.Name
		fmt.Printf("DEBUG: Variable.Name='%s', Variable.Sigil='%s', FullName='%s'\n",
			variable.Name, variable.Sigil, variable.FullName())

		if varName == "" {
			// Try using FullName() instead
			varName = variable.FullName()
			if varName == "" {
				fmt.Printf("DEBUG: Skipping variable with empty name and FullName\n")
				continue
			}
			fmt.Printf("DEBUG: Using FullName: %s\n", varName)
		}

		fmt.Printf("DEBUG: Processing variable: %s\n", varName)

		// Extract type from type annotation or infer from initializer
		var varType string

		// First, check for explicit type annotation on the declaration
		if varDecl.TypeExpr != nil {
			varType = fa.extractTypeFromTypeExpression(varDecl.TypeExpr)
			fmt.Printf("DEBUG: Type from annotation: %s\n", varType)
		} else if varDecl.Initializer != nil {
			// If no type annotation, try to infer from initializer
			fmt.Printf("DEBUG: Initializer found, type: %s\n", varDecl.Initializer.Type())

			// Debug: print more details about the initializer
			if varDecl.Initializer != nil {
				fmt.Printf("DEBUG: Initializer text: %q\n", varDecl.Initializer.Text())
				if literal, ok := varDecl.Initializer.(*ast.LiteralExpr); ok {
					fmt.Printf("DEBUG: LiteralExpr Kind: %d\n", literal.Kind)
				}
			}

			varType = fa.inferTypeFromExpression(varDecl.Initializer)
			fmt.Printf("DEBUG: Type from initializer: %s\n", varType)

			// Fallback: if initializer inference returns "Any", try source text analysis
			if varType == "Any" {
				fmt.Printf("DEBUG: Initializer returned 'Any', trying source text analysis\n")
				extractedType := fa.extractTypeFromSourceText(varName, stmt)
				if extractedType != "" && extractedType != "Any" {
					varType = extractedType
					fmt.Printf("DEBUG: Source text analysis found better type: %s\n", varType)
				}
			}
		} else {
			fmt.Printf("DEBUG: No TypeExpr and no Initializer found\n")
			// WORKAROUND: Try to extract initializer from source text
			extractedType := fa.extractTypeFromSourceText(varName, stmt)
			if extractedType != "" && extractedType != "Any" {
				varType = extractedType
				fmt.Printf("DEBUG: Extracted type from source text: %s\n", varType)
			}
		}

		// Fall back to "Any" if we can't determine the type
		if varType == "" {
			varType = "Any"
			fmt.Printf("DEBUG: Falling back to 'Any' type\n")
		}

		// Update type state with the variable type
		state.VariableTypes[varName] = varType
		fmt.Printf("DEBUG: Stored variable '%s' with type '%s'\n", varName, varType)

		// Validate the type annotation if provided
		if varDecl.TypeExpr != nil && varType != "" && varType != "Any" {
			if err := fa.validateTypeAnnotation(varType, varDecl.Start()); err != nil {
				errors = append(errors, err)
			}
		}

		// If there's an initializer, check type compatibility and safety
		if varDecl.Initializer != nil {
			// Check type compatibility if we have a specific type
			if varType != "Any" {
				initType := fa.inferTypeFromExpression(varDecl.Initializer)
				if initType != "" && initType != "Any" && initType != varType {
					if err := fa.checkTypeCompatibility(initType, varType, varDecl.Start()); err != nil {
						errors = append(errors, err)
					}
				}
			}

			// IMPORTANT: Perform safety analysis on the initializer expression
			initSafetyErrors := fa.analyzeSafetyInExpression(varDecl.Initializer, state)
			errors = append(errors, initSafetyErrors...)
		}
	}

	return errors
}

// extractTypeFromSourceText attempts to extract type information from source text
// This is a workaround for parser issues where variable declarations lose their initializers
func (fa *FlowAnalyzer) extractTypeFromSourceText(varName string, stmt ast.Node) string {
	// Use the source code stored in the analyzer
	sourceText := fa.SourceCode

	fmt.Printf("DEBUG: Analyzing source text for variable '%s' in: %q\n", varName, sourceText)

	if sourceText == "" {
		return ""
	}

	// Look for patterns like: my $varName = expression;
	return fa.inferTypeFromInitializerPattern(varName, sourceText)
}

// inferTypeFromInitializerPattern analyzes source text to find variable initializers
func (fa *FlowAnalyzer) inferTypeFromInitializerPattern(varName string, sourceText string) string {
	// Clean up the variable name (remove sigil if present)
	cleanVarName := strings.TrimPrefix(varName, "$")

	// First, try to identify function calls using the BuiltinTypeRegistry
	if fa.BuiltinTypes != nil {
		// Look for function call patterns: $var = funcname(...)
		funcCallPattern := fmt.Sprintf(`\$%s\s*=\s*(\w+)\s*\(`, cleanVarName)
		if match := regexp.MustCompile(funcCallPattern).FindStringSubmatch(sourceText); len(match) > 1 {
			funcName := match[1]
			fmt.Printf("DEBUG: Found function call '%s' in assignment to %s\n", funcName, varName)

			// Check if it's a builtin function
			if fa.BuiltinTypes.IsBuiltinFunction(funcName) {
				// For simplicity, assume 1 parameter for now (could be improved)
				if returnType := fa.BuiltinTypes.GetFunctionType(funcName, 1); returnType != "" {
					fmt.Printf("DEBUG: Using BuiltinTypeRegistry: %s() returns %s\n", funcName, returnType)
					return returnType
				}
			}
		}
	}

	// Look for assignment patterns - ORDER MATTERS: more specific patterns first
	patterns := []struct {
		pattern  string
		typeFunc func(string) string
	}{
		// Boolean comparison operations - MUST come before ref() to catch comparisons
		{fmt.Sprintf(`\$%s\s*=\s*.*\s+(eq|ne|lt|gt|le|ge|==|!=|<|>|<=|>=)\s+`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found comparison operation for variable %s\n", varName)
			return "Bool"
		}},
		// defined() function calls - also returns Bool
		{fmt.Sprintf(`\$%s\s*=\s*defined\s*\(`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found defined() call for variable %s\n", varName)
			return "Bool"
		}},
		// keys() function calls in scalar context
		{fmt.Sprintf(`\$%s\s*=\s*keys\s*\(`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found keys() call in scalar context for variable %s\n", varName)
			return "Int"
		}},
		// ref() function calls - simpler pattern, comes after comparisons
		{fmt.Sprintf(`\$%s\s*=\s*ref\s*\(`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found ref() call for variable %s\n", varName)
			return "Str"
		}},
		// Array slice assignments
		{fmt.Sprintf(`@%s\s*=\s*keys\s*\(`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found keys() call in list context for variable %s\n", varName)
			return "ArrayRef[Str]"
		}},
		{fmt.Sprintf(`@%s\s*=\s*values\s*\(`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found values() call in list context for variable %s\n", varName)
			return "ArrayRef[Any]"
		}},
		// length() function calls
		{fmt.Sprintf(`\$%s\s*=\s*length\s*\(`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found length() call for variable %s\n", varName)
			return "Int"
		}},
		// int() function calls
		{fmt.Sprintf(`\$%s\s*=\s*int\s*\(`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found int() call for variable %s\n", varName)
			return "Int"
		}},
		// abs() function calls
		{fmt.Sprintf(`\$%s\s*=\s*abs\s*\(`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found abs() call for variable %s\n", varName)
			return "Num"
		}},
		// Arithmetic operations - improved pattern
		{fmt.Sprintf(`\$%s\s*=\s*.*\s*[\+\-\*\/%%]\s*`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found arithmetic operation for variable %s\n", varName)
			return "Num"
		}},
		// String operations
		{fmt.Sprintf(`\$%s\s*=\s*.*(uc|lc|ucfirst|lcfirst)\s*\(`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found string manipulation function for variable %s\n", varName)
			return "Str"
		}},
		// chomp() function calls
		{fmt.Sprintf(`\$%s\s*=\s*chomp\s*\(`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found chomp() call for variable %s\n", varName)
			return "Int"
		}},
		// split() function calls in array context
		{fmt.Sprintf(`@%s\s*=\s*split\s*\(`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found split() call in array context for variable %s\n", varName)
			return "ArrayRef[Str]"
		}},
		// Array assignments from arrays
		{fmt.Sprintf(`@%s\s*=\s*`, cleanVarName), func(match string) string {
			fmt.Printf("DEBUG: Found array assignment for variable %s\n", varName)
			return "ArrayRef[Any]"
		}},
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern.pattern, sourceText); matched {
			return pattern.typeFunc(sourceText)
		}
	}

	fmt.Printf("DEBUG: No matching pattern found for variable %s in text: %q\n", varName, sourceText)
	return ""
}

// processAssignment processes assignment statements
func (fa *FlowAnalyzer) processAssignment(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Cast to AssignmentExpr node - check by node type first
	if stmt.Type() != "assignment_expr" {
		return errors // Not an assignment expression, skip
	}

	assign, ok := stmt.(*ast.AssignmentExpr)
	if !ok {
		return errors // Type assertion failed, skip
	}

	// Initialize type maps if needed
	if state.VariableTypes == nil {
		state.VariableTypes = make(map[string]string)
	}

	// Extract target variable name from LHS
	targetVar := fa.extractVariableFromExpression(assign.Left)
	if targetVar == "" {
		// Not assigning to a simple variable (e.g., array element, method call result)
		// Still validate RHS for any type errors
		fa.inferTypeFromExpression(assign.Right)
		return errors
	}

	// Get current type of target variable from type state
	targetType := fa.getVariableTypeFromState(targetVar, state)

	// Infer type of RHS expression
	sourceType := fa.inferTypeFromExpression(assign.Right)
	if sourceType == "" {
		sourceType = "Any"
	}

	// Check if RHS is a safe expression (e.g., ternary with defined check)
	isSafeAssignment := fa.isAssignmentFromSafeExpression(assign.Right)

	// For compound assignments (+=, -=, etc.), need special handling
	if assign.Operator != "=" {
		// For compound assignments, both operands should be compatible with the operation
		if targetType != "" && targetType != "Any" && sourceType != "Any" {
			if err := fa.checkCompoundAssignmentCompatibility(targetType, sourceType, assign.Operator, assign.Start()); err != nil {
				errors = append(errors, err)
			}
		}
		// Result type remains the same for compound assignments
		return errors
	}

	// For simple assignment (=), always update the type
	if sourceType != "" && sourceType != "Any" {
		// Store the inferred type
		state.VariableTypes[targetVar] = sourceType

		// Also check type compatibility if there was a previous type
		if targetType != "" && targetType != "Any" && targetType != sourceType {
			if err := fa.checkTypeCompatibility(sourceType, targetType, assign.Start()); err != nil {
				errors = append(errors, err)
			}
		}
	} else if targetType == "" || targetType == "Any" {
		// Variable not previously declared or has Any type
		if sourceType != "" {
			state.VariableTypes[targetVar] = sourceType
		} else {
			state.VariableTypes[targetVar] = "Any"
		}
	}

	// If this is a safe assignment, mark the variable as safe
	if isSafeAssignment {
		// For safe assignments (like ternary with defined check),
		// we consider the variable to be safely initialized
		if targetType == "Maybe[Str]" || strings.HasPrefix(targetType, "Maybe[") {
			// Extract the inner type from Maybe[T]
			if innerType := extractMaybeParameter(targetType); innerType != "" {
				state.VariableTypes[targetVar] = innerType
			}
		} else {
			// Even if no pre-existing type, treat safe assignment as definite
			// For ternary with defined() check, the result is guaranteed non-undef
			if sourceType != "" && sourceType != "Any" {
				state.VariableTypes[targetVar] = sourceType
			} else {
				// Default to Str for safe ternary expressions that produce strings
				state.VariableTypes[targetVar] = "Str"
			}
		}
	}

	return errors
}

// processFunctionCall processes function calls
func (fa *FlowAnalyzer) processFunctionCall(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Cast to CallExpr node - check by node type first
	if stmt.Type() != "call_expr" {
		return errors // Not a function call expression, skip
	}

	call, ok := stmt.(*ast.CallExpr)
	if !ok {
		return errors // Type assertion failed, skip
	}

	// Extract function name from the call expression
	functionName := fa.extractFunctionNameFromCall(call)
	if functionName == "" {
		// Can't determine function name (e.g., complex expression)
		// Still validate arguments for any type errors
		for _, arg := range call.Arguments {
			fa.inferTypeFromExpression(arg)
		}
		return errors
	}

	// Look up function signature in the type checker's function types
	signature, exists := fa.TypeChecker.FunctionTypes[functionName]
	if !exists {
		// Function signature not found - might be built-in, external, or undeclared
		// Still validate arguments for any type errors
		for _, arg := range call.Arguments {
			fa.inferTypeFromExpression(arg)
		}
		return errors
	}

	// Validate parameter count
	expectedParamCount := len(signature.ParameterTypes)
	actualParamCount := len(call.Arguments)

	if actualParamCount != expectedParamCount {
		err := &TypeCheckError{
			Message: fmt.Sprintf("Function '%s' expects %d parameters, got %d", functionName, expectedParamCount, actualParamCount),
			Path:    "",
			Line:    call.Start().Line,
			Column:  call.Start().Column,
		}
		errors = append(errors, err)
		return errors // Can't validate individual parameters if count is wrong
	}

	// Validate each parameter type
	paramNames := fa.getParameterNamesFromSignature(signature)
	for i, arg := range call.Arguments {
		if i >= len(paramNames) {
			break // Safety check
		}

		paramName := paramNames[i]
		expectedType, exists := signature.ParameterTypes[paramName]
		if !exists {
			continue // Parameter not found in signature, skip
		}

		actualType := fa.inferTypeFromExpression(arg)
		if actualType == "" || actualType == "Any" {
			continue // Can't validate unknown or Any type
		}

		// Check type compatibility
		if expectedType != "Any" && actualType != expectedType {
			if err := fa.checkTypeCompatibility(actualType, expectedType, call.Start()); err != nil {
				// Customize error message for function parameter
				paramErr := &TypeCheckError{
					Message: fmt.Sprintf("Parameter %d of function '%s' expects type '%s', got '%s'", i+1, functionName, expectedType, actualType),
					Path:    err.(*TypeCheckError).Path,
					Line:    call.Start().Line,
					Column:  call.Start().Column,
				}
				errors = append(errors, paramErr)
			}
		}
	}

	return errors
}

// AddFlowPatterns adds custom flow patterns for validation
func (tc *TypeChecker) AddFlowPatterns(patterns []string) {
	// Convert string patterns to ValidationPattern objects
	for _, pattern := range patterns {
		// Parse the pattern and create a ValidationPattern
		// For now, we'll just store the patterns as names
		validationPattern := ValidationPattern{
			Name:    pattern,
			Pattern: pattern,
			RefinementFunc: func(variable string, currentType string) string {
				// Default refinement - no change
				return currentType
			},
			Checker: func(node ast.Node) (string, bool) {
				// Default checker - matches nothing
				return "", false
			},
		}
		tc.ValidationPatterns = append(tc.ValidationPatterns, validationPattern)
	}
}

// Helper methods for flow control analysis

// extractTypeFromTypeExpression extracts a type string from a TypeExpression AST node
func (fa *FlowAnalyzer) extractTypeFromTypeExpression(typeExpr *ast.TypeExpression) string {
	if typeExpr == nil {
		return ""
	}

	// Use the existing String() method on TypeExpression
	return typeExpr.String()
}

// inferTypeFromExpression performs sophisticated type inference on expressions
func (fa *FlowAnalyzer) inferTypeFromExpression(expr ast.ExpressionNode) string {
	if expr == nil {
		return "Any"
	}

	// Use the existing inference engine if available
	if fa.TypeChecker.InferenceEngine != nil {
		// Try to get inferred type from the inference engine
		inferredTypes := fa.TypeChecker.InferenceEngine.GetAllInferredTypes()

		// For variable expressions, check if we have an inferred type
		if varExpr, ok := expr.(*ast.VariableExpr); ok {
			if inferredType, exists := inferredTypes[varExpr.Name]; exists {
				return inferredType
			}
		}
	}

	// Enhanced type inference based on expression type
	switch expr.Type() {
	case "literal":
		if literal, ok := expr.(*ast.LiteralExpr); ok {
			return fa.inferTypeFromLiteral(literal)
		}
	case "variable":
		if varExpr, ok := expr.(*ast.VariableExpr); ok {
			return fa.inferVariableType(varExpr)
		}
	case "call_expr":
		if callExpr, ok := expr.(*ast.CallExpr); ok {
			return fa.inferReturnTypeFromCall(callExpr)
		}
	case "binary_expr":
		if binaryExpr, ok := expr.(*ast.BinaryExpr); ok {
			return fa.inferTypeFromBinaryExpression(binaryExpr)
		}
	case "hash_ref_expr":
		if hashRef, ok := expr.(*ast.HashRefExpr); ok {
			return fa.inferTypeFromHashAccess(hashRef)
		}
	case "array_ref_expr":
		if arrayRef, ok := expr.(*ast.ArrayRefExpr); ok {
			return fa.inferTypeFromArrayAccess(arrayRef)
		}
	case "conditional_expr":
		if conditional, ok := expr.(*ast.ConditionalExpr); ok {
			return fa.inferTypeFromConditional(conditional)
		}
	}

	// Perl default: most scalar values are strings unless proven otherwise
	return "Str"
}

// inferTypeFromLiteral infers type from literal expressions
func (fa *FlowAnalyzer) inferTypeFromLiteral(literal *ast.LiteralExpr) string {
	if literal == nil {
		return "Any"
	}

	switch literal.Kind {
	case ast.NumberLiteral:
		return "Num" // Could be more specific with Int vs Num analysis
	case ast.StringLiteral:
		return "Str"
	case ast.BooleanLiteral:
		return "Bool"
	case ast.UndefLiteral:
		return "Undef"
	case ast.RegexLiteral:
		return "Regex"
	case ast.FunctionCallLiteral:
		// Parse function calls from literal text
		return fa.inferTypeFromFunctionCallLiteral(literal.Value)
	case ast.MethodCallLiteral:
		// Parse method calls from literal text (Class->method(), $obj->method())
		return fa.inferTypeFromMethodCallLiteral(literal.Value)
	case ast.ExpressionLiteral:
		// Generic expression - try to parse function calls
		return fa.inferTypeFromFunctionCallLiteral(literal.Value)
	default:
		return "Any"
	}
}

// inferTypeFromFunctionCallLiteral infers type from function call patterns in literal text
func (fa *FlowAnalyzer) inferTypeFromFunctionCallLiteral(text string) string {
	text = strings.TrimSpace(text)

	// FIRST: Check for comparison operators - these always return Bool
	comparisonOps := []string{"eq", "ne", "lt", "gt", "le", "ge", "==", "!=", "<", ">", "<=", ">=", "=~", "!~"}
	for _, op := range comparisonOps {
		if strings.Contains(text, " "+op+" ") {
			fmt.Printf("DEBUG: Found comparison operator '%s' in expression: %s\n", op, text)
			return "Bool"
		}
	}

	// SECOND: Pattern: functionName(args...)
	if strings.Contains(text, "(") && strings.Contains(text, ")") {
		// Extract function name
		parenIndex := strings.Index(text, "(")
		functionName := strings.TrimSpace(text[:parenIndex])

		fmt.Printf("DEBUG: Parsing function call literal: %s -> function: %s\n", text, functionName)

		// Check builtin function types
		if fa.BuiltinTypes != nil {
			if builtinSigs := fa.BuiltinTypes.GetFunctionSignatures(functionName); len(builtinSigs) > 0 {
				// Use the first signature for now (could be more sophisticated)
				builtinSig := builtinSigs[0]
				returnType := builtinSig.ReturnType
				fmt.Printf("DEBUG: Found builtin signature for %s: %s\n", functionName, returnType)
				return returnType
			}
		}

		// Hardcoded builtin types as fallback
		switch functionName {
		case "ref":
			return "Str"
		case "defined":
			return "Bool"
		case "length", "scalar":
			return "Int"
		case "keys", "values":
			return "Array[Str]"
		case "int":
			return "Int"
		case "abs":
			return "Num"
		}
	}

	return "Any"
}

// inferTypeFromMethodCallLiteral infers type from method call patterns in literal text
func (fa *FlowAnalyzer) inferTypeFromMethodCallLiteral(text string) string {
	text = strings.TrimSpace(text)

	fmt.Printf("DEBUG: Parsing method call literal: %s\n", text)

	// Handle method chains like JSON->new()->decode($data_str)
	// We need to parse from right to left to get the final method result

	// Find the last method call in the chain
	lastArrowIndex := strings.LastIndex(text, "->")
	if lastArrowIndex == -1 {
		return "Any"
	}

	// Get the final method call
	finalMethodPart := text[lastArrowIndex+2:]

	// Extract method name from finalMethodPart (remove arguments)
	finalMethodName := finalMethodPart
	if parenIndex := strings.Index(finalMethodPart, "("); parenIndex != -1 {
		finalMethodName = strings.TrimSpace(finalMethodPart[:parenIndex])
	}

	fmt.Printf("DEBUG: Final method in chain: %s\n", finalMethodName)

	// Get the object/class part (everything before the final ->)
	objectPart := strings.TrimSpace(text[:lastArrowIndex])

	// For constructor calls (Class->new), return the class name
	if finalMethodName == "new" {
		// Find the class name (should be at the beginning or after the last ->)
		if strings.Contains(objectPart, "->") {
			// This is a chained call ending with ->new(), get the result of previous chain
			return fa.inferTypeFromMethodCallLiteral(objectPart)
		} else {
			// This is Class->new()
			if !strings.HasPrefix(objectPart, "$") {
				fmt.Printf("DEBUG: Constructor call for class: %s\n", objectPart)
				return objectPart
			}
		}
	}

	// For other method calls, try to infer based on common patterns
	switch finalMethodName {
	case "connect":
		if strings.Contains(objectPart, "DBI") {
			return "DBI"
		}
	case "decode", "decode_json":
		return "HashRef[Str, Any]"
	case "encode", "encode_json":
		return "Str"
	}

	return "Any"
}

// inferReturnTypeFromCall infers the return type of a function call with sophisticated pattern recognition
func (fa *FlowAnalyzer) inferReturnTypeFromCall(call *ast.CallExpr) string {
	functionName := fa.extractFunctionNameFromCall(call)
	if functionName == "" {
		// Check for method calls (Class->method())
		return fa.inferReturnTypeFromMethodCall(call)
	}

	// Check for built-in Perl functions first
	if builtinType := fa.inferBuiltinFunctionType(functionName, call); builtinType != "" {
		return builtinType
	}

	// Check for common library functions
	if libraryType := fa.inferLibraryFunctionType(functionName, call); libraryType != "" {
		return libraryType
	}

	// Check user-defined function signatures
	if signature, exists := fa.TypeChecker.FunctionTypes[functionName]; exists {
		return signature.ReturnType
	}

	// Perl default: most function calls return strings unless proven otherwise
	return "Str"
}

// inferReturnTypeFromMethodCall infers return types from method calls (Class->method, $obj->method)
func (fa *FlowAnalyzer) inferReturnTypeFromMethodCall(call *ast.CallExpr) string {
	if call == nil || call.Function == nil {
		return "Str"
	}

	// Check if this is a constructor call (Class->new())
	if fa.isConstructorCall(call) {
		className := fa.extractClassNameFromConstructor(call)
		if className != "" {
			return className
		}
	}

	// Check for chained method calls
	if fa.isMethodChain(call) {
		return fa.inferMethodChainType(call)
	}

	// Default for unknown method calls
	return "Str"
}

// inferBuiltinFunctionType infers types for Perl built-in functions using the type registry
func (fa *FlowAnalyzer) inferBuiltinFunctionType(functionName string, call *ast.CallExpr) string {
	fmt.Printf("DEBUG: inferBuiltinFunctionType called for: %s\n", functionName)

	// Use the type registry if available
	if fa.BuiltinTypes != nil {
		// Count parameters to help with overload resolution
		paramCount := 0
		if call != nil && call.Arguments != nil {
			paramCount = len(call.Arguments)
		}

		fmt.Printf("DEBUG: Looking up function '%s' with %d parameters\n", functionName, paramCount)
		if returnType := fa.BuiltinTypes.GetFunctionType(functionName, paramCount); returnType != "" {
			fmt.Printf("DEBUG: Found return type '%s' for function '%s'\n", returnType, functionName)
			return returnType
		} else {
			fmt.Printf("DEBUG: No return type found in registry for function '%s'\n", functionName)
		}
	} else {
		fmt.Printf("DEBUG: BuiltinTypes registry is nil\n")
	}

	// Fallback to hardcoded types for critical built-ins if registry failed
	switch functionName {
	case "ref":
		return "Str"
	case "defined", "exists":
		return "Bool"
	case "keys", "values":
		return "ArrayRef[Str]"
	case "length", "index", "rindex":
		return "Int"
	case "substr", "sprintf":
		return "Str"
	case "int":
		return "Int"
	case "time":
		return "Int"
	default:
		return ""
	}
}

// inferLibraryFunctionType infers types for common library functions
func (fa *FlowAnalyzer) inferLibraryFunctionType(functionName string, call *ast.CallExpr) string {
	switch functionName {
	case "decode_json", "from_json":
		// JSON parsing functions return hash references
		return "HashRef[Str, Any]"
	case "encode_json", "to_json":
		// JSON encoding functions return strings
		return "Str"
	case "selectrow_hashref":
		// Database function returns hash ref with string values
		return "Maybe[HashRef[Str, Str]]"
	case "selectrow_array":
		// Database function returns array ref
		return "Maybe[ArrayRef[Str]]"
	case "selectall_hashref":
		// Database function returns hash of hashes
		return "HashRef[Str, HashRef[Str, Str]]"
	case "selectall_arrayref":
		// Database function returns array of arrays
		return "ArrayRef[ArrayRef[Str]]"
	case "get_user_from_db", "fetch_user", "load_user":
		// Common user fetching patterns - return hash refs
		return "Maybe[HashRef[Str, Str]]"
	case "try_open":
		// Common file opening pattern that can fail
		return "Maybe[FileHandle]"
	case "log_error", "log_info", "log_debug":
		// Logging functions return success boolean
		return "Bool"
	case "send_email", "send_notification":
		// Communication functions return success boolean
		return "Bool"
	}

	return ""
}

// isConstructorCall checks if a call is a constructor call (Class->new())
func (fa *FlowAnalyzer) isConstructorCall(call *ast.CallExpr) bool {
	if call == nil || call.Function == nil {
		return false
	}

	// Check if function is a method call ending with ->new
	functionText := call.Function.Text()
	return strings.Contains(functionText, "->new")
}

// extractClassNameFromConstructor extracts class name from constructor call
func (fa *FlowAnalyzer) extractClassNameFromConstructor(call *ast.CallExpr) string {
	if call == nil || call.Function == nil {
		return ""
	}

	functionText := call.Function.Text()

	// Extract class name from "ClassName->new" pattern
	if strings.Contains(functionText, "->new") {
		parts := strings.Split(functionText, "->")
		if len(parts) > 0 {
			className := strings.TrimSpace(parts[0])
			// Remove any leading $ or @ or % sigils
			className = strings.TrimLeft(className, "$@%")
			return className
		}
	}

	return ""
}

// isMethodChain checks if this is a method chain call
func (fa *FlowAnalyzer) isMethodChain(call *ast.CallExpr) bool {
	if call == nil || call.Function == nil {
		return false
	}

	functionText := call.Function.Text()
	return strings.Contains(functionText, "->") && !strings.Contains(functionText, "->new")
}

// inferMethodChainType infers types from method chains
func (fa *FlowAnalyzer) inferMethodChainType(call *ast.CallExpr) string {
	if call == nil || call.Function == nil {
		return "Str"
	}

	functionText := call.Function.Text()

	// Common method patterns
	if strings.Contains(functionText, "->email") {
		return "Str"
	}
	if strings.Contains(functionText, "->id") {
		return "Str"
	}
	if strings.Contains(functionText, "->name") {
		return "Str"
	}
	if strings.Contains(functionText, "->domain") {
		return "Str"
	}

	// Default for method chains
	return "Str"
}

// Exception Prevention and Safety Analysis

// performSafetyAnalysis performs comprehensive safety analysis to prevent runtime exceptions
func (fa *FlowAnalyzer) performSafetyAnalysis(cfg *ControlFlowGraph) []error {
	var errors []error

	// Analyze each basic block for safety issues
	for _, block := range cfg.Nodes {
		blockErrors := fa.analyzeSafetyInBlock(block)
		errors = append(errors, blockErrors...)
	}

	return errors
}

// analyzeSafetyInBlock analyzes a single basic block for safety issues
func (fa *FlowAnalyzer) analyzeSafetyInBlock(block *BasicBlock) []error {
	var errors []error

	for _, stmt := range block.Statements {
		stmtErrors := fa.analyzeSafetyInStatement(stmt, block.TypeState)
		errors = append(errors, stmtErrors...)
	}

	return errors
}

// analyzeSafetyInStatement analyzes a statement for potential safety issues
func (fa *FlowAnalyzer) analyzeSafetyInStatement(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Recursively analyze all expressions in the statement
	fa.walkExpressions(stmt, func(expr ast.ExpressionNode) {
		exprErrors := fa.analyzeSafetyInExpression(expr, state)
		errors = append(errors, exprErrors...)
	})

	return errors
}

// analyzeSafetyInExpression analyzes an expression for safety issues
func (fa *FlowAnalyzer) analyzeSafetyInExpression(expr ast.ExpressionNode, state *TypeState) []error {
	var errors []error

	if expr == nil {
		return errors
	}

	switch expr.Type() {
	case "hash_ref_expr":
		if hashRef, ok := expr.(*ast.HashRefExpr); ok {
			errors = append(errors, fa.checkHashFieldSafety(hashRef, state)...)
		}
	case "array_ref_expr":
		if arrayRef, ok := expr.(*ast.ArrayRefExpr); ok {
			errors = append(errors, fa.checkArrayAccessSafety(arrayRef, state)...)
		}
	case "variable":
		if varExpr, ok := expr.(*ast.VariableExpr); ok {
			errors = append(errors, fa.checkVariableSafety(varExpr, state)...)
		}
	case "call_expr":
		if callExpr, ok := expr.(*ast.CallExpr); ok {
			errors = append(errors, fa.checkFunctionCallSafety(callExpr, state)...)
		}
	case "binary_expr":
		if binaryExpr, ok := expr.(*ast.BinaryExpr); ok {
			errors = append(errors, fa.checkBinaryExpressionSafety(binaryExpr, state)...)
		}
	case "literal":
		// Handle literal expressions that represent complex expressions
		if literalExpr, ok := expr.(*ast.LiteralExpr); ok {
			errors = append(errors, fa.checkLiteralExpressionSafety(literalExpr, state)...)
		}
	}

	return errors
}

// checkLiteralExpressionSafety checks literal expressions that represent complex access patterns
func (fa *FlowAnalyzer) checkLiteralExpressionSafety(literalExpr *ast.LiteralExpr, state *TypeState) []error {
	var errors []error

	switch literalExpr.Kind {
	case ast.HashAccessLiteral:
		// Parse hash access pattern like $input->{name}
		errors = append(errors, fa.checkHashAccessLiteralSafety(literalExpr, state)...)
	case ast.ArrayAccessLiteral:
		// Parse array access pattern like $data->[0]
		errors = append(errors, fa.checkArrayAccessLiteralSafety(literalExpr, state)...)
	case ast.MethodCallLiteral:
		// Parse method call pattern like $obj->method()
		errors = append(errors, fa.checkMethodCallLiteralSafety(literalExpr, state)...)
	case ast.FunctionCallLiteral:
		// Parse function call pattern like length($string)
		errors = append(errors, fa.checkFunctionCallLiteralSafety(literalExpr, state)...)
	case ast.BinaryExpressionLiteral:
		// Parse binary expressions like $config->{timeout} // 30
		errors = append(errors, fa.checkBinaryExpressionLiteralSafety(literalExpr, state)...)
	}

	return errors
}

// checkHashAccessLiteralSafety checks hash access literals for safety
func (fa *FlowAnalyzer) checkHashAccessLiteralSafety(literalExpr *ast.LiteralExpr, state *TypeState) []error {
	var errors []error

	// Extract hash access pattern from literal value like "$input->{name}"
	value := literalExpr.Value

	// Simple pattern matching for hash access: $var->{key}
	if strings.Contains(value, "->{") && strings.Contains(value, "}") {
		// Extract the key name from the pattern
		start := strings.Index(value, "->{") + 3
		end := strings.LastIndex(value, "}")
		if start < end && start >= 3 {
			keyName := value[start:end]

			// Generate safety error for hash field access without exists() check
			errors = append(errors, &TypeCheckError{
				Message: fmt.Sprintf("unsafe hash field access: %s", keyName),
				Path:    "",
				Line:    literalExpr.Start().Line,
				Column:  literalExpr.Start().Column,
			})
		}
	}

	return errors
}

// checkArrayAccessLiteralSafety checks array access literals for safety
func (fa *FlowAnalyzer) checkArrayAccessLiteralSafety(literalExpr *ast.LiteralExpr, state *TypeState) []error {
	// For now, just return empty - array access is generally safer than hash access
	return []error{}
}

// checkMethodCallLiteralSafety checks method call literals for safety
func (fa *FlowAnalyzer) checkMethodCallLiteralSafety(literalExpr *ast.LiteralExpr, state *TypeState) []error {
	// For now, just return empty - method calls need different safety analysis
	return []error{}
}

// checkFunctionCallLiteralSafety checks function call literals for safety
func (fa *FlowAnalyzer) checkFunctionCallLiteralSafety(literalExpr *ast.LiteralExpr, state *TypeState) []error {
	// For now, just return empty - function calls need different safety analysis
	return []error{}
}

// checkBinaryExpressionLiteralSafety checks binary expression literals for safety
func (fa *FlowAnalyzer) checkBinaryExpressionLiteralSafety(literalExpr *ast.LiteralExpr, state *TypeState) []error {
	// For now, just return empty - binary expressions need more complex analysis
	return []error{}
}

// checkHashFieldSafety checks for unsafe hash field access
func (fa *FlowAnalyzer) checkHashFieldSafety(hashRef *ast.HashRefExpr, state *TypeState) []error {
	var errors []error

	if hashRef == nil || hashRef.Hash == nil || hashRef.Key == nil {
		return errors
	}

	// Extract variable name and field name
	varName := fa.extractVariableFromExpression(hashRef.Hash)
	fieldName := fa.extractConstantValue(hashRef.Key)

	if varName == "" || fieldName == "" {
		return errors // Can't analyze dynamic access
	}

	// Check if this hash access is in a safe Perl idiom context
	if fa.isHashAccessInSafeContext(hashRef) {
		return errors // Safe context, no warning needed
	}

	// Check if this field access was validated with exists()
	if state != nil && state.FieldAccess != nil {
		if fields, exists := state.FieldAccess[varName]; exists {
			if fields[fieldName] {
				return errors // Field access is safe
			}
		}
	}

	// Get the hash type to determine if this is risky
	hashType := fa.getVariableTypeFromState(varName, state)

	// Check if this is a potentially unsafe access pattern
	if fa.isPotentiallyUnsafeHashAccess(hashType, varName, fieldName) {
		errors = append(errors, &TypeCheckError{
			Message: fmt.Sprintf("unsafe hash field access: $%s->{%s}",
				varName, fieldName),
			Path:   "",
			Line:   hashRef.Start().Line,
			Column: hashRef.Start().Column,
		})
	}

	return errors
}

// checkArrayAccessSafety checks for unsafe array access
func (fa *FlowAnalyzer) checkArrayAccessSafety(arrayRef *ast.ArrayRefExpr, state *TypeState) []error {
	var errors []error

	if arrayRef == nil || arrayRef.Array == nil {
		return errors
	}

	// Extract variable name
	varName := fa.extractVariableFromExpression(arrayRef.Array)
	if varName == "" {
		return errors
	}

	// Get the array type
	arrayType := fa.getVariableTypeFromState(varName, state)

	// Check if variable is actually an array reference
	if !fa.isArrayType(arrayType) && arrayType != "Any" {
		errors = append(errors, &TypeCheckError{
			Message: fmt.Sprintf("Array access on non-array variable $%s of type '%s'. "+
				"This may cause 'Not an ARRAY reference' error at runtime. "+
				"Consider checking: ref($%s) eq 'ARRAY'",
				varName, arrayType, varName),
			Path:   "",
			Line:   arrayRef.Start().Line,
			Column: arrayRef.Start().Column,
		})
	}

	return errors
}

// checkVariableSafety checks for unsafe variable usage
func (fa *FlowAnalyzer) checkVariableSafety(varExpr *ast.VariableExpr, state *TypeState) []error {
	var errors []error

	if varExpr == nil {
		return errors
	}

	varName := varExpr.Name

	// Skip if variable name is empty or just a sigil
	if varName == "" || varName == "$" {
		return errors
	}

	varType := fa.getVariableTypeFromState(varName, state)

	// Check for uninitialized variable usage in contexts where it matters
	if fa.isUninitializedVariable(varName, varType, state) {
		// Skip safety warnings if this variable is in a safe context
		if fa.isInSafePerlIdiomContext(varExpr) {
			return errors
		}

		// Skip if this variable was assigned from a safe expression
		if fa.wasAssignedFromSafeExpression(varName, varExpr) {
			return errors
		}

		// Determine the context where this variable is being used
		context := fa.determineUsageContext(varExpr)

		if context == "numeric" {
			errors = append(errors, &TypeCheckError{
				Message: fmt.Sprintf("Variable $%s may be uninitialized in numeric context. "+
					"This may cause 'Use of uninitialized value' warnings at runtime. "+
					"Consider: defined($%s) ? $%s : 0",
					varName, varName, varName),
				Path:   "",
				Line:   varExpr.Start().Line,
				Column: varExpr.Start().Column,
			})
		} else if context == "string" && fa.isStrictStringContext(varExpr) {
			errors = append(errors, &TypeCheckError{
				Message: fmt.Sprintf("Variable $%s may be uninitialized in string context. "+
					"This may cause 'Use of uninitialized value' warnings at runtime. "+
					"Consider: defined($%s) ? $%s : ''",
					varName, varName, varName),
				Path:   "",
				Line:   varExpr.Start().Line,
				Column: varExpr.Start().Column,
			})
		}
	}

	return errors
}

// checkFunctionCallSafety checks for unsafe function call patterns
func (fa *FlowAnalyzer) checkFunctionCallSafety(callExpr *ast.CallExpr, state *TypeState) []error {
	var errors []error

	if callExpr == nil {
		return errors
	}

	functionName := fa.extractFunctionNameFromCall(callExpr)

	// Check for functions that require reference validation
	if fa.requiresReferenceValidation(functionName) {
		for i, arg := range callExpr.Arguments {
			if varExpr, ok := arg.(*ast.VariableExpr); ok {
				varType := fa.getVariableTypeFromState(varExpr.Name, state)
				expectedType := fa.getExpectedArgumentType(functionName, i)

				if !fa.isCompatibleReferenceType(varType, expectedType) {
					errors = append(errors, &TypeCheckError{
						Message: fmt.Sprintf("Function '%s' expects %s but got variable $%s of type '%s'. "+
							"This may cause reference type errors at runtime. "+
							"Consider: ref($%s) eq '%s' ? %s($%s) : undef",
							functionName, expectedType, varExpr.Name, varType,
							varExpr.Name, fa.getReferenceTypeString(expectedType), functionName, varExpr.Name),
						Path:   "",
						Line:   callExpr.Start().Line,
						Column: callExpr.Start().Column,
					})
				}
			}
		}
	}

	return errors
}

// checkBinaryExpressionSafety checks for unsafe binary expressions
func (fa *FlowAnalyzer) checkBinaryExpressionSafety(binaryExpr *ast.BinaryExpr, state *TypeState) []error {
	var errors []error

	if binaryExpr == nil {
		return errors
	}

	operator := binaryExpr.Operator
	leftType := fa.inferTypeFromExpression(binaryExpr.Left)
	rightType := fa.inferTypeFromExpression(binaryExpr.Right)

	// Check for numeric operations on potentially non-numeric types
	if fa.isNumericOperator(operator) {
		if !fa.isNumericType(leftType) && leftType != "Any" && leftType != "Str" {
			if varExpr, ok := binaryExpr.Left.(*ast.VariableExpr); ok {
				errors = append(errors, &TypeCheckError{
					Message: fmt.Sprintf("Numeric operation '%s' on non-numeric variable $%s of type '%s'. "+
						"This may cause unexpected coercion or warnings at runtime.",
						operator, varExpr.Name, leftType),
					Path:   "",
					Line:   binaryExpr.Start().Line,
					Column: binaryExpr.Start().Column,
				})
			}
		}

		if !fa.isNumericType(rightType) && rightType != "Any" && rightType != "Str" {
			if varExpr, ok := binaryExpr.Right.(*ast.VariableExpr); ok {
				errors = append(errors, &TypeCheckError{
					Message: fmt.Sprintf("Numeric operation '%s' on non-numeric variable $%s of type '%s'. "+
						"This may cause unexpected coercion or warnings at runtime.",
						operator, varExpr.Name, rightType),
					Path:   "",
					Line:   binaryExpr.Start().Line,
					Column: binaryExpr.Start().Column,
				})
			}
		}
	}

	return errors
}

// Helper methods for safety analysis

// isPotentiallyUnsafeHashAccess determines if a hash access is potentially unsafe
func (fa *FlowAnalyzer) isPotentiallyUnsafeHashAccess(hashType, varName, fieldName string) bool {
	// Always risky for untyped hashes or Any type
	if hashType == "Any" || hashType == "Str" || hashType == "" {
		return true
	}

	// Database result patterns are risky without exists() checks
	if strings.Contains(hashType, "HashRef") && strings.Contains(hashType, "Str") {
		return true
	}

	// Common risky patterns
	commonRiskyFields := []string{"user_id", "id", "name", "email", "data", "body", "result"}
	for _, riskyField := range commonRiskyFields {
		if fieldName == riskyField {
			return true
		}
	}

	return false
}

// isArrayType checks if a type represents an array
func (fa *FlowAnalyzer) isArrayType(typeStr string) bool {
	return strings.HasPrefix(typeStr, "ArrayRef") || strings.Contains(typeStr, "Array")
}

// isUninitializedVariable checks if a variable might be uninitialized
func (fa *FlowAnalyzer) isUninitializedVariable(varName, varType string, state *TypeState) bool {
	// Check if variable was assigned a value
	if state != nil {
		if currentType, exists := state.VariableTypes[varName]; exists {
			// If the variable has been assigned a concrete type (not Maybe[T]), it's safe
			if !strings.HasPrefix(currentType, "Maybe[") {
				return false
			}
		}
	}

	// Conservative approach: variables named "result" are often safely initialized
	// This is a heuristic to reduce false positives for common patterns
	// But only apply this to variables that aren't assigned from hash/array access
	if varName == "result" || varName == "value" || varName == "output" || varName == "id" {
		// Don't be conservative for variables assigned from hash/array access
		// as those might be genuinely unsafe
		if !fa.isAssignedFromHashOrArrayAccess(varName, state) {
			return false
		}
	}

	// Variables with Maybe types are potentially uninitialized, but only if they
	// haven't been safely assigned a concrete value
	if strings.HasPrefix(varType, "Maybe[") {
		return true
	}

	// If variable has an explicit type annotation, it's likely initialized properly
	if varType != "" && varType != "Str" && !strings.Contains(varType, "|") {
		return false
	}

	// Variables declared with type annotations are considered safe
	if fa.TypeChecker.TypeAnnotations != nil {
		if _, hasAnnotation := fa.TypeChecker.TypeAnnotations[varName]; hasAnnotation {
			return false
		}
	}

	// Variables not in the type state might be uninitialized
	if state == nil {
		return false // Conservative: don't warn without more context
	}

	// Check if variable was assigned a value
	if _, exists := state.VariableTypes[varName]; !exists {
		// Only report for variables that look like they should have been assigned
		// This is a heuristic to reduce false positives
		return fa.looksLikeUnassignedVariable(varName)
	}

	return false
}

// looksLikeUnassignedVariable uses heuristics to determine if a variable looks unassigned
func (fa *FlowAnalyzer) looksLikeUnassignedVariable(varName string) bool {
	// Common patterns that indicate a variable should be checked
	commonUnassignedPatterns := []string{
		"result", "response", "data", "value", "item", "element",
		"user", "config", "content", "output", "input",
	}

	for _, pattern := range commonUnassignedPatterns {
		if strings.Contains(varName, pattern) {
			return true
		}
	}

	// Variables with generic names like "x", "y", "temp" might be unassigned
	if len(varName) <= 2 {
		return true
	}

	// Conservative: don't warn for other patterns to reduce false positives
	return false
}

// determineUsageContext determines how a variable is being used
func (fa *FlowAnalyzer) determineUsageContext(varExpr *ast.VariableExpr) string {
	// This would need to analyze the parent context in a full implementation
	// For now, return a default
	return "string"
}

// isStrictStringContext checks if this is a context where string warnings matter
func (fa *FlowAnalyzer) isStrictStringContext(varExpr *ast.VariableExpr) bool {
	// Check if the variable is used in a safe Perl idiom context
	if fa.isInSafePerlIdiomContext(varExpr) {
		return false // Not strict if it's in a safe context
	}

	// In a full implementation, this would check other parent contexts
	// For now, assume remaining string contexts are strict
	return true
}

// isInSafePerlIdiomContext checks if a variable is used in a context that safely handles undef
func (fa *FlowAnalyzer) isInSafePerlIdiomContext(varExpr *ast.VariableExpr) bool {
	if varExpr == nil || varExpr.Parent() == nil {
		return false
	}

	parent := varExpr.Parent()

	// Check for defined-or operator: $var // default
	if fa.isDefinedOrOperatorContext(parent, varExpr) {
		return true
	}

	// Check for ternary with defined check: defined($var) ? $var : default
	if fa.isSafeTernaryContext(parent, varExpr) {
		return true
	}

	// Check for ||= assignment: $var ||= default
	if fa.isLogicalOrAssignmentContext(parent, varExpr) {
		return true
	}

	// Check for exists/defined guard context
	if fa.isInGuardedContext(varExpr) {
		return true
	}

	return false
}

// isDefinedOrOperatorContext checks if variable is left operand of // operator
func (fa *FlowAnalyzer) isDefinedOrOperatorContext(parent ast.Node, varExpr *ast.VariableExpr) bool {
	if parent == nil {
		return false
	}

	// Check if parent is a binary expression with // operator
	if parent.Type() == "binary_expr" {
		if binaryExpr, ok := parent.(*ast.BinaryExpr); ok {
			if binaryExpr.Operator == "//" {
				// Check if our variable is the left operand
				if leftVar, ok := binaryExpr.Left.(*ast.VariableExpr); ok {
					return leftVar.Name == varExpr.Name
				}
			}
		}
	}

	// Fallback: check parent text for // operator pattern
	parentText := parent.Text()
	varText := varExpr.Text()
	if varText == "" {
		varText = "$" + varExpr.Name
	}

	// Check if parent contains variable followed by //
	return strings.Contains(parentText, varText) && strings.Contains(parentText, "//")
}

// isSafeTernaryContext checks if variable is in a ternary with defined() check
func (fa *FlowAnalyzer) isSafeTernaryContext(parent ast.Node, varExpr *ast.VariableExpr) bool {
	if parent == nil {
		return false
	}

	parentText := parent.Text()
	varText := varExpr.Text()

	// Check for ternary pattern: defined($var) ? $var : default
	return strings.Contains(parentText, "defined("+varText+")") &&
		strings.Contains(parentText, "?") &&
		strings.Contains(parentText, ":")
}

// isLogicalOrAssignmentContext checks for ||= assignment operator
func (fa *FlowAnalyzer) isLogicalOrAssignmentContext(parent ast.Node, varExpr *ast.VariableExpr) bool {
	if parent == nil {
		return false
	}

	parentText := parent.Text()
	varText := varExpr.Text()

	// Check for ||= assignment pattern
	return strings.Contains(parentText, varText) && strings.Contains(parentText, "||=")
}

// isInGuardedContext checks if variable access is protected by early returns or guards
func (fa *FlowAnalyzer) isInGuardedContext(varExpr *ast.VariableExpr) bool {
	// This is a simplified implementation
	// In practice, would need to analyze control flow to see if there are
	// preceding exists/defined checks that protect this access

	// For now, look for simple patterns in the surrounding text
	if varExpr.Parent() != nil {
		grandParent := varExpr.Parent().Parent()
		if grandParent != nil {
			text := grandParent.Text()
			varName := varExpr.Name

			// Look for exists/defined checks before this usage
			if strings.Contains(text, "exists") && strings.Contains(text, varName) {
				return true
			}
			if strings.Contains(text, "defined("+varName+")") {
				return true
			}
		}
	}

	return false
}

// isAssignmentFromSafeExpression checks if an assignment RHS is a safe expression
func (fa *FlowAnalyzer) isAssignmentFromSafeExpression(expr ast.Node) bool {
	if expr == nil {
		return false
	}

	exprText := expr.Text()

	// Check if it's a conditional expression with defined() check
	// Pattern: defined($var) ? $var : default
	if strings.Contains(exprText, "?") && strings.Contains(exprText, ":") {
		if strings.Contains(exprText, "defined(") {
			return true
		}
	}

	// Check for other safe patterns like // operator
	if strings.Contains(exprText, "//") {
		return true
	}

	// Check for ||= style patterns
	if strings.Contains(exprText, "||") {
		return true
	}

	return false
}

// extractMaybeParameter extracts the inner type from Maybe[T]
func extractMaybeParameter(maybeType string) string {
	if strings.HasPrefix(maybeType, "Maybe[") && strings.HasSuffix(maybeType, "]") {
		return maybeType[6 : len(maybeType)-1]
	}
	return ""
}

// wasAssignedFromSafeExpression checks if a variable was assigned from a safe expression
func (fa *FlowAnalyzer) wasAssignedFromSafeExpression(varName string, varExpr *ast.VariableExpr) bool {
	// Look at the surrounding context to find the assignment statement
	current := varExpr.Parent()
	for current != nil && current.Type() != "subroutine" {
		text := current.Text()

		// Look for assignment pattern: my $varName = defined(...) ? ... : ...
		if strings.Contains(text, varName) && strings.Contains(text, "=") {
			if strings.Contains(text, "defined(") && strings.Contains(text, "?") && strings.Contains(text, ":") {
				return true
			}
			// Also check for // operator assignments
			if strings.Contains(text, "//") {
				return true
			}
		}

		current = current.Parent()
	}

	// Also check the immediate context for common safe patterns
	if varExpr.Parent() != nil {
		parentText := varExpr.Parent().Text()
		// Check if this variable reference is in a return statement following a safe assignment
		if strings.Contains(parentText, "return") && strings.Contains(parentText, varName) {
			// Look for the variable assignment in the same subroutine
			if subroutine := fa.findContainingSubroutine(varExpr); subroutine != nil {
				subroutineText := subroutine.Text()
				assignmentPattern := varName + " = defined("
				if strings.Contains(subroutineText, assignmentPattern) {
					return true
				}
				// Also check for ternary pattern without exact variable name
				if strings.Contains(subroutineText, "my $"+varName) &&
					strings.Contains(subroutineText, "defined(") &&
					strings.Contains(subroutineText, "?") &&
					strings.Contains(subroutineText, ":") {
					return true
				}
			}
		}
	}

	return false
}

// findContainingSubroutine finds the subroutine that contains this variable expression
func (fa *FlowAnalyzer) findContainingSubroutine(varExpr *ast.VariableExpr) ast.Node {
	current := varExpr.Parent()
	for current != nil {
		if current.Type() == "subroutine" || current.Type() == "sub_decl" {
			return current
		}
		current = current.Parent()
	}
	return nil
}

// isAssignedFromHashOrArrayAccess checks if a variable was assigned from hash or array access
func (fa *FlowAnalyzer) isAssignedFromHashOrArrayAccess(varName string, state *TypeState) bool {
	// This is a simplified check - in a full implementation we'd track assignment sources
	// For now, check if the variable name suggests it might be from unsafe access
	// Look for common patterns that might indicate hash/array assignment

	// If the variable doesn't exist in the type state, it might be untracked
	if state == nil {
		return false
	}

	// Check if the variable exists in our type state
	if _, exists := state.VariableTypes[varName]; !exists {
		// Variable not in type state might be from an assignment we didn't track properly
		return true // Be conservative and assume it might be from unsafe access
	}

	// For variables like "name", "id" that are common hash field names,
	// assume they might be from hash access unless proven otherwise
	commonHashFields := []string{"name", "id", "user_id", "email", "data"}
	for _, field := range commonHashFields {
		if strings.Contains(varName, field) {
			return true
		}
	}

	return false
}

// isHashAccessInSafeContext checks if hash access is in a safe Perl idiom context
func (fa *FlowAnalyzer) isHashAccessInSafeContext(hashRef *ast.HashRefExpr) bool {
	if hashRef == nil || hashRef.Parent() == nil {
		return false
	}

	parent := hashRef.Parent()

	// Check for defined-or operator: $hash->{key} // default
	if fa.isHashInDefinedOrContext(parent, hashRef) {
		return true
	}

	// Check for ternary with exists check: exists($hash->{key}) ? $hash->{key} : default
	if fa.isHashInSafeTernaryContext(parent, hashRef) {
		return true
	}

	// Check for ||= assignment: $hash->{key} ||= default
	if fa.isHashInLogicalOrAssignmentContext(parent, hashRef) {
		return true
	}

	return false
}

// isHashInDefinedOrContext checks if hash access is left operand of // operator
func (fa *FlowAnalyzer) isHashInDefinedOrContext(parent ast.Node, hashRef *ast.HashRefExpr) bool {
	if parent == nil {
		return false
	}

	// Check if parent is a binary expression with // operator
	if parent.Type() == "binary_expr" {
		if binaryExpr, ok := parent.(*ast.BinaryExpr); ok {
			if binaryExpr.Operator == "//" {
				// Check if our hash access is the left operand
				if leftHash, ok := binaryExpr.Left.(*ast.HashRefExpr); ok {
					// Compare hash access (simplified check)
					return leftHash == hashRef
				}
			}
		}
	}

	// Also check if the parent's parent is a binary expression (for nested expressions)
	if parent.Parent() != nil && parent.Parent().Type() == "binary_expr" {
		return fa.isHashInDefinedOrContext(parent.Parent(), hashRef)
	}

	parentText := parent.Text()
	hashText := hashRef.Text()

	// Check if parent contains both hash access and //
	return strings.Contains(parentText, hashText) && strings.Contains(parentText, "//")
}

// isHashInSafeTernaryContext checks if hash access is in a ternary with exists() check
func (fa *FlowAnalyzer) isHashInSafeTernaryContext(parent ast.Node, hashRef *ast.HashRefExpr) bool {
	if parent == nil {
		return false
	}

	parentText := parent.Text()
	hashText := hashRef.Text()

	// Check for ternary pattern: exists($hash->{key}) ? $hash->{key} : default
	return strings.Contains(parentText, "exists("+hashText+")") &&
		strings.Contains(parentText, "?") &&
		strings.Contains(parentText, ":")
}

// isHashInLogicalOrAssignmentContext checks for ||= assignment with hash access
func (fa *FlowAnalyzer) isHashInLogicalOrAssignmentContext(parent ast.Node, hashRef *ast.HashRefExpr) bool {
	if parent == nil {
		return false
	}

	parentText := parent.Text()
	hashText := hashRef.Text()

	// Check for ||= assignment pattern
	return strings.Contains(parentText, hashText) && strings.Contains(parentText, "||=")
}

// requiresReferenceValidation checks if a function requires reference type validation
func (fa *FlowAnalyzer) requiresReferenceValidation(functionName string) bool {
	referenceRequiredFunctions := []string{"keys", "values", "each", "exists", "delete"}
	for _, fn := range referenceRequiredFunctions {
		if functionName == fn {
			return true
		}
	}
	return false
}

// getExpectedArgumentType returns the expected type for a function argument
func (fa *FlowAnalyzer) getExpectedArgumentType(functionName string, argIndex int) string {
	switch functionName {
	case "keys", "values", "each":
		if argIndex == 0 {
			return "HashRef"
		}
	case "exists", "delete":
		if argIndex == 0 {
			return "HashRef"
		}
	}
	return "Any"
}

// isCompatibleReferenceType checks if actual type is compatible with expected reference type
func (fa *FlowAnalyzer) isCompatibleReferenceType(actualType, expectedType string) bool {
	if expectedType == "Any" {
		return true
	}

	if strings.Contains(actualType, expectedType) {
		return true
	}

	// Allow Any and Str types to pass (Perl's dynamic nature)
	if actualType == "Any" || actualType == "Str" {
		return true
	}

	return false
}

// getReferenceTypeString returns the string representation for ref() checks
func (fa *FlowAnalyzer) getReferenceTypeString(refType string) string {
	switch refType {
	case "HashRef":
		return "HASH"
	case "ArrayRef":
		return "ARRAY"
	case "ScalarRef":
		return "SCALAR"
	default:
		return refType
	}
}

// isNumericOperator checks if an operator is numeric
func (fa *FlowAnalyzer) isNumericOperator(operator string) bool {
	numericOps := []string{"+", "-", "*", "/", "%", "**", "==", "!=", "<", ">", "<=", ">=", "++", "--"}
	for _, op := range numericOps {
		if operator == op {
			return true
		}
	}
	return false
}

// walkExpressions recursively walks all expressions in a statement
func (fa *FlowAnalyzer) walkExpressions(node ast.Node, visitor func(ast.ExpressionNode)) {
	if node == nil {
		return
	}

	// If this node is an expression, visit it
	if expr, ok := node.(ast.ExpressionNode); ok {
		visitor(expr)
	}

	// Recursively walk children
	for _, child := range node.Children() {
		fa.walkExpressions(child, visitor)
	}
}

// Union Type Analysis with Throws[T] Support

// analyzeExceptionFlow analyzes exception flow and infers Throws[T] union types
func (fa *FlowAnalyzer) analyzeExceptionFlow(cfg *ControlFlowGraph) []error {
	var errors []error

	// Track all die statements and their types
	dieStatements := fa.findDieStatements(cfg)

	// For each function, determine if it can throw exceptions
	for _, block := range cfg.Nodes {
		blockErrors := fa.analyzeExceptionsInBlock(block, dieStatements)
		errors = append(errors, blockErrors...)
	}

	return errors
}

// findDieStatements finds all die statements in the control flow graph
func (fa *FlowAnalyzer) findDieStatements(cfg *ControlFlowGraph) map[int][]DieStatement {
	dieStatements := make(map[int][]DieStatement)

	for _, block := range cfg.Nodes {
		for _, stmt := range block.Statements {
			dies := fa.extractDieStatementsFromNode(stmt)
			if len(dies) > 0 {
				dieStatements[block.ID] = append(dieStatements[block.ID], dies...)
			}
		}
	}

	return dieStatements
}

// DieStatement represents a die statement with its exception type
type DieStatement struct {
	Node          ast.Node
	ExceptionType string
	Message       string
	Location      ast.Position
}

// extractDieStatementsFromNode extracts die statements from an AST node
func (fa *FlowAnalyzer) extractDieStatementsFromNode(node ast.Node) []DieStatement {
	var dies []DieStatement

	fa.walkExpressions(node, func(expr ast.ExpressionNode) {
		if callExpr, ok := expr.(*ast.CallExpr); ok {
			functionName := fa.extractFunctionNameFromCall(callExpr)
			if functionName == "die" {
				die := fa.createDieStatement(callExpr)
				dies = append(dies, die)
			}
		}
	})

	return dies
}

// createDieStatement creates a DieStatement from a die function call
func (fa *FlowAnalyzer) createDieStatement(callExpr *ast.CallExpr) DieStatement {
	var message string
	var exceptionType string = "Str" // Default exception type

	// Extract message from die statement
	if len(callExpr.Arguments) > 0 {
		if literal, ok := callExpr.Arguments[0].(*ast.LiteralExpr); ok {
			message = fa.extractConstantValue(literal)
			// Infer exception type from message pattern
			exceptionType = fa.inferExceptionTypeFromMessage(message)
		} else {
			// Dynamic message - infer type from expression
			exceptionType = fa.inferTypeFromExpression(callExpr.Arguments[0])
		}
	}

	return DieStatement{
		Node:          callExpr,
		ExceptionType: exceptionType,
		Message:       message,
		Location:      callExpr.Start(),
	}
}

// inferExceptionTypeFromMessage infers exception type from die message patterns
func (fa *FlowAnalyzer) inferExceptionTypeFromMessage(message string) string {
	// Common exception type patterns
	lowerMessage := strings.ToLower(message)

	if strings.Contains(lowerMessage, "invalid") || strings.Contains(lowerMessage, "bad") {
		return "ValidationError"
	}
	if strings.Contains(lowerMessage, "not found") || strings.Contains(lowerMessage, "missing") {
		return "NotFoundError"
	}
	if strings.Contains(lowerMessage, "permission") || strings.Contains(lowerMessage, "access") {
		return "PermissionError"
	}
	if strings.Contains(lowerMessage, "timeout") {
		return "TimeoutError"
	}
	if strings.Contains(lowerMessage, "network") || strings.Contains(lowerMessage, "connection") {
		return "NetworkError"
	}

	// Default to string exception
	return "Str"
}

// analyzeExceptionsInBlock analyzes exceptions in a single basic block
func (fa *FlowAnalyzer) analyzeExceptionsInBlock(block *BasicBlock, dieStatements map[int][]DieStatement) []error {
	var errors []error

	// Check if this block contains die statements
	if dies, hasDies := dieStatements[block.ID]; hasDies {
		for _, die := range dies {
			// Check if die statement is handled
			if !fa.isDieStatementHandled(die, block) {
				// Create Throws[T] type for unhandled exception
				throwsType := fa.createThrowsType(die.ExceptionType)

				// Add to block's exception types
				if block.TypeState.ExceptionTypes == nil {
					block.TypeState.ExceptionTypes = make(map[string]bool)
				}
				block.TypeState.ExceptionTypes[throwsType] = true

				// Generate error for unhandled exception
				errors = append(errors, &TypeCheckError{
					Message: fmt.Sprintf("Unhandled exception: %s. "+
						"Function may throw %s. "+
						"Consider handling with try/catch or declaring return type as T|%s",
						die.Message, throwsType, throwsType),
					Path:   "",
					Line:   die.Location.Line,
					Column: die.Location.Column,
				})
			}
		}
	}

	return errors
}

// isDieStatementHandled checks if a die statement is properly handled
func (fa *FlowAnalyzer) isDieStatementHandled(die DieStatement, block *BasicBlock) bool {
	// Check if die is in a try block (simplified check)
	// In a full implementation, we'd analyze the control flow
	nodeText := die.Node.Text()
	return strings.Contains(nodeText, "try") || strings.Contains(nodeText, "eval")
}

// createThrowsType creates a Throws[T] type name
func (fa *FlowAnalyzer) createThrowsType(exceptionType string) string {
	return fmt.Sprintf("Throws[%s]", exceptionType)
}

// inferUnionTypeWithExceptions infers union types that include exception types
func (fa *FlowAnalyzer) inferUnionTypeWithExceptions(baseType string, exceptionTypes []string) string {
	if len(exceptionTypes) == 0 {
		return baseType
	}

	// Build union type with base type and all exception types
	unionParts := []string{baseType}
	for _, excType := range exceptionTypes {
		throwsType := fa.createThrowsType(excType)
		unionParts = append(unionParts, throwsType)
	}

	return strings.Join(unionParts, "|")
}

// analyzeExceptionUnions analyzes and validates exception union types
func (fa *FlowAnalyzer) analyzeExceptionUnions(node ast.Node, state *TypeState) []error {
	var errors []error

	// Find all variables with union types that include exceptions
	if state != nil && state.VariableTypes != nil {
		for varName, varType := range state.VariableTypes {
			if fa.isExceptionUnionType(varType) {
				// Check if union type is properly handled
				if !fa.isExceptionUnionHandled(varName, varType, node) {
					errors = append(errors, &TypeCheckError{
						Message: fmt.Sprintf("Unhandled exception union type: Variable $%s has type '%s' "+
							"which includes exception types. Must be handled with pattern matching or try/catch.",
							varName, varType),
						Path:   "",
						Line:   node.Start().Line,
						Column: node.Start().Column,
					})
				}
			}
		}
	}

	return errors
}

// isExceptionUnionType checks if a type is a union that includes exception types
func (fa *FlowAnalyzer) isExceptionUnionType(typeStr string) bool {
	return strings.Contains(typeStr, "Throws[") || strings.Contains(typeStr, "Exception[")
}

// isExceptionUnionHandled checks if an exception union type is properly handled
func (fa *FlowAnalyzer) isExceptionUnionHandled(varName, varType string, context ast.Node) bool {
	// Check if variable is used in a pattern match or try/catch construct
	contextText := context.Text()

	// Look for pattern matching syntax
	if strings.Contains(contextText, "match") && strings.Contains(contextText, varName) {
		return true
	}

	// Look for try/catch usage
	if strings.Contains(contextText, "try") && strings.Contains(contextText, varName) {
		return true
	}

	// Look for explicit exception handling
	if strings.Contains(contextText, "catch") || strings.Contains(contextText, "Exception") {
		return true
	}

	return false
}

// Enhanced function return type inference with exception propagation
func (fa *FlowAnalyzer) inferFunctionReturnTypeWithExceptions(functionNode ast.Node, cfg *ControlFlowGraph) string {
	// Find all possible return types
	returnTypes := fa.collectReturnTypes(functionNode, cfg)

	// Find all exception types that can be thrown
	exceptionTypes := fa.collectExceptionTypes(functionNode, cfg)

	// Combine base return types with exception types
	if len(exceptionTypes) > 0 {
		baseType := fa.mergeReturnTypes(returnTypes)
		return fa.inferUnionTypeWithExceptions(baseType, exceptionTypes)
	}

	return fa.mergeReturnTypes(returnTypes)
}

// collectReturnTypes collects all possible return types from a function
func (fa *FlowAnalyzer) collectReturnTypes(functionNode ast.Node, cfg *ControlFlowGraph) []string {
	var returnTypes []string

	for _, block := range cfg.Nodes {
		for _, stmt := range block.Statements {
			if fa.isReturnStatement(stmt) {
				returnType := fa.inferReturnTypeFromStatement(stmt)
				if returnType != "" {
					returnTypes = append(returnTypes, returnType)
				}
			}
		}
	}

	// Add default return type if no explicit returns
	if len(returnTypes) == 0 {
		returnTypes = append(returnTypes, "()") // Void/empty return
	}

	return returnTypes
}

// collectExceptionTypes collects all exception types that can be thrown
func (fa *FlowAnalyzer) collectExceptionTypes(functionNode ast.Node, cfg *ControlFlowGraph) []string {
	var exceptionTypes []string

	dieStatements := fa.findDieStatements(cfg)
	for _, dies := range dieStatements {
		for _, die := range dies {
			if !fa.isDieStatementHandled(die, nil) {
				exceptionTypes = append(exceptionTypes, die.ExceptionType)
			}
		}
	}

	return fa.uniqueStrings(exceptionTypes)
}

// Helper methods for exception analysis

// isReturnStatement checks if a statement is a return statement
func (fa *FlowAnalyzer) isReturnStatement(stmt ast.Node) bool {
	return stmt.Type() == "return_statement" || strings.Contains(stmt.Text(), "return")
}

// inferReturnTypeFromStatement infers the return type from a return statement
func (fa *FlowAnalyzer) inferReturnTypeFromStatement(stmt ast.Node) string {
	// Extract return expression and infer its type
	// This is simplified - a full implementation would parse the return statement
	stmtText := stmt.Text()
	if strings.Contains(stmtText, "return") {
		// Default return type inference
		if strings.Contains(stmtText, "undef") || strings.Contains(stmtText, "()") {
			return "()"
		}
		// More sophisticated return type inference would go here
		return "Str"
	}
	return ""
}

// mergeReturnTypes merges multiple return types into a single type
func (fa *FlowAnalyzer) mergeReturnTypes(returnTypes []string) string {
	if len(returnTypes) == 0 {
		return "()"
	}
	if len(returnTypes) == 1 {
		return returnTypes[0]
	}

	// Create union type for multiple return types
	unique := fa.uniqueStrings(returnTypes)
	return strings.Join(unique, "|")
}

// uniqueStrings removes duplicates from a string slice
func (fa *FlowAnalyzer) uniqueStrings(strs []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, str := range strs {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}

// extractVariableFromExpression extracts variable name from an expression (for LHS of assignments)
func (fa *FlowAnalyzer) extractVariableFromExpression(expr ast.ExpressionNode) string {
	if expr == nil {
		return ""
	}

	// Handle simple variable expressions
	if varExpr, ok := expr.(*ast.VariableExpr); ok {
		return varExpr.Name
	}

	// For more complex expressions (array/hash refs, etc.), we don't extract the variable
	// This is intentional - we only want simple variable assignments
	return ""
}

// getVariableTypeFromState retrieves the current type of a variable from the type state
func (fa *FlowAnalyzer) getVariableTypeFromState(varName string, state *TypeState) string {
	if state == nil || varName == "" {
		return ""
	}

	// Check refined types first (more specific)
	if state.RefinedTypes != nil {
		if refinedType, exists := state.RefinedTypes[varName]; exists {
			return refinedType
		}
	}

	// Check base variable types
	if state.VariableTypes != nil {
		if varType, exists := state.VariableTypes[varName]; exists {
			return varType
		}
	}

	// Check type checker's type annotations as fallback
	if varType, exists := fa.TypeChecker.TypeAnnotations[varName]; exists {
		return varType
	}

	return ""
}

// extractFunctionNameFromCall extracts the function name from a CallExpr
func (fa *FlowAnalyzer) extractFunctionNameFromCall(call *ast.CallExpr) string {
	if call == nil || call.Function == nil {
		return ""
	}

	// Handle simple function name (identifier)
	if varExpr, ok := call.Function.(*ast.VariableExpr); ok {
		return varExpr.Name
	}

	// For method calls or more complex expressions, we might need more sophisticated handling
	// For now, return empty string to indicate we can't handle it
	return ""
}

// getParameterNamesFromSignature extracts ordered parameter names from a function signature
func (fa *FlowAnalyzer) getParameterNamesFromSignature(signature *FunctionSignature) []string {
	if signature == nil || signature.ParameterTypes == nil {
		return []string{}
	}

	// Since Go maps don't have guaranteed order, we need to extract parameter names
	// in a consistent way. For now, we'll extract them alphabetically.
	// In a real implementation, we'd want to maintain parameter order in the signature.
	var paramNames []string
	for paramName := range signature.ParameterTypes {
		paramNames = append(paramNames, paramName)
	}

	// Sort alphabetically for consistency (this is a limitation of the current design)
	// Ideally, FunctionSignature should maintain parameter order
	sort.Strings(paramNames)

	return paramNames
}

// validateTypeAnnotation validates that a type annotation is valid
func (fa *FlowAnalyzer) validateTypeAnnotation(typeStr string, pos ast.Position) error {
	if typeStr == "" || typeStr == "Any" {
		return nil // Empty or Any type is always valid
	}

	// Use the type hierarchy to validate the type
	if fa.TypeChecker.Hierarchy != nil {
		if err := fa.TypeChecker.Hierarchy.ValidateType(typeStr); err != nil {
			return &TypeCheckError{
				Message: fmt.Sprintf("Invalid type annotation '%s'", typeStr),
				Path:    "",
				Line:    pos.Line,
				Column:  pos.Column,
			}
		}
	}

	return nil
}

// checkTypeCompatibility checks if sourceType is compatible with targetType
func (fa *FlowAnalyzer) checkTypeCompatibility(sourceType, targetType string, pos ast.Position) error {
	if sourceType == "Any" || targetType == "Any" {
		return nil // Any type is compatible with everything
	}

	// Use the type hierarchy to check compatibility
	if fa.TypeChecker.Hierarchy != nil {
		if err := fa.TypeChecker.Hierarchy.CheckTypeCompatibility(sourceType, targetType); err != nil {
			return &TypeCheckError{
				Message: fmt.Sprintf("Cannot assign type '%s' to type '%s'", sourceType, targetType),
				Path:    "",
				Line:    pos.Line,
				Column:  pos.Column,
			}
		}
	}

	return nil
}

// checkCompoundAssignmentCompatibility checks type compatibility for compound assignments (+=, -=, etc.)
func (fa *FlowAnalyzer) checkCompoundAssignmentCompatibility(targetType, sourceType, operator string, pos ast.Position) error {
	// For compound assignments, both operands should be compatible with the operation
	switch operator {
	case "+=", "-=", "*=", "/=", "%=":
		// Numeric operations - both types should be numeric or compatible
		if !fa.isNumericType(targetType) || !fa.isNumericType(sourceType) {
			return &TypeCheckError{
				Message: fmt.Sprintf("Operator '%s' requires numeric types, got '%s' %s '%s'", operator, targetType, operator, sourceType),
				Path:    "",
				Line:    pos.Line,
				Column:  pos.Column,
			}
		}
	case ".=":
		// String concatenation - both types should be string-compatible
		if !fa.isStringCompatibleType(targetType) || !fa.isStringCompatibleType(sourceType) {
			return &TypeCheckError{
				Message: fmt.Sprintf("Operator '%s' requires string-compatible types, got '%s' %s '%s'", operator, targetType, operator, sourceType),
				Path:    "",
				Line:    pos.Line,
				Column:  pos.Column,
			}
		}
	}

	return nil
}

// isNumericType checks if a type is numeric (Int, Num, or compatible)
func (fa *FlowAnalyzer) isNumericType(typeStr string) bool {
	switch typeStr {
	case "Int", "Num", "Any":
		return true
	default:
		return false
	}
}

// isStringCompatibleType checks if a type can be used in string operations
func (fa *FlowAnalyzer) isStringCompatibleType(typeStr string) bool {
	// In Perl, most types can be stringified, but we'll be more strict
	switch typeStr {
	case "Str", "Int", "Num", "Bool", "Any":
		return true
	default:
		return false
	}
}

// Enhanced type inference methods for sophisticated flow analysis

// inferVariableType performs enhanced variable type inference with flow sensitivity
func (fa *FlowAnalyzer) inferVariableType(varExpr *ast.VariableExpr) string {
	if varExpr == nil {
		return "Str"
	}

	varName := varExpr.Name

	// Check current type state first (includes flow-sensitive refinements)
	if fa.TypeChecker.TypeState != nil {
		// Check refined types first (most specific)
		if refinedType, exists := fa.TypeChecker.TypeState.RefinedTypes[varName]; exists {
			return refinedType
		}
		// Check base variable types
		if varType, exists := fa.TypeChecker.TypeState.VariableTypes[varName]; exists {
			return varType
		}
	}

	// Fall back to type annotations
	if varType, exists := fa.TypeChecker.TypeAnnotations[varName]; exists {
		return varType
	}

	// Check if variable was inferred from usage patterns
	if fa.TypeChecker.InferenceEngine != nil {
		inferredTypes := fa.TypeChecker.InferenceEngine.GetAllInferredTypes()
		if inferredType, exists := inferredTypes[varName]; exists {
			return inferredType
		}
	}

	// Perl default: untyped scalars are Str unless proven otherwise
	return "Str"
}

// inferTypeFromBinaryExpression infers types from binary expressions
func (fa *FlowAnalyzer) inferTypeFromBinaryExpression(binaryExpr *ast.BinaryExpr) string {
	if binaryExpr == nil {
		return "Str"
	}

	operator := binaryExpr.Operator
	leftType := fa.inferTypeFromExpression(binaryExpr.Left)
	rightType := fa.inferTypeFromExpression(binaryExpr.Right)

	// Use operator registry if available
	if fa.OperatorTypes != nil {
		resultType := fa.OperatorTypes.GetOperatorType(operator, leftType, rightType)
		if resultType != "Any" {
			// Handle special logic cases that need custom processing
			switch operator {
			case "&&", "||", "and", "or":
				// Logical operations return the type of the last evaluated operand
				// For &&, ||: left operand if falsy, right operand if left is truthy
				// Simplified: return union of both operand types
				if leftType == rightType {
					return leftType
				}
				return fa.createUnionType(leftType, rightType)
			case "//":
				// Defined-or operator returns first defined operand's type
				// If left is Maybe[T], returns T|rightType
				if strings.HasPrefix(leftType, "Maybe[") && strings.HasSuffix(leftType, "]") {
					// Extract T from Maybe[T]
					innerType := leftType[6 : len(leftType)-1]
					if innerType == rightType {
						return rightType
					}
					return fa.createUnionType(innerType, rightType)
				}
				if leftType == rightType {
					return leftType
				}
				return fa.createUnionType(leftType, rightType)
			default:
				return resultType
			}
		}
	}

	// Fallback to hardcoded logic if operator registry is unavailable or incomplete
	switch operator {
	case "+", "-", "*", "/", "%", "**":
		// Arithmetic operations result in numbers
		if fa.isNumericType(leftType) && fa.isNumericType(rightType) {
			// If both operands are Int, result is Int (except division)
			if leftType == "Int" && rightType == "Int" && operator != "/" {
				return "Int"
			}
			return "Num"
		}
		return "Num" // Perl coerces to numbers

	case ".", "x":
		// String operations result in strings
		return "Str"

	case "==", "!=", "<", ">", "<=", ">=", "eq", "ne", "lt", "gt", "le", "ge":
		// Comparison operations result in boolean
		return "Bool"

	case "&&", "||", "and", "or", "not", "!":
		// Logical operations return the type of the last evaluated operand
		// For &&, ||: left operand if falsy, right operand if left is truthy
		// Simplified: return union of both operand types
		if leftType == rightType {
			return leftType
		}
		return fa.createUnionType(leftType, rightType)

	case "=~", "!~":
		// Regex matching returns boolean
		return "Bool"

	case "//":
		// Defined-or operator returns first defined operand's type
		// If left is Maybe[T], returns T|rightType
		if strings.HasPrefix(leftType, "Maybe[") && strings.HasSuffix(leftType, "]") {
			// Extract T from Maybe[T]
			innerType := leftType[6 : len(leftType)-1]
			if innerType == rightType {
				return rightType
			}
			return fa.createUnionType(innerType, rightType)
		}
		if leftType == rightType {
			return leftType
		}
		return fa.createUnionType(leftType, rightType)
	}

	// Default: string concatenation-like behavior
	return "Str"
}

// inferTypeFromHashAccess infers types from hash access with field tracking
func (fa *FlowAnalyzer) inferTypeFromHashAccess(hashRef *ast.HashRefExpr) string {
	if hashRef == nil {
		return "Str"
	}

	// Get the type of the hash variable
	hashType := fa.inferTypeFromExpression(hashRef.Hash)
	keyValue := fa.extractConstantValue(hashRef.Key)

	// Track field access for hash types
	if strings.HasPrefix(hashType, "HashRef[") {
		// Extract value type from HashRef[K, V]
		return fa.extractHashValueType(hashType)
	}

	// For untyped hashes, fields default to Str (database results, JSON, etc.)
	if hashType == "Any" || hashType == "Str" || strings.Contains(hashType, "Hash") {
		// Mark that this field was accessed (for safety analysis)
		if keyValue != "" {
			fa.trackFieldAccess(hashRef.Hash, keyValue)
		}
		return "Str"
	}

	// Default to string for hash access
	return "Str"
}

// inferTypeFromArrayAccess infers types from array access
func (fa *FlowAnalyzer) inferTypeFromArrayAccess(arrayRef *ast.ArrayRefExpr) string {
	if arrayRef == nil {
		return "Str"
	}

	arrayType := fa.inferTypeFromExpression(arrayRef.Array)

	// Extract element type from ArrayRef[T]
	if strings.HasPrefix(arrayType, "ArrayRef[") && strings.HasSuffix(arrayType, "]") {
		elementType := arrayType[9 : len(arrayType)-1]
		return elementType
	}

	// For untyped arrays, elements default to Str
	return "Str"
}

// inferTypeFromConditional infers types from conditional (ternary) expressions
func (fa *FlowAnalyzer) inferTypeFromConditional(conditional *ast.ConditionalExpr) string {
	if conditional == nil {
		return "Str"
	}

	trueType := fa.inferTypeFromExpression(conditional.TrueExpr)
	falseType := fa.inferTypeFromExpression(conditional.FalseExpr)

	// If both branches have the same type, return that type
	if trueType == falseType {
		return trueType
	}

	// Otherwise, create a union type
	return fa.createUnionType(trueType, falseType)
}

// extractHashValueType extracts the value type from HashRef[K, V]
func (fa *FlowAnalyzer) extractHashValueType(hashType string) string {
	if !strings.HasPrefix(hashType, "HashRef[") || !strings.HasSuffix(hashType, "]") {
		return "Str"
	}

	inner := hashType[8 : len(hashType)-1] // Remove "HashRef[" and "]"

	// Find the comma separating key and value types
	// Handle nested types like HashRef[Str, ArrayRef[Int]]
	depth := 0
	commaPos := -1
	for i, ch := range inner {
		switch ch {
		case '[':
			depth++
		case ']':
			depth--
		case ',':
			if depth == 0 {
				commaPos = i
				break
			}
		}
	}

	if commaPos > 0 && commaPos < len(inner)-1 {
		valueType := strings.TrimSpace(inner[commaPos+1:])
		return valueType
	}

	// If no comma found, assume HashRef[V] (single parameter)
	return strings.TrimSpace(inner)
}

// extractConstantValue extracts constant string values from expressions
func (fa *FlowAnalyzer) extractConstantValue(expr ast.ExpressionNode) string {
	if expr == nil {
		return ""
	}

	if literal, ok := expr.(*ast.LiteralExpr); ok {
		if literal.Kind == ast.StringLiteral {
			// Remove quotes from string literal
			value := literal.Value
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				return value[1 : len(value)-1]
			}
			if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
				return value[1 : len(value)-1]
			}
			return value
		}
	}

	return ""
}

// trackFieldAccess tracks hash field access for safety analysis
func (fa *FlowAnalyzer) trackFieldAccess(hashExpr ast.ExpressionNode, fieldName string) {
	if hashExpr == nil || fieldName == "" {
		return
	}

	// Extract variable name from hash expression
	if varExpr, ok := hashExpr.(*ast.VariableExpr); ok {
		varName := varExpr.Name

		// Track in type state for safety analysis
		if fa.TypeChecker.TypeState != nil {
			// Initialize field tracking if needed
			if fa.TypeChecker.TypeState.FieldAccess == nil {
				fa.TypeChecker.TypeState.FieldAccess = make(map[string]map[string]bool)
			}
			if fa.TypeChecker.TypeState.FieldAccess[varName] == nil {
				fa.TypeChecker.TypeState.FieldAccess[varName] = make(map[string]bool)
			}

			// Mark field as accessed
			fa.TypeChecker.TypeState.FieldAccess[varName][fieldName] = true
		}
	}
}

// Debugging and Visualization Capabilities

// FlowDebugger provides debugging and visualization support for flow analysis
type FlowDebugger struct {
	Analyzer    *FlowAnalyzer
	DebugOutput []string
	Enabled     bool
}

// NewFlowDebugger creates a new flow analysis debugger
func NewFlowDebugger(analyzer *FlowAnalyzer) *FlowDebugger {
	return &FlowDebugger{
		Analyzer:    analyzer,
		DebugOutput: make([]string, 0),
		Enabled:     false,
	}
}

// DumpControlFlowGraph exports the control flow graph to DOT format for GraphViz
func (fd *FlowDebugger) DumpControlFlowGraph(cfg *ControlFlowGraph, outputPath string) error {
	dotContent := fd.generateDotGraph(cfg)

	// Write to file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CFG output file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("failed to write CFG content: %w", err)
	}

	fd.logDebug(fmt.Sprintf("Control flow graph exported to %s", outputPath))
	return nil
}

// generateDotGraph generates DOT format content for the control flow graph
func (fd *FlowDebugger) generateDotGraph(cfg *ControlFlowGraph) string {
	var dot strings.Builder

	dot.WriteString("digraph CFG {\n")
	dot.WriteString("  rankdir=TB;\n")
	dot.WriteString("  node [shape=box, style=filled];\n\n")

	// Generate nodes
	for _, block := range cfg.Nodes {
		label := fd.generateBlockLabel(block)
		color := fd.getBlockColor(block)

		dot.WriteString(fmt.Sprintf("  block_%d [label=\"%s\", fillcolor=\"%s\"];\n",
			block.ID, label, color))
	}

	dot.WriteString("\n")

	// Generate edges
	for _, edge := range cfg.Edges {
		edgeLabel := fd.getEdgeLabel(edge)
		edgeColor := fd.getEdgeColor(edge)

		dot.WriteString(fmt.Sprintf("  block_%d -> block_%d [label=\"%s\", color=\"%s\"];\n",
			edge.From.ID, edge.To.ID, edgeLabel, edgeColor))
	}

	dot.WriteString("}\n")
	return dot.String()
}

// generateBlockLabel creates a label for a basic block
func (fd *FlowDebugger) generateBlockLabel(block *BasicBlock) string {
	var label strings.Builder

	label.WriteString(fmt.Sprintf("Block %d\\n", block.ID))

	// Add type state summary
	if block.TypeState != nil {
		label.WriteString("Types:\\n")
		for varName, varType := range block.TypeState.VariableTypes {
			label.WriteString(fmt.Sprintf("  $%s: %s\\n", varName, varType))
		}

		// Add exception types if any
		if len(block.TypeState.ExceptionTypes) > 0 {
			label.WriteString("Exceptions:\\n")
			for excType := range block.TypeState.ExceptionTypes {
				label.WriteString(fmt.Sprintf("  %s\\n", excType))
			}
		}
	}

	// Add statements summary
	if len(block.Statements) > 0 {
		label.WriteString("Statements:\\n")
		for i, stmt := range block.Statements {
			if i >= 3 { // Limit to first 3 statements
				label.WriteString("  ...\\n")
				break
			}
			stmtText := stmt.Text()
			if len(stmtText) > 30 {
				stmtText = stmtText[:27] + "..."
			}
			// Escape quotes for DOT format
			stmtText = strings.ReplaceAll(stmtText, "\"", "\\\"")
			label.WriteString(fmt.Sprintf("  %s\\n", stmtText))
		}
	}

	return label.String()
}

// getBlockColor determines the color for a basic block based on its properties
func (fd *FlowDebugger) getBlockColor(block *BasicBlock) string {
	// Entry block
	if block.ID == 0 {
		return "lightgreen"
	}

	// Exit block (no successors)
	if len(block.Successors) == 0 {
		return "lightcoral"
	}

	// Blocks with exceptions
	if block.TypeState != nil && len(block.TypeState.ExceptionTypes) > 0 {
		return "orange"
	}

	// Conditional blocks (multiple successors)
	if len(block.Successors) > 1 {
		return "lightblue"
	}

	// Regular blocks
	return "lightgray"
}

// getEdgeLabel creates a label for a control flow edge
func (fd *FlowDebugger) getEdgeLabel(edge *FlowEdge) string {
	switch edge.EdgeType {
	case ConditionalTrueEdge:
		return "true"
	case ConditionalFalseEdge:
		return "false"
	case LoopBackEdge:
		return "loop"
	case BreakEdge:
		return "break"
	case ContinueEdge:
		return "continue"
	case ExceptionEdge:
		return "exception"
	default:
		return ""
	}
}

// getEdgeColor determines the color for a control flow edge
func (fd *FlowDebugger) getEdgeColor(edge *FlowEdge) string {
	switch edge.EdgeType {
	case ConditionalTrueEdge:
		return "green"
	case ConditionalFalseEdge:
		return "red"
	case LoopBackEdge:
		return "blue"
	case BreakEdge:
		return "orange"
	case ContinueEdge:
		return "purple"
	case ExceptionEdge:
		return "red"
	default:
		return "black"
	}
}

// logDebug logs debug information if debugging is enabled
func (fd *FlowDebugger) logDebug(message string) {
	if fd.Enabled {
		timestamp := time.Now().Format("15:04:05.000")
		logMessage := fmt.Sprintf("[%s] %s", timestamp, message)
		fd.DebugOutput = append(fd.DebugOutput, logMessage)
		// Debug output collected in DebugOutput slice
	}
}

// TraceTypeRefinement logs type refinement events
func (fd *FlowDebugger) TraceTypeRefinement(varName, oldType, newType, reason string) {
	if fd.Enabled {
		fd.logDebug(fmt.Sprintf("Type refinement: $%s: %s → %s (reason: %s)", varName, oldType, newType, reason))
	}
}
