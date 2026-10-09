---
title: Remove the WRONG pin when the harness measures the new parser
state: done
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
- 01a0b835-1270-7b32-9f0a-18fe7c555bae
- 01a11f08-6eb7-74a3-90bd-fb1231908e7d
- 01a11f08-6f62-7c58-9346-c385d77bbca7
- 01a11f08-6fee-7542-9b4a-81c9008bbdea
- 01a11f08-708a-738a-a613-e9f0e12ed4b7
- 01a11f08-712d-7ba2-8d92-fb518ba74db1
- 01a11f08-71bb-7ac6-9c87-0ccd0f498713
- 01a11f08-7251-77b4-a234-ba01fae75e9f
- 01a11f08-72a2-7eb3-8a76-cef1f3f49fc6
blocks:
- 01a11eb2-21e9-75a7-ae19-ce6f6a358be2
created: 2026-09-16T06:00:39.116126082Z
updated: 2026-10-09T12:33:33.672538509Z
sessions:
- start_sha: 528b9d835526d7a2c3acb41646c44cf6fd145437
  end_sha: 528b9d835526d7a2c3acb41646c44cf6fd145437
  commits: 0
  started_at: 2026-09-19T05:43:28.637346023Z
  ended_at: 2026-09-19T05:47:32.896618539Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T05:43:28.637346023Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T05:47:32.896618539Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T03:26:34.301077783Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T12:33:33.672538509Z
