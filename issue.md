---
title: "Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-9671-733f-8b98-30a78f634b0f
- 01a0c110-839f-77fb-8a3d-30fa5a8d6e96
created: 2026-09-20T23:10:16.13915176Z
updated: 2026-09-20T23:12:51.705179Z
---

## Context

- paths: `internal/infer/lowered.go` (new), `internal/infer/lowered_test.go` (new), `internal/psc/check_command.go:94`
- commands: `go test ./internal/psc/ -run "^TestCheckCommandOnLoweredTree$" -count=1 -v`
- entrypoints: `internal/infer/infer.go:64` (`Analyze`), `internal/psc/check_command.go:83-97`

The strangler-fig step. `internal/infer` has three entry points — `Analyze`
(`infer.go:64`), `AnalyzeWithOptions` (`infer.go:121`) and `NewProjectIndex`
(`project.go:38`) — and exactly TWO non-test callers:

	internal/psc/lsp.go:53            annotations, diagnostics, symbols := infer.Analyze(tree, source, nil)
	internal/psc/check_command.go:94  _, diags, _ := infer.AnalyzeWithOptions(tree, source, nil, opts)

That is the whole boundary, which is what makes a staged cutover possible
instead of a big bang. This issue adds `infer.AnalyzeLowered` ALONGSIDE the
existing entry points and flips the BATCH caller only — `check_command.go`, the
smaller surface, with no incremental path, no hover, and no position queries to
get right at the same time. The LSP flip is a separate issue and comes second.

Nothing is deleted here. The old walkers keep working against
`internal/parser/` until the LSP flip lands and the retirement issue removes
them; a half-migrated `internal/infer` with both paths live is the intended
state for exactly one issue's duration.

**What `AnalyzeLowered` must reproduce.** The diagnostics `check_command`
already emits. The cutover is not the place to change what PSC reports — a
diagnostic that changes here is either a fix that belongs in its own issue or a
regression. The test for that is differential: run both paths over the same
corpus and compare.

**Where the vocabulary bites.** `Kind()` on the old tree returns tree-sitter
grammar names; the lowered tree has §4.14's kinds and none of those names. Every
rule ported here is REWRITTEN against the lowered vocabulary, not shimmed. §6.4.1
is explicit that an earlier draft concluded from the ten-method count that PSC
would not change — "not one of its 4,265 lines" — and that this was the wrong
count. Ten methods port; the vocabulary does not.

**What gets measurably simpler, and these are the things to check in review:**
- the five `GetBuiltin` sites (`infer.go:693`, `:881`, `:1009`, `:4004`,
  `:4048`) collapse to one `Call` arm
- `findOperatorText` (`infer.go:645-659`) has no successor — `Binary.Op` is a
  field
- `callElementType`'s `strings.HasPrefix(..., "grep")` (`infer.go:1313`) becomes
  a switch on `MapGrepSort.Op`
- the `infer.go:1155-1165` filehandle workaround has no successor —
  `Call.Handle` is a separate field from `Call.Args`
- the three `@_`/`$_[0]` text re-derivations (`infer.go:3659`, `:3753`, `:4193`)
  compare `Var{Sigil,Name}` fields
- the 8-of-9 `scalar`/`array`/`hash` switch sites that differ only by sigil
  string become one arm reading `Var.Sigil`

Scope discipline: **no dataflow IR** (§6.1.6). §6.1.7 owns flow and is unbuilt.
If a rule wants flow, it keeps whatever approximation it has today.

## Steps

- [ ] Write `TestCheckCommandOnLoweredTree`: run `checkFile` through the lowered
      path over a fixture exercising each lowered kind, and assert the
      diagnostics match what the current path emits for the same fixture.
- [ ] Run to verify failure: `go test ./internal/psc/ -run "^TestCheckCommandOnLoweredTree$" -count=1 -v`
- [ ] Add `infer.AnalyzeLowered` in `internal/infer/lowered.go`, porting rules
      slice by slice against the lowered vocabulary. Run to green. Commit at
      each slice — literals, variables, operators, declarations, calls, special
      forms — rather than once at the end.
- [ ] Write `TestLoweredAndLegacyAgreeOnCorpus`: a differential test running
      both paths over the parse package's existing corpus fixtures and asserting
      the diagnostic sets are equal. Any divergence is triaged as fix-or-
      regression before the flip, not after.
- [ ] Run to verify failure, implement, run to green, commit.
- [ ] Flip `check_command.go:94` to `AnalyzeLowered`. Run to green. Commit.
- [ ] Run the full suite: `make test`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] The batch checker produces its diagnostics from the lowered tree (`go test ./internal/psc/ -run "^TestCheckCommandOnLoweredTree$" -count=1 -v | grep -q "^--- PASS: TestCheckCommandOnLoweredTree"`)
- [ ] Both paths agree over the corpus (`go test ./internal/psc/ -run "^TestLoweredAndLegacyAgreeOnCorpus$" -count=1 -v | grep -q "^--- PASS: TestLoweredAndLegacyAgreeOnCorpus"`)
- [ ] The psc package is green (`go test ./internal/psc/ -count=1`)
- [ ] The whole repository is green (`make test`)
