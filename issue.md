---
title: "A block deref loses its braces: ${$pkg . '::'} re-emits as $$pkg . '::'"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-27T00:31:06.168570065Z
updated: 2026-09-27T00:31:06.168570065Z
---

## Measured finding

`${EXPR}` -- the block form of a scalar deref -- loses its braces, so the
emission re-groups around the sigil. Measured at `34737c46` in a clean
`git archive` of HEAD, independent of any in-flight change:

    source                                  first canon                      second canon
    my $x = ${$pkg . '::'};                 $$pkg . '::'                     ($$pkg) . '::'
    my $x = !!${$pkg . '::'}{$1 . '::'};    !!($$pkg . '::'){$1 . '::'}      !!(($$pkg) . '::'){$1 . '::'}
    f(${$pkg . '::'});                      f($$pkg . '::')                  f(($$pkg) . '::')

The braces of `${EXPR}` GROUP -- `term.go:206` says so in its own comment --
and the whole expression inside them is what the sigil applies to. Dropping
them makes `$` apply to `$pkg` alone and leaves `. '::'` as a concatenation
outside the deref, which is a different program. perl, measured on 5.42.0 at
`/home/perigrin/.local/bin/perl`:

    $ perl -MO=Deparse -e 'no strict "refs"; my $pkg = "main"; my $x = ${$pkg . "::"};'
    my $x = ${$pkg . '::';};

perl's own deparser writes the braces back, because there is no brace-free
spelling of this.

A SECOND defect in the same file: `$::{$pkg}` -- the main-stash subscript --
canons as `$:;:{$pkg}`, a `$:` variable followed by a stray `;:`. That one is a
LEXER-level split of the `$::` sigil and is noted here rather than filed apart
because the same corpus file carries both.

## How it was found

Found by fixing 01a0e013, whose canon defect had been DELETING every `return`
operand -- `return EXPR;` canon'd as `return ;`. With the operand emitted, this
became visible. `package_decl_stash.t` in the T1 corpus is the file:

    return !!${$pkg . '::'}{$1 . '::'};
    return !!$::{$pkg . '::'};

It sat in `t1canon.ratchet` at 0 only because BOTH emissions deleted the whole
operand, which made a wrong tree look like a fixpoint. It is `0 -> 1` after
01a0e013 and the ratchet records why.

## Acceptance criteria

- [ ] `my $x = ${$pkg . '::'};` canons with its braces and is a fixpoint `go test ./internal/parse -run TestBlockDerefKeepsItsBraces -count=1`
- [ ] `$::{$pkg}` canons as itself rather than `$:;:{$pkg}` `go test ./internal/parse -run TestBlockDerefKeepsItsBraces -count=1`
- [ ] `package_decl_stash.t` returns to 0 in `t1canon.ratchet` and rejoins `TestCanonTokenIdentity` `go test ./internal/parse -run 'TestCanonRatchet|TestCanonTokenIdentity' -count=1`
- [ ] REGRESSION GUARD: `$$r`, `@{$r}`, `%{$h}` and `${$h}{k}` are unchanged `go test ./internal/parse -run TestDeref -count=1`
- [ ] the whole tree is green `go test ./... -count=1`
