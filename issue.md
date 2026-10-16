---
title: Pin the core-module ranking in a test and record the rule in the RFC
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
- 01a1464e-db4c-78ae-9bd5-84d71ccf5432
- 01a1464e-db76-704f-bd85-023a54d2cf6e
- 01a1464e-dba9-740b-b8b6-8da484195ce3
- 01a1464e-dbd3-7b16-801c-b9a5893d687b
- 01a1464e-dc07-754e-9ef4-637df15d51a9
- 01a1464e-dc48-720b-9ab9-fcc7fe47f483
- 01a1464e-dca3-7284-85c3-a06107c16646
- 01a1464e-dd27-7b82-ae71-3e6173375fa7
created: 2026-10-16T20:02:15.040084527Z
updated: 2026-10-16T20:04:26.771832181Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one

## Context

- paths: `internal/parse/ranking_test.go` (new), `internal/parse/testdata/core_module_ranking.txt` (new), `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -run '^TestCoreModuleRanking' -count=1 -v`
- entrypoints: `internal/perltest` (from the baseline issue), RFC 0001 "Core-module libraries", "Which 50."

The ranking fixes which 50 modules get declarations, so every batch issue depends on it. The rule (perigrin, 2026-10-16), stated verbatim in the test's comment:

- A core module is any `.pm` under privlib, named with the arch directory stripped (`x86_64-linux-thread-multi/` here; in general archlib relative to privlib, so the rule holds on macOS too).
- An import is a line matching `^\s*(use|require)\s+[A-Z][A-Za-z0-9_:]+` in any `.pm` or `.pl` under privlib.
- An import is not counted when the importing file's top-level path component, with the arch dir stripped and `.pm`/`.pl` removed, equals the module's first `::` component.
- Only imports of core modules count. Ties are broken by module name, ascending.

Measured by the decomposer on 2026-10-16 with a perl script implementing that rule over `/home/perigrin/.local/share/pvm/versions/5.42.0/lib/perl5/5.42.0`: **128 modules, 1,427 imports -- the RFC's figure reproduces** -- and the top 50 take 1,326 (92.9%, the RFC's "93%"). Ranks 45-51 tie at 3 imports; ascending name puts Net::Ping 50th and Params::Check 51st. The top 50, in order:

Carp 417, Exporter 168, Scalar::Util 76, Config 74, File::Spec 65, XSLoader 48, Fcntl 37, Cwd 33, File::Basename 33, Encode 27, Symbol 27, POSIX 26, Socket 22, Data::Dumper 20, File::Find 16, File::Path 16, List::Util 16, Errno 12, Text::ParseWords 12, Class::Struct 11, FileHandle 10, DynaLoader 9, Compress::Raw::Zlib 8, Getopt::Long 8, IO::Socket 8, Test::More 8, File::Copy 7, File::Spec::Functions 7, MIME::Base64 7, SDBM_File 7, Time::HiRes 7, File::Temp 6, IO::File 6, IO::Handle 6, Text::Wrap 6, Time::Local 6, IO::Socket::IP 5, Storable 5, Term::ReadLine 5, Tie::Hash 5, ExtUtils::MakeMaker 4, File::Spec::Unix 4, IO::Select 4, Locale::Maketext::Simple 4, Archive::Tar 3, ExtUtils::Manifest 3, File::stat 3, Getopt::Std 3, HTTP::Tiny 3, Net::Ping 3.

The test walks privlib in Go (`filepath.WalkDir`, `filepath.ToSlash` on relative paths), computes the ranking, and checks the top 50 (name and count) against `testdata/core_module_ranking.txt`, plus the totals 128 / 1,427. It skips where perl 5.42.0 or privlib is absent. A second test runs everywhere: the rule applied to a synthetic tree under `t.TempDir()` (own-namespace exclusion, arch-dir stripping, `.pl` files counted, a lowercase pragma not counted, a tie broken ascending).

## Steps

- [ ] Write failing test `TestCoreModuleRankingRule` over a synthetic tree in `t.TempDir()`
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestCoreModuleRankingRule$' -count=1 -v`
- [ ] Implement the ranking in the test file; run to verify pass; commit
- [ ] Write `testdata/core_module_ranking.txt` (the 50 above, one `Name count` per line) and `TestCoreModuleRanking` over privlib; run to verify pass: `go test ./internal/parse/ -run '^TestCoreModuleRanking$' -count=1 -v`
- [ ] In the RFC's "Which 50." paragraph, state the rule, record the top-50 figure (1,326) and that ties are broken by module name, ascending
- [ ] `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] the rule ranks a synthetic tree as stated (`go test ./internal/parse/ -run '^TestCoreModuleRankingRule$' -count=1 -v | grep -q '^--- PASS: TestCoreModuleRankingRule'`)
- [ ] the ranking over perl 5.42's lib matches the committed top 50 and 128 / 1,427 (`go test ./internal/parse/ -run '^TestCoreModuleRanking$' -count=1 -v | grep -q '^--- PASS: TestCoreModuleRanking'`)
- [ ] the committed list has 50 modules (`test "$(grep -c . internal/parse/testdata/core_module_ranking.txt)" -eq 50`)
- [ ] the RFC records the tie-break (`grep -q 'Ties are broken by module name, ascending' docs/rfcs/0001-type-libraries.md`)
- [ ] the RFC records the top-50 figure (`grep -q '1,326' docs/rfcs/0001-type-libraries.md`)
