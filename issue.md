---
title: "Drop :context: a multi selects by return type in the calling context"
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a118fe-8d68-71a8-ab06-cc12d7a20589
- 01a11937-22da-7250-b922-fabda15009a7
blocks:
- 01a118fe-8e5f-717a-b046-6020db95c724
created: 2026-10-08T01:28:26.650675969Z
updated: 2026-10-16T18:48:34.066148815Z
sessions:
- start_sha: 04a6ca62d20d1d53be57651b769964909706bf69
  end_sha: 04a6ca62d20d1d53be57651b769964909706bf69
  commits: 0
  started_at: 2026-10-08T08:31:40.328391047Z
  ended_at: 2026-10-08T08:39:39.426869668Z
- start_sha: 5f2311c758d827fdcb132ebd657a21508edaa80a
  end_sha: cd34146a827f624890247b5c2383db9e2261ebc9
  commits: 1
  started_at: 2026-10-08T09:58:19.606471674Z
  ended_at: 2026-10-08T10:57:12.191161454Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-08T08:31:40.328391047Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-08T08:39:39.426869668Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-08T09:58:19.606471674Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-08T10:57:12.191161454Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- docs/specs/perl-parser/04-expressions.md
- internal/infer/core_test.go
- internal/infer/infer.go
- internal/lexer/proto.go
- internal/lexer/proto_test.go
- internal/lexer/scanrows_test.go
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/chain.go
- internal/parse/corearray_test.go
- internal/parse/coregrepmapsort_test.go
- internal/parse/corenoproto_test.go
- internal/parse/coreops_test.go
- internal/parse/coreprocess_test.go
- internal/parse/decl.go
- internal/parse/declarations.go
- internal/parse/declarations/CORE.pmt
- internal/parse/derive.go
- internal/parse/derive_test.go
- internal/parse/expr.go
- internal/parse/expr_test.go
- internal/parse/library_pmt_test.go
- internal/parse/operators.go
- internal/parse/parse.go
- internal/parse/pmt_test.go
- internal/parse/precedence.go
- internal/parse/precrelations_test.go
- internal/parse/testdata/declarations/My/Lib.pmt
- internal/types/declared.go
- internal/types/narrowing.go
- internal/types/select.go
- internal/types/select_test.go
---

## Context

- paths: `internal/types/declared.go` (`Contexts`, `ContextSet`, `EveryContext`), `internal/types` selection (`Select`/`SelectShaped`), `internal/parse/decl.go` (attribute reading), `internal/parse/declarations/CORE.pmt` (37 `:context` candidates), `docs/rfcs/0001-type-libraries.md` ("`:context(...)`", Implemented)
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`

perigrin, 2026-10-08: drop `:context`, since multi and the return type handle it. A call site's context is a demand on the result: list context selects the candidate whose return type is a list (`List`, `List[...]`), scalar context the one whose return type is not. Measured over CORE.pmt: all 37 `:context` candidates already obey this (every `:context(@)` returns a list type, every `:context($)` a scalar one), so the attribute states nothing the return type does not. `reverse` is the pure case: identical parameters, `Str` versus `List`. `x`'s list candidate is also chosen by its parenthesised left operand, which the `List @l` shape check already enforces.

Context is still decided syntactically at the call site; only the attribute goes, not context analysis. Void context (perigrin, 2026-10-08): the lattice gains `Void` (the arity-0 type, `{()}`; not `None`, which is "does not return") in 01a11937-22da-7250-b922-fabda15009a7; a candidate returning `Void` is the void-context candidate, as `:context()` was. The 5 test fixtures using `:context()` become `Void` returns.

## Steps

- [ ] Write failing test `TestContextSelectsByReturnType`: with the `:context` attributes removed from a fixture multi (`reverse`-shaped: same params, `Str` and `List` returns), scalar context selects the `Str` candidate and list context the `List` one; also `split`, `each`, `keys`, `localtime` as CORE.pmt declares them
- [ ] Write failing test `TestContextAttributeRefused`: a `.pmt` line stating `:context(...)` is a declaration error naming it
- [ ] Write failing test `TestVoidContextSelectsVoidReturn`: in void context a candidate returning `Void` is selected over one returning a scalar or list type
- [ ] Write failing test `TestReturnTypeContextAmbiguity`: two candidates with the same parameters and both list (or both scalar) returns are an ambiguity declaration error
- [ ] Implement selection by return type; delete `Contexts` and the attribute reader; drop `:context` from CORE.pmt's 37 lines; restate RFC 0001 "`:context(...)`"; infer and psc results must not change (existing tests unchanged); `make test`; commit

## Acceptance criteria

- [ ] void context selects the Void-returning candidate (`go test ./internal/types/ ./internal/parse/ -run '^TestVoidContextSelectsVoidReturn$' -count=1 -v | grep -q '^--- PASS: TestVoidContextSelectsVoidReturn'`)
- [ ] context selects by return type (`go test ./internal/types/ ./internal/parse/ -run '^TestContextSelectsByReturnType$' -count=1 -v | grep -q '^--- PASS: TestContextSelectsByReturnType'`)
- [ ] a :context attribute is refused (`go test ./internal/parse/ -run '^TestContextAttributeRefused$' -count=1 -v | grep -q '^--- PASS: TestContextAttributeRefused'`)
- [ ] same-shaped returns are ambiguous (`go test ./internal/types/ ./internal/parse/ -run '^TestReturnTypeContextAmbiguity$' -count=1 -v | grep -q '^--- PASS: TestReturnTypeContextAmbiguity'`)
- [ ] no :context remains in CORE.pmt (`sh -c '! grep -q ":context(" internal/parse/declarations/CORE.tpm'`)
