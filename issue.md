---
title: "`$r->&*` refuses: the code and glob postfix derefs lex as a sigil and a star"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T16:53:39.276060993Z
updated: 2026-09-30T17:21:30.624574895Z
sessions:
- start_sha: 4aff87bf40eedf08eb1e0b1db5cb540885f609d4
  end_sha: 7509ed4ad526521131693e0e6882e9b6ee3ea667
  commits: 2
  started_at: 2026-09-30T16:53:40.813031943Z
  ended_at: 2026-09-30T17:21:30.624574895Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T16:53:40.813031943Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-30T17:21:30.624574895Z
observed_paths:
- internal/lexer/expect_test.go
- internal/lexer/lexer.go
- internal/lexer/postderefstar_test.go
- internal/lexer/scan.go
- internal/memory/lazy.go
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1fidelity.ratchet
---

`$sub_ref->&*` and `$r->**` refuse. After `->`, `&*` lexes as the & sigil plus a star, and `**` as exponentiation. perl reads both as whole postfix dereferences, as it reads `->@*`. Measured with perl 5.42 and -MO=Deparse: `$x = $r->&*` is `$x = &$r`, and `$g = $r->**` is `$g = *$r`.

PerlOnJava unit/postderef.t:31 and unit/postderef_coderef_reference.t:9.

## Acceptance criteria

- [ ] `->&*` and `->**` lex as one Variable each (`go test ./internal/lexer/ -run '^TestPostDerefCodeAndGlob$' -count=1 -v | grep -q '^--- PASS: TestPostDerefCodeAndGlob'`)
