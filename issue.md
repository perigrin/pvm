---
title: "Round-2 coverage and doc gaps: dualvar declares DualVar, IO::Socket::IP -register, Cwd::chdir a handle, the possible-mismatch flag in help, strict perl on every runner"
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-21T05:19:06.155781713Z
updated: 2026-10-21T05:19:23.604408724Z
---

crochet:review round 2 of PR #488 at 2db182e6 (2026-10-21). Quality report: scratchpad/review2-core-module-types-paad.md, probes scratchpad/r2probes/, psc builds of head, cf08c0ac and pu in scratchpad/r2bin/. A finding pu gets right is a regression this milestone introduced: fix it. A gap pu shares is out of scope here and goes to the follow-up named in the issue.

Coverage round 2: (1, Important) Scalar::Util::dualvar returns Str and its misfit is suppressed under done issue 01a15ac0-6ccc (quality I11 too; the .tpm comment is false since 30d02a6c) -- declare DualVar, delete the pendingAudit entry; (2) IO::Socket::IP's -register import option dropped (`use IO::Socket::IP -register` installs Socket's names) -- declare an unmarked import as Moose's; (3) Cwd::chdir($dirhandle) works (passes to CORE::chdir) but Cwd.tpm:28 takes Optional[Str] -- Str|FileHandle as POSIX::chdir; (4) help/types.md omits --warn-possible-mismatch (also quality S9); (5) PVM_REQUIRE_PERL542 stays off on macOS and Windows with no issue (also S1) -- turn it on, CI is green there; (6) RFC 'Possible mismatch' cites no sha (7fc02ddf). Quality S7 (Symbol delete_package dropped silently), S8 (2db182e6 says skipped cuts are in 6d24's issue; record them there), S10 (RFC Rejected alternatives contradicts skeleton mode), S12 (RFC does not cite the typed-our consumer issue 01a15ac0-c1d6).

## Steps
- [ ] Each fix measured with perl
- [ ] Record 6d24's four skipped cuts and their reasons in its issue body

## Acceptance criteria

### Positive Scenarios
- [ ] TestDualvarReturnsDualVar (`go test ./internal/infer/ -run '^TestDualvarReturnsDualVar$' -count=1 -v | grep -q '^--- PASS: TestDualvarReturnsDualVar'`)
- [ ] TestIOSocketIPRegisterInstallsSocketNames (`go test ./internal/parse/ -run '^TestIOSocketIPRegisterInstallsSocketNames$' -count=1 -v | grep -q '^--- PASS: TestIOSocketIPRegisterInstallsSocketNames'`)
- [ ] TestCwdChdirTakesAHandle (`go test ./internal/infer/ -run '^TestCwdChdirTakesAHandle$' -count=1 -v | grep -q '^--- PASS: TestCwdChdirTakesAHandle'`)
- [ ] help lists --warn-possible-mismatch (`grep -q 'warn-possible-mismatch' internal/data/help/types.md`)
- [ ] macOS and Windows require perl 5.42 (`! grep -q "matrix.os == 'Linux' && '1'" .github/workflows/ci.yml`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestPendingAuditHasNoDoneIssue (`go test ./internal/infer/ -run '^TestPendingAuditHasNoDoneIssue$' -count=1 -v | grep -q '^--- PASS: TestPendingAuditHasNoDoneIssue'`)
