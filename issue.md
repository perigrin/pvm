---
title: "A declaration followed by an infix operator refuses: my $s =~ /x/ canons as my $s;=~ /x/"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T01:10:55.839984345Z
updated: 2026-09-29T01:10:55.839984345Z
---

A declaration without an initialiser, followed by any infix operator other than
an assignment, refuses:

    my $s =~ /x/;       Unknown=1, canon  my $s;=~ /x/;

perl reads the declaration as the operator's left operand. Measured on 5.42.0,
all three deparse unchanged:

    my $s =~ /x/;
    my $t . 'a';
    our @a == 1;

parseVarDeclNoSemi takes the target, attributes, and an initialiser at the
assignment level, and hands back to the statement only when a word operator
below the comma follows (`atOperatorBelowComma`). Every other infix operator is
left unconsumed and the statement becomes trailing_tokens.

Found writing the test for 01a0ea25-505d (canon dropping a declaration's
parens), where `my $s =~ /x/` was meant as a no-parens-needed control case.

## Acceptance criteria

- [ ] A declaration is the left operand of a following infix operator, and canons to a fixpoint (`go test ./internal/parse/ -run '^TestDeclarationAsLeftOperand$' -count=1 -v | grep -q '^--- PASS: TestDeclarationAsLeftOperand'`)
- [ ] REGRESSION GUARD: an initialiser still binds inside the declaration (`go test ./internal/parse/ -run '^TestParenthesisedDeclarationCanon$' -count=1 -v | grep -q '^--- PASS: TestParenthesisedDeclarationCanon'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
