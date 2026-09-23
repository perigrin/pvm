---
title: Forty-nine negative token facts name text their source never contains
state: in-progress
urgency: normal
milestone: corpus-complete
created: 2026-09-23T19:16:11.345924198Z
updated: 2026-09-23T21:33:48.307460229Z
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

**Ten looked like they were not debt, and they were.** Deleting the
arrow negatives broke `TestTierOoArrowIsLexical`, which required one,
and I read that as evidence the negatives were load-bearing: the test
argues a lone negative is vacuous but that a tier holding BOTH a
negative and a positive is falsifiable. `pairedNegative` was built to
record and check that.

The gate measured it and it does not hold. Delete `"->"` from
`operators` in `internal/lexer/scan.go` -- a lexer that cannot tell an
arrow from a minus and a `>`, the exact case the exemption claimed the
pair would catch -- and ten files fail, every one on a POSITIVE fact.
No negative fails, and none can: no arrow-negative source contains the
byte pair `->`, so `checkTokenFact` can never count one.

Two tells were visible beforehand. `10_deref_at_sigil.t` does the same
job with no arrow fact, and `10_io/07_say.t` had its identical fact
deleted for want of a partner -- a rule that keeps a fact in tier 08
and deletes it in tier 10 is tracking which negatives another test
requires, not falsifiability.

So all ten are deleted too, `pairedNegative` is gone, and
`TestTierOoArrowIsLexical` keeps its positive requirement and asks the
SOURCE for the other half: some file must spell the call indirectly,
which is a claim about source text and not about tokens.

`knownVacuous` is deleted outright rather than left empty. It existed
so the check could land over pre-existing debt; the debt is repaired.

## Result

Counted inside `--- expect tokens` blocks only, negatives go 110 to 63:
47 removed. An earlier count of 115 to 76 came from grepping whole
files, which counts header prose.

## Acceptance criteria

- [x] Every negative token fact names text its own source contains, with no exemption list at all (`grep -Lq knownVacuous internal/conformance/vacuous_test.go`)
- [x] The vacuity check passes with no ratcheted entries (`go test ./internal/conformance/ -run '^TestNegativeFactsNameTextTheSourceContains$' -count=1 -v | grep -q '^--- PASS'`)
- [x] The whole conformance suite is green with the repairs in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
