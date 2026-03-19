# Cross-File Analysis

## Problem

PSC analyzes each file in isolation. When code calls `Foo::Bar::baz()` or
`$obj->method()`, the return type is Unknown because PSC never reads
`Foo/Bar.pm`. Import statements (`use Foo::Bar`) are ignored. Method calls
on objects always return Any. This limits the value of type inference in any
project with more than one file.

## Approach

Three increments, each delivering standalone value:

1. **File resolution + per-file cache** — find .pm files, analyze on demand, cache
2. **Import tracking + FQ call resolution** — resolve `use` statements and
   fully-qualified function calls across file boundaries
3. **Class tracking + method resolution** — track class names on objects,
   resolve `$obj->method()` to the class's .pm file

## Increment 1: File Resolution and Per-File Cache

### ProjectIndex

A new type in `internal/infer/` that manages cross-file analysis:

```go
type ProjectIndex struct {
    root    string                        // project root (where .git or pvm.toml lives)
    libDirs []string                      // search paths, default ["lib"]
    cache   map[string]*FileAnalysis      // keyed by absolute file path
    mu      sync.RWMutex                  // concurrent access from LSP
}

type FileAnalysis struct {
    Annotations map[uint32]types.Type
    Diagnostics []Diagnostic
    Symbols     *SymbolTable
    Package     string                    // primary package declared in the file
}
```

### File Resolution

`Foo::Bar` resolves to `lib/Foo/Bar.pm` relative to the project root. The
resolution function:

1. Replace `::` with `/` in the module name
2. Append `.pm`
3. Search each directory in `libDirs` for the file
4. Return the first match, or an error if not found

### Lazy Analysis with Background Prefetch

On first reference (via `use` statement or FQ call):
1. Resolve the module name to a file path
2. If already in cache, return cached result
3. Otherwise: parse the file, run CollectDeclarations + Analyze, cache result
4. The ProjectIndex is passed to Analyze so dependencies can cascade

Background prefetch after startup:
1. Walk `lib/` recursively, find all `.pm` files
2. For each file not already in cache, analyze in a background goroutine
3. Use a worker pool to limit concurrency
4. The LSP server triggers prefetch after initialization

### Cache Invalidation

For `psc check` (batch mode): no invalidation needed — single run.

For LSP: when a file changes (didChange/didSave), invalidate its cache entry.
Files that depend on the changed file are NOT automatically invalidated
(conservative — the user re-opens or re-saves dependent files). Full
invalidation can be added later if needed.

### CST Shapes (verified by discovery)

```
// use Foo::Bar;
use_statement
  use (anon)
  package "Foo::Bar"              // ← module name
  ;

// require Foo::Bar;
expression_statement
  require_expression
    require (anon)
    bareword "Foo::Bar"           // ← module name

// package Foo::Bar;
package_statement
  package (anon)
  package "Foo::Bar"              // ← declares the package for this file

// use Foo::Bar qw(...);
// Broken CST — string grammar bug. Module name still extractable
// from the package child, but import list is not parseable.
```

### What Changes

**New file: `internal/infer/project.go`**
- `ProjectIndex` struct with `NewProjectIndex(root string)`
- `ResolveModule(name string) (string, error)`
- `AnalyzeFile(path string) (*FileAnalysis, error)` — lazy, cached
- `LookupSymbol(pkg, name string) (Symbol, bool)` — cross-file lookup
- `Prefetch()` — background analysis of lib/

**infer.go:** `Analyze` gains an optional `*ProjectIndex` parameter (nil for
single-file mode, preserving backward compatibility).

**psc/lsp.go:** LSPServer holds a `*ProjectIndex`. Initialized on workspace
open. Prefetch triggered after initialization.

**psc/check_command.go:** Creates a ProjectIndex for the project root.

## Increment 2: Import Tracking and FQ Call Resolution

### use Statement Processing

