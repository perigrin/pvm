---
title: "`//` after shift, pop, undef and the other UNIDOR builtins opens a pattern instead of defined-or"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T06:45:37.925193784Z
updated: 2026-09-29T06:45:37.925193784Z
---

The lexer never enters XTERMORDORDOR, so after the nine UNIDOR builtins a `//` opens an empty pattern instead of being defined-or. `is(shift // 7, 7)` refuses, and `my $y = shift // 7;` canonicalises to `my $y = shift(//);7;`, a wrong tree at Unknown=0.

toke.c:

    #define UNIDOR(f) UNI3(f,XTERMORDORDOR,1)
    if ((PL_expect == XOPERATOR || PL_expect == XTERMORDORDOR) && s[1] == '/')
        ... AOPERATOR(DORDOR);

UNIDOR covers getc, pop, pos, readline, readpipe (OP_BACKTICK), readlink, shift, undef and umask. Measured with perl 5.42, `my $y = W // 7` is defined-or for all nine; after any other named unary (`lc // 7`) `//` is a pattern.

perl.git t/op/dor.t refuses 3 times on this.

## Acceptance criteria

- [ ] `//` after a UNIDOR builtin is defined-or, and after another named unary a pattern (`go test ./internal/lexer/ -run '^TestDefinedOrAfterUnidor$' -count=1 -v | grep -q '^--- PASS: TestDefinedOrAfterUnidor'`)
- [ ] REGRESSION GUARD: a UNIDOR word as a sub name still opens its body (`go test ./internal/lexer/ -run '^TestUnidorSubNameStillOpensABlock$' -count=1 -v | grep -q '^--- PASS: TestUnidorSubNameStillOpensABlock'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
