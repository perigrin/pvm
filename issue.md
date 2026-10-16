---
title: "Drop XS::Parse::Infix's LOW and HIGH classes: relations already state those levels"
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T13:39:49.983282202Z
updated: 2026-10-16T18:48:34.899316676Z
---

## Context

- paths: internal/parse/operators.go (`classAnchors`: `"LOW": ""`, `"HIGH": ""`), internal/parse/coreops_test.go (`xpiClasses`), docs/rfcs/0001-type-libraries.md ("Fixity and precedence" class list)
- commands: `make test` (green before every commit)

perigrin, 2026-10-08: drop LOW and HIGH. They exist only because XS::Parse::Infix defines them (its plugin hooks below `or` and above perl's core operators); nothing uses them -- no CORE.tpm line, no library `.tpm` in the repo, and no installed module besides XS::Parse::Infix itself. Under "XS::Parse::Infix is an influence not a limitation", a library places an operator above or below every core operator with relations it can already state (`:tighter(**)`, `:looser(or)`). So `:infix(LOW)` and `:infix(HIGH)` become unknown classes, refused like the retired BITAND.

## Steps

- [ ] Write failing test `TestLowHighClassesRefused`: a library `.tpm` naming `:infix(LOW)` or `:infix(HIGH)` is an unknown-class declaration error
- [ ] Remove them from `classAnchors`, `xpiClasses` and RFC 0001's class list (say the plugin levels are expressed by relations); `make test`; commit

## Acceptance criteria

- [ ] LOW and HIGH are unknown classes (`go test ./internal/parse/ -run '^TestLowHighClassesRefused$' -count=1 -v | grep -q '^--- PASS: TestLowHighClassesRefused'`)
