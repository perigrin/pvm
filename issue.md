---
title: "`&name` is checked as the CORE builtin of that name"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:00:46.150855484Z
updated: 2026-10-14T17:00:46.150855484Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N2-AMP-BUILTIN`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 7 (7 added)  
**Likely fix site:** internal/infer/lowered_calls.go `call` (honour `Call.Ampersand`/`BareAmpersand`: `&name` never calls CORE)

User::grent/pwent/hostent/netent define `getgrgid` etc. and call them as `&getgrgid`, passing @_. perl calls the package sub (`user:7`). The lowered path checks CORE::getgrgid's arity. The repro also shows `print lookup(7)` drawing "got List" for a user sub whose body is `&name`.

**Repro `n2_amp_builtin.pl`:**
```perl
sub getgrgid { "user:@_" }
sub lookup { &getgrgid }
print lookup(7), "\n";
```
perl -w (rc=0): user:7
old: 
new: 2:14: error: call to getgrgid: expected at least 1 argument(s), got 0 [arity-mismatch];3:7: error: call to print: argument 1 expects Str, got List [type-mismatch];

**Repro `n38_exists_amp.pl`:**
```perl
sub getgrnam { "g:@_" } sub getgrgid { "i:@_" }
sub getgr :prototype($) { ($_[0] =~ /^\d+/) ? &getgrgid : &getgrnam }
print getgr("root"), "\n";
```
perl -w (rc=0): g:root
old: 
new: 2:47: error: call to getgrgid: expected at least 1 argument(s), got 0 [arity-mismatch];2:59: error: call to getgrnam: expected at least 1 argument(s), got 0 [arity-mismatch];3:7: error: call to print: argument 1 expects Str, got List [type-mismatch];

Examples from the diff:

- `privlib/Net/hostent.pm:56` (added) `&gethostbyaddr(Socket::inet_aton($addr));` -- call to gethostbyaddr: expected at least 2 argument(s), got 1
- `privlib/Net/netent.pm:48` (added) `&getnetbyaddr(Socket::inet_aton(shift));` -- call to getnetbyaddr: expected at least 2 argument(s), got 1
- `privlib/Net/netent.pm:50` (added) `&getnetbyname;` -- call to getnetbyname: expected at least 1 argument(s), got 0

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestAmpersandCallIsTheUserSub$' -count=1 -v | grep -q '^--- PASS: TestAmpersandCallIsTheUserSub'`)
