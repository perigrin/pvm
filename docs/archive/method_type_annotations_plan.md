# Method Type Annotations Implementation Blueprint ✅ COMPLETED

## Project Overview

We need to implement parsing support for method type annotations in the PVM (Perl Version Manager) type system. This includes:
- Method parameter types: `method foo(Type $param)`
- Method return types: `method foo() -> ReturnType`
- Combined syntax: `method foo(Type $param) -> ReturnType`

## High-Level Architecture

The implementation spans two main components:
1. **Tree-sitter Grammar Extension** - Add parsing rules for typed method syntax
2. **Parser Implementation** - Extract and convert parsed annotations to type system objects

## Detailed Step-by-Step Blueprint

### Phase 1: Grammar Foundation (Steps 1-3)
**Goal**: Extend tree-sitter grammar to parse method type annotations

### Phase 2: Parser Implementation (Steps 4-6)
**Goal**: Implement extraction and conversion of parsed method annotations

### Phase 3: Integration & Testing (Steps 7-8)
**Goal**: Wire everything together and ensure robust testing

---

## Implementation Steps

### Step 1: Add Typed Method Parameter Grammar Rules ✅ COMPLETED

**Scope**: Extend tree-sitter grammar to parse `Type $param` syntax in method signatures
**Risk**: Low - Additive grammar change
**Test Strategy**: Grammar-level parsing tests

```text
Implement tree-sitter grammar rules for typed method parameters.

Context: We need to extend the existing method signature parsing in tree-sitter-typed-perl to support type annotations like `method foo(Type $param, ArrayRef[Int] $data)`.

Current State:
- The grammar has `method_declaration_statement` and `signature` rules
- Signatures support basic parameters but not typed parameters
- The `_signature_vars` rule needs extension

Requirements:
1. Add a `typed_method_parameter` rule that matches `Type $variable` syntax
2. Extend `_signature_vars` to include typed parameters
3. Ensure the new rule works with existing parameter types (scalar, array, hash)
4. Maintain backwards compatibility with existing signature parsing

Implementation:
1. In tree-sitter-typed-perl/grammar.js, add the new rule:
   ```javascript
   typed_method_parameter: $ => seq(
     field('type', $.type_expression),
     field('parameter', choice(
       alias($._signature_scalar, $.scalar),
       alias($._signature_array, $.array),
       alias($._signature_hash, $.hash)
     ))
   ),
   ```

2. Update `_signature_vars` to include the new rule:
   ```javascript
   _signature_vars: $ => choice(
     $.mandatory_parameter,
     $.optional_parameter,
     $.slurpy_parameter,
     $.named_parameter,
     $.typed_method_parameter  // Add this line
   ),
   ```

3. Test with examples:
   - `method foo(Str $name)`
   - `method bar(HashRef[Int] $data, Bool $flag)`
   - `method baz($regular, Typed $typed)` (mixed parameters)

Testing:
- Create minimal test cases that parse method signatures with typed parameters
- Verify the AST structure contains the expected type_expression and parameter nodes
- Ensure backwards compatibility with existing method signature tests

Deliverable: Updated grammar.js with typed method parameter support that passes basic parsing tests.
```

### Step 2: Add Method Return Type Grammar Rules ✅ COMPLETED

**Scope**: Add grammar support for `-> ReturnType` syntax after method signatures
**Risk**: Low - Additive grammar change
**Test Strategy**: Grammar-level parsing tests

