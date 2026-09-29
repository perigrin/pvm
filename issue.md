---
title: "`${...}` refuses: a deref brace starting with the yada holds a statement"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T21:18:44.495327618Z
updated: 2026-09-29T21:18:44.495327618Z
---

`${...}` refuses. derefHoldsStatements reads the yada as an attempted expression. `...` is a statement and never an expression (perly.y:759, bare_statement_yadayada), so braces that begin with it hold statements.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'eval { ${...}++ };'
    eval { do { ++${die 'Unimplemented';} } };

perl.git t/base/lex.t:504.

## Acceptance criteria

- [ ] `${...}` reads a block (`go test ./internal/parse/ -run '^TestDerefYada$' -count=1 -v | grep -q '^--- PASS: TestDerefYada'`)
