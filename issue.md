---
title: A Glob in a Str slot is rejected, though perl stringifies a glob silently
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-15T00:27:13.433187146Z
updated: 2026-10-15T00:27:30.24616296Z
sessions:
- start_sha: 6fed0273b415fdabaec853b07adab5c4df6f5598
  end_sha: ""
  commits: 0
  started_at: 2026-10-15T00:27:30.24616296Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-15T00:27:30.24616296Z
---

## Context

- found by: cutover 1 (01a0c115-a44b), round-3 re-measure, bucket `R3-GLOB-STR` (2026-10-15); evidence in `scratchpad/cmp-a44b/ev-r3.md`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed before it merges. Pod/Simple.pm:451: under `ref(\$source) eq 'GLOB'` (narrowed to Glob by 01a13c6a-71a1), `'' . ($source)` is reported `right operand of ".": expected Str, got Glob`; perl -w is silent and prints `*main::STDIN`. A glob stringifies to its name without warning; a Glob satisfies a Str slot. Legacy was silent.

```perl
my $source = *STDIN;
if (ref(\$source) eq 'GLOB') { my $name = '' . ($source); print "$name\n" }
```

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42 -w's verdict holds, keeping a control perl warns on; precision must stay >= 64/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the repro is silent as perl is (`go test ./internal/infer/ -run '^TestGlobInStrSlotIsSilent$' -count=1 -v | grep -q '^--- PASS: TestGlobInStrSlotIsSilent'`)
