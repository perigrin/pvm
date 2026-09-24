---
title: Three .t headers state measured claims that are false
state: cancelled
urgency: normal
milestone: corpus-complete
created: 2026-09-24T05:53:42.371355094Z
updated: 2026-09-24T15:26:56.302923371Z
transitions:
- state: cancelled
  actor: human:git-zhi
  timestamp: 2026-09-24T15:26:56.302923371Z
---

Three `.t` headers stated measured claims that were false. Found by the
verify phase of the mdtest port, which measured every prose claim it
carried across rather than trusting it.

**CLOSED AS OBSOLETE.** The `.t` files were deleted in `a6b33a2a` when
the corpus moved to markdown topics, so the originals no longer exist.
The mdtest copies were repaired in `6105cc3d` at the time they were
found:

  - `reverse`/`aassign` targ numbers (t6/t4/t8 claimed, t4/t7/t5
    measured). The CONTEXT FLAG was always the claim and is correct;
    the targ is pad allocation and was never assertable, so the numbers
    were dropped rather than corrected.
  - "`*` occurred three times across 1,141 lines" -- 34 across 1,081
    when re-measured, and the original count included multiplication.
    Restated as the KIND (no `*` was the sigil) rather than the tally,
    with a note saying why.
  - "Three of the five cases emit an op" where four do; `<*.glob>`
    emits `glob` and was missed.

## What survives this issue

The pattern, which is recorded in the postmortem rather than here: two
of the three survived because they cited a NUMBER, which reads as
evidence. Token facts are mechanically checkable and header prose is
not, and a check that re-runs every `MEASURED` block in a case's prose
would have caught all three. The mdtest format makes that easier than
the old one did -- a case's program is already a fenced block a tool
can extract -- so it is a plausible thing to build rather than a wish.

Not filed as its own issue, because the port's verify phase already
demonstrated the cheaper version: measure the prose when you touch it.