observed_paths:
- .github/workflows/ci.yml
- conformance/01_literals/README.md
- conformance/02_variables/README.md
- conformance/03_context/README.md
- conformance/04_operators/README.md
- conformance/05_scoping/README.md
- conformance/06_control/README.md
- conformance/07_subroutines/README.md
- conformance/08_references/README.md
- conformance/09_regex/README.md
- conformance/10_io/README.md
- conformance/11_oo/README.md
- conformance/12_packages/README.md
- conformance/13_opaque/README.md
- conformance/14_recursive/README.md
- conformance/GLOSSARY.md
- conformance/README.md
- conformance/mdtest/FORMAT.md
- conformance/mdtest/adjacency-01_literals.md
- conformance/mdtest/adjacency-02_variables.md
- conformance/mdtest/adjacency-03_context.md
- conformance/mdtest/adjacency-04_operators.md
- conformance/mdtest/adjacency-05_scoping.md
- conformance/mdtest/adjacency-06_control.md
- conformance/mdtest/adjacency-07_subroutines.md
- conformance/mdtest/adjacency-08_references.md
- conformance/mdtest/adjacency-09_regex.md
- conformance/mdtest/adjacency-10_io.md
- conformance/mdtest/adjacency-11_oo.md
- conformance/mdtest/adjacency-12_packages.md
- conformance/mdtest/adjacency-13_opaque.md
- conformance/mdtest/adjacency-14_recursive.md
- conformance/mdtest/aggregate-operators.md
- conformance/mdtest/aggregates-in-context.md
- conformance/mdtest/argument-extent.md
- conformance/mdtest/arguments.md
- conformance/mdtest/arithmetic.md
- conformance/mdtest/arrays.md
- conformance/mdtest/bitwise.md
- conformance/mdtest/bless-dispatch.md
- conformance/mdtest/blocks-and-terminations.md
- conformance/mdtest/class-feature.md
- conformance/mdtest/code-references.md
- conformance/mdtest/comparison.md
- conformance/mdtest/compile-phase.md
- conformance/mdtest/compound-assignment.md
- conformance/mdtest/conditionals.md
- conformance/mdtest/context-from-inside.md
- conformance/mdtest/declarations.md
- conformance/mdtest/dereferencing.md
- conformance/mdtest/dynamic-scope.md
- conformance/mdtest/embedded-code.md
- conformance/mdtest/eval-replacement.md
- conformance/mdtest/formats.md
- conformance/mdtest/handles.md
- conformance/mdtest/hashes.md
- conformance/mdtest/heredocs.md
- conformance/mdtest/jumps.md
- conformance/mdtest/list-operators-in-context.md
- conformance/mdtest/loading.md
- conformance/mdtest/loops.md
- conformance/mdtest/named-operators.md
- conformance/mdtest/names.md
- conformance/mdtest/numeric-point.md
- conformance/mdtest/numeric-radix.md
- conformance/mdtest/packages.md
- conformance/mdtest/pod-and-data.md
- conformance/mdtest/quote-like-terms.md
- conformance/mdtest/range.md
- conformance/mdtest/read-position.md
- conformance/mdtest/regex-delimiters.md
- conformance/mdtest/regex-matching.md
- conformance/mdtest/regex-pattern-positions.md
- conformance/mdtest/regex-quote-likes.md
- conformance/mdtest/strings.md
- conformance/mdtest/subroutines.md
- conformance/mdtest/symbol-table.md
- conformance/mdtest/taking-references.md
- conformance/mdtest/tie.md
- conformance/mdtest/unary.md
- conformance/mdtest/ungated.md
- conformance/mdtest/vstrings.md
- conformance/mdtest/writing.md
- docs/audits/2026-09-22-corpus-completeness.md
- docs/audits/2026-09-23-corpus-complete-chain-review.md
- docs/audits/2026-09-24-corpus-complete-assessment.md
- docs/audits/2026-09-24-corpus-complete-postmortem.md
- docs/plans/2026-09-05-parser-conformance-plan.md
- docs/plans/2026-09-21-cst-is-the-source-of-truth.md
- docs/plans/2026-09-21-deferred-chain-m1-m2.md
- docs/plans/2026-09-21-graded-conformance-corpus.md
- docs/plans/2026-09-26-perl-git-t-parses.md
- docs/rfcs/0001-type-libraries.md
- docs/specs/perl-parser/00-findings.md
- docs/specs/perl-parser/04-expressions.md
- internal/cli/help_test.go
- internal/conformance/allcases.go
- internal/conformance/categories.go
- internal/conformance/claimonce_test.go
- internal/conformance/conformance_test.go
- internal/conformance/croak.go
- internal/conformance/croak_test.go
- internal/conformance/fact.go
- internal/conformance/file.go
- internal/conformance/glossary_test.go
- internal/conformance/hardmarkers_test.go
- internal/conformance/lint.go
- internal/conformance/lint_test.go
- internal/conformance/namedconstructs_test.go
- internal/conformance/numbering_test.go
- internal/conformance/perl.go
- internal/conformance/prereq_test.go
- internal/conformance/ratchet.go
- internal/conformance/ratchet_test.go
- internal/conformance/regen_ratchet_test.go
- internal/conformance/run.go
- internal/conformance/subops_test.go
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier01_test.go
- internal/conformance/tier02_test.go
- internal/conformance/tier03_test.go
- internal/conformance/tier04_test.go
- internal/conformance/tier05_test.go
- internal/conformance/tier06_test.go
- internal/conformance/tier07_callforms_test.go
- internal/conformance/tier07_test.go
- internal/conformance/tier08_test.go
- internal/conformance/tier09_test.go
- internal/conformance/tier10_test.go
- internal/conformance/tier11_test.go
- internal/conformance/tier12_test.go
- internal/conformance/tier13_test.go
- internal/conformance/tier14_test.go
- internal/conformance/topic.go
- internal/conformance/topic_test.go
- internal/conformance/vacuous_test.go
- internal/cpan/mirror_manager_test.go
- internal/data/help/types.md
- internal/infer/core_test.go
- internal/infer/diagnostics.go
- internal/infer/diagnostics_test.go
- internal/infer/infer.go
- internal/infer/infer_test.go
- internal/infer/precision_test.go
- internal/infer/testdata/observe_types.pl
- internal/lexer/ampquotename_test.go
- internal/lexer/bracecomment_test.go
- internal/lexer/brackets.go
- internal/lexer/caret_test.go
- internal/lexer/codeblockheredoc_test.go
- internal/lexer/constsub_test.go
- internal/lexer/declapos_test.go
- internal/lexer/delim_test.go
- internal/lexer/digitvar_test.go
- internal/lexer/expect.go
- internal/lexer/expect_test.go
- internal/lexer/fhslot_test.go
- internal/lexer/formatargblock_test.go
- internal/lexer/formatstate_test.go
- internal/lexer/globquotename_test.go
- internal/lexer/heredoc.go
- internal/lexer/heredocbrace_test.go
- internal/lexer/ident.go
- internal/lexer/ident_test.go
- internal/lexer/interpheredoc_test.go
- internal/lexer/intuit.go
- internal/lexer/intuit_test.go
- internal/lexer/keyword.go
- internal/lexer/keywordapos_test.go
- internal/lexer/lastidx_test.go
- internal/lexer/leadingsep_test.go
- internal/lexer/letternumber_test.go
- internal/lexer/lexer.go
- internal/lexer/lexsubquote_test.go
- internal/lexer/mainstash_test.go
- internal/lexer/methodnameop_test.go
- internal/lexer/midstringheredoc_test.go
- internal/lexer/niladicsub_test.go
- internal/lexer/numberend_test.go
- internal/lexer/openblock_test.go
- internal/lexer/perlkeywords.go
- internal/lexer/pidvar_test.go
- internal/lexer/pod.go
- internal/lexer/pod_test.go
- internal/lexer/postderefstar_test.go
- internal/lexer/postfixlastidx_test.go
- internal/lexer/proto.go
- internal/lexer/proto_test.go
- internal/lexer/quote.go
- internal/lexer/quoteop_test.go
- internal/lexer/repeatcount_test.go
- internal/lexer/repetitionglued_test.go
- internal/lexer/replheredoc_test.go
- internal/lexer/scan.go
- internal/lexer/scanrows_test.go
- internal/lexer/sigilspace_test.go
- internal/lexer/smartmatch_test.go
- internal/lexer/spacedbrace_test.go
- internal/lexer/spacedderef_test.go
- internal/lexer/stringbitwise_test.go
- internal/lexer/symbolop_test.go
- internal/lexer/testdata/corpus.ratchet
- internal/lexer/testdata/lex.t.tokens
- internal/lexer/trivia_test.go
- internal/lexer/unidor_test.go
- internal/lexer/vnum_test.go
- internal/memory/lazy.go
- internal/parse/ampblock_test.go
- internal/parse/ampcall_test.go
- internal/parse/anonhead_test.go
- internal/parse/anyall_test.go
- internal/parse/argcutops_test.go
- internal/parse/arity.go
- internal/parse/arityrefusal_test.go
- internal/parse/arrowargs_test.go
- internal/parse/arrowhashread_test.go
- internal/parse/attrsig_test.go
- internal/parse/barederef_test.go
- internal/parse/barehandle_test.go
- internal/parse/barepattern_test.go
- internal/parse/barewordfilehandle_test.go
- internal/parse/barewordinvocant_test.go
- internal/parse/beginrequire_test.go
- internal/parse/brace_before_arrow_test.go
- internal/parse/bracketslot_test.go
- internal/parse/builderimport_test.go
- internal/parse/builtinimport_test.go
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/canon_test.go
- internal/parse/chain.go
- internal/parse/classword_test.go
- internal/parse/cmd/subject/main.go
- internal/parse/cmd/subject/main_test.go
- internal/parse/commaassign_test.go
- internal/parse/commafat_test.go
- internal/parse/commanamelist_test.go
- internal/parse/compiletoken_test.go
- internal/parse/condeclcanon_test.go
- internal/parse/control.go
- internal/parse/control_test.go
- internal/parse/corearray_test.go
- internal/parse/coredecl_test.go
- internal/parse/coredeleteselect_test.go
- internal/parse/coregrepmapsort_test.go
- internal/parse/coreinfix_test.go
- internal/parse/coreio_test.go
- internal/parse/corenoproto_test.go
- internal/parse/coreops_test.go
- internal/parse/coreprefix_test.go
- internal/parse/coreprocess_test.go
- internal/parse/coreshapes_test.go
- internal/parse/coresubs_test.go
- internal/parse/coresystem_test.go
- internal/parse/coretable_test.go
- internal/parse/coretypes_test.go
- internal/parse/decl.go
- internal/parse/decl_test.go
- internal/parse/declarations.go
- internal/parse/declarations/CORE.pmt
- internal/parse/declarations/Future/AsyncAwait.pmt
- internal/parse/declarations/Mojolicious/Lite.pmt
- internal/parse/declarations/Moose.pmt
- internal/parse/declarations/Moose/Role.pmt
- internal/parse/declarations_test.go
- internal/parse/declleft_test.go
- internal/parse/declparen_test.go
- internal/parse/declsplit_test.go
- internal/parse/declsubscript_test.go
- internal/parse/deferstmt_test.go
- internal/parse/derefblock_test.go
- internal/parse/derefcanon_test.go
- internal/parse/derefyada_test.go
- internal/parse/derive.go
- internal/parse/derive_test.go
- internal/parse/easy_test.go
- internal/parse/emptybraces_test.go
- internal/parse/emptycomma_test.go
- internal/parse/export_test.go
- internal/parse/exportdefault_test.go
- internal/parse/exports.go
- internal/parse/exports_test.go
- internal/parse/exporttags_test.go
- internal/parse/expr.go
- internal/parse/expr_test.go
- internal/parse/faithfulnested_test.go
- internal/parse/featuregated_test.go
- internal/parse/fidelity.go
- internal/parse/fidelity_test.go
- internal/parse/forbarevar_test.go
- internal/parse/forderef_test.go
- internal/parse/formany_test.go
- internal/parse/formateofcanon_test.go
- internal/parse/fuzz_test.go
- internal/parse/globalias_test.go
- internal/parse/globat_test.go
- internal/parse/globbracket_test.go
- internal/parse/globderef_test.go
- internal/parse/globname_test.go
- internal/parse/handlederef_test.go
- internal/parse/handleword_test.go
- internal/parse/hashderef_test.go
- internal/parse/hashhedge_test.go
- internal/parse/heredoc_test.go
- internal/parse/heredocarg_test.go
- internal/parse/indirect.go
- internal/parse/indirect_test.go
- internal/parse/indirectprog_test.go
- internal/parse/indirectvar_test.go
- internal/parse/infile_test.go
- internal/parse/interphedge_test.go
- internal/parse/interppkg_test.go
- internal/parse/invocant_test.go
- internal/parse/invocantparse_test.go
- internal/parse/isaindirect_test.go
- internal/parse/isolation_test.go
- internal/parse/keepcomma_test.go
- internal/parse/keyquote_test.go
- internal/parse/keyquoteop_test.go
- internal/parse/keyword.go
- internal/parse/keywordgap_test.go
- internal/parse/labelbeforeclose_test.go
- internal/parse/lastinblock_test.go
- internal/parse/lastindexblock_test.go
- internal/parse/leadingdecimal_measure_test.go
- internal/parse/leadingdecimal_test.go
- internal/parse/lexsub_test.go
- internal/parse/library_pmt_test.go
- internal/parse/listdeclattrs_test.go
- internal/parse/listopspacedparen_test.go
- internal/parse/loopexexpr_test.go
- internal/parse/lowlogical_test.go
- internal/parse/mapbrace_test.go
- internal/parse/methodgate_test.go
- internal/parse/module.go
- internal/parse/module_test.go
- internal/parse/mymethod_test.go
- internal/parse/namedparam_test.go
- internal/parse/nestedblock_test.go
- internal/parse/nestedsubfacts_test.go
- internal/parse/operandless_test.go
- internal/parse/operators.go
- internal/parse/parenhandle_test.go
- internal/parse/parenlessarity_test.go
- internal/parse/parse.go
- internal/parse/perlgitt_test.go
- internal/parse/perllines_test.go
- internal/parse/pin_test.go
- internal/parse/pkgversion_test.go
- internal/parse/pmt_test.go
- internal/parse/postfixhashslice_test.go
- internal/parse/postfixlastidx_test.go
- internal/parse/precedence.go
- internal/parse/precrelations_test.go
- internal/parse/printmod_test.go
- internal/parse/protoarity_test.go
- internal/parse/protoslot_test.go
- internal/parse/qrequire_test.go
- internal/parse/qualifiedmodsub_test.go
- internal/parse/rangeargs_test.go
- internal/parse/ratchet_test.go
- internal/parse/readline_test.go
- internal/parse/refaliasfor_test.go
- internal/parse/refscalar.go
- internal/parse/refscalar_test.go
- internal/parse/refusal.go
- internal/parse/refusal_test.go
- internal/parse/required_test.go
- internal/parse/requireexpr_test.go
- internal/parse/requirequote_test.go
- internal/parse/returnexpr_test.go
- internal/parse/returnlist_test.go
- internal/parse/selfdefault_test.go
- internal/parse/session_test.go
- internal/parse/shape.go
- internal/parse/shape_test.go
- internal/parse/sigcanon_test.go
- internal/parse/sigcomment_test.go
- internal/parse/sigemptydef_test.go
- internal/parse/sigplaceholder_test.go
- internal/parse/sites.go
- internal/parse/sites_decided_test.go
- internal/parse/smartmatchaggregate_test.go
- internal/parse/sortcmp_test.go
- internal/parse/specialsub_test.go
- internal/parse/stmt_test.go
- internal/parse/stmtmodcanon_test.go
- internal/parse/subattr_test.go
- internal/parse/subspelling_test.go
- internal/parse/subthenstmt_test.go
- internal/parse/switch_test.go
- internal/parse/syntax.go
- internal/parse/t2_test.go
- internal/parse/term.go
- internal/parse/testdata/declarations/My/Lib.pmt
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1corpus.pin
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/trailingcomma_test.go
- internal/parse/trycatch_test.go
- internal/parse/typedlexical_test.go
- internal/parse/unaryplus_test.go
- internal/parse/unclosed_test.go
- internal/parse/unseenexport_test.go
- internal/parse/use.go
- internal/parse/use_test.go
- internal/parse/useconstant_test.go
- internal/parse/useifboth_test.go
- internal/parse/voidparam_test.go
- internal/parse/whenmod_test.go
- internal/parse/wordblock_test.go
- internal/parse/wrapper_test.go
- internal/parse/wrongtree_test.go
- internal/parse/xs.go
- internal/parse/xs_test.go
- internal/parse/yada_test.go
- internal/parseoracle/adapter.go
- internal/parseoracle/adapter_test.go
- internal/parseoracle/compare.go
- internal/parseoracle/compare_facts.go
- internal/parseoracle/compare_test.go
- internal/parseoracle/cvwalk_test.go
- internal/parseoracle/degenerate_test.go
- internal/parseoracle/deref_test.go
- internal/parseoracle/fidelity_test.go
- internal/parseoracle/findings.go
- internal/parseoracle/findings_test.go
- internal/parseoracle/goparser_test.go
- internal/parseoracle/hedge_test.go
- internal/parseoracle/inject_test.go
- internal/parseoracle/marker_gaps_test.go
- internal/parseoracle/markers_test.go
- internal/parseoracle/oracle.go
- internal/parseoracle/oracle_test.go
- internal/parseoracle/parenless_test.go
- internal/parseoracle/ratchet.go
- internal/parseoracle/ratchet_test.go
- internal/parseoracle/run.go
- internal/parseoracle/run_test.go
- internal/parseoracle/span_test.go
- internal/parseoracle/subject.go
- internal/parseoracle/taxonomy.go
- internal/parseoracle/testdata/fixture/subbody.pl
- internal/parseoracle/testdata/fixture/wider.pl
- internal/parseoracle/testdata/parse_facts.pl
- internal/parseoracle/testdata/ratchet/baseline.txt
- internal/parseoracle/verdict.go
- internal/performance/analyzer_test.go
- internal/psc/check_command_test.go
- internal/pvm/command.go
- internal/types/coercion.go
- internal/types/coercion_test.go
- internal/types/declared.go
- internal/types/gotchas_test.go
- internal/types/invocant_test.go
- internal/types/lattice_oracle_test.go
- internal/types/names.go
- internal/types/names_test.go
- internal/types/narrowing.go
- internal/types/narrowing_test.go
- internal/types/paper_alignment_test.go
- internal/types/select.go
- internal/types/select_test.go
- internal/types/signatures.go
- internal/types/signatures_audit_test.go
- internal/types/signatures_test.go
- internal/types/testdata/lattice_oracle.pl
- internal/types/types.go
- internal/types/types_test.go
- internal/types/void_test.go
- scripts/archive-chain.py
- scripts/check-dangling.py
- tools/mdtest/.gitignore
- tools/mdtest/01_literals.py
- tools/mdtest/02_variables.py
- tools/mdtest/03_context.py
- tools/mdtest/05_scoping.py
- tools/mdtest/06_control.py
- tools/mdtest/07_subroutines.py
- tools/mdtest/08_references.py
- tools/mdtest/09_regex.py
- tools/mdtest/10_io.py
- tools/mdtest/11_oo.py
- tools/mdtest/12_packages.py
- tools/mdtest/13_opaque.py
- tools/mdtest/14_recursive.py
- tools/mdtest/__pycache__/gen.cpython-312.pyc
- tools/mdtest/extract.py
- tools/mdtest/gen.py
- tools/mdtest/tiersrc.py
---

