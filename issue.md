---
title: A refusal at end of file is on the operator's line, but perl reports the last line it read
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T03:58:10.458182083Z
updated: 2026-10-09T03:58:10.458182083Z
---

## Context

- paths: `internal/parse/expr.go` (`operand()`'s `missing_operand` refusal), `internal/parse/refusal.go`, `internal/parse/expr_test.go` (`TestTrailingInfixAtEOFRefused`)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found finishing 01a11eb2-21ac (2026-10-09). A trailing infix operator at end of file is refused (`missing_operand`), but on the line of the operator. perl reports the last line it read: measured on 5.42, `print 1+` followed by a blank line, or by a comment line, is "syntax error at ... line 2, at EOF", where internal/parse's refusal is on line 1. comp/final_line_num.t has no trailing newline, so T2 is unaffected. Deciding whether a refusal carries perl's reported line separately from its span is a design question for perigrin.

## Steps

- [ ] Measure perl on trailing blank lines, comments and POD after the operator; bring perigrin the options (a separate reported line, or the refusal's span extended to end of file)
- [ ] Failing test first; implement the ruling; `make test`; commit

## Acceptance criteria

- [ ] a refusal at end of file reports the line perl reports (`go test ./internal/parse/ -run '^TestEOFRefusalReportsPerlsLine$' -count=1 -v | grep -q '^--- PASS: TestEOFRefusalReportsPerlsLine'`)
