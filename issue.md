---
title: Parse code inside interpolated strings and regexes, and report its sites as perl does
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T11:07:58.353028976Z
updated: 2026-10-09T11:07:58.353028976Z
---

## Context

- paths: `internal/parse/sites.go`, `internal/lexer` (interpolation), `docs/specs/perl-parser/00-findings.md` §0.11
- commands: `make test`; the oracle sweep `PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestCorpusBaselineIsIntact)$' -count=1 -v -timeout 40m -parseoracle.corpus`

The real fix the m1 hedge for interpolation defers to (filed 2026-10-09). perl compiles code inside an interpolated string or regex -- `${\ EXPR}`, `@{[ EXPR ]}`, a postfix deref or slice in a string, `/b${\do {...}}/` -- into ordinary ops, with reference sites, anonymous hashes and calls the oracle counts. internal/parse does not parse inside interpolation, so m1 hedges every marker in such a span (the oracle's `wider`). This issue parses the embedded code and reports its sites as perl does, turning those hedged files from wider to exact. Seen in comp/parser.t, op/coreamp.t, op/postfixderef.t, op/sub_lval.t, re/reg_eval_scope.t.

## Steps

- [ ] Measure perl's ops for each interpolation form; failing test first; parse the embedded code; report its sites; drop the hedge; re-baseline the oracle (every move toward exact); `make test`; commit

## Acceptance criteria

- [ ] code inside interpolation reports perl's sites (`go test ./internal/parse/ -run '^TestInterpolatedCodeReportsItsSites$' -count=1 -v | grep -q '^--- PASS: TestInterpolatedCodeReportsItsSites'`)
