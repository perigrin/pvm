---
title: atan2 and fc are in no arity table, so a parenless call drops its arguments at Unknown=0
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T20:39:01.402801113Z
updated: 2026-10-16T18:48:29.872161389Z
sessions:
- start_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  end_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  commits: 0
  started_at: 2026-10-11T10:56:05.288425847Z
  ended_at: 2026-10-11T10:56:05.486621395Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T10:56:05.288425847Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-11T10:56:05.486621395Z
---

`atan2` and `fc` are in none of the three parse arity tables, so a parenless
call to either DROPS ITS ARGUMENTS and scores Unknown=0. The arguments become a
separate statement.

Found by bucketing the first Unknown of every dirty file at `35549df0`, then
isolating the four tail shapes that looked like one cause.

## Measured at `35549df0`

    source                          U   canon
    my $d = atan2 $var, 1 ;         0   my $d = atan2();$var , 1;
    my $c = ord fc $char;           0   my $c = ord(fc());$char;

Both score ZERO and both are wrong. `atan2()` takes no arguments and `$var , 1`
becomes a statement of its own -- the argument list is detached, not refused.

Perl reads both as ordinary calls:

    perl -MO=Deparse -e 'my $var=1; my $d = atan2 $var, 1;'
      ->  my $d = atan2($var, 1);

    perl -MO=Deparse -e 'my $x = atan2 1, 2;'
      ->  my $x = 0.46364760900080609352;      constant-folded, so BOTH args were taken

That is the seventh zero-Unknown-and-wrong finding in this repo, and the third
where canon is the only instrument that could see it.

## The boundary, measured over 48 builtins

Exactly two are affected. Probing every two-argument builtin parenless
(`my $z = NAME $a, $b;`) and every named unary (`my $z = NAME $a;`), and
looking for the signature `NAME();` in the canon:

    1 of 33 two-arg builtins drop their arguments:   atan2
    1 of 15 one-arg builtins drop its argument:      fc

`index`, `sprintf`, `join`, `crypt`, `vec`, `substr`, `push`, `sort`, `grep`,
`map`, `die`, `warn`, `printf`, `open`, `pack`, `unpack`, `seek`, `reverse`,
`rename`, `symlink`, `kill`, `waitpid`, `chmod`, `chown`, `utime`, `link`,
`bind`, `connect`, `split`, `rindex`, `unshift` are all CORRECT, as are `ord`,
`uc`, `lc`, `chr`, `length`, `ref`, `defined`, `quotemeta`, `hex`, `oct`, `log`,
`sqrt`, `int` and `abs`. So this is not a systemic arity problem: it is two
names missing from two tables.

## The cause is an absence, not a misclassification

    internal/parse/keyword.go:40    namedUnary
    internal/parse/keyword.go:73    listOperator
    internal/parse/keyword.go:122   niladicParse

`grep -n 'atan2\|"fc"' internal/parse/keyword.go` finds NOTHING. `atan2` belongs
in `listOperator` beside `index`, `crypt` and `vec`, which are all two-argument
builtins already tabled there. `fc` belongs in `namedUnary` beside `lc` and `uc`.

`internal/lexer/keyword.go`'s `niladic` table is a separate thing and is
CORRECT here -- its comment already says `fc` "takes arguments and drops out"
under `use v5.38`, and `fc` is indeed absent from it. The lexer knew; the parser
tables did not.

## `fc` has a second, separate subtlety -- do NOT conflate them

Unenabled, perl does not read `fc` as a builtin at all:

    perl -MO=Deparse -e 'my $char="a"; my $c = ord fc $char;'
      ->  my $c = ord $char->fc;                    a METHOD call

    perl -MO=Deparse -e 'use feature "fc"; my $c = ord fc $char;'
      ->  my $c = ord fc $char;                     the builtin

So `fc` is feature-gated, exactly as `internal/lexer/keyword.go`'s comment
describes for `say`, `state`, `class` and friends: an unenabled keyword is an
ordinary bareword. Tabling `fc` unconditionally in `namedUnary` would make us
disagree with perl on unfeatured source. Measure which behaviour the corpus
needs before choosing; `atan2` carries no such gate and is the unambiguous half.

## Reach

