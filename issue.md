---
title: "Refuse map/grep forms perl rejects: hash brace then a term-position operator, block at EOF"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-05T22:57:12.511225441Z
updated: 2026-10-05T22:57:12.511225441Z
---

## Context

- paths: `internal/parse/call.go` (`parseListOpBlock`), `internal/parse/mapbrace_test.go`
- commands: `make test`; the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`

Found while building 01a0fd9b-4e65-7397-836f-140fdc7feb53 (map/grep comma rule). Two forms perl 5.42 refuses that pvm accepts with 0 Unknown:

1. After a map/grep hash constructor's `}`, perl expects a term, so `map { "a" => 1 } + 1, @a` is a syntax error, as are `- 1`, `* 1`, `x 2`, `// 1`, `<=> 1`, `% 2`, `/ 2`, `< 1` there; `. "x"`, `== 1`, `|| 1`, `&&`, `=~`, `or`, `? :` and `->` are accepted. (Recorded in RFC 0001 and spec §4.9.2.)
2. `map { 1 }` at end of input with no `;`: perl says "syntax error at EOF"; `grep { 1 } ;` is already refused.

## Steps

- [ ] Re-measure every listed operator with perl; write failing test `TestMapGrepHashBraceThenTermOperatorRefused`; implement in the map/grep path only; `make test`; commit
- [ ] Write failing test `TestMapGrepBlockAtEOFRefused`; implement; `make test`; commit

## Acceptance criteria

- [ ] a hash brace followed by an operator perl reads as a term is refused (`go test ./internal/parse/ -run '^TestMapGrepHashBraceThenTermOperatorRefused$' -count=1 -v | grep -q '^--- PASS: TestMapGrepHashBraceThenTermOperatorRefused'`)
- [ ] a map block at EOF with no list is refused (`go test ./internal/parse/ -run '^TestMapGrepBlockAtEOFRefused$' -count=1 -v | grep -q '^--- PASS: TestMapGrepBlockAtEOFRefused'`)
