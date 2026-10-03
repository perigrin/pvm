---
title: Add IO, FileHandle and Boolean to the type lattice, with lookup by name
state: pending
urgency: normal
milestone: type-libraries
blocks:
- 01a0fd9b-4139-773b-8676-e148e90c42bd
- 01a0fd9b-49b2-7ea9-bfa1-ba398baadc3f
created: 2026-10-02T17:13:25.182848094Z
updated: 2026-10-03T11:13:44.132103586Z
---

## Prerequisites

- none

## Context

- paths: `internal/types/types.go`, `internal/types/names.go` (new), `internal/types/names_test.go` (new), `internal/types/types_test.go`, `internal/types/lattice_oracle_test.go`, `internal/types/paper_alignment_test.go`, `internal/types/narrowing_test.go`, `internal/types/gotchas_test.go`, `internal/types/coercion_test.go`, `internal/types/signatures_test.go`, `internal/types/coercion.go`, `internal/types/join.go`, `internal/types/narrowing.go`, `internal/types/signatures.go`, `internal/infer/diagnostics.go`, `internal/infer/` (every `types.Bool` use: also `infer.go`, `diagnostics_test.go`, `infer_test.go`, `precision_test.go`), `internal/data/help/types.md` (list from `grep -rlw Bool internal/types` and `grep -rlw types.Bool --include='*.go' internal cmd`)
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
- [ ] Doc: add `IO` and `FileHandle` (a handle: `Glob|GlobRef|IO`) beside `Boolean` in the type list of `internal/data/help/types.md`; `go test ./internal/cli/ -run '^TestShowTypesHelp$' -count=1`; commit

### Negative Scenarios

- `FromName` rejects malformed and unknown spellings without panicking: empty string, whitespace, `|Str`, `Str|`, `Str||Undef`, a union with one unknown member (`Str|Strng`), and the wrong case (`str`); each error names the offending part (`go test ./internal/types/ -run '^TestTypeFromNameRejectsMalformed$' -count=1 -v | grep -q '^--- PASS: TestTypeFromNameRejectsMalformed'`)
- the retired spelling `Bool` is an unknown name once the lattice says `Boolean` (RFC: a `.pmt` writes `Boolean`; an unknown name is an error) (`go test ./internal/types/ -run '^TestTypeFromNameRejectsBool$' -count=1 -v | grep -q '^--- PASS: TestTypeFromNameRejectsBool'`)
- `FileHandle` excludes what the RFC says is not a handle: `IsSubtype(Str, FileHandle)` and `IsSubtype(Object, FileHandle)` are false (a string naming a handle dies under strict, measured; a blessed `IO::File` files under `Object`), and `IsSubtype(FileHandle, IO)` is false (`go test ./internal/types/ -run '^TestFileHandleExcludesStrAndObject$' -count=1 -v | grep -q '^--- PASS: TestFileHandleExcludesStrAndObject'`)

## Acceptance criteria

- [ ] IO is a leaf and FileHandle is Glob|GlobRef|IO (`go test ./internal/types/ -run '^TestIOLeafAndFileHandleUnion$' -count=1 -v | grep -q '^--- PASS: TestIOLeafAndFileHandleUnion'`)
- [ ] the lattice names Boolean as the paper does (`go test ./internal/types/ -run '^TestBooleanIsThePapersName$' -count=1 -v | grep -q '^--- PASS: TestBooleanIsThePapersName'`)
- [ ] a type name or A|B union resolves, and an unknown name is an error (`go test ./internal/types/ -run '^TestTypeFromName$' -count=1 -v | grep -q '^--- PASS: TestTypeFromName'`)
- [ ] no Go code still names types.Bool (`sh -c '! grep -rqw types.Bool internal cmd'`)
- [ ] malformed or partly unknown type names are errors, not panics (`go test ./internal/types/ -run '^TestTypeFromNameRejectsMalformed$' -count=1 -v | grep -q '^--- PASS: TestTypeFromNameRejectsMalformed'`)
- [ ] the retired name Bool is no longer accepted (`go test ./internal/types/ -run '^TestTypeFromNameRejectsBool$' -count=1 -v | grep -q '^--- PASS: TestTypeFromNameRejectsBool'`)
- [ ] a string or a blessed object is not a FileHandle (`go test ./internal/types/ -run '^TestFileHandleExcludesStrAndObject$' -count=1 -v | grep -q '^--- PASS: TestFileHandleExcludesStrAndObject'`)
- [ ] pvm help types names IO, FileHandle and Boolean, and not Bool (`sh -c 'grep -qw IO internal/data/help/types.md && grep -qw FileHandle internal/data/help/types.md && grep -qw Boolean internal/data/help/types.md && ! grep -qw Bool internal/data/help/types.md'`)
