---
title: The lint cannot see inside a sub, so five tier-07 ops and methstart are unclaimable
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T18:43:00.587037252Z
updated: 2026-09-22T00:45:18.402747232Z
sessions:
- start_sha: b41801021927f68fcbf799c60d6b947393d983de
  end_sha: a3a0ed0c97ce48a68f5db427979f9d8acf88d3c8
  commits: 1
  started_at: 2026-09-21T22:29:10.573159397Z
  ended_at: 2026-09-21T23:31:52.682542829Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T22:29:10.573159397Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T23:31:52.682542829Z
observed_paths:
- conformance/01_literals/README.md
- conformance/02_variables/README.md
- conformance/03_context/README.md
- conformance/04_operators/README.md
- conformance/05_scoping/README.md
- conformance/06_control/README.md
- conformance/07_subroutines/08_parenless_extent.t
- conformance/07_subroutines/10_undeclared_callee.t
- conformance/07_subroutines/README.md
- conformance/08_references/README.md
- conformance/09_regex/README.md
- conformance/10_io/README.md
- conformance/11_oo/README.md
- conformance/12_packages/README.md
- conformance/13_opaque/README.md
- conformance/14_recursive/README.md
- internal/conformance/lint.go
- internal/conformance/lint_test.go
- internal/conformance/numbering_test.go
- internal/conformance/prereq_test.go
- internal/conformance/subops_test.go
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier01_test.go
- internal/conformance/tier02_test.go
- internal/conformance/tier03_test.go
- internal/conformance/tier04_test.go
- internal/conformance/tier05_test.go
- internal/conformance/tier06_test.go
- internal/conformance/tier07_callforms_test.go
- internal/conformance/tier07_test.go
- internal/conformance/tier08_test.go
- internal/conformance/tier09_test.go
- internal/conformance/tier10_test.go
- internal/conformance/tier11_test.go
- internal/conformance/tier12_test.go
- internal/conformance/tier13_test.go
- internal/conformance/tier14_test.go
---

`opsOf` in `internal/conformance/lint.go:197` runs

    exec.Command("perl", "-MO=Concise,-exec", path)

with no sub named, so B::Concise prints the MAIN program's optree only. A
subroutine body is a separate CV and needs `-MO=Concise,-exec,<subname>`.

Two tiers hit this independently while writing their corpus files at
0ce515cb:

- **Tier 07** dropped from seven claimed ops to two. `argcheck`,
  `argdefelem`, `argelem`, `leavesub` and `return` are all really emitted
  by `conformance/07_subroutines/*.t` and all live in a CV. Its adjacency
  file is the sharpest case: source carrying a signature, a parameter
  default, a `for` loop, a `next`, three `return`s and an ampersand call
  compiles a main optree containing exactly `anoncode` and `entersub`.
- **Tier 11** dropped `methstart`, which opens every `method` body under
  the `class` feature. No file can be written that puts it where the lint
  looks.

Both READMEs record the real op streams in prose so the measurements are
not lost, and both say why the op is unclaimed.

The fix is to teach `opsOf` to enumerate the file's subs and dump each.
It is deliberately NOT done here: it changes the instrument for all
fourteen tiers at once, and several tiers' declared sets were measured
against the current behaviour. It wants measuring before being assumed an
improvement -- specifically, whether widening the visible op set causes
any tier to start emitting an op claimed by a LATER tier, which is the
direction the lint exists to catch.

## The measurement, taken

The widening was built and measured across all fourteen tiers before
being landed, which is what the paragraph above asked for. The answer to
the question it names -- does a wider visible op set make any tier emit
an op a LATER tier claims -- is NO, in every file in the corpus. Every
one of the twenty-one files whose op stream changed reports `uses ops no
tier claims`, and not one reports `uses ops introduced after`. The
widening therefore does not disturb the ordering; it only exposes ops
that had no owner because no tier could see them to claim them.

Eight ops become visible corpus-wide:

    argcheck argdefelem argelem leavesub methstart postinc return shift

`leavesub` is the broad one, reaching tiers 07, 08, 11 and 12 -- every
tier whose files declare a sub. The other seven are confined: the four
signature and `return` ops to tier 07, `methstart` to tier 11, `postinc`
to tier 07's `07_args_alias.t`, `shift` to tier 11's `06_indirect_new.t`.

