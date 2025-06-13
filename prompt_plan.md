# Parser Test Failure Resolution - Prompt Plan

## Overview

This plan addresses the resolution of 50 failing parser tests in the PVM project's tree-sitter-typed-perl grammar. The failures are caused by limitations in parsing valid untyped Perl constructs.

**Current Status**:
- Total Parser Tests: 908
- Failing Tests: 50
- Pass Rate: 94.5%

**Failure Categories**:
1. Given/When Constructs (11 failures)
2. Package-Qualified Variables (4 failures)
3. Control Flow Statements (35 failures)

## Implementation Phases

### Phase 1: Test Infrastructure Setup (Steps 1-2)
Establish the testing foundation to ensure we can validate each change incrementally.

### Phase 2: Given/When Support (Steps 3-5)
Add support for Perl's switch statement syntax to the grammar.

### Phase 3: Package-Qualified Variables (Steps 6-8)
Fix the grammar to support `our $Package::qualified;` syntax.

### Phase 4: Control Flow Completion (Steps 9-11)
Address remaining control flow constructs and edge cases.

### Phase 5: Integration and Validation (Steps 12-13)
Ensure all fixes work together and maintain performance.

---

## Detailed Steps

### Step 1: Create Test Infrastructure for Grammar Changes ✅ COMPLETED

```text
I need to set up a test infrastructure for tree-sitter grammar changes in the PVM project.

Context:
- We have 50 failing parser tests due to grammar limitations
- The tree-sitter-typed-perl grammar is in tree-sitter-typed-perl/
- Tests are currently in internal/parser/testdata/untyped-perl/

Please help me:
1. Create a new test file tree-sitter-typed-perl/test/corpus/untyped_perl_fixes
2. Add a simple test case for package-qualified variables:
   ```
   ================
   Package qualified our declaration
   ================
   our $Package::qualified;
   ---
   (source_file
     (variable_declaration
       (scalar
         (varname))))
   ```
3. Run the test with tree-sitter test and show me the actual output vs expected
4. Create a helper script to quickly test grammar changes

The goal is to have a fast feedback loop for grammar development.
```

**COMPLETED**: Test infrastructure successfully created with:
- Test corpus file for untyped Perl fixes with multiple test cases
- Helper script test_grammar.sh for quick testing
- Documentation for tree-sitter test corpus
- Package-qualified variable tests passing

### Step 2: Add Debugging Tools for Grammar Development ✅ COMPLETED

```text
I need debugging tools to understand why certain constructs fail to parse in tree-sitter-typed-perl.

Context:
- Package-qualified variables like "our $Package::qualified;" create ERROR nodes
- The grammar has different rules for declarations vs expressions
- We need to trace how tokens are consumed

Please help me:
1. Create a debug script that shows:
   - Token stream for a given input
   - Which grammar rules are attempted
   - Where parsing fails
2. Add a debug mode to tree-sitter-typed-perl that logs rule attempts
3. Create a visualization tool that shows the parse tree with ERROR nodes highlighted
4. Document common parsing failure patterns

Use "our $Package::qualified;" as the test case.
```

**COMPLETED**: Successfully created comprehensive debugging tools:
- **debug_grammar.js**: Shows token streams, grammar rules, and detailed parsing analysis
- **visualize_tree.js**: Visualizes parse trees with ERROR node highlighting in multiple formats
- **PARSING_FAILURE_PATTERNS.md**: Documents common parsing failures and solutions
- Both tools support interactive mode, file input, and various output formats

### Step 3: Add Given/When Grammar Rules ✅ COMPLETED

```text
I need to add support for Perl's given/when switch statement syntax to tree-sitter-typed-perl.

Context:
- 11 tests fail because given/when is not supported
- This is similar to switch/case in other languages
- We need to handle both given blocks and when clauses

Please help me add these grammar rules:
1. Define given_statement rule that accepts:
   - given keyword
   - Expression in parentheses
   - Block containing when clauses
2. Define when_clause rule for:
   - when keyword
   - Condition expression
   - Block of statements
3. Define default_clause for default blocks
4. Add these to the statement choices
5. Create comprehensive test cases

Ensure the rules integrate with existing expression and block rules.
```

**COMPLETED**: Successfully implemented given/when/default grammar support:
- Added `given_statement` rule with condition expression and special given_block body
- Added `given_block` that accepts when_clause, default_clause, and regular statements
- Added `when_clause` with condition expression and regular block body
- Added `default_clause` with regular block body
- Created comprehensive test corpus with basic, complex, nested, and break statement cases
- All tests passing

### Step 4: Test Given/When Implementation

```text
I need to thoroughly test the given/when grammar implementation.

Context:
- We just added given/when rules to the grammar
- Need to ensure all edge cases work
- Must verify integration with existing features

Please create test cases for:
1. Basic given/when with scalar matching
2. Given/when with array/hash matching
3. Nested given/when statements
4. Given/when with complex expressions
5. Default clause handling
6. Given/when inside subroutines
7. Break/continue in when blocks

Run all tests and fix any failing cases.
```

### Step 5: Integrate Given/When with Parser Tests

