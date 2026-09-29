---
title: "A CORE::-prefixed keyword is not its keyword: CORE::state $x refuses and CORE::length $s drops its argument"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T03:08:00.827847771Z
updated: 2026-09-29T03:08:00.827847771Z
---

`CORE::X` is the keyword X whatever else is in scope, and the parser matched the
whole word against its keyword tables, so a CORE::-prefixed keyword refused --
or, for a named unary, became a call with no argument beside a stray term:

    CORE::state $x = 42;           Unknown=1
    my $n = CORE::length $s;       canon  my $n = CORE::length();$s;   Unknown=0

Measured on 5.42.0, Deparse writes each as the bare keyword:

    $ perl -MO=Deparse -e 'use feature "state","say"; CORE::state $x = 42;
          CORE::say "ok"; my $n = CORE::length $s; CORE::my $y;'
    state $x = 42;
    say 'ok';
    my $n = length $s;
    my $y;

The prefix must stay in the tree and in canon: `CORE::say` works without the
`say` feature, where bare `say` does not.

Reach: 8 dirty perl.git t/ files first refuse at one -- CORE::state in op/sub.t
and op/svleak.t, CORE::say in io/say.t, CORE::require in comp/require.t.

## Acceptance criteria

- [ ] CORE::-prefixed keywords classify as their keywords and canon with the prefix kept (`go test ./internal/parse/ -run '^TestCorePrefixedKeyword$' -count=1 -v | grep -q '^--- PASS: TestCorePrefixedKeyword'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
