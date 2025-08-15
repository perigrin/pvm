# Test Fix Plan

## Current Status

### Passing Packages ✅
- `internal/ast` - 100% passing
- `internal/compiler` - 100% passing

### Failing Packages ❌

#### Parser (`internal/parser`)
- **Failures**: 9 typed-perl corpus tests
- **Root Cause**: Test expectations don't match new AST structure
- **Issues**:
  - `return_stmt` vs `expression_stmt` mismatches
  - `literal` vs `variable` mismatches
  - Comments/print statements incorrectly expected as return statements

#### TypeChecker (`internal/typechecker`)
- **Failures**: 27 test groups
- **Root Cause**: Flow analysis not detecting patterns after AST changes
- **Main Issues**:
  - Safety analysis tests (unsafe hash access, array access, etc.)
  - Type inference tests
  - Exception flow tracking

## Fix Strategy

### Phase 1: Parser Test Fixes (Quick Wins)
1. Fix the 9 remaining typed-perl corpus tests
2. These are all expectation mismatches, not functionality issues
3. Each requires manual review of the test file

### Phase 2: TypeChecker Fixes (More Complex)
1. Update flow analysis to work with new AST structure
2. Fix safety analysis detection patterns
3. Update type inference logic

## Specific Parser Tests to Fix

1. `method-signatures-unions` - testdata/corpus/parser/typed-perl/union-types/method-signatures-unions.md
2. `basic-class-declarations` - testdata/corpus/parser/typed-perl/classes-roles/basic-class-declarations.md
3. `method-signatures` - testdata/corpus/parser/typed-perl/parameterized-types/method-signatures.md
4. `generic-class-declarations` - testdata/corpus/parser/typed-perl/classes-roles/generic-class-declarations.md
5. `mixed-typed-untyped` - testdata/corpus/parser/typed-perl/methods-fields/mixed-typed-untyped.md
6. `complex-method-signatures` - testdata/corpus/parser/typed-perl/methods-fields/complex-method-signatures.md
7. `class-inheritance` - testdata/corpus/parser/typed-perl/classes-roles/class-inheritance.md
8. `class-context-methods` - testdata/corpus/parser/typed-perl/methods-fields/class-context-methods.md
9. `constructor-destructor-methods` - testdata/corpus/parser/typed-perl/classes-roles/constructor-destructor-methods.md

## Next Steps

1. Fix parser tests one by one (manual review required)
2. Once parser is 100%, focus on typechecker
3. Commit fixes incrementally
