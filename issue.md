---
title: Choose the map and grep variant by the comma after the closing brace
state: done
urgency: normal
milestone: type-libraries
blocks:
- 01a0fd9b-519d-7b42-886e-2222a3076986
created: 2026-10-02T17:13:28.933338968Z
updated: 2026-10-05T22:59:34.663793809Z
sessions:
- start_sha: 579dfb99ea6dd63d2b66cfcf3547201f41d18513
  end_sha: 579dfb99ea6dd63d2b66cfcf3547201f41d18513
  commits: 0
  started_at: 2026-10-05T22:57:22.914190622Z
  ended_at: 2026-10-05T22:57:46.220286883Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-05T22:57:22.914190622Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-05T22:57:46.220286883Z
---

## Prerequisites

- none

## Context

- paths: `internal/parse/call.go`, `internal/parse/mapbrace_test.go` (new), `docs/specs/perl-parser/04-expressions.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `takesBlock`/`parseListOpBlock` (`internal/parse/call.go:480-500`), `Token.OpensBlock` (`internal/lexer/lexer.go:188-200`), the lexer's brace guess (`internal/lexer/expect.go:190`)

RFC 0001 "Builtins that keep their own parse": `map`'s and `grep`'s brace is variant selection. perl guesses block or hash constructor from the first tokens; wherever its guess disagrees with what follows the closing brace, perl rejects the program (`map { "\L$_" => 1 } @a` guessed hash with no comma, `map { $_ => 1 }, @a` guessed block then a comma: both syntax errors). So the comma that ENDS THE FIRST ARGUMENT selects the expression variant and its absence the block variant; `map { $_ => 1 } @a` holds pairs and is still a valid block. The comma need not follow `}` directly: perl 5.42 accepts `map { foo => 1 }->{foo}, @a`, `map { "a" => 1 } ? 1 : 2, @a` and `map {}->{a}, @a` (a hash constructor continued by postfix or operators before the comma), and refuses `map { $_ => 1 }->{a}, @a`. Measured over 46 forms (2026-10-05), perl refuses exactly where its first-tokens guess disagrees with whether the first argument ends at a comma, so on every accepted program the guess and the comma pick the same variant: the parser may keep the guess to choose and refuses where the first argument's end disagrees. Today pvm already matches perl except three cases: `map {} @a`/`grep {} @a` accepted (perl refuses), `map {}, @a`/`grep {}, @a` refused (perl accepts as an anon hash), `map { $_` at EOF parsed silently. Empty `{}` is a hash for map/grep only (toke.c ~6706); the lexer's `intuitCurly` must not change, because it runs after every word (changing it moved T2 op/goto.t, op/loopctl.t, op/method.t) -- fix it in the parser's map/grep path. The RFC sentence "a comma after `}` ... agrees with perl on every program perl accepts" is false as written and is corrected by this issue. Measure every case with perl before writing its expectation. Needs nothing else in this milestone.

## Steps

- [ ] Measure with perl, then write failing test `TestMapGrepCommaSelectsVariant` in `internal/parse/mapbrace_test.go`: for map and grep, `{...} LIST` parses as the block variant and `{...}, LIST` as the expression variant (anon hash first argument), across brace contents perl guesses either way, and a hash constructor continued before the comma (`map { foo => 1 }->{foo}, @a`, `map { "a" => 1 } ? 1 : 2, @a`, `map {}->{a}, @a`) is the expression variant
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestMapGrepCommaSelectsVariant$' -count=1 -v`
- [ ] Implement in `internal/parse/call.go`; run test to verify pass; `make test` (watch the T1/T2 ratchets); commit
- [ ] Write failing test `TestMapGrepRefusesWhatPerlRefuses`: `map { "\L$_" => 1 } @a` and `map { $_ => 1 }, @a` are refused
- [ ] Implement with the first-tokens guess used only for refusal; run test to verify pass; `make test`; commit
- [ ] Doc: in RFC 0001 "Builtins that keep their own parse", correct "a comma after `}`" to the comma that ends the first argument, citing the postfix measurement; `make test`; commit
- [ ] Doc: rewrite `docs/specs/perl-parser/04-expressions.md` §4.9.2 "The `{` decision" so its rule for a Go implementation is the comma that ends the first argument, citing RFC 0001 "Builtins that keep their own parse", with perl's first-tokens guess kept only to refuse what perl refuses; commit

### Negative Scenarios

- more forms perl 5.42 refuses (measured): `map {} @a` (empty braces, no comma), `map {; "a" => 1 }, @a` (forced block, then a comma), `grep { 1 } ;` (block, no list) (`go test ./internal/parse/ -run '^TestMapGrepBraceEdgeRefusals$' -count=1 -v | grep -q '^--- PASS: TestMapGrepBraceEdgeRefusals'`)
- forms perl 5.42 accepts stay accepted as the expression variant (measured): `map {}, @a` and `grep { "a" => 1 }, @a` (`go test ./internal/parse/ -run '^TestMapGrepAnonHashBeforeComma$' -count=1 -v | grep -q '^--- PASS: TestMapGrepAnonHashBeforeComma'`)
- an unterminated brace (`map { $_` at EOF) is an error, not a panic or a hang (`go test ./internal/parse/ -run '^TestMapGrepUnterminatedBrace$' -count=1 -v | grep -q '^--- PASS: TestMapGrepUnterminatedBrace'`)

## Acceptance criteria

- [ ] the comma ending the first argument selects the map or grep variant (`go test ./internal/parse/ -run '^TestMapGrepCommaSelectsVariant$' -count=1 -v | grep -q '^--- PASS: TestMapGrepCommaSelectsVariant'`)
- [ ] map and grep forms perl rejects are refused (`go test ./internal/parse/ -run '^TestMapGrepRefusesWhatPerlRefuses$' -count=1 -v | grep -q '^--- PASS: TestMapGrepRefusesWhatPerlRefuses'`)
- [ ] empty braces, forced blocks before a comma, and a block with no list are refused (`go test ./internal/parse/ -run '^TestMapGrepBraceEdgeRefusals$' -count=1 -v | grep -q '^--- PASS: TestMapGrepBraceEdgeRefusals'`)
- [ ] an anon hash before the comma is the expression variant (`go test ./internal/parse/ -run '^TestMapGrepAnonHashBeforeComma$' -count=1 -v | grep -q '^--- PASS: TestMapGrepAnonHashBeforeComma'`)
- [ ] an unterminated map brace is an error, not a panic (`go test ./internal/parse/ -run '^TestMapGrepUnterminatedBrace$' -count=1 -v | grep -q '^--- PASS: TestMapGrepUnterminatedBrace'`)
- [ ] the RFC no longer says the comma follows the closing brace (`sh -c '! grep -q "a comma after .}. selects" docs/rfcs/0001-type-libraries.md'`)
- [ ] the parser spec's brace rule cites the RFC's comma rule (`sh -c "sed -n '/^### 4.9.2/,/^### 4.9.3/p' docs/specs/perl-parser/04-expressions.md | grep -q 'RFC 0001'"`)
