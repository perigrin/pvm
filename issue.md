---
title: "[I22] The term-position * split refuses or misparses ** and *= after __LINE__, do {}, eval {} and unresolved cons..."
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.994388424Z
updated: 2026-10-16T03:56:40.188807577Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: e54d7d4d43d75f14738485d21503199143aee75b
  commits: 5
  started_at: 2026-10-16T01:17:16.066990383Z
  ended_at: 2026-10-16T03:56:40.188807577Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.066990383Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T03:56:40.188807577Z
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

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I22**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/parse/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/lexer/scan.go:1105-1114`, the new term-position star added in this PR (the diff confirms it).
- **Repro:**
  - `my $y = __LINE__ ** 2;` lowers to `Stmts: Hole`.
  - `my $y = do {1} ** 2;` lowers to `Hole`.
  - In `v/pi.pl`, `use Math::Trig; my $y = pi ** 2;` lowers to `Hole` when `pi` is unresolved, which is common because psc parse has no @INC and psc check often has an empty one.
  - `print *STDOUT **2;` lowers to `Binary * (Var *STDOUT) (Var *2)`.
  - perl Deparse gives `(my($y) = 1.0)`, `do {1} ** 2`, and `print((*STDOUT ** 2))`.
- **Fix:**
  - Put `__LINE__`, `__FILE__`, `__PACKAGE__` and `__SUB__` in the term-ending table.
  - Set XOperator after the `}` of `do`/`eval BLOCK`.
  - Keep `**` as one token after a Word.
- **Confidence:** High for the behaviour. For NEW-ness, the lexer diff shows the branch is new; I did not rebuild origin/pu's parse to compare canon, so I relied on logic-A's comparison. **NEW.** **Issue:** none. **Found by:** logic-A #6.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] The term-position `*` split refuses or misparses `**` and `*=` after `__LINE__`, `do {}`, `eval {}` and unresolved constants is fixed (`go test ./internal/parse/ -run '^TestPowerAfterNiladicTerm$' -count=1 -v | grep -q '^--- PASS: TestPowerAfterNiladicTerm'`)
