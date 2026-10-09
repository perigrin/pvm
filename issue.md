---
title: "A bareword filehandle in open's first slot canonicalises as a sub call: open(FH, ...) emits open(FH() , ...)"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T08:56:11.112212718Z
updated: 2026-10-09T08:56:11.112212718Z
---

## Context

- paths: `internal/parse/call.go` (how a bareword in a filehandle slot becomes a Call), CORE.pmt's `open`/`close`/etc. filehandle slots, `internal/parse/canon.go`
- commands: `make test`; the canon/fidelity ratchets `go test ./internal/parse/ -run '^(TestCanonRatchet|TestCanonFidelityRatchet)$' -count=1`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found by 01a0f5eb (a bareword class invocant is a string), 2026-10-09: `open(FH, "<", $f)` canonicalises as `open(FH() , ...)` -- the bareword filehandle inside `open`'s first slot is read as a sub call. perl reads a bareword in a filehandle slot as the handle (Deparse: `open FH, '<', $f;`).

## Steps

- [ ] Measure perl on a bareword handle in each builtin's filehandle slot (open, close, binmode, eof, print FH, opendir, ...) and with a sub of that name declared
- [ ] Failing test first; read a bareword in a filehandle slot as perl does; ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] a bareword filehandle in a builtin's handle slot is not a call (`go test ./internal/parse/ -run '^TestBarewordFilehandleSlotIsNotACall$' -count=1 -v | grep -q '^--- PASS: TestBarewordFilehandleSlotIsNotACall'`)
