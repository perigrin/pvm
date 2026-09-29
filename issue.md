---
title: "`print'x'` lexes as one word: an apostrophe after a keyword opens a string"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T12:26:50.026430523Z
updated: 2026-09-29T12:40:28.104028006Z
sessions:
- start_sha: 4058d489710e76ad523e76f9f1f555f013cd58b9
  end_sha: afacf7d206a75c7deb219d34136f542ea5929db1
  commits: 1
  started_at: 2026-09-29T12:26:50.446895722Z
  ended_at: 2026-09-29T12:40:28.104028006Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T12:26:50.446895722Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T12:40:28.104028006Z
observed_paths:
- internal/lexer/keywordapos_test.go
- internal/lexer/perlkeywords.go
- internal/lexer/scan.go
- internal/parse/testdata/perlgitt.ratchet
---

The lexer takes an apostrophe inside any bareword as the old package separator, so `print'x'` lexes as one word `print'x` and `eval'scalar %a[4,5,6]'` as a qualified name. toke.c's yyl_keylookup scans the first word WITHOUT package parts and looks it up as a keyword; only a non-keyword reaches yyl_just_a_word, where `'` joins a package name.

Measured with perl 5.42:

    $ perl -MO=Deparse -e "my \$r = eval'fred3(5)'; my \$s = print'x'; my \$t = foo'bar;"
    my $r = eval 'fred3(5)';
    my $s = print('x');
    my $t = 'foo::bar';

perl.git t/op/eval.t:315 and t/op/kvaslice.t:48.

## Acceptance criteria

- [ ] An apostrophe after a keyword opens a string; after another word it separates (`go test ./internal/lexer/ -run '^TestKeywordThenApostrophe$' -count=1 -v | grep -q '^--- PASS: TestKeywordThenApostrophe'`)
