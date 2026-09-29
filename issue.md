---
title: "Typed lexical declarations are not read: `my Foo $f` splits into a call and a statement"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:06:00.197280687Z
updated: 2026-09-29T08:06:00.197280687Z
---

`my TYPE VAR` declares a typed lexical: `my Foo $f;` and `my Foo ($g, $h);` are valid when Foo is a package. The parser read the class as a call and split the statement, `my Foo();$f;`, at Unknown=0; it now refuses as trailing tokens.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'package Foo; package main; my Foo $f; my Foo ($g, $h);'
    my Foo $f;
    my Foo ($g, $h);

perl.git t/op/multideref.t:240. perly.y: `myattrterm: MY myterm myattrlist` with the class taken by toke.c's KEY_my branch (`PL_in_my_stash = find_in_my_stash(...)`).

## Acceptance criteria

- [ ] `my Foo $f` is a declaration of $f typed Foo (`go test ./internal/parse/ -run '^TestTypedLexicalDeclaration$' -count=1 -v | grep -q '^--- PASS: TestTypedLexicalDeclaration'`)
