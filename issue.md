---
title: require './t/test.pl' after chdir '..' is not resolved, so porting/ files lose test.pl's subs
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T22:31:49.144380373Z
updated: 2026-09-28T22:31:49.144380373Z
---

Seventeen porting/ files load their helper as `require './t/test.pl'`, after
`chdir '..' if -f 'test.pl'` moves them from t/ to the perl root. The loader
resolves a literal require against t/, so './t/test.pl' is not found, none of
test.pl's subs are declared, and every parenless call to one refuses:

    skip_all "not devel"   unless -d "./.git";      porting/bench.t:25

4 of the 17 are dirty, 67 Unknowns between them (porting/bench.t 61).

Decide before fixing: the loader does not model `chdir`. Options are to model
the one idiom perl's own suite uses (`chdir '..' if -f 'test.pl'` then a path
relative to the root), or to resolve a literal require against the corpus
root's parent when the t/-relative path does not exist. The second is a
guess about the working directory; the first reads it from the file.

## Acceptance criteria

- [ ] A literal `require './t/test.pl'` after `chdir '..'` contributes test.pl's sub declarations (`go test ./internal/parse/ -run '^TestRequireFromPerlRoot$' -count=1 -v | grep -q '^--- PASS: TestRequireFromPerlRoot'`)
- [ ] REGRESSION GUARD: the t/-relative `require './test.pl'` still resolves (`go test ./internal/parse/ -run '^TestRequiredFileDeclarations$' -count=1 -v | grep -q '^--- PASS: TestRequiredFileDeclarations'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