When the inference walker encounters a `use_statement`:
1. Extract the package name from the `package` child
2. Call `projectIndex.AnalyzeFile` for the resolved .pm file
3. The analyzed file's exported symbols become available for lookup

### Fully-Qualified Function Calls

`Foo::Bar::baz()` already parses as `function_call_expression` with
`function` child "Foo::Bar::baz". The existing `inferFunctionCallType`
splits the name on `::`, resolves the package via ProjectIndex, and looks up
`baz`'s ReturnType from the cached symbol table.

The lookup order becomes:
1. Check builtins (authoritative)
2. Check local symbol table (same-file subs)
3. Check ProjectIndex for cross-file symbols (split FQ name on `::`)

### What Changes

**infer.go:** `walkNode` handles `use_statement` by triggering ProjectIndex
analysis. `inferFunctionCallType` gains FQ name resolution via ProjectIndex.

**project.go:** `LookupSymbol` resolves a package + function name to a Symbol
from the cached analysis of the package's .pm file.

## Increment 3: Class Tracking and Method Resolution

### ClassType Field

Add `ClassType string` to the `Symbol` struct (alongside the existing
`ReturnType types.Type` and bitset `Type`). ClassType holds the Perl class
name when known.

### Constructor Pattern: Foo->new()

CST (verified):
```
method_call_expression
  bareword "Foo"                  // class name
  -> (anon)
  method "new"                    // constructor
  ( )
```

When `$obj = Foo->new()` is assigned, set `ClassType = "Foo"` on $obj's
Symbol. The assignment narrowing already handles the RHS type — extend it to
also propagate ClassType from the call result.

`inferMethodCallType` is a new function that handles `method_call_expression`:
1. If the invocant is a `bareword`, it's a class method call. The class name
   is the bareword text. Look up the class via ProjectIndex, find the method,
   return its ReturnType. For `new()`, return Object with ClassType set.
2. If the invocant is a `scalar`, look up the variable's ClassType from the
   symbol table. If set, resolve the class via ProjectIndex and look up the
   method's ReturnType.

### Guard Narrowing with Class Name

`if ($x isa Foo)` already narrows $x to Object via GuardIsa. Extend the
guard result to carry the class name from the `bareword` child of the
relational_expression. When walkBlockWithGuard applies the guard, set
ClassType on the narrowed symbol.

### What Changes

**symbols.go:** Add `ClassType string` to Symbol.

**infer.go:** Add `inferMethodCallType`. Update `inferNodeType` to dispatch
`method_call_expression` to it. Update assignment narrowing to propagate
ClassType. Update `extractIsaGuard` to capture the class name.

**project.go:** `LookupSymbol` also used for method resolution — looks up
a method name in a class's symbol table.

## Out of Scope

- Configurable search paths beyond lib/ (psc.toml support deferred)
- `use lib` directive parsing
- Import lists from `use Foo qw(bar baz)` (blocked by string grammar bug)
- Inheritance / `@ISA` / `use parent` resolution
- CPAN module analysis (only project-local .pm files)
- Automatic cache invalidation for dependents (manual re-save)
- `require` with runtime expressions
- Exporter.pm protocol (what symbols are actually exported)

## Implementation Order

### Increment 1 (File Resolution + Cache)
1. ProjectIndex struct with ResolveModule
2. AnalyzeFile with caching
3. LookupSymbol for cross-file queries
4. Background Prefetch
5. Wire into LSP and psc check

### Increment 2 (Import Tracking + FQ Calls)
1. Walk use_statement nodes, trigger AnalyzeFile
2. FQ name splitting in inferFunctionCallType
3. Cross-file ReturnType lookup via ProjectIndex
4. End-to-end tests: sub in lib/Foo.pm, call in main script

### Increment 3 (Class Tracking + Method Resolution)
1. ClassType field on Symbol
2. Constructor pattern recognition (Foo->new())
3. inferMethodCallType with ProjectIndex lookup
4. Guard narrowing with class name propagation
5. End-to-end tests: class in lib/, method call in script
