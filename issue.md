---
title: eof() with empty parentheses means end of @ARGV, which a .pmt signature cannot tell from bare eof
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T12:26:26.068275997Z
updated: 2026-10-16T18:48:33.485105427Z
---

## Context

- paths: `internal/parse/declarations/CORE.tpm` (eof), RFC 0001
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`; `perldoc -f eof`

From 01a10177-5057-7f6f-bfa1-363a4568ae80. perl distinguishes three forms of eof: `eof FH` / `eof($fh)` (that handle), bare `eof` (the last handle read, `${^LAST_FH}`), and `eof()` with EMPTY PARENTHESES (end of the files in @ARGV -- the end of the `<>` stream, not of the current file). A .tpm signature cannot tell `eof` from `eof()`: both are a call with no argument. CORE.tpm's eof line notes this in a comment and types bare `eof`.

## Steps

- [ ] Measure the three forms on 5.42 (`perl -ne` over two files with each form)
- [ ] Decide how typed Perl states a meaning of empty parentheses distinct from no argument (perigrin), or record it as a parse-shape fact the parser keeps (not a type); failing test; `make test`; commit

## Acceptance criteria

- [ ] `eof()` and bare `eof` are told apart where it matters (`go test ./internal/parse/ -run '^TestEofEmptyParensIsArgv$' -count=1 -v | grep -q '^--- PASS: TestEofEmptyParensIsArgv'`)
