---
title: "The lexer panics on a format with no name: format=\\n0\\n\\n indexes an empty slice in scanFormatBody"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T17:18:34.327458898Z
updated: 2026-10-10T05:34:51.288400628Z
sessions:
- start_sha: 7cbf3fdcf4df28374896f0db93391750fa1f926d
  end_sha: 79cb88417fc43ff03ba5154ef6eb2928fd3b3ea0
  commits: 2
  started_at: 2026-10-09T17:18:43.976218315Z
  ended_at: 2026-10-09T19:48:53.66484412Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T17:18:43.976218315Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T19:48:53.66484412Z
observed_paths:
- internal/lexer/brackets.go
- internal/lexer/delim_test.go
- internal/lexer/expect.go
- internal/lexer/heredoc_test.go
- internal/lexer/lastidx_test.go
- internal/lexer/lexer.go
- internal/lexer/pod.go
- internal/lexer/pod_test.go
- internal/lexer/quote.go
- internal/lexer/quote_test.go
- internal/lexer/scan.go
- internal/lexer/sigilspace_test.go
- internal/parse/ambiguoussigil_test.go
- internal/parse/canon.go
- internal/parse/control.go
- internal/parse/control_test.go
- internal/parse/decl.go
- internal/parse/decl_test.go
- internal/parse/derefcanon_test.go
- internal/parse/expr.go
- internal/parse/formateofcanon_test.go
- internal/parse/keyquoteop_test.go
- internal/parse/literalrange_test.go
- internal/parse/methodgate_test.go
- internal/parse/refusal.go
- internal/parse/refusal_test.go
- internal/parse/term.go
- internal/parse/testdata/fuzz/FuzzParser/0a7bf5a6847330ca
- internal/parse/testdata/fuzz/FuzzParser/1d8e878b245a4d18
- internal/parse/testdata/fuzz/FuzzParser/265c688a3cda12ee
- internal/parse/testdata/fuzz/FuzzParser/6e82dbe27cb52f90
- internal/parse/testdata/fuzz/FuzzParser/6f4123251d19517b
- internal/parse/testdata/fuzz/FuzzParser/8cdacffd12f615d6
- internal/parse/testdata/fuzz/FuzzParser/97207773ca613314
- internal/parse/testdata/fuzz/FuzzParser/ae85ae101f71eea1
- internal/parse/testdata/fuzz/FuzzParser/aec6e6ec684aa209
- internal/parse/testdata/fuzz/FuzzParser/bea487d46dc6ff51
- internal/parse/testdata/fuzz/FuzzParser/dd9600c0bfa3547c
- internal/parse/testdata/fuzz/FuzzParser/e7cce4b02f0018d1
- internal/parse/testdata/fuzz/FuzzParser/ef55e35659d56262
- internal/parse/testdata/fuzz/FuzzParser/f533dab875bee51d
- internal/parse/testdata/fuzz/FuzzParser/fc94f4f5f7e5a508
- internal/parse/trailingcomma_test.go
- internal/parse/unaryplus_test.go
- internal/parse/unclosed_test.go
- internal/parse/wrongtree_test.go
---

## Context

- paths: `internal/lexer/pod.go` (`scanFormatBody`, line 151), `internal/parse/fuzz_test.go` (FuzzParser), `internal/parse/testdata/fuzz/FuzzParser/`
- commands: `make fuzz-parser`; `make test`

Found 2026-10-09 by `make fuzz-parser` (01a0a70f-0ec3's criterion) on feat/m1-finish at 7cbf3fdc: the input `format=\n0\n\n` panics the lexer -- `index out of range [0] with length 0` in scanFormatBody (pod.go:151). A lexer must never panic on any input (the plan's M0 non-negotiable: no panic on any input).

## Steps

- [ ] Measure what perl does with `format=\n0\n\n` (`perl -c`); failing test first (commit the seed as a FuzzParser corpus entry and a named unit test); fix scanFormatBody; `make fuzz-parser` clean; `make test`; commit

### The fuzz criterion replays the seeds (2026-10-10)

`make fuzz-parser` explores at random, so as a criterion it can fail on a
new find unrelated to this issue -- and a failing run writes its seed into
testdata, where every later `go test ./internal/parse/` replays it. This
issue's claim is its own inputs, which are committed as fuzz seeds; the
criterion replays them deterministically. The open-ended fuzz claim
belongs to 01a0a70f-0ec3, where it stays.

## Acceptance criteria

- [ ] the lexer does not panic on a format whose name is empty (`go test ./internal/lexer/ -run '^TestFormatWithNoNameDoesNotPanic$' -count=1 -v | grep -q '^--- PASS: TestFormatWithNoNameDoesNotPanic'`)
- [ ] every committed fuzz seed replays clean (`go test ./internal/parse/ -run '^FuzzParser$' -count=1 -v | grep -q '^--- PASS: FuzzParser'`)
