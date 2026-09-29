---
title: "`$'b` splits into the postmatch and a bareword: a leading apostrophe is a package separator too"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T07:27:50.555588785Z
updated: 2026-09-29T07:38:48.797801401Z
sessions:
- start_sha: db395c977bc74bac6e005d6f1048bf7d70b43aa9
  end_sha: d6662dfe252501db6de1354bab1263b50feaa372
  commits: 1
  started_at: 2026-09-29T07:28:29.138436107Z
  ended_at: 2026-09-29T07:38:48.797801401Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T07:28:29.138436107Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T07:38:48.797801401Z
observed_paths:
- internal/lexer/ident_test.go
- internal/lexer/scan.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
---

The lexer takes `'` as a package separator only between identifier characters, so `$'b` lexed as the postmatch variable `$'` and a bareword `b`. toke.c's parse_ident takes the apostrophe at the first position as at any other, when an identifier start follows:

    (   *s == '\''
     && FEATURE_APOS_AS_NAME_SEP_IS_ENABLED
     && isIDFIRST_lazy_if_safe(s+1, s_end, is_utf8))

Measured with perl 5.42:

    $ perl -e '$main::b = 5; "xy" =~ /x/; print $'"'"'b, "|", $'"'"''
    5|y

The first refusal in perl.git t/comp/package.t is line 17, `$main'a = $'b;`. The existing lexer test quoted that line without its `b`.

## Acceptance criteria

- [ ] `$'b` is one variable and `$'` before a non-identifier is still the postmatch (`go test ./internal/lexer/ -run '^TestLeadingApostropheSeparator$' -count=1 -v | grep -q '^--- PASS: TestLeadingApostropheSeparator'`)
- [ ] REGRESSION GUARD: interior separators still hold (`go test ./internal/lexer/ -run '^TestIdentifierPackageSeparators$' -count=1 -v | grep -q '^--- PASS: TestIdentifierPackageSeparators'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