```text
Implement tree-sitter grammar rules for method return type annotations.

Context: We need to add support for method return type syntax like `method foo() -> ReturnType` and `method bar(Type $param) -> ReturnType`.

Current State:
- Method declarations support signatures but not return types
- Need to add optional return type parsing after signatures
- The `method_declaration_statement` rule needs extension

Requirements:
1. Add a `method_return_type` rule that matches `-> Type` syntax
2. Update `method_declaration_statement` to include optional return types
3. Ensure return types work with all method signature variations
4. Maintain proper precedence and parsing order

Implementation:
1. Add the method return type rule to grammar.js:
   ```javascript
   method_return_type: $ => seq(
     '->',
     field('return_type', $.type_expression)
   ),
   ```

2. Update the method declaration statement:
   ```javascript
   method_declaration_statement: $ => seq(
     optional(field('lexical', 'my')),
     subExtensions(),
     'method',
     field('name', $.bareword),
     optseq(':', optional(field('attributes', $.attrlist))),
     optional(choice($.prototype, $.signature)),
     optional(field('return_type', $.method_return_type)),  // Add this line
     field('body', $.block),
   ),
   ```

3. Test with examples:
   - `method foo() -> Str`
   - `method bar(Type $param) -> ReturnType`
   - `method baz() -> ArrayRef[Int]` (complex return types)

Testing:
- Create test cases for various return type syntax patterns
- Verify AST contains method_return_type nodes with proper type_expression children
- Test interaction between signatures and return types
- Ensure backwards compatibility with methods without return types

Deliverable: Updated grammar.js with method return type support that passes parsing tests.
```

### Step 3: Rebuild and Validate Grammar ✅ COMPLETED

**Scope**: Rebuild tree-sitter parser and validate grammar changes
**Risk**: Medium - Build process dependencies
**Test Strategy**: Integration tests with sample method code

```text
Rebuild the tree-sitter parser with the new grammar rules and validate the changes.

Context: After modifying the grammar, we need to rebuild the tree-sitter parser and ensure the changes work correctly with real method syntax examples.

Current State:
- Grammar rules have been added for typed parameters and return types
- The parser binary needs rebuilding to include the changes
- Need to validate that parsing works end-to-end

Requirements:
1. Rebuild the tree-sitter parser binary with the new grammar
2. Test parsing of complete method examples from the test case
3. Verify AST structure matches expectations
4. Ensure no regressions in existing method parsing

Implementation:
1. Follow the existing build process in the project:
   - Run `make tree-sitter` or equivalent build command
   - Ensure tree-sitter-cli is available and working
   - Verify the build completes without errors

2. Create comprehensive test examples:
   ```perl
   class TestClass {
       method Str simple() { }
       method with_params(Str $name, Int $age) { }
       method ArrayRef[Int] full_syntax(HashRef[Str] $data) { }
       method Bool mixed($regular, Typed $typed) { }
   }
   ```

3. Test parsing programmatically:
   - Use the tree-sitter Go bindings to parse test examples
   - Inspect the resulting AST structure
   - Verify that typed_method_parameter and method_return_type nodes appear correctly

Testing:
- Parse the complete test method examples from the failing test
- Verify AST contains all expected node types
- Check that existing method parsing still works
- Document any AST structure findings for the next step

Deliverable: Successfully rebuilt tree-sitter parser with validated grammar changes and documented AST structure.
```

### Step 4: Implement Method Signature Processing ✅ COMPLETED

**Scope**: Add parser functions to extract typed method parameters from AST
**Risk**: Medium - Parser logic complexity
**Test Strategy**: Unit tests for signature processing

