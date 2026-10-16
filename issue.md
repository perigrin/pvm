---
title: "[I10] my $x OP= EXPR loses OP in the parser; list and attribute declarations with compound ops lose their shape i..."
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.491801427Z
updated: 2026-10-16T03:56:38.15326351Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: e54d7d4d43d75f14738485d21503199143aee75b
  commits: 5
  started_at: 2026-10-16T01:17:16.004571748Z
  ended_at: 2026-10-16T03:56:38.15326351Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.004571748Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T03:56:38.15326351Z
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
- internal/infer/lowered_test.go
- internal/infer/ops.go
- internal/infer/ops_test.go
- internal/infer/testdata/lowered/ops/compound_assign.golden
- internal/infer/testdata/lowered/ops/defined_or.golden
- internal/infer/testdata/lowered/special/ref_arith.golden
- internal/infer/testdata/lowered/special/ref_paren.golden
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
- internal/lexer/keyword.go
- internal/lexer/lexer.go
- internal/lexer/scan.go
- internal/lexer/versionpadding_test.go
- internal/lower/call.go
- internal/lower/decl.go
- internal/lower/decl_test.go
- internal/lower/id.go
- internal/lower/node.go
- internal/lower/positions.go
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

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I10**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/parse/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:**
  - `internal/parse/decl.go:791-813` records only the initialiser.
  - `internal/lower/decl.go:31-75`: a non-`=` operator in list form puts an `Assign` into `Decl.Vars`, and the attribute path never sets `ListAssign`.
- **Repro:** `psc parse --tree lowered`:
  - `my $x -= 1;` gives `Decl Vars:$x Init:(NumLit 1)`, the same tree as `my $x = 1`. perl gives -1.
  - `my ($x) ||= 1;` gives `Decl Vars:(Assign Op="||=" ...)`.
  - `my ($x, $y) :shared = f();` gives `Decl Attrs=[:shared]` with no ListAssign, so declContext returns UnknownCtx where perl uses list context.
- **Fix:** record the operator on the Declaration, or build `[binary OP target init]`. In lowering, accept any assign operator in the one-child branch, and set ListAssign when the single Var is a Paren list.
- **Confidence:** High.
- **NEW or PRE-EXISTING:** the parser part is PRE-EXISTING, and this PR's lowering and inference are its first consumers. The lowering part is NEW.
- **Issue:** none. 01a110e9-e23e-7c99-b7db-39ffee6d567d ("Two declaration comma shapes still disagree with perl") is adjacent but different.
- **Found by:** logic-A #3, logic-A #8.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `my $x OP= EXPR` loses OP in the parser; list and attribute declarations with compound ops lose their shape in lowering is fixed (`go test ./internal/parse/ -run '^TestDeclCompoundOperatorKept$' -count=1 -v | grep -q '^--- PASS: TestDeclCompoundOperatorKept'`)
