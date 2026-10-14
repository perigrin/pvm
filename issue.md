---
title: fileno of a reference to a scalar that holds a glob is rejected
state: done
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T21:56:11.297071782Z
updated: 2026-10-14T23:37:05.229493895Z
sessions:
- start_sha: 85bda4f9b766ba76e39cccddbd73509b57462ebd
  end_sha: 6fed0273b415fdabaec853b07adab5c4df6f5598
  commits: 2
  started_at: 2026-10-14T22:36:29.912439882Z
  ended_at: 2026-10-14T23:37:05.229493895Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T22:36:29.912439882Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-14T23:37:05.229493895Z
observed_paths:
- internal/infer/infer.go
- internal/infer/infer_test.go
- internal/infer/lowered_calls.go
- internal/infer/lowered_flow.go
- internal/infer/lowered_ops.go
- internal/infer/lowered_r2_test.go
- internal/infer/lowered_special.go
- internal/infer/testdata/lowered/calls/chr_num.golden
- internal/lexer/brackets.go
- internal/lexer/expect.go
- internal/lexer/lexer.go
- internal/lexer/scan.go
- internal/parse/globpunct_test.go
- internal/parse/parse.go
- internal/parse/term.go
- internal/types/narrowing.go
- internal/types/referent_test.go
---

## Context

- found by: cutover 1 (01a0c115-a44b), round-2 re-measure, bucket `R2-FILENO-GLOB-IN-SCALAR` (2026-10-14); evidence in `scratchpad/cmp-a44b/ev-r2.md`, `ev-r2b.md`, repros in `cmp-a44b/repro/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed before it merges, on the lowered path. 2 added lines (Dumpvalue, dumpvar.pl). `fileno(\$v)` where `$v` holds a glob: `expects Glob, got ScalarRef`; perl is silent. 01a138e3-972f's handle set does not include a ScalarRef, which perl dereferences to the glob it holds.

#### r2_fileno_scalarref

```perl
my $v = \*STDOUT; my $fileno;
if (ref $v eq 'HASH') { 1 } elsif (defined ($fileno = fileno(\$v))) { print "fd\n" }
```

```perl
my $v = "x"; my $fileno;
print "fd\n" if defined($fileno = eval { fileno(\$v) });
```
perl -w (rc=0): 
old: 2:49: error: call to fileno: argument 1 expects Glob, got ScalarRef [type-mismatch];
new: 2:49: error: call to fileno: argument 1 expects Glob, got ScalarRef [type-mismatch];

```perl
my $v = \*STDOUT; my $fileno;
if (ref $v eq 'HASH') { 1 } elsif (defined ($fileno = fileno(\$v))) { print "fd\n" }
```
perl -w (rc=255): Not a GLOB reference at r2_fileno_scalarref.pl line 2.
old: 2:62: error: call to fileno: argument 1 expects Glob, got ScalarRef [type-mismatch];
new: 2:62: error: call to fileno: argument 1 expects Glob, got ScalarRef [type-mismatch];


**Corrected 2026-10-14.** The repro above is wrong: with `$v = \*STDOUT`, `fileno(\$v)` makes perl die "Not a GLOB reference". perl succeeds only when `$v` holds a glob VALUE (`my $v = *STDOUT`), and both real sites (Dumpvalue.pm:294, dumpvar.pl:315) are inside `elsif (ref \$v eq 'GLOB')`. So the fix is flow narrowing, not a wider handle slot: under a `ref \$v eq 'GLOB'` guard, `$v` is a Glob and `\$v` a GlobRef, which the handle slot already accepts. Outside such a guard a ScalarRef stays reported, because perl dies there. perl 5.42: `my $v = *STDOUT; if (ref \$v eq "GLOB") { print fileno(\$v) }` prints 1.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42 -w's verdict holds, keeping a control perl warns on; goldens under `Diagnostics-changed:` trailers; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestFilenoOfScalarRefToGlob$' -count=1 -v | grep -q '^--- PASS: TestFilenoOfScalarRefToGlob'`)
