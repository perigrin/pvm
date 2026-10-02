---
title: Derive prototypes from typed signatures and typed signatures from prototypes, and check they agree
state: cancelled
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-a8d1-77c5-97d8-66cf3b4ec55d
- 01a0fa46-b610-7b2f-ad27-3e6a5365dbeb
- 01a0fa46-9aff-7c99-b17b-7278e810608c
- 01a0fa46-9db1-710e-8f0d-eca8314606ca
blocks:
- 01a0fa46-e28a-7469-998f-ebdd26d18494
- 01a0fa46-fccf-7f53-a476-c02f8f07616d
- 01a0fafc-0dd7-7cd6-bfc3-262c0b8019eb
created: 2026-10-02T01:42:15.435660195Z
updated: 2026-10-02T17:18:12.557205834Z
transitions:
- state: cancelled
  actor: human:git-zhi
  timestamp: 2026-10-02T17:18:12.557205834Z
---

RFC 0001, "A typed signature and a prototype say the same thing": its table maps each prototype character to a parameter type, measured on 5.42. A prototype-only declaration gets the coarse typed signature; a typed one gets its prototype and needs no `:prototype`; one with both must agree, and disagreement is an error in the file. The derived prototypes are held to `prototype("CORE::name")` for all 188 builtins. Re-mark the RFC section *Implemented*.

## Acceptance criteria

- [ ] Derivation both ways, and agreement checked (`go test ./internal/parse/ -run '^TestPrototypeTypeDerivation$' -count=1 -v | grep -q '^--- PASS: TestPrototypeTypeDerivation'`)
- [ ] Derived prototypes equal perl's for every builtin (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypesArePerls'`)
- [ ] The RFC section is implemented (`grep -qF '### A typed signature and a prototype say the same thing (*Implemented*' docs/rfcs/0001-type-libraries.md`)
- [ ] The RFC's opt-out and scalar-container sections are implemented (`sh -c "grep -qF '### Builtins that keep their own parse (*Implemented*' docs/rfcs/0001-type-libraries.md && grep -qF '### The scalar container (*Implemented*' docs/rfcs/0001-type-libraries.md"`)
