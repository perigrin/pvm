---
title: "Canon is not a fixpoint on no(A(A0)): a parenthesised no operand re-reads differently"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T21:15:56.35321275Z
updated: 2026-10-10T03:36:22.408079274Z
sessions:
- start_sha: 79cb88417fc43ff03ba5154ef6eb2928fd3b3ea0
  end_sha: baf52e02b9bfe6cccdacae5aafb9bd5e8d725c68
  commits: 7
  started_at: 2026-10-09T21:16:08.35755842Z
  ended_at: 2026-10-10T03:36:22.408079274Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T21:16:08.35755842Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-10T03:36:22.408079274Z
observed_paths:
- internal/lexer/fhslot_test.go
- internal/lexer/lexer.go
- internal/parse/anonhead_test.go
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/chain.go
- internal/parse/chain_test.go
- internal/parse/commafat_test.go
- internal/parse/control.go
- internal/parse/control_test.go
- internal/parse/cstylefor_test.go
- internal/parse/decl.go
- internal/parse/declaratornotarget_test.go
- internal/parse/expr.go
- internal/parse/parse.go
- internal/parse/postfixcallparen_test.go
- internal/parse/refusal.go
- internal/parse/scalarhandle_test.go
- internal/parse/statementanonhash_test.go
- internal/parse/term.go
- internal/parse/testdata/fuzz/FuzzParser/0c172203842b1f8c
- internal/parse/testdata/fuzz/FuzzParser/101e4c4a6858cc12
- internal/parse/testdata/fuzz/FuzzParser/372e02c8df1282f9
- internal/parse/testdata/fuzz/FuzzParser/3dd7051a66dfa870
- internal/parse/testdata/fuzz/FuzzParser/78dd41b5463a5add
- internal/parse/testdata/fuzz/FuzzParser/7bd29f70ee42a374
- internal/parse/testdata/fuzz/FuzzParser/8cbe7cdacf7dc963
- internal/parse/testdata/fuzz/FuzzParser/9f23c2442c703025
- internal/parse/testdata/fuzz/FuzzParser/a162c0187a1bd0a1
- internal/parse/testdata/fuzz/FuzzParser/a447ccd41522a9fd
- internal/parse/testdata/fuzz/FuzzParser/b45427e536ce28a7
- internal/parse/testdata/fuzz/FuzzParser/c729db5cac65ae20
- internal/parse/testdata/fuzz/FuzzParser/cf2931562936c99d
- internal/parse/testdata/fuzz/FuzzParser/daec3eb606eaafb5
- internal/parse/testdata/fuzz/FuzzParser/ea1b13bee7797dd5
- internal/parse/testdata/fuzz/FuzzParser/f0168060c2eb4a04
- internal/parse/testdata/fuzz/FuzzParser/f8029b60b6b06abd
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/trailingcomma_test.go
- internal/parse/unaryplus_test.go
- internal/parse/unclosed_test.go
- internal/parse/use.go
- internal/parse/use_test.go
- internal/parseoracle/subject_test.go
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