Two files are pinned as WRONG in the corpus baseline, and the pin comes out
when they do.

    comp/our.t       metric 8
    op/universal.t   metric 16

Both are defects in the TREE-SITTER grammar, which is what the harness still
measures (`run.go:381` calls `parser.New()`). Findings §0.11 has the analysis
and `marker_gaps_test.go` has the regression tests.

## Why they are pinned rather than zero

Commit `a69d8f1c` introduced these and left `TestCorpusBaselineIsIntact`
tripped on purpose, as a signal. It stayed tripped, and a permanently-red
test is one nobody reads -- worse, a red suite cannot show the NEXT
regression. An absolute zero that has never held stops being a gate and
becomes scenery.

The pin names both PATHS rather than a count. A count of 2 would let one
defect be fixed and a different one appear with the gate silent; measured,
that swap passes a count and fails the pin.

## What removes it

Either:

1. **The M1 gate lands** and the harness measures `internal/parse` instead of
   the tree-sitter grammar. This is the expected route, and `comp/our.t`'s
   cause is already fixed there -- the M1 lexer reads `if (/TIE/)` as
   `Quote("/TIE/")`, not division. Verified.
2. **Someone fixes the tree-sitter grammar's two defects** directly. Real
   work on a parser M1-M6 exist to replace, so probably not worth it.

Either way the closing move is the same: delete the pin block in
`ratchet_test.go`, restore `!= 0`, re-baseline in the same commit, and update
findings §0.11.

If only ONE is fixed, replace the pin with the remaining path rather than
loosening it to a count.


Findings: §0.11. Introduced by a69d8f1c; pinned by the commit that closes this issue's parent.

## Acceptance criteria

- [ ] The WRONG pin is gone and the gate asserts zero (`go test ./internal/parseoracle/ -run '^TestCorpusBaselineIsIntact$' -count=1 -v | grep -q '^--- PASS: TestCorpusBaselineIsIntact'`)
- [ ] The baseline holds no WRONG rows (`grep -c '^WRONG' internal/parseoracle/testdata/ratchet/baseline.txt | grep -q '^0$'`)
- [ ] findings §0.11 records that the pin is gone (`grep -q 'WRONG pin removed' docs/specs/perl-parser/00-findings.md`)
