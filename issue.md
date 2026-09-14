---
title: "t/base/lex.t round-trips: the M0 gate"
state: pending
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:42:44.613699558Z
updated: 2026-09-14T15:42:56.445953975Z
---

M0's exit criterion, and the spec is emphatic about it (§7.3.2):

> Treat `lex.t` as a milestone in its own right, not as one file among 620.
> If it parses, the lexer is real.

16K, 129 tests, written specifically to abuse a lexer. Its opening lines are
six separate lexer-feedback traps:

    $x = $#[0];              # $# of an anonymous array? no: $#, then [0]
    $x = '\\'; # ';          # a quoted backslash, then a comment holding a quote
    eval '$foo{1} / 1;';     # division or pattern, after a hash subscript
    print <<'EOF';           # non-interpolating heredoc
    print <<EOF;             # interpolating heredoc
    eval <<\EOE, print $@;   # backslash-quoted terminator, mid-expression

Our current tree-sitter parser fails this file — it sits in the no-answer
bucket of the §0.11 baseline — so this is a red test on day one and stays
red until the preceding issues land. That is the point.

## Why round-trip and not parse

M0 asks only that the bytes survive tokenization. `lex.t` need not PARSE
here; it must tokenize losslessly, with every trap above delimited
correctly. Parsing it is M1's problem. Keeping the bar at round-trip is what
makes M0 finishable, and it is still a real gate: you cannot round-trip a
heredoc you did not recognise.

## Acceptance

- [ ] t/base/lex.t tokenizes without error (`go test ./internal/lexer/ -run TestLexDotTTokenizes -count=1`)
- [ ] t/base/lex.t round-trips byte-for-byte (`go test ./internal/lexer/ -run TestLexDotTRoundTrips -count=1`)
- [ ] Each of the six opening traps is pinned as its own case (`go test ./internal/lexer/ -run TestLexDotTOpeningTraps -count=1`)

Spec: §7.3.2. Plan: M0 exit criterion.
