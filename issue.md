---
title: "Lower literals, names and Hole: NumLit, StrLit, Interp, QwList, Bareword"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
- 01a0c10e-6a8d-721d-be45-a07ae1ee2033
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-09-20T23:03:39.376749228Z
updated: 2026-10-11T10:32:00.350336441Z
---

## Context

- paths: `internal/lower/literals.go` (new), `internal/lower/literals_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerLiterals$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Literals and names"; §4.14.1 rule 3

First slice of the lowering rules. It goes first because it is the smallest,
and it sets the walk shape the other slices copy: one expression-lowering
function that dispatches on `parse.Kind` and returns a `Hole` for any kind no
slice handles yet. Slice tests lower expression nodes directly. The
file-level entry, `lower.Lower`, belongs to the statements issue. Each row's status comes
from §4.14.2 as the status-table issue (01a0c10e-6a8d) re-measured it.
Re-measure every fragment below at execution time. The shapes quoted here
were measured at 24bfac4a.

- `NumLit` SHIPPED: `term "42"`, `term "0x1f"`, `term "1_000"`. `IsInt` comes
  from `Text`.
- `StrLit` SHIPPED: the delimiter is `Text[0]`.
- `StrLit.Backtick` LOWERING: `Text[0] == backtick`. `qx//` splitting stays
  PARSER: `qx/ls/` parses as one `term "qx/ls/"`. The lowering may emit a
  `StrLit` with `Backtick` set and the body unsplit, or a `Hole`. Pick one
  and say why.
- `QwList` LOWERING: `term "qw(a b)"` is opaque, and **§4.14.2 assigns the
  body split to the lowering**. perl: `(qw(a b)) x 2` and `("a","b") x 2`
  both give `a b a b`; scalar `qw(a b c)` and `("a","b","c")` both give `c`.
- `Bareword{Quoted}` LOWERING, by position. §4.14.1 rule 3 lists four
  overloads of an unresolved word. **At 24bfac4a three of the four are now a
  `term`, not a `call`**:

	$h{k}        [index "{" [term "$h"] [term "k"]]            hash key
	(a => 1)     [list paren [call "a" fat] [term "1"]]        fat comma
	Foo->new     [binary "->" [term "Foo"] [term "new"]]       class invocant
	-r _         [call "-r" resolved [term "_"]]                stat buffer

  Separate them by position: second child of an `Index{"{"}`, an element
  with `Fat` set, left operand of `Binary "->"`, operand of a file test.
  Produce `Bareword{Quoted:true}` for the first two. Inside a `list` the
  fat-comma arm reads the `Fat` flag (`parse.go:350`), not a position.
  Unparenthesised, `a => 1;` is `[binary "=>" [call "a"] [term "1"]]` with
  no `Fat` on `a`: the word is the left operand of a `=>`, which the
  operators slice's comma-chain flattening (01a0c111-8c3e) turns into a
  `Fat` mark. That slice tests the unparenthesised spelling; this one does
  not.
- `Interp` PARSER: `"a$b c"` is one `term`; the lexer defers interpolation.
  Emit a `StrLit{Interpolated:true}` with `Parts` empty rather than a `Hole`,
  since the text is intact.
- `Hole` LOWERING: one `Hole` per `Unknown`. A `Hole` is opaque. If the
  Unknown has children (`if ($x) {` gives
  `[unknown refusal=unclosed_opener [block]]`), they are not lowered.

**Hole tests use fragments that are Unknown by construction.** Earlier drafts
used `new Foo(1);`, `print {$fh} "x"` and `&$code(1)`. The last two now parse,
and the first parses whenever `Foo` is a declared package. A Hole test must
not depend on what the parser has not learned yet. Use input perl itself
rejects (confirmed with `perl -c` on 5.42): `qw(a b` with no closer and
`"abc` with no closer. Each slice after this one uses the same rule.

## Steps

- [ ] Write `TestLowerLiterals`: a table of (source, expected lowered shape)
      over each row above, including `qw(a b)` splitting to its words and a
      backtick string setting `Backtick`.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerLiterals$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerBarewordByPosition`: `$h{k}`, `(a => 1)`, `Foo->new` and
      `-r _` each produce the node §4.14.1 rule 3 asks for. Run to verify
      failure, implement, run to green, commit.
- [ ] Write `TestLowerUnknownToHole`: an unterminated `qw(a b` and `"abc`
      each parse to one Unknown over the whole file, and each lowers to
      exactly one `Hole` with the Unknown's span. No panic. (`$x = ;` needs
      the operators slice to lower its `=`, so the operators issue tests
      it.) Run to verify failure, implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Literals lower to their §4.14 kinds with IsInt, Backtick and the qw body split (`go test ./internal/lower/ -run '^TestLowerLiterals$' -count=1 -v | grep -q '^--- PASS: TestLowerLiterals'`)
- [ ] The four overloads of an unresolved word separate by position and flag (`go test ./internal/lower/ -run '^TestLowerBarewordByPosition$' -count=1 -v | grep -q '^--- PASS: TestLowerBarewordByPosition'`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### Negative Scenarios
- [ ] Input perl rejects lowers to a Hole covering the Unknown's bytes, and nothing panics (`go test ./internal/lower/ -run '^TestLowerUnknownToHole$' -count=1 -v | grep -q '^--- PASS: TestLowerUnknownToHole'`)
