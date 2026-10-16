---
title: "The lowered arity check is unreachable: the parser refuses what it would report"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-14T04:00:51.479006583Z
updated: 2026-10-16T12:09:59.740111522Z
---

## Context

- paths: `internal/infer/calls.go:49` (`CodeArityMismatch`), `internal/parse/` (arity refusal, RFC 0001), `internal/types/CORE.pmt`
- commands: `go test ./internal/infer/`

Found 2026-10-14 by the test-suite port (01a12a85-3417): no parsed input reaches the lowered arity check. internal/parse refuses a builtin call perl refuses at compile time (`push();` -- perl: "Not enough arguments for push"), so it lowers to a Hole; ~30 under-arity builtins tried, including `system()`, drew no arity diagnostic. The calls port (01a12a85-32cf) found the only remaining path is unfeatured `any()`/`all()`, which 01a13793 (any/all only when featured) routes to user subs -- after which the check is dead.

Decide after the cutover (01a0c115-df67): delete `CodeArityMismatch` and its check, or name an input that reaches it. The parser's refusal is the arity diagnostic now; confirm psc check surfaces it (01a0c115-a44b's negative scenario).


**Resolved 2026-10-16.** The check is reachable: `sub getgrgid { "u" } getgrgid()` -- perl 5.42 resolves it to CORE::getgrgid and refuses it ("Not enough arguments for getgrgid"), the parser reads it as the user sub and does not refuse it, so the arity check reports it. `TestArityCheckReachable` (internal/infer/calls_test.go, from review I28) pins it. Per perigrin's I28 ruling, a parser refusal perl always refuses is reported by psc check instead; the check stays for what the parser lets through. The criterion named `lowered_calls.go`, which no longer exists, so it passed vacuously; it now names calls.go.

## Steps

- [ ] Search CORE.pmt for any builtin the parser does NOT refuse at too few/many args yet perl does; if none, delete the check and its code, with a test that the parser's refusal reaches psc check's output; `make test`; commit

## Acceptance criteria

- [ ] no arity-mismatch code remains, or a named test reaches it (`! grep -q 'CodeArityMismatch' internal/infer/calls.go || go test ./internal/infer/ -run '^TestArityCheckReachable$' -count=1 -v | grep -q '^--- PASS: TestArityCheckReachable'`)
