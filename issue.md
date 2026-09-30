---
title: "`Y / X` after `use constant` opens a pattern: constants are not declared"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T17:39:31.523098476Z
updated: 2026-09-30T17:46:51.276139112Z
sessions:
- start_sha: 60f88793a1b0f49ac5e598d92f3f4ff72778c539
  end_sha: c9401c372fe12638de681ca135ac5a13004384a7
  commits: 1
  started_at: 2026-09-30T17:39:33.784704173Z
  ended_at: 2026-09-30T17:46:51.276139112Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T17:39:33.784704173Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-30T17:46:51.276139112Z
observed_paths:
- internal/lexer/constsub_test.go
- internal/lexer/expect.go
- internal/lexer/lexer.go
- internal/lexer/proto.go
- internal/parse/testdata/t1.ratchet
- internal/parse/use.go
- internal/parse/useconstant_test.go
---

`is(Y / X, 2, ...)` refuses when X and Y come from `use constant`. After an undeclared word the lexer expects a term, so `/ X, 2, 'constant division');\n...` lexes as a pattern. perl's lexer finds the constant sub in its symbol table and expects an operator. The parser doesn't record the names either.

Measured with perl 5.42: `prototype "main::X"` is "" for `use constant X => 2`, for the keys of `use constant { A => 1 }`, for a list constant, and for `sub Z () {...}`. `print Y / X` deparses as a constant-folded division.

PerlOnJava unit/constant.t:60.

## Acceptance criteria

- [ ] The lexer expects an operator after a declared constant (`go test ./internal/lexer/ -run '^TestDeclaredConstantDivides$' -count=1 -v | grep -q '^--- PASS: TestDeclaredConstantDivides'`)
- [ ] `use constant` declares its names with the empty prototype (`go test ./internal/parse/ -run '^TestUseConstantDeclares$' -count=1 -v | grep -q '^--- PASS: TestUseConstantDeclares'`)
