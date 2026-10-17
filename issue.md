---
title: Generate a .tpm skeleton from perl 5.42 for each core module
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-17T04:36:13.290610042Z
updated: 2026-10-17T05:06:48.403535093Z
sessions:
- start_sha: 5b91125f4014127e11963fc610c805ba205b10ff
  end_sha: 51c993cd397ad3e68b7ae8e64fdabf2624b1b4bf
  commits: 1
  started_at: 2026-10-17T04:36:37.923656263Z
  ended_at: 2026-10-17T04:58:40.29099013Z
- start_sha: 33ad8d4c2cb59a472bee475eca8881a92b333587
  end_sha: ""
  commits: 0
  started_at: 2026-10-17T05:06:48.403535093Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T04:36:37.923656263Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-17T04:58:40.29099013Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-17T05:06:48.350958821Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T05:06:48.403535093Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/parse/heldtoperl_test.go
- internal/parse/testdata/held_to_perl.pl
---

Batch workers transcribe hundreds of names per module by hand (a bare `use POSIX` installs 586). A test-only perl script generates a .tpm skeleton from perl 5.42 itself, so a worker only decides return types and oracle cases. perigrin chose this on 2026-10-17 to reduce token spend.

## Context
- db1c's driver `internal/parse/testdata/held_to_perl.pl` already measures prototypes (`can`, `prototype()`), installs (bare, qw(name), ':tag') and the callable rule. Reuse it; do not write a second measurer.
- RFC 0001 "Core-module libraries"; Carp.tpm is the reference shape.

## Steps
- [ ] Add a `skeleton` mode to held_to_perl.pl that prints, for a module, a .tpm skeleton: `package M;`, then one `sub NAME :export|:export_ok(tags) (PARAMS) TODO;` per installable name, params derived from prototype() (`$` -> `Any $a`, `@` -> `List @rest`, `\@` -> `Array \@a`, `;` -> optional, no prototype -> `List @args`), and `our TYPE NAME :export...` for exported variables with TODO type.
- [ ] `TODO` is not a type, so an unfilled skeleton fails to load: a worker must replace every TODO.
- [ ] A Go test helper `tpmSkeleton(t, module) string` runs it through perltest.
- [ ] Test: Carp's skeleton, with each TODO replaced by Carp.tpm's return type, agrees with Carp.tpm on every name, mark, tag and prototype (holdPrototypes and holdExports pass on it).
- [ ] Document the mode in batch-brief terms in the RFC's "Held to perl" paragraph (one sentence).

### Review Findings

Gate, 2026-10-17, on 51c993cd (fresh Sonnet reviewer; the three tests mutation-tested, none vacuous; skeleton and installs modes share the install measurement). VERDICT: REOPEN.

1. (CODE, in scope, 90) `internal/parse/testdata/held_to_perl.pl` skeletonParams never emits `:prototype(...)`. A lone `@` derives no prototype (derive.go:96, declarations.go:252, CORE.tpm states `:prototype(@)` explicitly) and `&`-leading prototypes fall back to `(List @args)`, so a filled List::Util skeleton fails holdPrototypes for 26 exports ("List::Util::max declares prototype none; perl's is (@)", "first ... perl's is (&@)"). Carp has no prototypes, so the criteria could not see it. Fix: when perl's prototype() is defined and non-empty, emit `:prototype(RAW)` with perl's literal string; test: List::Util's skeleton, TODOs filled with a placeholder type, passes holdPrototypes.

## Acceptance criteria

### Positive Scenarios
- [ ] The Carp skeleton matches Carp.tpm `go test ./internal/parse/ -run '^TestTpmSkeletonMatchesCarp$' -count=1 -v | grep -q '^--- PASS: TestTpmSkeletonMatchesCarp'`
- [ ] A skeleton marks tags from %EXPORT_TAGS `go test ./internal/parse/ -run '^TestTpmSkeletonMarksTags$' -count=1 -v | grep -q '^--- PASS: TestTpmSkeletonMarksTags'`

### Negative Scenarios
- [ ] An unfilled skeleton fails to load `go test ./internal/parse/ -run '^TestTpmSkeletonUnfilledRefused$' -count=1 -v | grep -q '^--- PASS: TestTpmSkeletonUnfilledRefused'`
