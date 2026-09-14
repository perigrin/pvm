---
title: "Lexer skeleton: token, cursor, and the forward-progress guarantee"
state: pending
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:41:55.194811622Z
updated: 2026-09-14T15:42:03.932341573Z
---

The lexer package does not exist. This issue creates it with the token type,
the cursor, and the forward-progress guarantee everything else rests on —
and nothing else. No Perl-specific lexing yet.

Spec §7.6.2 invariant 4 is the reason this is its own issue: a Perl lexer
WILL infinite-loop during development (an unterminated heredoc or quote-like
operator that fails to advance the cursor is the classic case), and `go test
-fuzz` does not detect infinite loops — it hangs. The guard has to be
structural and it has to exist before there is anything to guard.

## Shape

    package lexer   // internal/lexer

    type Token struct {
        Kind       Kind
        Start, End int   // byte offsets into the source
    }

    func Tokenize(src []byte) []Token

Trivia (whitespace, comments, POD) are TOKENS, not skipped: invariant 3 is
lossless round-trip, so every byte of input belongs to exactly one token.
That is the decision that makes the round-trip metric reachable at all, and
retrofitting it later is a rewrite.

Positions are byte offsets, not line/column. Line/column is a presentation
concern the LSP layer derives; carrying it in the token doubles the state
that has to stay consistent under edit.

## Acceptance

- [ ] Package builds and Tokenize returns one trivia token for whitespace-only input (`go test ./internal/lexer/ -run TestTokenizeTrivia -count=1`)
- [ ] Every lexer step consumes at least one byte, asserted structurally rather than by timeout (`go test ./internal/lexer/ -run TestForwardProgress -count=1`)
- [ ] Positions are monotonic, non-negative, and within bounds on every input (`go test ./internal/lexer/ -run TestPositionMonotonicity -count=1`)
- [ ] Concatenating every token's source text reproduces the input exactly (`go test ./internal/lexer/ -run TestLossless -count=1`)

Spec: §7.6.2 invariants 1-4. Plan: M0.
