---
title: A lexical hash is not an rv2hv, and the subject cannot tell which it is
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T18:15:49.871374091Z
updated: 2026-10-09T17:24:07.687535869Z
sessions:
- start_sha: 25483d2b5bfff159cfc023dec9b050ce0c57c4b3
  end_sha: cc4a3d3a4dff59d2e6da66443724d83ccbfb75ec
  commits: 1
  started_at: 2026-09-20T01:38:27.624702622Z
  ended_at: 2026-09-20T02:04:04.677782914Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-20T01:38:27.624702622Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-20T02:04:04.677782914Z
observed_paths:
- internal/parse/hashderef_test.go
- internal/parse/hashhedge_test.go
- internal/parse/sites.go
- internal/parse/sites_decided_test.go
---

**The title is wrong and git-zhi cannot retitle.** Read it as "the hash
marker over-reports what perl folds away; hedge the undecidable". The
lexical-versus-package framing was my filing hypothesis and the measurement
below refutes it.

48 WRONG rv2hv sites over the 620-file corpus at 25483d2b. This issue was
filed as a lexical-versus-package scope problem. **It is not**, and the
measurement below is what the work should be planned against rather than the
original diagnosis.

## What the 48 actually are

Every WRONG rv2hv site, dumped with its source line and counted by shape:

    hash SLICE  @h{...}, @$r{...}, ->@{...}    ~20
    delete, usually `delete local`              14
    nested deref %{$h{a}{b}}                     ~8
    a plain package %h                           ~2

Barely any are the `keys(%h)` case this issue was filed for. Examples:

    op/array.t:73       @bar{keys %foo} = values %foo;
    run/locale.t:44     delete local @ENV{'LANGUAGE', 'LANG', ...};
    porting/libperl.t:472  for my $symbol (sort keys %{$symbols{data}{$dtype}})
    op/postfixderef.t:104  $spring2{"foo"}->@* = (1,2,3);

## Scope matters, but not the way this issue assumed

It is not "lexical declarations are invisible to us". Measured on 5.42.0:

    my %h;    keys %h                       0 rv2hv   (padhv)
    our %h;   keys %h                       1 rv2hv
    state %h; keys %h                       1 rv2hv   <-- NOT lexical here
    local %h; keys %h                       2 rv2hv

    my %foo, %bar;  @bar{keys %foo} = ...   0 rv2hv
    our %h, %g;     @g{keys %h}    = ...    2 rv2hv

So `state` behaves like a package hash for this marker even though it is a
lexical declaration, and a slice's op count depends on the scope of the
hashes it slices. A scope stack keyed on `my` -- which is what I proposed --
gets `state` wrong and does nothing for the slice cases where both hashes
are already lexical and we over-report anyway.

## And some are not about scope at all

`@$a{'abc','def'}` emits 0 rv2hv on its own:

    my $r = {a=>1}; my @x = @$r{'a','b'};   0 rv2hv

but op/avhv.t:131 has the same shape inside an `eval { }` block and IS in
the WRONG list, which makes it an op-attribution question rather than a
detection one.

## Why hedging is the plan

`isHashTerm` tests a leaf's TEXT for a leading `%`. Text cannot carry scope,
cannot know which accesses perl's optimiser folds to multideref, and cannot
see the block structure that moves attribution. The marker's own doc records
the asymmetry that makes hedging safe:

> a lexical `%h` is padhv and `$h{a}` folds to multideref, so perl emits
> FEWER of these than the source has hash accesses, never more -- which is
> the direction presence-matching tolerates

A hedge scores WIDER, which is the honest bucket for "I see a hash access
and cannot tell whether perl will emit a dereference for it". It is also the
bucket the four-bucket design exists to reward, and exports.go's three-tier
rule is the existing precedent for refusing rather than guessing.

Tracking scope is NOT deferred, it is **dropped**: the measurement says it
would fix at most the ~2 plain-package sites while getting `state` wrong,
and the other 46 need facts a parser cannot have.

## Scope of the change

Hedge the hash site where the reading is undecidable, commit where it is
not. The line to draw needs its own measurement -- hedging EVERY hash site
would take exact down as well as WRONG, and the exact floor exists to stop
exactly that. Start by hedging the four shapes above and measure both
numbers.

