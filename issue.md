---
title: "Inference reads (1..5) as Str: the literal-range defect 01a0afc0 claimed fixed is still in the tree-sitter path"
state: done
urgency: normal
milestone: v0.1
created: 2026-10-09T17:43:55.324336034Z
updated: 2026-10-16T03:49:43.636155832Z
sessions:
- start_sha: 133837eb651f8dcffb4203d072b3701753486869
  end_sha: 133837eb651f8dcffb4203d072b3701753486869
  commits: 0
  started_at: 2026-10-16T03:49:43.511986033Z
  ended_at: 2026-10-16T03:49:43.636155832Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T03:49:43.511986033Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T03:49:43.636155832Z
---

## Context

- paths: `internal/infer` (reads the tree-sitter wrapper `internal/parser`), `internal/parse` (whose tree already reads `1..5` as `..` over two numbers)
- commands: `make test`

Found 2026-10-09 auditing 01a0afc0's criteria. 01a0afc0 was closed claiming `my @r = (1..5)` gives a Range whose endpoints are numbers, not a Str. The parser half holds in internal/parse (TestLiteralRangeHasEndpoints). The inference half does not: `infer.Analyze` runs on internal/parser (tree-sitter), and there `my @r = (1..5);` infers the node at `1..5` as `Str`, while `(1 .. 5)` and `($a..$b)` infer `List`. So the §4.14.2 defect 01a0afc0 was opened for -- `@r` inferred as `Str` -- remains whenever the range has literal endpoints and no spaces. Moving internal/infer onto internal/parse (m2-lower-the-tree) is the likely fix; check before fixing the tree-sitter path.

## Steps

- [ ] Failing test: inference of `my @r = (1..5);` gives a List of numbers, not Str; fix (on the tree inference reads); `make test`; commit

## Acceptance criteria

- [ ] a literal range infers as a list, not a string (`go test ./internal/infer/ -run '^TestLiteralRangeInfersAList$' -count=1 -v | grep -q '^--- PASS: TestLiteralRangeInfersAList'`)
