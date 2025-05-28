// ABOUTME: Expression AST node types for Perl expressions
// ABOUTME: Specific node types for different kinds of expressions

package ast

// ExpressionNode represents any kind of expression
type ExpressionNode interface {
	Node
	IsExpression() bool
}

// LiteralExpr represents literal values (strings, numbers, etc.)
type LiteralExpr struct {
	*BaseNode
	Value string
	Kind  LiteralKind
}

// LiteralKind represents the type of literal
type LiteralKind int

const (
	StringLiteral LiteralKind = iota
	NumberLiteral
	BooleanLiteral
	UndefLiteral
	RegexLiteral
)

// NewLiteralExpr creates a new literal expression
func NewLiteralExpr(value string, kind LiteralKind, start, end Position) *LiteralExpr {
	return &LiteralExpr{
		BaseNode: NewBaseNode("literal", start, end),
		Value:    value,
		Kind:     kind,
	}
}

// IsExpression implements ExpressionNode interface
func (le *LiteralExpr) IsExpression() bool {
	return true
}

// VariableExpr represents variable references ($var, @array, %hash)
type VariableExpr struct {
	*BaseNode
	Name  string
	Sigil string // $, @, %
}

// NewVariableExpr creates a new variable expression
func NewVariableExpr(name, sigil string, start, end Position) *VariableExpr {
	return &VariableExpr{
		BaseNode: NewBaseNode("variable", start, end),
		Name:     name,
		Sigil:    sigil,
	}
}

// IsExpression implements ExpressionNode interface
func (ve *VariableExpr) IsExpression() bool {
	return true
}

// FullName returns the full variable name with sigil
func (ve *VariableExpr) FullName() string {
	return ve.Sigil + ve.Name
}

// BinaryExpr represents binary operations (a + b, a == b, etc.)
type BinaryExpr struct {
	*BaseNode
	Left     ExpressionNode
	Right    ExpressionNode
	Operator string
}

// NewBinaryExpr creates a new binary expression
func NewBinaryExpr(left, right ExpressionNode, operator string, start, end Position) *BinaryExpr {
	expr := &BinaryExpr{
		BaseNode: NewBaseNode("binary_expr", start, end),
		Left:     left,
		Right:    right,
		Operator: operator,
	}

	// Add children
	if left != nil {
		expr.AddChild(left)
	}
	if right != nil {
		expr.AddChild(right)
	}

	return expr
}

// IsExpression implements ExpressionNode interface
func (be *BinaryExpr) IsExpression() bool {
	return true
}

// UnaryExpr represents unary operations (!expr, -expr, etc.)
type UnaryExpr struct {
	*BaseNode
	Operand  ExpressionNode
	Operator string
	Prefix   bool // true for prefix operators, false for postfix
}

// NewUnaryExpr creates a new unary expression
func NewUnaryExpr(operand ExpressionNode, operator string, prefix bool, start, end Position) *UnaryExpr {
	expr := &UnaryExpr{
		BaseNode: NewBaseNode("unary_expr", start, end),
		Operand:  operand,
		Operator: operator,
		Prefix:   prefix,
	}

	if operand != nil {
		expr.AddChild(operand)
	}

	return expr
}

// IsExpression implements ExpressionNode interface
func (ue *UnaryExpr) IsExpression() bool {
	return true
}

// CallExpr represents function/method calls
type CallExpr struct {
	*BaseNode
	Function  ExpressionNode   // The function being called
	Arguments []ExpressionNode // Arguments to the function
	Method    bool             // true if this is a method call
}

// NewCallExpr creates a new call expression
func NewCallExpr(function ExpressionNode, args []ExpressionNode, isMethod bool, start, end Position) *CallExpr {
	expr := &CallExpr{
		BaseNode:  NewBaseNode("call_expr", start, end),
		Function:  function,
		Arguments: args,
		Method:    isMethod,
	}

	// Add children
	if function != nil {
		expr.AddChild(function)
	}
	for _, arg := range args {
		if arg != nil {
			expr.AddChild(arg)
		}
	}

	return expr
}

// IsExpression implements ExpressionNode interface
func (ce *CallExpr) IsExpression() bool {
	return true
}

// ArrayRefExpr represents array references ($array[index])
type ArrayRefExpr struct {
	*BaseNode
	Array ExpressionNode
	Index ExpressionNode
}

