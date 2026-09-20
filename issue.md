---
title: A leading-decimal literal is a number, not an Unknown
state: pending
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0c13f-c7d7-7d9c-9a85-60a59aa26bc0
created: 2026-09-20T23:56:05.493446087Z
updated: 2026-09-20T23:56:45.608176819Z
---

A numeric literal with no digit before the point is a number in perl and an
Unknown here.

## Context

- paths: `internal/lexer/` (number scanning), `internal/parse/term.go`,
  `internal/parse/t2_test.go`, `internal/parse/testdata/t1.ratchet`
- commands: `go test ./internal/lexer/ ./internal/parse/ -count=1`
- entrypoints: wherever the lexer scans a numeric literal; the `.` is being
  read as an operator rather than as the start of a number.

## Measured, 2026-09-20 at b6e7c351

    my $x = .5;                    1 Unknown
    my $x = int($f * 100 + .5);    2 Unknown
    my $x = 0.5;                   0

`base/num.t` carries 48 Unknowns and `base/lex.t` 37; leading decimals are
named in the survey as a contributor to both. The exact share is what the
first step measures.

A leading `.` is unambiguous in TERM position: nothing else can start there.
In operator position it is concatenation, which must keep working. The rule
is the lexer's expect state, the same mechanism §3.2 uses for a leading `%`.

## Steps

- [ ] Measure how many of `base/num.t`'s 48 and `base/lex.t`'s 37 Unknowns
      are leading decimals, and record the number here
- [ ] Write a failing test: `my $x = .5;` has no Unknown node, and `.5`
      lexes as one numeric literal whose text is `.5`
- [ ] Run it and confirm it fails for that reason (`go test ./internal/lexer/ -run '^TestLeadingDecimalLiteral$' -count=1 -v`)
- [ ] Implement: in term-expecting position, a `.` followed by a digit starts
      a number
- [ ] Write the companion test that concatenation is unharmed: `$a . 5`,
      `$a .5` and `"x" . .5` each keep their operator reading
- [ ] Run both to green
- [ ] Measure the T1 delta in nodes AND bytes; regenerate
      `internal/parse/testdata/t1.ratchet` and the `shortfall` map in
      `internal/parse/t2_test.go` in the SAME commit
- [ ] Run `go test ./internal/lexer/ ./internal/parse/ -count=1`
- [ ] Commit

## Acceptance Criteria

### Positive Scenarios

- [ ] A leading-decimal literal is one number and the statement is clean (`go test ./internal/lexer/ -run '^TestLeadingDecimalLiteral$' -count=1 -v | grep -q '^--- PASS: TestLeadingDecimalLiteral'`)
- [ ] Concatenation still reads as concatenation (`go test ./internal/lexer/ -run '^TestDotIsStillConcatenation$' -count=1 -v | grep -q '^--- PASS: TestDotIsStillConcatenation'`)
- [ ] The T2 metric is green with the shortfall map regenerated in the same
      commit (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] The T1 ratchet is green with its baseline regenerated in the same commit (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)

Spec: §3.2. Plan: M1. Survey: 01a0c10b bucket 3.
