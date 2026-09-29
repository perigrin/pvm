---
title: "$#+ and $#- lex as $# and an operator: the last index of @+ refuses"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T04:42:21.816204177Z
updated: 2026-09-29T04:42:22.690790398Z
sessions:
- start_sha: 2ff17048eb0761ee20df9012bdcda59e36ef6136
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T04:42:22.690790398Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T04:42:22.690790398Z
---

`$#+` and `$#-` -- the last index of @+ and @- -- lex as the variable `$#`
followed by an operator, so the statement refuses:

    my $n = $#+;            canon  my $n = $# + ;;     Unknown=1

toke.c reads `$#` as last-index when the next byte is an identifier start or
one of six punctuation bytes:

    if (   s[1] == '#'
        && (   isIDFIRST_lazy_if_safe(s+2, PL_bufend, UTF)
            || memCHRs("{$:+-@", s[2])))

The lexer took only a word byte, `{` or `$`. Measured on 5.42.0, after
`"ab" =~ /(a)(b)/`, `print $#+, " ", $#-` prints `2 2`, and Deparse keeps
`$#+`, `$#-` and `$#::x` as written.

Reach: re/pat.t and re/pat_rt_report.t.

## Acceptance criteria

- [ ] `$#+`, `$#-`, `$#@` and `$#::x` are one Variable each, and `$#` alone is still the variable named `#` (`go test ./internal/lexer/ -run '^TestLastIndexOfPunctuationArray$' -count=1 -v | grep -q '^--- PASS: TestLastIndexOfPunctuationArray'`)
- [ ] REGRESSION GUARD: the lexer's golden token stream is unchanged (`go test ./internal/lexer/ -count=1`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
