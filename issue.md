---
title: "AST skeleton: nodes, spans, and the Unknown that keeps WRONG at zero"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-15T21:52:57.027412574Z
updated: 2026-09-15T21:53:21.824219143Z
---

The parser package does not exist. This issue creates it with the node type,
the Pratt loop's skeleton, and the `Unknown` node that the whole milestone's
discipline rests on — and nothing else. No Perl-specific parsing yet.

## Why Unknown comes first

The plan states M1's shape in one sentence:

> Note the shape: parse-rate is a *ratchet*, WRONG is a *gate*. Constructs
> not yet understood must be emitted as explicit `Unknown` nodes preserving
> their source text — **not guessed at**. That is what keeps WRONG at zero
> while coverage is still low, and it is the discipline the whole ladder
> rests on.

A parser that guesses at an unfamiliar construct produces a tree that is not
a parse of its source, and the fidelity harness scores that WRONG — the only
bucket that fails a build. A parser that says "I do not know what this is,
and here are its bytes" scores no-answer, which is an honest refusal.

M0 learned this the hard way in miniature: `internal/parser`'s tree-sitter
grammar silently shreds a statement and reports a clean parse
(`op/universal.t`, `comp/our.t` — both baselined WRONG today). Unknown is how
this parser declines to repeat that.

## Shape

    package ast   // internal/ast

    type Node struct {
        Kind     Kind
        Start, End int        // byte offsets, as the lexer uses
        Children []*Node
    }

Byte offsets and not line/column, for the reason M0's Token gives: line and
column are derived at the LSP layer, and carrying them doubles the state that
must stay consistent under edit.

An `Unknown` node spans its source exactly, so the round-trip property M0
established survives into the tree: concatenating every leaf's text
reproduces the input. That is the parser-level analogue of invariant 3, and
it is what lets an LSP show correct text for a construct nobody has taught
the parser yet.

## Acceptance

- [ ] An empty input parses to an empty source_file node (`go test ./internal/ast/ -run '^TestParseEmpty$' -count=1 -v | grep -q '^--- PASS: TestParseEmpty'`)
- [ ] An Unknown node spans exactly the bytes it could not parse (`go test ./internal/ast/ -run '^TestUnknownSpansItsSource$' -count=1 -v | grep -q '^--- PASS: TestUnknownSpansItsSource'`)
- [ ] Concatenating every leaf reproduces the input, Unknown included (`go test ./internal/ast/ -run '^TestTreeRoundTrips$' -count=1 -v | grep -q '^--- PASS: TestTreeRoundTrips'`)
- [ ] Parsing never panics on any input, including half-typed buffers (`go test ./internal/ast/ -run '^TestParseNeverPanics$' -count=1 -v | grep -q '^--- PASS: TestParseNeverPanics'`)
- [ ] Every parse step consumes at least one token, asserted structurally (`go test ./internal/ast/ -run '^TestParserForwardProgress$' -count=1 -v | grep -q '^--- PASS: TestParserForwardProgress'`)

Plan: M1. Spec: ch4 (the Pratt table), ch5 (statements).
