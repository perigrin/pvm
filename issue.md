---
title: "Parse typed Perl in .pmt files: typed parameters, defaults and return types"
state: cancelled
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-97e0-7526-bc54-398af57e9df4
blocks:
- 01a0fa46-b610-7b2f-ad27-3e6a5365dbeb
- 01a0fa46-be8b-7823-9c6c-5be1406b57b8
- 01a0fa46-ece2-7c78-87ed-84da0b5d65f9
- 01a0fa46-fccf-7f53-a476-c02f8f07616d
- 01a0fa47-146b-7047-84b6-20ac55c1b0f6
- 01a0fafc-0dd7-7cd6-bfc3-262c0b8019eb
created: 2026-10-02T01:42:09.873559281Z
updated: 2026-10-16T18:48:30.284832388Z
transitions:
- state: cancelled
  actor: human:git-zhi
  timestamp: 2026-10-02T17:18:12.42397858Z
---

RFC 0001, "Typed Perl, in `.pmt` only": `sub NAME :attributes (Type $var, ...) Type;`. perl 5.42 rejects a typed parameter and a return type, so `internal/parse` needs a declaration mode, used only when reading a `.pmt`, that accepts an optional type before each parameter and one after `)`, including type expressions `A|B` and `Name[...]`, and attaches them to the declaration. An unknown type name fails. Ordinary source is unaffected: without the signatures feature `sub f (Ref $x) {...}` is valid Perl whose parentheses are a prototype, and it must still parse that way. Re-mark the RFC section *Implemented*.

## Acceptance criteria

- [ ] Typed declarations parse in a .pmt (`go test ./internal/parse/ -run '^TestTypedDeclarations$' -count=1 -v | grep -q '^--- PASS: TestTypedDeclarations'`)
- [ ] Ordinary source still reads (Ref $x) as a prototype (`go test ./internal/parse/ -run '^TestTypedSyntaxOnlyInTpm$' -count=1 -v | grep -q '^--- PASS: TestTypedSyntaxOnlyInTpm'`)
- [ ] The RFC section is implemented (`grep -q '^### Typed Perl, in .*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] The RFC's declaration-order and type-names sections are implemented (`sh -c "grep -qF '### Declaration order: Perl's (*Implemented*' docs/rfcs/0001-type-libraries.md && grep -qF '### Type names (*Implemented*' docs/rfcs/0001-type-libraries.md"`)
