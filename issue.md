---
title: "Named unary and list operators: the keyword table M0 deferred"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a75b-7c7b-78ca-b47f-40cd34c5c963
blocks:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-15T21:53:08.668833928Z
updated: 2026-09-15T23:17:33.614856901Z
---

Named unary operators and list operators: where perl's precedence stops being
a table and starts consulting the keyword table.

Depends on the lexer-feedback issue, which builds the table. This issue is
the parser side: how the table's classification changes the shape of the
tree.

## What the table buys

Named unaries bind tighter than comparison but looser than arithmetic.
Measured on the optree rather than on Deparse, which round-trips the surface
form and hides the grouping:

    $ perl -MO=Concise -e 'my $x = "abc"; print length $x + 1'
    a  <1> length[t3] sK/1 ->b
    9     <2> add[t2] sK/2 ->a          # length(add($x, 1))

    $ perl -MO=Concise -e 'my $x = "abc"; print length ($x) + 1'
    a  <2> add[t3] sK/2 ->b
    8     <1> length[t2] sK/1 ->9       # add(length($x), 1)

`toke.c`'s `UNI3` (285) returns FUNC1 when the next character is `(` and
UNIOP otherwise; `S_lop` (2174) is the list-operator side. Spec §4.8.1 calls
this the paren cliff and §4.8.2 gives how far a list operator swallows.

## What it cannot buy

A user sub whose prototype changes its parse. That is the srefgen signal the
fidelity harness measures and it stays unresolved until M4.

Per §4.8.3, "unresolved" here does NOT mean Unknown:

> The expression parser's job is to produce a `Call` node with
> `Resolved: false` and let a later pass decide.

A `Call{Resolved: false}` is a real node with a known shape and an honest
flag. It is what the subject adapter already emits as an `Unresolved` site
(`adapter.go:119`), which scores `wider`. Unknown is for constructs with no
known shape at all; a call is not one of those.

## Feature-gated keywords

§0.13 rank 6: `try`, `defer`, `method`, `true` used as ordinary identifiers
in 5 corpus files. A table that does not know they are feature-gated turns
valid old code into a parse error. The feature state itself — threading
`use feature` through the parse — is M2's; this issue needs only that an
unenabled keyword is an identifier.

## Acceptance

- [ ] Named unaries bind tighter than comparison, looser than arithmetic, asserted on tree shape (`go test ./internal/parse/ -run '^TestNamedUnaryPrecedence$' -count=1 -v | grep -q '^--- PASS: TestNamedUnaryPrecedence'`)
- [ ] `length ($x) + 1` and `length $x + 1` produce different trees, per UNI3 (`go test ./internal/parse/ -run '^TestNamedUnaryParenForm$' -count=1 -v | grep -q '^--- PASS: TestNamedUnaryParenForm'`)
- [ ] A list operator swallows to the statement end or closing paren (`go test ./internal/parse/ -run '^TestListOperatorExtent$' -count=1 -v | grep -q '^--- PASS: TestListOperatorExtent'`)
- [ ] Post-5.36 keywords parse as identifiers where the feature is off (`go test ./internal/parse/ -run '^TestFeatureGatedKeywords$' -count=1 -v | grep -q '^--- PASS: TestFeatureGatedKeywords'`)
- [ ] A call to an unknown sub is `Call{Resolved:false}`, not Unknown, per §4.8.3 (`go test ./internal/parse/ -run '^TestUnknownSubIsUnresolvedCall$' -count=1 -v | grep -q '^--- PASS: TestUnknownSubIsUnresolvedCall'`)
- [ ] All four `&foo` forms of §4.8.3 parse distinctly (`go test ./internal/parse/ -run '^TestAmpersandCallForms$' -count=1 -v | grep -q '^--- PASS: TestAmpersandCallForms'`)

Findings: §0.13 rank 6 (5 files). Spec: ch4 §4.8.1, §4.8.2, §4.8.3; ch3 §3.8.1. toke.c:285, 2174.
