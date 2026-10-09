---
title: "Canon is not a fixpoint on no(A(A0)): a parenthesised no operand re-reads differently"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T21:15:56.35321275Z
updated: 2026-10-09T21:16:08.35755842Z
sessions:
- start_sha: 79cb88417fc43ff03ba5154ef6eb2928fd3b3ea0
  end_sha: ""
  commits: 0
  started_at: 2026-10-09T21:16:08.35755842Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T21:16:08.35755842Z
---

## Context

- paths: `internal/parse` (`no`/`use` parsing, canon), `internal/parse/testdata/fuzz/FuzzParser/`
- commands: `make fuzz-parser`; `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-09 by `make fuzz-parser` during the final m1 verify, at 79cb8841: `no(A(A0))` -- canon is not a fixpoint: token 2, the first emission has `Operator (`, the second `Word A0`. Seed parked at scratchpad/pending-seeds/9f23c2442c703025 (string "no(A(A0))").

## Steps

- [ ] Measure perl on `no(A(A0))`, `use(A(A0))` and related forms; failing test first; fix; commit the seed; `make fuzz-parser` clean; `make test`; commit

## Acceptance criteria

- [ ] canon is a fixpoint on a parenthesised `no`/`use` operand (`go test ./internal/parse/ -run '^TestParenthesisedNoOperandIsAFixpoint$' -count=1 -v | grep -q '^--- PASS: TestParenthesisedNoOperandIsAFixpoint'`)
- [ ] the parser fuzzer finds nothing (`make fuzz-parser`)
