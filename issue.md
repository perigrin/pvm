---
title: Narrow the macOS and Windows skips in the perl-held tests to the names that differ there
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-20T21:05:01.164297765Z
updated: 2026-10-20T21:26:12.719509525Z
---

Found 2026-10-21 merging fix/6b81b (crochet:review round 1 of core-module-types). CI run 39489951290 ran the perl-held tests on macOS and Windows for the first time. 6b81b fixed the harness bugs (CRLF in return_oracle.pl and two Go parsers, the macOS /tmp symlink, build-specific counts) but skipped twelve tests WHOLE on macOS/Windows where only some names differ: e.g. TestPurePerlExportersReturnTypes/Prototypes/Exports skip all ten pure-Perl exporters off Linux because Errno's Linux-only codes (EADV ...) are absent; TestPOSIXReturnTypes skips every POSIX sub on Windows because one case calls Fcntl::F_GETFD. Linux holds everything (PVM_REQUIRE_PERL542=1), so goal 4 holds; but off Linux these tests now gate nothing.

## Steps
- [ ] Replace each whole-test platform skip with a per-OS map of the names measured absent or different there (cited from the CI log), so every other name in the test still runs
- [ ] An exemption naming a name that is in fact present on that OS fails, so the map cannot go stale

## Acceptance criteria

### Positive Scenarios
- [ ] No perl-held test skips whole on macOS or Windows (`go test ./internal/perltest/ -run '^TestNoWholeTestPlatformSkips$' -count=1 -v | grep -q '^--- PASS: TestNoWholeTestPlatformSkips'`)

### Negative Scenarios
- [ ] A platform exemption for a name perl defines fails (`go test ./internal/perltest/ -run '^TestStalePlatformExemptionFails$' -count=1 -v | grep -q '^--- PASS: TestStalePlatformExemptionFails'`)
