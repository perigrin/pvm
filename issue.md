---
title: "The lexer panics on a format with no name: format=\\n0\\n\\n indexes an empty slice in scanFormatBody"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T17:18:34.327458898Z
updated: 2026-10-09T17:18:43.976218315Z
sessions:
- start_sha: 7cbf3fdcf4df28374896f0db93391750fa1f926d
  end_sha: ""
  commits: 0
  started_at: 2026-10-09T17:18:43.976218315Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T17:18:43.976218315Z
---

## Context

- paths: `internal/lexer/pod.go` (`scanFormatBody`, line 151), `internal/parse/fuzz_test.go` (FuzzParser), `internal/parse/testdata/fuzz/FuzzParser/`
- commands: `make fuzz-parser`; `make test`

Found 2026-10-09 by `make fuzz-parser` (01a0a70f-0ec3's criterion) on feat/m1-finish at 7cbf3fdc: the input `format=\n0\n\n` panics the lexer -- `index out of range [0] with length 0` in scanFormatBody (pod.go:151). A lexer must never panic on any input (the plan's M0 non-negotiable: no panic on any input).

## Steps

- [ ] Measure what perl does with `format=\n0\n\n` (`perl -c`); failing test first (commit the seed as a FuzzParser corpus entry and a named unit test); fix scanFormatBody; `make fuzz-parser` clean; `make test`; commit

## Acceptance criteria

- [ ] the lexer does not panic on a format whose name is empty (`go test ./internal/lexer/ -run '^TestFormatWithNoNameDoesNotPanic$' -count=1 -v | grep -q '^--- PASS: TestFormatWithNoNameDoesNotPanic'`)
- [ ] the parser fuzzer finds nothing (`make fuzz-parser`)
