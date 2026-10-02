---
title: Parse operator declarations and type perl's plain operators in CORE.pmt
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4139-773b-8676-e148e90c42bd
blocks:
- 01a0fd9b-5468-7355-a22e-1b16fa5ee354
created: 2026-10-02T17:13:30.281662921Z
updated: 2026-10-02T17:14:19.682338246Z
---

## Prerequisites

- [ ] Parse typed declarations in .pmt files: typed parameters, defaults and return types

## Context

- paths: `internal/parse/decl.go`, `internal/parse/exports.go`, `internal/parse/declarations/CORE.pmt`, `internal/parse/declarations.go`, `internal/parse/coreops_test.go` (new), `internal/parse/export_test.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `parseSubDecl` (`internal/parse/decl.go:383`), the precedence table `infix`/`prefix` (`internal/parse/precedence.go:56`, `:215`), `internal/types/signatures.go:128-210` (`binaryOps`, `unaryOps`, and their comments)

RFC 0001 "Operator declarations" (provisional spellings): an operator is declared under its symbol, `sub + :infix(ADD) (Num $x, Num $y) Num;`, with `:infix(CLASS)`, `:prefix` or `:postfix`. CLASS is XS::Parse::Infix's vocabulary without `XPI_CLS_` and `_MISC` (`ADD`, `MUL`, `POW`, `LOGICAL_AND`, `LOGICAL_OR`, `ASSIGN`, `LOW`, `HIGH`, `RELATION`, `EQUALITY`, `ORDERING`, `MATCHRE`, `ISA`). No result rule: `Num` is an upper bound that `Int + Int = Int` satisfies; narrowing is inference's. The parser's precedence table stays authoritative, and a test holds each CORE.pmt operator's class to its level there (read-only use of `precedence.go`). Symbol names are read only in `.pmt` mode. Declare the plain (non-forking) operators of signatures.go's `binaryOps` and `unaryOps`, comments included, copying their rows into the test as golden values -- never comparing against signatures.go, which a later issue deletes. `..`/`...` and `x` fork and are the next issue's.

## Steps

- [ ] Write failing test `TestPmtOperatorDeclarationParses`: `sub + :infix(ADD) (Num $x, Num $y) Num;`, `sub ! :prefix (Scalar $x) Boolean;` and a word operator (`sub eq :infix(EQUALITY) ...`) give the name, fixity, class and signature; an unknown class is an error
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtOperatorDeclarationParses$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreOperatorTypesMatchMeasured`: the golden rows of every plain binary and unary operator equal their CORE.pmt declarations
- [ ] Add the declarations and move the comments; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreOperatorClassesMatchPrecedence`: each declared infix operator's class names the level `precedence.go` gives it
- [ ] Fix any disagreeing declaration (the table is authoritative); run test to verify pass; `make test`; commit

## Acceptance criteria

- [ ] operator declarations parse with their fixity and class (`go test ./internal/parse/ -run '^TestPmtOperatorDeclarationParses$' -count=1 -v | grep -q '^--- PASS: TestPmtOperatorDeclarationParses'`)
- [ ] CORE.pmt types the plain operators as measured (`go test ./internal/parse/ -run '^TestCoreOperatorTypesMatchMeasured$' -count=1 -v | grep -q '^--- PASS: TestCoreOperatorTypesMatchMeasured'`)
- [ ] each declared operator class matches the precedence table (`go test ./internal/parse/ -run '^TestCoreOperatorClassesMatchPrecedence$' -count=1 -v | grep -q '^--- PASS: TestCoreOperatorClassesMatchPrecedence'`)
