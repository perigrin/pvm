---
title: Choose the map and grep variant by the comma after the closing brace
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T17:13:28.933338968Z
updated: 2026-10-02T17:13:28.933338968Z
---

## Prerequisites

- none

## Context

- paths: `internal/parse/call.go`, `internal/parse/mapbrace_test.go` (new)
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `takesBlock`/`parseListOpBlock` (`internal/parse/call.go:480-500`), `Token.OpensBlock` (`internal/lexer/lexer.go:188-200`), the lexer's brace guess (`internal/lexer/expect.go:190`)

RFC 0001 "Builtins that keep their own parse": `map`'s and `grep`'s brace is variant selection. perl guesses block or hash constructor from the first tokens; wherever its guess disagrees with what follows the closing brace, perl rejects the program (`map { "\L$_" => 1 } @a` guessed hash with no comma, `map { $_ => 1 }, @a` guessed block then a comma: both syntax errors). So a comma after `}` selects the expression variant and its absence the block variant, which agrees with perl on every program perl accepts; `map { $_ => 1 } @a` holds pairs and is still a valid block. Today `parseListOpBlock` follows `tok.OpensBlock` alone. After this issue the comma decides, and the first-tokens guess remains only to refuse what perl refuses. Measure every case with perl before writing its expectation. Needs nothing else in this milestone.

## Steps

- [ ] Measure with perl, then write failing test `TestMapGrepCommaSelectsVariant` in `internal/parse/mapbrace_test.go`: for map and grep, `{...} LIST` parses as the block variant and `{...}, LIST` as the expression variant (anon hash first argument), across brace contents perl guesses either way
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestMapGrepCommaSelectsVariant$' -count=1 -v`
- [ ] Implement in `internal/parse/call.go`; run test to verify pass; `make test` (watch the T1/T2 ratchets); commit
- [ ] Write failing test `TestMapGrepRefusesWhatPerlRefuses`: `map { "\L$_" => 1 } @a` and `map { $_ => 1 }, @a` are refused
- [ ] Implement with the first-tokens guess used only for refusal; run test to verify pass; `make test`; commit

## Acceptance criteria

- [ ] the comma after the closing brace selects the map or grep variant (`go test ./internal/parse/ -run '^TestMapGrepCommaSelectsVariant$' -count=1 -v | grep -q '^--- PASS: TestMapGrepCommaSelectsVariant'`)
- [ ] map and grep forms perl rejects are refused (`go test ./internal/parse/ -run '^TestMapGrepRefusesWhatPerlRefuses$' -count=1 -v | grep -q '^--- PASS: TestMapGrepRefusesWhatPerlRefuses'`)
