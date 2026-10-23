---
title: State the M3 gate's measured numbers in the plan and READMEs
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
- 01a16602-5e01-7e7a-92c5-27d7da7d6f36
- 01a16602-5f1a-7d56-a91d-5fa2c478d608
- 01a16602-5fd5-7cc4-9645-4539ca73e839
- 01a16602-6030-7faa-8408-bb3a8cb03f12
- 01a16602-608d-71ed-8ab2-66a9951d9957
- 01a16602-60e9-7343-a630-7a93da5ee389
- 01a16602-6146-76ab-bb68-f3e8b645ae79
- 01a16602-61a2-78de-a798-da4cef3208c5
- 01a1664d-6fdb-7a09-8211-882e814db87c
- 01a16795-75e8-7787-81ad-79e130324139
created: 2026-10-22T23:46:34.622478576Z
updated: 2026-10-23T07:07:01.085965939Z
---

## Context

- paths: `docs/plans/2026-09-05-parser-conformance-plan.md` (§5 M3 at ~606, the "never the parens the source had" sentence at ~620, the ladder summary at ~729), `docs/specs/perl-parser/07-conformance.md` (~308), `docs/specs/perl-parser/README.md` (Status at ~99, the stale `($x)` sentence at ~129), `internal/parseoracle/README.md` (~287), `internal/parseoracle/docs_test.go` (new)
- commands: `go test ./internal/parseoracle/ -run '^(TestPlanM3SectionQuotesSitesRatchet|TestM3DocsQuoteCommittedNumbers)$' -count=1 -v`
- entrypoints: `internal/parseoracle/testdata/ratchet/sites.txt`, `internal/parseoracle/testdata/ratchet/m3start.txt`, `internal/parse/testdata/compiledclean.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`, `internal/parse/testdata/faithful_queue.txt`, `internal/parseoracle/testdata/opcount/`

State the M3 gate's measured numbers where readers look, taken from the committed files rather than from memory, and check that every remaining condition of the milestone holds. The plan's M3 table says "≥ 526 of 584" and its ladder row says "584 files". Both are stale: the oracle is now pinned to perl 5.44.0, and the compiled count is m3start.txt's `compiled`. "593" was the 5.42 graded set, not a compile count, so the docs must not call it one. The README's defect sentence that `($x)` loses its `Paren` node (term.go:284-293) is also stale. And the plan (~620, "never the parens the source had") and 07-conformance.md:308 ("never the parens seen in the source") must be reconciled with perigrin's Paren-once ruling. canon writes a `Paren` node's parens once because the tree carries the paren, exactly as perl's op carries one OPf_PARENS bit, so the tree implies them. That needs saying.

## Steps

- [ ] Read each final number from its committed file:
  - sites.txt (whole corpus and M3-dir exact/wider/no-answer/WRONG, rv2hv wider) and m3start.txt (the starting values and `compiled`);
  - compiledclean.txt (clean of compiled);
  - perlgittfaithful.ratchet (faithful of 619, and of the M3 dirs, which must be 100% of what parses), with the start counts in m3start.txt;
  - the opcount fixtures (assertion count).
- [ ] Plan §5 M3: change the parse row to the compiled count and add a `#### Measured at the M3 gate` subsection. It gives each metric's measured value, and the perl (5.44.0) and commit it was measured at. It states that canonical re-emission is 100% faithful on every M3-dir file that parses, with no defects file, per perigrin's round-4 ruling, and gives the start counts beside it (178/347 M3 dirs, 342/619 overall).
- [ ] Plan ladder summary: change the M3 row's 584 to the compiled count. Confirm with `grep -nw 584` that no 584 is left anywhere in the plan.
- [ ] Reconcile plan ~620 and 07-conformance.md:308 with the Paren-once ruling: canon writes the parens the tree implies, a `Paren` node's pair included, never the source's extra nesting.
- [ ] docs/specs/perl-parser/README.md: add a Status paragraph starting `The M3 gate` with the measured numbers, and correct the stale `($x)` sentence. canon now writes a `Paren` node's parens once, and `Faithful` forgives only extra nesting.
- [ ] internal/parseoracle/README.md: check that the "Ground-truth ceiling" sentence (~287) states m3start.txt's compiled count at 5.44.0, as the re-pin issue set it. Leave the historical "became 584/620" sentence (~309) alone.
- [ ] Create `internal/parseoracle/docs_test.go` with two `// ABOUTME:` lines. RED/GREEN TestM3DocsQuoteCommittedNumbers: the plan subsection and the README `The M3 gate` paragraph state the compiled count (m3start.txt), the clean count (compiledclean.txt), the faithful count (perlgittfaithful.ratchet header), the M3-dir faithful count (ratchet rows) stated as 100%, and the opcount assertion count (fixtures), each equal to its file, and the parseoracle README ceiling equals m3start.txt's `compiled`.
- [ ] Run the regression list below and the milestone resolution command, and watch them pass. Open the PR, wait for CI (Windows and macOS included), then commit any fix and re-check.

