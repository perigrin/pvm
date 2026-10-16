---
title: Parse operator declarations and type perl's plain operators in CORE.pmt
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a10177-49d9-7915-975e-2a8e3da943bf
blocks:
- 01a0fd9b-5468-7355-a22e-1b16fa5ee354
created: 2026-10-02T17:13:30.281662921Z
updated: 2026-10-16T18:48:31.772693109Z
sessions:
- start_sha: b1131799b6f6c1d1db02fc7b204d43b85af37dcc
  end_sha: b1131799b6f6c1d1db02fc7b204d43b85af37dcc
  commits: 0
  started_at: 2026-10-06T04:08:36.482978549Z
  ended_at: 2026-10-06T04:08:36.528218834Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T04:08:36.482978549Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T04:08:36.528218834Z
---

## Prerequisites

- [ ] Parse return types and type expressions in .pmt declarations

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

### Negative Scenarios

- fixity is spelled one way: `:infix` with no class, `:prefix(ADD)`, and two fixities on one line (`:infix(ADD) :prefix`) are errors (`go test ./internal/parse/ -run '^TestPmtOperatorFixityErrors$' -count=1 -v | grep -q '^--- PASS: TestPmtOperatorFixityErrors'`)
- arity follows fixity: an infix with one parameter (`sub + :infix(ADD) (Num $x) Num;`) or a prefix with two is an error (`go test ./internal/parse/ -run '^TestPmtOperatorArityMatchesFixity$' -count=1 -v | grep -q '^--- PASS: TestPmtOperatorArityMatchesFixity'`)
- declaration order holds for operators: `sub + (Num $x, Num $y) :infix(ADD) Num;` is refused (`go test ./internal/parse/ -run '^TestPmtOperatorAttributeAfterSignatureRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtOperatorAttributeAfterSignatureRefused'`)
- symbol names are `.pmt`-only: `sub + { 1 }` in ordinary source declares no operator (`go test ./internal/parse/ -run '^TestOperatorSymbolOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestOperatorSymbolOnlyInPmt'`)
- the class check is not tautological: a deliberately wrong class (`sub * :infix(ADD) ...`) run through the same check is reported (`go test ./internal/parse/ -run '^TestCoreOperatorClassCheckCatchesMismatch$' -count=1 -v | grep -q '^--- PASS: TestCoreOperatorClassCheckCatchesMismatch'`)

## Acceptance criteria

- [ ] operator declarations parse with their fixity and class (`go test ./internal/parse/ -run '^TestTpmOperatorDeclarationParses$' -count=1 -v | grep -q '^--- PASS: TestTpmOperatorDeclarationParses'`)
- [ ] CORE.pmt types the plain operators as measured (`go test ./internal/parse/ -run '^TestCoreOperatorTypesMatchMeasured$' -count=1 -v | grep -q '^--- PASS: TestCoreOperatorTypesMatchMeasured'`)
- [ ] each declared operator class matches the precedence table (`go test ./internal/parse/ -run '^TestCoreOperatorClassesMatchPrecedence$' -count=1 -v | grep -q '^--- PASS: TestCoreOperatorClassesMatchPrecedence'`)
- [ ] a missing class, a class on a prefix, or two fixities is an error (`go test ./internal/parse/ -run '^TestTpmOperatorFixityErrors$' -count=1 -v | grep -q '^--- PASS: TestTpmOperatorFixityErrors'`)
- [ ] an operator whose arity does not match its fixity is an error (`go test ./internal/parse/ -run '^TestTpmOperatorArityMatchesFixity$' -count=1 -v | grep -q '^--- PASS: TestTpmOperatorArityMatchesFixity'`)
- [ ] an operator attribute after the signature is refused (`go test ./internal/parse/ -run '^TestTpmOperatorAttributeAfterSignatureRefused$' -count=1 -v | grep -q '^--- PASS: TestTpmOperatorAttributeAfterSignatureRefused'`)
- [ ] ordinary source declares no operator (`go test ./internal/parse/ -run '^TestOperatorSymbolOnlyInTpm$' -count=1 -v | grep -q '^--- PASS: TestOperatorSymbolOnlyInTpm'`)
- [ ] the precedence check reports a wrong class (`go test ./internal/parse/ -run '^TestCoreOperatorClassCheckCatchesMismatch$' -count=1 -v | grep -q '^--- PASS: TestCoreOperatorClassCheckCatchesMismatch'`)
