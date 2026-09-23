---
title: Ungated defer is declined as a statement form the parser can already read
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-23T23:08:33.884597854Z
updated: 2026-09-23T23:54:39.887673967Z
---

`defer` sits in `statementKeywords` because the GATED form -- a block
that runs at scope exit -- is unimplemented. That entry also declines
the UNGATED reading, which is an ordinary indirect method call the
expression parser can already read.

Found while closing `01a0cf3e`, which tracked four recorded refusals.
Three were one filehandle bug, fixed in `831eaa1d`. This is the fourth
and it is not the same kind of thing.

## Measured

Ungated, `defer BLOCK` is a method call whose arguments are the block's
value and whatever follows:

    $ perl -MO=Concise,-exec,f -e 'sub f { defer { print "D\n" } print "body\n" }'
    2  <0> pushmark s
    ...
    5  <@> print sKM
    ...
    8  <@> print sKM
    9  <.> method_named[PV "defer"]
    a  <1> entersub[t3] KRS/TARG

Both prints run BEFORE the call, which is why the ungated reading
prints `D` first and the gated one prints it last -- the claim
`conformance/11_oo/17_defer_ungated.t` pins.

The expression parser reads this correctly today:

    my $x = defer { 1 } print "b";      parses
    defer { 1 };                        refuses
    sub f { defer { 1 } print "b" }     refuses

Only the statement path refuses, because the `statementKeywords` entry
short-circuits it before the expression parser is reached.

## Why removing the entry is not enough

Removing it fixes the first two shapes and leaves the third, which
refuses with `trailing_tokens`: `WORD BLOCK ARG` at statement level has
no reader. `parseWordTerm` reads the brace group as an anonymous hash
and stops, orphaning `print "body\n"`.

The general form is INDIRECT OBJECT NOTATION, which this parser does
not implement -- measured, `new Foo;` refuses at statement level too,
while `my $o = new Foo;` parses. `listOperator` is the table with the
block-then-list shape, but it is a table of BUILTINS and `defer`
ungated is a user-facing method name, so adding it there would be
recording the wrong fact.

## And a round-trip defect behind it

With the entry removed, two T1 files improve -- `defer.t` 17 refusals
to 11, `defer_implicit_return_cleanup.t` to zero -- and one breaks:

    canon_test.go: defer.t: token 399:
      first emission has "CloseBracket }", second has "Semicolon ;"

`defer.t` stops re-emitting identically. Parsing further into the file
exposes a canonicalisation bug that the refusal was hiding.

TWO ratchets move, not one: `TestCanonRatchet` also fails with
`defer.t: 0 -> 1`. Same underlying defect counted a second way, so it
does not change the conclusion, but both baselines are in scope. That is a
correctness defect in its own right and should be understood before the
parse change lands, not after.

## Scope

Three pieces, in order:

1. The canon round-trip defect on `defer.t`, which is independent of
   any parse change and is currently masked.
2. Indirect object notation at statement level -- `WORD BLOCK ARG` and
   `WORD BAREWORD` -- which is what `new Foo;` needs too.
3. Removing `defer` from `statementKeywords`, and the T1 and canon
   ratchet updates the improvement earns -- both of them.

The GATED `defer` statement is not in scope here. It is a real
statement form with scope-exit semantics and belongs wherever
`05_scoping` grows to cover it; this issue is only about not declining
the ungated reading.

## Acceptance criteria

- [ ] `defer.t` re-emits identically from the tree (`go test ./internal/parse/ -run '^TestCanonTokenIdentity$' -count=1 | grep -q '^ok'`)
- [ ] Ungated `defer BLOCK ARG` parses at statement level (`go test ./internal/parse/ -run '^TestUngatedDeferParses$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] The corpus file stops refusing and pins its output (`go test ./internal/conformance/ -run '^TestCorpus$/11_oo/17' -count=1 -v | grep -q 'PASS: TestCorpus/11_oo/17'`)
- [ ] Both parser ratchets record the movement (`go test ./internal/parse/ -run '^TestParseRatchet$|^TestCanonRatchet$' -count=1 | grep -q '^ok'`)
- [ ] The whole suite is green (`make test`)
