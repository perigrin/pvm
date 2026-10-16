---
title: "[I17] Exits and guards are matched by exact name: croak, CORE::die and CORE::defined are missed, and CORE::push g..."
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.760410032Z
updated: 2026-10-16T03:49:37.286262186Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: 133837eb651f8dcffb4203d072b3701753486869
  commits: 4
  started_at: 2026-10-16T01:17:18.033021959Z
  ended_at: 2026-10-16T03:49:37.286262186Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:18.033021959Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T03:49:37.286262186Z
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
- internal/lower/call.go
- internal/lower/node.go
- internal/lower/positions.go
- internal/parse/arity.go
- internal/parse/declarations/Carp.pmt
- internal/parse/loader_regular_test.go
- internal/parse/module.go
- internal/parse/parse.go
- internal/parse/staysinroots_test.go
- internal/parse/term.go
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

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I17**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:**
  - `internal/infer/flow.go:354-371` (blockExits: only die, exit and return).
  - `flow.go:425-437` (callGuard).
  - `calls.go:38` (`CoreBuiltin(n.Name)` with the `CORE::` prefix still on).
- **Repro:**
  - `v/c6b.pl` (`unless (CORE::defined $y) { die }`) and `v/c6c.pl` (`unless (defined $y) { CORE::die }`) each give a new psc error `got Undef|Int`.
  - `v/c6a.pl` (croak, with PERL5LIB set to perl's privlib) gives the same.
  - origin/pu is silent on all three, and perl prints 4.
  - `v/eb6b.pl`: `CORE::push %h, 1` is silent in psc, but perl refuses it at compile time.
- **Fix:** strip `CORE::` (and `CORE::GLOBAL::`) before matching builtins. Treat Carp's croak and confess as exits, or declare no-return in .pmt.
- **Confidence:** High. **NEW.** **Issue:** none. **Found by:** errors-B #6.


**Ruled 2026-10-16 (perigrin): "inference should be getting the parsed tree from the parser."** The parser already resolves each call -- imports, `use subs` and default-`@EXPORT` overrides, and library `.pmt` signatures (`parse.Imports`). Inference takes that resolution from the parse result through lowering (e.g. the lowered Call carries the parser's resolved signature) instead of re-deriving it by name in infer. croak is then an exit because Carp.pmt declares it returning None. Carp.pmt must declare all of Carp's exports (`@EXPORT` carp croak confess; `@EXPORT_OK` cluck verbose longmess shortmess), since a .pmt replaces the module's source; the m1 gates must hold. The same mechanism fixes S3 (01a14247-ddfd, overrides re-guessed). The CORE half (exits by CORE.pmt None, CORE:: names) landed in 7c2063f2.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Exits and guards are matched by exact name: `croak`, `CORE::die` and `CORE::defined` are missed, and `CORE::push` goes unchecked is fixed (`go test ./internal/infer/ -run '^TestExitsAndGuardsByCoreAndCarp$' -count=1 -v | grep -q '^--- PASS: TestExitsAndGuardsByCoreAndCarp'`)