## Acceptance criteria

### Positive Scenarios

- [ ] The plan has a measured-at-the-gate subsection for M3 (`grep -q '^#### Measured at the M3 gate' docs/plans/2026-09-05-parser-conformance-plan.md`)
- [ ] The stale 584 is gone from the plan (`sh -c '! grep -qw 584 docs/plans/2026-09-05-parser-conformance-plan.md'`)
- [ ] The perl-parser README's Status states the M3 gate (`grep -q '^The M3 gate' docs/specs/perl-parser/README.md`)
- [ ] Every number the docs state equals its committed file (`go test ./internal/parseoracle/ -run '^TestM3DocsQuoteCommittedNumbers$' -count=1 -v | grep -q '^--- PASS: TestM3DocsQuoteCommittedNumbers'`)
- [ ] The parseoracle README's ceiling sentence no longer claims 584/620; the historical sentence at ~309 may stay (`sh -c 'test -f internal/parseoracle/README.md && ! grep -q "\*\*584/620" internal/parseoracle/README.md'`)
- [ ] The perl-parser README no longer cites term.go:284-293 for the dropped Paren (`sh -c '! grep -q term.go:284-293 docs/specs/perl-parser/README.md'`)
- [ ] Plan and conformance spec no longer say canon never writes source parens (`sh -c 'test -f docs/specs/perl-parser/07-conformance.md && ! grep -q "never the parens the source had" docs/plans/2026-09-05-parser-conformance-plan.md && ! grep -q "never the parens seen in the source" docs/specs/perl-parser/07-conformance.md'`)
- [ ] Regression guard: perl.git t/ stays 619/619 clean (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] Regression guard: T2 still parses (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] Regression guard: T2 holds all 56 (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestT2HoldsAll56$' -count=1 -v | grep -q '^--- PASS: TestT2HoldsAll56'`)
- [ ] Regression guard: T2 round-trips losslessly (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/lexer/ -run '^TestT2CoreRoundTrips$' -count=1 -v | grep -q '^--- PASS: TestT2CoreRoundTrips'`)
- [ ] Regression guard: the parse corpus ratchet holds (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] Regression guard: the conformance corpus ratchet holds (`go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1 -v | grep -q '^--- PASS: TestCorpusRatchet'`)
- [ ] Regression guard: the exact-match floor holds over T2 at the 5.44.0 pin (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim go test ./internal/parseoracle/ -run '^TestGoParserExactFloor$' -count=1 -v | grep -q '^--- PASS: TestGoParserExactFloor'`)
- [ ] Regression guard: inference precision stays at or above 65/76 (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)
- [ ] Regression guard: every committed perl-judged probe still holds (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 -v | grep -q '^--- PASS: TestM3ProbesHeldToPerl'`)
- [ ] Release check, pinned to the checked-out commit (it passes for today's HEAD, and at completion it judges the commit that closes the milestone): CI's main workflow succeeded on exactly this commit (`sh -c 'gh run list --workflow ci.yml --commit "$(git rev-parse HEAD)" -L 1 --json conclusion -q ".[0].conclusion" | grep -qx success'`)
- [ ] Release check, pinned to the checked-out commit: CI's parse-fidelity workflow succeeded on exactly this commit (`sh -c 'gh run list --workflow parse-fidelity.yml --commit "$(git rev-parse HEAD)" -L 1 --json conclusion -q ".[0].conclusion" | grep -qx success'`)

### Negative Scenarios

- [ ] The measured subsection's numbers are the committed ones, not a heading over stale prose. It states the whole-corpus and M3-dir site totals and the rv2hv wider count exactly as sites.txt records them, so editing either side alone fails (`go test ./internal/parseoracle/ -run '^TestPlanM3SectionQuotesSitesRatchet$' -count=1 -v | grep -q '^--- PASS: TestPlanM3SectionQuotesSitesRatchet'`)
- [ ] The corrected `($x)` sentence describes the extra-nesting ruling, not just a removed citation (`sh -c 'grep -qi "extra nesting" docs/specs/perl-parser/README.md'`)
- [ ] The measured subsection states re-emission at 100% on the M3 dirs, and the work queue it came from is empty (`sh -c 'f=internal/parse/testdata/faithful_queue.txt; test -f "$f" && ! grep -q "^[^#]" "$f" && sed -n "/^#### Measured at the M3 gate/,/^##/p" docs/plans/2026-09-05-parser-conformance-plan.md | grep -q "100%"'`)
