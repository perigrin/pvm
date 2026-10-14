---
title: The else branch of `if (ref $x)` keeps Undef and fails a Str slot
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T21:56:09.258973289Z
updated: 2026-10-14T21:56:09.258973289Z
---

## Context

- found by: cutover 1 (01a0c115-a44b), round-2 re-measure, bucket `R2-REF-ELSE-UNION` (2026-10-14); evidence in `scratchpad/cmp-a44b/ev-r2.md`, `ev-r2b.md`, repros in `cmp-a44b/repro/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed before it merges, on the lowered path. 3 added lines (B::Concise, B::Showlex). In the `else` of `if (ref $objname)`, `$objname` is reported `got Undef|Boolean|...` in a Str slot; perl is silent. A negated `ref` guard should leave the non-reference part of the type, as the positive guard leaves the reference part.

#### r2_ref_branch_objname

```perl
my @args = ("Foo", \"x");
for my $objname (@args) {
    if (ref $objname) { $objname = "main::" . $objname unless $objname =~ /::/; }
}
```

```perl
my @args = ("Foo", \"x");
for my $objname (@args) {
    if (ref $objname) { $objname = "main::" . $objname unless $objname =~ /::/; }
}
```
perl -w (rc=0): 
old: 3:47: error: right operand of ".": expected Str, got Ref [coercion-mismatch];
new:
#### r2_ref_else_objname

```perl
my @args = ("Foo", "x::y");
for my $objname (@args) {
    if (ref $objname) { } else { $objname = "main::" . $objname unless $objname =~ /::/; }
}
```

```perl
my @args = ("Foo", "x::y");
for my $objname (@args) {
    if (ref $objname) { } else { $objname = "main::" . $objname unless $objname =~ /::/; }
}
```
perl -w (rc=0): 
old: 3:56: error: right operand of ".": expected Str, got Undef|Boolean|Int|Num|Str|DualVar|NaN|Inf [coercion-mismatch];
new: 3:72: error: left operand of "=~": expected Str, got Undef|Boolean|Int|Num|Str|DualVar|NaN|Inf [coercion-mismatch];3:56: error: right operand of ".": expected Str, got Undef|Boolean|Int|Num|Str|DualVar|NaN|Inf [coercion-mismatch];

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42 -w's verdict holds, keeping a control perl warns on; goldens under `Diagnostics-changed:` trailers; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestNegatedRefGuardNarrowsToNonRef$' -count=1 -v | grep -q '^--- PASS: TestNegatedRefGuardNarrowsToNonRef'`)
