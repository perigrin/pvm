---
title: The apostrophe package separator is not gated on feature apostrophe_as_package_separator
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T07:27:50.679702926Z
updated: 2026-10-11T01:58:41.649935954Z
sessions:
- start_sha: 24bfac4a5293052e7c38dd54a56a58aef818e639
  end_sha: ""
  commits: 0
  started_at: 2026-10-11T00:57:41.886264706Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T00:57:41.886264706Z
---

The lexer reads `'` as a package separator in every file, both inside a name (`$main'a`, scanIdentRunes) and right after a sigil (`$'b`, scanVarName). perl gates both on feature `apostrophe_as_package_separator`, which the 5.41 bundle turns off, so `use v5.42;` or `no feature "apostrophe_as_package_separator"` makes `$'b` the postmatch and a bareword. Measured with perl 5.42:

    $ perl -e 'use v5.42; our $b = 5; "xy" =~ /x/; print $'"'"'b'
    Bareword found where operator expected (Missing operator before "b"?) at -e line 1, near "$'b"

That is a syntax error in almost every spelling, so the ungated reading is wrong only when the bareword is itself an operator. The one found so far is the repeat: `$'x3` is `$' x 3` under the bundle and `$main::x3` without it.

The lexer already tracks one pragma (`l.utf8Pragma`). Gating needs the same for the feature bundle and for explicit `no feature`.

### Fixed (2026-10-11)

Measured on 5.42: the separator holds with no `use` and under `use v5.40`
(`$main'a` is `$a`, `sub isn't` is `isn::t`); from `use v5.41`/`v5.42`/
`use 5.042`, or `no feature 'apostrophe_as_package_separator'`, it is off --
`$'x3` is `$' x 3` and `$'b` is "Bareword found where operator expected".
Fixed in d4b186ca, in the lexer, which tokenizes before the parser notes
features. The same commit fixed versionAtLeast reading `use 5.036` as minor
3, so `use 5.036` never turned signatures on. The criterion named a test
never written under that name; it now names the tests that landed.

## Acceptance criteria

- [ ] the apostrophe package separator follows its feature: on by default and under v5.40, off from a v5.41 bundle or `no feature`, and `$main'a` without a bundle is still `$main::a` (`go test ./internal/lexer/ -run '^TestApostropheSeparatorFollowsItsFeature$' -count=1 -v | grep -q '^--- PASS: TestApostropheSeparatorFollowsItsFeature'`)
- [ ] `use 5.036` turns signatures on, as perl reads it (`go test ./internal/lexer/ -run '^TestSignatureIsNotAPrototype$' -count=1 -v | grep -q '^--- PASS: TestSignatureIsNotAPrototype'`)
