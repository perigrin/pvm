---
title: "Canon drops the closing parenthesis around a low-precedence word operator: (%h and 1) emits (%h and 1;"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T07:02:15.524283243Z
updated: 2026-10-09T07:49:25.683124327Z
sessions:
- start_sha: d7c31487939158225e0589373b9e583fe7ed61fe
  end_sha: 0f8697ba4cff25a3304a02f6478bb57511bfe8ef
  commits: 1
  started_at: 2026-10-09T07:14:26.203229426Z
  ended_at: 2026-10-09T07:49:25.683124327Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T07:14:26.203229426Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T07:49:25.683124327Z
observed_paths:
- internal/parse/canon.go
- internal/parse/canon_test.go
- internal/parse/parse.go
- internal/parse/refusal.go
- internal/parse/refusal_test.go
- internal/parse/term.go
- internal/parse/unclosed_test.go
---

## Context

- paths: `internal/parse/canon.go` (parenthesisation of low-precedence word operators), `internal/parse/testdata/t1canon.ratchet`
- commands: `make test` (green before every commit; never `--no-verify`); `go test ./internal/parse/ -run '^(TestCanonRatchet|TestCanonFidelityRatchet)$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`
- NEVER use pkill or killall

Found by 01a11f42-2eb8 (an opener unclosed at end of file is refused), 2026-10-09: canon drops a closing parenthesis around a low-precedence word operator -- `(%h and 1);` canonicalises as `(%h and 1;`, which perl refuses ("syntax error ... near "1;""). The defect was hidden: the broken emission happened to be a canon fixpoint, so t1canon counted `stash_var_keyword_op.t` clean; once an unclosed opener at end of file is refused, the emission no longer re-parses clean and the file shows 0 -> 1. Fixing this restores it to 0 honestly.

## Steps

- [ ] Measure perl on `(%h and 1)`, `(%h or 1)`, `(%h xor 1)`, `(not %h)` and the scalar/array forms; find why canon drops the `)`
- [ ] Write failing test `TestCanonKeepsParensAroundWordOperators`: the canon emission keeps the parentheses and perl accepts it; fix; t1canon's stash_var_keyword_op.t stays 0 with 01a11f42-2eb8's refusal in place; `make test`; commit

## Acceptance criteria

- [ ] canon keeps the parentheses around a low-precedence word operator (`go test ./internal/parse/ -run '^TestCanonKeepsParensAroundWordOperators$' -count=1 -v | grep -q '^--- PASS: TestCanonKeepsParensAroundWordOperators'`)
