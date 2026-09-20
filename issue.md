---
title: "Lower literals, names and Hole: NumLit, StrLit, Interp, QwList, Bareword"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
- 01a0c10e-6a8d-721d-be45-a07ae1ee2033
created: 2026-09-20T23:03:39.376749228Z
updated: 2026-09-20T23:11:58.767757266Z
---

## Context

- paths: `internal/lower/literals.go` (new), `internal/lower/literals_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerLiterals$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Literals and names"; §4.14.1 rule 3

First slice of the lowering rules, chosen first because it is the smallest and
establishes the walk shape the other five slices copy. Status per §4.14.2, as
re-measured by the status-table issue:

- `NumLit` SHIPPED -- `term "42"`, `term "0x1f"`, `term "1_000"`; `IsInt` from `Text`
- `StrLit` SHIPPED -- delimiter is `Text[0]`
- `StrLit.Backtick` LOWERING -- `Text[0] == backtick`. `qx//` stays PARSER
  (`internal/lexer/quote.go:22-31` has no `qx` row); at HEAD `qx/ls/` parses as
  one `term "qx/ls/"`, so the lowering may emit a `StrLit` with `Backtick` set
  and the body unsplit, or a `Hole`. Pick one and say why.
- `QwList` LOWERING -- `term "qw(a b)"` is opaque and **§4.14.2 assigns the body
  split to the lowering**. perl: `(qw(a b)) x 2` and `("a","b") x 2` both give
  `a b a b`; scalar `qw(a b c)` and `("a","b","c")` both give `c`.
- `Bareword{Quoted}` LOWERING -- positional over `Call{Resolved:false}`, which
  §4.14.1 rule 3 shows is overloaded four ways: an unresolved call; a hash key
  (`$h{k}` -> `[index "{" [term "$h"] [call "k"]]`); the left of a fat comma;
  a class name as invocant; and the stat-buffer bareword (`-r _`). Separate them
  by position -- second child of an `Index{"{"}`, left operand of `Binary "=>"`,
  left operand of `Binary "->"`, operand of a file test -- and produce
  `Bareword{Quoted:true}` for the first two.
  **At HEAD the fat-comma case is easier than §4.14.2 says**: `parse.Node.Fat`
  ships and is set on the bareword inside a `list`, so that arm reads a flag
  rather than a position. Verified: `(a => 1, b => 2)` gives a `list` whose two
  `call` children carry `Fat`.
- `Interp` PARSER -- `"a$b c"` is one `term`; the lexer defers interpolation.
  Emit a `StrLit{Interpolated:true}` with `Parts` empty rather than a `Hole`,
  since the text is intact.
- `Hole` LOWERING -- one `Hole` per `Unknown`, per constraint 3 of the milestone.
  Every T1 `Unknown` is a childless leaf (0 with children, 0 nested), so this is
  a leaf-to-leaf rename plus the id.

## Steps

- [ ] Write `TestLowerLiterals`: a table of (source, expected lowered shape)
      over each row above, including `qw(a b)` splitting to its words and a
      backtick string setting `Backtick`.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerLiterals$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerBarewordByPosition`: `$h{k}`, `a => 1`, `Foo->new` and
      `-r _` each produce the node §4.14.1 rule 3 asks for. Run to verify
      failure, implement, run to green, commit.
- [ ] Write `TestLowerUnknownToHole`: a fragment that parses to `Unknown`
      (`new Foo(1);`) lowers to exactly one `Hole` with a span covering the same
      bytes. Run to verify failure, implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Literals lower to their §4.14 kinds with IsInt, Backtick and the qw body split (`go test ./internal/lower/ -run "^TestLowerLiterals$" -count=1 -v | grep -q "^--- PASS: TestLowerLiterals"`)
- [ ] The four overloads of Call{Resolved:false} separate by position (`go test ./internal/lower/ -run "^TestLowerBarewordByPosition$" -count=1 -v | grep -q "^--- PASS: TestLowerBarewordByPosition"`)
- [ ] Each Unknown lowers to one Hole covering the same span (`go test ./internal/lower/ -run "^TestLowerUnknownToHole$" -count=1 -v | grep -q "^--- PASS: TestLowerUnknownToHole"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)
