---
title: "`$a CORE::eq $b` refuses: CORE:: word operators are not in the tables"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T16:43:35.495579112Z
updated: 2026-09-30T16:49:42.422838424Z
sessions:
- start_sha: 0e02c67405b811bea16863ec69e6b2d906d7b762
  end_sha: 4aff87bf40eedf08eb1e0b1db5cb540885f609d4
  commits: 1
  started_at: 2026-09-30T16:43:37.554674986Z
  ended_at: 2026-09-30T16:49:42.422838424Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T16:43:37.554674986Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-30T16:49:42.422838424Z
observed_paths:
- internal/parse/coreinfix_test.go
- internal/parse/precedence.go
- internal/parse/testdata/t1.ratchet
---

`$dot CORE::eq '.'` refuses. perl accepts a CORE:: prefix on each word operator and reads it as the bare word. Measured with perl 5.42 and -MO=Deparse for eq ne lt gt le ge cmp x and or xor isa: `$r = $a CORE::eq $b` is `$r = $a eq $b`. The operator tables are keyed by the spelled text, so `CORE::eq` is not an operator.

PerlOnJava unit/core_qualified_infix.t:6-9.

## Acceptance criteria

- [ ] CORE:: word operators bind as the bare ones and canon keeps the spelling (`go test ./internal/parse/ -run '^TestCoreQualifiedInfix$' -count=1 -v | grep -q '^--- PASS: TestCoreQualifiedInfix'`)
