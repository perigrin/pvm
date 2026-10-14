---
title: "my ($x) followed by a comma swallows the list: read $fh, my ($buf), 3 loses its count"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T13:01:23.329883557Z
updated: 2026-10-14T13:01:37.809308499Z
sessions:
- start_sha: 1feeef97410aa0897fde9a5c646b36484162cc2b
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T13:01:37.809308499Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T13:01:37.809308499Z
---

## Context

- paths: `internal/parse/` (declarations: `my`/`our`/`state`/`local` with a parenthesised list), `internal/conformance/` (oracle ratchets)
- commands: `go test ./internal/parse/`; `psc parse --format sexpr`; perl 5.42 `/home/perigrin/.local/bin/perl -MO=Deparse`

Found 2026-10-14 by cutover 1 (01a0c115-a44b): internal/parse reads a parenthesised declaration followed by a comma as a declaration of the whole comma list. `read $fh, my ($buf), 3;` parses as `(declaration "my" (binary "," (term "$buf" paren) (term "3")))` -- the declaration swallows the count -- so psc check reports a false arity ("got 2"). perl reads `my` as taking only its parenthesised list: Deparse gives `read $fh, my $buf, 3;`. Same for `sysread FH, my ($x), 1` and `opendir my ($dh), $d` (perl: `opendir my $dh, '.'`). The parse oracle's WRONG=0 gate did not catch it.

perigrin ruled 2026-10-14: fix now, in parallel, before cutover 1 merges.

## Steps

- [ ] Measure the shapes with Deparse: `f($a, my ($b), $c)`, `my ($b), $c` as a statement, `my ($b, $c), $d`, `our (@x), 1`, `local ($y), 2`, and the unparenthesised `my $b, $c` (already right?)
- [ ] Failing test first in internal/parse: a parenthesised declaration ends at its closing paren
- [ ] Fix; the oracle ratchets (T2, WRONG=0, T1, perl.git, canon, fidelity, corpus, exact) must not regress; add the shapes to the oracle subject's checked set if it has one; `make test`; commit

## Acceptance criteria

- [ ] a parenthesised declaration does not swallow the comma after it (`go test ./internal/parse/ -run '^TestParenDeclarationEndsAtItsParen$' -count=1 -v | grep -q '^--- PASS: TestParenDeclarationEndsAtItsParen'`)
