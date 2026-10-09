---
title: "A call that fills a \\ or * prototype slot reports perl's reference site, lexical subs included"
state: done
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0a8cd-8fcc-72a8-9ba1-c916eee9992d
created: 2026-10-09T05:00:08.759092832Z
updated: 2026-10-09T07:04:44.709384876Z
sessions:
- start_sha: 563f5ab2119ee9c4fb919fc3dcd8e1c570d2c746
  end_sha: 9509016a049ed19754c10f87dace84e5d72e58b5
  commits: 3
  started_at: 2026-10-09T06:08:05.228114687Z
  ended_at: 2026-10-09T07:04:44.709384876Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T06:08:05.228114687Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T07:04:44.709384876Z
observed_paths:
- docs/specs/perl-parser/00-findings.md
- internal/parse/canon.go
- internal/parse/fuzz_test.go
- internal/parse/parse.go
- internal/parse/protoslot_test.go
- internal/parse/refscalar.go
- internal/parse/sites.go
- internal/parse/term.go
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/unaryplus_test.go
- internal/parseoracle/compare_facts.go
- internal/parseoracle/ratchet_test.go
- internal/parseoracle/testdata/ratchet/baseline.txt
---

## Context

- commands: `make test` (green before every commit; never `--no-verify`); the oracle's corpus ratchet `go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -count=1 -v -parseoracle.corpus` (~2 min; regenerate with `-parseoracle.update` in the same commit when verdicts move, listing every moved file); the parse ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2HoldsAll56)$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; corpus: perl.git t/ through the shim (see `internal/parseoracle/ratchet_test.go`); the oracle's markers and verdicts: `internal/parseoracle/compare_facts.go`, docs/specs/perl-parser/00-findings.md §0.11
- the subject's sites: `internal/parse/sites.go` (`decidedMarkers`, `hidesStructure`), `internal/parse/cmd/subject`
- NEVER use pkill or killall

Found by the WRONG triage of 2026-10-09 at 98d02849 (139 WRONG findings across the 27 WRONG files of the regenerated baseline, each cause confirmed by a minimal repro run through the oracle). Goal: WRONG = 0 measured against internal/parse (01a0a8cd).

**P1/P2.** perl takes a reference (srefgen) when a call fills a `\` or `*` prototype slot; `sites.go` never emits a site for any call, even when the subject's prototype table has the prototype. Repro: `sub f(\@){1} my @a; f(@a);` -> perl srefgen, subject nothing -> WRONG. `*`: `sub star(*&){} star *FOO, sub{1}` -> perl `\*FOO`. Lexical subs (`my sub p (\@)`, `state sub p (\@)`) are not in the prototype table at all. Files: comp/proto.t (x23), io/scalar.t:359,364,369,373, op/coreamp.t:673, op/each.t (x10), op/hash.t:141, op/hash-clear-placeholders.t:16, op/sub_lval.t:560, op/lexsub.t:369,370,760,761. The comment in compare_facts.go about "our grammar marks every unqualified f(...) call unresolved" describes the old tree-sitter subject; correct it.

## Steps

- [ ] Write failing test `TestPrototypeSlotReportsReference`: a call through a `\@`, `\%`, `\$`, `\[...]` or `*` slot reports perl's site; a lexical sub's prototype counts; a call with no prototype reports none (negative)
- [ ] Implement in sites.go; re-baseline the oracle, listing every moved file; `make test`; commit

## Acceptance criteria

- [ ] a prototype slot reports the reference perl takes (`go test ./internal/parse/ -run '^TestPrototypeSlotReportsReference$' -count=1 -v | grep -q '^--- PASS: TestPrototypeSlotReportsReference'`)
