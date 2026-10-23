---
title: The corpus format has expect parsent and the lint cannot carry such a file
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T22:11:13.008302813Z
updated: 2026-10-23T23:40:18.369007124Z
sessions:
- start_sha: b41801021927f68fcbf799c60d6b947393d983de
  end_sha: a3a0ed0c97ce48a68f5db427979f9d8acf88d3c8
  commits: 1
  started_at: 2026-09-21T22:29:10.466869947Z
  ended_at: 2026-09-21T23:33:01.056333366Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T22:29:10.466869947Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T23:33:01.056333366Z
observed_paths:
- conformance/01_literals/README.md
- conformance/02_variables/README.md
- conformance/03_context/README.md
- conformance/04_operators/README.md
- conformance/05_scoping/README.md
- conformance/06_control/README.md
- conformance/07_subroutines/08_parenless_extent.t
- conformance/07_subroutines/10_undeclared_callee.t
- conformance/07_subroutines/README.md
- conformance/08_references/README.md
- conformance/09_regex/README.md
- conformance/10_io/README.md
- conformance/11_oo/README.md
- conformance/12_packages/README.md
- conformance/13_opaque/README.md
- conformance/14_recursive/README.md
- internal/conformance/lint.go
- internal/conformance/lint_test.go
- internal/conformance/numbering_test.go
- internal/conformance/prereq_test.go
- internal/conformance/subops_test.go
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier01_test.go
- internal/conformance/tier02_test.go
- internal/conformance/tier03_test.go
- internal/conformance/tier04_test.go
- internal/conformance/tier05_test.go
- internal/conformance/tier06_test.go
- internal/conformance/tier07_callforms_test.go
- internal/conformance/tier07_test.go
- internal/conformance/tier08_test.go
- internal/conformance/tier09_test.go
- internal/conformance/tier10_test.go
- internal/conformance/tier11_test.go
- internal/conformance/tier12_test.go
- internal/conformance/tier13_test.go
- internal/conformance/tier14_test.go
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

Criteria updated 2026-10-23: e8ee9001 and a6b33a2a replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `10_undeclared_callee.t` with `--- expect parsent` -> the `An undeclared callee is a syntax error` case of `argument-extent.md` carrying `parses: no`). Left unchanged here: `TestTierSubroutinesAdjacency`; the note below maps each to the test that now holds it.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, and that each case's main program is multi-statement; the declared prerequisite is pinned by `TestTierPrerequisitesMatchTheSpec` instead. Old -> new: `TestTierSubroutinesAdjacency` -> `TestAdjacencyCoversItsTier/07_subroutines`. These commands pass once branch ccx/adj merges.

## Acceptance criteria

- [x] The dependency lint skips a `parsent` file rather than erroring on it, at ONE place all seventeen `lintOps` call sites route through rather than at each of them, and `opsOf` is left alone so a genuinely broken `parses` file still errors. (`go test ./internal/conformance/ -run 'TestCorpusLints' -count=1`)

- [x] The corpus carries its first `--- expect parsent` file: `conformance/07_subroutines/10_undeclared_callee.t`, the undeclared parenless callee, the fact `TestTierCallFormsParenlessExtent` asserts in Go because no corpus file could carry it. (`awk '/^## An undeclared callee is a syntax error$/{f=1;next} /^## /{f=0} f && /^parses: no$/{ok=1} END{exit !ok}' conformance/mdtest/argument-extent.md`)

- [x] Perl still ADJUDICATES that file's claim before anything else: the tier's own perl validator runs over it and would report a CORPUS BUG if perl accepted the source, and `TestPerlAdjudicatesFirst` still holds. (`go test ./internal/conformance/ -run 'TestTierSubroutinesPerlValidated|TestPerlAdjudicatesFirst' -count=1 -v 2>&1 | grep -E '10_undeclared_callee|TestPerlAdjudicatesFirst'`)

- [x] A `parses` file whose source perl genuinely refuses still ERRORS from the lint rather than linting vacuously -- the failure mode the caller-side skip exists to avoid. (`go test ./internal/conformance/ -run 'TestLintErrorsOnABrokenParsesFile' -count=1 -v`)

- [x] The tier's adjacency topic still emits every op tier 07 introduces, which `TestAdjacencyCoversItsTier` demands of every tier; a `parses: no` case compiles no ops, so it adds nothing the topic must cover. (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^07_subroutines$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/07_subroutines'`)

- [x] The ratchet records the new file and the whole conformance suite is green. (`go test ./internal/conformance/ -count=1`)
