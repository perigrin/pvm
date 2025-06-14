# Experimental Features

This directory contains experimental features that are not yet ready for production use. These features represent future functionality that may be integrated into the main codebase once they are stable and the underlying systems they depend on are fully implemented.

## Language Server Features

The following Language Server Protocol (LSP) features have been moved here as they require a complete type system implementation:

### Enhanced Completion (`enhanced_completion.go`)
- Type-aware autocompletion
- Intelligent code suggestions based on context
- Function signatures and parameter information
- Requires: Complete type inference and symbol resolution

### Enhanced Diagnostics (`enhanced_diagnostics.go`)
- Rich error messages with type context
- Quick fixes and suggestions for type mismatches
- Type hierarchy analysis
- Requires: Full type checking pipeline

### Enhanced Navigation (`enhanced_navigation.go`)
- Advanced "Go to Definition" with complete type information
- Cross-module symbol navigation
- Type-aware symbol information
- Requires: Multi-file type analysis

### Incremental Processing (`incremental.go`)
- Incremental parsing and analysis for performance
- Requires: Stable incremental type checking

## Future Integration

These features will be moved back to their appropriate packages once:

1. The tree-sitter-typed-perl grammar supports all required syntax
2. The type system implementation is complete and stable
3. Symbol binding and resolution is fully implemented
4. Performance characteristics are acceptable

## Development Notes

- These files may not compile without modifications
- Import paths may need updates when features are moved back
- TODO comments link to roadmap items for implementation
