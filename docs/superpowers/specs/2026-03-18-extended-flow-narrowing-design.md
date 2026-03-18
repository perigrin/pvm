# Extended Flow Narrowing Design

## Problem

PSC's flow narrowing handles if/unless/while guards but stops short of three
common Perl patterns: elsif chains get no guard narrowing, early-exit guard
clauses (`if (!defined($x)) { return; }`) don't refine the post-if scope, and
for-loop variables lack proper scoping in the inference walker.

## Features

### 1. elsif Guard Narrowing

Currently elsif nodes are walked without extracting guards. Each elsif branch
should apply its own guard independently.

**CST structure** (verified by discovery):

```
conditional_statement
  if (anon)
  condition (named)
  block (named)
  elsif (named)
    elsif (anon keyword)
    ( (anon)
    condition (named)   <- extract guard from this
    ) (anon)
    block (named)       <- walk with guard
    else (named, optional — may be an "else" or another "elsif")
```

**Approach:** Replace the current "walk elsif children blindly" code with a
`walkElsifNode` helper that mirrors `walkConditionalStatement`:
1. Scan the elsif's children to find condition, block, and trailing node
2. Walk the condition
3. Extract a guard pattern from the condition
4. Walk the block with `walkBlockWithGuard` using the guard
5. If the trailing child is an `else` node, walk its block with the negated guard
6. If the trailing child is another `elsif` node, recurse into `walkElsifNode`

Each elsif branch gets only its own guard. No guard combination across branches.
Negated guards in elsif conditions (e.g. `elsif (!ref($x))`) will work once
negation unwrapping (feature 2A) is implemented; until then, they are silently
ignored (no narrowing applied, same as any unrecognized condition).

### 2. Early-Exit Narrowing

The common Perl guard clause pattern:

```perl
if (!defined($x)) { return; }
# $x is known defined here
```

Requires three pieces:

**A. Negation unwrapping in `extractGuardPattern`:**

Recognize `unary_expression` with `!` operator and `ambiguous_function_call_expression`
with `function: "not"` as negation wrappers. Recurse into the inner expression to
find the guard, and set a `negated` flag on the `guardResult`.

CST for `!defined($x)` (verified by discovery):
```
unary_expression
  ! (anon)
  func1op_call_expression   <- extract guard from this
```

CST for `not defined($x)` (verified by discovery):
```
ambiguous_function_call_expression
  function: "not"
  func1op_call_expression   <- extract guard from this
```

The `guardResult` struct gains a `Negated bool` field. When `Negated` is true,
`walkConditionalStatement` flips the negate flag for the if-body (so negated
condition + negated application = positive guard in the if-body).

**B. Early-exit detection:**

A function `blockAlwaysExits(block, source)` scans a block's direct named
children for unconditional exit statements. The following node kinds are
recognized (all verified by CST discovery):

- `return_expression` — `return` and `return EXPR`
- `func1op_call_expression` with keyword `exit` — `exit` and `exit N`
- `bareword` with text `die` — bare `die;` (note: `die "msg"` fails to parse
  due to the gotreesitter string literal limitation, so only bare `die` is
  detectable)

`croak` is excluded because it is a Carp module function, not a Perl builtin.
Its parse representation varies and is unreliable.

Only checks top-level statements in the block (children of the `block` node
wrapped in `expression_statement`). Nested exits inside inner conditions
don't count (conservative — avoids false positives).

**C. Post-if narrowing:**

In `walkConditionalStatement`, after walking the if-block:
- If the block always exits AND there is no else/elsif
- Apply the "else-branch guard" to the guarded variable via `st.UpdateType`

The "else-branch guard" is the type that would be applied in a hypothetical
else block — the opposite of what the if-body received. Concretely:

