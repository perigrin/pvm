---
title: A labelled bare block is a statement, and last LABEL inside it resolves
state: pending
urgency: normal
milestone: perl-git-t-parses
blocked_by:
- 01a0de8a-7043-74ea-a42e-e2db217946ec
created: 2026-09-26T16:27:39.730386781Z
updated: 2026-09-26T16:27:39.730386781Z
---

`SKIP: { ... }` -- a bare block with a label -- refuses, and `last LABEL` inside
one costs a second Unknown. Valid perl, and present in 148 of perl.git t/'s 620
files.

## Measured in isolation, 2026-09-26 at `59029a2c`

    SKIP: {                          Unknown=1
        print "in\n";
    }

    SKIP: {                          Unknown=2
        last SKIP;
    }

    {                                Unknown=0   <- an UNLABELLED bare block parses
        print "in\n";
    }

So it is the LABEL, not the block. The unlabelled form has worked throughout.

Perl accepts both, verified on 5.42.0 at `/home/perigrin/.local/bin/perl`:

    perl -e 'SKIP: { print "in\n"; last SKIP; }'      ->  in

`last LABEL` targeting a bare block is the idiom perl's own suite uses for
skipping, which is why it is everywhere in `t/`.

## Reach

    148 of 620 files contain a labelled bare block
     83 of op/'s 227 contain one

`op/` is 37% of this goal's population and 32% of its dirty files, so a
construct in a third of it is worth doing early. BUT the share of the 1,118
bare-brace Unknowns that a labelled block actually owns is measured by
`01a0de8a-7043` and not by this issue -- do not start until that lands, and if
it reports a small share then this issue is not the first fix and should be
re-ordered rather than executed.

## Where the parser stands today

A `Label` handling path exists -- `withLabels` is the funnel every statement form
passes through, and `parseLabels` reads them. The gap is narrower than "labels
are unimplemented": measure what `withLabels` does with a label whose next token
is `{`, because a plain bare block already parses and a labelled one does not.

`01a0dd71` (ADJUST) and `01a0d087` (ungated defer) both turned out to be a brace
group after a word being read as an index call rather than a block, and
`01a0ddac` records a third spelling of the same thing inside a class body. A
label followed by `{` may be a fourth. Measure before assuming, and if it IS
that family, say so -- the fix may be shared.

## Scope

1. Measure what the parser does with `LABEL: {` today: which node kinds, and
   whether canon re-emits it. `parse.Canon(n, src)` and a token dump.
2. The label attaches to the block as a statement, so `last LABEL` and
   `next LABEL` inside it resolve. Both spellings are in perl's suite.
3. Check `redo LABEL` too -- perl allows it on a bare block and it is a third
   spelling with the same shape.
4. Canon must re-emit the label. `TestCanonTokenIdentity` and the canon fixpoint
   have bitten four workers this milestone; run
   `go test ./internal/parse/ -run '^TestCanon' -count=1` EARLY.

## Context

- paths: `internal/parse/parse.go` (`withLabels`, `parseLabels`,
  `statementKeywords`), `internal/parse/control.go` (`parseLoopControl`),
  `internal/parse/canon.go`
- commands: `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- related: `01a0dd71`, `01a0d087`, `01a0ddac` -- three spellings of a brace
  group after a word

## Acceptance criteria

- [ ] A labelled bare block parses (`go test ./internal/parse/ -run '^TestLabelledBareBlock$' -count=1 -v | grep -q '^--- PASS: TestLabelledBareBlock'`)
- [ ] `last LABEL` inside one resolves (`go test ./internal/parse/ -run '^TestLabelledBareBlock$' -count=1 | grep -q '^ok'`)
- [ ] Canon re-emits the label and holds its fixpoint (`go test ./internal/parse/ -run '^TestCanon' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: an unlabelled bare block still parses (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
