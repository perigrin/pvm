# PSC Type Inference Design

## Problem

Perl values carry latent semantic types — `42` is an Int, `"hello"` is a Str,
`push` returns Int — but no tool exposes this information from standard Perl
source. PSC already parses Perl via tree-sitter into a concrete syntax tree.
This design adds a type inference engine that analyzes that CST to surface
Perl's implicit types for both interactive (LSP) and batch (`psc check`) use.

## Philosophical Foundation

We do not add types to Perl. We expose the types values already have.

Every Perl value belongs to types determined by two tests (from
`docs/perl-types-practical.md` in the Chalk project):

1. **Syntactic Preservation** — can the value survive conversion to the type
   and back without losing information?
2. **Semantic Fulfillment** — does the value behave correctly with the type's
   operations?

This is a flow-based type system similar to TypeScript's control flow analysis,
but purely inference-driven. No annotations exist to seed from. Every type fact
originates from sigils, operators, literals, builtins, and control flow guards.

## Type Hierarchy

The type lattice, identical to Chalk's TypeLibrary:

```
Any
├── Scalar
│   ├── Undef
│   ├── Bool
│   ├── Str
│   │   └── Num
│   │       └── Int
│   ├── DualVar
│   ├── Regex
│   └── Ref
│       ├── ScalarRef
│       ├── ArrayRef
│       ├── HashRef
│       ├── CodeRef
│       ├── GlobRef
│       └── Object
├── List
│   ├── Array
│   └── Hash
├── Code
├── Glob
└── None
```

Critical subtyping chain: **Int <: Num <: Str <: Scalar <: Any**

Polymorphic types (Scalar, Any, List) pass permissively against subtypes.
A variable typed as Scalar could hold Str, Int, or Num at runtime.

## Architecture: Two-Pass Inference

The engine receives a completed tree-sitter CST and walks it twice.

### Pass 1: Declaration Collection

Walk the CST top-down. Build a symbol table keyed by scope + name:

- **Subroutine/method declarations** — name, parameter list, enclosing
  package/class, location
- **Variable declarations** — `my`, `our`, `state`, `field` with sigil-derived
  type and lexical scope
- **Package/class declarations** — namespace boundaries for method resolution
- **Use/require statements** — module dependencies (reuses existing
  `collectDependencies` logic)

Tree-sitter provides block structure directly. Every `{...}` defines a scope
boundary. Inner declarations shadow outer ones.

### Pass 2: Type Inference

Walk the CST bottom-up (post-order). At each node, compute its type from its
children's already-computed types. Store results in `map[uint32]TypeInfo` keyed
by node start byte. Emit diagnostics when type violations are detected.

Diagnostics are values, not side effects. `Analyze(tree, source)` returns both
the annotation map and a slice of diagnostics. The caller decides presentation.

## Type Sources

Five sources feed the inference:

1. **Sigils** — the only "declarations" Perl provides. `$` → Scalar,
   `@` → Array, `%` → Hash, `$#` → Int, `*` → Glob.
2. **Operators** — return type comes from the signature table. `+` → Num,
   `.` → Str, `==` → Bool, etc.
3. **Literals** — `42` → Int, `3.14` → Num, `"hello"` → Str,
   `/pattern/` → Regex, `undef` → Undef, `true`/`false` → Bool.
4. **Builtins** — `push` → Int, `keys` → List, `split` → List, etc.
   Arity and argument types validated against the signature table.
5. **Control flow** — basic narrowing through guards: `defined($x)` narrows
   from Scalar to non-Undef, `ref($x)` to Ref, `$x isa Foo` to Object.

## Node Kind Mapping

Tree-sitter's Perl grammar uses these node types. The inference walk dispatches
on `node.Kind()`.

### Literals

| tree-sitter kind              | Type       | Notes                                  |
|-------------------------------|------------|----------------------------------------|
| `number`                      | Int or Num | `.` or `e/E` in text → Num, else Int   |
| `string_literal`              | Str        | Single-quoted                          |
| `interpolated_string_literal` | Str        | Double-quoted                          |
| `quoted_regexp`               | Regex      |                                        |
| `quoted_word_list`            | List       | qw(...)                                |
| `version`                     | Str        |                                        |

Keyword literals detected by text: `undef` → Undef, `true`/`false` → Bool.

### Variables

