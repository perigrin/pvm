---
title: A statement modifier on a loop control is dropped, and the tree is wrong with zero Unknowns
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T18:09:45.133282332Z
updated: 2026-09-26T18:09:45.133282332Z
---

Every statement modifier on a loop control produces a WRONG tree with ZERO
Unknown nodes, which is why nothing has caught it. `parseGoto` and
`parseLoopControl` both decline the modifier on purpose so the caller can apply
it, and the caller never does.

## Measured 2026-09-26 at `9325864f`, and identical after `01a0de8b`

The defect is independent of that issue's change -- measured either side of it
with the same result -- and was found because the labelled-block fix stopped
hiding `t/class/field.t:289` inside one Unknown.

    source                canon                      Unknown
    goto HERE if $x;      goto HERE; if $x;          0
    last if $x;           last if; $x;               0
    next LOOP if $x;      next LOOP; if $x;          0
    return 1 if $x;       return 1; if $x;           0

Three statements come out of one, and `if` arrives as a Conditional with NO
children. `last if $x;` is worse still: `if` is taken as the LABEL.

perl parses all four as one statement, measured on 5.42.0 at
`/home/perigrin/.local/bin/perl`:

    $ perl -MO=Deparse -e 'my $x; HERE: ; goto HERE if !$x++;'
    my $x;
    HERE: ;
    goto HERE unless $x++;

## Where it is

`internal/parse/parse.go`, the statement dispatch. The control-flow exit is

    if c := p.parseControlFlow(tok); c != nil {
        return p.withLabels(labels, c, start)
    }

with no `applyModifier` between them. The declaration path has one and the
expression path has one; this path does not.

`internal/parse/control.go:203` says so explicitly, in `parseGoto`'s own
comment:

    A modifier is left alone: `goto HERE if $x` must reach applyModifier, so
    the label test refuses a modifier word.

It refuses the word correctly. Nothing then applies the modifier.

## Why it matters more than the node count says

A zero-Unknown wrong tree is the failure mode this milestone treats as the only
scoring bucket that fails a build (`Kind.Unknown`'s own doc comment). Four
workers this milestone shipped or nearly shipped one. This one is already in the
tree and the ratchets cannot see it, because they count Unknown nodes and this
construct produces none.

## Scope

1. Apply `applyModifier` on the control-flow exit, the way the declaration and
   expression exits do.
2. `last if $x` must not take `if` as its label -- `parseLoopControl`'s label
   test needs the same `modifiers` check `parseGoto` already has, or the fix
   above has nothing to apply the modifier to.
3. Check `unless`, `while`, `until`, `for` and `foreach` as modifiers, not just
   `if`. All six are in `modifiers`.
4. Canon must re-emit ONE statement. `goto HERE if $x;` is currently
   `goto HERE; if $x;` and its canon-of-canon differs again.
5. Count the corpus delta. The construct produces no Unknowns today, so the
   node ratchets will not move and the canon ratchets are where the evidence
   will be.
