---
title: An undeclared bareword argument is read as a sub call where perl reads a string
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T10:01:22.426443886Z
updated: 2026-10-09T10:01:22.426443886Z
---

## Context

- paths: `internal/parse/call.go` (`readHandleSlots`, `isBarewordHandle`, `classInvocant`), `internal/parse/term.go`
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found by 01a11fe0 (2026-10-09). An undeclared bareword outside a handle slot is read by pvm as a sub call `foo()`, where perl (without `use strict 'subs'`, which makes it a compile error) reads it as a string. Measured on 5.42 with Deparse: `seek(FH, foo, 0); chmod 0644, FH;` gives `seek FH, 'foo', 0; chmod 0644, 'FH';`. Also `foo(FH)`. A sub declared before the word makes it a call. Under `use strict` perl refuses the bareword ("Bareword not allowed while strict subs in use"). Related: `use` lists (`p.listed`) are counted as subs by classInvocant but not by isBarewordHandle.

## Steps

- [ ] Measure perl with and without strict subs, with the sub declared before/after; failing test first; read the word as perl does (string, call, or refusal under strict); ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] an undeclared bareword argument is read as perl reads it (`go test ./internal/parse/ -run '^TestUndeclaredBarewordArgument$' -count=1 -v | grep -q '^--- PASS: TestUndeclaredBarewordArgument'`)