| tree-sitter kind | Type   |
|-------------------|--------|
| `scalar`          | Scalar |
| `array`           | Array  |
| `hash`            | Hash   |
| `arraylen`        | Int    |
| `glob`            | Glob   |

### Operators

Binary expressions: find operator child node, look up text in `BinaryOpSig`
map, assign result type. Unary: same pattern with `UnaryOpSig`.

Operator signatures ported directly from Chalk's tables:
- Arithmetic (`+`, `-`, `*`, `/`, `%`, `**`) → Num
- String (`.`) → Str, (`x`) → Str
- Numeric comparison (`==`, `!=`, `<`, `>`, `<=`, `>=`, `<=>`) → Bool
- String comparison (`eq`, `ne`, `lt`, `gt`, `le`, `ge`, `cmp`) → Bool
- Logical (`&&`, `||`, `//`, `and`, `or`) → Any
- Bitwise (`&`, `|`, `^`, `<<`, `>>`) → Int
- Regex binding (`=~`, `!~`) → Bool
- Range (`..`, `...`) → List

### Function Calls

| tree-sitter kind              | Action                                                |
|-------------------------------|-------------------------------------------------------|
| `function_call_expression`    | Extract name, look up BuiltinSig, validate, return    |
| `func0op_call_expression`     | Zero-arg builtin                                      |
| `func1op_call_expression`     | One-arg builtin                                       |

Wrong arity or argument type mismatch → emit diagnostic.

### Method Calls

| tree-sitter kind            | Action                                               |
|-----------------------------|------------------------------------------------------|
| `method_call_expression`    | Extract method name, consult symbol table for return  |

Intra-file only. Unknown methods get type Any (permissive, no diagnostic).

### Structural Constructs

| tree-sitter kind          | Type   | Notes                    |
|---------------------------|--------|--------------------------|
| `conditional_expression`  | Any    | Ternary branches differ  |
| anonymous sub             | Code   | `sub { ... }`            |
| subscript `[...]`         | Scalar | Array element access     |
| subscript `{...}`         | Scalar | Hash element access      |
| `postfix_deref` `->@*`   | Array  |                          |
| `postfix_deref` `->%*`   | Hash   |                          |
| `postfix_deref` `->$*`   | Scalar |                          |

### Context Narrowing

Assignment triggers context analysis:
- LHS is `$x` (Scalar) → scalar context on RHS
- LHS is `@x` (Array) → list context on RHS
- Statement position → void context

In scalar context: Array/Hash → Int (count), List → Scalar.
In void context: type discarded.

### Basic Flow Narrowing

Inside `if`/`unless`/`while` blocks, when the condition matches a recognized
pattern:

| Condition             | Narrowing                    |
|-----------------------|------------------------------|
| `defined($x)`        | Scalar → non-Undef           |
| `ref($x)`            | Scalar → Ref                 |
| `ref($x) eq 'HASH'`  | Scalar → HashRef             |
| `$x isa Foo`         | Scalar → Object              |

Implemented as scoped type overrides in the annotation map. The override
expires when leaving the block.

## Package Layout

```
internal/
  types/                  # Type system (no parser dependency)
    types.go              # Type enum, hierarchy DAG, IsSubtype, TypeSatisfies
    signatures.go         # Builtin function + operator signatures
    narrowing.go          # Context narrowing + basic flow narrowing rules

  infer/                  # Inference engine (depends on types/ and parser/)
    symbols.go            # Pass 1: symbol table, scopes, declarations
    infer.go              # Pass 2: bottom-up type inference walk
    diagnostics.go        # Diagnostic types and codes

  psc/                    # CLI commands (existing package, extended)
    check_command.go      # NEW: "psc check" batch type checking
    parse_command.go      # existing
    analyze_command.go    # existing
    lsp_command.go        # existing
    lsp.go                # existing, extended to use infer/
```

`types/` contains pure data — the type hierarchy, signatures, and
satisfiability rules. No dependency on tree-sitter or the parser.

`infer/` walks tree-sitter CSTs using `types/`. Stateless per invocation:
`Analyze(tree, source) → (map[uint32]TypeInfo, []Diagnostic)`. Safe for
concurrent use by the LSP server.

## Go Type Representations

### Type Enum

```go
type Type int

const (
    Any Type = iota
    Scalar
    Undef
    Bool
    Str
    Num
    Int
    DualVar
    Regex
    Ref
    ScalarRef
    ArrayRef
    HashRef
    CodeRef
    GlobRef
    Object
    List
    Array
    Hash
    Code
    Glob
    None
)
```

