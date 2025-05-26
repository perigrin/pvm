// ABOUTME: Statement AST node types for Perl statements and declarations
// ABOUTME: Covers all major statement types including typed declarations

package ast

// StatementNode represents any kind of statement
type StatementNode interface {
	Node
	IsStatement() bool
}

// ExpressionStmt represents expression statements (expressions used as statements)
type ExpressionStmt struct {
	*BaseNode
	Expression ExpressionNode
}

// NewExpressionStmt creates a new expression statement
func NewExpressionStmt(expr ExpressionNode, start, end Position) *ExpressionStmt {
	stmt := &ExpressionStmt{
		BaseNode:   NewBaseNode("expression_stmt", start, end),
		Expression: expr,
	}

	if expr != nil {
		stmt.AddChild(expr)
	}

	return stmt
}

// IsStatement implements StatementNode interface
func (es *ExpressionStmt) IsStatement() bool {
	return true
}

// VarDecl represents variable declarations (my, our, state)
type VarDecl struct {
	*BaseNode
	DeclType    string          // "my", "our", "state"
	Variables   []*VariableExpr // Variables being declared
	TypeExpr    *TypeExpression // Optional type annotation
	Initializer ExpressionNode  // Optional initializer
}

// NewVarDecl creates a new variable declaration
func NewVarDecl(declType string, vars []*VariableExpr, typeExpr *TypeExpression, init ExpressionNode, start, end Position) *VarDecl {
	decl := &VarDecl{
		BaseNode:    NewBaseNode("var_decl", start, end),
		DeclType:    declType,
		Variables:   vars,
		TypeExpr:    typeExpr,
		Initializer: init,
	}

	// Add children
	for _, v := range vars {
		if v != nil {
			decl.AddChild(v)
		}
	}
	if init != nil {
		decl.AddChild(init)
	}

	return decl
}

// IsStatement implements StatementNode interface
func (vd *VarDecl) IsStatement() bool {
	return true
}

// IsTyped returns true if this declaration has a type annotation
func (vd *VarDecl) IsTyped() bool {
	return vd.TypeExpr != nil
}

// SubDecl represents subroutine declarations
type SubDecl struct {
	*BaseNode
	Name       string
	Parameters []*Parameter
	ReturnType *TypeExpression
	Body       *BlockStmt
	IsMethod   bool
}

// Parameter represents a subroutine parameter
type Parameter struct {
	Name     string
	TypeExpr *TypeExpression
	Variable *VariableExpr
	Pos      Position
}

// NewSubDecl creates a new subroutine declaration
func NewSubDecl(name string, params []*Parameter, returnType *TypeExpression, body *BlockStmt, isMethod bool, start, end Position) *SubDecl {
	decl := &SubDecl{
		BaseNode:   NewBaseNode("sub_decl", start, end),
		Name:       name,
		Parameters: params,
		ReturnType: returnType,
		Body:       body,
		IsMethod:   isMethod,
	}

	// Add body as child
	if body != nil {
		decl.AddChild(body)
	}

	return decl
}

// IsStatement implements StatementNode interface
func (sd *SubDecl) IsStatement() bool {
	return true
}

// IsTyped returns true if this subroutine has type annotations
func (sd *SubDecl) IsTyped() bool {
	if sd.ReturnType != nil {
		return true
	}
	for _, param := range sd.Parameters {
		if param.TypeExpr != nil {
			return true
		}
	}
	return false
}

// MethodDecl represents method declarations (method keyword)
type MethodDecl struct {
	*SubDecl // Embed SubDecl since methods are similar to subs
}

// NewMethodDecl creates a new method declaration
func NewMethodDecl(name string, params []*Parameter, returnType *TypeExpression, body *BlockStmt, start, end Position) *MethodDecl {
	subDecl := NewSubDecl(name, params, returnType, body, true, start, end)
	subDecl.nodeType = "method_decl"

	return &MethodDecl{
		SubDecl: subDecl,
	}
}

// FieldDecl represents field declarations (field keyword)
type FieldDecl struct {
	*BaseNode
	Name        string
	TypeExpr    *TypeExpression
	Variable    *VariableExpr
	Initializer ExpressionNode
}

// NewFieldDecl creates a new field declaration
func NewFieldDecl(name string, typeExpr *TypeExpression, variable *VariableExpr, init ExpressionNode, start, end Position) *FieldDecl {
	decl := &FieldDecl{
		BaseNode:    NewBaseNode("field_decl", start, end),
		Name:        name,
		TypeExpr:    typeExpr,
		Variable:    variable,
		Initializer: init,
	}

	if variable != nil {
		decl.AddChild(variable)
	}
	if init != nil {
		decl.AddChild(init)
	}

	return decl
}

// IsStatement implements StatementNode interface
func (fd *FieldDecl) IsStatement() bool {
	return true
}

// TypeDecl represents type declarations (type Name = Type)
type TypeDecl struct {
	*BaseNode
	Name     string
	TypeExpr *TypeExpression
}

