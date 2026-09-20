---
title: Measure the 147 unbucketed Unknowns in four comp files before fixing any of them
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0c13f-6816-79d2-945a-b84a12d39e6c
created: 2026-09-20T23:56:17.751068121Z
updated: 2026-09-20T23:56:40.423688587Z
---

Four files carry 147 of T2's 649 Unknown nodes and nobody knows what is in
them. This issue MEASURES first and files second. Do not guess a cause and do
not write a fix inside this issue.

## Context

- paths: `internal/parse/t2_test.go` (the shortfall map is the inventory)
- commands: `go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v`
- entrypoints: `internal/parse/t2_test.go`'s `shortfall` map; the corpus at
  `$PERL5_CORPUS/t/` (default `~/dev/perl5/t/`)

## What is unmeasured, at b6e7c351

    comp/parser.t   65 Unknowns
    comp/hints.t    36
    comp/colon.t    25
    comp/redef.t    21
    -------------------
                   147

The survey (01a0c10b) bucketed 502 of T2's 649 nodes by construct. These 147
it listed as "(unbucketed)". Two of them already have partial notes in the
`shortfall` map's comments -- `comp/parser.t` is recorded as holding a
symbol-table dereference `$::{"..."}` that `TestDerefBlockContentsAreDecided`
names as still opaque -- but no file has a construct breakdown.

Note the overlap risk: the parenless-call fix and the leading-decimal fix may
account for a share of these. Measure AFTER those land, or measure now and
re-measure after, but say which.

## Why measurement is the deliverable

The survey exists because filing 39 files of work as one issue is what
produced a milestone with two targets missed. Filing a FIX for 147 unbucketed
nodes would repeat that at smaller scale. The unit of work is a construct
with a minimal reproduction, and this issue produces those units.

## Steps

- [ ] For each of the four files, extract every Unknown node's source span
      and group the spans by construct
- [ ] Reduce each group to a minimal reproduction of two or three lines,
      measured -- the form `f 13 % 4;  1 Unknown` the survey uses
- [ ] Record the per-file breakdown in this issue: construct, node count,
      minimal reproduction
- [ ] For each construct that is M1's, file its own issue in
      `m1-parse-the-core` with the measurement and a failing-test step
- [ ] For each construct that is NOT M1's (prototypes are M4's; anything the
      m2 refinement already blocked, such as block-form filehandles), record
      the exclusion here with the milestone it belongs to
- [ ] Commit the breakdown as a note where it can be found -- the shortfall
      map's comments are where the other per-file findings live

## Acceptance Criteria

### Positive Scenarios

- [ ] Every one of the 147 nodes is attributed to a named construct or
      explicitly recorded as residual, with the arithmetic shown
- [ ] Each M1-owned construct has its own issue in `m1-parse-the-core` (`git zhi issue list --milestone m1-parse-the-core | grep -c .`)
- [ ] The findings are committed beside the counts they explain, so the next
      reader does not re-measure (`grep -q 'comp/parser.t' internal/parse/t2_test.go`)
- [ ] The T2 metric is still green -- this issue changes no parse behaviour (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)

Plan: M1. Survey: 01a0c10b, the "(unbucketed)" row.
