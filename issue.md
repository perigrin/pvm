---
title: No tier introduces the switch ops, so the corpus cannot hold a given/when case
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-26T17:33:59.477827339Z
updated: 2026-10-23T16:04:08.161215176Z
sessions:
- start_sha: 62cf900f71db986393d6d645b6814afefaea26e4
  end_sha: 394f64d005a7e2ed73b94278ca6a14099bc6cecc
  commits: 2
  started_at: 2026-10-23T15:16:22.429387141Z
  ended_at: 2026-10-23T16:04:08.161215176Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-23T15:16:22.429387141Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-23T16:04:08.161215176Z
observed_paths:
- conformance/04_operators/README.md
- conformance/06_control/README.md
- conformance/mdtest/adjacency-06_control.md
- conformance/mdtest/blocks-and-terminations.md
- conformance/mdtest/comparison.md
- conformance/mdtest/conditionals.md
- conformance/mdtest/dynamic-scope.md
- conformance/mdtest/jumps.md
- conformance/mdtest/loops.md
- conformance/mdtest/subroutines.md
- conformance/mdtest/switch.md
- conformance/mdtest/symbol-table.md
- conformance/mdtest/taking-references.md
- internal/conformance/ratchet_test.go
- internal/conformance/testdata/corpus.ratchet
---

The corpus has no `given`/`when`/`default` case and structurally cannot hold
one: the construct compiles to five ops that no tier introduces, so `lintOps`
would reject any case asserting it.

Found by asking "do we not have given/when in the corpus?" after `01a0de8c-4755`
landed a parser fix for it.

## Measured

`grep` across `conformance/mdtest/` finds `given`, `when` and `default` only in
PROSE and in unrelated identifiers -- `$default` as a variable name, "default"
in English, "when" mid-sentence. No `perl` block contains the construct.

The ops it compiles to, from `perl -MO=Concise,-exec` on 5.42.0:

    const enter entergiven enterwhen gvsv leave leavegiven leavewhen
    nextstate padsv padsv_store print pushmark smartmatch

Checked against every tier's `Introduces` line:

    entergiven   UNCLAIMED BY ANY TIER
    leavegiven   UNCLAIMED BY ANY TIER
    enterwhen    UNCLAIMED BY ANY TIER
    leavewhen    UNCLAIMED BY ANY TIER
    smartmatch   UNCLAIMED BY ANY TIER

Tier 06 control introduces `cond_expr`, `die`, `enteriter`, `entertry`, `exit`,
`goto`, `iter`, `last`, `leavetry`, `next`, `redo`, `time`, `unstack`. The
switch family is absent from it and from everything else.

So this is not "nobody wrote the case yet". `lintOps` reads those lines and
rejects a case using an op no reachable tier introduces, which means the corpus
cannot express the construct until a tier claims the ops.

## Why it matters that the fix landed without one

`01a0de8c-4755` left three unit tests in `internal/parse/switch_test.go` --
`TestGivenWhenDefault`, `TestSwitchIsNotAnIndexCall`,
`TestSwitchKeywordsAreNotCalls`. That is real coverage and it is not corpus
coverage: those assert OUR tree, where a corpus case asserts PERL'S OUTPUT,
byte-exact, with perl adjudicating. The corpus is the instrument that catches a
tree that is confidently wrong, which is exactly how this construct was failing
before the fix -- `given (1) { print "a" }` was Unknown=0 and parsed as
`index(call(given,1), call(print))`.

A construct we just taught the parser to read, with no case pinning what perl
prints for it, is the gap this issue closes.

## smartmatch is a SEPARATE gap, measured

`smartmatch` is reachable without `given`/`when`:

    perl -MO=Concise,-exec -e 'no warnings; my @a=(1); my $r = (1 ~~ @a);'
      -> smartmatch

So `~~` is its own unclaimed construct and its own missing case, not a
by-product of the switch family. Whichever tier claims `smartmatch` should be
chosen for `~~` on its own merits, and the switch ops may or may not belong
with it.

## Scope

1. Decide which tier introduces `entergiven`, `leavegiven`, `enterwhen`,
   `leavewhen`. Tier 06 control is the obvious home -- it already owns the
   jump family and `cond_expr` -- but measure the dependency honestly: a
   `when` block's body can contain anything, and `given` takes an expression,
   so check what tier 06 depends on and whether the switch family needs more.
2. Decide `smartmatch` separately. `~~` is a binary operator, not a statement
   form, so it may belong in an operator tier rather than with the blocks.
3. Write the case or cases. At minimum: `given` with `when` and `default`,
   the parenless `default BLOCK`, and `when` with a non-scalar argument, since
   those are the three shapes `01a0de8c-4755` had to distinguish.
4. Record in the case's prose that the feature is DISABLED from bundle 5.36
   onward and enabled in 5.10-5.34, and that `use warnings` does NOT warn about
   it. That was measured false in the issue that fixed it -- `perl -we` and
   `use warnings "deprecated"` both compile it silently -- so the case should
   carry the truth rather than let the next reader assume.

## Context

- paths: `conformance/mdtest/adjacency-06_control.md` and every tier-06 topic
  carrying the `Introduces` line (they are duplicated per topic);
  `internal/conformance/lint.go:143` (`lintOps`), `internal/conformance/lint.go`
  (`reachable`)
- commands: `go test ./internal/conformance/ -count=1`
- related: `01a0de8c-4755` (the parser fix, landed), `01a0dde3` (`shift` is
  introduced at tier 11, the same class of grading defect)
- note: `01a0de11` records that `lintOps` reads Concise's RENDERED text, so it
  sees op names but not flags. Op names are what this issue needs, so that
  limitation does not bite here.

## Acceptance criteria

- [ ] A tier introduces the four switch ops (`grep -qE 'entergiven|leavegiven' conformance/mdtest/*.md`)
- [ ] A corpus case asserts what perl prints for given/when/default (`go test ./internal/conformance/ -run 'TestMdtestTopicsMakeTheSameClaims' -count=1 | grep -q '^ok'`)
- [ ] The case records that the feature is off from 5.36 and does not warn (`grep -qiE '5\.36|does not warn' conformance/mdtest/*.md`)
- [ ] REGRESSION GUARD, passes today: the corpus stays complete (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
