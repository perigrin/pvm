# Tree-sitter Debug Utility

## Purpose
Use `debug_treesitter_dump.go` to analyze tree-sitter parse output when debugging parser conversion issues.

## Usage
```bash
go run debug_treesitter_dump.go
```

## Output Analysis

### Expected Structure for Subroutines
For the code:
```perl
sub hello {
    print "Hello\n";
}
hello();
```

Tree-sitter correctly produces:
```
source_file
├── subroutine_declaration_statement (sub hello { ... })
│   ├── sub
│   ├── bareword ("hello")
│   └── block ({ ... })
├── expression_statement (hello())
│   └── ambiguous_function_call_expression
│       ├── function ("hello")
│       └── stub_expression ("()")
└── ; (semicolon)
```

### Key Findings
1. **Tree-sitter works correctly** - parses 3 main children as expected
2. **Subroutine found** - `subroutine_declaration_statement` contains the sub definition
3. **Function call found** - `ambiguous_function_call_expression` contains the call

### Problem Location
The issue is in `internal/parser/treesitter/parser.go` in the `convertTreeSitterNode` function, which should:
1. Convert `subroutine_declaration_statement` to proper AST nodes
2. Extract subroutine name from `bareword` child
3. Preserve all node text and hierarchy

But instead is producing:
- `expression_stmt` (wrong type)
- Empty text content
- Only 1 child instead of 3

## Next Steps
Fix the conversion logic in `convertTreeSitterNode` to properly map tree-sitter node types to our AST format.
