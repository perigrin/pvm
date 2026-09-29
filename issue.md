---
title: "`local *@` is not a glob: `*@` lexes as a star and a punctuation variable"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:06:00.071366895Z
updated: 2026-09-29T09:18:12.323952066Z
sessions:
- start_sha: 6f5e809bf3b21ccb29e252f189b7d6c7edc7461e
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T09:18:12.323952066Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T09:18:12.323952066Z
---

`local *@;` is valid perl (it localises the glob named `@`), but the lexer does not read `*@` as a glob, so the statement canonicalised as `local *@;;` at Unknown=0 and now refuses. globWithSpecialName covers `*-`, `*+`, `*/`, `*!`, `*&`, digits and `*^X`, not `*@`.

Measured with perl 5.42, `perl -MO=Deparse -e 'sub f { local *@; 1 }'` keeps `local *@;`.

perl.git t/op/local.t:824.

## Acceptance criteria

- [ ] `*@` is a glob (`go test ./internal/parse/ -run '^TestGlobAtSign$' -count=1 -v | grep -q '^--- PASS: TestGlobAtSign'`)
