---
title: Choose the map and grep variant by the comma after the closing brace
state: pending
urgency: normal
milestone: type-libraries
blocks:
- 01a0fd9b-519d-7b42-886e-2222a3076986
created: 2026-10-02T17:13:28.933338968Z
updated: 2026-10-02T17:26:26.727232285Z
---

## Prerequisites

- none

## Context

- paths: `internal/parse/call.go`, `internal/parse/mapbrace_test.go` (new), `docs/specs/perl-parser/04-expressions.md`
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
- [ ] Doc: rewrite `docs/specs/perl-parser/04-expressions.md` §4.9.2 "The `{` decision" so its rule for a Go implementation is the comma after `}`, citing RFC 0001 "Builtins that keep their own parse", with perl's first-tokens guess kept only to refuse what perl refuses; commit

### Negative Scenarios

- more forms perl 5.42 refuses (measured): `map {} @a` (empty braces, no comma), `map {; "a" => 1 }, @a` (forced block, then a comma), `grep { 1 } ;` (block, no list) (`go test ./internal/parse/ -run '^TestMapGrepBraceEdgeRefusals$' -count=1 -v | grep -q '^--- PASS: TestMapGrepBraceEdgeRefusals'`)
- forms perl 5.42 accepts stay accepted as the expression variant (measured): `map {}, @a` and `grep { "a" => 1 }, @a` (`go test ./internal/parse/ -run '^TestMapGrepAnonHashBeforeComma$' -count=1 -v | grep -q '^--- PASS: TestMapGrepAnonHashBeforeComma'`)
- an unterminated brace (`map { $_` at EOF) is an error, not a panic or a hang (`go test ./internal/parse/ -run '^TestMapGrepUnterminatedBrace$' -count=1 -v | grep -q '^--- PASS: TestMapGrepUnterminatedBrace'`)

## Acceptance criteria

- [ ] the comma after the closing brace selects the map or grep variant (`go test ./internal/parse/ -run '^TestMapGrepCommaSelectsVariant$' -count=1 -v | grep -q '^--- PASS: TestMapGrepCommaSelectsVariant'`)
- [ ] map and grep forms perl rejects are refused (`go test ./internal/parse/ -run '^TestMapGrepRefusesWhatPerlRefuses$' -count=1 -v | grep -q '^--- PASS: TestMapGrepRefusesWhatPerlRefuses'`)
- [ ] empty braces, forced blocks before a comma, and a block with no list are refused (`go test ./internal/parse/ -run '^TestMapGrepBraceEdgeRefusals$' -count=1 -v | grep -q '^--- PASS: TestMapGrepBraceEdgeRefusals'`)
- [ ] an anon hash before the comma is the expression variant (`go test ./internal/parse/ -run '^TestMapGrepAnonHashBeforeComma$' -count=1 -v | grep -q '^--- PASS: TestMapGrepAnonHashBeforeComma'`)
- [ ] an unterminated map brace is an error, not a panic (`go test ./internal/parse/ -run '^TestMapGrepUnterminatedBrace$' -count=1 -v | grep -q '^--- PASS: TestMapGrepUnterminatedBrace'`)
- [ ] the parser spec's brace rule cites the RFC's comma rule (`sh -c "sed -n '/^### 4.9.2/,/^### 4.9.3/p' docs/specs/perl-parser/04-expressions.md | grep -q 'RFC 0001'"`)
