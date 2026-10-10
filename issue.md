---
title: "A $var before a term-start after print is the filehandle to perl: print $0 &y reads as an infix &"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-10T01:40:32.06078993Z
updated: 2026-10-10T01:40:32.06078993Z
---

## Context

- paths: `internal/parse/call.go` (`parseFilehandleSlot`), `internal/lexer` (perl decides part of this in toke.c), `internal/parse/testdata/fuzz/FuzzParser/`
- commands: `make fuzz-parser`; `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-10 by `make fuzz-parser` (after 01a12344). Seed parked at scratchpad/pending-seeds/0c172203842b1f8c: `print$0 &y` -- perl reads `$0` as print's filehandle and `&y` as a call (Deparse: `print $0 &y;`); internal/parse reads `print($0 & y())`, and canon's `y()` re-lexes as a transliteration. This is perl's toke.c rule for a `$var` followed by whitespace after a list operator (print/printf/say/exec/system/sort): the variable is the filehandle when what follows looks like the start of a term -- `&`, `*`, `<`, `%`, `'` before an identifier, `.3`, `-1`/`+1`, `?x`, `/re/`, `<<HEREDOC`. Measured: `print $0 -1;` Deparses as `print $0 -1;` (handle).

## Steps

- [ ] Read toke.c's rule (S_intuit_more / the list-operator `$var` case) and measure each trigger with Deparse; failing test first; implement in the parser and/or lexer as the rule requires; commit the seed; ratchets toward perl only; `make fuzz-parser` until one full run passes; `make test`; commit

## Acceptance criteria

- [ ] a `$var` before a term-start after a list operator is the filehandle (`go test ./internal/parse/ -run '^TestScalarHandleBeforeATermStart$' -count=1 -v | grep -q '^--- PASS: TestScalarHandleBeforeATermStart'`)
- [ ] the parser fuzzer finds nothing (`make fuzz-parser`)
