---
title: Have internal/infer read builtin and operator types from CORE.pmt and remove internal/types' builtin table
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-5468-7355-a22e-1b16fa5ee354
- 01a0fd9b-519d-7b42-886e-2222a3076986
- 01a10177-4e7b-7d2c-8d2a-2033647ffc4a
blocks:
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
- 01a0fd9b-55cb-70cd-a9da-929190cb581c
created: 2026-10-02T17:13:30.992763407Z
updated: 2026-10-06T10:31:00.732877718Z
---

## Prerequisites

- [ ] Type CORE.pmt's array, list and hash builtins
- [ ] Declare grep, map and sort as multis in CORE.pmt, with sort's invocant slot
- [ ] Declare the forking operators .. and x as multis in CORE.pmt

## Context

- paths: `internal/infer/infer.go`, `internal/infer/core_test.go` (new), `internal/parse/declarations.go`, `internal/types/signatures.go` (deleted), `internal/types/signatures_test.go` (deleted), `internal/types/signatures_audit_test.go` (deleted), `docs/specs/perl-parser/04-expressions.md`
- commands: `go test ./internal/infer/ -count=1`, `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: the lookups `types.GetBinaryOp`/`GetUnaryOp`/`GetBuiltin` at `internal/infer/infer.go:432`, `:604`, `:652`, `:693`, `:881`, `:1009`, `:3944`, `:4004`, `:4048`; `coreTable` (`internal/parse/declarations.go:39`)

The builtin-type fold of RFC 0001 "Typed Perl, in `.pmt` only": `internal/types/signatures.go` (40 typed builtins, `binaryOps`, `unaryOps`) folds into CORE.pmt, whose lines now carry the same types (held by the CORE type tests). Those 40 rows are typed by the measured-rows issue, the invocant-colon issue (`print`, `say`), the no-prototype builtins issue, the `grep`/`map`/`sort` issue and the array, list and hash batch, so this issue follows those and not the remaining typing batches: it needs the 40 rows, not all 188 lines, and a builtin CORE.pmt has not typed yet was untyped in signatures.go too. internal/infer looks builtins and operators up in CORE.pmt by name through an exported accessor in `internal/parse` (infer does not import it yet; `internal/parse` does not import infer, so there is no cycle). This changes only where lookups come from: infer still walks tree-sitter until m2 retires it (01a0c115), and m2 is also editing `infer.go`, so expect merge friction and keep the diff to the lookup sites. Where CORE.pmt is more precise than signatures.go, measure and update the expectation. `signatures_audit_test.go` holds measurements; carry any not already in the CORE tests into `internal/infer/core_test.go` before deleting it. The parser's keyword table folds after this issue (the keyword-shape issue), and that issue re-marks the section once both tables are gone.

From 01a0fd9b-4f8d-71ca-bac4-24eceb1dc4a4: 12 of signatures.go's 25 measured rows disagree with perl 5.42 and CORE.pmt now carries perl's values -- push, unshift and join need one argument, not two (`push(@a)`, `join(":")` compile); abs, int, uc, lc, ucfirst, lcfirst and reverse need none (`abs()`, `reverse()` compile); substr's fourth argument is a Str; scalar's argument is Scalar. Until infer reads CORE.pmt, it still uses signatures.go's MinArity and flags valid calls such as `push @a` and `uc()`. When switching infer over, add a test that these calls are not flagged.

## Steps

- [ ] Write failing test `TestInferTypesBuiltinsFromCore` in `internal/infer/core_test.go`: a builtin CORE.pmt types that signatures.go does not (pick one, measured) is typed by infer
- [ ] Run test to verify failure: `go test ./internal/infer/ -run '^TestInferTypesBuiltinsFromCore$' -count=1 -v`
- [ ] Add the CORE.pmt accessor in `internal/parse`; switch the builtin lookups; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestInferTypesOperatorsFromCore`: an operator result typed from its CORE.pmt declaration, including `x` with a parenthesised left operand
- [ ] Switch the operator lookups; run test to verify pass; `make test`; commit
- [ ] Carry remaining audit measurements over; delete signatures.go and its tests; `make test`; commit
- [ ] Doc: repoint the `Range` row of `docs/specs/perl-parser/04-expressions.md` §4.14.2 from `signatures.go:182` to `..`'s declaration in `CORE.pmt`; commit

### Negative Scenarios

- a call to a name CORE.pmt does not declare gets no builtin type (no fallback to a stale or default entry) (`go test ./internal/infer/ -run '^TestInferUndeclaredNameHasNoCoreType$' -count=1 -v | grep -q '^--- PASS: TestInferUndeclaredNameHasNoCoreType'`)
- lossy coercions are still reported through the CORE.pmt lookup: `bless "x"` (no coercion to Ref) and a `Str` where `Num` is declared still produce a coercion mismatch (`go test ./internal/infer/ -run '^TestInferCoreCoercionMismatchStillReported$' -count=1 -v | grep -q '^--- PASS: TestInferCoreCoercionMismatchStillReported'`)
- a parenthesised `x` in scalar context is typed `Str` (measured `my $x = (1,2) x 2` gives `22`), not `List` (`go test ./internal/infer/ -run '^TestInferRepeatScalarContextIsStr$' -count=1 -v | grep -q '^--- PASS: TestInferRepeatScalarContextIsStr'`)

## Acceptance criteria

- [ ] infer types builtin calls from CORE.pmt (`go test ./internal/infer/ -run '^TestInferTypesBuiltinsFromCore$' -count=1 -v | grep -q '^--- PASS: TestInferTypesBuiltinsFromCore'`)
- [ ] infer types operators from CORE.pmt (`go test ./internal/infer/ -run '^TestInferTypesOperatorsFromCore$' -count=1 -v | grep -q '^--- PASS: TestInferTypesOperatorsFromCore'`)
- [ ] the builtin table in internal/types is gone (`sh -c '! grep -rqwE "GetBuiltin|GetBinaryOp|GetUnaryOp" --include="*.go" internal'`)
- [ ] infer gives no builtin type to an undeclared name (`go test ./internal/infer/ -run '^TestInferUndeclaredNameHasNoCoreType$' -count=1 -v | grep -q '^--- PASS: TestInferUndeclaredNameHasNoCoreType'`)
- [ ] coercion mismatches survive the switch to CORE.pmt (`go test ./internal/infer/ -run '^TestInferCoreCoercionMismatchStillReported$' -count=1 -v | grep -q '^--- PASS: TestInferCoreCoercionMismatchStillReported'`)
- [ ] infer types a parenthesised x in scalar context as Str (`go test ./internal/infer/ -run '^TestInferRepeatScalarContextIsStr$' -count=1 -v | grep -q '^--- PASS: TestInferRepeatScalarContextIsStr'`)
- [ ] the parser spec no longer cites the deleted builtin table (`sh -c '! grep -q "signatures\.go" docs/specs/perl-parser/04-expressions.md'`)
