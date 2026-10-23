---
title: lintOps reads Concise's rendered text, so the op budget cannot see a flags-carried distinction
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-26T14:14:49.420353566Z
updated: 2026-10-23T14:45:39.887770686Z
sessions:
- start_sha: 054e75e6f0548ae9b596ac0d16dadd06e0d100a9
  end_sha: ""
  commits: 0
  started_at: 2026-10-23T14:45:39.887770686Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-23T14:45:39.887770686Z
---

`lintOps` reads `B::Concise`'s RENDERED TEXT, so the op budget cannot see any
distinction Concise does not print. Two programs that behave differently can be
one op set to it.

Reported by the B::SoN session, which hit it from the other direction, and
verified here on 5.42.0.

## Measured

    perl -MO=Concise,-exec -e 'my @a = split / /, "  a b ";'
    perl -MO=Concise,-exec -e 'my @a = split " ", "  a b ";'

Both print, byte for byte:

    5  </> split(/" "/ => @a:1,2)[t2] vK/LVINTRO,ASSIGN,LEX,IMPLIM

And they are different operations:

    split / /, "  a b "     4 fields: ["", "", "a", "b"]
    split " ", "  a b "     2 fields: ["a", "b"]

awk mode strips LEADING whitespace and splits on RUNS. It is not `/ /` and not
`/\s+/` either -- only the leading strip makes it awk, so no pattern spelling
reproduces it. The separator is `PMf_SKIPWHITE` (2048) in `pmflags`, which
Concise does not render:

    split / /, ...          pmflags=0
    split " ", ...          pmflags=2048
    my $p=" "; split($p)    pmflags=2048

Note the third: perl sets the bit whenever the pattern is a plain string
EXPRESSION rather than a `//` literal, so it is a compile-time fact even for a
runtime pattern.

## What this means for the corpus

`internal/conformance/lint.go:242`'s `fileOpsBackend` ends with

    B::Concise::compile('-exec', '-main', @names)->();

and `opsOf` parses that output. For op NAMES this is sufficient -- Concise
always prints the name, and the op budget is a set of names. So no existing
lint result is wrong.

What it cannot do is distinguish two cases whose difference lives in a flag.
Concretely: a corpus case asserting awk-mode `split` and one asserting
pattern-mode `split` would satisfy the same budget, and a tier introducing
`split` introduces both readings whether or not anyone measured that it did.

That is a limit on the INSTRUMENT rather than a defect in a result, which is
why this is filed rather than fixed under something else. The honest form of
every "measured, the op sets are identical" claim in this repo is "identical in
what Concise prints".

## Scope

1. Decide whether the op budget should see flags at all. It may not need to:
   the budget's job is "which constructs does this tier introduce", and a flag
   is arguably a property of a construct the tier already has. Write the
   decision down either way -- that is the deliverable even if the answer is
   no change.
2. If it should: the backend already has `B` loaded, so reading
   `$op->pmflags` for a PMOP is available without a new dependency. Measure
   which flags would actually discriminate corpus cases before adding any --
   `PMf_SKIPWHITE` is one, and `PMf_CONST`/`PMf_GLOBAL` may be others.
3. Either way, record in `lint.go`'s docstring that the source is Concise's
   rendering and what that excludes. A reader currently has no way to know the
   instrument has a blind direction.

## Context

- paths: `internal/conformance/lint.go` (`fileOpsBackend` at 242, `opsOf` at
  283, `lintOps` at 143)
- commands: `go test ./internal/conformance/ -count=1`
- measure with: `perl -MO=Concise,-exec` versus a backend reading `pmflags`
  directly

## Acceptance criteria

- [ ] The decision on whether the budget sees flags is recorded in `lint.go` (`grep -qiE 'pmflags|Concise.*render|flags the budget' internal/conformance/lint.go`)
- [ ] REGRESSION GUARD, passes today: the corpus lint holds (`go test ./internal/conformance/ -run TestCorpusLints -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
