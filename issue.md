---
title: "A niladic builtin's name after `sub` loses its body: `sub time { 1 }` parses as a forward declaration and a bare block"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T06:45:13.42664501Z
updated: 2026-09-29T06:57:05.938031108Z
sessions:
- start_sha: d1037956815eca1e624d44e35f143745d546d2f7
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T06:57:05.938031108Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T06:57:05.938031108Z
---

In lexer/expect.go the Word transition tests isNiladic before afterDeclName, so the word after `sub` returns XOperator and the `{` is never classified as the body. The tree is wrong at Unknown=0:

    sub time { 1 }    canon: sub time;{1};

Measured with perl 5.42, `perl -MO=Deparse -e 'sub time { 1 } sub wait { 2 }'` keeps both bodies.

The UNIDOR transition (issue for `shift // 7`) had the same ordering bug and was placed after the name and brace cases; the niladic test should move the same way. Corpus users include PerlOnJava unit/mojolicious_promise_lifecycle_regressions.t.

## Acceptance criteria

- [ ] A niladic builtin used as a sub name keeps its body (`go test ./internal/lexer/ -run '^TestNiladicSubNameOpensABlock$' -count=1 -v | grep -q '^--- PASS: TestNiladicSubNameOpensABlock'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
