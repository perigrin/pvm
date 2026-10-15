---
title: "The lowering pass: internal/infer must read internal/parse, not tree-sitter"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c115-df67-7d69-8883-73d40b0ae15d
blocks:
- 019d1449-1880-729b-860b-ac29464f6cd7
created: 2026-09-20T22:19:52.755941663Z
updated: 2026-10-15T15:18:14.377243048Z
sessions:
- start_sha: 8d07b21bc858f3646cad58b8e4a7538e38eaedc3
  end_sha: ""
  commits: 0
  started_at: 2026-10-15T15:07:56.788229224Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-15T15:07:56.788229224Z
---

`internal/infer` reads the tree-sitter tree through `internal/parser/`. That
tree is being burned down, so every inference rule written against it is work
that gets deleted. Nothing in `internal/infer` should be extended until it
reads `internal/parse/` instead.

This issue is the milestone's closing gate. The work is decomposed into the
issues it is blocked by; this one checks, across all of them, that the goal
holds. It stays last, and it is the only edge from this milestone into
psc-type-alignment's 019d1449.

## Why no partial work survives

§6.1.6 counted it, and the count is the argument:

> `internal/infer/*.go` switches on 46 distinct tree-sitter-perl grammar
> names (`subroutine_declaration_statement`,
> `ambiguous_function_call_expression`, `array_element_expression`,
> `hash_deref_expression`, `func0op_call_expression`, `refgen_expression`,
> ...), and `internal/parse/parse.go:146-201` emits 23 coarse kinds with none
> of those names (`statement`, `term`, `binary`, `index`, `declaration`,
> `call`, ...). Porting the facade hands PSC a `Kind()` that returns
> `"binary"` where it expects `"assignment_expression"`. **Ten methods port;
> the vocabulary does not.**

So this is not a migration that can be done incrementally behind a shim. The
vocabulary changes underneath every switch.

## And the old tree is re-derived from TEXT in places

§4.14.2 records where `internal/infer` reaches past the tree entirely:

    findOperatorText    infer.go:656
    callElementType     infer.go:1307, telling map from grep by
                        strings.HasPrefix on source (:1315)

Those are the places new rules would have to reproduce, and they are exactly
what a lowered tree exists to remove.

## Scope

