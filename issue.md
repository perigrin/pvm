---
title: Sub bodies and bare blocks are scoped as legacy did, not as perl does
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-13T00:31:01.879793799Z
updated: 2026-10-13T00:31:01.879793799Z
---

## Context

- paths: `internal/infer/lowered.go` (`collectLoweredDeclarations`, `loweredAnalysis.walk`), legacy `internal/infer/infer.go`
- commands: `go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$'`

Found 2026-10-13 by the AnalyzeLowered issue (01a12a85-304b), which copied legacy's scoping on purpose so the goldens would match:

- pass 2 walks a named sub's body in a fresh scope, so `sub f { my $x = 1 }` leaves `$x` Scalar where its initialiser makes it Int;
- a bare block's body is not a scope in pass 1, though perl gives `{ my $x }` its own lexical scope.

Both make PSC report less than perl knows. Fix them on the lowered path only after the m2 cutover (01a0c115-df67), so the ratchet is no longer holding the lowered path to legacy; each golden it changes gets a `Diagnostics-changed:` trailer.

## Steps

- [ ] Failing test first for each; one scope per sub body and bare block, as perl has; edit goldens under trailers; precision must not fall; `make test`; commit

## Acceptance criteria

- [ ] a `my` in a named sub narrows to its initialiser (`go test ./internal/infer/ -run '^TestSubBodyDeclNarrows$' -count=1 -v | grep -q '^--- PASS: TestSubBodyDeclNarrows'`)
- [ ] a bare block scopes its lexicals (`go test ./internal/infer/ -run '^TestBareBlockIsAScope$' -count=1 -v | grep -q '^--- PASS: TestBareBlockIsAScope'`)
