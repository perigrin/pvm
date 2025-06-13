1. Identify the specific parsing issue:
   - **Error nodes**: Tree-sitter detected parse errors
   - **Unexpected tokens**: Grammar doesn't recognize valid Perl syntax
   - **Type annotation failures**: Type extraction or validation problems
   - **AST structure issues**: Incorrect node relationships or missing information
2. Create minimal reproduction case:
   - Add focused test case in appropriate testdata directory
   - Use simplest possible code that demonstrates the issue
   - Ensure test case isolates the specific problem
3. Determine issue category and approach:
   - **Grammar issue**: Update tree-sitter-typed-perl/grammar.js
   - **Type annotation issue**: Check type extraction logic in parser
   - **AST compilation issue**: Review compiler package logic
4. For grammar issues:
   - Study Perl language specification for correct syntax
   - Update grammar.js with new rules or fix existing ones
   - Run `make tree-sitter` to regenerate parser
   - Test against comprehensive examples, not just the failing case
5. For type annotation issues:
   - Check type annotation extraction in parser package
   - Verify type validation logic handles the specific case
   - Ensure error messages are clear and actionable
6. Test fix comprehensively:
   - Run specific failing test to verify fix
   - Run full test suite to ensure no regressions
   - Test edge cases and related functionality
7. Verify no performance degradation with the fix
8. Document any grammar extensions or parsing improvements made
