# Tree-sitter-perl Parser for PSC

This directory contains the tree-sitter-perl parser used by PSC (Perl Script Compiler) for parsing and type checking Perl code. PSC exclusively uses tree-sitter for parsing.

## Overview

PSC uses tree-sitter-perl to parse Perl code with type annotations. The tree-sitter parser provides a robust parsing capability that works even with incomplete or syntactically incorrect code, making it ideal for use in IDE-like environments.

The integration consists of:

1. Go bindings for the tree-sitter-perl library
2. Grammar extensions to handle typed Perl syntax
3. AST conversion to map tree-sitter nodes to our internal AST format
4. Type annotation extraction from the parsed AST

## Building the tree-sitter-perl Grammar with Type Extensions

To build the grammar with type extensions:

1. Clone the tree-sitter-perl repository:
   ```bash
   git clone https://github.com/tree-sitter-perl/tree-sitter-perl
   cd tree-sitter-perl
   ```

2. Copy the grammar extension file:
   ```bash
   cp /path/to/pvm/internal/parser/treesitter/grammar_extension.js .
   ```

3. Modify the grammar.js file to include our extensions:
   ```bash
   # Add the following line to the end of grammar.js:
   module.exports.rules = Object.assign(module.exports.rules, require('./grammar_extension').rules);
   ```

4. Generate the parser:
   ```bash
   npx tree-sitter generate
   ```

5. Build the shared library:
   ```bash
   npx tree-sitter build-wasm
   # Or for a native shared library:
   # cc -dynamiclib -fPIC -g -O2 -I./src src/*.c -shared -o libtree-sitter-perl.so
   ```

6. Move the shared library to a location where Go can find it:
   ```bash
   mkdir -p /Users/perigrin/dev/pvm/lib
   cp libtree-sitter-perl.so /Users/perigrin/dev/pvm/lib/
   ```

## Using the Tree-sitter Parser in Go

The tree-sitter-perl parser is accessed through the Go bindings in this directory:

```go
import (
    "tamarou.com/pvm/internal/parser/treesitter"
)

// Create a new parser
parser, err := treesitter.NewParser(false)
if err != nil {
    log.Fatal(err)
}
defer parser.Close()

// Parse a file
ast, err := parser.ParseFile("path/to/file.pl")
if err != nil {
    log.Fatal(err)
}

// Access type annotations
for _, annotation := range ast.TypeAnnotations {
    fmt.Printf("%s has type %s\n", annotation.AnnotatedItem, annotation.TypeExpression.String())
}
```

## Grammar Rules for Type Annotations

The grammar extension adds support for the following syntax:

1. **Variable Declarations**:
   ```perl
   my Int $count = 0;
   my Str $name = "example";
   my ArrayRef[Int] $numbers = [1, 2, 3];
   ```

2. **Subroutine Declarations**:
   ```perl
   sub add(Int $a, Int $b) -> Int {
       return $a + $b;
   }
   ```

3. **Method Declarations**:
   ```perl
   method greet(Str $name) -> Str {
       return "Hello, $name!";
   }
   ```

4. **Field Declarations**:
   ```perl
   field Bool $flag;
   field Str $name = "default";
   ```

5. **Type Declarations**:
   ```perl
   type ID = Int;
   type Names = ArrayRef[Str];
   type Callback = CodeRef[Str, Int];
   ```

6. **Type Expressions**:
   - Simple types: `Int`, `Str`, `Bool`
   - Parameterized types: `ArrayRef[Int]`, `HashRef[Str, Int]`
   - Union types: `Int|Str`, `Maybe[Int]|ArrayRef[Int]`
   - Intersection types: `Serializable&Printable`
   - Negation types: `!Int`

7. **Type Assertions**:
   ```perl
   my $value = $input as Int;
   ```

## Future Improvements

1. **Integration with CPAN Type Libraries**: Automatically extract type definitions from CPAN modules.
2. **Performance Optimizations**: Add caching for repeated parsing of the same files.
3. **Incremental Parsing**: Support for incremental parsing as files are edited.
4. **Syntax Error Recovery**: Improve error recovery for syntax errors in type expressions.
5. **IDE Integration**: Provide an API for IDE plugins to use the parser for features like syntax highlighting and code completion.

## References

- [Tree-sitter Perl](https://github.com/tree-sitter-perl/tree-sitter-perl)
- [Go Tree-sitter Bindings](https://github.com/tree-sitter/go-tree-sitter)
- [Tree-sitter Documentation](https://tree-sitter.github.io/tree-sitter/)
