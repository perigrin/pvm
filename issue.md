---
title: "`chr $u =~ /re/` argument typed List"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:01.328516143Z
updated: 2026-10-14T17:01:01.328516143Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N10-CHR-MATCH`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 1 (1 added)  
**Likely fix site:** internal/infer/lowered_special.go `match` (a match in scalar/argument context is Boolean, not List)

JSON::PP:898. Legacy fires on the repro too.

**Repro `n10_chr_match.pl`:**
```perl
my $u = 65;
print "ascii\n" if chr $u =~ /[[:^ascii:]]/;
```
perl -w (rc=0): ascii
old: 2:24: error: call to chr: argument 1 expects Int, got List [type-mismatch];
new: 2:24: error: call to chr: argument 1 expects Int, got List [type-mismatch];

Examples from the diff:

- `privlib/JSON/PP.pm:898` (added) `if ( chr $u =~ /[[:^ascii:]]/ ) {` -- call to chr: argument 1 expects Int, got List

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestChrOfBoundMatchIsScalar$' -count=1 -v | grep -q '^--- PASS: TestChrOfBoundMatchIsScalar'`)
