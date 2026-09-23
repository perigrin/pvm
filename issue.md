---
title: 03_call_forms.t documents a claim its own source cannot observe
state: done
urgency: normal
milestone: corpus-complete
created: 2026-09-23T02:07:02.721145536Z
updated: 2026-09-23T19:35:48.226631233Z
sessions:
- start_sha: 66e6d10dfed49b493965299ae1c5d6fcc9088c25
  end_sha: 66e6d10dfed49b493965299ae1c5d6fcc9088c25
  commits: 0
  started_at: 2026-09-23T19:35:42.005747328Z
  ended_at: 2026-09-23T19:35:48.226631233Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T19:35:42.005747328Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-23T19:35:48.226631233Z
---

`07_subroutines/03_call_forms.t` documents a distinction and then builds
a program that cannot observe it.

Its header states, verbatim:

> `&f` with no parens passes the CALLER's `@_` through untouched rather
> than an empty list, so it is not merely a noisier `f()`.

Its source is:

    sub answer { 42 }
    print answer(), "\n";
    print answer, "\n";
    print &answer, "\n";
    print &answer(), "\n";

`sub answer { 42 }` ignores its arguments. All four spellings print `42`,
and the file's own comment says this "keeps the printed output the same
across all four" -- which is exactly the problem. The file names the
semantic difference and pins an output that is identical whether or not
a parser implements it.

## The distinction is one line to measure

    $ perl -e 'sub inner { "inner[@_]" } sub outer { return &inner }
               print outer(1,2), "\n"'
    inner[1 2]

    $ perl -e 'sub inner { "inner[@_]" } sub outer2 { return inner() }
               print outer2(1,2), "\n"'
    inner[]

`&inner` forwards the caller's `@_`; `inner()` does not. Output separates
them cleanly.

## Why this is its own issue

It adds no files. It repairs a claim the corpus already makes and cannot
support -- the same failure as a gate that passes on incidental data,
which this milestone's measurement found twice elsewhere. A corpus whose
files document more than they assert is worse than one that documents
less, because the documentation is what a reader trusts.

The fix is to give the callee a body that reports `@_`, and to pin an
output that differs between the two spellings. The existing four-spelling
structure survives; only the callee and the expectation change.

`&f` with an EXPLICIT empty list (`&f()`) is the third reading and
belongs in the same file: it passes nothing, unlike bare `&f`.

## Acceptance criteria

- [ ] The call-forms file's callee reports its arguments rather than ignoring them (`awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/07_subroutines/03_call_forms.t | grep -q '@_'`)
- [ ] Its pinned output differs between the &f and f() spellings, so the claim is observable (`go test ./internal/conformance/ -run '^TestCorpus/07_subroutines' -count=1 -v | grep -q '^    --- PASS: TestCorpus/07_subroutines'`)
- [ ] The file is validated against pinned perl 5.42.0 (`go test ./internal/conformance/ -run '^TestTierSubroutinesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesPerlValidated'`)
- [ ] The dependency lint passes over the tier (`go test ./internal/conformance/ -run '^TestTierSubroutinesLint$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesLint'`)
- [ ] The whole conformance suite is green with the change in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