No tier's INTRODUCES set becomes incomplete. Every op the fourteen
READMEs claim is still emitted by some file under the wider instrument,
so nothing had to be removed.

`postinc` and `shift` are the one finding the issue did not predict.
Both are assigned in PROSE to earlier tiers -- 07's README says
"`shift` is tier 02's" and "`postinc` ... is the increment operator and
not this tier's subject", and 04's README says `shift` is "an op that NO
tier in the corpus claims". But no tier-02 or tier-04 file emits either
op, so claiming them there fails `TestCorpusLints`, which reports a
README claiming an op no file emits. Under the corpus's own rule --
the tier that emits an op FIRST owns it, which `claimonce_test.go`
states -- they belong to the tiers that emit them, and the prose that
assigns them elsewhere describes where they would belong in a corpus
that had a file for them.

## Acceptance criteria

- [x] `opsOf` sees inside a named sub: a file whose signature compiles `argcheck`, `argdefelem` and `argelem` and whose body ends in `leavesub` yields all four, none of which the main optree carries. (`go test ./internal/conformance/ -run TestOpsOfSeesInsideNamedSub -count=1 -v`)
- [x] `opsOf` sees inside a `method` body, so `methstart` is reachable at all. (`go test ./internal/conformance/ -run TestOpsOfSeesInsideMethod -count=1 -v`)
- [x] `opsOf` sees subs declared in a package other than `main`, since B::Concise names them `Foo::bar` and most of tiers 11 and 12 declare subs there. (`go test ./internal/conformance/ -run TestOpsOfSeesSubsInOtherPackages -count=1 -v`)
- [x] The enumeration stops at the file boundary. Every CV the process loaded is reachable from the symbol table, and dumping `strict::import` or `Exporter::import` would add their ops to EVERY corpus file's measured set. A file that merely says `use strict; use warnings` emits no `goto`, `require` or `dofile`. (`go test ./internal/conformance/ -run TestOpsOfStopsAtTheFileBoundary -count=1 -v`)
- [x] The main program is still dumped, which thirteen tiers' op sets were measured against, so the widening ADDS and never replaces. (`go test ./internal/conformance/ -run TestOpsOfStillSeesTheMainProgram -count=1 -v`)
- [x] Each of the five checks above is proven by mutation: reverting `opsOf` to the bare `-MO=Concise,-exec` fails the first three, and dropping the `FILE` filter fails the fourth. (`go test ./internal/conformance/ -run TestOpsOfMutations -count=1 -v`)
- [x] Tier 07's INTRODUCES grows back the ops the issue names, plus `postinc`, and tier 11's gains `methstart` and `shift`, each claimed by the tier that emits it first. (`grep -A3 '^## INTRODUCES' conformance/07_subroutines/README.md conformance/11_oo/README.md`)
- [x] Both READMEs' statements that these ops are INVISIBLE to the lint are rewritten, since the widening makes them false. (`git diff -- conformance/07_subroutines/README.md conformance/11_oo/README.md`)
- [x] No op is claimed by two tiers after the widening. (`go test ./internal/conformance/ -run TestOpClaimedByExactlyOneTier -count=1 -v`)
- [x] Every corpus file still lints at its declared tier, and no README claims an op no file emits. (`go test ./internal/conformance/ -run TestCorpusLints -count=1`)
- [ ] The whole conformance suite stays green. (`go test ./internal/conformance/ -count=1`)

      NOT TICKED, and not by this change. 153 of the package's 154 tests
      pass. The one failure is `TestCorpusRatchet`, reporting that
      `undeclared_callee.t` -- the corpus's first `--- expect parsent`
      file, landed concurrently -- passes on arrival and belongs in the
      baseline before it lands. The ratchet path never calls `opsOf`
      (`grep -n 'opsOf|lintOps|lintFile' internal/conformance/ratchet.go
      internal/conformance/ratchet_test.go` matches nothing), so the
      widening is not what fails it; regenerating the ratchet belongs to
      the change that added the file. Everything else is green:
      (`go test ./internal/conformance/ -count=1 -run 'Test[^C]|TestC[^o]|TestCo[^r]|TestCorpusL'`)
      and (`go test ./internal/conformance/ -count=1 -run 'TestCorpus$|TestCorpusSkipsAreDocumented'`).
