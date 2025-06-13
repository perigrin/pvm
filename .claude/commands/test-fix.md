1. Run `make test` to identify current failing tests and get comprehensive failure breakdown
2. Categorize failures by type:
   - **Expectation mismatch**: "Expected error but parsing succeeded" (quick wins)
   - **Grammar missing**: ERROR nodes, unexpected tokens, unsupported constructs
   - **Type annotation issues**: UnknownTypeError, complex type parsing failures
   - **Architecture problems**: Symbol binding, compilation, or structural issues
3. Start with quick wins (test expectation fixes):
   - Review tests that expect failures but now succeed due to improved parsing
   - Remove `<!-- should_error: true -->` from tests that should pass
   - Update expected outcomes to match successful parsing
4. For each remaining failure:
   - Create minimal reproduction case if needed
   - Determine root cause (grammar, type system, or logic)
   - Fix incrementally with focused changes
   - Test fix against related test cases
5. Ensure 100% pass rate before proceeding to next category
6. Commit changes with clear descriptions of what was fixed
7. Run `make test` again to verify no regressions
8. Update any relevant documentation or plans to reflect completed fixes
