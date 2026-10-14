---
title: "The lowered arity check is unreachable: the parser refuses what it would report"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-14T04:00:51.479006583Z
updated: 2026-10-14T04:00:51.479006583Z
---

## Context

- paths: `internal/infer/lowered_calls.go:49` (`CodeArityMismatch`), `internal/parse/` (arity refusal, RFC 0001), `internal/types/CORE.pmt`
- commands: `go test ./internal/infer/`

Found 2026-10-14 by the test-suite port (01a12a85-3417): no parsed input reaches the lowered arity check. internal/parse refuses a builtin call perl refuses at compile time (`push();` -- perl: "Not enough arguments for push"), so it lowers to a Hole; ~30 under-arity builtins tried, including `system()`, drew no arity diagnostic. The calls port (01a12a85-32cf) found the only remaining path is unfeatured `any()`/`all()`, which 01a13793 (any/all only when featured) routes to user subs -- after which the check is dead.

Decide after the cutover (01a0c115-df67): delete `CodeArityMismatch` and its check, or name an input that reaches it. The parser's refusal is the arity diagnostic now; confirm psc check surfaces it (01a0c115-a44b's negative scenario).

## Steps

- [ ] Search CORE.pmt for any builtin the parser does NOT refuse at too few/many args yet perl does; if none, delete the check and its code, with a test that the parser's refusal reaches psc check's output; `make test`; commit

## Acceptance criteria

- [ ] no arity-mismatch code remains, or a named test reaches it (`! grep -q 'CodeArityMismatch' internal/infer/lowered_calls.go || go test ./internal/infer/ -run '^TestArityCheckReachable$' -count=1 -v | grep -q '^--- PASS: TestArityCheckReachable'`)
