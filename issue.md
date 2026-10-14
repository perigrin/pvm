---
title: A file-scope `my $x = undef` is read as Undef inside a sub, ignoring assignments in other subs
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:00:54.5981833Z
updated: 2026-10-14T17:00:54.5981833Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N5b-FILE-UNDEF-IN-SUB`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 5 (5 added)  
**Likely fix site:** internal/infer/lowered.go `decl`/`collectLoweredDeclarations` (a file-scope variable assigned in any sub widens to the join of its assignments); possibly 01a132ab-7cf7-7af0-b284-c350971557ab

Sys::Syslog `$current_proto` (set in connect_log, read in syslog), Unicode::UCD `$UNICODEVERSION`. No standalone repro fires; real-file evidence: Syslog.pm:122 `my $current_proto = undef;`, :655 `$current_proto = $proto;` (in connect_log), :456 `if ($current_proto eq 'native')` draws "expected Str, got Undef".

**Repro `n5b_file_undef_in_sub.pl`:**
```perl
my $proto = undef;
sub connect_log { $proto = 'native' }
sub log_it { if ($proto eq 'native') { print "n\n" } } connect_log(); log_it();
```
perl -w (rc=0): n
old: 
new:

**Repro `n5_undef_init.pl`:**
```perl
my $proto = undef;
sub connect_log { $proto = 'native' }
sub log_it { connect_log(); print "n\n" if $proto eq 'native' } log_it();
```
perl -w (rc=0): n
old: 
new:

Examples from the diff:

- `privlib/Unicode/UCD.pm:4735` (added) `unless $UNICODEVERSION =~ /^\d+(?:\.\d+)+$/;` -- left operand of "=~": expected Str, got Undef
- `privlib/x86_64-linux-thread-multi/Sys/Syslog.pm:456` (added) `if ($current_proto eq 'native') {` -- left operand of "eq": expected Str, got Undef
- `privlib/x86_64-linux-thread-multi/Sys/Syslog.pm:459` (added) `elsif ($current_proto eq 'eventlog') {` -- left operand of "eq": expected Str, got Undef

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestFileScopeUndefReadInSubIsNotUndef$' -count=1 -v | grep -q '^--- PASS: TestFileScopeUndefReadInSubIsNotUndef'`)
