---
title: "Lower variables and subscripts: Var, Deref, Index, Key with Arrow"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
- 01a0c10e-6a8d-721d-be45-a07ae1ee2033
created: 2026-09-20T23:04:40.094536526Z
updated: 2026-09-20T23:12:06.052903954Z
---

## Context

- paths: `internal/lower/vars.go` (new), `internal/lower/vars_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerVarAndSubscript$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Variables and dereference"; §4.14.1 rules 2 and 7

Status per §4.14.2, corrected at HEAD:

- `Var{Sigil,Name}` SHIPPED — `term "$x"`, `"@a"`, `"%h"`, `"*foo"`, `"$#a"`,
  `"&foo"`, `"$Foo::x"`; the sigil is the leading bytes. `$::x` is still PARSER.
  §4.14.1 rule 7: `(Sigil, Name)` is the identity pair and the sigil is
  load-bearing — `$_`, `@_` and the file-test bareword `_` are three different
  things. This matters concretely: **three paths in `internal/infer` re-derive
  `@_` and `$_[0]` from source text** (`infer.go:3659`, `:3753`, `:4193`). The
  lowering produces `Var{"@","_"}` and `Index{Var{"$","_"}, 0}` so those rules
  compare fields instead.

  The consumer evidence that the sigil is a field and not three kinds: at 8 of 9
  switch sites in `internal/infer` the `scalar`/`array`/`hash` arms are identical
  modulo the sigil string (`infer.go:1669-1673`, `:3560-3564`, `:3851-3855`,
  `:4072-4077`, `:1942-1946`, `:3088-3090`, `:1402-1406`), and
  `lookupNarrowedType` (`:1890-1896`) and `sigildName` (`symbols.go:403`) already
  take the sigil as a parameter.

- `Index`, `Key` SHIPPED. **`Arrow` is no longer PARSER**: `parse.Node.Arrow`
  ships (`parse.go:270`) and is set in the arrow-before-subscript branch
  (`expr.go:83-91`). Verified at HEAD: `$h{k}` gives `index "{"` with no flag and
  `$h->{k}` gives `index "{" arrow`.

  §4.14.1 rule 2 is why it matters, and it is a semantic difference rather than a
  spelling: `$h{k}` reads `%h`, `$h->{k}` reads `$h`. Measured on perl 5.42.0,
  `my %h = (k => "hash"); my $h = { k => "ref" }; print $h{k}, $h->{k};` prints
  `hashref`.

  What inference actually branches on is one step removed — "is the base a
  reference being dereferenced". `inferElementType` (`infer.go:1270-1280`) reads
  the old grammar's `container_variable`-vs-`scalar` split for exactly this, and
  `$$x[0]` is a dereference with no arrow. **So the lowering derives deref-ness
  as `Arrow || base is a Deref`, and keeps `Arrow` itself for the LSP.** Ship
  both; they are not the same predicate.

  `{` selects `Key`, `[` selects `Index`; chains nest (`$x->{a}[0]{b}`).

- `Deref{Sigil,Expr}` PARSER — `${$x}`, `$$x`, `@{$x}`, `@$x`, `%$x`, `$#{$x}`,
  `$#$x` are each one childless `term` (issue 01a0ad52). Emit the best node the
  text supports and a `Hole` for the inner expression rather than guessing; do
  not re-lex the span.

- `Deref{Postfix}` LOWERING for `->@*` / `->%*`: `$x->@*` is
  `[binary "->" [term "$x"] [term "@*"]]`, verified at HEAD. `->$#*` and postfix
  slices stay PARSER.

Out of scope: `Slice{Kind}` is its own issue — it is a five-way switch and the
largest free block in §4.14.2.

## Steps

- [ ] Write `TestLowerVarAndSubscript`: `$x`, `@a`, `%h`, `*foo`, `$#a`, `&foo`,
      `$Foo::x` lower to `Var` with the right `(Sigil, Name)` pair; `$_` and `@_`
      are distinguishable; `$h{k}` and `$h->{k}` lower to `Key` differing in
      exactly `Arrow`; `$h->{a}[0]{b}` nests.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerVarAndSubscript$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerDerefness`: `$h->{k}`, `$$x[0]` and `${$x}[0]` all report
      deref-ness true, and `$h{k}` reports false — the predicate
      `Arrow || base is Deref`, which is what `inferElementType` needs.
      Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerPostfixDeref`: `$x->@*` and `$x->%*` lower to
      `Deref{Postfix:true}` with the right sigil. Run to verify failure,
      implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Var carries (Sigil, Name) and the two subscript spellings differ in exactly Arrow (`go test ./internal/lower/ -run "^TestLowerVarAndSubscript$" -count=1 -v | grep -q "^--- PASS: TestLowerVarAndSubscript"`)
- [ ] Deref-ness is Arrow-or-base-is-Deref, so `$$x[0]` counts and `$h{k}` does not (`go test ./internal/lower/ -run "^TestLowerDerefness$" -count=1 -v | grep -q "^--- PASS: TestLowerDerefness"`)
- [ ] Postfix deref lowers for `->@*` and `->%*` (`go test ./internal/lower/ -run "^TestLowerPostfixDeref$" -count=1 -v | grep -q "^--- PASS: TestLowerPostfixDeref"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)
