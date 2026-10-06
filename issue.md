---
title: delete of a slice and four-argument select return lists, which their CORE.pmt rows do not say
state: done
urgency: normal
milestone: type-libraries
created: 2026-10-06T12:53:33.017885437Z
updated: 2026-10-06T14:14:42.072836223Z
sessions:
- start_sha: 7c49ae4f180e9c84cf8c01a4d9ce30fed5c9b881
  end_sha: 7c49ae4f180e9c84cf8c01a4d9ce30fed5c9b881
  commits: 0
  started_at: 2026-10-06T14:14:41.650993839Z
  ended_at: 2026-10-06T14:14:42.072836223Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T14:14:41.650993839Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T14:14:42.072836223Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt` (delete, select), `internal/parse/corenoproto_test.go`, internal/types/signatures.go's copied rows
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

From the alignment gate on 01a0fd9b-5017-78d0-ae38-331cbe6a9866, measured on 5.42; both rows were copied as the issue required and are now known short:

1. `delete` is typed `(Scalar $element) Scalar`, but `delete @h{qw(a b)}` (a slice) yields the list `(1, 2)`. RFC open question 10 (an element operand) does not cover a slice's list result.
2. The four-argument `select` candidate returns `Int`, but in list context it returns two values, nfound and timeleft (`my @r = select(undef, undef, undef, 0.01)` is `(0, 0)`). RFC 0001's own example row has it as Int.

## Steps

- [ ] For each, decide the declaration: `:context` candidates (delete of a slice vs an element is a parse-shape fork; select's scalar vs list result), measured first; failing tests; update CORE.pmt and the RFC example row; `make test`; commit

## Acceptance criteria

- [ ] delete of a slice and four-argument select in list context are typed as perl returns (`go test ./internal/parse/ -run '^TestCoreDeleteSliceAndSelectListRows$' -count=1 -v | grep -q '^--- PASS: TestCoreDeleteSliceAndSelectListRows'`)
