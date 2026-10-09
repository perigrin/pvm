---
title: A brace before -> is a hashref, and two anonhash cases the tree misreads
state: done
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0badc-a093-7e09-bdfb-002f4a73759f
- 01a0b835-1270-7b32-9f0a-18fe7c555bae
created: 2026-09-19T18:17:07.068073736Z
updated: 2026-10-09T09:41:35.621009881Z
sessions:
- start_sha: 25483d2b5bfff159cfc023dec9b050ce0c57c4b3
  end_sha: 25483d2b5bfff159cfc023dec9b050ce0c57c4b3
  commits: 0
  started_at: 2026-09-19T21:11:22.140962055Z
  ended_at: 2026-09-19T21:12:55.13834552Z
- start_sha: d2f26fe30e0f079d575302ad99d23d551a83d885
  end_sha: 069713648f3530ffaeb1156ca8f96cbcd64494c4
  commits: 3
  started_at: 2026-10-09T09:12:50.992886196Z
  ended_at: 2026-10-09T09:41:35.621009881Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T21:11:22.140962055Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T21:12:55.13834552Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T03:26:34.37807385Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T09:12:50.992886196Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T09:41:35.621009881Z
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
- internal/parse/attrsig_test.go
- internal/parse/barederef_test.go
- internal/parse/barehandle_test.go
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
- internal/parse/interppkg_test.go
- internal/parse/invocant_test.go
- internal/parse/invocantparse_test.go
- internal/parse/isaindirect_test.go
- internal/parse/isolation_test.go
- internal/parse/keepcomma_test.go
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

35 WRONG anonhash sites, and most are span attribution -- but three are real
and each is the "is `{` a block or a hashref" question §7.1.2 names.

## The real cases

**`grep {a=>$_}->{a}`** (op/grep.t:105). The `->` makes the brace a HASHREF,
not grep's block. Measured:

    $ perl -MO=Concise,-exec -e 'my @res = grep {a=>$_}->{a}, ("x");' | grep -c anonhash
    1

**`use overload '%{}' => sub { +{} }`** (op/coreamp.t:23) and
**`'${}' => sub { \my $x }`**. `+{}` is the disambiguating unary plus that
forces a hashref reading, and perl emits NO anonhash for it:

    $ perl -MO=Concise,-exec -e 'my $x = sub { +{} };' | grep -c anonhash
    0

So the subject reporting one there is over-reporting, the opposite direction
from the grep case. Both are the same ambiguity decided differently.

**`\state %y = {1,2}`** (op/lvref.t:370) and **`bless {`** (op/sort.t:944,
op/switch.t:1228) need measuring individually.

## The rest is span attribution

    op/sprintf2.t:347           x8   "my @tests = ("
    mro/package_aliases.t       x6   "for("
    mro/package_aliases_utf8.t  x6   "for("
    op/bop.t:693                x3   "my %op_pairs = ("
    op/pos.t:138                x2   "my %expected = ("

A hash literal assigned to a hash emits no anonhash at all:

    $ perl -MO=Concise,-exec -e 'my %op_pairs = ( a => 1 );' | grep -c anonhash
    0

so these are the first line of a multi-line statement whose ops belong
elsewhere -- 01a0badc's shape. Fix that first and re-measure before touching
this; the residue is what this issue is actually about.

Found by 01a0ba79's ranked table. Plan: M1. Spec: §7.1.2, §4.9.2.


## Re-measured after 01a0badc, 2026-09-19: 35 sites -> 2

This issue was filed at 35 anonhash sites and blocked on the span fix. That
fix landed (09e38fa9) and took it to 2. The span attribution WAS almost all
of it, as this issue predicted.

The two that remain:

    op/coreamp.t:23   use overload '%{}' => sub { +{} }
    op/undef.t:200    my $h= { baz => 1 }; my @k= keys %{($h=undef)||{}};

## What I established

**coreamp.t is a genuine over-report.** `+{}` is the disambiguating unary
plus, and perl emits NO anonhash for it:

    $ perl -MO=Concise,-exec -e 'my $x = sub { +{} };' | grep -c anonhash
    0

We report one. That is a claim perl contradicts, and it is the whole of
this site.

**undef.t is not what I assumed and I did not finish it.** I expected perl
to fold `{ baz => 1 }`, but measuring says otherwise:

    my $h = { baz => 1 };          1 anonhash
    my $h = {};                    0 anonhash, 1 emptyavhv/ANONHASH
    my $h = {}; my $j = {a=>1};    1 anonhash  (the empty one is emptyavhv)

and parse_facts.pl:142 folds emptyavhv-flagged-ANONHASH back into the
marker, so perl's total for that line should be 2 -- which is what we
report. The count agrees and the site is still WRONG, so the disagreement is
about WHICH STATEMENT owns one of them, not how many exist.

That needs the oracle's own line list for the file, which I did not run.
Paused there rather than guessing: two sites does not justify more
speculation, and a wrong diagnosis committed to an issue is worse than an
open question.

## Recommendation

Worth doing as a pair with whoever next touches anonhash, or folding into
the "hedge versus detect" decision 01a0bae1-b76f raises -- `+{}` may be
better answered by declining than by adding a special case for the unary
plus.

### Re-measured (2026-10-09, at d2f26fe3)

The oracle now measures internal/parse (01a0b835), and this issue's cases have moved:

- op/grep.t is `exact`: `grep {a=>$_}->{a}` already reads as perl does (one anonhash: measured with Concise on 5.42).
- The `+{}` diagnosis was wrong. perl DOES emit an anonymous hash for `+{}` (`my $x = sub { +{} };` measures one anonhash/emptyavhv site), and the op/coreamp.t:23 site it cited was the oracle counting a cloned anon sub twice -- fixed by 01a11f08-6fee. The criterion that `+{}` reports no site contradicts perl and is replaced.
- `TestGoParserWrongIsZero` only runs with `PARSEORACLE_SHIM` set, so as a criterion verify sees it skip; the corpus ratchet's pin (`TestCorpusBaselineIsIntact`) is the WRONG gate now, and 01a0a8cd removes it.

What is left is to hold the two brace readings with a test, so they cannot regress.

## Steps

- [ ] Write `TestBraceBeforeArrowIsAHashref`: `grep {a=>$_}->{a}` reports one anonhash site and `+{}` reports one, each as perl 5.42 does (measure with Concise first); if it passes at once, show it is not vacuous by breaking the reading and reverting; `make test`; commit

## Acceptance criteria

- [ ] a brace before `->`, and `+{}`, each report the anonymous hash perl compiles (`go test ./internal/parse/ -run '^TestBraceBeforeArrowIsAHashref$' -count=1 -v | grep -q '^--- PASS: TestBraceBeforeArrowIsAHashref'`)
