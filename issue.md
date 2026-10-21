---
title: Select compares a slurpy's element type, so fileparse and utime can declare that they stringify a reference
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-21T22:51:29.804352309Z
updated: 2026-10-21T22:51:29.804352309Z
---

From the 01a16016-4587 rework (2026-10-21): `types.Select` compares a slurpy parameter's container type, not its element type, so a sub that takes a stray reference in a slurpy list cannot declare a stringifying candidate beside a handle one. `File::Basename::fileparse({})` and `Time::HiRes::utime(undef, undef, {})` are silent in perl 5.42 -w and give a default coercion error on 2db182e6, 692a560c and fix/4587 (pu silent: no declarations). Neither appears over perl 5.42's lib (TestPSCCheckLibBaseline passes).

## Steps
- [ ] Select compares a slurpy's element type against each flattened argument it absorbs
- [ ] Declare fileparse and utime with a stringifying candidate, measured

## Acceptance criteria

### Positive Scenarios
- [ ] A reference in a stringifying slurpy is a lossy coercion, silent by default (`go test ./internal/infer/ -run '^TestSlurpyStringifiesARef$' -count=1 -v | grep -q '^--- PASS: TestSlurpyStringifiesARef'`)

### Negative Scenarios
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
