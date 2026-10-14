---
title: An array argument to a list operator is checked as one Str
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:00:50.508117584Z
updated: 2026-10-14T17:00:50.508117584Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N4-ARRAY-LISTOP`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 18 (1 moved (message), 17 added)  
**Likely fix site:** internal/infer/lowered_calls.go `args`/`checkArg` (an Array/Hash in a slurpy LIST slot flattens; check its element type, not the container)

`print @parts`, `print $fh @x`, `print "x", @a`, `warn @code`, `exec @_`, `system(@cmd)`, `system 1, @_`, `chomp(@$arr)`. Pre-existing on legacy (repro), added where the lowered path reaches the line.

**Repro `n4_array_listop.pl`:**
```perl
my @parts = ("a", "b");
print @parts; print "x", @parts, "\n";
warn @parts;
```
perl -w (rc=0): ab at n4_array_listop.pl line 3. abxab
old: 2:7: error: call to print: argument 1 expects Str, got Array [type-mismatch];3:6: error: call to warn: argument 1 expects Str, got Array [type-mismatch];
new: 2:7: error: call to print: argument 1 expects Str, got Array [type-mismatch];2:26: error: call to print: argument 2 expects Str, got Array [type-mismatch];3:6: error: call to warn: argument 1 expects Str, got Array [type-mismatch];

**Repro `n4_array_system.pl`:**
```perl
sub run { my @cmd = @_; system(@cmd) == 0 or die "failed"; exec @cmd if 0 }
run("true");
```
perl -w (rc=0): 
old: 1:32: error: call to system: argument 1 expects Str, got Array [type-mismatch];1:65: error: call to exec: argument 1 expects Str, got Array [type-mismatch];
new: 1:32: error: call to system: argument 1 expects Str, got Array [type-mismatch];1:65: error: call to exec: argument 1 expects Str, got Array [type-mismatch];

**Repro `n4_chomp_aref.pl`:**
```perl
my $arr = ["a\n"];
chomp(@$arr);
```
perl -w (rc=0): 
old: 2:7: error: call to chomp: argument 1 expects Str, got Array [type-mismatch];
new: 2:7: error: call to chomp: argument 1 expects Str, got Array [type-mismatch];

Examples from the diff:

- `privlib/CPAN/Tarzip.pm:451` (slotmsg) `return system(@system) == 0;` -- call to system: argument 1 expects Str, got Boolean
- `privlib/CPAN/Distribution.pm:1381` (added) `system(@cmd) == 0 or` -- call to system: argument 1 expects Str, got Array
- `privlib/CPAN/Distribution.pm:4536` (added) `unless ( ($wstatus = system(@args)) == 0 ) {` -- call to system: argument 1 expects Str, got Array

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestArrayArgumentToListOpFlattens$' -count=1 -v | grep -q '^--- PASS: TestArrayArgumentToListOpFlattens'`)
