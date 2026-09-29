---
title: new Foo builds new(Foo()) where perl builds a method call on the class name
state: done
urgency: normal
milestone: v0.1
created: 2026-09-26T12:51:56.944311859Z
updated: 2026-09-29T05:49:19.033553395Z
sessions:
- start_sha: ef36d1ef0486ea3dddfc702d261fdc29c7b92cb0
  end_sha: 9c48146ef630f3df38870b9dc5d83c4ecb775b6b
  commits: 1
  started_at: 2026-09-29T05:35:16.049019218Z
  ended_at: 2026-09-29T05:49:19.033553395Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T05:35:16.049019218Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T05:49:19.033553395Z
observed_paths:
- internal/conformance/conformance_test.go
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/decl.go
- internal/parse/indirect.go
- internal/parse/indirect_test.go
- internal/parse/parse.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/use.go
---

`new Foo;` builds `new(Foo())` where perl builds `'Foo'->new`. Zero Unknown
nodes and the wrong tree, so nothing in the suite notices. Found while
measuring `01a0d087`, which asserts indirect object notation is unimplemented
-- it is worse than unimplemented, it is silently mis-implemented for the
declared case.

## Measured

Perl, 5.42.0 at `/home/perigrin/.local/bin/perl`:

    perl -MO=Deparse -e 'package Foo; sub new {1} package main; new Foo;'
      -> 'Foo'->new;                     a METHOD call on the string "Foo"

    perl -MO=Deparse -e 'package Foo; sub new {1} package main; new Foo 1, 2;'
      -> 'Foo'->new(1, 2);

Ours, counting Unknown nodes and reading canon:

    new Foo;      (sub new declared in-file)   Unknown=0  canon "new(Foo());"
    new Foo;      (undeclared)                 Unknown=1  canon "new Foo;"
    new Foo 1, 2;                              Unknown=1  canon "new Foo 1, 2;"

So the DECLARED case parses clean and builds a function call whose argument
is a bareword call, where perl builds a method call on a class name. The
undeclared case refuses, which is correct-by-accident: perl also cannot
resolve it, though for a different reason.

## Why it went unnoticed, and why it matters

`01a0ad3f` and `01a0c13f-50ee` taught the parser that a `sub NAME` earlier in
the file is a declaration the call site can see. That is right, and it is
what makes `new Foo;` resolve. But `new` resolving as a FUNCTION is the wrong
resolution when the next token is a bareword that names a package: perl's
rule is that `WORD BAREWORD` is indirect object notation, and the bareword is
the invocant.

Zero Unknowns means every ratchet reports this as clean. It is the same class
of defect as `undef &f` (found under `01a0dd43`, which scored Unknown=0 while
building `binary call call`): a confident wrong tree that round-trips.

## Not urgent, and the reason is measured

Nothing in T2's 56 files writes `WORD BAREWORD` in this shape -- grepped for
`^\s*(new|open)\s+[A-Z]\w*\s*;` across base, comp, cmd, opbasic and class,
zero hits. So it costs no ratchet node today and is latent rather than
active. It will stop being latent the moment the corpus grows a case for it,
which is the right time to fix it.

## The `defer` relationship, which is what I was actually measuring

`01a0d087` says `defer BLOCK ARG` and `new Foo;` need the same rule. Measured,
they do NOT:

    defer { 1 } print "b";     Unknown=1  canon "defer { 1 }print(\"b\");"
    zzz { 1 } print "b";       Unknown=1  canon "zzz { 1 } print \"b\";"
    zzz { 1 };                 Unknown=0  canon "zzz(){1};"

`WORD BLOCK` alone is read as an INDEX CALL -- `zzz(){1}`, a call subscripted
by an anon hash -- which is exactly the mechanism `01a0dd71` fixed for
`ADJUST` and `01a0ddac` records for a bare block in a class body. Three
spellings of one gap: a brace group after a word is not being read as a
block.

`WORD BAREWORD` is a different gap and this issue is it.

## Scope

1. In statement position, `WORD BAREWORD` where the bareword looks like a
   package name is a method call on that name. Measure the boundary first:
   `new Foo::Bar;`, `new $class;`, `print STDERR "x";` (a filehandle, NOT
   this rule) and `open FH, "<", $f;` all wear the same shape and do not all
   mean the same thing.
2. Decide what to do about the DECLARED case specifically, since that is the
   one that currently builds a wrong tree rather than refusing. Refusing
   would be a regression in node count and an improvement in honesty; the
   oracle gate (WRONG = 0) says honesty wins.
3. A corpus case, in tier 11, pinning `'Foo'->new` against `new Foo`.

## Context

- paths: `internal/parse/call.go` (`knowsShape`, the parenless-call path),
  `internal/parse/parse.go` (statement dispatch)
- commands: `go test ./internal/parse/ -count=1`,
  `go test ./internal/parseoracle/ -count=1`
- entrypoint: wherever a resolved bareword callee takes its arguments

## Where the fixtures point here

`internal/conformance/conformance_test.go`'s `TestCorpusSkipsAreDocumented` and
`TestRefusalCodeMismatchFails` need a construct that still refuses, and use
`my $x = new Foo "a";` naming THIS issue. When it lands both fail with "no
longer refuses; replace this fixture", which is the harness working.

## Acceptance criteria

- [ ] `new Foo;` builds a method call on the class name (`go test ./internal/parse/ -run '^TestIndirectObjectNotation$' -count=1 -v | grep -q '^--- PASS: TestIndirectObjectNotation'`)
- [ ] The filehandle shapes are unaffected (`go test ./internal/parse/ -run '^TestIndirectObjectNotation$' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the oracle holds (`go test ./internal/parseoracle/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
