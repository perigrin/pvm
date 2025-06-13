# Common Parsing Failure Patterns in tree-sitter-typed-perl

This document catalogs common parsing failure patterns found in the tree-sitter-typed-perl grammar and their solutions.

## 1. Package-Qualified Variables

### Pattern
```perl
our $Package::qualified;
my $Foo::Bar::var = 42;
```

### Issue
The grammar's `varname` rule only accepts `_identifier`, not qualified names with `::`.

### Symptoms
- ERROR nodes in variable declarations
- Works fine in expressions but fails in declarations

### Solution
Extend `varname` to accept qualified names:
```javascript
varname: $ => choice(
  $._identifier,
  $.qualified_name,
  $._ident_special
),

qualified_name: $ => seq(
  $._identifier,
  repeat1(seq('::', $._identifier))
)
```

## 2. Given/When Constructs

### Pattern
```perl
given ($value) {
    when (/pattern/) { ... }
    when ($_ > 10) { ... }
    default { ... }
}
```

### Issue
Missing grammar rules for Perl's switch statement syntax.

### Symptoms
- Entire given block parsed as ERROR
- 11 test failures in control flow tests

### Solution
Add dedicated given/when rules to the grammar.

## 3. Method Return Type Annotations

### Pattern
```perl
method foo() -> Int { ... }
sub bar() :returns(Str) { ... }
```

### Issue
The arrow syntax for return types not fully supported in method signatures.

### Symptoms
- ERROR node after parameter list
- Return type not recognized

### Solution
Extend method_signature to handle return type syntax.

## 4. Complex Type Expressions in Assertions

### Pattern
```perl
$value as (Int|Str)
$obj as ArrayRef[HashRef[Int]]
```

### Issue
Type assertions with parentheses or complex nested types may fail.

### Symptoms
- Parentheses in type expressions cause errors
- Nested parameterized types partially parsed

### Solution
Ensure type_expr handles full precedence and nesting.

## 5. Slurpy Parameters in Signatures

### Pattern
```perl
method foo($x, @rest) { ... }
sub bar($a, %opts) { ... }
```

### Issue
Array and hash slurpy parameters in signatures not fully supported.

### Symptoms
- ERROR nodes on @rest or %opts
- Parameter list parsing fails

### Solution
Add slurpy parameter support to signature rules.

## 6. Package Blocks with Version

### Pattern
```perl
package Foo::Bar 1.23 {
    ...
}
```

### Issue
Package declarations with version numbers before block.

### Symptoms
- Version number causes ERROR
- Block not associated with package

### Solution
Update package_statement to optionally accept version.

## 7. Regex Modifiers

### Pattern
```perl
/pattern/msixpodual
m{pattern}gcs
```

### Issue
Not all regex modifiers recognized.

### Symptoms
- Some modifiers cause ERROR nodes
- Regex parsing incomplete

### Solution
Ensure all Perl regex modifiers are in the grammar.

## 8. Special Variables in Interpolation

### Pattern
```perl
"Value: ${"
"Array: @{"
```

### Issue
Special interpolation syntax with braces.

### Symptoms
- Interpolation sequences not recognized
- String parsing fails

### Solution
Extend interpolation rules to handle brace syntax.

## Debugging Tips

1. **Use the debug tools**:
   ```bash
   ./debug_grammar.js "problematic code here"
   ./visualize_tree.js "code" --errors-only
   ```

2. **Check token stream**:
   ```bash
   ./debug_grammar.js "code" --tokens
   ```

3. **Compare with working code**:
   - If `$var` works but `$Package::var` doesn't, the issue is in qualified names
   - If expressions work but declarations fail, check declaration-specific rules

4. **Look for precedence issues**:
   - Parentheses often indicate precedence problems
   - Try simplifying complex expressions to isolate the issue

5. **Test incrementally**:
   - Start with simplest form that works
   - Add complexity step by step
   - Identify exactly where parsing breaks

## Testing After Grammar Changes

Always run after making grammar changes:
```bash
# Regenerate parser
cd tree-sitter-typed-perl
tree-sitter generate

# Test specific constructs
tree-sitter test -f test/corpus/untyped_perl_fixes

# Run full test suite
cd ..
make test
```
