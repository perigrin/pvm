---
title: Statement skeleton, and the recovery that stops a cascade
state: done
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-0f56-71e3-871e-50e1a1ed9b8a
blocks:
- 01a0a75b-5e51-79ef-a59a-3a997f738cca
- 01a0a75b-95b2-730a-9303-b6aba9adb26c
created: 2026-09-15T21:52:57.320252116Z
updated: 2026-09-16T03:13:01.905950482Z
sessions:
- start_sha: 68994750d92c543b5e60a6043acc8b215945a4ac
  end_sha: 5b936a5d9c9c023339bf9efbde2faf62e31f3c78
  commits: 1
  started_at: 2026-09-16T02:27:17.488294006Z
  ended_at: 2026-09-16T03:13:01.905950482Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-16T02:27:17.488294006Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-16T03:13:01.905950482Z
observed_paths:
- internal/lexer/brackets.go
- internal/lexer/lexer.go
- internal/parse/parse.go
- internal/parse/stmt_test.go
---

The statement skeleton: `source_file` as a sequence of statements, the
expression statement, `;`, the bare block, and the recovery rule that keeps
one unfamiliar construct from cascading.

Split out of a larger statements issue so that the keyword table has
something to sit on. A list operator "swallows to the statement end" needs a
statement end to exist; without this issue first, that issue and this one
depend on each other.

## Recovery is the point

Any statement form this milestone does not implement emits an Unknown
spanning the whole statement, and parsing resumes at the next `;` or `}`.

Resynchronising on a statement boundary is what stops a cascade, and the
cascade is not hypothetical: `op/filetest.t:96` derailed a reference at line
131 under the tree-sitter grammar — one bad line, 35 lines of wreckage.

Per the skeleton issue, a statement containing an Unknown also emits
Unresolved sites of every marker kind, so it scores wider rather than WRONG.

## Acceptance

- [ ] A file parses to a sequence of statements with no gaps between spans (`go test ./internal/parse/ -run '^TestSourceFileIsStatements$' -count=1 -v | grep -q '^--- PASS: TestSourceFileIsStatements'`)
- [ ] An expression statement and a bare block parse (`go test ./internal/parse/ -run '^TestExpressionStatementAndBlock$' -count=1 -v | grep -q '^--- PASS: TestExpressionStatementAndBlock'`)
- [ ] An unknown statement becomes Unknown and parsing resumes at the next boundary (`go test ./internal/parse/ -run '^TestStatementRecovery$' -count=1 -v | grep -q '^--- PASS: TestStatementRecovery'`)
- [ ] One unparseable statement does not affect the next one's spans (`go test ./internal/parse/ -run '^TestRecoveryDoesNotCascade$' -count=1 -v | grep -q '^--- PASS: TestRecoveryDoesNotCascade'`)

Spec: ch5 §5.2, §5.13.3.
