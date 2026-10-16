---
title: A .pmt that declares one name both as sub and as multi sub reads it as a multi, with a warning
state: done
urgency: normal
milestone: type-libraries
created: 2026-10-06T03:27:30.293600081Z
updated: 2026-10-16T18:48:32.481420679Z
sessions:
- start_sha: 07cc780f72630b0802c7e086b4174f04ac531b3f
  end_sha: 07cc780f72630b0802c7e086b4174f04ac531b3f
  commits: 0
  started_at: 2026-10-07T21:47:23.569744893Z
  ended_at: 2026-10-07T21:47:23.627263146Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-07T21:47:23.569744893Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-07T21:47:23.627263146Z
---

## Context

- paths: `internal/parse/decl.go` (signature recording, ~line 233: a plain `sub` deletes the name's entry, `multi sub` appends), `internal/parse/parse.go` (parser state `typedErrs`, `multi`; root `declErrs`), `internal/parse/declarations.go` (`readDeclaration` -> moduleFacts), `internal/parse/module.go` (resolver `declErrs`, `DeclarationErrors(root)`), `internal/parse/exports.go` (`moduleFacts.errs`), `docs/rfcs/0001-type-libraries.md` ("Multi declarations")
- commands: `make test`

Found by the alignment gate on 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c. In a .pmt a plain `sub f` replaces the name's entry and `multi sub f` appends, so mixing the two is order-dependent and silent: `sub f ...; multi sub f ...;` keeps two candidates, `multi sub f ...; sub f ...;` keeps only the last.

**perigrin's ruling (2026-10-07):** "declaring a name as both sub and multi sub should just make it a multi-sub with a warning." A name declared both ways keeps EVERY candidate, in either order, and the declaration file carries a warning naming the sub. A plain `sub f` still replaces an earlier plain `sub f` when no `multi sub f` is declared.

Declarations have an error channel (`typedErrs` -> `moduleFacts.errs` -> resolver `declErrs` -> `DeclarationErrors(root)`) but no warning channel; a warning travels the same route beside it (`DeclarationWarnings(root)`), and a warning does not drop the declaration as an error does.

## Steps

- [ ] Write failing test `TestPmtPlainAndMultiSubMixIsAMultiWithAWarning`: both orders keep both candidates and give exactly one warning `sub f: declared both as sub and as multi sub; read as a multi`, with no error; two plain `sub f` lines still keep only the last, with no warning
- [ ] Implement in decl.go/declarations.go; run test to verify pass
- [ ] Write failing test `TestDeclarationWarningsReachTheRoot`: a module whose .pmt mixes the two, `use`d from source, puts the warning in `DeclarationWarnings(root)` (prefixed with the module name, as errors are) and nothing in `DeclarationErrors(root)`; implement the channel
- [ ] Record the ruling in RFC 0001 "Multi declarations"; `make test`; commit

## Acceptance criteria

- [ ] mixing sub and multi sub for one name makes a multi with a warning, in either order (`go test ./internal/parse/ -run '^TestTpmPlainAndMultiSubMixIsAMultiWithAWarning$' -count=1 -v | grep -q '^--- PASS: TestTpmPlainAndMultiSubMixIsAMultiWithAWarning'`)
- [ ] a declaration file's warnings reach the parse root beside its errors (`go test ./internal/parse/ -run '^TestDeclarationWarningsReachTheRoot$' -count=1 -v | grep -q '^--- PASS: TestDeclarationWarningsReachTheRoot'`)
- [ ] the RFC records the ruling (`grep -q 'multi sub.*warning\|warning.*multi sub' docs/rfcs/0001-type-libraries.md`)