// NewTypeDecl creates a new type declaration
func NewTypeDecl(name string, typeExpr *TypeExpression, start, end Position) *TypeDecl {
	return &TypeDecl{
		BaseNode: NewBaseNode("type_decl", start, end),
		Name:     name,
		TypeExpr: typeExpr,
	}
}

// IsStatement implements StatementNode interface
func (td *TypeDecl) IsStatement() bool {
	return true
}

// BlockStmt represents block statements ({ ... })
type BlockStmt struct {
	*BaseNode
	Statements []StatementNode
}

// NewBlockStmt creates a new block statement
func NewBlockStmt(statements []StatementNode, start, end Position) *BlockStmt {
	stmt := &BlockStmt{
		BaseNode:   NewBaseNode("block_stmt", start, end),
		Statements: statements,
	}

	// Add all statements as children
	for _, s := range statements {
		if s != nil {
			stmt.AddChild(s)
		}
	}

	return stmt
}

// IsStatement implements StatementNode interface
func (bs *BlockStmt) IsStatement() bool {
	return true
}

// IfStmt represents if statements
type IfStmt struct {
	*BaseNode
	Condition ExpressionNode
	ThenStmt  StatementNode
	ElseStmt  StatementNode // Optional
}

// NewIfStmt creates a new if statement
func NewIfStmt(condition ExpressionNode, thenStmt, elseStmt StatementNode, start, end Position) *IfStmt {
	stmt := &IfStmt{
		BaseNode:  NewBaseNode("if_stmt", start, end),
		Condition: condition,
		ThenStmt:  thenStmt,
		ElseStmt:  elseStmt,
	}

	if condition != nil {
		stmt.AddChild(condition)
	}
	if thenStmt != nil {
		stmt.AddChild(thenStmt)
	}
	if elseStmt != nil {
		stmt.AddChild(elseStmt)
	}

	return stmt
}

// IsStatement implements StatementNode interface
func (is *IfStmt) IsStatement() bool {
	return true
}

// WhileStmt represents while loops
type WhileStmt struct {
	*BaseNode
	Condition ExpressionNode
	Body      StatementNode
}

// NewWhileStmt creates a new while statement
func NewWhileStmt(condition ExpressionNode, body StatementNode, start, end Position) *WhileStmt {
	stmt := &WhileStmt{
		BaseNode:  NewBaseNode("while_stmt", start, end),
		Condition: condition,
		Body:      body,
	}

	if condition != nil {
		stmt.AddChild(condition)
	}
	if body != nil {
		stmt.AddChild(body)
	}

	return stmt
}

// IsStatement implements StatementNode interface
func (ws *WhileStmt) IsStatement() bool {
	return true
}

// ForStmt represents for loops
type ForStmt struct {
	*BaseNode
	Iterator ExpressionNode // Iterator variable
	Iterable ExpressionNode // What we're iterating over
	Body     StatementNode
}

// NewForStmt creates a new for statement
func NewForStmt(iterator, iterable ExpressionNode, body StatementNode, start, end Position) *ForStmt {
	stmt := &ForStmt{
		BaseNode: NewBaseNode("for_stmt", start, end),
		Iterator: iterator,
		Iterable: iterable,
		Body:     body,
	}

	if iterator != nil {
		stmt.AddChild(iterator)
	}
	if iterable != nil {
		stmt.AddChild(iterable)
	}
	if body != nil {
		stmt.AddChild(body)
	}

	return stmt
}

// IsStatement implements StatementNode interface
func (fs *ForStmt) IsStatement() bool {
	return true
}

// ReturnStmt represents return statements
type ReturnStmt struct {
	*BaseNode
	Value ExpressionNode // Optional return value
}

// NewReturnStmt creates a new return statement
func NewReturnStmt(value ExpressionNode, start, end Position) *ReturnStmt {
	stmt := &ReturnStmt{
		BaseNode: NewBaseNode("return_stmt", start, end),
		Value:    value,
	}

	if value != nil {
		stmt.AddChild(value)
	}

	return stmt
}

// IsStatement implements StatementNode interface
func (rs *ReturnStmt) IsStatement() bool {
	return true
}

// UseStmt represents use statements
type UseStmt struct {
	*BaseNode
	Module  string
	Version string   // Optional version specification
	Imports []string // Optional import list
}

// NewUseStmt creates a new use statement
func NewUseStmt(module, version string, imports []string, start, end Position) *UseStmt {
	return &UseStmt{
		BaseNode: NewBaseNode("use_stmt", start, end),
		Module:   module,
		Version:  version,
		Imports:  imports,
	}
}

// IsStatement implements StatementNode interface
func (us *UseStmt) IsStatement() bool {
	return true
}

// PackageStmt represents package statements
type PackageStmt struct {
	*BaseNode
	Name    string
	Version string // Optional version
}

// NewPackageStmt creates a new package statement
func NewPackageStmt(name, version string, start, end Position) *PackageStmt {
	return &PackageStmt{
		BaseNode: NewBaseNode("package_stmt", start, end),
		Name:     name,
		Version:  version,
	}
}

// IsStatement implements StatementNode interface
func (ps *PackageStmt) IsStatement() bool {
	return true
}
