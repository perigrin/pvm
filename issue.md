---
title: Expand PSC inference to cover more Perl type semantics
state: pending
urgency: normal
milestone: psc-type-alignment
blocked_by:
- 01a0c0e7-8234-7f04-80e5-1fab8682f9bd
created: 2026-03-22T02:43:47.968515-04:00
updated: 2026-09-20T22:20:33.870251335Z
---

**2026-09-20: blocked on 01a0c0e7, and the split recommendation below is
superseded.**

That recommendation was "undef propagation now against the tree-sitter tree,
parameter inference later against the lowered one", on the argument that
undef propagation is "a lattice question more than a tree question" and so
mostly survives the conversion.

perigrin's direction: **do no work against tree-sitter, it is being burned
down.** That is a stronger constraint than "most of it survives" -- surviving
partially is not a reason to write code against a tree that is being deleted,
and §6.1.6 counts why nothing survives cleanly anyway:

> Ten methods port; the vocabulary does not.

`internal/infer` switches on 34 tree-sitter grammar names that
`internal/parse` does not emit. So BOTH halves of this issue wait on the
lowering pass, not just the parameter half.

Also stale below: area 3's note to "check for double-ownership before
starting" against 01a0ad3f. That issue is DONE (f066f2ab, 2026-09-19).
Cross-file inference remains out of scope here regardless.

---

Expand PSC's type inference beyond arity and container-level mismatches to value-level checking, as the formal papers describe.

Filed 2026-03-22. Re-measured 2026-09-17 at 3c997a6b by building `./cmd/psc`
and running it: two of the six expansion areas have since shipped, one
acceptance criterion already passes, and two do not. What follows is the
current state, not the original list.

## Measured state of the six areas

**1. String-to-number coercion — SHIPPED.** The first acceptance criterion
below already passes:

    my $x = "hello";
    my $y = $x + 1;
    -> ac1.pl:2:9: warning: left operand of "+": expected Num, got Str
       [coercion-mismatch]

**2. Undef propagation — NOT IMPLEMENTED.** `my $x; my $y = $x + 1;` produces
no diagnostic at all. This is the real remaining work in this issue.

**3. Cross-file return type inference — UNMEASURED here.** Untested in this
pass because it needs a multi-file fixture, and because it now overlaps
01a0ad3f (resolve imports by parsing the module's own source), which owns
how a second file is reached at all. Check for double-ownership before
starting.

**4. User-defined function signatures from usage — NOT IMPLEMENTED.**
`sub add { my ($a,$b) = @_; return $a+$b } add("x", 2);` produces no
diagnostic; nothing infers that `$a` is used numerically.

**5. Context-sensitive inference — PARTLY SHIPPED.** `types.NarrowByContext`
exists and is called at `internal/infer/infer.go:524`. §4.14.2 records the
gap that remains: `NarrowByContext` has no *source* of context, because the
CST has no `Wantarray` node and no call-site context propagation. That half
is blocked on the lowering, not on this issue.

**6. Pattern-matching and type guards — LARGELY SHIPPED.** `defined`, `ref`
and `isa` guards are implemented (`infer.go:2065-2082`, `:2536-2538`).
Guard-based flow narrowing shipped in v1.0.0-rc52. What remains is CST
extraction for `ref($x) eq 'TYPE'` and `$x isa Foo` — the type-system half is
done and the walker's `extractGuardPattern` needs the two node-kind branches.
That is GitHub #384, effort:medium, and it should be its own issue rather
than a line here.

## What this issue is now

Areas 2 and 4: undef propagation, and parameter types inferred from usage.
Everything else has shipped, moved, or belongs elsewhere.

## The framing that changed underneath this issue

Filed before the lowering decision. Chapter 6 §6.1.6 (committed ef839f93)
establishes that inference does not walk the CST directly: a lowering pass
produces the tree it reads, with the vocabulary of §4.14. `internal/infer`
still consumes the OLD tree-sitter tree through `internal/parser/`, switching
on 57 grammar names.

This matters for scope. Work done here against the tree-sitter tree is work
that the conversion will have to redo. Two options, and the issue should pick
one rather than drift:

- **Do it now against the current tree.** Undef propagation is a lattice
  question more than a tree question, so most of it survives the conversion.
  Cheapest if the conversion is far off.
- **Wait for the lowered tree.** §4.14.2 records that `internal/infer`
  re-derives structure from source TEXT in several places -- `findOperatorText`
  (`infer.go:645-659`), and `callElementType` (`:1313`) telling map from grep
  by `strings.HasPrefix` on source. Those are exactly the places new inference
  rules would have to reproduce.

Recommend the first for undef propagation and the second for parameter
inference, which needs call-site argument shapes the lowered `Call` node
provides and the grammar names do not.

## Acceptance

- [ ] A non-numeric string literal in numeric context diagnoses (`go test ./internal/infer/ -run '^TestCoercionMismatchOnStringInNumericContext$' -count=1 -v | grep -q '^--- PASS: TestCoercionMismatchOnStringInNumericContext'`)
- [ ] `my $x; my $y = $x + 1;` diagnoses undef in a numeric context (`go test ./internal/infer/ -run '^TestUndefInNumericContext$' -count=1 -v | grep -q '^--- PASS: TestUndefInNumericContext'`)
- [ ] `my $x; if (defined $x) { my $y = $x + 1 }` does NOT diagnose — the guard narrows it (`go test ./internal/infer/ -run '^TestDefinedGuardSuppressesUndefDiagnostic$' -count=1 -v | grep -q '^--- PASS: TestDefinedGuardSuppressesUndefDiagnostic'`)
- [ ] Undef propagates through assignment: `my $x; my $y = $x; my $z = $y + 1;` diagnoses at the addition (`go test ./internal/infer/ -run '^TestUndefPropagatesThroughAssignment$' -count=1 -v | grep -q '^--- PASS: TestUndefPropagatesThroughAssignment'`)
- [ ] A parameter used numerically in its body is inferred Num, and a string argument at a call site diagnoses (`go test ./internal/infer/ -run '^TestParameterTypeInferredFromBody$' -count=1 -v | grep -q '^--- PASS: TestParameterTypeInferredFromBody'`)
- [ ] A parameter used in no typed position stays Unknown rather than being guessed (`go test ./internal/infer/ -run '^TestUnconstrainedParameterStaysUnknown$' -count=1 -v | grep -q '^--- PASS: TestUnconstrainedParameterStaysUnknown'`)
- [ ] No new diagnostic fires on the existing corpus fixtures (`go test ./internal/infer/ -count=1 2>&1 | grep -qv FAIL`)
- [ ] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

The third and sixth criteria are the negative scenarios: an undef check with
no guard awareness would fire on every `my $x;` in the corpus, and parameter
inference that guesses is worse than one that declines. §4.8.3's rule applies
to types as much as to parses -- an honest Unknown is never a wrong answer.

Not in scope: cross-file inference (area 3, overlaps 01a0ad3f), context
sources for `NarrowByContext` (area 5, needs the lowered tree), and
`ref($x) eq 'TYPE'` / `$x isa Foo` extraction (area 6, GitHub #384, its own
issue).

Key files: internal/infer/infer.go, internal/types/signatures.go
See also: GitHub #409, GitHub #384
