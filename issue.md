---
title: Two incompatible file-numbering conventions, and nothing records which tier uses which
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T20:47:54.980994967Z
updated: 2026-10-23T23:40:18.315090293Z
sessions:
- start_sha: b41801021927f68fcbf799c60d6b947393d983de
  end_sha: a3a0ed0c97ce48a68f5db427979f9d8acf88d3c8
  commits: 1
  started_at: 2026-09-21T22:29:10.669616329Z
  ended_at: 2026-09-21T23:31:52.826625642Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T22:29:10.669616329Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T23:31:52.826625642Z
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

Tier 01's `TestTierNumberingRegenerates` asserts that a tier's file numbers
can be DERIVED from its file names -- `regenerateNumbering` sorts names and
renumbers from 01 with no gap. That test was written for issue
01a0c360-01bf and passes for tier 01.

It does not generalise, and tier 06 proved it by measurement.

Tier 06's numbering is SEMANTIC, not alphabetical:

    01-03   conditionals   (if, unless, if/else)
    04-07   loops          (while, until, C-style for, foreach)
    08-10   jumps          (next/last/redo, postfix while, goto)

`conformance/06_control/README.md`'s adjacency reasoning depends on that
arrangement. Its agent wrote the mirrored test, watched it fail, and
correctly DELETED the test rather than renumber a corpus whose order carries
meaning. Numbering was not among that issue's acceptance criteria, so
asserting alphabetical order would have been scope the issue did not ask for
and a contradiction of the tier's deliberate layout.

So the corpus now holds two incompatible numbering conventions and nothing
records which tiers use which:

    derived      tier 01 (and tier 02, renumbered to match during
                 01a0c361-6191)
    semantic     tier 06, and probably others not yet examined

THE QUESTION TO SETTLE: is derived numbering a corpus-wide rule that tier 06
violates, or is semantic grouping legitimate and tier 01's test simply
tier-specific?

Evidence for semantic being legitimate: the spec (lines ~154-165) says a
file's identity is its NAME and the number is its current position. It does
NOT say positions must be alphabetical. The ratchet is name-keyed with
numeric prefixes stripped precisely so renumbering is invisible.

Evidence for derived being the rule: `regenerateNumbering` exists, tier 01's
AC demanded it, and a tier whose numbering is hand-maintained can drift
without anything noticing -- which is the failure mode the derived test was
written to prevent.

Whichever way it goes, the outcome should be checkable corpus-wide rather
than per-tier, since the current state is two conventions and no record.
## The measurement

Running `regenerateNumbering`'s logic over all fourteen tiers, then reading
the READMEs of the ones it rejects, gives three populations rather than two:

    derived      01_literals 02_variables
    grouped      06_control
    accidental   03_context 04_operators 05_scoping 07_subroutines
                 08_references 09_regex 10_io 11_oo 12_packages
                 13_opaque 14_recursive

So derived numbering is not a corpus-wide rule that tier 06 violates. It is
a property two tiers have and twelve do not, and asserting it corpus-wide
would renumber twelve directories to satisfy a rule the spec never states.

But the twelve are not tier 06 either. Tier 06 ARGUES for its order -- the
conditionals/loops/jumps runs are what its `and`/`or`/`cond_expr` and
`enterloop`/`enteriter` passages compare between neighbours. The other
eleven cite their files by NAME and never by position, and several
contradict any grouping reading outright: `13_opaque` pairs
`06_format_write.t` with `09_format_picture.t` under one construct, and
`08_references` pairs 05 with 07 and 06 with 10. Their numbers are the
order the files were written in.

A measurement correction worth recording, because the check caught it and a
hand-rolled census did not. A first pass over the corpus with a shell
script reported tier 02 as NOT derived, which would have made this issue's
own claim about 01a0c361-6191 wrong. The Go check disagreed, and the Go
check was right: `regenerateNumbering` sorts the identity WITH its
extension, so `array.t` sorts before `array_element.t` (`.` is 0x2E, `_` is
0x5F) and `01_array.t` legitimately precedes `02_array_element.t`. The
shell script stripped the extension and reversed them. The issue's claim
stands; the reimplementation was the thing that drifted, which is the
argument for the check living in one place.

