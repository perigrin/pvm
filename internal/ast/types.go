// ABOUTME: Core AST type definitions consolidating scattered node types
// ABOUTME: Provides unified interfaces following TypeScript-Go astnav patterns

//go:generate stringer -type=AnnotationKind -output=annotation_kind_string.go

package ast

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Position represents a position in source code
// Consolidated from multiple package definitions
type Position struct {
	Line   int
	Column int
	Offset int
}

// Node represents a node in the Abstract Syntax Tree
// Unified interface from parser and treesitter packages
type Node interface {
	// Type returns the syntactic type of the node
	Type() string

	// Start returns the start position of the node in source
	Start() Position

	// End returns the end position of the node in source
	End() Position

	// Children returns the child nodes
	Children() []Node

	// Text returns the source text covered by this node
	Text() string

	// Parent returns the parent node (nil for root)
	Parent() Node

	// SetParent sets the parent node (for tree construction)
	SetParent(parent Node)
}

// AST represents the Abstract Syntax Tree of a parsed file
// Consolidated from parser and treesitter AST types
type AST struct {
	// Path is the source file path
	Path string

	// Root is the root node of the syntax tree
	Root Node

	// TypeAnnotations contains all type annotations found
	TypeAnnotations []*TypeAnnotation

	// Errors contains any parse errors encountered
	Errors []error

	// Source contains the original source text
	Source string
}

// GetPath returns the source file path (compiler.AST interface)
func (a *AST) GetPath() string {
	return a.Path
}

// IsValid returns true if the AST is valid for compilation (compiler.AST interface)
func (a *AST) IsValid() bool {
	return len(a.Errors) == 0
}

// GetContent returns the original source content (compiler.AST interface)
func (a *AST) GetContent() (string, error) {
	if a.Source != "" {
		return a.Source, nil
	}
	if a.Path == "" {
		return "", fmt.Errorf("AST has no source content or file path")
	}
	// Could read from file here if needed, but typically Source should be populated
	return "", fmt.Errorf("AST source content not available")
}

// GetRootNode returns the root AST node (compiler.AST interface)
func (a *AST) GetRootNode() (Node, error) {
	if a.Root == nil {
		return nil, fmt.Errorf("AST has no root node")
	}
	return a.Root, nil
}

// AST implements Node interface to allow navigation
func (a *AST) Type() string {
	return "ast"
}

func (a *AST) Start() Position {
	if a.Root != nil {
		return a.Root.Start()
	}
	return Position{Line: 1, Column: 1, Offset: 0}
}

func (a *AST) End() Position {
	if a.Root != nil {
		return a.Root.End()
	}
	return Position{Line: 1, Column: 1, Offset: 0}
}

func (a *AST) Children() []Node {
	if a.Root != nil {
		return []Node{a.Root}
	}
	return []Node{}
}

func (a *AST) Text() string {
	return a.Source
}

func (a *AST) Parent() Node {
	return nil // AST is the root, has no parent
}

func (a *AST) SetParent(parent Node) {
	// AST is the root, cannot set parent
}

