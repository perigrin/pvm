---
title: "`$r->&*` refuses: the code and glob postfix derefs lex as a sigil and a star"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T16:53:39.276060993Z
updated: 2026-09-30T16:53:40.813031943Z
sessions:
- start_sha: 4aff87bf40eedf08eb1e0b1db5cb540885f609d4
  end_sha: ""
  commits: 0
  started_at: 2026-09-30T16:53:40.813031943Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T16:53:40.813031943Z
---

`$sub_ref->&*` and `$r->**` refuse. After `->`, `&*` lexes as the & sigil plus a star, and `**` as exponentiation. perl reads both as whole postfix dereferences, as it reads `->@*`. Measured with perl 5.42 and -MO=Deparse: `$x = $r->&*` is `$x = &$r`, and `$g = $r->**` is `$g = *$r`.

PerlOnJava unit/postderef.t:31 and unit/postderef_coderef_reference.t:9.

## Acceptance criteria

- [ ] `->&*` and `->**` lex as one Variable each (`go test ./internal/lexer/ -run '^TestPostDerefCodeAndGlob$' -count=1 -v | grep -q '^--- PASS: TestPostDerefCodeAndGlob'`)
