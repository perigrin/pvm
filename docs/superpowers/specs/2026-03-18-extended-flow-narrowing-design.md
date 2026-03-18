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
    condition (named)   ← extract guard from this
    ) (anon)
    block (named)       ← walk with guard
    else (named, optional)
```

**Approach:** Replace the current "walk elsif children blindly" code with a
`walkElsifNode` helper that mirrors `walkConditionalStatement`:
1. Scan the elsif's children to find condition, block, and optional else
2. Walk the condition
3. Extract a guard pattern from the condition
4. Walk the block with `walkBlockWithGuard` using the guard
5. Walk the else child (if present) with the negated guard

Each elsif branch gets only its own guard. No guard combination across branches.

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

CST for `!defined($x)`:
```
unary_expression
  ! (anon)
  func1op_call_expression   ← extract guard from this
```

CST for `not defined($x)`:
```
ambiguous_function_call_expression
  function: "not"
  func1op_call_expression   ← extract guard from this
```

The `guardResult` struct gains a `Negated bool` field. When `Negated` is true,
`walkConditionalStatement` flips the negate flag for the if-body (so negated
condition + negated application = positive guard in the if-body).

**B. Early-exit detection:**

A function `blockAlwaysExits(block, source)` scans a block's direct named
children for unconditional exit statements:
- `return_expression` node kind
- `bareword` with text `die`, `croak`, or `exit`

Only checks top-level statements in the block. Nested exits inside inner
conditions don't count (conservative — avoids false positives).

**C. Post-if narrowing:**

In `walkConditionalStatement`, after walking the if-block:
- If the block always exits AND there is no else/elsif
- Apply the negated guard type via `st.UpdateType` to the guarded variable

For `if (!defined($x)) { return; }`:
- Condition: negated defined($x) → guardResult has GuardDefined + Negated=true
- If-body gets the positive guard (negated + negated = positive): Scalar
- But the if-body always exits, so the if-body narrowing is moot
- Post-if: the negated guard (which is the positive defined guard) is applied
  → $x narrowed to Scalar (non-undef) for the rest of the scope

For `if (!ref($x)) { return; }`:
- Post-if: $x narrowed to Ref

The mechanism uses `st.UpdateType` (same as assignment narrowing) so the type
change persists in the current scope. No new scope machinery.

### 3. Loop Variable Narrowing (for_statement)

**CST structure** (verified by discovery):

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

**Approach:** Special-case `for_statement` in `walkNode`:
1. Walk the iteration source (array/list expression) to type it
2. Enter a scope for the loop body
3. Define the loop variable as Scalar (element type of a Perl array)
4. Walk the block body
5. Exit scope

The loop variable is properly scoped — it doesn't leak after the for block.

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

1. elsif guard narrowing — `walkElsifNode` helper + tests
2. Negation unwrapping — `Negated` field on `guardResult`, update `extractGuardPattern`
3. Early-exit detection — `blockAlwaysExits` function + tests
4. Post-if narrowing — wire early-exit + negated guards in `walkConditionalStatement`
5. For-loop scoping — `walkForStatement` + tests
6. C-style for verification — test that assignment narrowing works
