---
title: "`use if` is read only as failing: parse it both ways"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-30T00:48:36.055750268Z
updated: 2026-09-30T00:54:52.536516741Z
sessions:
- start_sha: ecb04d74383f860ea152ed4286c0b779d42a4bd1
  end_sha: 2149d9471a9dd509899efbe7860716bf6b6daeeb
  commits: 1
  started_at: 2026-09-30T00:48:38.230831185Z
  ended_at: 2026-09-30T00:54:52.536516741Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T00:48:38.230831185Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-30T00:54:52.536516741Z
observed_paths:
- internal/parse/exports.go
- internal/parse/module.go
- internal/parse/perlgitt_test.go
- internal/parse/use.go
- internal/parse/useifboth_test.go
---

The ratchet reads every `use if COND, MODULE, LIST` as failing, because its condition is known only at compile time. perigrin decided (2026-09-30) that a file using it must be clean both ways: failing, where it defines nothing, and passing, where it is `use MODULE LIST`.

The passing reading needs a parse option and a harness check. Measured with the option, perl.git t/op/coreamp.t had 2 refusals: its `qw "catfile"` import list has a space before the delimiter, and literalNameList read the name as `"catfile`. perl 5.42 reads `qw "a b"` as ("a", "b").

## Acceptance criteria

- [ ] `use if` parses both ways (`go test ./internal/parse/ -run '^TestUseIfBothWays$' -count=1 -v | grep -q '^--- PASS: TestUseIfBothWays'`)
- [ ] Every perl.git t/ file using `use if` is clean with it passing (`go test ./internal/parse/ -run '^TestPerlGitTUseIfPasses$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTUseIfPasses'`)
- [ ] A spaced `qw "..."` import list imports its names (`go test ./internal/parse/ -run '^TestSpacedQwImportList$' -count=1 -v | grep -q '^--- PASS: TestSpacedQwImportList'`)
