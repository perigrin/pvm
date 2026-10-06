---
title: A .pmt that declares one name both as sub and as multi sub is an error
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-06T03:27:30.293600081Z
updated: 2026-10-06T03:27:30.293600081Z
---

## Context

- paths: `internal/parse/decl.go` (multi/plain sub recording into `moduleFacts.signatures`), `internal/parse/pmt_test.go`, `docs/rfcs/0001-type-libraries.md` ("Multi declarations")
- commands: `make test`

Found by the alignment gate on 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c. In a .pmt the signatures table is `map[string][]types.Signature`: a plain `sub f` replaces the name's entry and `multi sub f` appends. Mixing the two is order-dependent and silent: `sub f ...; multi sub f ...;` keeps two candidates, `multi sub f ...; sub f ...;` keeps only the last. RFC 0001 does not say what mixing means. Proposed (pending perigrin): mixing plain and multi declarations of one name in a .pmt is a declaration error naming the sub, as Raku refuses an `only` beside a `multi`.

## Steps

- [ ] Confirm the rule with perigrin; record it in RFC 0001 "Multi declarations"
- [ ] Write failing test `TestPmtPlainAndMultiSubMixIsError` covering both orders; implement; `make test`; commit

## Acceptance criteria

- [ ] mixing sub and multi sub for one name is an error in either order (`go test ./internal/parse/ -run '^TestPmtPlainAndMultiSubMixIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtPlainAndMultiSubMixIsError'`)
