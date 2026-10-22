---
title: Re-emit heredoc bodies and sub-less special subs faithfully
state: pending
urgency: normal
milestone: m3-oracle-at-scale
created: 2026-10-22T23:46:34.069628305Z
updated: 2026-10-22T23:46:34.069628305Z
---

## Context

- paths: `internal/parse/fidelity.go`, `internal/parse/canon.go`, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_defects.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^(TestFaithfulHeredocBodyIsNotAToken|TestFaithfulSublessSpecialSub)$' -count=1 -v`
- entrypoints: `internal/parse/fidelity.go` (significant-token extraction), `internal/parse/canon.go` (sub declarations)

Two token-level re-emission causes, 10 M3-dir files, rows `heredoc` and `subless-special-sub` in faithful_defects.txt. (1) Heredoc bodies show up as significant tokens on the source side: op/debug.t, op/hook/require.t, op/lc.t, op/lex.t, op/threads.t. A heredoc body is a string literal's content, and the two sides must tokenize it alike. (2) canon writes `sub` where the source has `DESTROY` without it: op/hash.t, op/pos.t, op/ref.t, op/sub.t, op/substr.t. perl accepts `sub`-less declarations for some names, and whether canon should write `sub` there (a forgiven difference) or write what the source wrote is decided by measuring perl. Neither cause is a grouping question, so neither may weaken misgrouping detection.

## Steps

- [ ] Measure with perl first. Run `perl -MO=Deparse` and `perl -c` on `DESTROY { 1 }`, `AUTOLOAD { 1 }`, `BEGIN { 1 }` and `foo { 1 }` (the last should not be a declaration), and on a heredoc inside a call's argument list. Quote the results in the tests' comments.
- [ ] RED: write TestFaithfulHeredocBodyIsNotAToken, minimal fixtures reduced from op/debug.t and op/threads.t. Run it and watch it fail.
- [ ] GREEN: fix the token extraction on whichever side is wrong. Run the test and watch it pass.
- [ ] RED: write TestFaithfulSublessSpecialSub for `DESTROY {...}` and the other names perl accepted above. Run it and watch it fail.
- [ ] GREEN: fix it, either as a forgiven difference limited to exactly the names perl accepts or in canon, whichever the measurement supports. Run the test and watch it pass.
- [ ] Remove the fixed rows from faithful_defects.txt and re-baseline perlgittfaithful.ratchet in the same commit, naming each moved file in the message.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus. Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] A heredoc body is compared as one literal, not as source tokens (`go test ./internal/parse/ -run '^TestFaithfulHeredocBodyIsNotAToken$' -count=1 -v | grep -q '^--- PASS: TestFaithfulHeredocBodyIsNotAToken'`)
- [ ] A sub-less special-sub declaration re-emits faithfully, for exactly the names perl accepts (`go test ./internal/parse/ -run '^TestFaithfulSublessSpecialSub$' -count=1 -v | grep -q '^--- PASS: TestFaithfulSublessSpecialSub'`)
- [ ] The defects list and the M3-dir check agree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3Dirs$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3Dirs'`)
- [ ] The perl.git t/ faithful ratchet records the gain (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