```
if guard.Negated is true:
    # Condition was !guard. If-body got positive narrowing.
    # Else-branch (= post-exit scope) also gets positive narrowing.
    # Use NarrowByGuard(currentType, guard.Guard)
    elseType, _ = NarrowByGuard(currentType, guard.Guard)

if guard.Negated is false:
    # Condition was guard. If-body got positive narrowing.
    # Else-branch (= post-exit scope) gets negated narrowing.
    # Use NegateGuard(currentType, guard.Guard)
    elseType, _ = NegateGuard(currentType, guard.Guard)
```

Example: `if (!defined($x)) { return; }`
- Condition: `!defined($x)` -> guardResult{GuardDefined, Negated: true}
- If-body: negated condition, so negate flag is flipped -> positive guard applied
  (Scalar). But the body always exits, so this is moot.
- Post-if: guard.Negated is true, so apply NarrowByGuard -> Scalar (non-undef)
- Result: $x narrowed to Scalar for the rest of the scope.

Example: `unless (defined($x)) { return; }`
- Condition: `defined($x)` -> guardResult{GuardDefined, Negated: false}
- keyword is "unless", so isUnless=true
- If-body: negate flipped by unless -> negated guard applied (Undef). Body exits.
- Post-if: guard.Negated is false AND isUnless is true. The else-branch for
  unless is the positive guard. Use NarrowByGuard -> Scalar (non-undef).
- Result: $x narrowed to Scalar for the rest of the scope.

The general rule: the post-exit type is whatever type the code AFTER the if
can rely on — the opposite of what the exiting branch proved.

The mechanism uses `st.UpdateType` (same as assignment narrowing) so the type
change persists in the current scope. No new scope machinery needed.

Note: if `walkConditionalStatement` is called from within an already-narrowed
guard scope (nested if-statements), `st.UpdateType` will update the variable
in the innermost scope that contains it, which may be a guard scope from a
parent if. This is acceptable — the narrowed type is still correct within
that scope, and it will be discarded when the parent scope exits.

### 3. Loop Variable Narrowing (for_statement)

**CST structure** (verified by discovery):

With `my`:
```
for_statement
  for (anon)
  my (anon)
  scalar (named — loop variable)
  ( (anon)
  array / list_expression (named — iteration source)
  ) (anon)
  block (named)
```

Without `my` (verified: same structure, `my` anon child absent):
```
for_statement
  for (anon)
  scalar (named — loop variable)
  ( (anon)
  array / list_expression (named — iteration source)
  ) (anon)
  block (named)
```

**Approach:** Special-case `for_statement` in `walkNode`:
1. Walk the iteration source (array/list expression) to type it
2. Enter a scope for the loop body
3. Define the loop variable as Scalar (element type of a Perl array)
4. Walk the block body
5. Exit scope (use `defer` for safety)

The loop variable is properly scoped — it doesn't leak after the for block.

Note: inner `my` declarations inside the for-loop body were collected in
pass 1's block scope that no longer exists in pass 2. `UpdateType` calls
for those inner variables will be no-ops. This is the same known limitation
as in `walkBlockWithGuard`.

**C-style for (`cstyle_for_statement`):** The initializer `my $i = 0` already
goes through assignment narrowing via the generic walk. No special handling
needed. Verify with a test.

## Out of Scope

- Guard combination across branches (`if (defined($x)) ... elsif (ref($x))` does
  NOT produce "not-defined AND ref" in the elsif)
- Nested early-exit analysis (only top-level exits in a block count)
- Element type tracking for arrays (loop var is always Scalar for now)
- `unless (!guard)` double-negation (rare, not worth the complexity)
- Post-loop narrowing (what type does the variable have after the loop?)

## Implementation Order

1. elsif guard narrowing — `walkElsifNode` helper + tests (including 3+ branch chain)
2. Negation unwrapping — `Negated` field on `guardResult`, update `extractGuardPattern`
3. Early-exit detection — `blockAlwaysExits` function + tests (including negative test)
4. Post-if narrowing — wire early-exit + negated guards in `walkConditionalStatement`,
   including `unless` + always-exits case
5. For-loop scoping — `walkForStatement` + tests (with/without `my`, scope leak test,
   list_expression vs array iteration source)
6. C-style for verification — test that assignment narrowing works
