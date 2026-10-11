---
title: Bucket not_a_term's 1,584 nodes by span, and test each candidate in isolation
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T16:26:40.127300233Z
updated: 2026-10-11T15:16:53.284189619Z
sessions:
- start_sha: 2dc301fdeeec2303d2bfcdc1cae708b09b190c0b
  end_sha: 9325864f86f42d125118a800b15eb0b62d95d6c3
  commits: 3
  started_at: 2026-09-26T16:30:23.790542956Z
  ended_at: 2026-09-26T17:22:19.440131106Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T16:30:23.790542956Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T17:22:19.440131106Z
observed_paths:
- docs/plans/2026-09-26-perl-git-t-parses.md
- internal/parse/control.go
- internal/parse/switch_test.go
- internal/parse/testdata/t1.ratchet
---

`not_a_term` is 1,584 Unknown nodes across 192 of perl.git t/'s files and has
had no bucketing pass. MEASUREMENT ONLY -- no fix, no parser change.

## Why this is its own issue

The `trailing_tokens` bucketing produced
`docs/plans/2026-09-26-perl-git-t-parses.md`, and it overturned three of four
guesses: `eval BLOCK` appears in 177 of 435 dirty files and parses PERFECTLY,
as do `for my ($k,$v)` and `sub()`. Had the plan been written from presence
counts it would have aimed its first fix at the largest population and moved
nothing.

`not_a_term` has had none of that. Fixing anything it covers before bucketing it
is the defect this chain keeps finding: read a count, name a cause, never
measure the cause separately.

## Baseline, measured 2026-09-26 at `59029a2c`

    perl.git t/   620 files, 185 clean (29.8%), 7745 Unknown nodes

    trailing_tokens          5982 nodes   412 files
    not_a_term               1584 nodes   192 files   <- this issue
    unimplemented_statement   169 nodes    15 files
    chain_class_mismatch         5 nodes     2 files
    ternary_no_colon             4 nodes     3 files
    missing_operand              1 node      1 file

## What to produce

A bucketing of those 1,584 nodes by what the Unknown SPAN CONTAINS, in the shape
the plan's `trailing_tokens` table already has:

    N   "leading 24 bytes of the span"

Then, for the top buckets only, the thing presence counts cannot give: does the
construct refuse IN ISOLATION? A five-line probe per candidate, one construct
per file, measured with `parse.Parse` and an Unknown count. That step is what
eliminated `eval BLOCK`, and it is the deliverable rather than the bucket table.

## How to measure it

`internal/parse` has no runner for this; a throwaway `zz*_test.go` in
`package parse_test` walking `/home/perigrin/dev/perl5/t` is how the plan's
numbers were taken. Delete it afterwards -- these are measurements, not tests.

`parse.Parse(src)` returns a `*Node`; walk `Children` and count
`Kind == parse.Unknown`, reading `x.Refusal` for the code and
`src[x.Start:x.End]` for the span.

perl is the authority on whether a candidate is valid:
`/home/perigrin/.local/bin/perl` is 5.42.0 and plenv is NOT on PATH.

## Scope

1. Bucket all 1,584 `not_a_term` nodes by span text. Report the top 25.
2. For each of the top buckets, write the smallest program containing that
   construct alone and measure its Unknown count. Report which refuse and which
   do not -- the negative results are the valuable half.
3. Append both tables to `docs/plans/2026-09-26-perl-git-t-parses.md`, in the
   section named "What is NOT yet measured", replacing the line that says
   `not_a_term` has had no bucketing pass.
4. File an issue per construct that BOTH refuses in isolation AND appears in
   more than a handful of files. Do not fix any of them here.

## Context

- paths: `docs/plans/2026-09-26-perl-git-t-parses.md`,
  `internal/parse/parse.go` (`Node`, `Kind`, `Unknown`, `Refusal`)
- commands: `go test ./internal/parse/ -run <your probe> -count=1 -v`
- corpus: `/home/perigrin/dev/perl5/t`, 620 `.t` files walked recursively

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance criteria

- [ ] The plan holds a `not_a_term` bucket table (`grep -q 'not_a_term.*bucket\|=== not_a_term' docs/plans/2026-09-26-perl-git-t-parses.md`)
- [ ] The plan records at least one construct measured NOT to be a cause (`grep -qi 'not a cause\|parses clean in isolation' docs/plans/2026-09-26-perl-git-t-parses.md`)
- [ ] No probe file is left behind (`test -z "$(ls internal/parse/zz*_test.go 2>/dev/null)"`)
