# Parser for PSC Type Annotations

This package implements a parser for Perl code with type annotations for the PSC (Perl Script Compiler) component of the PVM Ecosystem.

## Overview

The parser is designed to handle Perl code with type annotations in the following contexts:

1. **Variable Declarations**:
   - Scalar variables: `my Type $name`
   - Array variables: `my Type @array`
   - Hash variables: `my Type %hash`
   - With assignments: `my Type $var = value`

2. **Subroutine Declarations**:
   - Parameter types: `sub name(Type $param, AnotherType @array)`
   - Return types: `sub name() -> ReturnType`
   - Combined: `sub name(Type $param) -> ReturnType`

3. **Method Declarations**:
   - In regular packages: `sub method(Type $self, Type $param) -> ReturnType`
   - In class syntax: `method name(Type $param) -> ReturnType`

4. **Attribute Declarations**:
   - In class syntax: `field Type $attribute`
   - With default values: `field Type $attribute = default_value`

5. **Type Expressions**:
   - Simple types: `Int`, `Str`, `Bool`
   - Parameterized types: `ArrayRef[Type]`, `HashRef[KeyType, ValueType]`
   - Union types: `Type1|Type2`
   - Intersection types: `Type1&Type2`
   - Negation types: `!Type`

## Implementation Notes

The current implementation is a simplified version that demonstrates the structure and expected functionality.
For a production implementation, a more robust parser would be needed, likely using a parsing library like tree-sitter.

The implementation includes:

- A simple recursive descent parser for type expressions
- A placeholder implementation for parsing type annotations in Perl code
- Integration with the PSC component for type checking

## Future Enhancements

Future enhancements to the parser would include:

1. Integration with tree-sitter for more robust parsing
2. Full grammar extensions for Perl with type annotations
3. Support for more complex type expressions
4. Better error reporting and recovery
5. Performance optimizations for parsing large files

## Usage

```go
// Create a parser
parser, err := parser.NewParser()
if err != nil {
    // Handle error
}

// Parse a file
ast, err := parser.ParseFile("/path/to/file.pl")
if err != nil {
    // Handle error
}

// Process type annotations
for _, annotation := range ast.TypeAnnotations {
    // Process each type annotation
    fmt.Printf("Found type annotation: %s has type %s\n",
        annotation.AnnotatedItem, annotation.TypeExpression.String())
}
```

## Type Checking

The parser provides integration with PSC for type checking:

```go
// Create a type checker
checker, err := parser.NewTypeChecker()
if err != nil {
    // Handle error
}

// Check a file
result, err := checker.CheckFile("/path/to/file.pl")
if err != nil {
    // Handle error
}

// Process results
if len(result.Errors) > 0 {
    // Handle type errors
} else {
    // No type errors found
}
```

## Stripping Type Annotations

To strip type annotations from Perl code:

```go
// Strip type annotations
result, err := parser.StripAnnotations("/path/to/file.pl")
if err != nil {
    // Handle error
}

// Write the result to a file
err = os.WriteFile("/path/to/stripped.pl", []byte(result), 0644)
if err != nil {
    // Handle error
}
```
