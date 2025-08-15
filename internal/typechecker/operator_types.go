// ABOUTME: Parser and loader for Perl binary operator type definitions
// ABOUTME: Replaces hardcoded operator type logic with declarative type definition files

package typechecker

import (
	"embed"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/parser"
)

//go:embed operators.ptd
var operatorTypesFile embed.FS

// OperatorSignature represents a binary operator's type signature
type OperatorSignature struct {
	Operator   string // The Perl operator symbol (+, -, eq, etc.)
	LeftType   string // Type of left operand
	RightType  string // Type of right operand
	ResultType string // Result type
	Context    string // "scalar", "list", or ""
}

// OperatorTypeRegistry manages binary operator type signatures
type OperatorTypeRegistry struct {
	operators map[string][]*OperatorSignature // operator symbol -> possible signatures
}

// NewOperatorTypeRegistry creates and loads the operator type registry
func NewOperatorTypeRegistry() (*OperatorTypeRegistry, error) {
	registry := &OperatorTypeRegistry{
		operators: make(map[string][]*OperatorSignature),
	}

	// Load operator types from embedded file
	content, err := operatorTypesFile.ReadFile("operators.ptd")
	if err != nil {
		return nil, fmt.Errorf("failed to read operators.ptd: %v", err)
	}

	if err := registry.parseOperatorDefinitions(string(content)); err != nil {
		return nil, fmt.Errorf("failed to parse operator definitions: %v", err)
	}

	return registry, nil
}

// parseOperatorDefinitions parses .ptd file content and extracts operator signatures
func (r *OperatorTypeRegistry) parseOperatorDefinitions(content string) error {
	// Create a parser to parse the .ptd file
	p, err := parser.NewParser()
	if err != nil {
		return fmt.Errorf("failed to create parser: %v", err)
	}

	// Parse the content as Perl code
	astResult, err := p.ParseString(content)
	if err != nil {
		return fmt.Errorf("failed to parse operator definitions: %v", err)
	}

	// Extract operator signatures from the AST
	if err := r.extractOperatorSignaturesFromAST(astResult); err != nil {
		return fmt.Errorf("failed to extract operator signatures: %v", err)
	}

	return nil
}

// extractOperatorSignaturesFromAST extracts operator signatures from the parsed AST
func (r *OperatorTypeRegistry) extractOperatorSignaturesFromAST(astResult *ast.AST) error {
	if astResult == nil || astResult.Root == nil {
		return fmt.Errorf("empty AST")
	}

	// Walk through the AST looking for subroutine definitions
	for _, child := range astResult.Root.Children() {
		if err := r.extractOperatorSignatures(child); err != nil {
			return err
		}
	}

	return nil
}

// extractOperatorSignatures extracts operator signatures from AST
func (r *OperatorTypeRegistry) extractOperatorSignatures(node ast.Node) error {
	switch n := node.(type) {
	case *ast.ProgramStmt:
		for _, stmt := range n.LogicalStatements() {
			if err := r.extractOperatorSignatures(stmt); err != nil {
				return err
			}
		}
	case *ast.SubDecl:
		if err := r.processOperatorDeclaration(n); err != nil {
			return err
		}
	}
	return nil
}

// processOperatorDeclaration processes a single operator function declaration
func (r *OperatorTypeRegistry) processOperatorDeclaration(sub *ast.SubDecl) error {
	if sub == nil || sub.Name == "" {
		return nil
	}

	// Skip non-operator functions
	if !strings.HasPrefix(sub.Name, "operator_") {
		return nil
	}

	// Extract operator symbol from function name
	operatorFunc := sub.Name
	operator, err := r.mapOperatorFunctionToSymbol(operatorFunc)
	if err != nil {
		return fmt.Errorf("unknown operator function %s: %v", operatorFunc, err)
	}

	// Extract parameter types
	var leftType, rightType string
	params := sub.Parameters()
	if len(params) >= 1 {
		leftType = r.extractParameterType(params[0])
	}
	if len(params) >= 2 {
		rightType = r.extractParameterType(params[1])
	}

	// For unary operators, rightType remains empty
	resultType := "Any" // Default
	if sub.ReturnType != nil {
		resultType = r.extractTypeFromAST(sub.ReturnType)
	}

	signature := &OperatorSignature{
		Operator:   operator,
		LeftType:   leftType,
		RightType:  rightType,
		ResultType: resultType,
		Context:    "scalar", // Most operators work in scalar context
	}

	// Add to registry
	r.operators[operator] = append(r.operators[operator], signature)

	return nil
}

// mapOperatorFunctionToSymbol maps operator function names to Perl operator symbols
func (r *OperatorTypeRegistry) mapOperatorFunctionToSymbol(funcName string) (string, error) {
	operatorMap := map[string]string{
		// Arithmetic
		"operator_add": "+",
		"operator_sub": "-",
		"operator_mul": "*",
		"operator_div": "/",
		"operator_mod": "%",
		"operator_pow": "**",

		// String
		"operator_concat": ".",
		"operator_repeat": "x",

		// Numeric comparison
		"operator_num_eq": "==",
		"operator_num_ne": "!=",
		"operator_num_lt": "<",
		"operator_num_gt": ">",
		"operator_num_le": "<=",
		"operator_num_ge": ">=",

		// String comparison
		"operator_str_eq": "eq",
		"operator_str_ne": "ne",
		"operator_str_lt": "lt",
		"operator_str_gt": "gt",
		"operator_str_le": "le",
		"operator_str_ge": "ge",

		// Logical
		"operator_and":        "&&",
		"operator_or":         "||",
		"operator_and_word":   "and",
		"operator_or_word":    "or",
		"operator_not":        "not",
		"operator_not_symbol": "!",

		// Regex
		"operator_match":     "=~",
		"operator_not_match": "!~",

		// Special
		"operator_defined_or":  "//",
		"operator_smart_match": "~~",

		// Bitwise
		"operator_bit_and":        "&",
		"operator_bit_or":         "|",
		"operator_bit_xor":        "^",
		"operator_bit_lshift":     "<<",
		"operator_bit_rshift":     ">>",
		"operator_bit_complement": "~",

		// Assignment
		"operator_assign":            "=",
		"operator_add_assign":        "+=",
		"operator_sub_assign":        "-=",
		"operator_mul_assign":        "*=",
		"operator_div_assign":        "/=",
		"operator_mod_assign":        "%=",
		"operator_pow_assign":        "**=",
		"operator_concat_assign":     ".=",
		"operator_repeat_assign":     "x=",
		"operator_and_assign":        "&=",
		"operator_or_assign":         "|=",
		"operator_xor_assign":        "^=",
		"operator_lshift_assign":     "<<=",
		"operator_rshift_assign":     ">>=",
		"operator_defined_or_assign": "//=",

		// Range
		"operator_range":     "..",
		"operator_flip_flop": "..", // Same symbol, different context

		// Comma
		"operator_comma":     ",",
		"operator_fat_comma": "=>",

		// Reference
		"operator_ref":          "\\",
		"operator_ref_scalar":   "\\$",
		"operator_ref_array":    "\\@",
		"operator_ref_hash":     "\\%",
		"operator_ref_code":     "\\&",
		"operator_deref_scalar": "${}",
		"operator_deref_array":  "@{}",
		"operator_deref_hash":   "%{}",
		"operator_deref_code":   "&{}",

		// File test operators
		"operator_file_test": "-test", // General file test

		// Specific file test operators (mapped to special symbols for now)
		"file_test_r": "-r",
		"file_test_w": "-w",
		"file_test_x": "-x",
		"file_test_e": "-e",
		"file_test_f": "-f",
		"file_test_d": "-d",
		"file_test_l": "-l",
		"file_test_s": "-s",
		"file_test_z": "-z",
		"file_test_p": "-p",
		"file_test_S": "-S",
		"file_test_b": "-b",
		"file_test_c": "-c",
		"file_test_t": "-t",
		"file_test_u": "-u",
		"file_test_g": "-g",
		"file_test_k": "-k",
		"file_test_o": "-o",
		"file_test_O": "-O",
		"file_test_R": "-R",
		"file_test_W": "-W",
		"file_test_X": "-X",
		"file_test_T": "-T",
		"file_test_B": "-B",
		"file_test_M": "-M",
		"file_test_A": "-A",
		"file_test_C": "-C",
	}

	if symbol, exists := operatorMap[funcName]; exists {
		return symbol, nil
	}

	return "", fmt.Errorf("unknown operator function: %s", funcName)
}

// extractParameterType extracts type from parameter declaration
func (r *OperatorTypeRegistry) extractParameterType(param *ast.Parameter) string {
	if param == nil || param.TypeExpr == nil {
		return "Any"
	}
	return r.extractTypeFromAST(param.TypeExpr)
}

// extractTypeFromAST extracts type string from AST type node
func (r *OperatorTypeRegistry) extractTypeFromAST(typeNode ast.Node) string {
	if typeNode == nil {
		return "Any"
	}

	switch t := typeNode.(type) {
	case *ast.TypeExpression:
		return t.Name
	default:
		return "Any"
	}
}

// GetOperatorType returns the result type for a binary operator with given operand types
func (r *OperatorTypeRegistry) GetOperatorType(operator, leftType, rightType string) string {
	signatures, exists := r.operators[operator]
	if !exists {
		return "Any" // Unknown operator
	}

	// Find best matching signature
	for _, sig := range signatures {
		if r.typesMatch(sig.LeftType, leftType) && r.typesMatch(sig.RightType, rightType) {
			return sig.ResultType
		}
	}

	// Try with type coercion (Any matches anything)
	for _, sig := range signatures {
		if (sig.LeftType == "Any" || leftType == "Any") &&
			(sig.RightType == "Any" || rightType == "Any") {
			return sig.ResultType
		}
	}

	return "Any" // No match found
}

// typesMatch checks if two types are compatible
func (r *OperatorTypeRegistry) typesMatch(expected, actual string) bool {
	if expected == actual {
		return true
	}
	if expected == "Any" || actual == "Any" {
		return true
	}

	// Handle numeric type compatibility
	if (expected == "Num" && actual == "Int") || (expected == "Int" && actual == "Num") {
		return true
	}

	return false
}

// IsKnownOperator checks if an operator is in the registry
func (r *OperatorTypeRegistry) IsKnownOperator(operator string) bool {
	_, exists := r.operators[operator]
	return exists
}

// ListAllOperators returns all registered operators
func (r *OperatorTypeRegistry) ListAllOperators() []string {
	operators := make([]string, 0, len(r.operators))
	for op := range r.operators {
		operators = append(operators, op)
	}
	return operators
}

// GetOperatorSignatures returns all signatures for a given operator
func (r *OperatorTypeRegistry) GetOperatorSignatures(operator string) []*OperatorSignature {
	if signatures, exists := r.operators[operator]; exists {
		return signatures
	}
	return nil
}

// GetTypeSignatureString returns a human-readable signature string
func (sig *OperatorSignature) GetTypeSignatureString() string {
	if sig.RightType == "" {
		// Unary operator
		return fmt.Sprintf("%s %s -> %s", sig.LeftType, sig.Operator, sig.ResultType)
	}
	// Binary operator
	return fmt.Sprintf("%s %s %s -> %s", sig.LeftType, sig.Operator, sig.RightType, sig.ResultType)
}