### Criteria re-pointed (2026-10-09)

These criteria sat under a heading `git zhi verify` does not read, so they
never ran; run now, some named tests that no longer exist or need a shim:
TestGoParserWrongIsZero and TestGoParserExactFloor run only with
PARSEORACLE_SHIM (the oracle measures internal/parse by default since
01a0b835), TestUnknownScoresWider was deleted in 3b9422eb (TestUnknownHedgesEveryMarker
asserts the rule directly), and the T1-easy tests went with easy_test.go
in b4180102 (the T1 ratchet is the gate). Each now names the gate that
holds the same claim today: TestCorpusBaselineIsIntact asserts zero WRONG;
an exact floor is checked against the committed baseline's rows; the T1
rate against the T1 ratchet's own header; TestSitesSpanMultilineStatement
lives in internal/parseoracle.

## Acceptance criteria

- [x] An undecidable hash access hedges rather than committing (`go test ./internal/parse/ -run '^TestUndecidableHashHedges$' -count=1 -v | grep -q '^--- PASS: TestUndecidableHashHedges'`)
- [x] A decidable one still commits, so exact does not collapse (`go test ./internal/parse/ -run '^TestDecidableHashStillCommits$' -count=1 -v | grep -q '^--- PASS: TestDecidableHashStillCommits'`)
- [x] WRONG rv2hv falls and the exact floor holds, both measured (`awk '$1=="exact"{e++} $1~/^(exact|wider|WRONG|no-answer)$/{m++} END{exit !(m>0 && e*1000 >= m*589)}' internal/parseoracle/testdata/ratchet/baseline.txt`)

Measured 2026-09-19. Plan: M1. Spec: §7.1.2.


## Done, 2026-09-20 (cc4a3d3a)

    corpus WRONG      38 files / 96 sites -> 25 / 59
    corpus rv2hv      48 -> 10
    corpus exact      405 -> 390
    corpus wider      134 -> 162
    T2 exact floor    73.2%, floor 40.0%, holds
    full suite green, 0 FAIL across the repo

## The exact cost was real, predicted, and taken deliberately

sites_decided_test.go:425 warned that hedging subscripts would "trade exact
for wider on every hash and array access in the corpus". That is exactly the
-15, and perigrin took the trade with the numbers in hand.

The reasoning: WRONG is the GATE and exact is a RATCHET. The 15 are not lost
knowledge -- they are claims that could not be justified, right only when a
hash happened to be lexical.

## The line, drawn from measurement

A DEREFERENCE always emits the op and is decided; a NAMED hash depends on a
declaration and is hedged.

    keys %$r   1    keys %{$s{a}}     1    $r->%*  1
    keys %{$r} 1    keys %{$s{a}{b}}  2    %ENV    1

A nested access is both at once: `%{$s{a}}` is a dereference that always
emits, of an element that folds to multideref when the base is lexical. One
decided site plus one hedge, and my first test asserted no hedge there and
was wrong.

## Two under-reports fixed on the way

`@$r{"a","b"}` -- a slice through a reference -- reported no site at all,
the same class as `%$href` in 461ea788.

And my first version of that rule claimed `@{$r->{x}}`, an ARRAY dereference
of a hash element: rv2av, a different op. The Arrow flag separates them,
since a real slice has no arrow. That cost 2 sites back (8 -> 10) and the
2 were correct to give up.

## Four existing tests changed

Three were genuine inversions, not incidental breakage:

    TestSitesDecideMarkers       hash rows are dereferences now
    TestPlainDerefDoesNotHedge   hash subscript row gone, ARRAY row stays;
                                 that test owns OPACITY, this rule owns
                                 undecidable scope, and the comment now says so
    TestSitesReportTheirOwnLine  example swapped to `keys %$r`
    TestHashDerefIsAHashSite     asserts the site is PRESENT, decided or
                                 hedged -- silence is still scored as the
                                 claim that nothing is there

## What remains: 10 sites, all M4

op/each.t and op/hash-clear-placeholders.t are XS prototype calls --
`Hash::Util::num_buckets` has prototype `(\%)` and
`Internals::hv_clear_placeholders` `(\%)`. Not decidable from Perl source at
all, since the prototype lives in compiled C.
