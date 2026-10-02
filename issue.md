---
title: Derive parse shapes from CORE.pmt and delete keyword.go
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-be8b-7823-9c6c-5be1406b57b8
- 01a0fa46-d9f7-7aab-b4c6-19b4050b7907
created: 2026-10-02T01:42:24.650555371Z
updated: 2026-10-02T01:42:27.263235045Z
---

`internal/parse/keyword.go` holds `namedUnary`, `listOperator` and `niladicParse`, measured by deparsing every keyword. For builtins with a prototype the shape follows from it (`ShapeOf`); for the rest `:unary` says it. So the tables derive from CORE.pmt. A test first shows the derived shapes equal keyword.go's tables -- any disagreement is a finding to measure against perl -- and then keyword.go goes, its measurement comments moving into CORE.pmt.

## Acceptance criteria

- [ ] Shapes derived from CORE.pmt equal keyword.go (`go test ./internal/parse/ -run '^TestCoreShapesMatchKeywordTables$' -count=1 -v | grep -q '^--- PASS: TestCoreShapesMatchKeywordTables'`)
- [ ] keyword.go is gone (`sh -c '! test -e internal/parse/keyword.go'`)
