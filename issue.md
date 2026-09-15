---
title: "Lexer skeleton: token, cursor, and the forward-progress guarantee"
state: done
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:41:55.194811622Z
updated: 2026-09-15T01:36:40.488939392Z
sessions:
- start_sha: 8532564348567b936f8e40e10f67e25b8b52e54c
  end_sha: afbc38473f8c843c24f137f4be73f6eb4a46d35a
  commits: 1
  started_at: 2026-09-15T01:26:48.145170376Z
  ended_at: 2026-09-15T01:36:40.488939392Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-15T01:26:48.145170376Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-15T01:36:40.488939392Z
observed_paths:
- internal/lexer/lexer.go
- internal/lexer/lexer_test.go
---

The lexer package does not exist. This issue creates it with the token type,
the cursor, the error kind, and the forward-progress guarantee everything
else rests on — and nothing else. No Perl-specific lexing yet.

Spec §7.6.2 invariant 4 is why forward progress is its own concern: a Perl
lexer WILL infinite-loop during development (an unterminated heredoc or
quote-like operator that fails to advance the cursor is the classic case),
and `go test -fuzz` does not detect infinite loops — it hangs. The guard has
to be structural and it has to exist before there is anything to guard.

## Shape

    package lexer   // internal/lexer

    type Token struct {
        Kind       Kind
        Start, End int   // byte offsets into the source
    }

    func Tokenize(src []byte) []Token

Trivia (whitespace, comments, POD) are TOKENS, not skipped: invariant 3 is
lossless round-trip, so every byte of input belongs to exactly one token.
That is what makes the round-trip metric reachable at all, and retrofitting
it later is a rewrite.

Positions are byte offsets, not line/column. Line/column is a presentation
concern the LSP layer derives; carrying it in the token doubles the state
that has to stay consistent under edit.

**An error is a token kind, not a return value.** `Tokenize` has no error
result, so the only way to report "I could not lex this" while keeping every
byte accounted for is a token. Spec §2.15 item 18 names two: `Error` for a
byte that cannot start any token, and `UnknownRest` for a construct whose
end could not be found. Later issues (the identifier class, the lex.t gate)
assert on these, so they are declared here or they are declared twice.

## On the forward-progress test

There is no Perl lexing yet, so no input can stall the cursor: a test that
feeds sources and checks termination proves nothing. The test that fails
when the guard is deleted installs a scan state that does not advance and
asserts the guard fires. That is an in-package test by necessity.

## Acceptance

- [ ] Tokenize returns one trivia token for whitespace-only input (`go test ./internal/lexer/ -run '^TestTokenizeTrivia$' -count=1 -v | grep -q '^--- PASS: TestTokenizeTrivia'`)
- [ ] Error and UnknownRest kinds exist and carry their source span (`go test ./internal/lexer/ -run '^TestErrorKindsSpanSource$' -count=1 -v | grep -q '^--- PASS: TestErrorKindsSpanSource'`)
- [ ] A non-advancing scan state trips the forward-progress guard (`go test ./internal/lexer/ -run '^TestForwardProgressGuardFires$' -count=1 -v | grep -q '^--- PASS: TestForwardProgressGuardFires'`)
- [ ] Positions are monotonic, non-negative, and within bounds (`go test ./internal/lexer/ -run '^TestPositionMonotonicity$' -count=1 -v | grep -q '^--- PASS: TestPositionMonotonicity'`)
- [ ] Concatenating every token's source text reproduces the input exactly (`go test ./internal/lexer/ -run '^TestLossless$' -count=1 -v | grep -q '^--- PASS: TestLossless'`)

Spec: §7.6.2 invariants 1-4, §2.15 item 18. Plan: M0.
