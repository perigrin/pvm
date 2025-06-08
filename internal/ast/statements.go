// ABOUTME: Statement AST node types for Perl statements and declarations
// ABOUTME: Covers all major statement types including typed declarations

package ast

// StatementNode represents any kind of statement
type StatementNode interface {
	Node
	IsStatement() bool
}

// ProgramStmt represents the top-level program containing multiple statements
// Unlike BlockStmt, ProgramStmt does not create a new scope - it represents
// the global program context that can contain subroutines, classes, and statements
type ProgramStmt struct {
	*BaseNode
	statements []StatementNode // Private - logical statements only
}

// NewProgramStmt creates a new program statement
func NewProgramStmt(statements []StatementNode, start, end Position) *ProgramStmt {
	stmt := &ProgramStmt{
		BaseNode:   NewBaseNode("program_stmt", start, end),
		statements: make([]StatementNode, 0),
	}

	// Add statements using the new pattern
	for _, s := range statements {
		if s != nil {
			stmt.AddStatement(s)
		}
	}

	return stmt
}

// AddStatement adds a statement to the private collection
func (ps *ProgramStmt) AddStatement(statement StatementNode) {
	ps.statements = append(ps.statements, statement)
	ps.AddChild(statement)
}

// LogicalStatements returns the logical statements
func (ps *ProgramStmt) LogicalStatements() []StatementNode {
	return ps.statements
}

// Statements returns logical statements (for backward compatibility)
func (ps *ProgramStmt) Statements() []StatementNode {
	return ps.statements
}

