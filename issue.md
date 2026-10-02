---
title: "Parse typed Perl in .pmt files: typed parameters, defaults and return types"
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-97e0-7526-bc54-398af57e9df4
created: 2026-10-02T01:42:09.873559281Z
updated: 2026-10-02T01:42:13.141318811Z
---

RFC 0001, "Typed Perl, in `.pmt` only": `sub NAME :attributes (Type $var, ...) Type;`. perl 5.42 rejects a typed parameter and a return type, so `internal/parse` needs a declaration mode, used only when reading a `.pmt`, that accepts an optional type before each parameter and one after `)`, including type expressions `A|B` and `Name[...]`, and attaches them to the declaration. An unknown type name fails. Ordinary source is unaffected: without the signatures feature `sub f (Ref $x) {...}` is valid Perl whose parentheses are a prototype, and it must still parse that way. Re-mark the RFC section *Implemented*.

## Acceptance criteria

- [ ] Typed declarations parse in a .pmt (`go test ./internal/parse/ -run '^TestTypedDeclarations$' -count=1 -v | grep -q '^--- PASS: TestTypedDeclarations'`)
- [ ] Ordinary source still reads (Ref $x) as a prototype (`go test ./internal/parse/ -run '^TestTypedSyntaxOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestTypedSyntaxOnlyInPmt'`)
- [ ] The RFC section is implemented (`grep -qF '### Typed Perl, in `.pmt` only (*Implemented*' docs/rfcs/0001-type-libraries.md`)
