---
title: Refusal codes on the ten Unknown sites, so two refusals are distinguishable
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c35e-1b37-7304-ba39-65a6ef57d4e0
blocks:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
created: 2026-09-21T09:49:26.581612352Z
updated: 2026-09-21T19:27:48.988057764Z
sessions:
- start_sha: 9102578c35f74c5afa0bb9ec81e2bb8098a3221e
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T19:27:48.988057764Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T19:27:48.988057764Z
---

The parser is TOTAL: `parse.Parse(src)` returns one value and no error. A
refusal is an `Unknown` node carrying its text. That is the right design and
it has one consequence the corpus has to fix.

**Every refusal currently looks the same.** A corpus file that refuses can
say "1 Unknown node(s)" and nothing else, so two files refusing for
unrelated reasons are indistinguishable, and a refusal that CHANGES CAUSE
while staying a refusal is invisible.

## What a code buys

`04_signed_exponent.t` is the case. `5e-1` lexes as
`Number("5e") Operator("-") Number("1")` -- the parser sees valid
subtraction and produces ZERO Unknown nodes. The refusal is real but the
parser cannot report it; only the token assertion catches it. A file like
that needs to say WHICH refusal it waits on, and today it says "this file".

## Acceptance

- [ ] Two different refusals produce different codes (`go test ./internal/parse/ -run '^TestRefusalCodesDistinct$' -count=1 -v | grep -q '^--- PASS: TestRefusalCodesDistinct'`)
- [ ] Each of the ten Unknown construction sites carries a distinct stable code, not a message (`go test ./internal/parse/ -run '^TestEveryUnknownSiteHasACode$' -count=1 -v | grep -q '^--- PASS: TestEveryUnknownSiteHasACode'`)
- [ ] A file refusing with a different code than it names FAILS rather than skips, because the cause changed (`go test ./internal/conformance/ -run '^TestRefusalCodeMismatchFails$' -count=1 -v | grep -q '^--- PASS: TestRefusalCodeMismatchFails'`)
- [ ] The ten Unknown sites and their codes are enumerated where a reader finds them without grepping (`go test ./internal/parse/ -run '^TestRefusalCodeInventory$' -count=1 -v | grep -q '^--- PASS: TestRefusalCodeInventory'`)
