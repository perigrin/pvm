---
title: Mine at least 200 perf/opcount.t assertions into fixtures held to perl
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:34.43792349Z
updated: 2026-10-22T23:47:19.964572596Z
---

## Context

- paths: `internal/parseoracle/opcount_test.go` (new), `internal/parseoracle/testdata/opcount/` (new)
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parseoracle/ -run '^(TestOpcountFixturesParse|TestOpcountFixturesHeldToPerl)$' -count=1 -v`
- entrypoints: `$HOME/dev/perl5/t/perf/opcount.t` (`test_opcount`, ~line 49)

M3 requires at least 200 `perf/opcount.t` assertions mined into fixtures. opcount.t holds 109 `test_opcount(` calls with 375 op=>count pairs; nothing in the repo mines them today. Each call gives a code snippet and the count of each named op perl compiles it to. This issue writes those as readable text/Markdown fixtures, not JSON, under `testdata/opcount/`. TestOpcountFixturesHeldToPerl asserts there are >= 200 op=>count assertions and runs perl on each snippet to confirm every count. The fixtures then hold their own provenance, and a later issue can grade internal/parse's sites against them. The test runs perl; no production code does. This issue shares no code with the rest of the milestone.

## Steps

- [ ] Read opcount.t's `test_opcount` and its callback to learn how it counts ops, including which ops it skips and how it treats nested subs. Quote that in the test's comment.
- [ ] Create `internal/parseoracle/opcount_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write TestOpcountFixturesParse. It reads every fixture in testdata/opcount/ into (name, snippet, op=>count) records, and asserts at least 200 op=>count pairs and the source line in opcount.t for each. Run it with no fixtures and watch it fail.
- [ ] GREEN: mine the fixtures from opcount.t, by hand or with a small one-off script that is not committed as production code. Each fixture names its opcount.t line. Run the test and watch it pass.
- [ ] RED: write TestOpcountFixturesHeldToPerl. It runs /home/perigrin/.local/bin/perl with B on each snippet, counting ops the way opcount.t does, and compares against the fixture. Flip one count by hand, watch the test fail, and restore it.
- [ ] GREEN: run the test and watch every fixture pass. Drop a fixture perl disagrees with only after quoting perl's output in the commit message.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The fixture directory is committed (`test -d internal/parseoracle/testdata/opcount`)
- [ ] The fixtures hold at least 200 op=>count assertions, each traced to its opcount.t line (`go test ./internal/parseoracle/ -run '^TestOpcountFixturesParse$' -count=1 -v | grep -q '^--- PASS: TestOpcountFixturesParse'`)
- [ ] perl 5.42 confirms every mined op count (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parseoracle/ -run '^TestOpcountFixturesHeldToPerl$' -count=1 -v | grep -q '^--- PASS: TestOpcountFixturesHeldToPerl'`)
