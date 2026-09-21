---
title: Two incompatible file-numbering conventions, and nothing records which tier uses which
state: pending
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T20:47:54.980994967Z
updated: 2026-09-21T20:47:54.980994967Z
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
