---
title: An empty {} after print, printf, system or exec is a hash to perl but a block to the lexer
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T09:09:07.186145077Z
updated: 2026-10-09T09:34:57.436607204Z
sessions:
- start_sha: cfa7b6880d8cf587db4d4537ba1fa256261b1d64
  end_sha: ""
  commits: 0
  started_at: 2026-10-09T09:34:57.436607204Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T09:34:57.436607204Z
---

## Context

- paths: `internal/lexer/keyword.go` (`emptyBraceIsHash`), `internal/lexer/intuit.go` (`intuitCurly`), `internal/lexer/lexer.go`
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found by 01a11fbd (map/grep/sort's empty brace is an anonymous hash), 2026-10-09. Measured on 5.42: `print {} "x"`, `printf {} "x"`, `system {} "x"`, `exec {} "x"` -- and `say {} "x"` with feature say on -- read the empty brace as an anonymous hash and are syntax errors; pvm's lexer reads it as a block (the filehandle/program block form). `CORE::map {}` is not matched either (`takesBlock` doesn't strip `CORE::`), unmeasured. `say`'s answer depends on its feature, which the lexer does not track.

## Steps

- [ ] Measure each; failing test first; extend emptyBraceIsHash (and CORE:: forms as measured); ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] an empty brace after print/printf/system/exec is read as perl reads it (`go test ./internal/lexer/ -run '^TestEmptyBraceAfterFilehandleBuiltins$' -count=1 -v | grep -q '^--- PASS: TestEmptyBraceAfterFilehandleBuiltins'`)