```text
I need to integrate the given/when grammar changes with the Go parser tests.

Context:
- Grammar changes are complete
- Go bindings need to be regenerated
- Parser tests need to be updated

Please help me:
1. Regenerate the Go bindings with `make tree-sitter`
2. Run the specific failing given/when tests
3. Update test expectations if needed
4. Verify no regression in other tests
5. Document any behavioral changes

Focus on the 11 control flow tests that were failing.
```

### Step 6: Analyze Package-Qualified Variable Grammar Issue

```text
I need to understand why package-qualified variables fail in declaration context.

Context:
- "our $Package::qualified;" creates ERROR nodes
- "$Package::qualified" works fine in expressions
- The issue is specific to variable declarations

Please help me:
1. Trace through the grammar rules for:
   - variable_declaration
   - _declare_scalar
   - varname
   - _identifier vs _bareword
2. Identify why :: is not accepted in varname
3. Compare with how :: works in expressions
4. Propose grammar changes that won't break existing functionality

Show the exact token consumption process.
```

### Step 7: Implement Package-Qualified Variable Support

```text
I need to fix the grammar to support package-qualified variables in declarations.

Context:
- varname currently only accepts _identifier
- We need to support :: in variable names
- Must work for my, our, state, and local

Please implement:
1. Extend varname to accept qualified names:
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
2. Handle potential conflicts with attribute syntax
3. Ensure :: is treated as a single unit
4. Test with all declaration keywords

Verify the fix works for our test case.
```

### Step 8: Test Package-Qualified Variables Thoroughly

```text
I need comprehensive tests for package-qualified variable declarations.

Context:
- We just added support for qualified names in declarations
- Need to ensure all contexts work correctly
- Must verify no regression

Please create tests for:
1. Simple qualified: our $Foo::bar;
2. Deep nesting: my $A::B::C::D::var;
3. With initialization: our $Pkg::var = 42;
4. In lists: my ($Foo::x, $Bar::y);
5. With type annotations: my Int $Pkg::typed;
6. Local declarations: local $Module::setting;
7. In different scopes

Ensure all tests pass and integrate with parser.
```

### Step 9: Identify Remaining Control Flow Failures

```text
I need to catalog and categorize the remaining 35 control flow test failures.

Context:
- We've fixed given/when (11 tests)
- We've fixed package variables (4 tests)
- 35 failures remain

Please help me:
1. Run all failing control flow tests
2. Group them by failure pattern:
   - Missing grammar rules
   - Tokenization issues
   - Ambiguous syntax
3. Prioritize by complexity
4. Create minimal reproductions
5. Identify common patterns

Provide a fix strategy for each group.
```

### Step 10: Implement High-Priority Control Flow Fixes

```text
I need to fix the highest priority control flow issues identified.

Context:
- We have categorized the 35 remaining failures
- Starting with the most common patterns
- Need incremental fixes

Please implement fixes for:
1. [Specific pattern 1 from Step 9]
2. [Specific pattern 2 from Step 9]
3. [Specific pattern 3 from Step 9]

For each fix:
- Add grammar rules
- Create test cases
- Verify no regression
- Document the change

Run tests after each fix to ensure progress.
```

### Step 11: Complete Remaining Control Flow Fixes

```text
I need to complete all remaining control flow fixes.

Context:
- We've fixed the high-priority issues
- Some edge cases remain
- Need to achieve 100% pass rate

Please:
1. Implement fixes for all remaining patterns
2. Handle any ambiguous cases with precedence rules
3. Add comprehensive test coverage
4. Verify each fix independently
5. Run full test suite

Document any limitations or future work needed.
```

### Step 12: Performance Validation and Optimization

```text
I need to ensure the grammar changes don't impact parser performance.

Context:
- We've added significant grammar complexity
- Parser performance is critical for LSP
- Need to maintain <5% regression

Please:
1. Run performance benchmarks:
   - Baseline (before changes)
   - With all fixes
   - Compare results
2. Identify any bottlenecks
3. Optimize grammar rules if needed:
   - Reduce ambiguity
   - Simplify complex patterns
   - Use precedence effectively
4. Re-run benchmarks
5. Document performance characteristics

Ensure we meet performance requirements.
```

### Step 13: Final Integration and Documentation

```text
I need to complete the integration and update all documentation.

Context:
- All grammar fixes are complete
- Tests are passing
- Performance is validated

Please help me:
1. Run the complete test suite:
   - Parser tests: `go test ./internal/parser`
   - E2E tests: `go test ./test/e2e`
   - All tests: `make test`
2. Update documentation:
   - Grammar changes in README
   - Supported Perl constructs
   - Migration guide for users
3. Create examples of newly supported syntax
4. Update CLAUDE.md with any new patterns
5. Prepare a summary of all changes

Verify we've achieved 100% parser test pass rate.
```

---

## Success Criteria

1. **All 50 parser tests pass** (100% pass rate)
2. **No regression** in existing tests
3. **Performance within 5%** of baseline
4. **Documentation complete** and accurate
5. **Examples provided** for all new constructs

## Implementation Notes

- Each step builds on the previous one
- Test-driven development throughout
- Small, incremental changes
- Continuous validation
- No orphaned code

## Risk Mitigation

- Create backups before each grammar change
- Test each change in isolation
- Use feature branches for complex changes
- Profile performance regularly
- Document all decisions
