---
title: "`local *@` is not a glob: `*@` lexes as a star and a punctuation variable"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:06:00.071366895Z
updated: 2026-09-29T09:28:21.65716838Z
sessions:
- start_sha: 6f5e809bf3b21ccb29e252f189b7d6c7edc7461e
  end_sha: 4b37f0c8c8caaf2aa7cfe550cd3a4d63757bfe13
  commits: 1
  started_at: 2026-09-29T09:18:12.323952066Z
  ended_at: 2026-09-29T09:28:21.65716838Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T09:18:12.323952066Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T09:28:21.65716838Z
observed_paths:
- internal/lexer/lexer.go
- internal/lexer/scan.go
- internal/parse/globat_test.go
- internal/parse/term.go
- internal/parse/testdata/perlgitt.ratchet
---

`local *@;` is valid perl (it localises the glob named `@`), but the lexer does not read `*@` as a glob, so the statement canonicalised as `local *@;;` at Unknown=0 and now refuses. globWithSpecialName covers `*-`, `*+`, `*/`, `*!`, `*&`, digits and `*^X`, not `*@`.

Measured with perl 5.42, `perl -MO=Deparse -e 'sub f { local *@; 1 }'` keeps `local *@;`.

perl.git t/op/local.t:824.

## Acceptance criteria

- [ ] `*@` is a glob (`go test ./internal/parse/ -run '^TestGlobAtSign$' -count=1 -v | grep -q '^--- PASS: TestGlobAtSign'`)
