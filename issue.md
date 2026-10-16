---
title: ":context edge cases in .pmt: no signature, repeated, blank inside"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T05:27:24.45974331Z
updated: 2026-10-16T18:48:32.87781086Z
---

## Context

- paths: `internal/parse/decl.go` (`:context` reading in `parseSubDecl`/`parseTypedSignature`), `internal/parse/tpm_test.go`
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found by the alignment gate on 01a0fd9b-4ca3-7d55-82ac-ba125816ee71; none blocks RFC 0001 ":context(...)" being Implemented, but each is a silent or misleading edge:

1. A well-formed `:context` on a declaration with no typed signature (`sub f :prototype($) :context(@);`) is accepted and recorded nowhere -- no error, no signature.
2. A repeated `:context` keeps the FIRST silently: `:context($) :context(@)` reads as scalar only. perl keeps the LAST of a repeated `:prototype` (`:prototype($) :prototype(@)` gives `@`).
3. `:context($ @)` is refused with an error naming a blank character.

## Steps

- [ ] Measure perl's handling of repeated attributes; decide each case (likely: 1 is an error, 2 is an error or keeps the last as perl does, 3's message names the space); failing tests first; implement; `make test`; commit

## Acceptance criteria

- [ ] the three :context edge cases report clearly (`go test ./internal/parse/ -run '^TestTpmContextAttributeEdges$' -count=1 -v | grep -q '^--- PASS: TestTpmContextAttributeEdges'`)