// NewArrayRefExpr creates a new array reference expression
func NewArrayRefExpr(array, index ExpressionNode, start, end Position) *ArrayRefExpr {
	expr := &ArrayRefExpr{
		BaseNode: NewBaseNode("array_ref", start, end),
		Array:    array,
		Index:    index,
	}

	if array != nil {
		expr.AddChild(array)
	}
	if index != nil {
		expr.AddChild(index)
	}

	return expr
}

// IsExpression implements ExpressionNode interface
func (are *ArrayRefExpr) IsExpression() bool {
	return true
}

// HashRefExpr represents hash references ($hash{key})
type HashRefExpr struct {
	*BaseNode
	Hash ExpressionNode
	Key  ExpressionNode
}

// NewHashRefExpr creates a new hash reference expression
func NewHashRefExpr(hash, key ExpressionNode, start, end Position) *HashRefExpr {
	expr := &HashRefExpr{
		BaseNode: NewBaseNode("hash_ref", start, end),
		Hash:     hash,
		Key:      key,
	}

	if hash != nil {
		expr.AddChild(hash)
	}
	if key != nil {
		expr.AddChild(key)
	}

	return expr
}

// IsExpression implements ExpressionNode interface
func (hre *HashRefExpr) IsExpression() bool {
	return true
}

// TypeAssertionExpr represents type assertions ($value as Type)
type TypeAssertionExpr struct {
	*BaseNode
	Expression ExpressionNode
	TargetType *TypeExpression
}

// NewTypeAssertionExpr creates a new type assertion expression
func NewTypeAssertionExpr(expr ExpressionNode, targetType *TypeExpression, start, end Position) *TypeAssertionExpr {
	assertion := &TypeAssertionExpr{
		BaseNode:   NewBaseNode("type_assertion", start, end),
		Expression: expr,
		TargetType: targetType,
	}

	if expr != nil {
		assertion.AddChild(expr)
	}

	return assertion
}

// IsExpression implements ExpressionNode interface
func (tae *TypeAssertionExpr) IsExpression() bool {
	return true
}

// ConditionalExpr represents ternary conditional expressions (cond ? true : false)
type ConditionalExpr struct {
	*BaseNode
	Condition ExpressionNode
	TrueExpr  ExpressionNode
	FalseExpr ExpressionNode
}

// NewConditionalExpr creates a new conditional expression
func NewConditionalExpr(condition, trueExpr, falseExpr ExpressionNode, start, end Position) *ConditionalExpr {
	expr := &ConditionalExpr{
		BaseNode:  NewBaseNode("conditional_expr", start, end),
		Condition: condition,
		TrueExpr:  trueExpr,
		FalseExpr: falseExpr,
	}

	if condition != nil {
		expr.AddChild(condition)
	}
	if trueExpr != nil {
		expr.AddChild(trueExpr)
	}
	if falseExpr != nil {
		expr.AddChild(falseExpr)
	}

	return expr
}

// IsExpression implements ExpressionNode interface
func (ce *ConditionalExpr) IsExpression() bool {
	return true
}

// ListExpr represents list expressions (1, 2, 3)
type ListExpr struct {
	*BaseNode
	Elements []ExpressionNode
}

// NewListExpr creates a new list expression
func NewListExpr(elements []ExpressionNode, start, end Position) *ListExpr {
	expr := &ListExpr{
		BaseNode: NewBaseNode("list_expr", start, end),
		Elements: elements,
	}

	for _, element := range elements {
		if element != nil {
			expr.AddChild(element)
		}
	}

	return expr
}

// IsExpression implements ExpressionNode interface
func (le *ListExpr) IsExpression() bool {
	return true
}

// AssignmentExpr represents assignment expressions ($a = $b, $a += $b, etc.)
type AssignmentExpr struct {
	*BaseNode
	Left     ExpressionNode
	Right    ExpressionNode
	Operator string // =, +=, -=, *=, etc.
}

// NewAssignmentExpr creates a new assignment expression
func NewAssignmentExpr(left, right ExpressionNode, operator string, start, end Position) *AssignmentExpr {
	expr := &AssignmentExpr{
		BaseNode: NewBaseNode("assignment_expr", start, end),
		Left:     left,
		Right:    right,
		Operator: operator,
	}

	// Add children
	if left != nil {
		expr.AddChild(left)
	}
	if right != nil {
		expr.AddChild(right)
	}

	return expr
}

// IsExpression implements ExpressionNode interface
func (ae *AssignmentExpr) IsExpression() bool {
	return true
}
