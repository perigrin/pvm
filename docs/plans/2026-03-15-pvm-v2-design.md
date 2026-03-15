# PVM Rewrite Design: Copy-Forward to gotreesitter

## Decision

Rewrite PVM on a clean branch, copying forward proven core code while
rebuilding PSC without type annotations. Type annotations move to the Chalk
project. PSC becomes a structural Perl analysis tool using gotreesitter's
pure-Go parser.

## Motivation

PVM's codebase accumulated significant complexity around typed Perl:
tree-sitter-typed-perl (custom C grammar with CGO bindings), a full type
checker, binder, compiler pipeline, and type-specific PSC commands. This
infrastructure is better suited to Chalk, which does type inference natively.

Meanwhile, gotreesitter (github.com/odvcencio/gotreesitter) provides a pure-Go
tree-sitter runtime with a bundled Perl grammar and external scanner. This
eliminates CGO, cross-compilation issues, Node.js build dependencies, and the
custom grammar maintenance burden.

## Architecture

Four components in a single binary, same as the original PRD:

- **PVM** -- Perl version management (install, use, shims, shell integration)
- **PVX** -- Isolated Perl script execution
- **PM** -- CPAN module management
- **PSC** -- Perl structural analysis (parse, LSP, project analysis)

### PSC v2 Commands

- `psc parse <file>` -- Parse Perl and output AST (tree, JSON, S-expression)
- `psc lsp` -- Language Server Protocol server (structural, no types)
- `psc analyze <file|dir>` -- Dependency graphing, module structure, project analysis

Type inference may return later, layered on top of the parser, using Perl's
implicit type system (scalars, arrays, hashes, references, blessed objects)
rather than explicit annotations.

### gotreesitter Integration

The parser package wraps gotreesitter behind a clean interface:

```go
type Parser struct { /* gotreesitter internals */ }
type Tree struct { /* gotreesitter.Tree */ }
type Node struct { /* gotreesitter.Node */ }

func NewParser() *Parser
func (p *Parser) Parse(source []byte) (*Tree, error)
func (p *Parser) ParseIncremental(source []byte, old *Tree, edits []Edit) (*Tree, error)
```

Uses `grammars.PerlLanguage()` directly -- standard Perl grammar with external
scanner, no custom grammar, no CGO. Incremental parsing supported natively.

## What Gets Copied Forward

### Copy as-is

- `internal/config/` -- XDG paths, pvm.toml, defaults
- `internal/version/` -- version checking, comparison, resolution
- `internal/updater/` -- self-update, auto-update
- `internal/perl/` -- Perl build, detection, patchperl, shim generation
- `internal/ui/` -- terminal UI components
- `internal/memory/` -- lazy loading, caching
- `internal/errors/` -- error formatting

### Copy with adaptation

- `internal/pvm/` -- strip PSC/type references
- `internal/pvx/` -- remove --type-check integration
- `internal/pm/` -- remove pm type and type definition commands
- `internal/cli/` -- update command routing for new PSC
- `cmd/` -- entry points, wire four components
- `test/e2e/` -- remove type-checking e2e tests

### Not copied (deleted)

- `tree-sitter-typed-perl/` -- entire directory
- `internal/typechecker/` -- type system
- `internal/binder/` -- symbol binding for type checking
- `internal/compiler/` -- typed/clean Perl compilation
- `internal/ast/` -- type-annotation-aware AST
- `internal/parser/` -- rebuilt with gotreesitter
- `internal/inference/` -- type inference
- `internal/migration/` -- type migration layer
- `internal/validation/` -- type validation
- `internal/mcp/` -- MCP server
- `internal/ls/` -- current LSP (rebuilt in new PSC)
- `internal/psc/` -- rebuilt from scratch

## Build System

The CGO build chain is eliminated entirely:

- `go build ./cmd/pvm` -- builds everything
- `go test ./...` -- standard Go testing
- `GOOS=linux GOARCH=arm64 go build` -- cross-compilation works
- No Node.js, no tree-sitter CLI, no C compiler required
- Makefile retained for convenience targets

## Migration Sequence

Each step compiles and passes tests before proceeding.

1. **Bootstrap** -- orphan branch `pure-go`, go.mod, Makefile, CLAUDE.md
2. **Core infrastructure** -- config, version, errors, memory, ui
3. **Perl toolchain** -- perl, updater
4. **PVM core** -- pvm, cli, cmd/pvm entry point
5. **PVX and PM** -- pvx, pm, their entry points
6. **New parser** -- gotreesitter wrapper with tests
7. **New PSC** -- parse, analyze, lsp commands
8. **E2E and polish** -- adapted e2e tests, cross-compilation, consistency checks

## Open Questions

- Binary size impact of embedding all 206 grammar blobs (~14MB). Acceptable
  for now; gotreesitter has build tags to control this if needed.
- Scope of `psc analyze` -- start with dependency graphing, grow into linting
  over time.
- When and how to reintroduce type inference without annotations.
