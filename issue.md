---
title: fc is feature-gated and in no arity table, so a parenless call drops its argument at Unknown=0
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T22:02:44.311964663Z
updated: 2026-09-26T22:02:44.311964663Z
---

The `fc` builtin is feature-gated and is in none of the three parse arity
tables, so `fc $x` emits `fc();$x;` -- an empty call with the argument detached
as its own statement, at Unknown=0. That is the wrong-tree-at-zero defect, and
the canon is the only instrument that sees it.

It was left out of `01a0df71`'s fix DELIBERATELY, because tabling it
unconditionally would make us disagree with perl on unfeatured source, and
`01a0df71` had no feature-state machinery to condition on. This issue records
the disagreement so it is tracked rather than forgotten.

## Measured on perl 5.42.0

Unenabled, perl does not read `fc` as a builtin at all:

    perl -MO=Deparse,-p -e 'our $a; my $z = fc $a;'
      (my($z) = $a->fc);                 a METHOD call

Under the 5.42 bundle it is a named unary -- the comma stays outside:

    perl -MO=Deparse,-p -e 'use v5.42; our ($a,$b); my $z = fc $a, $b;'
      (fc($a), $b);

So there are THREE readings of `fc $x` and we produce a fourth:

    perl, unfeatured     $x->fc          a method call
    perl, featured       fc($x)          a named unary
    us, today            fc();$x;        an empty call, argument detached

## Reach, measured not grepped

ONE file in perl.git t/ has a parenless `fc` at all:

    re/anyof.t:858     $mod_cp = ord fc $char;

`grep -rlE '(^|[^>:_[:alnum:]])fc[[:space:]]+[$@]' /home/perigrin/dev/perl5/t`
returns that one file. So neither reading moves `perlgitt.ratchet`, which is
exactly why this is recorded rather than fixed: the value is the decision, not
the file count.

`evalbytes`, `break` and `default` are gated the same way and absent from all
three tables for the same reason. `isa` is gated too but is an INFIX operator
under `use feature 'isa'`, not a call, so it does not belong in any arity table
at all. All four were measured in the same pass:

    perl -MO=Deparse,-p -e 'our ($a,$b); my $z = evalbytes $a, $b;'
      ($a->evalbytes, $b);               method call, unfeatured
    perl -MO=Deparse,-p -e 'use v5.42; our ($a,$b); my $z = evalbytes $a, $b;'
      (evalbytes($a), $b);               named unary, featured

## Why this is not simply "add it to namedUnary"

The parser has no feature state. `use feature 'fc'`, `use v5.16` and any
bundle at or above it enable the builtin; a file without one gets the method
call. Tabling `fc` gets the featured half right and the unfeatured half wrong,
and leaving it out gets BOTH wrong -- today's `fc();$x;` matches neither perl
reading. So the smallest honest fix is not a table entry: it is either

  (a) table it and accept a known disagreement on unfeatured source, which is
      strictly better than today because one of two readings becomes right, or
  (b) thread feature state to `parseWordTerm` so the same name can be a
      builtin or a method call by scope.

(a) is one line and a comment. (b) is the real answer and also what `say`,
`state` and `class` will each need, so it is worth doing once rather than four
times.

## Acceptance criteria

- [ ] `fc $x` no longer emits an empty call with the argument detached, under whichever reading is chosen (`go test ./internal/parse/ -run '^TestFeatureGatedFcArity$' -count=1 | grep -q '^ok'`)
- [ ] The choice between the method-call reading and the builtin reading is recorded in `internal/parse/keyword.go` beside the entry, with the deparse that justifies it (`grep -qiE 'feature|gated' internal/parse/keyword.go`)
- [ ] REGRESSION GUARD, passes today: the audit's own cases still hold -- `atan2`, `glob` and `readpipe` keep their arguments (`go test ./internal/parse/ -run '^TestParenless' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the perl.git t/ ratchet does not move, since `fc` reaches one file and that file's first failure is elsewhere (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus stays complete (`go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1 | grep -q '^ok'`)
