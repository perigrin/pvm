---
title: A Unicode letter number, `ⅲ`, is not an identifier character under use utf8
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T14:46:43.376463174Z
updated: 2026-09-29T14:55:54.945570567Z
sessions:
- start_sha: e60f44cabd8d3937bca2a7fcf7e39f7eccbe7b58
  end_sha: 180c36303e06048aeb3fb31e9a8d771c7dc70293
  commits: 1
  started_at: 2026-09-29T14:46:45.155806016Z
  ended_at: 2026-09-29T14:55:54.945570567Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T14:46:45.155806016Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T14:55:54.945570567Z
observed_paths:
- internal/lexer/ident.go
- internal/lexer/letternumber_test.go
- internal/lexer/testdata/corpus.ratchet
- internal/parse/testdata/perlgitt.ratchet
---

Under `use utf8`, a letter number (Unicode category Nl, such as `ⅲ` U+2172) does not lex as an identifier character. identStart and identContinue ask unicode.IsLetter, which is the L categories alone; perl's identifier is \p{XID_Start} and \p{Word}, and \p{Word} takes Nl through Alphabetic.

Measured with perl 5.42:

    $ perl -Mutf8 -e 'package bugⅲⅱⅴⅵⅱ { sub f { __PACKAGE__ } } binmode STDOUT, ":utf8";
          print bugⅲⅱⅴⅵⅱ::f(), "\n"; my $ⅷ = 8; print "$ⅷ\n";'
    bugⅲⅱⅴⅵⅱ
    8

perl.git t/uni/package.t:94; the lexer ratchet counts 15 error tokens there.

## Acceptance criteria

- [ ] A letter number is an identifier character under `use utf8` (`go test ./internal/lexer/ -run '^TestLetterNumberIdentifier$' -count=1 -v | grep -q '^--- PASS: TestLetterNumberIdentifier'`)
