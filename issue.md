---
title: "`any( {...} LIST)` refuses: any and all take a block under their features"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T13:33:10.89887846Z
updated: 2026-09-29T13:33:10.89887846Z
---

Under `use feature qw(keyword_any keyword_all)`, `any` and `all` are builtins that take BLOCK LIST as grep does. The parser reads the block only for map, grep and sort, and the lexer carries brace lookahead into `(` only for its block-taking words, so `any( { $_ > 10 } 1 .. 20)` refuses.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'use feature qw(keyword_any keyword_all); no warnings;
          my $x = any( { $_ > 10 } 1 .. 20); my $y = all { $_ < 10 } 1 .. 9;'
    my $x = CORE::any({$_ > 10;} 1..20);
    my $y = CORE::all({$_ < 10;} 1..9);

Without the feature, `use v5.42; any( { $_ > 10 } 1 .. 20)` is a syntax error.

perl.git t/op/any_all.t:19-20.

## Acceptance criteria

- [ ] `any` and `all` take a block under their features, parenthesised or not (`go test ./internal/parse/ -run '^TestAnyAllTakeABlock$' -count=1 -v | grep -q '^--- PASS: TestAnyAllTakeABlock'`)
