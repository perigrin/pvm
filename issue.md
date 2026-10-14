---
title: "`die $object` expects Str"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:00:48.946214862Z
updated: 2026-10-14T17:00:48.946214862Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N3b-DIE-OBJECT`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 6 (6 added)  
**Likely fix site:** internal/parse/declarations/CORE.pmt `die` signature (`List[Str] @message` should admit any scalar: die throws objects)

perl: `eval { die E->new }` leaves the object in `$@`. Legacy fires too; the cutover adds 6 lines in CPAN.pm/Cmd.pm.

**Repro `n3b_die_object.pl`:**
```perl
package E; sub new { bless {}, shift }
package main;
eval { die E->new }; print ref($@), "\n";
```
perl -w (rc=0): E
old: 3:12: warning: call to die: argument 1 expects Str, got Object [type-mismatch];
new: 3:12: warning: call to die: argument 1 expects Str, got Object [type-mismatch];

Examples from the diff:

- `privlib/CPAN.pm:571` (added) `die CPAN::Exception::yaml_process_error->new($yaml_module,$local_file,"parse",$@);` -- call to die: argument 1 expects Str, got Object
- `privlib/CPAN.pm:581` (added) `die CPAN::Exception::yaml_process_error->new($yaml_module,$local_file,"parse",$@);` -- call to die: argument 1 expects Str, got Object
- `privlib/CPAN.pm:592` (added) `die CPAN::Exception::yaml_not_installed->new($yaml_module, $local_file, "parse");` -- call to die: argument 1 expects Str, got Object

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestDieTakesAnObject$' -count=1 -v | grep -q '^--- PASS: TestDieTakesAnObject'`)
