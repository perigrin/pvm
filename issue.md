---
title: Triage statement-boundary re-emission differences in the M3 dirs
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:34.159970536Z
updated: 2026-10-22T23:47:16.33239884Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_defects.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^TestFaithfulStatementBoundaries$' -count=1 -v`
- entrypoints: `internal/parse/fidelity.go` (`matchForgivingSemicolons`), `internal/parse/canon.go` (block and statement emission)

These are the `stmt-boundary` rows of faithful_defects.txt, 12 M3-dir files where canon and source disagree on a `;`, a `}` or a block dereference's braces rather than on a paren. `;`/`}` placement: op/closure.t, op/goto-sub.t, op/goto.t, op/loopctl.t, op/stash.t, op/state.t, op/while.t, uni/stash.t. A source `{` where canon writes the variable (block deref `%{$a}` vs `%$a`): op/kvaslice.t, op/kvhslice.t, op/localref.t. A token-count mismatch (318 vs 369): lib/universal.t. Each is reduced to a minimal fixture and judged against perl. It is then fixed, or kept in faithful_defects.txt with its line and the perl measurement that explains it.

## Steps

- [ ] For each file, reduce the first difference to a minimal snippet and measure it with `perl -MO=Deparse,-p` or `perl -MO=Concise`. Note whether the tree is wrong (a parser defect), canon is wrong, or the difference is cosmetic.
- [ ] RED: add each reduced snippet to TestFaithfulStatementBoundaries, with the perl measurement in a comment. Run it and watch the cases fail.
- [ ] GREEN: fix one cause at a time, running the test after each. Do not widen the forgiven set beyond what the measurement shows is cosmetic.
- [ ] Remove the fixed rows from faithful_defects.txt. Give every kept row its line and reason, and re-baseline perlgittfaithful.ratchet in the same commit, naming each moved file.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus. Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Each reduced statement-boundary case re-emits faithfully or is pinned with its perl measurement (`go test ./internal/parse/ -run '^TestFaithfulStatementBoundaries$' -count=1 -v | grep -q '^--- PASS: TestFaithfulStatementBoundaries'`)
- [ ] The defects list and the M3-dir check agree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3Dirs$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3Dirs'`)
- [ ] The perl.git t/ faithful ratchet records the result (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