The spec is the deciding evidence for allowing more than one convention. It
says a file's identity is its NAME and the number is its current position,
and does NOT say positions are alphabetical. The ratchet is name-keyed with
numeric prefixes stripped precisely so a renumbering is invisible, which is
what makes a chosen arrangement free to exist.

## The resolution

All three states are legitimate, and the tier DECLARES which it uses, in
the milestone's house pattern for a machine-read README block:

    ## FILE ORDER

        derived

    ## FILE ORDER

        01-03	conditionals
        04-07	loops
        08-10	jumps

    ## FILE ORDER

        accidental

The three spellings are deliberately asymmetric in cost, which is what
keeps them honest.

`derived` is one word and buys a machine check: the numbers are thrown away
and rebuilt from the names.

A grouping must NAME its ranges, and they must tile the tier's construct
files exactly -- start at 01, run contiguously, end at the last file. A
tier whose order is an accident cannot write that down without the
falsehood being visible to a reader, and a tier that gains or loses a file
is told immediately rather than drifting.

`accidental` is one word and buys NOTHING. That is the point: it records
debt rather than blessing it. A second word meaning merely "not derived"
would be a rubber stamp eleven tiers would reach for, so instead the word
asserts nothing and a tier has an incentive to say something truer. It is
kept honest from the other side too -- a tier declaring `accidental` whose
files ALREADY sit in regenerated order is told to declare `derived`, so the
word cannot be used to opt out of a check.

Renumbering the eleven into `derived` order is the right end state and
costs nothing on disk, but it regenerates the ratchet, which is another
issue's work. The declaration is what stops that from being rediscovered by
the next agent whose numbering test fails.

Criteria updated 2026-10-23: a6b33a2a replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestDerivedTierNumberingRegenerates` -> `TestDeclaredOrderingIsStillMeaningful`; `TestAccidentalTierNumberingIsNotDerivable` -> `TestAccidentalTierClaimsNoArrangement`; `TestGroupedTierNumberingCoversItsFiles` -> `TestGroupedTierNamesItsTopics`).

Criteria updated 2026-10-23 (review round 1, agent:ccx-rev1): the Grouped criterion's text now says what `TestGroupedTierNamesItsTopics` proves over topics, which carry no numbers to tile.

## Acceptance criteria

- [ ] Every one of the fourteen tiers declares a `## FILE ORDER` block holding `derived`, `accidental`, or `NN-NN name` grouping lines; a tier that declares none fails by name. (`go test ./internal/conformance/ -run TestEveryTierDeclaresFileOrder -count=1 -v`)
- [ ] Every tier declaring `derived` has numbering `regenerateNumbering` reproduces exactly, and a tier whose numbering has a gap or a misordering is named along with the number a regeneration would give it. (`go test ./internal/conformance/ -run '^TestDeclaredOrderingIsStillMeaningful$' -count=1 -v | grep -q '^--- PASS: TestDeclaredOrderingIsStillMeaningful'`)
- [ ] Every tier declaring `accidental` has an order a regeneration would NOT produce, so the word cannot be used to opt out of the derived check. (`go test ./internal/conformance/ -run '^TestAccidentalTierClaimsNoArrangement$' -count=1 -v | grep -q '^--- PASS: TestAccidentalTierClaimsNoArrangement'`)
- [ ] Every tier declaring a grouping has groups that cover its topics exactly -- every group names a topic the tier holds, and every topic but the adjacency topic is named by exactly one group, so no gap and no overlap -- so a tier that gains, loses or renames a topic is named rather than drifting. (`go test ./internal/conformance/ -run '^TestGroupedTierNamesItsTopics$' -count=1 -v | grep -q '^--- PASS: TestGroupedTierNamesItsTopics'`)
- [ ] The census is readable straight off the tree: two `derived`, one grouping, eleven `accidental`, fourteen lines in all. (`grep -A2 '^## FILE ORDER' conformance/*/README.md | grep -cE 'derived|accidental|[0-9][0-9]-[0-9][0-9]'`)
- [ ] No tier was renumbered to satisfy the rule; only READMEs changed. (`test 0 -eq "$(git status --porcelain conformance/ | grep -c '^R')"`)
- [ ] The whole conformance suite is green. (`go test ./internal/conformance/ -count=1`)
