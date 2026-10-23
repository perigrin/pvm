---
title: The corpus asserts local restores on block exit, never on call return
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-26T17:52:33.001336328Z
updated: 2026-10-23T23:40:18.605172108Z
sessions:
- start_sha: 62cf900f71db986393d6d645b6814afefaea26e4
  end_sha: 58b9d93efe068e8222e292e76272380827c4d10f
  commits: 1
  started_at: 2026-10-23T15:16:23.576324386Z
  ended_at: 2026-10-23T15:40:18.568412039Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-23T15:16:23.576324386Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-23T15:40:18.568412039Z
observed_paths:
- conformance/mdtest/dynamic-scope.md
- conformance/mdtest/subroutines.md
- conformance/mdtest/symbol-table.md
- internal/conformance/ratchet_test.go
- internal/conformance/testdata/corpus.ratchet
---

The corpus asserts `local` restores on BLOCK EXIT and never asserts it restores
on CALL RETURN. The construct is covered; the read position is not. Contributed
by the B::SoN session, where the missing shape prints a plausible wrong answer.

## The gap, measured over perl BLOCK contents

Four `local` sites in the whole corpus, and none reads the localised value
through a call:

    adjacency-05_scoping.md    local $g = $g + 10;              no sub in block
    dynamic-scope.md           local $x = 2;                    no sub in block
    adjacency-07_subroutines   local $SIG{__WARN__} = sub { };  sub is the VALUE
    argument-extent.md         local $SIG{__WARN__} = sub { };  sub is the VALUE

The two `$SIG` cases have a sub in the block, but it is the value being
assigned -- nothing calls it to read `$SIG`, so the localised value is never
observed from inside a call.

`dynamic-scope.md`'s case is titled "`local` restores on block exit". Block
exit, not call return. That is the gap in one line, and it is the same axis as
`read-position.md`: a construct already represented, with the POSITION of the
read being what was missing.

## What perl does, verified on 5.42.0

    our $v = "outer";
    sub show  { print "$v\n" }
    sub inner { local $v = "inner"; show() }
    inner();
    show();

      inner
      outer

And identically for the glob spelling, `local *v = \"inner"`. Both print
`inner` then `outer`.

## Why it is worth a case even though our parser reads it

Measured: all of `local $v`, `local *v = \"x"` and the cross-call shape parse
here with Unknown=0. So this is not a parse gap for us -- it is a corpus gap,
and the corpus's job is to assert what PERL PRINTS so an implementation that
parses it and then lowers it wrongly is caught.

The B::SoN session's measurement is the argument. It prints `inner inner` for
perl's `inner outer` -- `local` dropped entirely when the localised value is
only read through a call, emitting a bare assignment with no save and no
restore, so the outer value is clobbered permanently. Compiles, exits 0, wrong
answer.

And the same-scope shape PASSES there, which is why neither of us caught it:
the walk sees pre- and post-scope values as two bindings and emits neither save
nor restore, which happens to print the right thing. Across a call there is no
such binding.

## Tier: 07, not 05, and that is the structural reason for the gap

Ops for the cross-call form, from `-MO=Concise,-exec`:

    const enter entersub gv gvsv leave nextstate pushmark sassign

`entersub` is tier 07's. `dynamic-scope.md` is tier 05 and cannot reach it, so
the case that would have caught this could not have been written in the topic
that owns `local`. The construct is tier 05's and the READ POSITION needs tier
07 -- which is why a per-construct corpus missed it.

`local`'s own identity is a FLAG, not an op, and needs no tier work:

    *v = \@A          rv2gv sKRM*/1
    local *v = \"x"   rv2gv sKRM*/LVINTRO,1
    local $v = "x"    gvsv[*v] s/LVINTRO

Identical ops, one bit apart -- verified. Same shape as `continue`
(`01a0ded4`), whose destination is metadata on `enterloop`, and the opposite of
`given`/`when` (`01a0dec7`), which needs five ops claimed before a case can
exist. And it is invisible to `lintOps` for the reason `01a0de11` records: the
lint reads Concise's rendered text, so it sees op NAMES and not flags.

## Scope

1. A case in `adjacency-07_subroutines.md` or a new tier-07 topic: `local`
   set in one sub, read in another, asserting `inner` then `outer`.
2. Both spellings. `local $v` and `local *v = \"inner"` are the same property
   through different slots, and the glob form is what perl's own `t/` uses.
3. Consider a second case where the localised value is read through TWO frames,
   since one frame may be special-cased by an implementation that tracks the
   immediate caller.
4. Leave `dynamic-scope.md` alone. Its case is correct and its title is
   accurate about what it asserts; add a pointer rather than rewriting it.

## Context

- paths: `conformance/mdtest/dynamic-scope.md` (tier 05, the block-exit case),
  `conformance/mdtest/adjacency-07_subroutines.md` (tier 07, where `entersub`
  is reachable), `conformance/mdtest/read-position.md` (the same axis)
- commands: `go test ./internal/conformance/ -count=1`
- related: `01a0ded4` (`continue`, identity in metadata), `01a0dec7`
  (given/when, five unclaimed ops), `01a0de11` (lintOps cannot see flags)

Criteria updated 2026-10-23 (review round 1, agent:ccx-rev1): criterion 1 passed at 054e75e6, before the work. It now names the call-return case in subroutines.md.

## Acceptance criteria

- [ ] A case asserts local restores on CALL RETURN, not just block exit (`go test ./internal/conformance/ -run '^TestCorpus$/^subroutines.md$/^.local._restores_on_call_return' -count=1 -v | grep -q '^    --- PASS: TestCorpus/subroutines.md/.local._restores_on_call_return'`)
- [ ] Both the scalar and glob spellings are covered (`grep -qE 'local \*[a-z]' conformance/mdtest/*.md`)
- [ ] REGRESSION GUARD, passes today: the corpus stays complete (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
