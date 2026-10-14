---
title: A maybe-Inf value or a constant power still draws a mismatch in an Int slot
state: done
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T21:56:06.213388853Z
updated: 2026-10-14T23:36:59.460477298Z
sessions:
- start_sha: 85bda4f9b766ba76e39cccddbd73509b57462ebd
  end_sha: 6fed0273b415fdabaec853b07adab5c4df6f5598
  commits: 2
  started_at: 2026-10-14T22:36:28.961183293Z
  ended_at: 2026-10-14T23:36:59.460477298Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T22:36:28.961183293Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-14T23:36:59.460477298Z
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

- found by: cutover 1 (01a0c115-a44b), round-2 re-measure, bucket `R2-INF-INT-SLOT` (2026-10-14); evidence in `scratchpad/cmp-a44b/ev-r2.md`, `ev-r2b.md`, repros in `cmp-a44b/repro/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed before it merges, on the lowered path. 44 added lines (about 19 are paren-span moves of the same diagnostic). 01a135fc's finite rule reaches Num slots but not Int slots (`checkArg`/`checkOperand` for `splice`, `sysread`, `read`, `caller`, `localtime`, `rindex`, `ioctl`, `truncate`, and `<<`, `&`, `|`); and `sub PRIVATE () { 2**1 }; $x & PRIVATE` -- perl folds it to 2 -- is typed `Int|Num|NaN|Inf`. A Num judged finite should also satisfy an Int slot where perl truncates silently, as the Num repeat count already does.

#### r2_inf_int_slot

```perl
my @h = (1 .. 5); my $histsize = 2;
splice @h, 0, @h - $histsize if @h > $histsize;
my $x = 1 << (4 * 7);
```

```perl
my @h = (1 .. 5); my $histsize = 2;
splice @h, 0, @h - $histsize if @h > $histsize;
my $x = 1 << (4 * 7);
```
perl -w (rc=0): 
Deparse: (my(@h) = (1..5)); (my($histsize) = 2); ((@h > $histsize) and splice(@h, 0, (@h - $histsize))); (my($x) = 268435456); 
old: 3:15: warning: right operand of "<<": expected Int, got Int|Num|Inf [coercion-mismatch];
new: 2:15: warning: call to splice: argument 3 expects Int, got Int|Num|Inf [type-mismatch];3:14: warning: right operand of "<<": expected Int, got Int|Num|Inf [coercion-mismatch];
#### r2_inf_constant_bitop

```perl
use constant { PUBLIC => 1 << 0, PRIVATE => 1 << 1 };
my $fattr = 3;
print "private\n" if $fattr & PRIVATE;
```

```perl
use constant { PUBLIC => 1 << 0, PRIVATE => 1 << 1 };
my $fattr = 3;
print "private\n" if $fattr & PRIVATE;
```
perl -w (rc=0): private
old: 
new:
#### r2_const_pow_bitop

```perl
sub PRIVATE () { 2**1 }
my $battr = 3;
print "private\n" if $battr & PRIVATE;
```

```perl
sub PRIVATE () { 2**1 }
my $battr = 3;
print "private\n" if $battr & PRIVATE;
```
perl -w (rc=0): private
Deparse: sub PRIVATE () {     2.0; } (my($battr) = 3); (($battr & 2.0) and print("private
")); 
old: 
new: 3:31: warning: right operand of "&": expected Int, got Int|Num|NaN|Inf [coercion-mismatch];

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42 -w's verdict holds, keeping a control perl warns on; goldens under `Diagnostics-changed:` trailers; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestMaybeInfInIntSlotIsSilent$' -count=1 -v | grep -q '^--- PASS: TestMaybeInfInIntSlotIsSilent'`)
