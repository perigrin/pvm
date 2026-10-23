---
title: Resolve a named hash's declaring scope and whether state is enabled
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a1664d-81b7-7c96-8f48-636ae47d5cb1
blocks:
- 01a16602-61a2-78de-a798-da4cef3208c5
created: 2026-10-23T01:08:36.424056648Z
updated: 2026-10-23T01:09:12.391725524Z
---

## Context

- paths: `internal/parse/hashscope.go` (new), `internal/parse/hashscope_test.go` (new), `internal/parse/testdata/m3probes/hash-scope.txt` (new)
- commands: `go test ./internal/parse/ -run '^(TestHashScopeResolvesInnermostDeclaration|TestHashScopeTracksStateFeature)$' -count=1 -v`
- entrypoints: `internal/parse/sites.go` (the `Index` case), `internal/parse/call.go:745` ("this site has no feature state")

Deciding a named hash's rv2hv from its declaration needs something internal/parse does not have: a lexical scope resolver. The only scope types live in internal/infer (symbols.go, flow.go), and internal/infer imports parse, so they cannot be reused here. This issue builds the minimum the rv2hv decision needs. For each `%h` use, it answers which declaration is in effect (`my`, `our`, `state`, `local`, or none visible), honouring the innermost declaration, blocks that have closed, and use-before-declaration. It also answers whether the `state` feature is on at that point. perl measured on 5.42.0:
- `keys %h; my %h; keys %h` gives 1 rv2hv, then 0.
- With the feature on, `state %h; keys %h` is padhv with 0 rv2hv.
- Under `no feature "state"`, `state %h` is not a declaration (perl reads `state % h`), and the later `%h` is a package hash with 1 rv2hv.
- `use v5.10` and later bundles turn the feature on, and `use v5.36` does too.

The next issue uses this resolver to decide rv2hv.

## Steps

- [ ] Re-measure every case above with the pinned perl 5.44.0 (`perl -MO=Concise`), and quote the output in the tests' comments.
- [ ] Create both files with two `// ABOUTME:` lines each.
- [ ] RED: write TestHashScopeResolvesInnermostDeclaration, with fixtures for the innermost `my` shadowing `our`, a closed block's `my` not leaking, use before `my`, and `local` of an `our` hash. Run it and watch it fail.
- [ ] GREEN: implement the resolver over the parse tree. Run the test and watch it pass.
- [ ] RED: write TestHashScopeTracksStateFeature, with fixtures for `use feature "state"`, `no feature "state"`, `use v5.10`, `use v5.36`, a feature turned on inside a block and off after it, and no feature at all. Run it and watch it fail.
- [ ] GREEN: track the feature lexically. Run the test and watch it pass.
- [ ] Add the fixtures, with perl's rv2hv counts, as probe file `testdata/m3probes/hash-scope.txt`, and run TestM3ProbesHeldToPerl.
- [ ] Per-fix review against perl before merge. The reviewer writes NEW probes outside the corpus, such as `our %h` redeclared in a nested package, string-eval'd declarations, and `use feature ':5.10'`.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] A hash use resolves to the innermost declaration in effect, as perl measures (`go test ./internal/parse/ -run '^TestHashScopeResolvesInnermostDeclaration$' -count=1 -v | grep -q '^--- PASS: TestHashScopeResolvesInnermostDeclaration'`)
- [ ] The resolver knows whether the state feature is on, including via use v5.10 and use v5.36 bundles (`go test ./internal/parse/ -run '^TestHashScopeTracksStateFeature$' -count=1 -v | grep -q '^--- PASS: TestHashScopeTracksStateFeature'`)
- [ ] The resolver's cases are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/hash-scope.txt`)
