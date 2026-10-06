---
title: "sub g : method lvalue {} is refused; perl accepts space-separated attributes after one colon"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T10:59:34.371019964Z
updated: 2026-10-06T10:59:34.371019964Z
---

## Context

- paths: `internal/parse/use.go` (`parseAttributes`), `internal/lexer`
- commands: `make test`; ratchets; perl 5.42 `/home/perigrin/.local/bin/perl`

Found while building 01a10dc1-3530-762f-9a03-177445629c6b. perl accepts a sub attribute list with one colon and space-separated attributes, `sub g : method lvalue {}` (perlsub: "the colon may be omitted between attributes"); pvm refuses it as `trailing_tokens`. A wrong refusal of ordinary source.

## Steps

- [ ] Measure the forms perl accepts (`: a b`, `:a :b`, `:a b`, with arguments) with `perl -c` and Deparse
- [ ] Write a failing test `TestSubAttributeListSpaceSeparated`; fix parseAttributes; ratchets must not regress (report movement); `make test`; commit

## Acceptance criteria

- [ ] a space-separated attribute list parses (`go test ./internal/parse/ -run '^TestSubAttributeListSpaceSeparated$' -count=1 -v | grep -q '^--- PASS: TestSubAttributeListSpaceSeparated'`)
