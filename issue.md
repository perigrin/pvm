---
title: A Unicode letter number, `ⅲ`, is not an identifier character under use utf8
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T14:46:43.376463174Z
updated: 2026-09-29T14:46:43.376463174Z
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
