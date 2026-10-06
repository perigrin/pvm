---
title: "A \\$ slot still accepts non-lvalue expressions perl refuses"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T11:12:05.947779933Z
updated: 2026-10-06T11:12:05.947779933Z
---

## Context

- paths: `internal/parse/refscalar.go` (`ref_scalar_slot`), `docs/rfcs/0001-type-libraries.md`
- commands: `make test`; ratchets; perl 5.42

From the alignment gate on 01a10dc1-3530-762f-9a03-177445629c6b. The `\$`-slot refusal is deliberately conservative: it refuses literals, aggregates, the wrong arity and a declared non-lvalue sub's result. perl 5.42 also refuses, at compile time, `-1`, `__FILE__`, `v1.2`, `use constant` names, an imported module sub's result (`POSIX::floor(1)`), builtins (`time`, `ord`), `$x+1`, `\$x`, `$x++`, interpolated strings and `do {}` -- but pvm parses these. RFC 0001 says "Where perl refuses ... the parser refuses it too", so this records the deferral the `ponytail:` comment in refscalar.go names only partly (imported subs). A missed refusal keeps perl's parse and adds no false Unknown, so it is safe to defer; a wrong refusal is not.

## Steps

- [ ] For each form, measure with `perl -c`; add the ones that can be decided syntactically (operators, `-1`, `__FILE__`, version strings, `\$x`, `do {}`) with failing tests first; ratchets must not regress
- [ ] Decide `use constant` and imported-sub results once Import carries constant-ness and :lvalue across modules

## Acceptance criteria

- [ ] the syntactic non-lvalue forms are refused in a \$ slot (`go test ./internal/parse/ -run '^TestRefScalarSlotRefusesNonLvalueExpressions$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotRefusesNonLvalueExpressions'`)
