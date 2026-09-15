---
title: "AST skeleton: nodes, spans, and the Unknown that keeps WRONG at zero"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-15T21:52:57.027412574Z
updated: 2026-09-15T23:14:47.001202849Z
---

The parser package does not exist. This issue creates it with the node type,
the Pratt loop's skeleton, the `Unknown` node, and the rule that makes Unknown
mean something to the fidelity harness — and nothing else.

## Why Unknown comes first

The plan states M1's shape in one sentence:

> Note the shape: parse-rate is a *ratchet*, WRONG is a *gate*. Constructs
> not yet understood must be emitted as explicit `Unknown` nodes preserving
> their source text — **not guessed at**. That is what keeps WRONG at zero
> while coverage is still low.

## Unknown does NOT score no-answer on its own

An earlier draft of this issue said it did. That was wrong, and the mistake
mattered enough to state plainly. Measured in `compare_facts.go:222-229`:

    case owner != nil && owner.decided[m] > 0:   // agreement
    case owner != nil && owner.hedged[m] > 0:    // wider
    default:                                     // WRONG

An `Unknown` node emits no sites, so `owner` is nil for any perl site inside
it and the verdict is WRONG. No-answer is FILE-level only — `Declined`,
`!OK`, or no site of a marker's kind anywhere in the file. Once the subject
reports one hash site, every rv2hv in that file is scored per statement.

So an Unknown-heavy tree is WRONG-heavy, which is the opposite of the
discipline the plan describes.

## The rule that makes it work

For every statement whose subtree contains an Unknown, the subject adapter
emits an `Unresolved: true` site of EACH marker kind, spanning that
statement. Those land in `hedged` and score `wider` — an honest "something
here I could not settle" rather than a claim that nothing is there.

The mechanism already exists: `adapter.go:119` sets `Unresolved` for
unqualified calls today, and `subject.go:76` documents it as "the same site
with Unresolved set says 'I could not settle it'". This issue extends it to
Unknown subtrees.

That rule IS the reason Unknown exists. Without it, Unknown is not a refusal;
it is a silent wrong answer.

## Shape

    package parse   // internal/parse

    type Node struct {
        Kind       Kind
        Start, End int        // byte offsets, as the lexer uses
        Children   []*Node
    }

`internal/parse`, not `internal/ast`: `internal/parser` is the existing
tree-sitter wrapper and will be for some time, and two packages differing by
two letters is a trap every future reader springs once.

Byte offsets for the reason M0's Token gives: line and column are derived at
the LSP layer, and carrying them doubles the state that must stay consistent
under edit.

An Unknown node spans its source exactly, so M0's round-trip property
survives into the tree: concatenating every leaf's text reproduces the input.

## Acceptance

- [ ] An empty input parses to an empty source_file node (`go test ./internal/parse/ -run '^TestParseEmpty$' -count=1 -v | grep -q '^--- PASS: TestParseEmpty'`)
- [ ] An Unknown node spans exactly the bytes it could not parse (`go test ./internal/parse/ -run '^TestUnknownSpansItsSource$' -count=1 -v | grep -q '^--- PASS: TestUnknownSpansItsSource'`)
- [ ] Concatenating every leaf reproduces the input, Unknown included (`go test ./internal/parse/ -run '^TestTreeRoundTrips$' -count=1 -v | grep -q '^--- PASS: TestTreeRoundTrips'`)
- [ ] A statement containing an Unknown yields Unresolved sites of every marker kind (`go test ./internal/parse/ -run '^TestUnknownHedgesEveryMarker$' -count=1 -v | grep -q '^--- PASS: TestUnknownHedgesEveryMarker'`)
- [ ] Such a statement scores wider, not WRONG, against a perl site inside it (`go test ./internal/parse/ -run '^TestUnknownScoresWider$' -count=1 -v | grep -q '^--- PASS: TestUnknownScoresWider'`)
- [ ] Every parse step consumes at least one token, asserted structurally (`go test ./internal/parse/ -run '^TestParserForwardProgress$' -count=1 -v | grep -q '^--- PASS: TestParserForwardProgress'`)
- [ ] A fuzz target runs 1M execs with no panic and no lost bytes (`make fuzz-parser`)

Plan: M1. Spec: ch4, ch5, §7.5.5. `compare_facts.go:222-229`.
