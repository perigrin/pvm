# Method Return Type Inference

## Problem

PSC types every function call as `Any` because it never analyzes subroutine
bodies. `my $x = foo()` leaves `$x` at its sigil type `Scalar`, even when
`sub foo { return 42; }` clearly returns `Int`. The inference engine already
walks inside sub bodies for variable scoping (Pass 1) and narrowing (Pass 2),
but discards the return type information.

## Approach

During Pass 2, when the inference walker encounters a
`subroutine_declaration_statement`, walk the body, collect the types of all
return paths (explicit and implicit), OR them into a union, and store the
result as `ReturnType` on the subroutine's Symbol. At call sites, look up the
callee's `ReturnType` and use it instead of `Any`.

### Return Path Analysis

Three sources of return types:

**Explicit returns:** Every `return_expression` node contributes the type of
its value child. A bare `return;` (no value) contributes `Undef`.

**Implicit return:** Perl subs return the value of the last evaluated
expression. The last `expression_statement` in the body block contributes its
type. If the last statement is a `conditional_statement`, recurse into each
branch's last expression (union all paths).

**No return paths:** If the sub body is empty or contains only side-effect
statements with no return, the return type is `Unknown` (no information —
callers get permissive behavior, same as today).

The final return type is the bitwise OR of all path types.

### CST Shapes (verified by discovery)

```
// sub foo { return 42; }
subroutine_declaration_statement
  sub (anon)
  bareword "foo"
  block
    expression_statement
      return_expression
        return (anon)
        number "42"              // ← explicit return, type Int

// sub bar { 42; }
subroutine_declaration_statement
  sub (anon)
  bareword "bar"
  block
    expression_statement
      number "42"                // ← implicit return (last expr), type Int

// sub qux { if ($x) { return 1; } return 3.14; }
// Two explicit returns: Int and Num → union = Num (Num mask includes Int)

// Call site: my $y = foo();
assignment_expression
  variable_declaration "my $y"
  = (anon)
  function_call_expression
    function "foo"               // ← look up foo's ReturnType
    ( )
```

### What Changes

**symbols.go:** Add `ReturnType types.Type` field to the `Symbol` struct.
No change to `CollectDeclarations` — return types are computed in Pass 2.

**infer.go:**

Add `walkSubroutineDeclaration(node, source, st, annotations, diags)`:
1. Find the bareword (sub name) and block (body) children
2. Enter a sub scope in the symbol table
3. Walk the body with `walkNode` (types all expressions, processes returns)
4. Collect explicit return types by finding all `return_expression` nodes
   and looking up their value types in the annotation map
5. Collect the implicit return type from the last statement in the body
6. OR all return types into a union
7. Exit the sub scope
8. Store the union as `ReturnType` on the sub's Symbol via a new
   `UpdateReturnType(name, typ)` method

Add to the `walkNode` switch: `case "subroutine_declaration_statement"`.

Update `inferFunctionCallType`:
- After extracting the function name, look up the symbol
- If the symbol has a non-Unknown `ReturnType`, return it
- Otherwise fall through to existing builtin lookup

Add `collectReturnTypes(block, source, annotations)`:
- Walk direct children of the block, find all `return_expression` nodes
  (including nested inside conditionals)
- Find the last expression for implicit return
- Return the union of all found types

**No changes to:** `types/`, `narrowing.go`, `walkConditionalStatement`, or
any guard machinery. Return type inference is orthogonal to flow narrowing.

### Symbol Table Extension

```go
type Symbol struct {
    Name       string
    Type       types.Type   // sigil type or narrowed type
    ReturnType types.Type   // inferred return type (subs only, zero = unknown)
    Kind       SymbolKind
    StartByte  uint32
    EndByte    uint32
}
```

Add `UpdateReturnType(name string, typ types.Type) bool` to SymbolTable —
same pattern as `UpdateType`, but updates `ReturnType` field.

### Forward Reference Limitation

Pass 2 walks the CST top-down for subroutine bodies. If `foo()` is called
before `sub foo` is defined in the file, the call site won't see the return
type. This is acceptable for now — cross-file analysis (#394) can revisit
with multi-pass resolution.

### Interaction with Builtins

Builtins already have return types in `signatures.go`. The lookup order is:
1. Check `signatures.go` builtins first (authoritative)
2. Then check symbol table `ReturnType` for user-defined subs

This prevents user-defined subs from shadowing builtin return types.

## Out of Scope

- Cross-file return type propagation (deferred to #394)
- Forward reference resolution (requires multi-pass)
- Return type annotation syntax (Perl has no standard for this)
- Anonymous sub return types
- Method return types via `$obj->method()` (requires class analysis)
- Context-dependent return types (wantarray)

## Implementation Order

1. Add `ReturnType` field to Symbol + `UpdateReturnType` method
2. Add `collectReturnTypes` helper (explicit + implicit)
3. Add `walkSubroutineDeclaration` to the walker
4. Update `inferFunctionCallType` to use `ReturnType`
5. End-to-end tests: sub with return, call site gets narrowed type