```text
Implement parser functions to extract typed method parameter annotations from the tree-sitter AST.

Context: With the grammar changes in place, we need to implement the Go parser code that extracts typed method parameters and converts them to TypeAnnotation objects.

Current State:
- Tree-sitter grammar can parse typed method parameters
- The `processMethodDeclaration` function exists but uses regex patterns
- Need to replace with structured AST traversal

Requirements:
1. Rewrite `processMethodDeclaration` to use tree-sitter node traversal
2. Add `processMethodSignature` function to handle signature nodes
3. Add `processTypedMethodParameter` function for individual parameters
4. Create proper PerlTypeAnnotation objects for each parameter
5. Assign correct `MethodParamAnnotation` kind

Implementation:
1. Replace the existing `processMethodDeclaration` function in internal/parser/treesitter/perl.go:
   ```go
   func (t *PerlTree) processMethodDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
       var methodName string

       for i := 0; i < int(node.ChildCount()); i++ {
           child := node.Child(uint(i))
           if child == nil {
               continue
           }

           switch child.Kind() {
           case "bareword":
               if methodName == "" {
                   methodName = t.getNodeText(child)
               }
           case "signature":
               t.processMethodSignature(child, methodName, annotations)
           }
       }
   }
   ```

2. Implement the signature processing function:
   ```go
   func (t *PerlTree) processMethodSignature(signatureNode *sitter.Node, methodName string, annotations *[]*PerlTypeAnnotation) {
       for i := 0; i < int(signatureNode.ChildCount()); i++ {
           child := signatureNode.Child(uint(i))
           if child == nil {
               continue
           }

           if child.Kind() == "typed_method_parameter" {
               t.processTypedMethodParameter(child, methodName, annotations)
           }
       }
   }
   ```

3. Implement the typed parameter processing:
   ```go
   func (t *PerlTree) processTypedMethodParameter(paramNode *sitter.Node, methodName string, annotations *[]*PerlTypeAnnotation) {
       var paramName, typeName string

       for i := 0; i < int(paramNode.ChildCount()); i++ {
           child := paramNode.Child(uint(i))
           if child == nil {
               continue
           }

           switch child.Kind() {
           case "type_expression":
               typeName = t.extractTypeExpression(child)
           case "scalar", "array", "hash":
               paramName = t.getNodeText(child)
           }
       }

       if paramName != "" && typeName != "" {
           annotation := &PerlTypeAnnotation{
               ItemName: paramName,
               TypeName: typeName,
               Kind:     "method_parameter",
               StartPos: int(paramNode.StartByte()),
               EndPos:   int(paramNode.EndByte()),
               Content:  t.getNodeText(paramNode),
           }
           *annotations = append(*annotations, annotation)
       }
   }
   ```

Testing:
- Create unit tests for each function with minimal AST node examples
- Test with various parameter types: scalar, array, hash
- Test with complex type expressions: parameterized types, unions
- Verify that `MethodParamAnnotation` kind is assigned correctly in conversion

Deliverable: Working method signature processing that extracts typed parameter annotations.
```

### Step 5: Implement Method Return Type Processing ✅ COMPLETED

**Scope**: Add parser functions to extract method return type annotations
**Risk**: Low - Similar to parameter processing
**Test Strategy**: Unit tests for return type processing

```text
Implement parser functions to extract method return type annotations from the tree-sitter AST.

Context: Building on the signature processing, we need to handle method return types like `-> ReturnType` and convert them to MethodReturnAnnotation objects.

Current State:
- Method signature processing is implemented
- Grammar supports method_return_type nodes
- Need to add return type extraction logic

Requirements:
1. Add return type handling to `processMethodDeclaration`
2. Implement `processMethodReturnType` function
3. Create proper PerlTypeAnnotation objects for return types
4. Assign correct `MethodReturnAnnotation` kind
5. Handle method name association for return types

Implementation:
1. Update `processMethodDeclaration` to handle return types:
   ```go
   func (t *PerlTree) processMethodDeclaration(node *sitter.Node, annotations *[]*PerlTypeAnnotation) {
       var methodName string

       for i := 0; i < int(node.ChildCount()); i++ {
           child := node.Child(uint(i))
           if child == nil {
               continue
           }

           switch child.Kind() {
           case "bareword":
               if methodName == "" {
                   methodName = t.getNodeText(child)
               }
           case "signature":
               t.processMethodSignature(child, methodName, annotations)
           case "method_return_type":
               t.processMethodReturnType(child, methodName, annotations)
           }
       }
   }
   ```

2. Implement the return type processing function:
   ```go
   func (t *PerlTree) processMethodReturnType(returnTypeNode *sitter.Node, methodName string, annotations *[]*PerlTypeAnnotation) {
       for i := 0; i < int(returnTypeNode.ChildCount()); i++ {
           child := returnTypeNode.Child(uint(i))
           if child == nil {
               continue
           }

           if child.Kind() == "type_expression" {
               typeName := t.extractTypeExpression(child)

               if typeName != "" {
                   annotation := &PerlTypeAnnotation{
                       ItemName: methodName + "_return",  // Unique identifier for return type
                       TypeName: typeName,
                       Kind:     "method_return",
                       StartPos: int(returnTypeNode.StartByte()),
                       EndPos:   int(returnTypeNode.EndByte()),
                       Content:  t.getNodeText(returnTypeNode),
                   }
                   *annotations = append(*annotations, annotation)
               }
               break
           }
       }
   }
   ```

3. Update the annotation conversion in parser.go to handle the new kinds:
   ```go
   case "method_parameter":
       kind = MethodParamAnnotation
   case "method_return":
       kind = MethodReturnAnnotation
   ```

Testing:
- Create unit tests for return type processing with various type expressions
- Test simple types: `-> Str`, `-> Int`
- Test complex types: `-> HashRef[Str]`, `-> ArrayRef[Int]`
- Verify that `MethodReturnAnnotation` kind is assigned correctly
- Test methods with both parameters and return types

Deliverable: Working method return type processing that extracts return type annotations.
```

