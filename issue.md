---
title: "[I26] The LSP's long-lived parse.Session never re-reads a module, and caches \"not found\" forever"
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.193695538Z
updated: 2026-10-16T12:06:02.559293187Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: 77107dfbeb77e7b3f1d94437583eae6d28eb41c3
  commits: 8
  started_at: 2026-10-16T01:17:18.730330071Z
  ended_at: 2026-10-16T12:06:02.559293187Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:18.730330071Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T12:06:02.559293187Z
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
- internal/infer/golden_test.go
- internal/infer/infer.go
- internal/infer/join_test.go
- internal/infer/lowered_test.go
- internal/infer/ops.go
- internal/infer/ops_test.go
- internal/infer/project.go
- internal/infer/project_test.go
- internal/infer/returns_test.go
- internal/infer/scope_test.go
- internal/infer/symbols.go
- internal/infer/testdata/lowered/calls/My/Block.pm
- internal/infer/testdata/lowered/calls/imported_block.golden
- internal/infer/testdata/lowered/calls/imported_block.pl
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
- internal/infer/testdata/lowered/matching.txt
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
- internal/psc/analyze_command.go
- internal/psc/check_command.go
- internal/psc/check_command_test.go
- internal/psc/check_escape_test.go
- internal/psc/check_once_test.go
- internal/psc/inc_roots_test.go
- internal/psc/lsp.go
- internal/psc/lsp_command.go
- internal/psc/lsp_handler.go
- internal/psc/lsp_handler_test.go
- internal/psc/lsp_inference_test.go
- internal/psc/lsp_protocol.go
- internal/psc/lsp_test.go
- internal/psc/parse_command.go
- internal/psc/parse_command_test.go
- internal/psc/resolve_guard_test.go
- test/e2e/psc_check_test.go
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I26**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/psc/lsp.go:36,58` keeps one Session per server. `internal/parse/module.go:228-236` shares `seen` and `facts`. `module.go:302,312-313` sets `seen[module]` before loading and never clears it on not-found; `:385-389` does the same for a required file.
- **Confirmed by reading:**
  - A module missing at first open hits `seen` and takes the cycle branch (:295-298) on every later parse.
  - The concurrency specialist reproduced this with a `go test -overlay` (`scratchpad/review/conc/neg_test.go`).
- **Impact:**
  - Edits to a used module never reach the editor, so check and the LSP disagree.
  - A module created after the first open never resolves.
  - The map grows with every distinct directory list.
- **Fix:** use a fresh Session per OpenDocument, or invalidate `facts` and `seen` together. Also delete `seen[module]` on not-found.
- **Confidence:** High. **NEW.**
- **Issue:** 01a13e48-c9a5-79c4-b80b-636c1c9ec5be ("The LSP's parse Session never re-reads a module edited on disk"). This finding adds two things the issue's planned mtime fix cannot catch: the not-found case (there is no file to stat), and the trap where deleting `facts` but not `seen` turns a stale module into a missing one. It also adds the unbounded growth.
- **Found by:** concurrency #1, errors-A #6, errors-B #9, contract-A #9, contract-B #6.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] The LSP's long-lived parse.Session never re-reads a module, and caches "not found" forever is fixed (`go test ./internal/psc/ -run '^TestLSPSessionRereadsModules$' -count=1 -v | grep -q '^--- PASS: TestLSPSessionRereadsModules'`)
