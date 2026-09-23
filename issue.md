---
title: Forty-nine negative token facts name text their source never contains
state: in-progress
urgency: normal
milestone: corpus-complete
created: 2026-09-23T19:16:11.345924198Z
updated: 2026-09-23T20:53:04.907782179Z
sessions:
- start_sha: 66e6d10dfed49b493965299ae1c5d6fcc9088c25
  end_sha: ""
  commits: 0
  started_at: 2026-09-23T20:25:52.910408381Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T20:25:52.910408381Z
---

Forty-nine of the corpus's 115 negative token facts named text their own
source never contains, so none of them could fail. Found by
`TestNegativeFactsNameTextTheSourceContains`, which landed with them
ratcheted in `knownVacuous` so the check could arrive without going red.

## The measurement

`checkTokenFact` compares `src[tk.Start:tk.End]` against the claimed
text. So a token whose text is `X` requires the bytes `X` somewhere in
the source. A source not containing those bytes can produce no such
token under ANY lexing, correct or broken -- the count is always zero,
a `no` fact always passes, and it holds nothing up.

## What the repair found: the 49 are not one bug

The issue predicted three repair shapes. Working through them
one at a time split the 49 differently, and the split is the result.

**Thirty-nine were debt and are deleted.** Three rationales, each
plausible and each mistaking the PARSER's behaviour for the LEXER's:

- The sibling-builtin family -- `no word whose text is "unshift"` in a
  file about `push`. `14_push.t` argued a keyword table mapping one
  onto the other "would spell it the other way here", but `scanWord`
  emits `Word` by slicing `l.src`, so a Word token's text is always
  literal source bytes.
- The absent-operator family -- `no operator whose text is "**"` in a
  precedence file writing only `+` and `*`.
- The version-gate family. Four files shared a passage arguing their
  source "contains no arrow anywhere -- so a lexer that produced one
  would have manufactured it". The arrow came from perl's DEPARSE of
  the ungated reading, written as if it were a claim about tokens, and
  it spread by citation with each file naming the others as models.

**Twelve would have been left with no fact at all**, so each got a
reachable replacement. Seven first drafts had the wrong count and the
runner caught every one; the final facts are measured off its token
dumps. `${$sr}` is four tokens, not one variable.

**Ten are not debt.** `TestTierOoArrowIsLexical` had already diagnosed
the same vacuity and solved it better than deletion: it requires the
TIER to hold both a file declaring no arrow and one declaring an arrow.
Deleting those negatives broke that test, which is how the distinction
surfaced. `pairedNegative` records the exemption and CHECKS it -- an
entry is excused only while a positive partner exists in the same tier.

## Result

Negative facts go 115 to 76. `knownVacuous` is empty.

## Acceptance criteria

- [x] Every negative token fact names text its own source contains (`grep -q 'knownVacuous = map\[string\]bool{}' internal/conformance/vacuous_test.go`)
- [x] The vacuity check passes with no ratcheted entries (`go test ./internal/conformance/ -run '^TestNegativeFactsNameTextTheSourceContains$' -count=1 -v | grep -q '^--- PASS'`)
- [x] The whole conformance suite is green with the repairs in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
