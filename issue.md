---
title: "Lower declarations: Decl kinds, the list form by child count, sub declarations"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
- 01a0c10f-9671-733f-8b98-30a78f634b0f
blocks:
- 01a12a85-304b-77f1-9bda-e6b73c317305
created: 2026-09-20T23:06:52.901032434Z
updated: 2026-10-12T17:38:57.52371282Z
sessions:
- start_sha: 9895b3086efb0ed7feb8aeb6b994c6b245ce47a5
  end_sha: 0633986dbaf314741a7e588c665ff89622a80903
  commits: 4
  started_at: 2026-10-11T16:34:02.348401953Z
  ended_at: 2026-10-12T17:38:57.52371282Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T16:34:02.348401953Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-12T17:38:57.52371282Z
observed_paths:
- docs/specs/perl-parser/04-expressions.md
- docs/specs/perl-parser/06-incremental-lsp.md
- internal/lower/decl.go
- internal/lower/decl_test.go
- internal/lower/export_test.go
- internal/lower/expr.go
- internal/lower/literals.go
- internal/lower/literals_test.go
- internal/lower/node.go
- internal/lower/node_test.go
- internal/lower/special.go
- internal/lower/special_test.go
- internal/lower/vars.go
- internal/lower/vars_test.go
---

## Context

- paths: `internal/lower/decl.go` (new), `internal/lower/decl_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerDeclListForm$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Declarations"; §4.12.2 rule 2

Status per §4.14.2 as the status-table issue re-measured it. Shapes below were
measured at 24bfac4a; re-measure at execution time.

- `Decl{Kind}` SHIPPED — `declaration` with `Text` the declarator. T1 census of
  `Declaration.Text`: my 4,905, sub 1,775, package 569, local 398, our 318,
  state 9, field 7, class 5, method 4. `internal/infer` handles `my` only;
  `local`/`state`/`our`/`field` are unhandled. Lower all of them into
  `Decl.Kind` — the lowering's job is to carry the fact, not to type it.
  §4.14.6 leaves the *inference* treatment of `local`/`state`/`our`/`field`
  open, and `local`'s restore-on-scope-exit is a flow fact belonging to §6.1.7,
  which is unbuilt. Out of scope here.

- `Decl{Vars,Init}` LOWERING, and this is the semantically load-bearing part.
  The CST has TWO shapes, measured at 24bfac4a:

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

- `Decl{Attrs}` LOWERING — `my $x :shared` gives an `attribute ":shared"`
  second child (the `Attribute` kind, no longer a `term`).
- `Decl{RefAlias}` LOWERING — `my \$x = $y` gives
  `[declaration "my" [binary "=" [unary "ref" [term "$x"]] [term "$y"]]]`.
  No consumer in `internal/infer`; §4.14.2 marks the consumer SPECULATIVE. Lower
  it because it is nearly free and the shape is unambiguous, but write no
  inference rule for it.
- sub declarations SHIPPED — `sub f { 1 }` is
  `[declaration "sub" [term "f"] [block ...]]`; `sub f;` likewise without a
  block. `internal/infer` uses the old `subroutine_declaration_statement` at 8
  sites; all eight lower to this one shape. **Attributes now parse**:
  `sub f :prototype($) { 1 }` gives
  `[declaration "sub" [term "f"] [attribute ":prototype($)"] [block ...]]`.
  Lower them into the sub's `Attrs`; no `Hole`.
- signatures — without the feature, `sub f ($x, $y) { 1 }` gives
  `[declaration "sub" [term "f"] [prototype "($x, $y)"] [block ...]]`, which
  is what perl does too. **The 5.36 mis-parse is gone**: with `use v5.36;`
  the same text gives `[declaration "sub" [term "f"] [list paren [term "$x"] [term "$y"]] [block ...]]`.
  `internal/infer` reads `signature` (2) and `mandatory_parameter` (2), both
  unwrapped to their `scalar`. Lower the `list paren` form into the sub's
  parameters. Re-measure defaults (`$x = 1`), slurpies and `@_`-free bodies
  before deciding what else the signature form carries; emit a `Hole` for
  any part that does not parse rather than guessing.

- `AnonSub` — **SHIPPED in expression position, and the statement-position
  defect is FIXED.** §4.14.2 records `sub { 1 };` producing two statements, a
  wrong tree that round-trips. At 24bfac4a it is one `declaration "sub"` with
  a `block` child. In
  expression position `my $f = sub { 1 }` gives
  `[declaration "my" [term "$f"] [declaration "sub" [block ...]]]`. `Sig` stays
  PARSER per the signatures row.

**Hole tests use input perl rejects** (confirmed with `perl -c` on 5.42):
`my $x = (;` gives `[declaration "my" [term "$x"] [unknown paren refusal=not_a_term]]`,
and `sub {;` gives `[declaration "sub" [block [unknown refusal=unclosed_opener [statement]]]]`,
an Unknown with a child.

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
      `sub f :prototype($) { 1 }` carrying its attribute, and
      `use v5.36; sub f ($x, $y) { 1 }` carrying two parameters. Run to verify failure, implement, run to green,
      commit.
- [ ] Write `TestLowerDeclAttrsAndRefAlias`: `my $x :shared` fills `Attrs`,
      `my \$x = $y` sets `RefAlias`. Run to verify failure, implement, run to
      green, commit.
- [ ] Write `TestLowerDeclOverBrokenInput`: `my $x = (;` lowers to a `Decl`
      with its `Var` intact and a `Hole` `Init`; `sub {;` lowers to an
      `AnonSub` whose body holds one `Hole`. No panic. Run to verify
      failure, implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] The list-vs-scalar declaration form is derived from child count (`go test ./internal/lower/ -run '^TestLowerDeclListForm$' -count=1 -v | grep -q '^--- PASS: TestLowerDeclListForm'`)
- [ ] Every declarator in the T1 census lowers to a Decl.Kind (`go test ./internal/lower/ -run '^TestLowerDeclKinds$' -count=1 -v | grep -q '^--- PASS: TestLowerDeclKinds'`)
- [ ] Named, forward, anonymous, expression-position and signature subs all lower, attributes included (`go test ./internal/lower/ -run '^TestLowerSubDeclaration$' -count=1 -v | grep -q '^--- PASS: TestLowerSubDeclaration'`)
- [ ] Attributes and ref-alias declarations lower (`go test ./internal/lower/ -run '^TestLowerDeclAttrsAndRefAlias$' -count=1 -v | grep -q '^--- PASS: TestLowerDeclAttrsAndRefAlias'`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### Negative Scenarios
- [ ] A declaration perl rejects lowers with its variable intact and a Hole where the parse failed, and nothing panics (`go test ./internal/lower/ -run '^TestLowerDeclOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLowerDeclOverBrokenInput'`)
