---
title: An object in an operator slot, though its class overloads the operator
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:00:47.844513167Z
updated: 2026-10-14T20:41:47.635867076Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T17:13:40.168902325Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:40.168902325Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N3-OVERLOAD-OBJECT`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 24 (24 added)  
**Likely fix site:** internal/infer/lowered_ops.go `checkOperand` (an Object operand is not checked against Num/Str: perl dispatches to overloading or numifies the address)

Math::BigInt `% *`, version `<= > >= <=>`, an overloaded `+`, `"" . Obj->new`, `abs(BigInt)`, `truncate`/`length` on objects. Legacy fires on some repros (n33, n34) but not on BigInt/version, whose lines the cutover adds.

**Repro `n3_overload_object.pl`:**
```perl
use Math::BigInt;
my $r = Math::BigInt->new(10) % Math::BigInt->new(3);
print "$r\n";
```
perl -w (rc=0): 1
old: 
new: 2:9: warning: left operand of "%": expected Num, got Object [coercion-mismatch];2:33: warning: right operand of "%": expected Num, got Object [coercion-mismatch];

**Repro `n3_version_compare.pl`:**
```perl
use version;
print version->new("1.2") <= version->new("1.3") ? "le\n" : "gt\n";
```
perl -w (rc=0): le
old: 
new: 2:7: warning: left operand of "<=": expected Num, got Object [coercion-mismatch];2:30: warning: right operand of "<=": expected Num, got Object [coercion-mismatch];

**Repro `n33_overload_plus.pl`:**
```perl
package N; use overload '+' => sub { 7 }, fallback => 1; sub new { bless {}, shift }
package main;
my $value = N->new; print 0 + $value, "\n";
```
perl -w (rc=0): 7
old: 3:31: warning: right operand of "+": expected Num, got Object [coercion-mismatch];
new: 3:31: warning: right operand of "+": expected Num, got Object [coercion-mismatch];

**Repro `n34_concat_method_eq.pl`:**
```perl
package A; use overload '""' => sub { "a" }; sub new { bless {}, shift }
package main;
print "" . A->new eq 'a' ? "ok\n" : "no\n";
```
perl -w (rc=0): ok
old: 3:12: warning: right operand of ".": expected Str, got Object [coercion-mismatch];
new: 3:12: warning: right operand of ".": expected Str, got Object [coercion-mismatch];

Examples from the diff:

- `privlib/CPAN.pm:880` (added) `truncate $fh, 0;` -- call to truncate: argument 1 expects Int|Num|Str|GlobRef|Glob|NaN|Inf|IO, got Object
- `privlib/ExtUtils/MakeMaker.pm:627` (added) `: version->new($required_version) > version->new($pr_version)` -- left operand of ">": expected Num, got Object
- `privlib/ExtUtils/MakeMaker.pm:627` (added) `: version->new($required_version) > version->new($pr_version)` -- right operand of ">": expected Num, got Object


**Ruled 2026-10-14:** perigrin chose the lossy opt-in. An object or reference in a numeric slot (`$h + 1`, `Math::BigInt->new(5) % 2`, `$version <= 3`) is the same lossy-coercion class as a reference in a Str slot (01a13793-4023): silent by default, a `lossy-coercion` warning under `--warn-lossy-coercion`. perl -w is silent on all of them, overloaded or not, and PSC cannot see most overloads (version's come from XS; the index does not read perl's lib). Goldens `ops/paren_operand` and `calls/method_new` move under trailers citing this ruling.

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestOverloadedObjectOperandIsSilent$' -count=1 -v | grep -q '^--- PASS: TestOverloadedObjectOperandIsSilent'`)