### Step 6: Update Annotation Conversion Pipeline ✅ COMPLETED

**Scope**: Ensure proper conversion from PerlTypeAnnotation to main parser TypeAnnotation
**Risk**: Low - Existing conversion patterns
**Test Strategy**: Integration tests with type conversion

```text
Update the annotation conversion pipeline to properly handle method parameter and return type annotations.

Context: The tree-sitter parser extracts PerlTypeAnnotation objects, but these need to be converted to the main parser's TypeAnnotation format with correct kinds and type expressions.

Current State:
- Method parameter and return type extraction is implemented
- The `convertPerlTypeAnnotation` function needs updates for new annotation kinds
- Type expression parsing should already work from previous implementations

Requirements:
1. Add handling for "method_parameter" and "method_return" kinds in conversion
2. Ensure type expression parsing works for method annotations
3. Verify correct AnnotationKind assignment
4. Test the complete pipeline from AST to TypeAnnotation

Implementation:
1. Update the conversion function in internal/parser/treesitter/parser.go:
   ```go
   switch perlAnn.Kind {
   case "variable":
       kind = VarAnnotation
   case "subroutine":
       kind = SubParamAnnotation
   case "method":
       kind = MethodParamAnnotation // Keep existing for backwards compatibility
   case "method_parameter":
       kind = MethodParamAnnotation
   case "method_return":
       kind = MethodReturnAnnotation
   case "type_declaration":
       kind = TypeDeclAnnotation
   default:
       kind = VarAnnotation
   }
   ```

2. Ensure the type expression parsing handles method-specific cases:
   - The existing `parseTypeExpression` function should work for method types
   - Test with parameterized types from method signatures
   - Verify complex type expressions are parsed correctly

3. Add debug output for method annotation conversion:
   ```go
   if os.Getenv("DEBUG_PARSER") == "1" && (perlAnn.Kind == "method_parameter" || perlAnn.Kind == "method_return") {
       fmt.Printf("DEBUG: Converting method annotation: %s = %s (kind: %s)\n",
                  perlAnn.ItemName, perlAnn.TypeName, perlAnn.Kind)
   }
   ```

Testing:
- Create integration tests that parse complete method examples
- Verify that PerlTypeAnnotation objects are created correctly
- Test conversion to main parser TypeAnnotation format
- Check that annotation kinds are assigned properly
- Test with the actual method examples from the failing test case

Deliverable: Complete annotation conversion pipeline that properly handles method type annotations.
```

### Step 7: Enable and Debug Test Case ✅ COMPLETED

**Scope**: Enable TestParseMethodTypeAnnotations and fix any remaining issues
**Risk**: Medium - Integration debugging
**Test Strategy**: Full test case execution and debugging

