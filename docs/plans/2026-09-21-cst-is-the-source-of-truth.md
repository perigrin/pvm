# The CST is the source of truth, and serializing it is the test

**Status:** design, not decomposed. Written 2026-09-21.

## The rule

The CST is the only source of truth about a file's text. If the tree cannot
be serialized back to the source, the tree does not hold the source, and that
is the whole statement — not a test result, a property.

Two consequences, both structural:

- A node answers "what text am I" without being handed the file. No `src`
  parameter.
- A gap is a MISSING NODE, which is visible in the tree, rather than a span
  arithmetic error, which is not.

## What is wrong with the current design

`Node.SourceText(src []byte)` reconstructs a node's text by walking its
children and emitting the gaps between them:

    at := n.Start
    for _, c := range n.Children {
        if c.Start > at { b = append(b, src[at:c.Start]...) }
        walk(c)
        at = c.End
    }
    if n.End > at { b = append(b, src[at:n.End]...) }

Two things claim to describe the same region: the byte span, and the child
structure. Nothing forces them to agree.

**Measured 2026-09-21 over T1's 986 files, 286,169 nodes: the walk and
`src[n.Start:n.End]` never disagree.** Not once. So the walk is a second
implementation of a fact the span already carries.

That is worse than redundant. A node whose children escaped its span would
emit wrong bytes, and `TestParsedFilesRoundTrip` would not catch it, because
the round-trip test calls the same walk. A function validated against itself.

## What slice-only means

    type Node struct {
        Text      string  // a subslice of the one source buffer -- the truth
        Line, Col int     // recorded by the lexer, which walks every byte anyway
        Kind      Kind
        Children  []*Node
        ...
    }

`Text` is `src[a:b]` — an ordinary Go subslice. No copy, no `unsafe`, bounds
checked. Every leaf's string header points into one allocation, so the source
is interned by construction.

**Size is roughly neutral.** `Start, End int` is 16 bytes; a string header is
16. Node is currently 72 bytes, with a measured note (parse.go:256) that five
bools fit in padding it already had.

**No `unsafe` anywhere.** An earlier sketch recovered the byte offset from the
slice by pointer arithmetic; that is unnecessary. Positions are RECORDED at
parse time rather than derived afterwards — the lexer already knows the line,
and `sites.go` currently builds a `lineIndex(src)` and binary-searches it per
file, which the recorded form removes.

## The finding that sizes the work

Serialization under slice-only is concatenation of leaves in document order.
That requires every byte to be in a leaf.

**It is not. Measured over the same 986 files: leaf concatenation reproduces
the source in ZERO of them.** The first file loses 135 of 746 bytes.

Interior nodes hold bytes their children do not: the `+` in `$a + $b`, the
`;`, the parens of a call. The current walk's comment says so — "the gaps
between children are the parent's own bytes and are emitted in place, which
is why an operator is NOT made a child node: an operator is not an operand,
and a tree that said so would force every consumer to filter it back out."

So the design has a real cost, and it is the one that needs deciding:

- **Operators become leaves.** Every byte is in a leaf, serialization is
  concatenation, the invariant is structural. The comment above is the
  argument against: consumers that want operands would filter operators back
  out. Note that `Binary.Text` already stores the operator, so the
  information is duplicated today regardless.
- **Interior nodes carry their own text fragments.** A node holds the slices
  between its children. Keeps operators out of the child list, but
  reintroduces two places text lives.

The first is the honest version of "the CST is the truth". The second is a
smaller change that keeps the current shape.

## The test, and the order it is built in

**Serializing the CST is the destination.** The round-trip oracle is the
safety net used BEFORE the CST is known to be right.

    now         run perl on src, run perl on canon(parse(src)),
                compare behaviour. Perl adjudicates; the tree is not
                trusted.

    destination serialize the CST and compare to src. No perl, no
                emitter, no oracle. The tree answers for itself.

The order matters because the first does not depend on the tree being
correct, and the second only means something once it is. A serialization test
written today would pass against a tree that had lost the source, because the
same walk produces both sides.

This is B::SoN's shape one level up. Their lowering gate runs the original,
lowers to IR, deparses back, runs the emission and compares OUTPUT — the IR
is never trusted structurally, perl adjudicates. Measured on their side: 350
test files, 290 structural and 60 behavioural, and the behavioural layer was
added after structural-only shipped six live wrong answers from structurally
plausible graphs.

So both layers, and in that order. Behavioural first as the safety net,
structural serialization as the destination.

## What this does not change

`canon` stays, and stays a different operation. Serialization is "concatenate
what is there"; canon is "emit what the structure implies, with the
parenthesisation the TREE says". A mis-grouped tree serializes perfectly and
emits differently — that disjointness is the point of having both
(§7.2(c), and `wrongtree_test.go` has four measured examples).

Line/column recording does not answer the LSP's position question by itself:
LSP speaks UTF-16 code units, which is a conversion either way.

## Related, not blocking

`canon.go` reads the parser's own `infix[]` and `prefix[]` tables at eight
sites, and the coupling was documented as deliberate — "the emitter asks the
same question the parser did". B::SoN named the failure mode: a shared
assumption cancels out, so one wrong binding power mis-groups going in and
un-mis-groups coming out, and the round trip agrees with itself.

Their recommended check needs no refactor and does not involve the emitter:
for each adjacent precedence pair, a fixture whose two groupings produce
different observable output, run under perl. That validates the table against
perl rather than against ourselves.

Caution recorded from them, verified here: comparison operators CHAIN on
5.32+, they are not nonassoc. `3 > 2 > 1` is 1 and `(3 > 2) > 1` is 0. Our
table already knows this (`chain.go` has chRelop/chEqop, and `CmpChain` is a
node kind), so the concern does not apply to us — but an error-shaped fixture
would have confirmed a stale table if it had.

## Open

- Operators as leaves, or interior nodes carrying fragments? The measurement
  says the current tree cannot serialize by concatenation either way; this
  decides which shape fixes it.
- Does anything need a byte offset rather than line/col? The only current
  consumers are `internal/infer` and `internal/psc`, both of which are being
  rewritten, so they do not constrain it.
- 89 uses of `.Start` inside `internal/parse`. Most are construction; the
  count is the migration size, not a design question.
