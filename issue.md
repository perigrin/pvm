---
title: The last-semicolon forgiveness hides a hash misread as a block
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T04:11:53.607169795Z
updated: 2026-10-24T04:11:53.607169795Z
---

Found by the independent review of NL-c (01a1692f-bf96, exec/nlc 3de59a88) on 2026-10-24. **This needs perigrin's ruling.** It would replace one of the two forgivenesses he ruled into Faithful's closed set (round 7).

## Context

- paths: `internal/parse/fidelity.go` (the last-semicolon rule), `internal/parse/canon.go`, and the parser's Block node
- entrypoints: `faithfulTokens`' last-semicolon rule

The last-semicolon forgiveness can hide a hash or a subscript that psc misreads as a block.
- A hand-built tree for `$h = {a => 1};` with its AnonHash replaced by a Block: canon writes `$h = {a => 1;};`, and Faithful returns true.
- At token level, `$h{a}` against `$h{a;}` is also forgiven.
- perl 5.44 rejects both `{ a => 1; }` and `$h{a;}` as syntax errors.

So Faithful, and every ratchet built on it, cannot see a block/hash misparse.

The reviewer's proposed fix follows the call-paren precedent: the tree records whether the source wrote a block's last `;` (as ArgParen records call parens), canon writes it back, and the last-semicolon rule is deleted. That would leave Faithful forgiving only extra paren nesting (01a1664d-75d4).

A narrower option: forgive the last `;` only inside a node the tree knows is a Block, never inside `{` … `}` that the tokens alone can't classify.

## Acceptance criteria

### Positive Scenarios

- `go test ./internal/parse/ -run '^TestFaithfulCatchesBlockForHashMisread$' -count=1 -v | grep -q '^--- PASS'` (a hand-built AnonHash-as-Block tree is unfaithful)
