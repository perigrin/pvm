---
title: "`bless $self, $class` where $class is a package variable initialised undef"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:10.098063617Z
updated: 2026-10-14T17:01:10.098063617Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N12-BLESS-UNDEF-CLASS`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 2 (2 added)  
**Likely fix site:** same narrowing gap as N5/N5b (`... if defined $downgrade` / assignments elsewhere)

Math::BigInt `bless $self, $downgrade` (the call sites are guarded by `defined $downgrade`). Legacy is silent.

**Repro `n12_bless_undef_class.pl`:**
```perl
our $downgrade = undef;
sub f { my $self = bless {}, "main"; bless $self, $downgrade if defined $downgrade; $self }
f();
```
perl -w (rc=0): 
old: 
new: 2:51: error: call to bless: argument 2 expects Str, got Undef [type-mismatch];

Examples from the diff:

- `privlib/Math/BigInt.pm:737` (added) `bless $self, $downgrade;                  # bless into downgrade class` -- call to bless: argument 2 expects Str, got Undef
- `privlib/Math/BigInt.pm:773` (added) `bless $self, $upgrade;                    # bless into upgrade class` -- call to bless: argument 2 expects Str, got Undef

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestBlessClassFromAssignedPackageVar$' -count=1 -v | grep -q '^--- PASS: TestBlessClassFromAssignedPackageVar'`)
