---
title: A heredoc in statement position parses, body and all
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-20T23:56:10.360319979Z
updated: 2026-09-20T23:56:10.360319979Z
---

A heredoc introducer in statement position sinks its statement.

## Context

- paths: `internal/lexer/` (heredoc scanning), `internal/parse/call.go`,
  `internal/parse/t2_test.go`, `internal/parse/testdata/t1.ratchet`
- commands: `go test ./internal/lexer/ ./internal/parse/ -count=1`
- entrypoints: the lexer's heredoc handling; then whatever consumes the
  introducer token in `internal/parse`.

## Measured, 2026-09-20 at b6e7c351

    print <<'EOF';
    body
    EOF
                                   1 Unknown

`comp/require.t`'s shortfall comment in `internal/parse/t2_test.go` already
records this shape by name: reaching inside a `BEGIN { ... }` block "exposes
a heredoc the parser does not yet read". `base/lex.t` (37) is named in the
survey as heredocs plus leading decimals.

The first step is a SPLIT question: does the lexer emit a heredoc token whose
body runs to the terminator and the parser drops it, or does the lexer not
recognise the introducer at all? The fix lands in a different file each way,
so answer it before changing anything.

## Steps

- [ ] Determine whether the lexer produces a heredoc token for the three-line
      form above, and record the answer here. Measure how many of
      `base/lex.t`'s 37 and `comp/require.t`'s 12 Unknowns are heredocs.
- [ ] Write a failing test: the heredoc statement above parses with no
      Unknown node and round-trips byte-identically
- [ ] Run it and confirm the failure is the one measured (`go test ./internal/parse/ -run '^TestHeredocInStatementPosition$' -count=1 -v`)
- [ ] Implement minimally, in whichever layer the first step named
- [ ] Add the quoting variants the corpus uses: `<<'EOF'`, `<<"EOF"`,
      `<<EOF` and the indented `<<~EOF`
- [ ] Run to green
- [ ] Measure the T1 delta in nodes AND bytes; regenerate
      `internal/parse/testdata/t1.ratchet` and the `shortfall` map in
      `internal/parse/t2_test.go` in the SAME commit
- [ ] Run `go test ./internal/lexer/ ./internal/parse/ -count=1`
- [ ] Commit

## Acceptance Criteria

### Positive Scenarios

- [ ] A heredoc in statement position parses with no Unknown node (`go test ./internal/parse/ -run '^TestHeredocInStatementPosition$' -count=1 -v | grep -q '^--- PASS: TestHeredocInStatementPosition'`)
- [ ] All four quoting forms parse and round-trip (`go test ./internal/parse/ -run '^TestHeredocQuotingForms$' -count=1 -v | grep -q '^--- PASS: TestHeredocQuotingForms'`)
- [ ] The T2 metric is green with the shortfall map regenerated in the same
      commit (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] The T1 ratchet is green with its baseline regenerated in the same commit (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)

Plan: M1. Survey: 01a0c10b bucket 4.
