---
title: "t/base/lex.t round-trips: the M0 gate"
state: in-progress
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:42:44.613699558Z
updated: 2026-09-15T19:13:13.087421927Z
sessions:
- start_sha: a28f534a27eb8cf78a2413daaa60fd7d56012f33
  end_sha: ""
  commits: 0
  started_at: 2026-09-15T19:13:13.087421927Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-15T19:13:13.087421927Z
---

M0's exit criterion, and spec §7.3.2 is emphatic:

> Treat `lex.t` as a milestone in its own right, not as one file among 620.
> If it parses, the lexer is real.

15,633 bytes, 129 tests, written specifically to abuse a lexer.

## Round-trip alone cannot gate this

The chain review caught the flaw that made an earlier draft of this issue
worthless. **Invariant 3 — concatenating tokens reproduces the input — holds
for ANY partition of the bytes**, including a single token spanning the
whole file. A lexer that does

    return []Token{{Kind: Trivia, Start: 0, End: len(src)}}

passes every round-trip criterion in this milestone.

Worse, it holds for every near-miss that matters: `=cute` ending POD early,
the `/e` heredoc body lexed as barewords, `$main'a` opening a string at `'a
= $'`. All round-trip cleanly and emit no error token.

So the gate is a **golden token stream**: kind and span for every token in
lex.t, committed, compared exactly. That is what makes "if it parses, the
lexer is real" mean anything at the round-trip bar.

Regenerating the golden file is a deliberate act in the same commit as the
change that moved it, exactly as the parse-fidelity baseline works. A golden
file that regenerates itself on mismatch is decoration.

## What lex.t actually contains

Verified against the real file at the pinned revision:

    line 10    $x = $#[0];              # $# then a subscript, not $# of @[
    line 12    $x = '\\'; # ';          # quoted backslash, then a comment holding a quote
    line 20    eval '$foo{1} / 1;';     # division after a hash subscript
    line 36    print <<'EOF';           # non-interpolating heredoc
    line 41    print <<EOF;             # interpolating heredoc
    line 45    eval <<\EOE, print $@;   # backslash-quoted terminator, mid-expression
    line 56    two heredocs on one line
    line 114   s/^not /substr(<<EOF, 0, 0)/e;   # heredoc inside /e replacement
    lines 590-630   =cute / =cut2 / =cut_ terminator battery

There is no `<<~` in the file.

## Why round-trip and not parse

M0 asks that the bytes survive tokenization with correct boundaries. lex.t
need not PARSE here; that is M1. Keeping the bar at correct delimitation is
what makes M0 finishable while still being a real gate — you cannot delimit
a heredoc you did not recognise, and the golden stream proves the
delimitation rather than assuming it.

## Acceptance

- [ ] lex.t tokenizes with no Error or UnknownRest token (`go test ./internal/lexer/ -run '^TestLexDotTNoErrorTokens$' -count=1 -v | grep -q '^--- PASS: TestLexDotTNoErrorTokens'`)
- [ ] lex.t round-trips byte-for-byte (`go test ./internal/lexer/ -run '^TestLexDotTRoundTrips$' -count=1 -v | grep -q '^--- PASS: TestLexDotTRoundTrips'`)
- [ ] lex.t's token stream matches the committed golden file exactly, kind and span (`go test ./internal/lexer/ -run '^TestLexDotTGoldenStream$' -count=1 -v | grep -q '^--- PASS: TestLexDotTGoldenStream'`)
- [ ] The golden file is only rewritten under an explicit update flag (`go test ./internal/lexer/ -run '^TestGoldenRequiresUpdateFlag$' -count=1 -v | grep -q '^--- PASS: TestGoldenRequiresUpdateFlag'`)
- [ ] Each trap above is pinned as its own named case (`go test ./internal/lexer/ -run '^TestLexDotTTraps$' -count=1 -v | grep -q '^--- PASS: TestLexDotTTraps'`)
- [ ] The test skips with a clear message when the corpus is absent (`go test ./internal/lexer/ -run '^TestLexDotTSkipsWithoutCorpus$' -count=1 -v | grep -q '^--- PASS: TestLexDotTSkipsWithoutCorpus'`)

Spec: §7.3.2. Plan: M0 exit criterion.
