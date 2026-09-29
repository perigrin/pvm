---
title: "`CORE::my` and `CORE::our` refuse in expression position"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T10:40:48.835553402Z
updated: 2026-09-29T10:40:48.835553402Z
---

The statement path names a declarator through keywordName, so `CORE::state $x = 1;` parses, but the term path (parseTerm's declarator branch, and canon's binding power for a declaration) looked the spelling up as written. `f(CORE::my $x)` and `(CORE::our $z = 1) + 1` refused.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'use feature "state"; ok(ref(CORE::state $y = "a"));
          f(CORE::my $x); my $n = (CORE::our $z = 1) + 1;'
    ok(ref(state $y = 'a'));
    f(my $x);
    my $n = (our $z = 1) + 1;

perl.git t/opbasic/concat.t:203.

## Acceptance criteria

- [ ] CORE:: declarators work in expression position (`go test ./internal/parse/ -run '^TestCoreDeclaratorInExpression$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclaratorInExpression'`)
