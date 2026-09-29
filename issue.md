---
title: "A package or class version is not part of the head: `package Foo = 1.0` in canon, and the block after it is lost"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T10:14:08.7025049Z
updated: 2026-09-29T10:26:23.313612571Z
sessions:
- start_sha: 5f106d9d86951f7f7804461574c45246da5500d3
  end_sha: b6e97e516aebc8886a7d447e799c99ffacb8761b
  commits: 1
  started_at: 2026-09-29T10:14:09.303717955Z
  ended_at: 2026-09-29T10:26:23.313612571Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T10:14:09.303717955Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T10:26:23.313612571Z
observed_paths:
- internal/lexer/expect.go
- internal/lexer/lexer.go
- internal/lexer/proto.go
- internal/parse/canon.go
- internal/parse/pkgversion_test.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/use_test.go
---

A package or class version is read as a plain child of the declaration, so two readers get it wrong:

1. Canon joins it with the initialiser's ` = `: `package Foo 1.0;` canonicalises as `package Foo = 1.0;` and `class T6B 1.23 :isa(T6A) {}` as `class T6B = 1.23 ...`, which is not Perl -- at Unknown=0.
2. The lexer closes the declaration's head at the number, so the `{` after a version is not the body: `class Testcase2A 1.23 { }` read an anonymous hash, and `package Bar 1.0 { 1 }` refuses.

Measured with perl 5.42, all compile:

    $ perl -MO=Deparse -e 'package Foo 1.0; package Bar 1.0 { 1 } package main;
          use feature "class"; no warnings; class T6A 1.23 {} class T6B 1.23 :isa(T6A) {}'
    -e syntax OK

perl.git t/class/inherit.t:55 and 110.

## Acceptance criteria

- [ ] A version is part of a package or class head, in canon and before a block (`go test ./internal/parse/ -run '^TestPackageVersionIsHead$' -count=1 -v | grep -q '^--- PASS: TestPackageVersionIsHead'`)
