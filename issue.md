---
title: Add IO, FileHandle and Boolean to the type lattice, with lookup by name
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T17:13:25.182848094Z
updated: 2026-10-02T17:13:25.182848094Z
---

## Prerequisites

- none

## Context

- paths: `internal/types/types.go`, `internal/types/names.go` (new), `internal/types/names_test.go` (new), `internal/types/types_test.go`, `internal/types/lattice_oracle_test.go`, `internal/types/paper_alignment_test.go`, `internal/infer/` (every `types.Bool` use), `internal/data/help/types.md`
- commands: `go test ./internal/types/ -count=1`, `go test ./internal/infer/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `internal/types/types.go:17-160` (leaf bits, masks, `typeNames`, `allLeafBits`)

RFC 0001 "Type names" (perigrin, 2026-10-02): `IO` joins the lattice as a leaf (the paper has it), `FileHandle` is a named union `Glob|GlobRef|IO` defined as `Ref` and `Scalar` are, and the lattice's `Bool` becomes the paper's `Boolean` (Go identifier and printed name). A `.pmt` names types by string, so the lattice needs a lookup from a name or an `A|B` union to a `Type`, with an error for an unknown name. This is lattice work only; the parser consumes it in the next issue. Bits 19-30 are free; check where `IO` sits under `Any` against `~/dev/types-paper/2-perl-types-formal.md`. Measured on 5.42 per the RFC: a bareword handle is a `Glob`, `open my $fh` gives a `GLOB` ref, `*STDOUT{IO}` is an IO object; a string naming a handle is not a `FileHandle`.

## Steps

- [ ] Write failing test `TestIOLeafAndFileHandleUnion` in `internal/types/names_test.go`: `IO` is a distinct leaf, `FileHandle == Glob|GlobRef|IO`, `FileHandle.String() == "FileHandle"`, `IsSubtype(IO, FileHandle)`
- [ ] Run test to verify failure: `go test ./internal/types/ -run '^TestIOLeafAndFileHandleUnion$' -count=1 -v`
- [ ] Add the `IO` leaf, the `FileHandle` mask, their `typeNames`/`allLeafBits` entries; update the leaf-enumerating tests (`TestBitsetLeafTypesAreDistinct`, lattice oracle, paper alignment) to include `IO`
- [ ] Run test to verify pass; `make test`; commit
- [ ] Write failing test `TestBooleanIsThePapersName`: `Boolean.String() == "Boolean"`
- [ ] Rename `Bool` to `Boolean` in `internal/types` and every consumer (`internal/infer/*.go`, help text in `internal/data/help/types.md`), updating expected strings in tests
- [ ] Run test to verify pass; `make test`; commit
- [ ] Write failing test `TestTypeFromName`: every name in `typeNames` (plus `FileHandle`, `Boolean`) round-trips; `Str|Undef` gives `Str|Undef`; an unknown name (`Strng`) and a malformed union (`Str|`) return an error
- [ ] Implement `FromName(string) (Type, error)` in `internal/types/names.go` (ABOUTME header)
- [ ] Run test to verify pass; `make test`; commit

## Acceptance criteria

- [ ] IO is a leaf and FileHandle is Glob|GlobRef|IO (`go test ./internal/types/ -run '^TestIOLeafAndFileHandleUnion$' -count=1 -v | grep -q '^--- PASS: TestIOLeafAndFileHandleUnion'`)
- [ ] the lattice names Boolean as the paper does (`go test ./internal/types/ -run '^TestBooleanIsThePapersName$' -count=1 -v | grep -q '^--- PASS: TestBooleanIsThePapersName'`)
- [ ] a type name or A|B union resolves, and an unknown name is an error (`go test ./internal/types/ -run '^TestTypeFromName$' -count=1 -v | grep -q '^--- PASS: TestTypeFromName'`)
- [ ] no Go code still names types.Bool (`sh -c '! grep -rqw types.Bool internal cmd'`)
