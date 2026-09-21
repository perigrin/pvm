---
title: Tier 01 written against the finished tooling, which is what proves it works
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c35f-a487-7aa8-934b-4e68dd0fbaa1
- 01a0c35e-d2b6-7993-9059-5675239aea65
- 01a0c35e-66f9-71b2-b5f3-3dd0ef1cba05
- 01a0c35f-57cc-7a0f-83ac-53654f46f235
blocks:
- 01a0c361-6191-74b6-a1c0-263f58d465b3
created: 2026-09-21T09:50:44.159860455Z
updated: 2026-09-21T10:00:07.272858562Z
---

The tooling issues before this one are each separately checkable, and none
of them proves the tooling works together. Tier 01 does. It is the first
tier whose constructs are lexical, which means the token stream classifies
it and a declaration would be checking itself.

## What the spike established, and what it did not

Three files exist: `02_decimal.t` passes, `03_leading_decimal.t` and
`04_signed_exponent.t` refuse. Enough to show the format holds, nowhere near
enough to be a tier. Numbering starts at 02 because 01 was never written --
regenerate rather than preserve the gap.

## The finding this tier carries forward

`5e-1`, `5e+1`, `5E-1` and `1.5e-3` all split into three tokens; `5e1` is
fine. `perl -e 'print 5e'` is a syntax error, so `Number("5e")` is not a
token perl would ever produce.

**The parser cannot see this.** It gets `Number Operator Number`, reads
valid subtraction, and produces ZERO Unknown nodes. Behaviour cannot see it
either -- both print `0.5`. Nor can B::Concise, because perl never builds
the wrong tree to compare against. The token assertion is the only check
that catches it, which is the measured argument for the whole token layer
and belongs in this tier's README.

## Acceptance

- [ ] Tier 01 covers the numeric and string literal boundaries the glossary names, including v-strings and the -1-is-two-tokens case (`go test ./internal/conformance/ -run '^TestTierLiteralsCoversGlossary$' -count=1 -v | grep -q '^--- PASS: TestTierLiteralsCoversGlossary'`)
- [ ] Files are renumbered from 01 with no gap, by the regeneration the spec says the numbering supports (`go test ./internal/conformance/ -run '^TestTierNumberingRegenerates$' -count=1 -v | grep -q '^--- PASS: TestTierNumberingRegenerates'`)
- [ ] 03_leading_decimal.t's citation of 01a0c13f-97f5 resolves to a live issue (`git-zhi issue show 01a0c13f-97f5 >/dev/null`)
- [ ] Every file in 01_literals/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierLiteralsPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierLiteralsPerlValidated'`)
- [ ] The dependency lint passes over 01_literals/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierLiteralsLint$' -count=1 -v | grep -q '^--- PASS: TestTierLiteralsLint'`)
- [ ] 01_literals/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`test -s conformance/01_literals/README.md`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierLiteralsRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierLiteralsRefusalsCited'`)
- [ ] The whole corpus is green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
