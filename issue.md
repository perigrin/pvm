---
title: Select compares a slurpy's element type, so fileparse and utime can declare that they stringify a reference
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-21T22:51:29.804352309Z
updated: 2026-10-21T22:59:10.827833294Z
---

From the 01a16016-4587 rework (2026-10-21): `types.Select` compares a slurpy parameter's container type, not its element type, so a sub that takes a stray reference in a slurpy list cannot declare a stringifying candidate beside a handle one. `File::Basename::fileparse({})` and `Time::HiRes::utime(undef, undef, {})` are silent in perl 5.42 -w and give a default coercion error on 2db182e6, 692a560c and fix/4587 (pu silent: no declarations). Neither appears over perl 5.42's lib (TestPSCCheckLibBaseline passes).

**Root cause corrected by the 01a16016-4587 re-review (2026-10-21):** the error comes from checkArg's IsCoercible(HashRef, Str|FileHandle), not from Select (fileparse's candidates share one slurpy, utime has one candidate). 4587 is fixing the Undef-union part (sockaddr_in, utime's numeric-string times) and fileparse's path parameter; what remains here is a reference in a Str|FileHandle slurpy element (utime's file list, fileparse's suffix list) that the operation stringifies.

## Steps
- [ ] Select compares a slurpy's element type against each flattened argument it absorbs
- [ ] Declare fileparse and utime with a stringifying candidate, measured

## Acceptance criteria

### Positive Scenarios
- [ ] A reference in a stringifying slurpy is a lossy coercion, silent by default (`go test ./internal/infer/ -run '^TestSlurpyStringifiesARef$' -count=1 -v | grep -q '^--- PASS: TestSlurpyStringifiesARef'`)

### Negative Scenarios
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
