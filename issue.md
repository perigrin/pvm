---
title: "Lower declarations: Decl kinds, the list form by child count, sub declarations"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
created: 2026-09-20T23:06:52.901032434Z
updated: 2026-09-20T23:11:35.519149056Z
---

## Context

- paths: `internal/lower/decl.go` (new), `internal/lower/decl_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerDeclListForm$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Declarations"; §4.12.2 rule 2

Status per §4.14.2:

- `Decl{Kind}` SHIPPED — `declaration` with `Text` the declarator. T1 census of
  `Declaration.Text`: my 4,905, sub 1,775, package 569, local 398, our 318,
  state 9, field 7, class 5, method 4. `internal/infer` handles `my` only;
  `local`/`state`/`our`/`field` are unhandled. Lower all of them into
  `Decl.Kind` — the lowering's job is to carry the fact, not to type it.
  §4.14.6 leaves the *inference* treatment of `local`/`state`/`our`/`field`
  open, and `local`'s restore-on-scope-exit is a flow fact belonging to §6.1.7,
  which is unbuilt. Out of scope here.

- `Decl{Vars,Init}` LOWERING, and this is the semantically load-bearing part.
  The CST has TWO shapes, verified at HEAD:

	my $x = f()       -> [declaration "my" [term "$x"] [call "f"]]
	                     two children, no "="
	my ($x) = f()     -> [declaration "my" [binary "=" [term "$x" paren] [call "f"]]]
	                     one child, the "="
	my ($x,$y) = f()  -> the same with a list LHS

  **The list-vs-scalar fact §4.12.2 rule 2 needs is the child count.** Measured
  on perl: with `sub f {(1,2,3)}`, `my ($x)=f()` gives 1 and `my $y=f()` gives 3.
  The predicate §4.14.2 states is "one child and it is `binary "="`".

  §4.14.2 recommends normalising the parser to one shape before the lowering is
  written. **That recommendation is declined here and the reason is recorded**:
  it is a parser change with a corpus-wide blast radius, the predicate is two
  field reads, and `parse.Node.Paren` now also marks the parenthesised LHS
  independently — so the lowering has two agreeing signals and needs neither a
  parser change nor a source rescan. If the executor finds the predicate does
  not hold over the corpus, file the parser normalisation rather than widening
  this issue.

- `Decl{Attrs}` LOWERING — `my $x :shared` gives `term ":shared"` as a second
  child.
- `Decl{RefAlias}` LOWERING — `my \$x = $y` gives
  `[declaration "my" [binary "=" [unary "ref" [term "$x"]] [term "$y"]]]`.
  No consumer in `internal/infer`; §4.14.2 marks the consumer SPECULATIVE. Lower
  it because it is nearly free and the shape is unambiguous, but write no
  inference rule for it.
- sub declarations SHIPPED — `sub f { 1 }` is
  `[declaration "sub" [term "f"] [block ...]]`; `sub f;` likewise without a
  block. `internal/infer` uses the old `subroutine_declaration_statement` at 8
  sites; all eight lower to this one shape. Attributes
  (`sub f :prototype($) { 1 }`) are still PARSER — the declaration spans `sub f`
  and the rest is `Unknown`, so emit a `Hole`.
- signatures PARSER — without the feature, `sub f ($x, $y) { 1 }` gives
  `[declaration "sub" [term "f"] [prototype "($x, $y)"] [block ...]]`, which is
  what perl does too. With `use v5.36;` the same text mis-parses (the body is
  read as a hash slice of the signature) because `internal/lexer/proto.go`
  returns false when `l.signatures` is set and nothing downstream picks the
  signature up. `internal/infer` reads `signature` (2) and
  `mandatory_parameter` (2), both unwrapped to their `scalar`. Lower the
  prototype form; emit a `Hole` for the 5.36 mis-parse rather than lowering a
  wrong tree.

- `AnonSub` — **SHIPPED in expression position at HEAD, and the statement-position
  defect is FIXED.** §4.14.2 records `sub { 1 };` producing two statements, a
  wrong tree that round-trips. Verified at HEAD it is now one
  `declaration "sub"` with a `block` child. Re-verify before implementing. In
  expression position `my $f = sub { 1 }` gives
  `[declaration "my" [term "$f"] [declaration "sub" [block ...]]]`. `Sig` stays
  PARSER per the signatures row.

## Steps

- [ ] Write `TestLowerDeclListForm`: `my $x = f()` lowers to a `Decl` with
      `ListAssign` false, `my ($x) = f()` and `my ($x,$y) = f()` to true — by the
      child-count predicate. Assert against the `Paren` flag too, so the two
      signals are shown to agree.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerDeclListForm$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerDeclKinds`: one case per declarator in the T1 census
      (my, sub, package, local, our, state, field, class, method) lowering to
      the right `Decl.Kind`. Run to verify failure, implement, run to green,
      commit.
- [ ] Write `TestLowerSubDeclaration`: `sub f { 1 }`, `sub f;`, `sub { 1 };` and
      `my $f = sub { 1 };` each lower to the shape §4.14 specifies, with
      `sub f :prototype($) { 1 }` producing a `Hole` for the attribute region
      rather than a wrong node. Run to verify failure, implement, run to green,
      commit.
- [ ] Write `TestLowerDeclAttrsAndRefAlias`: `my $x :shared` fills `Attrs`,
      `my \$x = $y` sets `RefAlias`. Run to verify failure, implement, run to
      green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] The list-vs-scalar declaration form is derived from child count (`go test ./internal/lower/ -run "^TestLowerDeclListForm$" -count=1 -v | grep -q "^--- PASS: TestLowerDeclListForm"`)
- [ ] Every declarator in the T1 census lowers to a Decl.Kind (`go test ./internal/lower/ -run "^TestLowerDeclKinds$" -count=1 -v | grep -q "^--- PASS: TestLowerDeclKinds"`)
- [ ] Named, forward, anonymous and expression-position subs all lower (`go test ./internal/lower/ -run "^TestLowerSubDeclaration$" -count=1 -v | grep -q "^--- PASS: TestLowerSubDeclaration"`)
- [ ] Attributes and ref-alias declarations lower (`go test ./internal/lower/ -run "^TestLowerDeclAttrsAndRefAlias$" -count=1 -v | grep -q "^--- PASS: TestLowerDeclAttrsAndRefAlias"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)
