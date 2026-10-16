---
title: Sub bodies and bare blocks are scoped as legacy did, not as perl does
state: done
urgency: normal
milestone: v0.1
created: 2026-10-13T00:31:01.879793799Z
updated: 2026-10-16T11:47:40.538148775Z
sessions:
- start_sha: 9ff9be88b6f25ac747648420c2ff2433c1b55456
  end_sha: 9ff9be88b6f25ac747648420c2ff2433c1b55456
  commits: 0
  started_at: 2026-10-16T11:47:40.484085319Z
  ended_at: 2026-10-16T11:47:40.538148775Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T11:47:40.484085319Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T11:47:40.538148775Z
---

## Context

- paths: `internal/infer/lowered.go` (`collectLoweredDeclarations`, `loweredAnalysis.walk`), legacy `internal/infer/infer.go`
- commands: `go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$'`

Found 2026-10-13 by the AnalyzeLowered issue (01a12a85-304b), which copied legacy's scoping on purpose so the goldens would match:

- pass 2 walks a named sub's body in a fresh scope, so `sub f { my $x = 1 }` leaves `$x` Scalar where its initialiser makes it Int;
- a bare block's body is not a scope in pass 1, though perl gives `{ my $x }` its own lexical scope.

- `for my $i ($n)` scopes `$_` rather than `$i` (found by the flow port, 01a12a85-3100, which copied it so `flow/` goldens match).
These make PSC report less than perl knows. Fix them on the lowered path only after the m2 cutover (01a0c115-df67), so the ratchet is no longer holding the lowered path to legacy; each golden it changes gets a `Diagnostics-changed:` trailer.

## Steps

- [ ] Failing test first for each; one scope per sub body and bare block, as perl has; edit goldens under trailers; precision must not fall; `make test`; commit

## Acceptance criteria

- [ ] a `my` in a named sub narrows to its initialiser (`go test ./internal/infer/ -run '^TestSubBodyDeclNarrows$' -count=1 -v | grep -q '^--- PASS: TestSubBodyDeclNarrows'`)
- [ ] a bare block scopes its lexicals (`go test ./internal/infer/ -run '^TestBareBlockIsAScope$' -count=1 -v | grep -q '^--- PASS: TestBareBlockIsAScope'`)
- [ ] a `for my $i` loop scopes `$i` (`go test ./internal/infer/ -run '^TestForMyScopesItsVariable$' -count=1 -v | grep -q '^--- PASS: TestForMyScopesItsVariable'`)
