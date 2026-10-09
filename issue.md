---
title: A format terminator in a CRLF file may not end the format where perl's does
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T07:13:55.90088278Z
updated: 2026-10-09T07:13:55.90088278Z
---

## Context

- paths: `internal/lexer/pod.go` (`scanFormatBody`), `internal/parse/canon.go` (format emission)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found fixing 01a11f42-2f16 (2026-10-09): perl ends a format at a `.` line followed by blanks and then a newline or end of file; `scanFormatBody` trims spaces and tabs only, so in a CRLF file the `.\r\n` terminator may not end the format where perl's does. Not yet measured. Relates to 01a10f61 (Windows: CRLF corpus files).

## Steps

- [ ] Measure perl 5.42 on a CRLF file's format terminator (`.\r\n`, `. \r\n`, `.\r` at EOF) with `-c` and `-MO=Deparse`
- [ ] If pvm differs, failing test first; fix; `make test`; commit

## Acceptance criteria

- [ ] a CRLF format terminator ends the format where perl's does (`go test ./internal/lexer/ -run '^TestFormatTerminatorCRLF$' -count=1 -v | grep -q '^--- PASS: TestFormatTerminatorCRLF'`)
