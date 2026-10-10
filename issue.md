---
title: "sub(f{}\\n,) parses as a declaration, though perl calls it an illegal anonymous subroutine"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-10T00:44:02.11049166Z
updated: 2026-10-10T03:36:07.464778547Z
sessions:
- start_sha: 99a185a40e19329e5cc261aee3cac312f9c13c82
  end_sha: baf52e02b9bfe6cccdacae5aafb9bd5e8d725c68
  commits: 4
  started_at: 2026-10-10T00:49:57.106172009Z
  ended_at: 2026-10-10T03:36:07.464778547Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-10T00:49:57.106172009Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-10T03:36:07.464778547Z
observed_paths:
- internal/lexer/fhslot_test.go
- internal/lexer/lexer.go
- internal/parse/anonhead_test.go
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/chain.go
- internal/parse/chain_test.go
- internal/parse/commafat_test.go
- internal/parse/decl.go
- internal/parse/expr.go
- internal/parse/parse.go
- internal/parse/refusal.go
- internal/parse/scalarhandle_test.go
- internal/parse/term.go
- internal/parse/testdata/fuzz/FuzzParser/0c172203842b1f8c
- internal/parse/testdata/fuzz/FuzzParser/7bd29f70ee42a374
- internal/parse/testdata/fuzz/FuzzParser/b45427e536ce28a7
- internal/parse/testdata/fuzz/FuzzParser/daec3eb606eaafb5
- internal/parse/testdata/fuzz/FuzzParser/f0168060c2eb4a04
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/trailingcomma_test.go
- internal/parse/unaryplus_test.go
- internal/parseoracle/subject_test.go
---

## Context

- paths: `internal/parse/decl.go` (anonymous `sub` declarations), `internal/lexer` (prototype/signature lexing after `sub`), `internal/parse/testdata/fuzz/FuzzParser/`
- commands: `make fuzz-parser`; `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-10 by `make fuzz-parser` (after 01a12312-770f). Seed parked at scratchpad/pending-seeds/daec3eb606eaafb5: `sub(f{}\n,)`. With the newline it parses as a Declaration `sub` whose child is a parenthesised Call `f` holding a Block, and canon writes `sub f {} ,` -- not a fixpoint (6 significant tokens, then 5). Without the newline, `sub(f{},)` lexes the parens as a Prototype. perl refuses both: "Illegal declaration of anonymous subroutine". Fails at 419a99b6 too.

## Steps

- [ ] Measure perl on anonymous `sub (` forms with and without a body, across newlines, in statement and expression position; failing test first; refuse what perl refuses (existing refusal site); commit the seed; `make fuzz-parser` until one full run passes; `make test`; commit

## Acceptance criteria

- [ ] an anonymous sub with a paren and no body is refused (`go test ./internal/parse/ -run '^TestAnonSubWithoutBodyIsRefused$' -count=1 -v | grep -q '^--- PASS: TestAnonSubWithoutBodyIsRefused'`)
- [ ] the parser fuzzer finds nothing (`make fuzz-parser`)
