---
title: "[I23] $h{cb}(1) and $r->[0](2) lower to an array element Index, not a call; $c->(1)(2) lowers to a dynamic Method..."
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.043376609Z
updated: 2026-10-16T11:53:18.862679027Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: 2cea5857113d1d898a8a5651da708629c49a642d
  commits: 7
  started_at: 2026-10-16T01:17:16.515251901Z
  ended_at: 2026-10-16T11:53:18.862679027Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.515251901Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T11:53:18.862679027Z
observed_paths:
- docs/plans/2026-09-05-parser-conformance-plan.md
- docs/plans/2026-09-21-graded-conformance-corpus.md
- docs/specs/perl-parser/04-expressions.md
- docs/specs/perl-parser/05-statements.md
- docs/specs/perl-parser/06-incremental-lsp.md
- internal/infer/calls.go
- internal/infer/calls_test.go
- internal/infer/flow.go
- internal/infer/flow_test.go
- internal/infer/infer.go
- internal/infer/join_test.go
- internal/infer/lowered_test.go
- internal/infer/ops.go
- internal/infer/ops_test.go
- internal/infer/returns_test.go
- internal/infer/scope_test.go
- internal/infer/symbols.go
- internal/infer/testdata/lowered/calls/param_usage.golden
- internal/infer/testdata/lowered/decl/package.golden
- internal/infer/testdata/lowered/decl/sub_anon.golden
- internal/infer/testdata/lowered/decl/sub_named.golden
- internal/infer/testdata/lowered/flow/block_decl_in_branch.golden
- internal/infer/testdata/lowered/flow/for_my_scalar.golden
- internal/infer/testdata/lowered/flow/isa_else.golden
- internal/infer/testdata/lowered/flow/not_and.golden
- internal/infer/testdata/lowered/flow/or_else.golden
- internal/infer/testdata/lowered/flow/sub_guard_body.golden
- internal/infer/testdata/lowered/flow/user_compound.golden
- internal/infer/testdata/lowered/ops/compound_assign.golden
- internal/infer/testdata/lowered/ops/defined_or.golden
- internal/infer/testdata/lowered/special/capture_arith.golden
- internal/infer/testdata/lowered/special/do_block.golden
- internal/infer/testdata/lowered/special/ref_arith.golden
- internal/infer/testdata/lowered/special/ref_paren.golden
- internal/infer/testdata/lowered/special/wantarray.golden
- internal/infer/testdata/lowered/special/wantarray_top.golden
- internal/infer/testdata/lowered/vars/arraylen.golden
- internal/infer/testdata/lowered/vars/elem_from_var.golden
- internal/infer/testdata/lowered/vars/our_element.golden
- internal/infer/vars.go
- internal/infer/vars_test.go
- internal/lexer/brackets.go
- internal/lexer/expect.go
- internal/lexer/heredoc.go
- internal/lexer/heredocterminator_test.go
- internal/lexer/ident.go
- internal/lexer/ident_test.go
- internal/lexer/keyword.go
- internal/lexer/lexer.go
- internal/lexer/scan.go
- internal/lexer/versionpadding_test.go
- internal/lower/call.go
- internal/lower/call_test.go
- internal/lower/decl.go
- internal/lower/decl_test.go
- internal/lower/expr.go
- internal/lower/id.go
- internal/lower/literals.go
- internal/lower/literals_test.go
- internal/lower/node.go
- internal/lower/positions.go
- internal/lower/positions_test.go
- internal/lower/special.go
- internal/lower/special_test.go
- internal/lower/stmt.go
- internal/lower/vars.go
- internal/lower/vars_test.go
- internal/parse/arity.go
- internal/parse/canon.go
- internal/parse/decl.go
- internal/parse/declarations/Carp.pmt
- internal/parse/declcompound_test.go
- internal/parse/deepnesting_test.go
- internal/parse/expr.go
- internal/parse/loader_regular_test.go
- internal/parse/module.go
- internal/parse/namelesspackage_test.go
- internal/parse/nestedquadratic_test.go
- internal/parse/parse.go
- internal/parse/powerafterterm_test.go
- internal/parse/refusal.go
- internal/parse/refusal_test.go
- internal/parse/staysinroots_test.go
- internal/parse/term.go
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/use.go
- internal/perl/inc.go
- internal/perl/inc_test.go
- internal/perl/static.go
- internal/perl/static_test.go
- internal/psc/check_command.go
- internal/psc/check_command_test.go
- internal/psc/check_escape_test.go
- internal/psc/inc_roots_test.go
- internal/psc/lsp.go
- internal/psc/resolve_guard_test.go
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I23**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/lower/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/lower/vars.go:401-410` (`subscript` accepts `index "("`), after `indexCall` declines (`call.go:82-100`). Also `arrowCall` (call.go:120).
- **Repro:**
  - `$h{cb}(1);` gives `Index Base:(Key ...) Idx:(NumLit 1)`.
  - `$c->(1)(2);` gives `MethodCall HasParens Invocant:$c Dynamic:(NumLit 1) Args:2`.
  - perl calls the code ref in each case.
- **Impact:** a wrong tree on realistic code (TAP/Harness.pm `$self->{bail_summary}()`). IsDeref reports it as an array deref, and the call is invisible to the call walks. `psc check --strict` on `v/cb.pl` reports nothing.
- **Fix:** `subscript` declines `(`. `indexCall` lowers any other base to `Call{Fn: base, Args, Parenthesized}`. arrowCall requires a `$`-led base before it reads a dynamic method.
- **Confidence:** High. **NEW** (the lowering). **Issue:** none. **Found by:** contract-A #1.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `$h{cb}(1)` and `$r->[0](2)` lower to an array element `Index`, not a call; `$c->(1)(2)` lowers to a dynamic MethodCall is fixed (`go test ./internal/lower/ -run '^TestCodeRefCallThroughSubscript$' -count=1 -v | grep -q '^--- PASS: TestCodeRefCallThroughSubscript'`)
