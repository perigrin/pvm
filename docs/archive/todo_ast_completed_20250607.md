# AST Statement Types Encapsulation TODO

## Overview
Following the successful implementation of encapsulated `BlockStmt` with the `AddChild` pattern, several other statement types have been identified that suffer from the same issue: structural tokens (commas, parentheses, semicolons) getting mixed with logical content in public collection fields.

## Root Problem
Tree-sitter parsing creates nodes that include both:
- **Logical content** (actual statements, parameters, variables)
- **Structural tokens** (`,`, `(`, `)`, `;`, `{`, `}`)

When these get mixed in public collection fields, it causes issues for:
- **Source-to-source compilation** (extra tokens in logical iteration)
- **Semantic analysis** (structural tokens treated as content)
- **Type checking** (iteration over mixed content)

## Successful Pattern from BlockStmt
```go
type BlockStmt struct {
    *BaseNode
    children   []Node          // Private - all tokens in order (concrete syntax)
    statements []StatementNode // Private - cached logical statements only
}

func (bs *BlockStmt) AddChild(child Node) { /* handles both collections */ }
func (bs *BlockStmt) Children() []Node { return bs.children }
func (bs *BlockStmt) LogicalStatements() []StatementNode { return bs.statements }
func (bs *BlockStmt) Statements() []StatementNode { return bs.statements } // backward compatibility
```

## Critical Issues Found

### 1. **SubDecl.Parameters** - MOST CRITICAL ⚠️
**Location**: `internal/ast/statements.go:153-258`
**Current Problem**:
```go
Parameters []*Parameter // Public field with potential structural token contamination
```
**Problematic Usage**:
- `ast_compiler.go:343`: `for i, param := range subDecl.Parameters`
- `navigator.go:137`: `for _, param := range typed.Parameters`
- `binder.go:161`: `for _, param := range node.Parameters`

**Risk**: High - Parameter parsing involves commas, parentheses, default values
**Impact**: Compilation failures when structural tokens are treated as parameters

### 2. **VarDecl.Variables** - HIGH PRIORITY
**Location**: `internal/ast/statements.go:67-149`
**Current Problem**:
```go
Variables []*VariableExpr // Public field accessed directly
```
**Problematic Usage**:
- Multiple parser accesses to `decl.Variables[0]`
- Binder iterates Variables directly

**Risk**: Medium - Variable declarations may mix structural elements
**Impact**: Variable processing errors

### 3. **ProgramStmt.Statements** - MEDIUM PRIORITY
**Location**: `internal/ast/statements.go:12-40`
**Current Problem**:
```go
Statements []StatementNode // Public field - same old pattern as BlockStmt
```
**Risk**: Medium - Top-level parsing could mix structural tokens
**Impact**: Program-level compilation issues

### 4. **ClassDecl** - LOW PRIORITY
**Location**: `internal/ast/statements.go:636-717`
**Current Problem**:
```go
Fields  []*FieldDecl     // Public fields
Methods []*MethodDecl
Roles   []*TypeExpression
```
**Risk**: Low - Class structure less likely to have mixing issues
**Impact**: Type system and OOP feature processing

### 5. **RoleDecl** - LOW PRIORITY
**Location**: `internal/ast/statements.go:719-795`
**Current Problem**:
```go
RequiredMethods []*MethodSignature // Public fields
ProvidedMethods []*MethodDecl
Fields          []*FieldDecl
```
**Risk**: Low - Similar to ClassDecl
**Impact**: Role/trait system processing

## Implementation TODO List

### Phase 1: SubDecl.Parameters (Critical - Start Here)
- [ ] **1.1** Create new SubDecl structure with private fields
  - [ ] Add `parameters []Node` (private - all tokens)
  - [ ] Add `logicalParameters []*Parameter` (private - cached logical parameters)
- [ ] **1.2** Implement encapsulated methods
  - [ ] `AddParameter(param Node)` - handles both collections
  - [ ] `LogicalParameters() []*Parameter` - returns logical parameters only
  - [ ] `Parameters() []*Parameter` - backward compatibility
  - [ ] `AllTokens() []Node` - for source-to-source compilation
