---
title: shift is introduced at tier 11, so plain array mutation cannot be written before the OO tier
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-26T13:24:36.843414308Z
updated: 2026-10-23T16:20:47.531524902Z
sessions:
- start_sha: 62cf900f71db986393d6d645b6814afefaea26e4
  end_sha: 9c3e164290f60fe6c5c1625dbd8e15a2f8a4dc3f
  commits: 3
  started_at: 2026-10-23T15:16:23.050278532Z
  ended_at: 2026-10-23T16:20:47.531524902Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-23T15:16:23.050278532Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-23T16:20:47.531524902Z
observed_paths:
- conformance/02_variables/README.md
- conformance/04_operators/README.md
- conformance/06_control/README.md
- conformance/mdtest/adjacency-02_variables.md
- conformance/mdtest/adjacency-06_control.md
- conformance/mdtest/aggregate-operators.md
- conformance/mdtest/arrays.md
- conformance/mdtest/blocks-and-terminations.md
- conformance/mdtest/comparison.md
- conformance/mdtest/conditionals.md
- conformance/mdtest/dynamic-scope.md
- conformance/mdtest/hashes.md
- conformance/mdtest/jumps.md
- conformance/mdtest/loops.md
- conformance/mdtest/names.md
- conformance/mdtest/subroutines.md
- conformance/mdtest/switch.md
- conformance/mdtest/symbol-table.md
- conformance/mdtest/taking-references.md
- internal/conformance/ratchet_test.go
- internal/conformance/testdata/corpus.ratchet
---

`shift` is introduced at tier 11 oo, so plain array mutation cannot be
written before the OO tier. Tier 02 introduces `push` and `unshift` and uses
both on a plain array; `shift` and `pop` are the other half of that family and
are not there.

Found while placing a contributed case from the B::SoN session whose fourth
body is `my @q = (3,1,2); my $first = shift @q;` -- ordinary array work that
the op budget currently forces into tier 11.

## Measured

Tier 02's introduces line (`adjacency-02_variables.md:6`):

    aassign, aelem, aelemfast, aelemfast_lex, aelemfastlex_store, aslice,
    av2arylen, delete, each, gv, gvsv, helem, hslice, multideref, padav,
    padhv, push, rv2av, rv2hv, sassign, unshift, values

`push` and `unshift` are there. `shift` and `pop` are not. And tier 02's own
case already writes `unshift @a, $#a;` on a plain array, so the tier is
committed to array mutation as its subject.

Tier 11's introduces line (`bless-dispatch.md:6`, and two siblings):

    anonhash, bless, emptyavhv, isa, method, method_named, method_super,
    methstart, shift, stub, tie, tied

`shift` sits among the OO ops because `sub new { my $class = shift }` needs
it. That is a real dependency for THAT case and the wrong place for the op.

## Why it is a grading defect and not a preference

The partial order is supposed to say "this construct needs those constructs".
A tier-02 case cannot currently write `shift @q` even though `shift` needs
nothing tier 02 does not already have -- `padav` and `aassign` are both tier
02's own. So the order asserts a dependency that does not exist, which is the
one thing the grading is for.

`lintOps` reads these lines, so the effect is not cosmetic: any case using
`shift` on a plain array is rejected below tier 11 or must be written at 11,
where a control-flow or array finding reads as an OO finding.

## Scope

1. Move `shift` to tier 02, and measure `pop` at the same time -- it is
   probably absent for the same reason and has the same shape. Check whether
   any other array or hash mutator is missing from tier 02's list by walking
   perlfunc's array section against it rather than by inspection.
2. Tier 11's `shift` use is `my $class = shift` with no argument, reading
   `@_`. Decide whether that is the same op for grading purposes -- it is the
   same OP, so moving it to 02 leaves tier 11 depending on 02, which it
   already does transitively through 08. Verify that the dependency chain
   still reaches it.
3. Re-run the whole corpus after the move: `lintOps` will now ALLOW things it
   rejected, which cannot break a passing case, but a case may have been
   written at a higher tier to work around this and should move down.

## Context

- paths: `conformance/mdtest/adjacency-02_variables.md:6` and every topic
  carrying a Tier 02 introduces line; `bless-dispatch.md`, `class-feature.md`,
  `tie.md` for the tier 11 line
- code: `internal/conformance/lint.go:143` (`lintOps`), `reachable`
- commands: `go test ./internal/conformance/ -count=1`
- note: the introduces lines are DUPLICATED across every topic in a tier, so
  a move must edit all copies. `TestGlossaryMatchesCategories` and the tier
  tests read them.

## Acceptance criteria

- [ ] `shift` on a plain array is legal at tier 02 (`go test ./internal/conformance/ -run 'TestMdtestTopicsMakeTheSameClaims' -count=1 | grep -q '^ok'`)
- [ ] Every copy of the tier 02 introduces line agrees (`test $(grep -rc 'Introduces .aassign' conformance/mdtest/ | grep -v ':0' | wc -l) -gt 0`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
