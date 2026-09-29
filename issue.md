---
title: "`local *@` is not a glob: `*@` lexes as a star and a punctuation variable"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:06:00.071366895Z
updated: 2026-09-29T08:06:00.071366895Z
---

`local *@;` is valid perl (it localises the glob named `@`), but the lexer does not read `*@` as a glob, so the statement canonicalised as `local *@;;` at Unknown=0 and now refuses. globWithSpecialName covers `*-`, `*+`, `*/`, `*!`, `*&`, digits and `*^X`, not `*@`.

Measured with perl 5.42, `perl -MO=Deparse -e 'sub f { local *@; 1 }'` keeps `local *@;`.

perl.git t/op/local.t:824.

## Acceptance criteria

- [ ] `*@` is a glob (`go test ./internal/parse/ -run '^TestGlobAtSign$' -count=1 -v | grep -q '^--- PASS: TestGlobAtSign'`)
