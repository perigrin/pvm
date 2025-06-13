# Type Annotation Extraction Fix - Implementation Plan

## Project Context

This plan focuses on fixing the immediate bug in type annotation extraction that's causing 203+ parser test failures. The tree-sitter grammar correctly parses typed Perl code, but the type annotation extraction logic is broken due to a field name mismatch.

## Root Cause Analysis

**Problem**: Type annotation extraction fails because the code tries to access `node.ChildByFieldName("type")` but the tree-sitter grammar doesn't define a field name for `type_expression`.

**Evidence**: 
- Tree-sitter correctly parses `my Int $count = 42;` with `type_expression` and `scalar` children
- AST.TypeAnnotations slice remains empty when it should contain extracted annotations
- 203 parser test failures related to missing type annotations

## Target Architecture

**Goal**: Fix type annotation extraction to populate AST.TypeAnnotations correctly
**Strategy**: Small, targeted fixes with comprehensive testing  
**Target Outcome**: Major reduction in parser test failures (203 → ~50-100)

---

## Step-by-Step Implementation Plan

### ✅ Step 1: Fix Field-Based Type Extraction (COMPLETED)

```
You are fixing a critical bug in the PVM parser's type annotation extraction logic.

CONTEXT: The parser correctly uses tree-sitter to parse typed Perl code like `my Int $count = 42;`, but the TypeAnnotations slice in the AST remains empty. The tree-sitter grammar parses this correctly, creating a `variable_declaration` with `type_expression` and `scalar` children, but the extraction code is broken.

PROBLEM: In `/home/perigrin/dev/pvm/internal/parser/treesitter/perl.go`, the functions `processVariableDeclaration()` and `processTypedVariableDeclaration()` use `node.ChildByFieldName("type")` to find type expressions. However, the tree-sitter grammar defines `type_expression` without a field name, so this call always returns nil.

TASK: Fix the type extraction logic to properly find type_expression nodes.

REQUIREMENTS:
1. Examine the current implementation in `internal/parser/treesitter/perl.go`
2. Identify the lines using `node.ChildByFieldName("type")` that always return nil
3. Replace field-based access with manual iteration to find `type_expression` nodes by type
4. Keep the working `node.ChildByFieldName("variable")` calls unchanged (these work correctly)
5. Ensure the fix works for both `processVariableDeclaration()` and `processTypedVariableDeclaration()`

TECHNICAL DETAILS:
- The tree-sitter grammar has `field('variable', $._declared_vars)` but no field name for `type_expression`
- Replace `typeNode := node.ChildByFieldName("type")` with iteration to find nodes where `child.Type() == "type_expression"`
- The `variableNode := node.ChildByFieldName("variable")` calls should remain unchanged
- Ensure backward compatibility with existing fallback logic

EXPECTED OUTCOME:
- `TestParser_Baselines/type_annotations` should start showing extracted type annotations
- AST.TypeAnnotations should contain entries like `VarAnnotation: $count :: Int at 1:1`
- Major reduction in parser test failures from ~203 to much lower number

Focus on the minimal fix to get type annotation extraction working. The tree-sitter parsing is correct - only the extraction logic needs fixing.
```

### Step 2: Verify Type Annotation Conversion

```
You are continuing the type annotation extraction fix from Step 1.

CONTEXT: The field-based type extraction has been fixed. Now verify that the extracted type information is properly converted from tree-sitter format to the internal AST format.

TASK: Ensure the type annotation conversion pipeline works correctly.

REQUIREMENTS:
1. Test the basic type extraction with a simple example: `my Int $count = 42;`
2. Verify that `convertToASTTypeAnnotation()` correctly converts tree-sitter TypeAnnotation to ast.TypeAnnotation
3. Check that enum conversion between `treesitter.AnnotationKind` and `ast.AnnotationKind` works properly
4. Ensure position information is correctly preserved in the conversion
5. Test that the TypeAnnotation.String() method produces the expected output format

TECHNICAL DETAILS:
- Run the test: `go test -v ./internal/parser -run TestParser_Baselines/type_annotations`
- The expected output should show: `VarAnnotation: $count :: Int at 1:1`
- Check that the conversion in `convertToASTTypeAnnotation()` handles all annotation kinds correctly
- Verify position information is accurate (line:column matches source code)

VALIDATION STEPS:
- Create a simple test with `my Int $count = 42;`
- Verify AST.TypeAnnotations contains exactly one annotation
- Check that annotation.AnnotatedItem == "$count"
- Check that annotation.TypeExpression.Name == "Int"
- Check that annotation.Kind == ast.VarAnnotation
- Check that position is (1,1) for the first line

SUCCESS CRITERIA:
- Type annotation extraction produces correct TypeAnnotation objects
- Conversion from tree-sitter to AST format preserves all information
- Position information is accurate
- String representation matches expected baseline format

This ensures the conversion pipeline is working correctly after the extraction fix.
```

### Step 3: Test Complex Type Expressions

