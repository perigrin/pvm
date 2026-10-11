---
title: A bare block nested in a loop lets a statement modifier escape it
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-27T01:18:51.711058099Z
updated: 2026-10-11T11:31:03.089632857Z
sessions:
- start_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  end_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  commits: 0
  started_at: 2026-10-11T11:31:03.018042945Z
  ended_at: 2026-10-11T11:31:03.089632857Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T11:31:03.018042945Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-11T11:31:03.089632857Z
---

A bare block nested inside a loop lets a statement modifier ESCAPE it: the
modifier is re-emitted outside the inner block's closing brace, and the block's
own `}` is left over as an Unknown.

Found by tracing the `}` bucket of perl.git t/ at `e87dee9c` (11 files, 222
nodes) to its MATCHING opener with a depth counter, then narrowing.

## Measured at `e87dee9c`

    source                             U   canon
    for my $t (1) { { 3 if 1; } }      1   for my $t (1) {{3} if 1;} }

The modifier `if 1` belongs to the statement `3` INSIDE the inner block. Canon
emits `{3}` and then `if 1;` after the brace, so the modifier now guards the
whole block, and the outer loop's `}` has nowhere to go and becomes the Unknown.

Controls, all CORRECT today, which is what makes the shape narrow:

    for my $t (1) { 3 if 1; }              0   correct, no inner block
    for my $t (1) { for my $i (2) { 3 if 1; } }  0   correct, inner is a LOOP
    for my $a (1) { for my $b (2) { for my $c (3) { 4; } } }  0  correct

So it takes a BARE block (not a loop, not a sub) nested inside a loop, with a
modifier on a statement inside it.

## Perl keeps the modifier inside

    perl -MO=Deparse -e 'for my $t (1) { { print 3 if 1; } }'
      ->  foreach my $t (1) { { print 3; } }

The modifier is constant-folded away there, but the nesting is preserved: the
inner block stays, and nothing moves outside it.

## Scope

1. Find why a bare block's contents let a modifier reach the enclosing level.
   The six-keyword modifier canon fix (`01a0dfc7`) made `Modifier` a flag on the
   node; check whether a bare block's statement list loses it, or whether
   `blockForm` classifies the inner block such that the modifier is attributed
   to the block rather than to the statement.
2. Measure the same shape with `while`, `until` and a bare `{ }` at file scope,
   and with `unless`/`foreach` as the modifier. If only some combinations break,
   say which.
3. A corpus case. The count is 1 here so it is not invisible, but the canon is
   the thing to assert -- the Unknown is the leftover `}`, not the defect.

## Context

- paths: `internal/parse/canon.go` (the `Conditional, Loop` case and
  `blockForm`), `internal/parse/parse.go` (`applyModifier`)
- commands: `go test ./internal/parse/ -count=1`
- related: `01a0dfc7` (the modifier/block-form canon split -- the same code, and
  this may be a residue of it), `01a0e04d`-sibling `for $i (LIST)` swapping its
  variable and list, found in the same trace

### Re-measured (2026-10-11)

Measured at pu `24bfac4a`. `for my $t (1) { { 3 if 1; } }` scores Unknown=0 and canons as `for my $t (1) {{3 if 1;} }`, a fixpoint: the modifier stays inside the inner block, as perl keeps it. Scope item 2's combinations all hold too: `while` and `until` loops, a bare block at file scope, and `unless`, `foreach`, `for`, `while` and `until` as the modifier. Fixed by d8029d82, where a block's opening brace leaves a statement boundary in the lexer.

TestNestedBareBlock, from d8029d82, has no modifier case, and TestModifierInNestedBareBlock existed only in an uncommitted worktree. It was written in test/pgt-reconcile and asserts each shape's canon, zero Unknowns and fixpoint. With d8029d82's `openedBlock` transition removed it fails with this issue's own symptom, the modifier outside the inner brace and an Unknown left over.

The conformance and whole-suite commands carry `-timeout=30m`: internal/conformance took 696s in a full run, past go test's 10-minute default, and timed out without it.

## Acceptance criteria

- [ ] A modifier inside a nested bare block stays inside it (`go test ./internal/parse/ -run '^TestModifierInNestedBareBlock$' -count=1 -v | grep -q '^--- PASS: TestModifierInNestedBareBlock'`)
- [ ] The outer loop's closing brace is no longer an Unknown (`go test ./internal/parse/ -run '^TestModifierInNestedBareBlock$' -count=1 -v | grep -q '^--- PASS: TestModifierInNestedBareBlock'`)
- [ ] REGRESSION GUARD, passes today: a modifier with no inner block is unchanged (`go test ./internal/parse/ -run '^TestStatementModifierKeepsItsBody$' -count=1 -v | grep -q '^--- PASS: TestStatementModifierKeepsItsBody'`)
- [ ] REGRESSION GUARD, passes today: the corpus stays complete (`go test ./internal/conformance/ -count=1 -timeout=30m`)