### TypeInfo (annotation map value)

```go
type TypeInfo struct {
    Type    Type
    EvalCtx Context // Scalar, List, Void
}
```

Stores conclusions only. Chalk's focus hashes carry intermediate working state
(`op_text`, `call_symbol`, `list_arity`) because the semiring threads partial
information through an in-progress parse. We have the full tree, so we read
child nodes directly.

### Signatures

```go
type BuiltinSig struct {
    MinArity   int
    ArgTypes   []Type // last entry is variadic
    ReturnType Type
}

type BinaryOpSig struct {
    Left, Right, Result Type
}

type UnaryOpSig struct {
    Operand, Result Type
}
```

### Diagnostics

```go
type Diagnostic struct {
    StartByte uint32
    EndByte   uint32
    Severity  Severity // Error, Warning, Info
    Message   string
    Code      string   // machine-readable: "arity-mismatch", "type-mismatch"
}
```

## Consumer Integration

### `psc check <file|directory>`

1. Parse file(s) with tree-sitter
2. Run Pass 1 + Pass 2
3. Print diagnostics: `file.pl:12:5: warning: push expects Array as first argument, got Scalar [arity-mismatch]`
4. Exit 0 if clean, 1 if diagnostics found

### LSP Server

**didOpen / didChange:**
Parse → infer → publish diagnostics via `textDocument/publishDiagnostics`.

**hover:**
Find node at cursor → look up annotation map → return type as hover text.

**completion:**
After `->` → suggest methods from symbol table. In call position → suggest
builtins with signatures.

**definition:**
Find symbol at cursor → look up symbol table → return declaration location.

The annotation map and symbol table are cached per-file, invalidated on edit.
Tree-sitter's incremental parsing keeps re-parse fast. The inference engine is
stateless, so concurrent LSP requests are safe.

## Relationship to Chalk

This design ports Chalk's type knowledge but not its inference machinery:

| Concept                     | Chalk (Earley semiring)           | PSC (tree-sitter post-parse)     |
|-----------------------------|-----------------------------------|----------------------------------|
| Type hierarchy              | `TypeLibrary.pm` %PARENT          | `types/types.go` parent map      |
| Builtin signatures          | `TypeLibrary.pm` %BUILTIN_SIGS    | `types/signatures.go`            |
| Operator signatures         | `TypeLibrary.pm` %BINARY_OP_SIGS  | `types/signatures.go`            |
| Subtype checking            | `is_subtype`, `type_satisfies`    | `IsSubtype`, `TypeSatisfies`     |
| Context narrowing           | `narrow_type`                     | `types/narrowing.go`             |
| Inference integration       | Semiring on_scan/on_complete      | Two-pass tree walk               |
| Value threading             | Comonad Context + hash-consing    | `map[uint32]TypeInfo`            |
| Disambiguation              | Rejects ill-typed parses          | Not applicable (parse complete)  |

The type knowledge is shared. The integration differs because tree-sitter
delivers a complete CST, eliminating the need for semiring algebra.

## Future Work

### Full Flow Narrowing (separate PRD)

The basic narrowing covers `defined`, `ref`, `isa` guards. A full
implementation would add:

- **Branch merging at join points** — compute the union type where if/elsif/else
  branches rejoin
- **Negated guards** — `unless (defined $x)` narrows to Undef inside the block
- **Early exit narrowing** — after `return` or `die` in a branch, the
  remaining code knows the guard failed
- **Elsif chains** — progressive narrowing through multiple conditions
- **Loop variable narrowing** — `for my $item (@array)` means `$item` is Scalar
- **Nested conditions** — `if (ref($x) && ref($x) eq 'HASH')` combines guards

### Type Guards (separate PRD)

- **User-defined type guard functions** — like TypeScript's `x is Foo` return
  type annotation, but inferred from function body patterns
- **Type guard suggestions** — when a diagnostic fires ("got Scalar, expected
  Array"), suggest adding a `ref` guard
- **Guard pattern library** — recognize common Perl guard idioms
  (`looks_like_number`, `blessed`, `reftype`)

### Cross-File Analysis

- Import type information from dependencies via `use` statements
- Build a project-wide symbol table across all `.pm` files
- Method resolution across packages

### Method Return Type Inference

- Analyze subroutine/method bodies to infer return types
- Track through `return` statements (like Chalk's MethodDefinition action)
- Propagate to call sites
