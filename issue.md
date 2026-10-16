---
title: Consecutive comment lines lex in quadratic time
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T03:47:25.62643726Z
updated: 2026-10-16T03:47:25.62643726Z
---

## Context

- found by: the PR #486 review's parser fixes (2026-10-16)
- commands: `go test ./internal/lexer/`; perl 5.42 `/home/perigrin/.local/bin/perl`

internal/lexer `peekIsOpenBrace` runs on every comment token and scans past all the comments that follow, so a run of consecutive comment lines is O(n^2): 10k comment lines take 6s to lex and 40k take 63s. The LSP re-lexes on every keystroke. Limit the lookahead to the next non-trivia token.

## Steps

- [ ] Failing test first; fix against perl 5.42; the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] lexing comment runs is linear (`go test ./internal/lexer/ -run '^TestCommentRunLexIsLinear$' -count=1 -v | grep -q '^--- PASS: TestCommentRunLexIsLinear'`)