// MarshalJSON implements json.Marshaler interface for AST
func (a *AST) MarshalJSON() ([]byte, error) {
	type astJSON struct {
		Path            string                `json:"path"`
		Root            *nodeJSON             `json:"root,omitempty"`
		TypeAnnotations []*typeAnnotationJSON `json:"type_annotations"`
		Errors          []string              `json:"errors"`
		SourceLength    int                   `json:"source_length"`
	}

	// Convert errors to strings
	var errorStrings []string
	if len(a.Errors) == 0 {
		errorStrings = []string{}
	} else {
		for _, err := range a.Errors {
			errorStrings = append(errorStrings, err.Error())
		}
	}

	// Convert type annotations
	var typeAnnotations []*typeAnnotationJSON
	if len(a.TypeAnnotations) == 0 {
		typeAnnotations = []*typeAnnotationJSON{}
	} else {
		for _, ta := range a.TypeAnnotations {
			typeAnnotations = append(typeAnnotations, &typeAnnotationJSON{
				AnnotatedItem:  ta.AnnotatedItem,
				TypeExpression: ta.TypeExpression,
				Position:       ta.Pos,
				Kind:           ta.Kind.String(),
			})
		}
	}

	// Convert root node
	var rootJSON *nodeJSON
	if a.Root != nil {
		rootJSON = nodeToJSON(a.Root)
	}

	jsonBytes, err := json.Marshal(astJSON{
		Path:            a.Path,
		Root:            rootJSON,
		TypeAnnotations: typeAnnotations,
		Errors:          errorStrings,
		SourceLength:    len(a.Source),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AST to JSON: %w", err)
	}
	return jsonBytes, nil
}

// String returns a string representation of the AST for baseline testing
func (a *AST) String() string {
	var builder strings.Builder

	builder.WriteString("AST {\n")
	if a.Path == "" {
		builder.WriteString("  Path:\n")
	} else {
		builder.WriteString("  Path: " + a.Path + "\n")
	}
	builder.WriteString("  Source length: ")
	builder.WriteString(fmt.Sprintf("%d", len(a.Source)))
	builder.WriteString(" characters\n")

	if len(a.TypeAnnotations) > 0 {
		builder.WriteString("  Type Annotations:\n")
		for _, annotation := range a.TypeAnnotations {
			builder.WriteString("    ")
			builder.WriteString(annotation.String())
			builder.WriteString("\n")
		}
	}

	if len(a.Errors) > 0 {
		builder.WriteString("  Errors:\n")
		for _, err := range a.Errors {
			builder.WriteString("    ")
			builder.WriteString(err.Error())
			builder.WriteString("\n")
		}
	}

	if a.Root != nil {
		builder.WriteString("  Root: ")
		builder.WriteString(a.Root.Type())
		builder.WriteString("\n")

		// Add tree structure to help with debugging and tests
		builder.WriteString("  Tree Structure:\n")
		a.writeNodeTree(&builder, a.Root, 2)
	}

	builder.WriteString("}\n")
	return builder.String()
}

// writeNodeTree recursively writes the tree structure to the builder
func (a *AST) writeNodeTree(builder *strings.Builder, node Node, depth int) {
	if node == nil {
		return
	}

	// Write indentation
	for i := 0; i < depth; i++ {
		builder.WriteString(" ")
	}

	// Write node type
	builder.WriteString(node.Type())
	builder.WriteString("\n")

	// Recursively write children
	for _, child := range node.Children() {
		a.writeNodeTree(builder, child, depth+2)
	}
}

// TypeAnnotation represents a type annotation in the code
// Consolidated from parser and treesitter TypeAnnotation types
type TypeAnnotation struct {
	// AnnotatedItem is the name of the annotated item (variable, function, etc.)
	AnnotatedItem string

	// TypeExpression is the type expression itself
	TypeExpression *TypeExpression

	// Position is the location of the annotation in source
	Pos Position

	// Kind indicates what kind of item is being annotated
	Kind AnnotationKind
}

// String returns a string representation of the type annotation
func (ta *TypeAnnotation) String() string {
	var builder strings.Builder

	builder.WriteString(ta.Kind.String())
	builder.WriteString(": ")
	builder.WriteString(ta.AnnotatedItem)
	if ta.TypeExpression != nil {
		builder.WriteString(" :: ")
		builder.WriteString(ta.TypeExpression.String())
	}
	builder.WriteString(fmt.Sprintf(" at %d:%d", ta.Pos.Line, ta.Pos.Column))

	return builder.String()
}

// AnnotationKind represents the kind of type annotation
type AnnotationKind int

const (
	// VarAnnotation is a variable type annotation (my Type $var)
	VarAnnotation AnnotationKind = iota

	// SubParamAnnotation is a subroutine parameter annotation
	SubParamAnnotation

	// SubReturnAnnotation is a subroutine return type annotation
	SubReturnAnnotation

	// MethodParamAnnotation is a method parameter annotation
	MethodParamAnnotation

	// MethodReturnAnnotation is a method return type annotation
	MethodReturnAnnotation

	// FieldAnnotation is a field/attribute annotation
	FieldAnnotation

	// TypeDeclAnnotation is a type declaration
	TypeDeclAnnotation

	// TypeAssertionAnnotation is a type assertion (expr as Type)
	TypeAssertionAnnotation
)

// TypeExpression represents a type expression
// Consolidated and enhanced from parser and treesitter implementations
type TypeExpression struct {
	*BaseNode

	// Kind represents the type expression kind
	Kind TypeExpressionKind

	// Name is the primary type name (renamed from BaseType for consistency)
	Name string

	// Parameters are type parameters for parameterized types (e.g., ArrayRef[Int])
	Parameters []*TypeExpression

	// IsUnion indicates this is a union type (Type1|Type2)
	IsUnion bool

	// IsIntersection indicates this is an intersection type (Type1&Type2)
	IsIntersection bool

	// IsNegation indicates this is a negation type (!Type)
	IsNegation bool

	// UnionTypes contains the types in a union (for multi-way unions)
	UnionTypes []*TypeExpression

	// IntersectionTypes contains the types in an intersection
	IntersectionTypes []*TypeExpression

	// NegatedType is the type being negated (for negation types)
	NegatedType *TypeExpression

	// Constraint represents type constraints (where clauses)
	Constraint ExpressionNode

	// StructuralMembers contains the members for structural types
	StructuralMembers []*StructuralTypeMember

	// ConditionalCondition is the check type for conditional types
	ConditionalCondition *TypeExpression

	// ConditionalTarget is the type being checked against for conditional types
	ConditionalTarget *TypeExpression

	// ConditionalRelationship is the relationship operator (extends, implements, isa, does)
	ConditionalRelationship string

	// ConditionalTrueType is the type returned when condition is true
	ConditionalTrueType *TypeExpression

	// ConditionalFalseType is the type returned when condition is false
	ConditionalFalseType *TypeExpression

	// TypeGuardTarget is the type being guarded for TypeGuard<T> types
	TypeGuardTarget *TypeExpression

	// OriginalString preserves the original source text
	OriginalString string
}

// TypeExpressionKind represents the kind of type expression
type TypeExpressionKind int

const (
	SimpleTypeKind TypeExpressionKind = iota
	UnionTypeKind
	IntersectionTypeKind
	NegationTypeKind
	ParameterizedTypeKind
	ConstrainedTypeKind
	StructuralTypeKind
	ConditionalTypeKind
	TypeGuardKind
)

// NewTypeExpression creates a new type expression
func NewTypeExpression(name string, params []*TypeExpression, start, end Position) *TypeExpression {
	kind := SimpleTypeKind
	if len(params) > 0 {
		kind = ParameterizedTypeKind
	}

	te := &TypeExpression{
		BaseNode:   NewBaseNode("type_expr", start, end),
		Kind:       kind,
		Name:       name,
		Parameters: params,
	}

	// Add parameter children
	for _, param := range params {
		if param != nil {
			te.AddChild(param)
		}
	}

	return te
}

// String returns a string representation of the type expression
func (te *TypeExpression) String() string {
	if te == nil {
		return ""
	}

	// Use original string if available
	if te.OriginalString != "" {
		return te.OriginalString
	}

	// Handle negation
	if te.IsNegation && te.NegatedType != nil {
		return "!" + te.NegatedType.String()
	}

	// Handle union types
	if te.IsUnion && len(te.UnionTypes) > 0 {
		var parts []string
		for _, unionType := range te.UnionTypes {
			parts = append(parts, unionType.String())
		}
		return strings.Join(parts, "|")
	}

	// Handle intersection types
	if te.IsIntersection && len(te.IntersectionTypes) > 0 {
		var parts []string
		for _, intType := range te.IntersectionTypes {
			parts = append(parts, intType.String())
		}
		return strings.Join(parts, "&")
	}

	// Handle structural types
	if te.Kind == StructuralTypeKind && len(te.StructuralMembers) > 0 {
		var members []string
		for _, member := range te.StructuralMembers {
			members = append(members, member.Key+": "+member.Type.String())
		}
		return "struct { " + strings.Join(members, ", ") + " }"
	}

	// Handle conditional types
	if te.Kind == ConditionalTypeKind && te.ConditionalCondition != nil && te.ConditionalTarget != nil {
		return "(" + te.ConditionalCondition.String() + " " + te.ConditionalRelationship + " " +
			te.ConditionalTarget.String() + " ? " + te.ConditionalTrueType.String() + " : " +
			te.ConditionalFalseType.String() + ")"
	}

	// Handle type guards
	if te.Kind == TypeGuardKind && te.TypeGuardTarget != nil {
		return "TypeGuard<" + te.TypeGuardTarget.String() + ">"
	}

	// Handle parameterized types
	if len(te.Parameters) > 0 {
		var params []string
		for _, param := range te.Parameters {
			params = append(params, param.String())
		}
		return te.Name + "[" + strings.Join(params, ", ") + "]"
	}

	// Simple type
	return te.Name
}

// IsSimple returns true if this is a simple, non-compound type
func (te *TypeExpression) IsSimple() bool {
	return !te.IsUnion && !te.IsIntersection && !te.IsNegation && len(te.Parameters) == 0
}

// GetAllTypes returns all types mentioned in this expression (for unions/intersections)
func (te *TypeExpression) GetAllTypes() []string {
	var types []string

	switch {
	case te.IsUnion && len(te.UnionTypes) > 0:
		for _, unionType := range te.UnionTypes {
			types = append(types, unionType.GetAllTypes()...)
		}
	case te.IsIntersection && len(te.IntersectionTypes) > 0:
		for _, intType := range te.IntersectionTypes {
			types = append(types, intType.GetAllTypes()...)
		}
	case te.IsNegation && te.NegatedType != nil:
		types = append(types, te.NegatedType.GetAllTypes()...)
	default:
		types = append(types, te.Name)
	}

	return types
}

// WhereClauseMultiple represents multiple type constraints
type WhereClauseMultiple struct {
	*BaseNode
	Constraints []*TypeConstraint // list of constraints
}

// NewWhereClauseMultiple creates a new where clause with constraints
func NewWhereClauseMultiple(constraints []*TypeConstraint, start, end Position) *WhereClauseMultiple {
	return &WhereClauseMultiple{
		BaseNode:    NewBaseNode("where_clause_multi", start, end),
		Constraints: constraints,
	}
}

// WhereClause represents a type constraint expression (where { ... })
type WhereClause struct {
	*BaseNode
	Expression Node // The constraint expression - using Node interface to avoid circular dependency
}

// NewWhereClause creates a new where clause
func NewWhereClause(expr Node, start, end Position) *WhereClause {
	clause := &WhereClause{
		BaseNode:   NewBaseNode("where_clause", start, end),
		Expression: expr,
	}

	if expr != nil {
		clause.AddChild(expr)
	}

	return clause
}

// ConstrainedType represents a type with constraints (Type where { ... })
type ConstrainedType struct {
	*BaseNode
	BaseType   *TypeExpression
	Constraint *WhereClause
}

// NewConstrainedType creates a new constrained type
func NewConstrainedType(baseType *TypeExpression, constraint *WhereClause, start, end Position) *ConstrainedType {
	ct := &ConstrainedType{
		BaseNode:   NewBaseNode("constrained_type", start, end),
		BaseType:   baseType,
		Constraint: constraint,
	}

	if baseType != nil {
		ct.AddChild(baseType)
	}
	if constraint != nil {
		ct.AddChild(constraint)
	}

	return ct
}

// String returns a string representation of the constrained type
func (ct *ConstrainedType) String() string {
	if ct.BaseType == nil {
		return ""
	}

	result := ct.BaseType.String()
	if ct.Constraint != nil {
		result += " where { ... }"
	}

	return result
}

// BaseNode provides a common implementation for AST nodes
// This reduces boilerplate for concrete node implementations
type BaseNode struct {
	nodeType string
	start    Position
	end      Position
	children []Node
	parent   Node
	text     string
}

// NewBaseNode creates a new base node with the given type
func NewBaseNode(nodeType string, start, end Position) *BaseNode {
	return &BaseNode{
		nodeType: nodeType,
		start:    start,
		end:      end,
		children: make([]Node, 0),
	}
}

// Type implements Node interface
func (bn *BaseNode) Type() string {
	return bn.nodeType
}

// Start implements Node interface
func (bn *BaseNode) Start() Position {
	if bn == nil {
		return Position{}
	}
	return bn.start
}

// End implements Node interface
func (bn *BaseNode) End() Position {
	if bn == nil {
		return Position{}
	}
	return bn.end
}

// Children implements Node interface
func (bn *BaseNode) Children() []Node {
	if bn == nil {
		return nil
	}
	return bn.children
}

// Text implements Node interface
func (bn *BaseNode) Text() string {
	if bn == nil {
		return ""
	}
	return bn.text
}

// Parent implements Node interface
func (bn *BaseNode) Parent() Node {
	if bn == nil {
		return nil
	}
	return bn.parent
}

// SetParent implements Node interface
func (bn *BaseNode) SetParent(parent Node) {
	if bn == nil {
		return
	}
	bn.parent = parent
}

// AddChild adds a child node and sets its parent
func (bn *BaseNode) AddChild(child Node) {
	if child != nil {
		child.SetParent(bn)
		bn.children = append(bn.children, child)
	}
}

// SetText sets the text content of this node
func (bn *BaseNode) SetText(text string) {
	bn.text = text
}

// TokenNode represents a structural token (punctuation, keywords, whitespace)
// These nodes preserve the concrete syntax tree structure for formatting
type TokenNode struct {
	*BaseNode
	TokenType TokenType
}

// TokenType represents the type of structural token
type TokenType int

const (
	// Punctuation tokens
	LeftBrace  TokenType = iota // {
	RightBrace                  // }
	LeftParen                   // (
	RightParen                  // )
	Semicolon                   // ;
	Arrow                       // ->
	Equals                      // =
	Dollar                      // $

	// Keywords
	SubKeyword    // sub
	MethodKeyword // method
	MyKeyword     // my
	FieldKeyword  // field

	// Whitespace and formatting
	Whitespace // spaces, tabs
	Newline    // \n

	// Other tokens
	Identifier // variable names, type names
	Number     // numeric literals
	String     // string literals
)

// NewTokenNode creates a new token node
func NewTokenNode(tokenType TokenType, text string, start, end Position) *TokenNode {
	node := &TokenNode{
		BaseNode:  NewBaseNode("token", start, end),
		TokenType: tokenType,
	}
	node.SetText(text)
	return node
}

// NodeVisitor defines a function that visits AST nodes
// Used for AST traversal operations
type NodeVisitor func(node Node) bool

// WalkFunc defines a function for AST walking with enter/exit hooks
type WalkFunc struct {
	Enter func(node Node) bool // Return false to skip children
	Exit  func(node Node)      // Called when leaving node
}

// TypeConstraint represents a type constraint
type TypeConstraint struct {
	Parameter  string         // type parameter being constrained
	Kind       ConstraintKind // type, protocol, value, capability
	Expression ExpressionNode // constraint expression
	Position   Position       // source position
}

// ConstraintKind represents the kind of constraint
type ConstraintKind int

const (
	TypeConstraintKind   ConstraintKind = iota // T: SomeType
	ProtocolConstraint                         // T does SomeRole
	CapabilityConstraint                       // T can 'method'
	ValueConstraint                            // $param > 0
	VersionConstraint                          // T->VERSION >= 1.0
)

// StructuralTypeMember represents a member in a structural type
type StructuralTypeMember struct {
	Key      string          // field key/name
	Type     *TypeExpression // field type
	Position Position        // source position
}

// TypeParameter represents a generic type parameter
type TypeParameter struct {
	Name        string            // parameter name (e.g., T, U)
	Constraints []*TypeConstraint // constraints on this parameter
	Position    Position          // source position
}

// ParameterInfo represents detailed parameter information for methods
type ParameterInfo struct {
	Name       string          // parameter name
	Type       *TypeExpression // parameter type (optional)
	Default    ExpressionNode  // default value (optional)
	IsOptional bool            // optional parameter flag
	IsNamed    bool            // named parameter flag (:$param)
	IsVariadic bool            // variadic parameter flag (*@args)
	Position   Position        // source position
}

// MethodSignature represents a complete method signature
type MethodSignature struct {
	Name           string            // method name
	TypeParameters []*TypeParameter  // generic type parameters
	Parameters     []*ParameterInfo  // method parameters
	ReturnType     *TypeExpression   // return type specification
	Constraints    []*TypeConstraint // type constraints
	Position       Position          // source position
}

// TypeVisitor defines an interface for visiting type information in AST
type TypeVisitor interface {
	VisitTypeExpression(node *TypeExpression) error
	VisitTypedVariable(node *VarDecl) error
	VisitTypedMethod(node *SubDecl) error
	VisitTypeAssertion(node *TypeAssertionExpr) error
	VisitFieldDeclaration(node *FieldDecl) error
	VisitTypeDeclaration(node *TypeDecl) error
}

// TypeWalker provides functionality to traverse AST and visit type information
type TypeWalker struct {
	visitor TypeVisitor
}

// NewTypeWalker creates a new type walker
func NewTypeWalker(visitor TypeVisitor) *TypeWalker {
	return &TypeWalker{visitor: visitor}
}

// WalkTypes traverses AST and calls visitor methods for nodes with type information
func (tw *TypeWalker) WalkTypes(ast *AST) error {
	if ast.Root == nil {
		return nil
	}
	return tw.walkNode(ast.Root)
}

// walkNode recursively walks AST nodes looking for type information
func (tw *TypeWalker) walkNode(node Node) error {
	if node == nil {
		return nil
	}

	// Visit specific node types with type information
	// TODO: Re-enable type-specific visiting once go-critic caseOrder issue is resolved
	_ = node // Temporary to avoid unused variable warning

	// Recursively walk children
	for _, child := range node.Children() {
		if err := tw.walkNode(child); err != nil {
			return err
		}
	}

	return nil
}

// TypeInformation represents serializable type information from an AST
type TypeInformation struct {
	Variables      []*VariableTypeInfo  `json:"variables"`
	Methods        []*MethodSignature   `json:"methods"`
	Fields         []*FieldTypeInfo     `json:"fields"`
	TypeAliases    []*TypeAliasInfo     `json:"type_aliases"`
	TypeAssertions []*TypeAssertionInfo `json:"type_assertions"`
	Classes        []*ClassTypeInfo     `json:"classes"`
	Roles          []*RoleTypeInfo      `json:"roles"`
	FilePath       string               `json:"file_path"`
	Timestamp      int64                `json:"timestamp"`
}

// ExtractTypeInformation extracts all type information from an AST for serialization
func ExtractTypeInformation(ast *AST) *TypeInformation {
	extractor := &typeInformationExtractor{
		typeInfo: &TypeInformation{
			FilePath:  ast.Path,
			Timestamp: time.Now().Unix(),
		},
	}

	walker := NewTypeWalker(extractor)
	walker.WalkTypes(ast)

	return extractor.typeInfo
}

// typeInformationExtractor implements TypeVisitor to extract type information
type typeInformationExtractor struct {
	typeInfo *TypeInformation
}

func (tie *typeInformationExtractor) VisitTypeExpression(node *TypeExpression) error {
	// Type expressions are handled as part of other nodes
	return nil
}

func (tie *typeInformationExtractor) VisitTypedVariable(node *VarDecl) error {
	if info := node.GetTypeInfo(); info != nil {
		tie.typeInfo.Variables = append(tie.typeInfo.Variables, info)
	}
	return nil
}

func (tie *typeInformationExtractor) VisitTypedMethod(node *SubDecl) error {
	if sig := node.GetMethodSignature(); sig != nil {
		tie.typeInfo.Methods = append(tie.typeInfo.Methods, sig)
	}
	return nil
}

func (tie *typeInformationExtractor) VisitTypeAssertion(node *TypeAssertionExpr) error {
	if info := node.GetTypeAssertionInfo(); info != nil {
		tie.typeInfo.TypeAssertions = append(tie.typeInfo.TypeAssertions, info)
	}
	return nil
}

func (tie *typeInformationExtractor) VisitFieldDeclaration(node *FieldDecl) error {
	if info := node.GetFieldTypeInfo(); info != nil {
		tie.typeInfo.Fields = append(tie.typeInfo.Fields, info)
	}
	return nil
}

func (tie *typeInformationExtractor) VisitTypeDeclaration(node *TypeDecl) error {
	if info := node.GetTypeAliasInfo(); info != nil {
		tie.typeInfo.TypeAliases = append(tie.typeInfo.TypeAliases, info)
	}
	return nil
}

// JSON marshaling helper types and functions

// nodeJSON represents a JSON-serializable node structure
type nodeJSON struct {
	Type     string      `json:"type"`
	Start    Position    `json:"start"`
	End      Position    `json:"end"`
	Text     string      `json:"text,omitempty"`
	Children []*nodeJSON `json:"children,omitempty"`

	// Node-specific fields (will be populated based on node type)
	Value    interface{} `json:"value,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Name     string      `json:"name,omitempty"`
	Sigil    string      `json:"sigil,omitempty"`
	DeclType string      `json:"decl_type,omitempty"`
	Package  string      `json:"package,omitempty"`
	Method   bool        `json:"method,omitempty"`
	Prefix   bool        `json:"prefix,omitempty"`
	Kind     string      `json:"kind,omitempty"`
}

// typeAnnotationJSON represents a JSON-serializable type annotation
type typeAnnotationJSON struct {
	AnnotatedItem  string          `json:"annotated_item"`
	TypeExpression *TypeExpression `json:"type_expression,omitempty"`
	Position       Position        `json:"position"`
	Kind           string          `json:"kind"`
}

// nodeToJSON converts a Node to JSON-serializable format
// This avoids circular references by excluding parent pointers
func nodeToJSON(node Node) *nodeJSON {
	if node == nil {
		return nil
	}

	// Create base node JSON
	nodeJS := &nodeJSON{
		Type:  node.Type(),
		Start: node.Start(),
		End:   node.End(),
		Text:  node.Text(),
	}

	// Add children (recursively)
	for _, child := range node.Children() {
		if child != nil {
			nodeJS.Children = append(nodeJS.Children, nodeToJSON(child))
		}
	}

	// Add type-specific fields for core types only
	if n, ok := node.(*TypeExpression); ok {
		nodeJS.Name = n.Name
		nodeJS.Kind = typeExpressionKindToString(n.Kind)
	}

	return nodeJS
}

// Helper functions to convert enums to strings
func literalKindToString(kind LiteralKind) string {
	switch kind {
	case StringLiteral:
		return "string"
	case NumberLiteral:
		return "number"
	case BooleanLiteral:
		return "boolean"
	case UndefLiteral:
		return "undef"
	case RegexLiteral:
		return "regex"
	case HashAccessLiteral:
		return "hash_access"
	case ArrayAccessLiteral:
		return "array_access"
	case MethodCallLiteral:
		return "method_call"
	case FunctionCallLiteral:
		return "function_call"
	case BinaryExpressionLiteral:
		return "binary_expression"
	case ExpressionLiteral:
		return "expression"
	default:
		return "unknown"
	}
}

func typeExpressionKindToString(kind TypeExpressionKind) string {
	switch kind {
	case SimpleTypeKind:
		return "simple"
	case UnionTypeKind:
		return "union"
	case IntersectionTypeKind:
		return "intersection"
	case NegationTypeKind:
		return "negation"
	case ParameterizedTypeKind:
		return "parameterized"
	case ConstrainedTypeKind:
		return "constrained"
	case StructuralTypeKind:
		return "structural"
	case ConditionalTypeKind:
		return "conditional"
	case TypeGuardKind:
		return "typeguard"
	default:
		return "unknown"
	}
}
