---
title: Commit a perl-judged probe corpus that every parser fix extends
state: in-progress
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a1664d-69f0-715c-b060-c8f648ddfcd3
- 01a16666-9a1b-76b3-9389-e4dbfb01ab9d
blocks:
- 01a16602-5ebe-7169-b857-996ae2953394
- 01a12059-30a8-722e-bbc3-52066b8610cb
- 01a1664d-7bc8-7498-af02-7eaf187b43ed
- 01a1690b-1d10-7a84-be7e-0cbc31f36812
created: 2026-10-23T01:08:37.943092553Z
updated: 2026-10-23T17:31:28.974633733Z
sessions:
- start_sha: 4505f236a5ac6d2c6b78214071acbab509c7bb37
  end_sha: ""
  commits: 0
  started_at: 2026-10-23T17:31:28.974633733Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-23T17:31:28.974633733Z
---

## Context

- paths: `internal/parse/probes_test.go` (new), `internal/parse/testdata/m3probes/` (new), `internal/parse/testdata/m3probes/README.md` (new)
- commands: `go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 -v`
- entrypoints: `docs/postmortems/core-module-types.md` (action 3, ~line 48), `internal/parse/sites.go` (`Sites`)

The core-module-types postmortem's action 3 asks for the regression-probe corpus to be committed as a perl-judged table test that every later fix runs mechanically. A per-fix review that leaves only prose cannot be checked, and in that milestone, reviews that re-checked only the named finding let through regressions that the next lens found. This issue builds the table. Each probe file under `testdata/m3probes/` holds snippets, and each snippet holds facts of two kinds, both checked mechanically (K3, chain-review round 6):
- **Site facts:** marker-site counts per marker (rv2hv, srefgen, anonhash, match, readline), each with psc's expected verdict, `exact` or `wider`. perl's counts come from the pinned perl through parse_facts.pl, and psc's from `Sites`.
- **Re-emission facts:** a `faithful` row. The snippet must compile under the pinned perl's `perl -c` and must parse and be faithful through `Faithful`, which exists today; later issues' changes to `Faithful` apply to these rows automatically. This is what the re-emission issues' probes record.

`Deparse,-p` output may be quoted in a comment for the reader, but it is never compared, since internal/parse has nothing to compare it with. TestM3ProbesHeldToPerl does three things for every snippet: it runs the pinned perl (corpus.pin's interpreter, 5.44.0, found through `perltest.PinnedPerl`, so it skips under the 5.42 jobs and fails under `PVM_REQUIRE_PINNED_PERL=1`) to confirm the site facts and `perl -c`; it checks that psc's facts agree; and it checks each `faithful` row through `Faithful`. A known disagreement is allowed only as a row marked `known-defect <issue>`. Every later parser, sites.go or canon fix in this milestone adds its own probe file, named in that issue's criteria. Tests run perl; production psc never does.

## Steps

- [ ] Create `internal/parse/probes_test.go` with two `// ABOUTME:` lines, and a README in testdata/m3probes/ describing the readable text format (snippet, perl facts, optional `known-defect` row).
- [ ] Each probe row records, per marker, perl's count and psc's expected verdict, `exact` or `wider`. A psc hedge counts as agreement only where the row says `wider` explicitly. A row that says `exact` fails on a hedge, so a hedge cannot stand in silently for a decision, and a probe added by a fix that removes a hedge says `exact` (H7).
- [ ] RED: write TestM3ProbesHeldToPerl. It loads every probe file and fails on an empty directory. It runs perl on each snippet and compares perl's facts with the recorded ones, then compares psc's facts with perl's. Seed one probe file, `baseline.txt`, from adversarial cases the reviews of this milestone's predecessors used (for example, `\(@a)` vs `\@a`, `my ($x) = @_`, and `keys %$r` in each context). Flip one recorded count, watch the test fail, and restore it.
- [ ] RED/GREEN TestM3ProbesCheckFaithfulRows: a fixture `faithful` row whose snippet re-emits unfaithfully, or that `perl -c` rejects, fails.
- [ ] GREEN: implement the fact extraction on both sides by reusing the oracle's perl-side probe (internal/parseoracle/testdata/parse_facts.pl) rather than writing a second one. Run the test and watch it pass.
- [ ] RED/GREEN the negatives: TestM3ProbesRejectEmptyDir (an empty or missing probe directory fails), TestM3ProbesDetectDriftedFact (a fixture probe with one count changed from perl's fails and names the snippet, the marker, and both counts), and TestM3ProbesKnownDefectNamesAnIssue (a `known-defect` row without a git-zhi issue id fails).
- [ ] Add TestM3ProbesHeldToPerl to parse-fidelity.yml's 5.44-held step (`PVM_REQUIRE_PINNED_PERL=1`).
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The seed probe file is committed (`sh -c 'test -s internal/parse/testdata/m3probes/baseline.txt && go test ./internal/parse/ -run "^TestM3ProbesHeldToPerl$" -count=1 -v | grep -q "^--- PASS: TestM3ProbesHeldToPerl"'`)
- [ ] Every probe's recorded facts match perl 5.44.0, and psc agrees except on rows marked known-defect (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 -v | grep -q '^--- PASS: TestM3ProbesHeldToPerl'`)
- [ ] Regression guard (passes before this issue's work; it holds the new tests once they exist): every test that calls PinnedPerl, this issue's included, is in the -run list of a CI step that sets the REQUIRE variable (`go test ./internal/parseoracle/ -run '^TestEveryPinnedPerlTestIsRequired$' -count=1 -v | grep -q '^--- PASS: TestEveryPinnedPerlTestIsRequired'`)

### Negative Scenarios

- [ ] An empty or missing probe directory fails rather than passing with nothing checked (`go test ./internal/parse/ -run '^TestM3ProbesRejectEmptyDir$' -count=1 -v | grep -q '^--- PASS: TestM3ProbesRejectEmptyDir'`)
- [ ] A probe whose recorded count differs from perl's fails and names the snippet, the marker, and both counts (`go test ./internal/parse/ -run '^TestM3ProbesDetectDriftedFact$' -count=1 -v | grep -q '^--- PASS: TestM3ProbesDetectDriftedFact'`)
- [ ] A known-defect row that names no git-zhi issue fails, so no disagreement is parked silently (`go test ./internal/parse/ -run '^TestM3ProbesKnownDefectNamesAnIssue$' -count=1 -v | grep -q '^--- PASS: TestM3ProbesKnownDefectNamesAnIssue'`)
- [ ] A `faithful` probe row whose snippet re-emits unfaithfully, or that perl -c rejects, fails (`go test ./internal/parse/ -run '^TestM3ProbesCheckFaithfulRows$' -count=1 -v | grep -q '^--- PASS: TestM3ProbesCheckFaithfulRows'`)
