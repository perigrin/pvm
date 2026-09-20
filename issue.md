---
title: Decide what a parenless call to an unknown callee does with its arguments
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-20T23:55:47.309990102Z
updated: 2026-09-20T23:55:47.309990102Z
---

`internal/parse/call.go`'s `default:` arm decides what a bareword this parser
has never seen does with what follows it. Today it decides NOTHING, and that
one decision is most of the T2 gap.

## Measured, 2026-09-20 at b6e7c351

The survey (01a0c10b) bucketed this as "`%` in a parenless call" plus "a bare
number argument". Both bucketings are wrong, and the correction matters
because it changes what a fix has to cover:

    ok 8;              1 Unknown      ok $x;        1 Unknown
    ok "s";            1 Unknown      ok 8,'name';  1 Unknown
    f 13 % 4;          1 Unknown      f $a % 4;     1 Unknown
    ok($x);            0              print "x";    0   (print is a builtin)

It is not about `%`. It is not about numbers. **Any parenless call to a sub
this parser does not know, with any argument at all, falls to Unknown.** The
`%` case is one flavour; `ok 8` is another; there is one cause.

Scale: ~330 of T2's 649 Unknown nodes, all 179 of `opbasic/arith.t`, and the
~151 rows the survey attributed to `ok`/`is`/`print`.

## The code, and why it is not a typo

`internal/parse/call.go:160-175`:

    default:
        // A bareword this parser does not know. It might be a user sub taking
        // a list, a class name, or a hash key -- perl decides with the symbol
        // table and this parser cannot.
        //
        // §4.8.3: produce a Call with Resolved false rather than an Unknown.
        //
        // No arguments are consumed. Guessing that a bareword takes a list
        // would swallow the rest of the statement on every hash key and
        // class name in the corpus.
        n.Resolved = false

The `Call{Resolved:false}` node IS emitted, correctly, exactly as
`docs/specs/perl-parser/04-expressions.md` §4.8.3 requires and as the file's
own ABOUTME promises. What is left undone is the ARGUMENTS: nothing consumes
them, the statement loop then finds unconsumed tokens, and the enclosing
STATEMENT falls to Unknown. The Call is right and its container is lost.

The comment is an argument, not an oversight. §4.8.3 says NOAMP is emitted
"only when the bareword is a known sub", and calls this "the fundamental
undecidability". Consuming a list here is a guess against that.

## This issue is a DECISION, not an implementation

Three routes, and the cost of each is what has to be measured before one is
picked. **Do not pick by taste; pick by the numbers below.**

**A. Consume a list greedily.** Every parenless call parses. Cost: a bareword
that is NOT a call — a hash key `$h{foo}`, a class name `Foo->new`, a fat
comma LHS `foo => 1` — swallows the rest of its statement into a Call that is
a WRONG TREE, not merely a wider one. M1's gate is "Oracle WRONG: 0" and this
is the one route that can break it. Most such barewords are already handled
elsewhere (subscripts, `->`, `=>`); the question is what the residue costs.

**B. Keep consuming nothing.** No false parses, WRONG stays 0, and ~330
Unknowns and all of arith.t stay. This is the status quo; naming it as a
choice is the point.

**C. Widen what counts as "known".** `p.knowsShape(text)` already exists
(`call.go:152`, `shape.go:149`) and is true for a sub whose PROTOTYPE was
read from its declaration — imports resolved by 01a0ad3f. A sub DECLARED IN
THE FILE BEING PARSED, or imported with no prototype but a visible `sub`
keyword, is just as known and today is not. This is not a guess: perl's own
rule is "already declared", and a `sub ok {...}` or a resolved `Test::More`
export IS a declaration this parser has seen. `ok`, `is`, `like`, `pass` all
come from `Test::More`, which 01a0ad3f already resolves.

C is the route that matches §4.8.3's rule rather than working around it. It
is listed third because it must be MEASURED, not assumed: if C recovers most
of the ~330 then A's risk is never worth taking.

## Steps

- [ ] Write a test that records the CURRENT behaviour of a parenless
      unknown-callee call with an argument, one case per shape above
      (`ok 8`, `ok $x`, `ok "s"`, `ok 8,'name'`, `f 13 % 4`), asserting the
      statement is Unknown. Run it; it passes — it is the baseline the
      decision moves.
- [ ] Measure route C alone: extend `knowsShape` to cover subs declared in
      the file under parse and subs imported from a module whose source was
      read (prototype or not), behind a temporary build-local toggle if that
      is cheapest. Record T2 clean-file count, T2 Unknown total, T1 ratchet
      delta, and T1 delta in BYTES.
- [ ] Measure route A alone the same way, and additionally count how many
      T1/T2 sites become a Call swallowing a hash key, class name, or fat
      comma. Record whether the parse oracle's WRONG bucket moves.
- [ ] Record both measurements in this issue and pick the route. If C covers
      most of the gap, A is not taken.
- [ ] Open the implementation issue's work against the chosen route (this
      issue blocks it), and delete any temporary toggle.

## Acceptance Criteria

### Positive Scenarios

- [ ] The baseline test naming every shape of the parenless-unknown-callee
      failure exists and passes (`go test ./internal/parse/ -run '^TestParenlessUnknownCalleeBaseline$' -count=1 -v | grep -q '^--- PASS: TestParenlessUnknownCalleeBaseline'`)
- [ ] The measurement that chose the route is committed as a note beside the
      code it governs, so the tradeoff is in the history rather than only in
      a comment (`grep -q 'route' internal/parse/call.go`)
- [ ] The chosen route is named in this issue with the measurement that chose
      it, and the whole T2 suite is still green at decision time (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)

## Why this is its own issue

The comment in `call.go` is a design position held for a measured reason —
it names the hash-key and class-name failure it is avoiding. Overruling it
inside an implementation issue would make the tradeoff invisible in the
history, which is how the comment came to be the only record of it.

Spec: §4.8.3. Plan: M1. Survey: 01a0c10b.
