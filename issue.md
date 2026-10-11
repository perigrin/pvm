---
title: "Lower variables and subscripts: Var, Deref, Index, Key with Arrow"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
blocks:
- 01a0c111-72d3-70fa-88a8-a9a387f04261
created: 2026-09-20T23:04:40.094536526Z
updated: 2026-10-11T10:32:52.367846365Z
---

## Context

- paths: `internal/lower/vars.go` (new), `internal/lower/vars_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerVarAndSubscript$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Variables and dereference"; §4.14.1 rules 2 and 7

Status per §4.14.2 as the status-table issue re-measured it. The shapes below
were measured at 24bfac4a; re-measure at execution time. Line numbers in
`internal/infer` drift, so each citation also names the function.

- `Var{Sigil,Name}` SHIPPED: `term "$x"`, `"@a"`, `"%h"`, `"*foo"`, `"$#a"`,
  `"&foo"`, `"$Foo::x"`, and now `"$::x"` too. The sigil is the leading
  bytes. §4.14.1 rule 7: `(Sigil, Name)` is the identity pair and the sigil
  is load-bearing. `$_`, `@_` and the file-test bareword `_` are three
  different things. Three paths in `internal/infer` re-derive `@_` and
  `$_[0]` from source text. The lowering produces `Var{"@","_"}` and
  `Index{Var{"$","_"}, 0}` so those rules compare fields instead.

  Evidence that the sigil is a field and not three kinds: at 8 of 9 switch
  sites in `internal/infer` the `scalar`/`array`/`hash` arms are identical
  apart from the sigil string, and `lookupNarrowedType` and `sigildName`
  (`symbols.go:403`) already take the sigil as a parameter.

- `Index`, `Key` SHIPPED. `Arrow` ships (`parse.go:330`): `$h{k}` gives
  `index "{"` with no flag, and `$h->{k}` gives `index "{" arrow`. The key is
  `term "k"`.

  §4.14.1 rule 2 is why it matters, and it is semantic: `$h{k}` reads `%h`,
  `$h->{k}` reads `$h`. On perl 5.42.0,
  `my %h = (k => "hash"); my $h = { k => "ref" }; print $h{k}, $h->{k};`
  prints `hashref`.

  What inference branches on is one step removed: "is the base a reference
  being dereferenced". `inferElementType` (`infer.go:1262`) reads the old
  grammar's `container_variable`-vs-`scalar` split for exactly this, and
  `$$x[0]` is a dereference with no arrow. **So the lowering derives
  deref-ness as `Arrow || base is a Deref`, and keeps `Arrow` itself for the
  LSP.** Ship both; they are not the same predicate.

  `{` selects `Key`, `[` selects `Index`; chains nest (`$x->{a}[0]{b}`).

- `Deref{Sigil,Expr}` is **LOWERING now, not PARSER**. At 24bfac4a `${$x}`
  and `$$x` give `unary "$"` over `term "$x"`, `@$x` gives `unary "@"`, and
  `%$x` gives `unary "%"`. Each has a child. Lower the unary to a `Deref`.
  `$#{$x}` and `$#$x` are still one childless `term` each: emit a `Deref`
  with sigil `$#` and a `Hole` for the inner expression. Do not re-lex the
  span.

- `Deref{Postfix}` LOWERING for `->@*`, `->%*` and now `->$#*`: `$x->@*` is
  `[binary "->" [term "$x"] [term "@*"]]`, and `$x->$#*` is the same with
  `term "$#*"`. Postfix slices stay PARSER.

Out of scope: `Slice{Kind}` is its own issue (01a0c111-72d3).

## Steps

- [ ] Write `TestLowerVarAndSubscript`: `$x`, `@a`, `%h`, `*foo`, `$#a`,
      `&foo`, `$Foo::x`, `$::x` lower to `Var` with the right
      `(Sigil, Name)` pair; `$_` and `@_` are distinguishable; `$h{k}` and
      `$h->{k}` lower to `Key` differing in exactly `Arrow`; `$h->{a}[0]{b}`
      nests.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerVarAndSubscript$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerDerefness`: `$h->{k}`, `$$x[0]` and `${$x}[0]` all
      report deref-ness true, and `$h{k}` reports false. Run to verify
      failure, implement, run to green, commit.
- [ ] Write `TestLowerDeref`: `${$x}`, `$$x`, `@$x`, `%$x` lower to `Deref`
      with the right sigil and `Expr`; `$x->@*`, `$x->%*`, `$x->$#*` lower
      to `Deref{Postfix:true}`; `$#{$x}` lowers to a `Deref` whose `Expr` is
      a `Hole`. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerSubscriptOverBrokenInput`: `$x->[;` and `$h{;` (perl
      rejects both) lower to an `Index` / `Key` whose subscript is a `Hole`,
      with the base and `Arrow` intact. No panic. Run to verify failure,
      implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Var carries the sigil and name pair, and the two subscript spellings differ in exactly Arrow (`go test ./internal/lower/ -run '^TestLowerVarAndSubscript$' -count=1 -v | grep -q '^--- PASS: TestLowerVarAndSubscript'`)
- [ ] Deref-ness is Arrow or a Deref base, so a block deref counts and a plain hash key does not (`go test ./internal/lower/ -run '^TestLowerDerefness$' -count=1 -v | grep -q '^--- PASS: TestLowerDerefness'`)
- [ ] Prefix, block and postfix dereferences all lower to Deref (`go test ./internal/lower/ -run '^TestLowerDeref$' -count=1 -v | grep -q '^--- PASS: TestLowerDeref'`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### Negative Scenarios
- [ ] A subscript perl rejects lowers to a Hole inside an intact Index or Key, without a panic (`go test ./internal/lower/ -run '^TestLowerSubscriptOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLowerSubscriptOverBrokenInput'`)
