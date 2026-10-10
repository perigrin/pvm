---
title: "sub(f{}\\n,) parses as a declaration, though perl calls it an illegal anonymous subroutine"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-10T00:44:02.11049166Z
updated: 2026-10-10T00:49:57.106172009Z
sessions:
- start_sha: 99a185a40e19329e5cc261aee3cac312f9c13c82
  end_sha: ""
  commits: 0
  started_at: 2026-10-10T00:49:57.106172009Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-10T00:49:57.106172009Z
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
