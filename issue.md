---
title: "A block as a block's first statement reads as a hash: the lexer leaves term position after a block's {"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T22:31:48.966698079Z
updated: 2026-09-28T22:31:48.966698079Z
---

A block that is the FIRST statement of another block reads as an anonymous
hash, so it closes after its own first statement and every brace after it is
off by one:

    sub r {
        {
            $s = 1;
            $x = 2;
        }
    }

canons as `sub r {{$s = 1};$x = 2;} }` with one Unknown. `if (1) { { ... } }`,
`{ { ... } }` and `sub { { ... } }` all do it; a block after any other first
statement does not.

The cause is in the lexer's expect machine. `after(Operator, ...)` returns
XTerm for EVERY operator, including a `{` that opened a block, so the first
token inside a block stands in term position -- and a `{` there is classified
as a hash. perl's yyl_leftcurly sets `PL_expect = XSTATE` after a block brace
(toke.c:6688-6697, the XBLOCK / XSTATE cases). `afterOpenBlock` in lexer.go was
added to patch one symptom of this (a label as a block's first statement) and
its comment says a general fix "would change how every block's first token is
lexed, which is a far larger measurement than this needs". This is that
measurement.

Reach: grep finds a block opening directly after another block's `{` in 31
dirty perl.git t/ files carrying 488 Unknowns -- an upper bound, since not every
Unknown in those files is this. re/pat_advanced.t (99) is the largest.

## Acceptance criteria

- [ ] A block as a block's first statement is a block, in sub, if, bare and anonymous-sub bodies (`go test ./internal/parse/ -run '^TestNestedBareBlock$' -count=1 -v | grep -q '^--- PASS: TestNestedBareBlock'`)
- [ ] The lexer leaves a statement boundary after a block's opening brace (`go test ./internal/lexer/ -run '^TestOpenBlockLeavesStatementState$' -count=1 -v | grep -q '^--- PASS: TestOpenBlockLeavesStatementState'`)
- [ ] NEGATIVE: an anonymous hash as a block's first statement is still a hash, per intuit_curly (`go test ./internal/parse/ -run '^TestNestedBareBlock$' -count=1 -v | grep -q '^--- PASS: TestNestedBareBlock'`)
- [ ] REGRESSION GUARD: a label as a block's first statement still labels (`go test ./internal/parse/ -run '^TestLabelledBareBlock$' -count=1 -v | grep -q '^--- PASS: TestLabelledBareBlock'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