- [ ] **1.3** Update parser to use new pattern
  - [ ] Modify parameter parsing to use `AddParameter()`
  - [ ] Ensure structural tokens (commas, parens) are properly categorized
  - [ ] Test that default values are preserved correctly
- [ ] **1.4** Update all consumers to use new API
  - [ ] `ast_compiler.go:343` - change to `LogicalParameters()`
  - [ ] `navigator.go:137` - change to `LogicalParameters()`
  - [ ] `binder.go:161` - change to `LogicalParameters()`
  - [ ] Update any test files using direct field access
- [ ] **1.5** Test and validate
  - [ ] Run existing tests to ensure no regressions
  - [ ] Test parameter parsing with default values
  - [ ] Test signature compilation edge cases

### Phase 2: VarDecl.Variables (High Priority)
- [ ] **2.1** Create new VarDecl structure with private fields
  - [ ] Add `variables []Node` (private)
  - [ ] Add `logicalVariables []*VariableExpr` (private)
- [ ] **2.2** Implement encapsulated methods
  - [ ] `AddVariable(variable Node)`
  - [ ] `LogicalVariables() []*VariableExpr`
  - [ ] `Variables() []*VariableExpr` - backward compatibility
- [ ] **2.3** Update parser and consumers
  - [ ] Modify variable declaration parsing
  - [ ] Update all direct field access locations
- [ ] **2.4** Test and validate

### Phase 3: ProgramStmt.Statements (Medium Priority)
- [ ] **3.1** Apply same pattern as BlockStmt
  - [ ] Add private `statements` and `children` fields
  - [ ] Implement `AddStatement()`, `LogicalStatements()`, `Statements()`
- [ ] **3.2** Update parser and consumers
- [ ] **3.3** Test program-level compilation

### Phase 4: ClassDecl (Low Priority)
- [ ] **4.1** Encapsulate Fields, Methods, Roles collections
- [ ] **4.2** Implement appropriate access methods
- [ ] **4.3** Update OOP-related code

### Phase 5: RoleDecl (Low Priority)
- [ ] **5.1** Encapsulate RequiredMethods, ProvidedMethods, Fields
- [ ] **5.2** Implement appropriate access methods
- [ ] **5.3** Update role/trait system code

## Testing Strategy

### For Each Phase:
1. **Regression Testing**: Ensure all existing tests still pass
2. **Direct Access Testing**: Verify no direct field access remains
3. **Parser Testing**: Test that structural tokens are properly separated
4. **Compilation Testing**: Test source-to-source transformation works correctly
5. **Edge Case Testing**: Test complex cases (nested structures, default values, etc.)

## Success Criteria

✅ **Structural Separation**: Logical content and structural tokens properly separated
✅ **API Consistency**: All statement types follow the same encapsulation pattern
✅ **Backward Compatibility**: Existing code works with new APIs
✅ **Source Preservation**: Source-to-source compilation preserves exact formatting
✅ **Performance**: No significant performance regressions
✅ **Test Coverage**: All changes covered by tests

## Implementation Notes

### Pattern Template:
```go
type XStatement struct {
    *BaseNode
    children     []Node           // Private - all tokens in order
    logicalItems []LogicalType    // Private - cached logical items only
}

func (xs *XStatement) AddChild(child Node) {
    xs.children = append(xs.children, child)
    if logicalItem, ok := child.(LogicalType); ok {
        xs.logicalItems = append(xs.logicalItems, logicalItem)
    }
    xs.BaseNode.AddChild(child) // Backward compatibility
}

func (xs *XStatement) LogicalItems() []LogicalType { return xs.logicalItems }
func (xs *XStatement) Items() []LogicalType { return xs.logicalItems } // Backward compatibility
func (xs *XStatement) Children() []Node { return xs.children }
```

### Parser Integration:
- Use `AddChild()` for all nodes during parsing
- Let the method handle logical vs structural token separation
- Ensure proper token type identification in tree-sitter conversion

### Factory Method Updates:
- Initialize private fields in constructors
- Use `AddChild()` instead of direct field assignment
- Maintain backward compatibility for existing factory consumers
