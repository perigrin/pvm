---
title: Two incompatible file-numbering conventions, and nothing records which tier uses which
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T20:47:54.980994967Z
updated: 2026-09-21T22:34:37.38277674Z
sessions:
- start_sha: b41801021927f68fcbf799c60d6b947393d983de
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T22:29:10.669616329Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T22:29:10.669616329Z
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

Running `regenerateNumbering`'s logic over all fourteen tiers settles it.
EXACTLY ONE tier satisfies derived numbering:

    derived      01_literals
    semantic     02_variables 03_context 04_operators 05_scoping
                 06_control 07_subroutines 08_references 09_regex
                 10_io 11_oo 12_packages 13_opaque 14_recursive

So derived numbering is not a corpus convention that tier 06 violates. It
is a tier-01 property that thirteen tiers do not share, and asserting it
corpus-wide would renumber thirteen directories to satisfy a rule the spec
never states.

This issue's own claim that tier 02 was "renumbered to match" during
01a0c361-6191 is WRONG, and the census is what catches it. `git show
b4af7e51 -- conformance/02_variables/` shows the renames closed a GAP to
make room for a new file -- `05_hash.t` to `06_hash.t` and so on, shifting
a contiguous run up by one. It did not alphabetise: `01_array.t` still
precedes `02_array_element.t`, which a regeneration would reverse.

The spec is the deciding evidence. It says a file's identity is its NAME
and the number is its current position, and it does NOT say positions are
alphabetical. The ratchet is name-keyed with numeric prefixes stripped
precisely so that a renumbering is invisible, which is what makes a
semantic arrangement free to exist.

## The resolution

Both conventions are legitimate, and the tier DECLARES which it uses, in
the milestone's house pattern for a machine-read README block:

    ## FILE ORDER

        semantic

`derived` promises the numbers are a function of the names and is checked
by regeneration. `semantic` promises nothing a machine can check, and the
README prose beneath the block is where the reason lives -- which is the
honest shape, since no test can confirm that grouping conditionals before
loops is the RIGHT grouping.

`semantic` is not an opt-out. A tier claiming it whose files already sit in
the order a regeneration produces is told to claim `derived` instead, so
the word cannot be used to escape a check.

## Acceptance

- [ ] Every one of the fourteen tiers declares a `## FILE ORDER` block
      holding exactly `derived` or `semantic`; a tier that declares none
      fails by name.
      (`go test ./internal/conformance/ -run TestEveryTierDeclaresFileOrder -count=1 -v`)
- [ ] Every tier declaring `derived` has numbering `regenerateNumbering`
      reproduces exactly, and a tier whose numbering has a gap or a
      misordering is named along with the number a regeneration would give
      it.
      (`go test ./internal/conformance/ -run TestDerivedTierNumberingRegenerates -count=1 -v`)
- [ ] Every tier declaring `semantic` has an order a regeneration would NOT
      produce, so the declaration cannot be used to opt out of the derived
      check.
      (`go test ./internal/conformance/ -run TestSemanticTierNumberingIsNotAccidental -count=1 -v`)
- [ ] The census is reproducible from the tree: the declared orderings
      match what regeneration measures, one `derived` and thirteen
      `semantic`.
      (`grep -A2 '^## FILE ORDER' conformance/*/README.md | grep -E 'derived|semantic' | sort | uniq -c`)
- [ ] No tier was renumbered to satisfy the rule.
      (`git status --porcelain conformance/ | grep -E '^R|\.t$' ; test -z "$(git status --porcelain conformance/ | grep -E '\.t$')"`)
- [ ] The whole conformance suite is green.
      (`go test ./internal/conformance/ -count=1`)
