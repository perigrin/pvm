---
title: "A glob with a digit, caret or punctuation name refuses: *1, *^R and *- lose their name"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T05:21:17.339485318Z
updated: 2026-09-29T05:27:17.655446895Z
sessions:
- start_sha: 1160b99169804f5ed3dae905df9451608e6f9efe
  end_sha: ef36d1ef0486ea3dddfc702d261fdc29c7b92cb0
  commits: 1
  started_at: 2026-09-29T05:21:17.750541895Z
  ended_at: 2026-09-29T05:27:17.655446895Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T05:21:17.750541895Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T05:27:17.655446895Z
observed_paths:
- internal/parse/globname_test.go
- internal/parse/term.go
- internal/parse/testdata/perlgitt.ratchet
---

A glob whose name is not an identifier refuses: the glob term read only a Word,
a Variable or `{` after the star, so the name fell out as its own token:

    local *a = *1;        canon  local *a = *;1;
    *^R = *foo;           canon  * ^ R() = *foo;
    *X = *-;              Unknown=2

perl accepts any variable name there. Measured on 5.42.0, all three deparse as
written. The lexer emits the star and the name as separate tokens, so the glob
is assembled in parseTerm: a digit name, a caret name (`^` and a letter), or a
single punctuation byte, each touching the star.

Reach: op/method.t (`local *1 = sub { 123 }`), re/pat.t (`*^R`), re/subst.t
(`*1`), re/reg_namedcapture.t (`*X = *-`), op/tie_fetch_count.t.

## Acceptance criteria

- [ ] `*1`, `*^R` and `*-` are globs and canon to a fixpoint (`go test ./internal/parse/ -run '^TestGlobWithNonIdentifierName$' -count=1 -v | grep -q '^--- PASS: TestGlobWithNonIdentifierName'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
