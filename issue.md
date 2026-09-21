---
title: The corpus format has expect parsent and the lint cannot carry such a file
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T22:11:13.008302813Z
updated: 2026-09-21T22:54:39.675647307Z
sessions:
- start_sha: b41801021927f68fcbf799c60d6b947393d983de
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T22:29:10.466869947Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T22:29:10.466869947Z
---

The corpus format documents `--- expect parsent` for a file perl REFUSES,
`ParseFile` reads it, `File.ExpectParsent` exists, and `verdict()` handles
it. But **no corpus file uses it, and the first one to try will fail the
dependency lint**.

MEASURED while executing 01a0c432-fbd5 (the call-form slice).

## Where the failure actually is

Not in perl, and not in `B::Concise`. All three invocations exit 255 on a
syntax error, measured against a file containing `f 1, 2;` with `f`
undeclared:

    perl -c file.pl              exit 255
    perl file.pl                 exit 255
    perl -MO=Concise,-exec ...   exit 255

The difference is entirely in how the two Go callers treat that exit.

`askPerl` in `run.go` treats it as an ANSWER:

    if err := exec.Command(perl, "-c", path).Run(); err != nil {
        return false, ""
    }

so a refusing file yields `compiles=false` and `verdict()` proceeds
correctly. The runner is already right.

`opsOf` in `lint.go` treats it as an ERROR:

    out, err := exec.Command(perl, "-MO=Concise,-exec", path).Output()
    if err != nil {
        return nil, fmt.Errorf("perl -MO=Concise refused the source: %w", err)
    }

and `TestCorpusLints` plus every tier's `*Lint` test call `opsOf`
unconditionally. So the lint fails on any `parsent` file, in any tier.

## Why it matters rather than being theoretical

The croak extraction (01a0c35f-1685, done) measured **203 cases** in perl's
own `t/lib/croak/` that fail `perl -c` -- the genuine must-not-parse
negatives, and the spec's whole argument for extracting them. None can
become a corpus file until this is fixed.

The case that found it: `f 1, 2` with `f` undeclared is a syntax error,
true with and without `use strict` and even when `sub f` is defined later
in the same file. That fact deserves a corpus file and could not have one;
it is asserted in Go instead, inside `TestTierCallFormsParenlessExtent`.

## The fix

`opsOf` should not be called for a file whose `ExpectParsent` is set, or
should return an empty op set rather than an error for one. Either makes
the lint vacuous for such files, which is correct: a program perl will not
compile has no optree, and the dependency check has nothing to say about
it. Prefer the caller-side skip if the callers can see the `*File` -- an
`opsOf` that silently returns empty for a genuinely broken file would hide
a real error in a `parses` file.

## What the fix must not break

`verdict()` consults perl FIRST, so a file claiming `parsent` that perl
ACCEPTS is reported as a CORPUS BUG before the parser is consulted at all
(`TestPerlAdjudicatesFirst` pins this). Whatever skips the lint must not
also skip that adjudication, or a wrong `parsent` claim becomes
unfalsifiable.

## Acceptance

- [x] The dependency lint skips a `parsent` file rather than erroring on
      it, at ONE place all seventeen `lintOps` call sites route through
      rather than at each of them, and `opsOf` is left alone so a
      genuinely broken `parses` file still errors.
      (`go test ./internal/conformance/ -run 'TestCorpusLints' -count=1`)

- [x] The corpus carries its first `--- expect parsent` file:
      `conformance/07_subroutines/10_undeclared_callee.t`, the undeclared
      parenless callee, the fact `TestTierCallFormsParenlessExtent`
      asserts in Go because no corpus file could carry it.
      (`test -f conformance/07_subroutines/10_undeclared_callee.t && grep -qx -- '--- expect parsent' conformance/07_subroutines/10_undeclared_callee.t`)

- [x] Perl still ADJUDICATES that file's claim before anything else: the
      tier's own perl validator runs over it and would report a CORPUS
      BUG if perl accepted the source, and `TestPerlAdjudicatesFirst`
      still holds.
      (`go test ./internal/conformance/ -run 'TestTierSubroutinesPerlValidated|TestPerlAdjudicatesFirst' -count=1 -v 2>&1 | grep -E '10_undeclared_callee|TestPerlAdjudicatesFirst'`)

- [x] A `parses` file whose source perl genuinely refuses still ERRORS
      from the lint rather than linting vacuously -- the failure mode the
      caller-side skip exists to avoid.
      (`go test ./internal/conformance/ -run 'TestLintErrorsOnABrokenParsesFile' -count=1 -v`)

- [x] The new file joins the tier's adjacency table rather than being
      invisible to it: `subroutineConstructs` gains its entry, which is
      what `TestTierSubroutinesAdjacency` demands of every construct file.
      (`go test ./internal/conformance/ -run 'TestTierSubroutinesAdjacency' -count=1`)

- [x] The ratchet records the new file and the whole conformance suite is
      green.
      (`go test ./internal/conformance/ -count=1`)