```text
Enable the TestParseMethodTypeAnnotations test case and debug any remaining parsing issues.

Context: With all the parsing infrastructure in place, we need to enable the skipped test and ensure it passes completely, finding all expected annotations.

Current State:
- All parser components are implemented
- The test is currently skipped with a message about incomplete implementation
- Need to enable the test and debug any issues

Requirements:
1. Remove the test skip and enable the full test
2. Parse the complete test method examples
3. Debug and fix any annotation extraction issues
4. Ensure all 10 expected annotations are found
5. Verify correct annotation kinds and type expressions

Implementation:
1. Remove the skip from TestParseMethodTypeAnnotations in internal/parser/parser_test.go:
   ```go
   func TestParseMethodTypeAnnotations(t *testing.T) {
       // Method type annotations are now implemented in tree-sitter parser

       // Create a parser
       p, err := NewParser()
       require.NoError(t, err)
       require.NotNil(t, p)
       // ... rest of test
   ```

2. Run the test with debug output enabled:
   ```bash
   DEBUG_PARSER=1 go test -v -run TestParseMethodTypeAnnotations ./internal/parser
   ```

3. Debug common issues:
   - Check that method_declaration nodes are being found
   - Verify signature and method_return_type nodes are parsed correctly
   - Ensure typed_method_parameter nodes appear in the AST
   - Debug type expression parsing for complex types

4. Fix any issues found:
   - Grammar precedence or parsing problems
   - AST traversal logic errors
   - Type expression extraction issues
   - Annotation kind assignment problems

Testing Strategy:
- Run with debug output to see what annotations are found
- Compare expected vs actual annotations
- Debug step by step through the parsing pipeline
- Test individual method examples in isolation if needed
- Verify that field annotations (already working) still pass

Deliverable: Enabled and passing TestParseMethodTypeAnnotations test case.
```

### Step 8: Integration Testing and Cleanup ✅ COMPLETED

**Scope**: Run full test suite, ensure no regressions, clean up code
**Risk**: Low - Final validation
**Test Strategy**: Complete test suite execution

```text
Perform comprehensive integration testing and code cleanup to ensure the implementation is robust and complete.

Context: With the main functionality working, we need to ensure no regressions were introduced and that the implementation follows project standards.

Current State:
- TestParseMethodTypeAnnotations should be passing
- All implementation components are in place
- Need to validate the complete system and clean up

Requirements:
1. Run the complete parser test suite to check for regressions
2. Run related type system tests to ensure integration
3. Clean up any debug code or temporary implementations
4. Ensure code follows project formatting and style standards
5. Update documentation if needed

Implementation:
1. Run the full parser test suite:
   ```bash
   go test -v ./internal/parser
   ```

2. Run related tests to check for regressions:
   ```bash
   go test -v ./internal/typedef
   go test -v ./internal/typechecker
   ```

3. Clean up implementation:
   - Remove or formalize any debug output
   - Ensure consistent error handling
   - Add any missing documentation comments
   - Verify code formatting with `go fmt`

4. Performance validation:
   - Ensure parsing performance hasn't degraded
   - Check that the new grammar doesn't significantly slow down parsing
   - Verify memory usage is reasonable

5. Integration checks:
   - Test method type annotations with the type checker
   - Verify annotations work with the rest of the type system
   - Check that PSC (Perl Static Checker) can use the new annotations

Testing Strategy:
- Run comprehensive test suite
- Test with various method signature patterns beyond the test case
- Verify backwards compatibility with existing code
- Check integration with other type system components

Deliverable: Complete, tested, and cleaned implementation of method type annotations parsing.
```

---

## Implementation Prompts Summary

The implementation has been broken down into 8 carefully sized steps:

1. **Add Typed Method Parameter Grammar Rules** - Grammar extension for `Type $param`
2. **Add Method Return Type Grammar Rules** - Grammar extension for `-> ReturnType`
3. **Rebuild and Validate Grammar** - Build process and validation
4. **Implement Method Signature Processing** - Parser logic for parameters
5. **Implement Method Return Type Processing** - Parser logic for return types
6. **Update Annotation Conversion Pipeline** - Integration with type system
7. **Enable and Debug Test Case** - Test enablement and debugging
8. **Integration Testing and Cleanup** - Final validation and cleanup

Each step builds incrementally on the previous ones, with strong testing at every stage. The steps are sized to be implementable safely while making meaningful progress toward the goal.

The approach prioritizes:
- **Incremental Progress**: Each step adds a specific capability
- **Early Testing**: Grammar and parser components are tested in isolation
- **No Big Jumps**: Complexity increases gradually across steps
- **Integration Focus**: Steps 6-8 ensure everything works together
- **Best Practices**: Follows existing code patterns and project standards

This blueprint provides a robust foundation for implementing method type annotations parsing with confidence and reliability.