// IsStatement implements StatementNode interface
func (ps *ProgramStmt) IsStatement() bool {
	return true
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
// Enhanced with comprehensive type information
type VarDecl struct {
	*BaseNode
	DeclType    string          // "my", "our", "state"
	variables   []*VariableExpr // Private - logical variables only
	TypeExpr    *TypeExpression // Optional type annotation
	Initializer ExpressionNode  // Optional initializer
	Scope       string          // variable scope context
	Package     string          // package qualification if any
}

// NewVarDecl creates a new variable declaration
func NewVarDecl(declType string, vars []*VariableExpr, typeExpr *TypeExpression, init ExpressionNode, start, end Position) *VarDecl {
	decl := &VarDecl{
		BaseNode:    NewBaseNode("var_decl", start, end),
		DeclType:    declType,
		variables:   make([]*VariableExpr, 0),
		TypeExpr:    typeExpr,
		Initializer: init,
		Scope:       declType, // scope defaults to declaration type
	}

	// Add variables using the new pattern
	for _, v := range vars {
		if v != nil {
			decl.AddVariable(v)
		}
	}

	// Add children
	if typeExpr != nil {
		decl.AddChild(typeExpr)
	}
	if init != nil {
		decl.AddChild(init)
	}

	return decl
}

// AddVariable adds a variable to the private collection
func (vd *VarDecl) AddVariable(variable *VariableExpr) {
	vd.variables = append(vd.variables, variable)
	vd.AddChild(variable)
}

// LogicalVariables returns the logical variables
func (vd *VarDecl) LogicalVariables() []*VariableExpr {
	return vd.variables
}

// Variables returns logical variables (for backward compatibility)
func (vd *VarDecl) Variables() []*VariableExpr {
	return vd.variables
}

// GetTypeInfo returns comprehensive type information for this declaration
func (vd *VarDecl) GetTypeInfo() *VariableTypeInfo {
	return &VariableTypeInfo{
		DeclType:       vd.DeclType,
		TypeExpr:       vd.TypeExpr,
		VariableNames:  vd.getVariableNames(),
		Scope:          vd.Scope,
		Package:        vd.Package,
		HasInitializer: vd.Initializer != nil,
		Position:       vd.Start(),
	}
}

// getVariableNames extracts variable names from the declaration
func (vd *VarDecl) getVariableNames() []string {
	var names []string
	for _, v := range vd.LogicalVariables() {
		if v != nil {
			names = append(names, v.FullName())
		}
	}
	return names
}

// VariableTypeInfo contains comprehensive type information for a variable declaration
type VariableTypeInfo struct {
	DeclType       string          // my, our, state, etc.
	TypeExpr       *TypeExpression // type annotation
	VariableNames  []string        // names of declared variables
	Scope          string          // scope context
	Package        string          // package qualification
	HasInitializer bool            // whether declaration has initializer
	Position       Position        // source position
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
// Enhanced with comprehensive type signature information
type SubDecl struct {
	*BaseNode
	Name           string            // subroutine/method name
	parameters     []*Parameter      // Private - logical parameters only
	ReturnType     *TypeExpression   // return type annotation
	Body           *BlockStmt        // subroutine body
	IsMethod       bool              // true if this is a method
	TypeParameters []*TypeParameter  // generic type parameters
	Constraints    []*TypeConstraint // type constraints
	Signature      *MethodSignature  // complete method signature info
	Package        string            // package context
	AccessLevel    string            // public, private, protected (future)
}

// Parameter represents a subroutine parameter
// Enhanced with comprehensive parameter information
type Parameter struct {
	Name       string          // parameter name
	TypeExpr   *TypeExpression // parameter type annotation
	Variable   *VariableExpr   // parameter variable
	Default    ExpressionNode  // default value expression
	IsOptional bool            // optional parameter flag
	IsNamed    bool            // named parameter flag (:$param)
	IsVariadic bool            // variadic parameter flag (*@args)
	Pos        Position        // source position
}

// NewSubDecl creates a new subroutine declaration
func NewSubDecl(name string, params []*Parameter, returnType *TypeExpression, body *BlockStmt, isMethod bool, start, end Position) *SubDecl {
	decl := &SubDecl{
		BaseNode:   NewBaseNode("sub_decl", start, end),
		Name:       name,
		parameters: make([]*Parameter, 0),
		ReturnType: returnType,
		Body:       body,
		IsMethod:   isMethod,
	}

	// Add parameters using the new pattern
	for _, param := range params {
		if param != nil {
			decl.AddParameter(param)
		}
	}

	// Add children
	if returnType != nil {
		decl.AddChild(returnType)
	}
	if body != nil {
		decl.AddChild(body)
	}

	// Build method signature if this is typed
	if decl.IsTyped() {
		decl.Signature = decl.buildMethodSignature()
	}

	return decl
}

// AddParameter adds a parameter to the private collection
func (sd *SubDecl) AddParameter(param *Parameter) {
	sd.parameters = append(sd.parameters, param)
}

// LogicalParameters returns the logical parameters
func (sd *SubDecl) LogicalParameters() []*Parameter {
	return sd.parameters
}

// Parameters returns logical parameters (for backward compatibility)
func (sd *SubDecl) Parameters() []*Parameter {
	return sd.parameters
}

// buildMethodSignature creates a MethodSignature from the subroutine declaration
func (sd *SubDecl) buildMethodSignature() *MethodSignature {
	var paramInfos []*ParameterInfo
	for _, param := range sd.LogicalParameters() {
		if param != nil {
			paramInfos = append(paramInfos, &ParameterInfo{
				Name:       param.Name,
				Type:       param.TypeExpr,
				Default:    param.Default,
				IsOptional: param.IsOptional,
				IsNamed:    param.IsNamed,
				IsVariadic: param.IsVariadic,
				Position:   param.Pos,
			})
		}
	}

	return &MethodSignature{
		Name:           sd.Name,
		TypeParameters: sd.TypeParameters,
		Parameters:     paramInfos,
		ReturnType:     sd.ReturnType,
		Constraints:    sd.Constraints,
		Position:       sd.Start(),
	}
}

// GetMethodSignature returns the complete method signature information
func (sd *SubDecl) GetMethodSignature() *MethodSignature {
	if sd.Signature == nil && sd.IsTyped() {
		sd.Signature = sd.buildMethodSignature()
	}
	return sd.Signature
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
	for _, param := range sd.LogicalParameters() {
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
// Enhanced with comprehensive field type information
type FieldDecl struct {
	*BaseNode
	Name        string            // field name
	TypeExpr    *TypeExpression   // field type annotation
	Variable    *VariableExpr     // field variable
	Initializer ExpressionNode    // field initializer
	AccessLevel string            // public, private, protected (future)
	IsStatic    bool              // static field flag (future)
	Package     string            // package context
	Constraints []*TypeConstraint // field-specific constraints
}

// NewFieldDecl creates a new field declaration
func NewFieldDecl(name string, typeExpr *TypeExpression, variable *VariableExpr, init ExpressionNode, start, end Position) *FieldDecl {
	decl := &FieldDecl{
		BaseNode:    NewBaseNode("field_decl", start, end),
		Name:        name,
		TypeExpr:    typeExpr,
		Variable:    variable,
		Initializer: init,
		AccessLevel: "public", // default to public
	}

	if typeExpr != nil {
		decl.AddChild(typeExpr)
	}
	if variable != nil {
		decl.AddChild(variable)
	}
	if init != nil {
		decl.AddChild(init)
	}

	return decl
}

// GetFieldTypeInfo returns comprehensive type information for this field
func (fd *FieldDecl) GetFieldTypeInfo() *FieldTypeInfo {
	return &FieldTypeInfo{
		Name:           fd.Name,
		TypeExpr:       fd.TypeExpr,
		HasInitializer: fd.Initializer != nil,
		AccessLevel:    fd.AccessLevel,
		IsStatic:       fd.IsStatic,
		Package:        fd.Package,
		Constraints:    fd.Constraints,
		Position:       fd.Start(),
	}
}

// FieldTypeInfo contains comprehensive type information for a field declaration
type FieldTypeInfo struct {
	Name           string            // field name
	TypeExpr       *TypeExpression   // type annotation
	HasInitializer bool              // whether field has initializer
	AccessLevel    string            // access level
	IsStatic       bool              // static field flag
	Package        string            // package context
	Constraints    []*TypeConstraint // field constraints
	Position       Position          // source position
}

// IsStatement implements StatementNode interface
func (fd *FieldDecl) IsStatement() bool {
	return true
}

// TypeDecl represents type declarations (type Name = Type)
// Enhanced with comprehensive type alias information
type TypeDecl struct {
	*BaseNode
	Name           string            // type alias name
	TypeExpr       *TypeExpression   // aliased type expression
	TypeParameters []*TypeParameter  // generic type parameters
	Constraints    []*TypeConstraint // type constraints
	Package        string            // package context
	IsExported     bool              // whether type is exported
}

// NewTypeDecl creates a new type declaration
func NewTypeDecl(name string, typeExpr *TypeExpression, start, end Position) *TypeDecl {
	decl := &TypeDecl{
		BaseNode:   NewBaseNode("type_decl", start, end),
		Name:       name,
		TypeExpr:   typeExpr,
		IsExported: isExportedName(name), // determine if name is exported
	}

	if typeExpr != nil {
		decl.AddChild(typeExpr)
	}

	return decl
}

// isExportedName determines if a type name is exported (starts with uppercase)
func isExportedName(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

// GetTypeAliasInfo returns comprehensive information about this type alias
func (td *TypeDecl) GetTypeAliasInfo() *TypeAliasInfo {
	return &TypeAliasInfo{
		Name:           td.Name,
		AliasedType:    td.TypeExpr,
		TypeParameters: td.TypeParameters,
		Constraints:    td.Constraints,
		Package:        td.Package,
		IsExported:     td.IsExported,
		Position:       td.Start(),
	}
}

// TypeAliasInfo contains comprehensive information about a type alias
type TypeAliasInfo struct {
	Name           string            // alias name
	AliasedType    *TypeExpression   // the type being aliased
	TypeParameters []*TypeParameter  // generic parameters
	Constraints    []*TypeConstraint // type constraints
	Package        string            // package context
	IsExported     bool              // export status
	Position       Position          // source position
}

// IsStatement implements StatementNode interface
func (td *TypeDecl) IsStatement() bool {
	return true
}

// BlockStmt represents block statements ({ ... })
type BlockStmt struct {
	*BaseNode
	children   []Node          // Private - all tokens in order
	statements []StatementNode // Private - cached logical statements
}

// NewBlockStmt creates a new block statement
func NewBlockStmt(statements []StatementNode, start, end Position) *BlockStmt {
	stmt := &BlockStmt{
		BaseNode:   NewBaseNode("block_stmt", start, end),
		children:   make([]Node, 0),
		statements: make([]StatementNode, 0),
	}

	// Add all statements using the new pattern
	for _, s := range statements {
		if s != nil {
			stmt.AddChild(s)
		}
	}

	return stmt
}

// AddChild handles adding to both collections appropriately
func (bs *BlockStmt) AddChild(child Node) {
	bs.children = append(bs.children, child)
	
	// If it's a statement, also cache it
	if stmt, ok := child.(StatementNode); ok {
		bs.statements = append(bs.statements, stmt)
	}
	
	// Also add to BaseNode for backward compatibility
	bs.BaseNode.AddChild(child)
}

// Children returns all nodes (implements Node interface)
func (bs *BlockStmt) Children() []Node {
	return bs.children
}

// LogicalStatements returns only the statement nodes
func (bs *BlockStmt) LogicalStatements() []StatementNode {
	return bs.statements
}

// Statements returns logical statements (for backward compatibility)
func (bs *BlockStmt) Statements() []StatementNode {
	return bs.statements
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

// ClassDecl represents class declarations
type ClassDecl struct {
	*BaseNode
	Name           string            // class name
	TypeParameters []*TypeParameter  // generic parameters
	Superclass     *TypeExpression   // parent class
	Roles          []*TypeExpression // implemented roles
	Fields         []*FieldDecl      // class fields
	Methods        []*MethodDecl     // class methods
	Constructors   []*MethodDecl     // constructor methods
	Constraints    []*TypeConstraint // type constraints
	Package        string            // package context
	IsExported     bool              // export status
}

// NewClassDecl creates a new class declaration
func NewClassDecl(name string, start, end Position) *ClassDecl {
	return &ClassDecl{
		BaseNode:   NewBaseNode("class_decl", start, end),
		Name:       name,
		IsExported: isExportedName(name),
	}
}

// IsStatement implements StatementNode interface
func (cd *ClassDecl) IsStatement() bool {
	return true
}

// AddField adds a field to the class
func (cd *ClassDecl) AddField(field *FieldDecl) {
	if field != nil {
		cd.Fields = append(cd.Fields, field)
		cd.AddChild(field)
	}
}

// AddMethod adds a method to the class
func (cd *ClassDecl) AddMethod(method *MethodDecl) {
	if method != nil {
		cd.Methods = append(cd.Methods, method)
		cd.AddChild(method)
	}
}

// AddRole adds a role to the class
func (cd *ClassDecl) AddRole(role *TypeExpression) {
	if role != nil {
		cd.Roles = append(cd.Roles, role)
		cd.AddChild(role)
	}
}

// GetClassTypeInfo returns comprehensive type information for this class
func (cd *ClassDecl) GetClassTypeInfo() *ClassTypeInfo {
	return &ClassTypeInfo{
		Name:           cd.Name,
		TypeParameters: cd.TypeParameters,
		Superclass:     cd.Superclass,
		Roles:          cd.Roles,
		FieldCount:     len(cd.Fields),
		MethodCount:    len(cd.Methods),
		Constraints:    cd.Constraints,
		Package:        cd.Package,
		IsExported:     cd.IsExported,
		Position:       cd.Start(),
	}
}

// ClassTypeInfo contains comprehensive information about a class declaration
type ClassTypeInfo struct {
	Name           string            // class name
	TypeParameters []*TypeParameter  // generic parameters
	Superclass     *TypeExpression   // parent class
	Roles          []*TypeExpression // implemented roles
	FieldCount     int               // number of fields
	MethodCount    int               // number of methods
	Constraints    []*TypeConstraint // type constraints
	Package        string            // package context
	IsExported     bool              // export status
	Position       Position          // source position
}

// RoleDecl represents role declarations
type RoleDecl struct {
	*BaseNode
	Name            string             // role name
	TypeParameters  []*TypeParameter   // generic parameters
	RequiredMethods []*MethodSignature // required method signatures
	ProvidedMethods []*MethodDecl      // provided method implementations
	Fields          []*FieldDecl       // role fields
	Constraints     []*TypeConstraint  // type constraints
	Package         string             // package context
	IsExported      bool               // export status
}

// NewRoleDecl creates a new role declaration
func NewRoleDecl(name string, start, end Position) *RoleDecl {
	return &RoleDecl{
		BaseNode:   NewBaseNode("role_decl", start, end),
		Name:       name,
		IsExported: isExportedName(name),
	}
}

// IsStatement implements StatementNode interface
func (rd *RoleDecl) IsStatement() bool {
	return true
}

// AddRequiredMethod adds a required method signature to the role
func (rd *RoleDecl) AddRequiredMethod(signature *MethodSignature) {
	if signature != nil {
		rd.RequiredMethods = append(rd.RequiredMethods, signature)
	}
}

// AddProvidedMethod adds a provided method implementation to the role
func (rd *RoleDecl) AddProvidedMethod(method *MethodDecl) {
	if method != nil {
		rd.ProvidedMethods = append(rd.ProvidedMethods, method)
		rd.AddChild(method)
	}
}

// AddField adds a field to the role
func (rd *RoleDecl) AddField(field *FieldDecl) {
	if field != nil {
		rd.Fields = append(rd.Fields, field)
		rd.AddChild(field)
	}
}

// GetRoleTypeInfo returns comprehensive type information for this role
func (rd *RoleDecl) GetRoleTypeInfo() *RoleTypeInfo {
	return &RoleTypeInfo{
		Name:                rd.Name,
		TypeParameters:      rd.TypeParameters,
		RequiredMethodCount: len(rd.RequiredMethods),
		ProvidedMethodCount: len(rd.ProvidedMethods),
		FieldCount:          len(rd.Fields),
		Constraints:         rd.Constraints,
		Package:             rd.Package,
		IsExported:          rd.IsExported,
		Position:            rd.Start(),
	}
}

// RoleTypeInfo contains comprehensive information about a role declaration
type RoleTypeInfo struct {
	Name                string            // role name
	TypeParameters      []*TypeParameter  // generic parameters
	RequiredMethodCount int               // number of required methods
	ProvidedMethodCount int               // number of provided methods
	FieldCount          int               // number of fields
	Constraints         []*TypeConstraint // type constraints
	Package             string            // package context
	IsExported          bool              // export status
	Position            Position          // source position
}