```
You are continuing the type annotation extraction fix from Step 2.

CONTEXT: Basic type annotation extraction is working. Now test that complex type expressions are properly extracted and converted.

TASK: Verify complex type expressions work correctly with the fixed extraction logic.

REQUIREMENTS:
1. Test parameterized types: `my ArrayRef[Int] @numbers;`
2. Test union types: `my Int|Str $flexible;`  
3. Test custom types: `my MyClass $object;`
4. Test multiple annotations in one file
5. Verify that all complex type structures are preserved in AST

TEST CASES TO VERIFY:
```perl
my Int $count = 42;
my Str $message = "hello";
my ArrayRef[Int] @numbers = [1, 2, 3];
my HashRef[Str] %config = {};
my MyClass $object;
```

TECHNICAL DETAILS:
- Each type expression should be correctly parsed and stored
- Parameterized types should have Parameters populated
- Union types should have UnionTypes populated  
- Position information should be accurate for each annotation
- Test that the extraction handles multiple types in one file

VALIDATION:
- Run test with the complex example above
- Verify AST.TypeAnnotations contains 5 annotations
- Check that parameterized types preserve parameter information
- Ensure all type names are correctly extracted
- Verify position information is accurate for each annotation

SUCCESS CRITERIA:
- All basic and complex type expressions extract correctly
- Parameterized types maintain their parameter information
- Multiple annotations in one file are all captured
- Complex type structures are preserved in the AST

This ensures the fix works for realistic typed Perl code patterns.
```

### Step 4: Run Full Parser Test Suite

```
You are completing the type annotation extraction fix from Step 3.

CONTEXT: Type annotation extraction is working for basic and complex cases. Now run the full parser test suite to measure the impact of the fix.

TASK: Run the complete parser tests and measure the improvement in pass rate.

REQUIREMENTS:
1. Run the full parser test suite: `make test` or equivalent
2. Compare the results to the baseline of 203 failures out of ~900 parser tests
3. Identify remaining failure categories and root causes
4. Document the improvement in test pass rate
5. Identify any regressions or new issues introduced by the fix

EXECUTION STEPS:
- Run: `go test -v ./internal/parser -count=1`
- Capture the test output and failure summary
- Calculate the new failure count and pass rate
- Categorize remaining failures by type/cause
- Compare to previous baseline of 203 failures

EXPECTED IMPROVEMENTS:
- **Before**: ~203 failures related to missing type annotations
- **After**: Significant reduction, possibly to 50-100 failures
- **Categories that should improve**:
  - TestParser_Baselines/* tests
  - TestParser_SpecificBaselines/* tests  
  - Type annotation extraction tests
  - Variable declaration parsing tests

ANALYSIS REQUIRED:
- Calculate exact improvement: old_failures - new_failures
- Identify what types of tests are still failing
- Determine if failures are due to:
  - Advanced features not yet implemented (classes, roles)
  - Grammar limitations  
  - Other extraction issues
  - Test expectation problems

SUCCESS CRITERIA:
- Major reduction in parser test failures (target: 50%+ improvement)
- No new test regressions introduced
- Clear understanding of remaining failure categories
- Documented improvement metrics

This validates that the type annotation extraction fix has the expected impact on overall parser test reliability.
```

### Step 5: Update Test Documentation

```
You are completing the type annotation extraction fix from Step 4.

CONTEXT: The type annotation extraction fix is complete and tested. The parser test suite shows significant improvement. Now update documentation to reflect the current status.

TASK: Update test documentation and progress tracking with the current status.

REQUIREMENTS:
1. Update `todo_tests.md` with the new test results and failure counts
2. Document the specific fix that was implemented
3. Update the failure breakdown by package with new numbers
4. Document the root causes of remaining failures
5. Update progress metrics and next priorities

DOCUMENTATION UPDATES:
- Update total test count and pass rate
- Update parser package failure count (should be significantly reduced)
- Document the type annotation extraction fix
- Update priority focus areas based on new results
- Revise target goals based on current progress

KEY METRICS TO UPDATE:
- Total tests and pass rate
- Parser package failures (was 203, now should be much lower)
- Overall failure breakdown by package
- Progress since last update
- Next highest priority areas

ANALYSIS TO INCLUDE:
- What was fixed: Type annotation extraction field name mismatch
- Impact: Reduction from 203 to X parser failures  
- Remaining issues: Categories of tests still failing
- Next steps: Priority areas for further improvement

SUCCESS CRITERIA:
- Documentation accurately reflects current test status
- Fix is properly documented for future reference
- Next priority areas are clearly identified
- Progress metrics show the improvement achieved

This ensures accurate tracking of progress and clear guidance for future work.
```

---

## Success Metrics & Validation

### Target Improvements
- **Parser Test Failures**: Reduce from 203 to 50-100 (50%+ improvement)
- **Type Annotation Tests**: All basic type annotation tests should pass
- **Baseline Tests**: Type annotation baseline tests should show expected output
- **No Regressions**: Existing passing tests should continue to pass

### Validation Steps
1. **Unit Testing**: Individual type extraction functions work correctly
2. **Integration Testing**: Full parser pipeline processes type annotations
3. **Regression Testing**: No existing functionality is broken
4. **Performance Testing**: No significant performance degradation

### Key Files Modified
- `/home/perigrin/dev/pvm/internal/parser/treesitter/perl.go` - Fix field-based type extraction
- Potentially type conversion functions if enum mapping issues found
- Test documentation updates

## Completion Criteria

The type annotation extraction fix is successful when:

1. **Immediate Fix**: `TestParser_Baselines/type_annotations` passes
2. **Major Improvement**: Parser test failures reduced by 50%+ 
3. **No Regressions**: Existing tests continue to pass
4. **Complex Types**: Parameterized and union types extract correctly
5. **Documentation**: Progress is accurately documented

This focused plan should resolve the immediate blocker preventing type annotation extraction and significantly improve the parser test pass rate.