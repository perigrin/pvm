---
title: "Range, flip-flop and the line-number form: three operators, one spelling"
state: pending
urgency: normal
milestone: corpus-complete
created: 2026-09-23T02:07:02.512094755Z
updated: 2026-09-23T02:07:02.512094755Z
---

`..` is three operators wearing one spelling, and the corpus claims none
of them. One occurrence exists corpus-wide -- `for my $i (1 .. 2)` in
`07_subroutines/00_adjacency.t` -- and it is loop scaffolding, not a
claim.

Measured under perl 5.42.0, four distinct behaviours from one token:

    my @r = (1..5)                      1 2 3 4 5     list: range
    print if /START/ .. /END/           stateful      scalar: flip-flop
    print if 2 .. 3                     lines 2-3     scalar + constant:
                                                      implicit $. compare
    "a".."e"                            a b c d e     magic string range

Context decides which, and whether an operand is a CONSTANT decides
between the last two. The corpus has a whole tier about context (03) and
does not use it here.

## The measurement that decides how to write it

Constant endpoints FOLD and emit nothing at all:

    my @r = (1..3)          ->  no range, flip or flop op
    my @r = (1..$n)         ->  range flip flop

So a file written over literal endpoints would claim `range` and emit
none, and the lint would have nothing to check. The endpoint must be a
runtime value -- the same trap tier 04 records for arithmetic.

Note the second line: ONE `..` in source emits all three ops. The
optimiser builds the flip-flop machinery for a list range too. That makes
the op set a poor discriminator between the readings, and puts the weight
on OUTPUT, which separates them cleanly.

## Files

Three: the list range with a runtime endpoint; the string range, whose
magic increment is a separate claim (`"az".."bb"` is not arithmetic); and
the flip-flop in scalar context, whose statefulness across lines is the
whole construct and is observable only by feeding it several.

`...` -- the three-dot flip-flop, which differs in whether it can turn
off on the same line it turned on -- is the fourth claim and belongs with
the flip-flop file rather than alone.

Placement is `03_context`: the list-versus-scalar fork IS the construct,
and tier 03 owns discriminating pairs. The tier's README calls a
discriminating pair its entire subject, and this is one.

## Acceptance criteria

- [ ] Tier 03 holds a range file whose endpoint is a runtime value, so the operator is not folded away (`awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/03_context/*.t | grep -q '\.\.'`)
- [ ] The flip-flop's statefulness is pinned by output over several lines rather than asserted (`go test ./internal/conformance/ -run '^TestCorpus/03_context' -count=1 -v | grep -q '^    --- PASS: TestCorpus/03_context'`)
- [ ] Every new file is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierContextPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierContextPerlValidated'`)
- [ ] The dependency lint passes over the tier (`go test ./internal/conformance/ -run '^TestTierContextLint$' -count=1 -v | grep -q '^--- PASS: TestTierContextLint'`)
- [ ] The tier's adjacency file holds every construct it introduces (`go test ./internal/conformance/ -run '^TestTierContextAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierContextAdjacency'`)
- [ ] The whole conformance suite is green with the new files in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
