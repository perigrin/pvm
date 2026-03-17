# Flow Narrowing Design

## Problem

PSC's type inference engine narrows variable types through assignments (`$x = 42`
makes `$x` an Int) but ignores control flow. When a programmer writes
`if (defined($x))`, the code inside that block knows `$x` is not Undef — but PSC
does not. This design adds guard-based flow narrowing so PSC tracks type
refinements through if, unless, and while conditions.

## Approach

Special-case the inference walker for if/unless/while statements. Instead of the
generic post-order child recursion, walk children in a specific order:

1. Walk the condition expression (so its children get typed)
2. Extract a guard pattern from the condition
3. Enter a new scope, shadow the guarded variable with its narrowed type
4. Walk the block body
5. Exit the scope (shadow disappears, original type restored)

For if/else, apply the negated guard type in the else block.

## Guard Patterns

Four condition shapes, recognized by CST structure:

| Perl condition          | CST shape                              | Guard               |
|-------------------------|----------------------------------------|----------------------|
| `defined($x)`          | func1op: keyword `defined`, scalar arg | GuardDefined         |
| `ref($x)`              | func1op: keyword `ref`, scalar arg     | GuardRef (plain)     |
| `ref($x) eq 'HASH'`   | equality_expression: ref call eq string| GuardRef + RefType   |
| `$x isa Foo`           | isa-expression with scalar LHS         | GuardIsa             |

A new function `extractGuardPattern` reads the condition node and returns the
variable name, guard pattern, and whether a guard was recognized. It lives in
`infer.go` alongside the other `infer*` helpers.

## Negated Guards

`unless` applies the negated guard to its body. `else` applies the negated guard
after a positive `if`. A new function `NegateGuard` in `types/narrowing.go`
computes the negated type:

| Guard                  | Positive type       | Negated type         |
|------------------------|---------------------|----------------------|
| GuardDefined           | Scalar (non-Undef)  | Undef                |
| GuardRef (plain)       | Ref                 | Scalar (non-ref)     |
| GuardRef with RefType  | specific ref type   | no narrowing         |
| GuardIsa               | Object              | no narrowing         |

RefType-specific and isa negations are not useful ("not a HashRef" could be
anything), so we skip narrowing in those negated branches.

## Scope Mechanism

The symbol table already supports `EnterScope`/`ExitScope` with lexical
shadowing. Flow narrowing reuses this: enter a `"guard"` scope, define a shadow
symbol with the narrowed type, walk the body, exit the scope. The shadow
disappears and the original type is restored. No new scope machinery required.

## Walker Changes

`walkNode` in `infer.go` gains a check before the generic child recursion loop.
When the node is `conditional_statement` (if/unless) or `loop_statement` (while):

1. Identify the condition expression, if-block, and else-block children
2. Walk the condition with `walkNode`
3. Call `extractGuardPattern` on the condition
4. For if/while: shadow with positive narrowing in the body block
5. For unless: shadow with negated narrowing in the body block
6. For else blocks: shadow with the opposite narrowing
7. Return `types.Unknown` (statements have no type)
8. Skip the generic child loop

## CST Discovery

Before implementation, parse sample Perl snippets and dump the CST to confirm:

- The exact node kind for `$x isa Foo` expressions
- How if_statement children are laid out (condition, block, else positions)
- Whether `unless` is a separate node kind or a variant of `if_statement`

## Out of Scope

- elsif chains (deferred — walk normally with no narrowing)
- Nested guard combination (`ref($x) && ref($x) eq 'HASH'`)
- Early-exit narrowing (after `return`/`die`)
- Loop variable narrowing (`for my $item (@arr)`)
- Branch merging at join points (union types after if/else rejoin)

## Implementation Order

1. CST discovery — parse samples, map node structures
2. `NegateGuard` in `types/narrowing.go` — negated guard logic + tests
3. `extractGuardPattern` in `infer.go` — read condition CST, return guard + variable
4. Special-case `walkNode` for if/unless/while — condition-first walk with scoped shadowing
5. Tests — Perl snippets exercising each guard in if, unless, if/else
