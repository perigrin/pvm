---
title: "Quote-like operators and heredocs: delimiters, nesting, and the escape ordering"
state: pending
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:42:13.261933199Z
updated: 2026-09-14T22:45:07.888956032Z
---

The keyword-introduced quote-like operators and the plain string forms:
`q qq qw m s tr y qr`, `'...'`, `"..."`, backticks. Delimiter scanning only.

Split from heredocs and the bare forms deliberately. Everything here is
introduced by a keyword or a quote character, so it needs no term-versus-
operator knowledge: seeing `q` or `"` is enough. Heredocs and bare `/` do
need that bit, and they are later in the chain.

## The three rules, each measured

**Any non-whitespace character delimits.** `toke.c` `Perl_scan_str` (12272):
"after skipping whitespace, the next character is the delimiter". No
allow-list. `q~...~`, `q#...#`, `s\a\b\` are all legal. Even a WORD
character works after whitespace: `perl -e 'print q xfoox'` prints `foo`.
Spec §2.10.3 says word characters may not delimit; the spec is wrong and
should be corrected, with the measurement recorded.

**Bracketing pairs nest; everything else does not.** `( [ { <` have a
distinct closer and count depth. `toke.c:12412`:
`if (PL_multi_open == PL_multi_close)` turns nesting off for the rest.
Three-part operators (`s tr y`) take a SECOND delimiter pair when the first
is bracketing — `s{a}{b}` — and reuse the single delimiter otherwise:
`s/a/b/`.

**The closer is checked BEFORE the escape.** When the delimiter is itself a
backslash, escapes are impossible and every `\` is a delimiter; perl guards
it with `close_delim_code != '\\'` at `toke.c:12445`. Get the ordering
backwards and `s\pat\repl\` scans to EOF, derailing every line after it.

That ordering is a required test rather than an obvious detail. PerlOnJava
gets it right (`frontend/parser/StringParser.java:205-225`, closer then
escape); gotreesitter and perl-lsp both got it wrong the same way. Three
implementations, two identical bugs, and the one we depend on was one of the
wrong ones.

## Two asymmetries the corpus exercises

Spec §2.10.3: `q #a#` is a COMMENT, not a `#`-delimited string, because
whitespace before `#` makes it a comment — but `q#a#` is the string. Spec
§2.10.5: a comment may sit between the two parts of a bracketing three-part
operator, `s{a} # c\n {b}`.

## Acceptance

- [ ] Any non-whitespace delimiter is accepted, including a word char after whitespace (`go test ./internal/lexer/ -run '^TestQuoteArbitraryDelimiter$' -count=1 -v | grep -q '^--- PASS: TestQuoteArbitraryDelimiter'`)
- [ ] Bracketing delimiters nest and non-bracketing ones do not (`go test ./internal/lexer/ -run '^TestQuoteNesting$' -count=1 -v | grep -q '^--- PASS: TestQuoteNesting'`)
- [ ] Three-part operators take a second pair only when the first is bracketing (`go test ./internal/lexer/ -run '^TestQuoteThreePart$' -count=1 -v | grep -q '^--- PASS: TestQuoteThreePart'`)
- [ ] A backslash delimiter admits no escapes; the closer is checked first (`go test ./internal/lexer/ -run '^TestQuoteBackslashDelimiter$' -count=1 -v | grep -q '^--- PASS: TestQuoteBackslashDelimiter'`)
- [ ] `q #a#` is a comment, `q#a#` is a string, and a comment may sit between `s{a}` and `{b}` (`go test ./internal/lexer/ -run '^TestQuoteCommentAsymmetry$' -count=1 -v | grep -q '^--- PASS: TestQuoteCommentAsymmetry'`)

Spec: §2.10.3, §2.10.5, §2.15 items 9-10. toke.c: 12272, 12412, 12445.
