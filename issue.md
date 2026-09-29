---
title: "`->$#*` lexes as `$#` and a multiplication"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T10:27:53.087996019Z
updated: 2026-09-29T10:27:53.087996019Z
---

`->$#*`, the postfix last-index dereference, lexes as the variable `$#` followed by a multiplication, so `(\my @a)->$#*++` refuses and canonicalises as `$# * ++`. The `$#` scanner takes `+`, `-`, `@` or a name after it but not the star that closes a postfix dereference, which `->@*` already gets.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my $r = [1,2]; my $n = $r->$#*; (\my @a)->$#*++;'
    my $n = $r->$#*;
    ++(\my @a)->$#*;

perl.git t/op/array.t:613 and 628; t/op/postfixderef.t.

## Acceptance criteria

- [ ] `$#*` after an arrow is one token (`go test ./internal/lexer/ -run '^TestPostfixLastIndex$' -count=1 -v | grep -q '^--- PASS: TestPostfixLastIndex'`)
- [ ] It parses and its canon is a fixpoint (`go test ./internal/parse/ -run '^TestPostfixLastIndexParses$' -count=1 -v | grep -q '^--- PASS: TestPostfixLastIndexParses'`)