Small, and that is the point. Two files in perl.git t/ have one of these as
their FIRST failure (`$dummy = atan2 $var, 1 ;` at 11 nodes, and
`$mod_cp = ord fc $char;` at 1). The value here is not the file count -- it is
that a table nobody has audited produced a confident wrong tree, and the same
audit finds anything else missing.

The long tail is now the shape of this goal: at `35549df0` there are 229 dirty
files across 170 DISTINCT first-failure shapes, with the two largest buckets
(`)` at 28 files and `}` at 26) being cascade rather than cause. No single
construct is worth a dedicated fix any more, so an audit that clears a CLASS is
worth more than another one-construct issue.

## Two of my own reach figures were measured wrong, again

The same bucketing pass corrected two counts I filed from greps:

    &$subref      filed 36 files, measured 7
    $::{n} / %::  filed 58 files, measured 9

And a bucket I nearly filed as "old `'` package separator, 33 files" was 1 file
once the classifier stopped lumping every span containing a quote. `$main'a` IS
still valid in 5.42 and silent (verified), but it is one file, not a cluster.
Every reach figure in the design doc taken from `grep -l` should be read as an
upper bound.

## Scope

1. Add `atan2` to `listOperator`. It is a two-argument builtin and the table
   already holds every sibling.
2. Decide `fc` deliberately: it is feature-gated, so tabling it changes how
   UNFEATURED source parses. Perl reads unenabled `fc $x` as `$x->fc`. Either
   match that (harder, needs feature state) or table it and record the
   disagreement with a measurement.
3. AUDIT the three tables against perl's own `regen/keywords.pl` rather than
   fixing two names. The lexer's `niladic` table documents having been built
   that way -- 251 keywords compiled one at a time -- and the parse tables carry
   no such note. Anything else absent is another zero-Unknown wrong tree
   waiting.
4. A corpus case. It must assert the CANON or the token stream, since Unknown=0
   passes today for both spellings.

## Context

- paths: `internal/parse/keyword.go:40` (`namedUnary`), `:73` (`listOperator`),
  `:122` (`niladicParse`), `internal/parse/call.go:95` and `:115` (the
  `n.Resolved` fall-through), `internal/lexer/keyword.go:34` (the `niladic`
  table that is correct, and its provenance comment)
- commands: `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- related: `01a0de11` (lintOps sees op names, not flags -- the same
  instrument-blindness), `01a0deeb` (this goal is gated by no ratchet)

### Re-measured (2026-10-11)

Measured at pu `24bfac4a`. `my $d = atan2 $var, 1 ;` canons as `my $d = atan2($var , 1);` at Unknown=0, as perl Deparses it. `fc` follows perl's feature gate: with `use feature 'fc'`, `ord fc $char` is `ord(fc($char))`; unenabled it refuses (Unknown=1), where perl reads `$char->fc`, so no wrong tree is built. Fixed by c79edeed.

Criterion 3 grepped internal/parse/keyword.go, which d2094988 deleted: the parser's arity tables are now CORE.pmt's shapes. The audit is recorded in internal/parse/declarations/CORE.pmt's header (all 266 names in regen/keywords.pl, each absent one put to perl), and parenlessarity_test.go:2 points to it. The `^ok` greps after `-run` were vacuous; the criteria now grep TestParenlessBuiltinArity's and TestParenlessArityNoDetachedCall's PASS lines.

The conformance and whole-suite commands carry `-timeout=30m`: internal/conformance took 696s in a full run, past go test's 10-minute default, and timed out without it.

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance criteria

- [ ] `atan2 $a, $b` parenless keeps both arguments (`go test ./internal/parse/ -run '^TestParenlessBuiltinArity$' -count=1 -v | grep -q '^--- PASS: TestParenlessBuiltinArity'`)
- [ ] The canon of a parenless `atan2` holds no empty call (`go test ./internal/parse/ -run '^TestParenlessArityNoDetachedCall$' -count=1 -v | grep -q '^--- PASS: TestParenlessArityNoDetachedCall'`)
- [ ] The parse arity tables are audited against perl's keyword list, with the result recorded (`grep -q 'all 266 names in regen/keywords.pl' internal/parse/declarations/CORE.tpm`)
- [ ] REGRESSION GUARD, passes today: `index`, `sprintf` and `crypt` parenless still take both arguments (`go test ./internal/parse/ -run '^TestParenlessBuiltinArity$' -count=1 -v | grep -q '^--- PASS: TestParenlessBuiltinArity'`)
