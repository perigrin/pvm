---
title: "`watchdog 3;` refuses: a helper's sub declared inside a block is not recorded"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T17:16:30.013623379Z
updated: 2026-09-29T17:27:41.049212235Z
sessions:
- start_sha: f159095d1e39123371d1d367a81f4c2e6b16c97b
  end_sha: cce0f97fdbad6745670e0dcf0d0c57d05099ad5b
  commits: 1
  started_at: 2026-09-29T17:16:30.518447222Z
  ended_at: 2026-09-29T17:27:41.049212235Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T17:16:30.518447222Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T17:27:41.049212235Z
observed_paths:
- internal/parse/exports.go
- internal/parse/nestedsubfacts_test.go
- internal/parse/testdata/perlgitt.ratchet
---

A named sub declared inside a bare block of a required helper is not recorded: readModule reads only top-level statements, and t/test.pl declares `sub watchdog ($;$)` inside `{ # Closure ... }`. So `watchdog 3;` is an unknown word followed by a number in every test that calls it.

A named sub is package-global wherever it is declared; a lexical one is not. Measured with perl 5.42:

    $ perl -e '{ my $x; sub watchdog ($;$) { "wd@_" } } print watchdog 3;
          { my sub lex { 1 } } print defined(&lex) ? "visible" : "not visible";'
    wd3not visible

perl.git t/mro/package_aliases.t:299, t/mro/package_aliases_utf8.t, t/op/sub.t:417.

## Acceptance criteria

- [ ] A named sub in a helper's nested block resolves (`go test ./internal/parse/ -run '^TestRequiredHelperNestedSub$' -count=1 -v | grep -q '^--- PASS: TestRequiredHelperNestedSub'`)
