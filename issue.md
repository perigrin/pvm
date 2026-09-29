---
title: The apostrophe package separator is not gated on feature apostrophe_as_package_separator
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T07:27:50.679702926Z
updated: 2026-09-29T07:27:50.679702926Z
---

The lexer reads `'` as a package separator in every file, both inside a name (`$main'a`, scanIdentRunes) and right after a sigil (`$'b`, scanVarName). perl gates both on feature `apostrophe_as_package_separator`, which the 5.41 bundle turns off, so `use v5.42;` or `no feature "apostrophe_as_package_separator"` makes `$'b` the postmatch and a bareword. Measured with perl 5.42:

    $ perl -e 'use v5.42; our $b = 5; "xy" =~ /x/; print $'"'"'b'
    Bareword found where operator expected (Missing operator before "b"?) at -e line 1, near "$'b"

That is a syntax error in almost every spelling, so the ungated reading is wrong only when the bareword is itself an operator. The one found so far is the repeat: `$'x3` is `$' x 3` under the bundle and `$main::x3` without it.

The lexer already tracks one pragma (`l.utf8Pragma`). Gating needs the same for the feature bundle and for explicit `no feature`.

## Acceptance criteria

- [ ] Under `use v5.42` an apostrophe is never a package separator (`go test ./internal/lexer/ -run '^TestApostropheSeparatorFeatureGate$' -count=1 -v | grep -q '^--- PASS: TestApostropheSeparatorFeatureGate'`)
