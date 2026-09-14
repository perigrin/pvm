---
title: "Quote-like operators and heredocs: delimiters, nesting, and the escape ordering"
state: pending
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:42:13.261933199Z
updated: 2026-09-14T15:42:37.215129032Z
---

A quote-like operator's body cannot be lexed without knowing the operator
and its delimiter. This is the largest single block of M0 work and it is
what `t/base/lex.t` exists to break.

Covers `q qq qw m s tr y qr` and the bare forms `//`, `??`, `''`, `""`,
backticks, plus heredocs.

## The three rules, each measured this session

**Any non-whitespace character delimits.** `toke.c` `Perl_scan_str` (~12272):
"after skipping whitespace, the next character is the delimiter". No
allow-list. `q~...~`, `q#...#`, `s\a\b\` are all legal.

**Bracketing pairs nest; everything else does not.** `( [ { <` have a
distinct closer and count depth; `if (PL_multi_open == PL_multi_close)`
turns nesting off for the rest. Three-part operators (`s tr y`) take a
SECOND delimiter pair when the first is bracketing: `s{a}{b}` but `s/a/b/`.

**The closer is checked BEFORE the escape.** When the delimiter is itself a
backslash, escapes are impossible and every `\` is a delimiter — perl guards
this with `close_delim_code != '\\'` at `toke.c:12445`. Get the ordering
backwards and `s\pat\repl\` scans to EOF, derailing the rest of the file.

That ordering is a required test, not an obvious detail: PerlOnJava gets it
right (`StringParser.java:205-225`, closer then escape); gotreesitter and
perl-lsp both got it wrong the same way. Three implementations, two
identical bugs.

## Heredocs

The terminator is read at the `<<` but the body starts after the current
LINE ends, so the lexer holds a pending-heredoc queue. `<<'EOF'` does not
interpolate, `<<EOF` does, `<<\EOF` is the backslash-quoted form, `<<~EOT`
strips indentation. `t/base/lex.t` uses four of these in its opening lines,
including one mid-expression: `eval <<\EOE, print $@;`

## Acceptance

- [ ] Any non-whitespace delimiter is accepted for every quote-like operator (`go test ./internal/lexer/ -run TestQuoteArbitraryDelimiter -count=1`)
- [ ] Bracketing delimiters nest and non-bracketing ones do not (`go test ./internal/lexer/ -run TestQuoteNesting -count=1`)
- [ ] Three-part operators take a second pair only when the first is bracketing (`go test ./internal/lexer/ -run TestQuoteThreePart -count=1`)
- [ ] A backslash delimiter admits no escapes, closer checked before escape (`go test ./internal/lexer/ -run TestQuoteBackslashDelimiter -count=1`)
- [ ] All four heredoc forms lex, including one mid-expression (`go test ./internal/lexer/ -run TestHeredocForms -count=1`)
- [ ] Every quote-like body round-trips losslessly (`go test ./internal/lexer/ -run TestQuoteLossless -count=1`)

Findings: §0.13 ranks 3 and 5. Spec: ch2, ch3 §3.x lexer feedback.
