---
title: Derive prototypes from typed signatures and typed signatures from prototypes, and check they agree
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T01:42:15.435660195Z
updated: 2026-10-02T01:42:15.435660195Z
---

RFC 0001, "A typed signature and a prototype say the same thing": its table maps each prototype character to a parameter type, measured on 5.42. A prototype-only declaration gets the coarse typed signature; a typed one gets its prototype and needs no `:prototype`; one with both must agree, and disagreement is an error in the file. The derived prototypes are held to `prototype("CORE::name")` for all 188 builtins. Re-mark the RFC section *Implemented*.

## Acceptance criteria

- [ ] Derivation both ways, and agreement checked (`go test ./internal/parse/ -run '^TestPrototypeTypeDerivation$' -count=1 -v | grep -q '^--- PASS: TestPrototypeTypeDerivation'`)
- [ ] Derived prototypes equal perl's for every builtin (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypesArePerls'`)
- [ ] The RFC section is implemented (`grep -qF '### A typed signature and a prototype say the same thing (*Implemented*' docs/rfcs/0001-type-libraries.md`)
