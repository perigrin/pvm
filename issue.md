---
title: "`$r->&*` refuses: the code and glob postfix derefs lex as a sigil and a star"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T16:53:39.276060993Z
updated: 2026-09-30T16:53:39.276060993Z
---

`$sub_ref->&*` and `$r->**` refuse. After `->`, `&*` lexes as the & sigil plus a star, and `**` as exponentiation. perl reads both as whole postfix dereferences, as it reads `->@*`. Measured with perl 5.42 and -MO=Deparse: `$x = $r->&*` is `$x = &$r`, and `$g = $r->**` is `$g = *$r`.

PerlOnJava unit/postderef.t:31 and unit/postderef_coderef_reference.t:9.

## Acceptance criteria

- [ ] `->&*` and `->**` lex as one Variable each (`go test ./internal/lexer/ -run '^TestPostDerefCodeAndGlob$' -count=1 -v | grep -q '^--- PASS: TestPostDerefCodeAndGlob'`)
