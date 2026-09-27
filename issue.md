---
title: "An arrow code-deref call loses its parens: `__SUB__->($v->{k})` re-emits as a subscript chain"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-27T00:30:17.267241755Z
updated: 2026-09-27T00:30:17.267241755Z
---

## Measured finding

`X->(ARGS)` -- a code-deref call whose argument contains an arrow -- loses the
call's parentheses, so the emission re-reads as a subscript on the callee.
Measured at `34737c46` in a clean `git archive` of HEAD, so this is not any
in-flight change:

    source                              first canon                        second canon
    my $x = __SUB__->($value->{$_});    __SUB__() -> $value->{$_}          (__SUB__() -> $value)->{$_}
    f(__SUB__->($value->{$_}));         f(__SUB__() -> $value->{$_})       f((__SUB__() -> $value)->{$_})

The FIRST emission is already wrong: `__SUB__->($value->{$_})` calls `__SUB__`
with one argument, and `__SUB__() -> $value->{$_}` is a subscript chain on
`__SUB__()`'s result. The second emission then groups it differently again, so
canon is not a fixpoint and the tree is not a parse of its source.

`->` followed by `(` is a CODE DEREF CALL and the parens are the argument
boundary, not grouping -- the same distinction `parseSubscript` already makes
for `->[` and `->{`, where `expr.go`'s `->` branch handles `[` and `{` and
falls through for `(`.

perl reads it as a call, measured on 5.42.0 at `/home/perigrin/.local/bin/perl`:

    $ perl -MO=Deparse -e 'use feature "current_sub"; sub f { __SUB__->($_->{k}) }'
    sub f { __SUB__->($_->{'k'}); }

## How it was found

Found by fixing 01a0e013, whose canon defect had been DELETING every `return`
operand -- `return EXPR;` canon'd as `return ;`. With the operand emitted, this
construct became visible for the first time. `current_sub_map_block.t` in the
T1 corpus is the file: line 10 is

    return {map { $_ => __SUB__->($value->{$_}) } keys %$value};

and it was in `t1canon.ratchet` at 0 only because BOTH emissions deleted the
whole operand, which made a wrong tree look like a fixpoint. It is `0 -> 1`
after 01a0e013 and the ratchet records why.

## Acceptance criteria

- [ ] `my $x = __SUB__->($value->{$_});` canons to itself and is a fixpoint `go test ./internal/parse -run TestArrowCodeDerefCall -count=1`
- [ ] `current_sub_map_block.t` returns to 0 in `t1canon.ratchet` and rejoins `TestCanonTokenIdentity` `go test ./internal/parse -run 'TestCanonRatchet|TestCanonTokenIdentity' -count=1`
- [ ] REGRESSION GUARD: `$h->{k}`, `$a->[0]` and `$r->()` with a simple argument are unchanged `go test ./internal/parse -run TestDeref -count=1`
- [ ] the whole tree is green `go test ./... -count=1`
