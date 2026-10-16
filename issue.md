---
title: "[I2] A sub's assignment to an outer variable replaces that variable's type where the sub is declared"
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.174295834Z
updated: 2026-10-16T11:47:19.099832845Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: 9ff9be88b6f25ac747648420c2ff2433c1b55456
  commits: 6
  started_at: 2026-10-16T01:17:17.233388527Z
  ended_at: 2026-10-16T11:47:19.099832845Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:17.233388527Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T11:47:19.099832845Z
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

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I2**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/infer.go:264-273`. `written` is applied with `UpdateType(name, typ)`, which overwrites rather than joins.
- **Repro:** `logicB/w2.pl`:
  ```perl
  my $p = 5; sub reset_p { $p = undef } print $p + 1
  ```
  - New psc: `error: ... got Undef`.
  - origin/pu: silent.
  - perl prints 6.
- **Fix:** `UpdateType(name, types.Join(outer, typ))`, or leave the outer type to widenWritten.
- **Confidence:** High. **NEW.** **Issue:** none covers it directly; 01a132ab is related. **Found by:** logic-B #2.


**Decided 2026-10-16 (option c, widen).** Three options were measured (scratchpad flowA report): (a) join outer with what the sub writes adds a false positive for `$p=undef; sub s{$p=5} s(); $p+1`; (b) no write-back adds one too; (c) widening a variable any sub writes to its sigil type -- as `widenWritten` already does for a sub that reads it -- adds no false positive and fixes the repro, at the cost of the "reset sub was called" true positive, which needs call-site effects (filed separately). perl decides each case by whether the sub ran, which PSC cannot see today; (c) is the sound approximation.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] A sub's assignment to an outer variable replaces that variable's type where the sub is declared is fixed (`go test ./internal/infer/ -run '^TestSubAssignmentWidensOuterType$' -count=1 -v | grep -q '^--- PASS: TestSubAssignmentWidensOuterType'`)
