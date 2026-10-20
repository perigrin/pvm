---
title: "Hold the libraries to perl in CI: find the runner's perl 5.42.0, fail when a perl-held test skips, derive the hold lists from the declarations"
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.305509522Z
updated: 2026-10-20T19:18:42.305509522Z
---

crochet:review of PR #488, round 1 (2026-10-21). Quality I4 and coverage finding 1: perltest.Perl542 (internal/perltest/perltest.go:34) finds only a pvm-installed perl; CI installs perl 5.42.0 with actions-setup-perl (and Windows ships Strawberry 5.42), so every Test*Prototypes/Exports/ReturnTypes, TestCoreModuleRanking and TestPSCCheckLibBaseline SKIPs in CI (job 118774701699) and green CI says nothing about goals 3-5. Also: coverage 10 (hold batch lists are hand-kept, not tied to the ranking/declarations), quality S10 (lib_baseline.txt:13 has an archname path), S13 and ponytail (batch lists duplicated in internal/parse and internal/infer, two IgnoresTheCallersPerlLib tests, oracleFixture/perlFixture). Tests may run perl; production never does.

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

## Steps
- [ ] Perl542 also accepts a perl on PATH that reports version 5.042000 (tests only)
- [ ] PVM_REQUIRE_PERL542=1 turns Perl542's skip into a failure; the Linux ratchets job sets it
- [ ] Both packages derive their hold lists from the shipped declarations (one shared list in internal/perltest)
- [ ] Make the perl-held tests pass on macOS and Windows, or skip there with a measured reason as TestCoreModuleRanking does
- [ ] Fold the duplicate PERL5LIB tests and fixtures into internal/perltest

## Acceptance criteria

### Positive Scenarios
- [ ] TestPerl542FindsPerlOnPath (`go test ./internal/perltest/ -run '^TestPerl542FindsPerlOnPath$' -count=1 -v | grep -q '^--- PASS: TestPerl542FindsPerlOnPath'`)
- [ ] TestHoldListsCoverEveryDeclaration (`go test ./internal/parse/ -run '^TestHoldListsCoverEveryDeclaration$' -count=1 -v | grep -q '^--- PASS: TestHoldListsCoverEveryDeclaration'`)
- [ ] CI requires perl 5.42.0 for the held tests (`grep -q 'PVM_REQUIRE_PERL542' .github/workflows/ci.yml`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestPerl542FailsWhenRequired (`go test ./internal/perltest/ -run '^TestPerl542FailsWhenRequired$' -count=1 -v | grep -q '^--- PASS: TestPerl542FailsWhenRequired'`)
- [ ] TestPerl542SkipsAnotherVersion (`go test ./internal/perltest/ -run '^TestPerl542SkipsAnotherVersion$' -count=1 -v | grep -q '^--- PASS: TestPerl542SkipsAnotherVersion'`)