The lowering pass of §6.1.6: a pass over `internal/parse`'s tree producing
the vocabulary §4.14 specifies (and chapter 5's statement kinds), which
`internal/infer` and `internal/psc` then consume instead of
`internal/parser`.

§6.1.6 is explicit about what NOT to build with it: "One lowering pass,
specified in §6.1.6, and no dataflow IR until a feature demands one." Neither
HIR nor PIR as perl-lsp built them.

Key facts by NODE IDENTITY, not byte offset. §6.1.6 item 4, and rust-analyzer
rejected the offset key by name: "offsets change after edits. So, as a rule
of thumb, we avoid using offsets, text ranges or syntax trees as keys and
values for queries". §6.4.3 observes the same defect in PSC's existing
`map[uint32]types.Type`.

## Size

About 5,200 non-test lines in `internal/infer` at 24bfac4a (`infer.go`
4,459, `symbols.go` 451, `project.go` 135, `diagnostics.go` 161), and 246
tests. Decomposed: a status re-measure, node identity, the vocabulary, seven
lowering slices, the position map, seven inference ports, the test-suite
move, four caller flips, and the retirement.

## What it unblocks

  019d1449  Expand PSC inference       -- blocked on this, now explicitly
  area 5 of that issue                 -- NarrowByContext has no source of
                                          context until the lowered tree
                                          carries Wantarray and call-site
                                          context

Plan: M2/M3 adjacent -- it is what makes §6.1.6's "PSC is being converted to
the new tree regardless" true. Spec: §6.1.6, §4.14, §6.4.3.

## Context

- paths: `internal/infer/`, `internal/psc/`, `internal/lower/`, `internal/infer/testdata/lowered/`
- commands: `git zhi verify --timeout 900 m2-lower-the-tree`
- entrypoints: the goal, perigrin 2026-10-11: (3) every m2 issue is done or cancelled; (5) `internal/infer` and `internal/psc` read `internal/parse` through `internal/lower`, and both the batch checker and the LSP run on the lowered tree; (6) PSC does not get worse: precision exact at least 62 of 76, and any changed psc diagnostic is explained in the commit that changes it; (7) the m1 parse gates hold; (8) `make test` and CI pass.

The criteria below are the goal's cross-cutting checks. Each issue in the
chain has its own; these are the ones no single issue can own.

The m1 gate that is not here is the oracle's exact-match ratchet over the
620-file corpus (`TestRatchetCorpus`), which skips without
`-parseoracle.corpus` and runs only in CI's parse-fidelity workflow. This
milestone changes no non-test file under `internal/parse`, `internal/lexer`
or `internal/parseoracle`, so CI passing (goal 8) is its check.

## Steps

- [ ] Confirm every other m2 issue is done or cancelled, each cancellation
      with its reason recorded.
- [ ] Confirm `git status` is clean, then run
      `git zhi verify --timeout 900 m2-lower-the-tree` and confirm every
      done issue's criteria ran: compare each issue's `- [ ]` lines against
      the verify log by command, since a malformed criterion is dropped
      silently.
- [ ] Confirm CI is green on the final commit.
- [ ] Run the criteria below. Mark this issue done.

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

**Criterion narrowed 2026-10-15:** the golden-trailer check counted any commit touching the testdata directories, including `matching.txt` (the ratchet's own list, which every port updates). It now counts commits that modify a `.golden`, which is what its title claims; all 16 such commits carry the trailer.

## Acceptance Criteria

### Positive Scenarios
- [ ] Every other issue in the milestone is done or cancelled (`git zhi issue list --milestone m2-lower-the-tree --all --format json | python3 -c "import json, sys; d = json.load(sys.stdin); left = [i['id'] for i in d if i['state'] not in ('done', 'cancelled') and not i['id'].startswith('01a0c0e7')]; print(left); sys.exit(1 if left or len(d) < 16 else 0)"`)
- [ ] Both internal/infer and internal/psc import internal/lower (`test "$(go list -f '{{join .Imports " "}}' ./internal/infer ./internal/psc | grep -c 'pvm/internal/lower')" -eq 2`)
- [ ] The batch checker runs on the lowered tree and reports the golden diagnostics (`go test ./internal/psc/ -run '^TestCheckCommandOnLoweredTree$' -count=1 -v | grep -q '^--- PASS: TestCheckCommandOnLoweredTree'`)
- [ ] The LSP runs on the lowered tree and reports what the batch checker reports (`go test ./internal/psc/ -run '^TestLSPDiagnosticsMatchGolden$' -count=1 -v | grep -q '^--- PASS: TestLSPDiagnosticsMatchGolden'`)
- [ ] Precision on the lowered path is at least 62 of 76 exact with WRONG at zero (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | awk '/^--- PASS: TestInferencePrecision/ {p = 1} /^  exact:/ {n = $2} END {exit !(p && n >= 62)}'`)
- [ ] Every golden fixture holds on the lowered path (`go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$' -count=1 -v | grep -q '^--- PASS: TestLoweredGoldenRatchet'`)
- [ ] T2 still parses clean (`go test ./internal/parse/ -run '^TestT2HoldsAll56$' -count=1 -v | grep -q '^--- PASS: TestT2HoldsAll56'`)
- [ ] The T1 parse ratchet holds (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] The perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] Canon is still a fixpoint over T1 (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] The canon fidelity ratchet holds (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] The corpus ratchet holds every verdict (`go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1 -v | grep -q '^--- PASS: TestCorpusRatchet'`)
- [ ] The oracle's WRONG stays zero against an intact baseline (`go test ./internal/parseoracle/ -run '^TestCorpusBaselineIsIntact$' -count=1 -v | grep -q '^--- PASS: TestCorpusBaselineIsIntact'`)
- [ ] The oracle's fixture ratchet holds (`go test ./internal/parseoracle/ -run '^TestRatchet$' -count=1 -v | grep -q '^--- PASS: TestRatchet'`)

### Negative Scenarios
- [ ] Neither internal/infer nor internal/psc reaches the tree-sitter path, directly or through a dependency (`! go list -deps ./internal/infer ./internal/psc | grep -q 'pvm/internal/parser$'`)
- [ ] No golden was changed without an explanation: every commit that modifies one carries a Diagnostics-changed trailer (`test -n "$(git log --format=%H --diff-filter=A -- internal/infer/testdata/lowered)" && for c in $(git log --format=%H --diff-filter=M -- 'internal/infer/testdata/lowered/*/*.golden' 'internal/psc/testdata/*.golden'); do git log -1 --format=%B "$c" | grep -q '^Diagnostics-changed:' || exit 1; done`)
