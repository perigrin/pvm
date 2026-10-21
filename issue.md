---
title: perltest.Perl542 and conformance.findPerl still search for perl 5.42 differently
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-21T21:06:00.484242922Z
updated: 2026-10-21T21:06:00.484242922Z
---

crochet:review round 3 V17; destination given by 01a16016-481d.

Round 2's S19 asked that every test needing perl 5.42 find the same perl. 8652d0e8 shares only the version probe (`perltest.PerlVersion`). The search itself still differs:

- `perltest.Perl542` (internal/perltest/perltest.go) asks pvm's registry for 5.42.0, then PATH for a perl whose `$]` is 5.042000 (`perl542Version`).
- `conformance.findPerl` (internal/conformance/perl.go) takes `PVM_PERL`, then a fixed candidate list, checked against its own `measuredVersion = "5.042000"`.

So the T1 ratchet and the held-to-perl tests can run different perls, and the comment at perltest.go:92-95 ("never silently disagree") overstates what is shared.

## Steps
- [ ] One finder, one version constant, used by perltest and conformance
- [ ] A test that both packages resolve the same executable

## Acceptance criteria

### Positive Scenarios
- [ ] Both finders resolve the same perl (`go test ./internal/conformance/ -run '^TestFindPerlMatchesPerltest$' -count=1 -v | grep -q '^--- PASS: TestFindPerlMatchesPerltest'`)

### Negative Scenarios
- [ ] One 5.042000 constant outside tests (`test "$(grep -rln '"5.042000"' internal --include='*.go' | grep -v _test | wc -l)" -eq 1`)
