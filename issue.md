---
title: "An arrow code-deref call loses its parens: `__SUB__->($v->{k})` re-emits as a subscript chain"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-27T00:30:17.267241755Z
updated: 2026-10-11T01:54:41.961781772Z
sessions:
- start_sha: 24bfac4a5293052e7c38dd54a56a58aef818e639
  end_sha: ""
  commits: 0
  started_at: 2026-10-11T00:57:40.914684122Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T00:57:40.914684122Z
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

### Re-measured and fixed (2026-10-11)

Re-measured at pu 24bfac4a: the tree was already right -- the `->` call's
right operand carries the Paren flag -- and canon dropped the parens of any
argument that was not a Term or a List, so `$r->($v->[0])` re-emitted as
`$r -> $v->[0]`, a method call through $v to perl (Deparse: `&$r($v->[0])`).
Fixed in 8a9645bb; the criteria below were bare backtick spans verify could
not run, and now name the test that landed.

## Acceptance criteria

- [ ] an arrow code-deref call keeps its argument parens through canon, `__SUB__->($value->{$_})` and `$r->($v->[0])` included, as perl reads them (`go test ./internal/parse/ -run '^TestArrowCodeDerefCallKeepsItsParens$' -count=1 -v | grep -q '^--- PASS: TestArrowCodeDerefCallKeepsItsParens'`)
- [ ] `current_sub_map_block.t` returns to 0 in `t1canon.ratchet` (`grep -q '^0 current_sub_map_block.t$' internal/parse/testdata/t1canon.ratchet`)
- [ ] the canon ratchet holds (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] REGRESSION GUARD: dereferences with a simple argument are unchanged (`go test ./internal/parse/ -run '^TestCanonTokenIdentity$' -count=1 -v | grep -q '^--- PASS: TestCanonTokenIdentity'`)
