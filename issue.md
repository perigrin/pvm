---
title: "`${...}` refuses: a deref brace starting with the yada holds a statement"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T21:18:44.495327618Z
updated: 2026-09-29T21:26:29.656225097Z
sessions:
- start_sha: cdce17ba38e34775d6fdb03af3c29de1cf161880
  end_sha: 8d48b3dc1df62144513426d5387e0d1f67fd38b9
  commits: 1
  started_at: 2026-09-29T21:18:45.212747058Z
  ended_at: 2026-09-29T21:26:29.656225097Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T21:18:45.212747058Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T21:26:29.656225097Z
observed_paths:
- internal/parse/derefyada_test.go
- internal/parse/t2_test.go
- internal/parse/term.go
- internal/parse/testdata/perlgitt.ratchet
---

`${...}` refuses. derefHoldsStatements reads the yada as an attempted expression. `...` is a statement and never an expression (perly.y:759, bare_statement_yadayada), so braces that begin with it hold statements.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'eval { ${...}++ };'
    eval { do { ++${die 'Unimplemented';} } };

perl.git t/base/lex.t:504.

## Acceptance criteria

- [ ] `${...}` reads a block (`go test ./internal/parse/ -run '^TestDerefYada$' -count=1 -v | grep -q '^--- PASS: TestDerefYada'`)
