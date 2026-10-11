---
title: require q(./test.pl) is not a literal path, so its helper's subs are never declared
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T23:52:22.724446205Z
updated: 2026-10-11T15:17:45.971849896Z
sessions:
- start_sha: 0aaccdcbebea1a3bfae177cf641ba27c12f1b049
  end_sha: 0c70adb2c77f4e06e9adf757e2199a2580c48a09
  commits: 1
  started_at: 2026-09-29T00:16:35.907778568Z
  ended_at: 2026-09-29T00:24:19.068156618Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T00:16:35.907778568Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T00:24:19.068156618Z
observed_paths:
- internal/parse/exports.go
- internal/parse/qrequire_test.go
- internal/parse/requirequote_test.go
- internal/parse/testdata/perlgitt.ratchet
---

`require q(./test.pl)` is not resolved, so none of test.pl's subs are declared
and every PARENLESS call to one refuses. Calls with parens still parse, which
is why the files look mostly clean -- mro/package_aliases.t reads `ok (...)`
and `object_ok (...)` fine and then refuses at its first `fresh_perl_is q~...~`.

    require q(./test.pl);
    fresh_perl_is "x", "b";          Unknown=1

`literalNameList` reads `'...'`, `"..."` and `qw(...)`, but not `q(...)` or
`qq(...)` with perl's other delimiters. The fix belongs there, where every
caller -- the require resolver and the export-list readers -- gets it. A `qq`
body that interpolates is not a literal name, for the same reason
resolveRequiredFile's `interpolates` check exists for `"..."`.

Reach: 61 perl.git t/ files spell their require with `q`/`qq`; 13 are dirty,
110 Unknowns between them (an upper bound -- not every Unknown there is this).

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance criteria

- [ ] `q(...)` and `qq(...)` with any delimiter are literal names, and an interpolating `qq` is not (`go test ./internal/parse/ -run '^TestLiteralNameQuoteForms$' -count=1 -v | grep -q '^--- PASS: TestLiteralNameQuoteForms'`)
- [ ] `require q(./test.pl)` contributes test.pl's sub declarations (`go test ./internal/parse/ -run '^TestRequireQuoteForm$' -count=1 -v | grep -q '^--- PASS: TestRequireQuoteForm'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
