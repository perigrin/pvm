---
title: A parenthesised and/or/xor refuses, and the bare form is a different program
state: done
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T04:10:38.190903837Z
updated: 2026-09-26T06:56:47.936555367Z
sessions:
- start_sha: 53de041e579c32a987c8caed904ff7895895e8f3
  end_sha: a29b8aa63bbcdb25e5e87de15eb1ff570e657180
  commits: 495
  started_at: 2026-09-26T04:43:11.609276216Z
  ended_at: 2026-09-26T05:32:11.538954039Z
- start_sha: a21b05dea5e6ec80fb8f1326a65685e6bbcd4087
  end_sha: d1537f9631443c467a73f3e56cf382da4eeab2ae
  commits: 3
  started_at: 2026-09-26T06:06:56.589504832Z
  ended_at: 2026-09-26T06:56:47.936555367Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T04:43:11.609276216Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T05:32:11.538954039Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-09-26T05:57:25.207924343Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T06:06:56.589504832Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T06:56:47.936555367Z
observed_paths:
- .github/scripts/parseoracle-pin.sh
- .github/workflows/ci.yml
- .github/workflows/parse-fidelity.yml
- .gitignore
- CONTRIBUTING.md
- Makefile
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
- docs/contributing/coding-conventions.md
- docs/contributing/development-workflow.md
- docs/plans/2026-09-04-perl-parser-decision.md
- docs/plans/2026-09-04-psc-parser-fix-options.md
- docs/plans/2026-09-04-psc-t-corpus-triage.md
- docs/plans/2026-09-04-triage-arity.md
- docs/plans/2026-09-04-triage-builtins.md
- docs/plans/2026-09-04-triage-numeric.md
- docs/plans/2026-09-04-triage-scanner.md
- docs/plans/2026-09-04-triage-string.md
- docs/plans/2026-09-05-parser-conformance-plan.md
- docs/plans/2026-09-05-parser-prior-art.md
- docs/plans/2026-09-06-fidelity-harness-milestone.md
- docs/plans/2026-09-21-cst-is-the-source-of-truth.md
- docs/plans/2026-09-21-deferred-chain-m1-m2.md
- docs/plans/2026-09-21-graded-conformance-corpus.md
- docs/plans/review/chain-alignment.md
- docs/plans/review/chain-pushback.md
- docs/specs/perl-parser/00-findings.md
- docs/specs/perl-parser/01-scope.md
- docs/specs/perl-parser/02-lexical-structure.md
- docs/specs/perl-parser/03-lexer-feedback.md
- docs/specs/perl-parser/04-expressions.md
- docs/specs/perl-parser/05-statements.md
- docs/specs/perl-parser/06-incremental-lsp.md
- docs/specs/perl-parser/07-conformance.md
- docs/specs/perl-parser/README.md
- docs/specs/perl-parser/review/citecheck.sh
- docs/specs/perl-parser/review/r1-ch02.md
- docs/specs/perl-parser/review/r1-ch04-ch06.md
- docs/specs/perl-parser/review/r1-ch05.md
- docs/specs/perl-parser/review/r1-cross.md
- docs/specs/perl-parser/review/r1-self.md
- docs/specs/perl-parser/review/r2-applied.md
- docs/specs/perl-parser/review/r2-coordinator-note.md
- docs/specs/perl-parser/review/r2-placement.md
- docs/specs/perl-parser/review/r3-ch03-audit.md
- docs/specs/perl-parser/review/r4-consumers.md
- docs/specs/perl-parser/review/r5-cites.md
- go.mod
- go.sum
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
- internal/infer/infer.go
- internal/infer/infer_test.go
- internal/infer/mutation_test.go
- internal/lexer/brackets.go
- internal/lexer/delim_test.go
- internal/lexer/expect.go
- internal/lexer/expect_test.go
- internal/lexer/fatcomma_test.go
- internal/lexer/fuzz_test.go
- internal/lexer/heredoc.go
- internal/lexer/heredoc_test.go
- internal/lexer/ident.go
- internal/lexer/ident_test.go
- internal/lexer/intuit.go
- internal/lexer/intuit_test.go
- internal/lexer/keyword.go
- internal/lexer/keyword_test.go
- internal/lexer/lexdott_test.go
- internal/lexer/lexer.go
- internal/lexer/lexer_test.go
- internal/lexer/pod.go
- internal/lexer/pod_test.go
- internal/lexer/proto.go
- internal/lexer/proto_test.go
- internal/lexer/quote.go
- internal/lexer/quote_test.go
- internal/lexer/quoteop_test.go
- internal/lexer/ratchet_test.go
- internal/lexer/scan.go
- internal/lexer/scanrows_test.go
- internal/lexer/testdata/corpus.ratchet
- internal/lexer/testdata/fuzz/FuzzLexer/fb95265c512141f4
- internal/lexer/testdata/lex.t.tokens
- internal/lexer/vnum_test.go
- internal/parse/barepattern_test.go
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/canon_test.go
- internal/parse/chain.go
- internal/parse/cmd/subject/main.go
- internal/parse/compiletoken_test.go
- internal/parse/control.go
- internal/parse/control_test.go
- internal/parse/decl.go
- internal/parse/decl_test.go
- internal/parse/declsubscript_test.go
- internal/parse/deref_test.go
- internal/parse/exports.go
- internal/parse/exports_test.go
- internal/parse/expr.go
- internal/parse/expr_test.go
- internal/parse/fidelity.go
- internal/parse/fidelity_test.go
- internal/parse/flags_test.go
- internal/parse/fuzz_test.go
- internal/parse/handleword_test.go
- internal/parse/hashderef_test.go
- internal/parse/hashhedge_test.go
- internal/parse/hedge_test.go
- internal/parse/keyquote_test.go
- internal/parse/keyword.go
- internal/parse/keyword_test.go
- internal/parse/keywordgap_test.go
- internal/parse/leadingdecimal_measure_test.go
- internal/parse/leadingdecimal_test.go
- internal/parse/lowlogical_test.go
- internal/parse/modifier_test.go
- internal/parse/module.go
- internal/parse/module_test.go
- internal/parse/parse.go
- internal/parse/parse_test.go
- internal/parse/pin_test.go
- internal/parse/precedence.go
- internal/parse/quoteop_test.go
- internal/parse/ratchet_test.go
- internal/parse/readline_test.go
- internal/parse/refusal.go
- internal/parse/refusal_test.go
- internal/parse/shape.go
- internal/parse/shape_test.go
- internal/parse/sites.go
- internal/parse/sites_decided_test.go
- internal/parse/stmt_test.go
- internal/parse/subscript_test.go
- internal/parse/t2_test.go
- internal/parse/term.go
- internal/parse/termform_test.go
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1corpus.pin
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/use.go
- internal/parse/use_test.go
- internal/parse/wrongtree_test.go
- internal/parseoracle/README.md
- internal/parseoracle/cache.go
- internal/parseoracle/cache_test.go
- internal/parseoracle/ci_test.go
- internal/parseoracle/cmd/receipt/main.go
- internal/parseoracle/compare_facts.go
- internal/parseoracle/context_test.go
- internal/parseoracle/corpus.go
- internal/parseoracle/corpus_test.go
- internal/parseoracle/cvwalk_test.go
- internal/parseoracle/deficit_test.go
- internal/parseoracle/findings.go
- internal/parseoracle/findings_test.go
- internal/parseoracle/gate_test.go
- internal/parseoracle/goparser_test.go
- internal/parseoracle/oracle.go
- internal/parseoracle/oracle_test.go
- internal/parseoracle/persite_test.go
- internal/parseoracle/population_test.go
- internal/parseoracle/procgroup_other.go
- internal/parseoracle/procgroup_unix.go
- internal/parseoracle/ratchet.go
- internal/parseoracle/ratchet_test.go
- internal/parseoracle/receipt.go
- internal/parseoracle/receipt_test.go
- internal/parseoracle/run.go
- internal/parseoracle/run_test.go
- internal/parseoracle/scratch_test.go
- internal/parseoracle/span_test.go
- internal/parseoracle/subject.go
- internal/parseoracle/subject_test.go
- internal/parseoracle/taint_test.go
- internal/parseoracle/testdata/corpus.pin
- internal/parseoracle/testdata/fixture/environmental.pl
- internal/parseoracle/testdata/fixture/exact.pl
- internal/parseoracle/testdata/fixture/explicit_ref.pl
- internal/parseoracle/testdata/fixture/noanswer.pl
- internal/parseoracle/testdata/fixture/subbody.pl
- internal/parseoracle/testdata/fixture/wider.pl
- internal/parseoracle/testdata/parse_facts.pl
- internal/parseoracle/testdata/perl_subject.pl
- internal/parseoracle/testdata/ratchet/baseline.txt
- internal/parseoracle/testdata/ratchet/fixture.txt
- internal/parseoracle/testdata/stderr/environmental.txt
- internal/parseoracle/testdata/stderr/other.txt
- internal/parseoracle/testdata/stderr/syntaxerror.txt
- internal/parseoracle/testdata/stderr/versionskew.txt
- internal/parseoracle/threads_pin_test.go
- internal/parseoracle/verdict.go
- internal/parser/degenerate.go
- internal/parser/degenerate_test.go
- internal/parser/deref_test.go
- internal/parser/parser.go
- internal/types/narrowing.go
- internal/types/narrowing_test.go
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

`($a and $b)` refuses with `not_a_term`, and so do `or` and `xor`. The
bare forms parse. The gap is the three LOW-PRECEDENCE word operators in a
parenthesised expression.

Filed by `crochet:chain-review` round 2, which found this case owned by
nobody: the glossary issue (`01a0dba9`) explicitly disclaims it, the
citation issue (`01a0d0c7`) never mentions it, and the issue that would
have covered it was cancelled on a false premise.

## Measured, at 226f38db

    bare   $a and $b       clean
    paren ($a and $b)      UNKNOWN   not_a_term
    bare   $a or $b        clean
    paren ($a or $b)       UNKNOWN   not_a_term
    bare   $a xor $b       clean
    paren ($a xor $b)      UNKNOWN   not_a_term
    paren ($a cmp $b)      clean
    paren ($a x 3)         clean
    paren, no assign: print(($a and 1), "\n")   UNKNOWN

`not $a` parenthesised is CLEAN -- it is prefix, not infix, so it is not
part of this. `cmp` and `x` are clean because they bind tighter than `=`
and the parser reaches them by a different route; only the three operators
BELOW assignment in precedence fail.

So this is not "word-shaped operators are unimplemented". It is narrower:
a word operator at perlop's lowest levels is not accepted as an infix
operator inside a parenthesised term.

## Why the parenthesised form is the one that matters

perl, 5.42.0 -- the paren changes the PROGRAM, not just the style:

    $ perl -MO=Deparse -e 'my $a=1; my $b=0; my $y = ($a and $b);'
    my $y = $a && $b;

    $ perl -MO=Deparse -e 'my $a=1; my $b=0; my $y = $a and $b;'
    $b if my $y = $a;

Bare `and` binds looser than `=`, so the assignment swallows `$a` and the
statement becomes a modifier on `$b`. Parenthesised, it is an ordinary
`&&`. The corpus case `comparison.md`'s "`&&` and `and` are one op at two
precedences" exists to document exactly that cliff, and we parse the half
that regroups while refusing the half that does not.

## What refuses today

Three corpus cases, all `refuses: unfiled` with `refusal: not_a_term`
where recorded:

    comparison.md   The five logical operators          ($a xor $b)
    comparison.md   `&&` and `and` are one op at two
                    precedences                         see note
    adjacency-04_operators.md   The whole tier in one body

The middle case is NOT this issue's: its source has a bare `$a and 9`,
which parses clean, and its only failing claim is the token fact `one
operator whose text is "and"`. That is `01a0dba9`'s. It is listed here so
a reader does not mistake it for part of this gap -- measured,
`comparison.md` line 140 is that case and line 111 is this one, different
cases in one file, so the two issues can land in either order.

Chain-review round 2 added a `blocked_by` edge here and then removed it:
the shared FILE looked like a conflict and the measurement said different
lines. Recorded because the same wrong edge was added to `01a0d0c7` for the
same reason.

## Context

- paths: `internal/parse/expr.go`, `internal/parse/call.go`, `conformance/mdtest/comparison.md`
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/conformance/ -count=1`
- entrypoints: `internal/parse/precedence.go:59-61` (`or`/`xor` at level 4, `and` at 5) and `internal/parse/call.go:343` (`startsTerm`)

## What is already right, so the fix is narrower than it looks

Measured before filing, so the work does not re-litigate settled code:

  - `and`, `or` and `xor` ARE in the precedence table already
    (`precedence.go:59-61`), at perlop's levels 4 and 5, with `xor`
    correctly sharing `or`'s level and a `toke.c:9224` citation.
  - `startsTerm` ALREADY knows a Word that is an infix operator does not
    start a term (`call.go:347`, changed for this exact shape in
    `01a0cf3e`).

So both halves of the knowledge exist. What fails is reaching them from
inside a parenthesised term: the paren's sub-expression is not parsed down
to level 4/5. That is the thing to measure first.

## Do not

Do not fix this by adding `and`/`or`/`xor` to the term-starter set. The
`01a0cf3e` change went the other way deliberately -- a Word that is an
operator is NOT a term -- and reversing it would reintroduce the bug that
change fixed (`print FOO eq "x" ? ...`).

### Review Findings

TIER_2 PAAD gate, 2026-09-26. Verdict IN-SCOPE FINDINGS. Two of four are
in scope; I re-measured both before reopening rather than taking them.

**1. HIGH -- `parseBracketed` is the same loop and did not get the fix.**

`internal/parse/term.go`'s `parseBracketed` (serving `[...]` and `{...}`)
is a structural twin of the fixed `parseParenList`: same element loop at
`p.parseExpr(infix[","].BP)`, same `default:` arm that bails with the
operator unconsumed. Measured at a29b8aa6:

    my $y = ($a and $b);      Unknown=0   <- fixed
    my $y = [$a and $b];      Unknown=1
    my $y = {$a and $b};      Unknown=1
    my $y = [1, 2 and 3];     Unknown=1
    my $y = [$a && $b];       Unknown=0   <- punctuation fine

All legal perl: `perl -MO=Deparse -e 'my ($a,$b); my $y = [$a and $b];'`
is `my $y = [$a && $b];`.

Worse than the original symptom, per the gate: the operand escapes the
bracket, so the tree claims `and` is ABOVE the declaration with a
one-element arrayref as its left operand -- not a parse of the source.

CLAUDE.md is explicit that this is the shape of fix to reject: "the lazy
fix IS the root-cause fix: one guard in the shared function is a smaller
diff than a guard in every caller -- and patching only the path the ticket
names leaves every sibling caller still broken." Two copies of one element
loop, one fixed, is the finding.

**2. MEDIUM -- the comma case parses but with a tree perl disagrees with.**

`TestBareWordInfixStillRegroups` asserts only `!containsKind(Unknown)` for
`my $x = 1, my $y = 2`. Measured shape:

    my $x = 1, my $y = 2;   ( ( (my $x (, 1 (my $y 2)))))   ours
    $x = 1, $y = 2;         ( ( (, (= $x 1) (= $y 2))))      correct

perl puts the comma ABOVE both: `-MO=Concise` shows `ex-list` with two
sibling `padsv_store` ops. Ours nests `, my $y = 2` inside `$x`'s
initialiser, and it compounds with a third declarator. The plain-assignment
form is already right, so the declarator path is the outlier.

`23805082`'s message says that shape "went clean again". It went
NON-UNKNOWN, with a wrong tree -- which is the same
tree-that-round-trips-but-is-not-a-parse class the commit itself names.

**3-4. Out of scope, recorded rather than fixed here.** `not` is only
incidentally correct -- structurally safe, since `not` is in `prefix` and
both guards look up `infix`, but nothing asserts it. And three shapes the
design rests on are unasserted: `($a xor $b xor $c)`, `($a and $b or $c)`,
and `(1, 2 and 3)` -- the last being the measurement quoted in the commit
message, the code comment and the `parseInfix` doc comment as the whole
justification, with no test behind it.

## Scope of this reopen

1. Fix `parseBracketed` by the SHARED route, not a third copy of the guard.
   The gate notes two copies were defensible and three earn a helper.
2. Assert the shape for `my $x = 1, my $y = 2`, and either fix the nesting
   to match perl or record the known-wrong shape explicitly. A bare
   `!Unknown` must not read as a shape guarantee.
3. Add `(1, 2 and 3)` as an assertion, since it is the design's own
   evidence.

`not` (finding 3) may stay unasserted or gain one case; it is not a
blocker either way.

## Criteria: which are progress and which are guards

Four fail today and are progress checks; two pass today and are labelled
guards. The T2 ratchet is the interesting guard: `TestT2CoreParses` fails
in BOTH directions, so if this fix makes real-world files parse the
shortfall map must be updated in the same commit -- the guard catches the
fix succeeding as well as regressing.

## Acceptance criteria

- [ ] The parenthesised low-precedence word operators parse (`go test ./internal/parse/ -run '^TestParenthesisedWordInfixParses$' -count=1 -v | grep -q '^--- PASS: TestParenthesisedWordInfixParses'`)
- [ ] The bare forms still parse, so the fix did not trade one shape for the other (`go test ./internal/parse/ -run '^TestBareWordInfixStillRegroups$' -count=1 -v | grep -q '^--- PASS: TestBareWordInfixStillRegroups'`)
- [ ] The five-logical-operators case passes rather than skipping (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/comparison.md/The_five_logical_operators'`)
- [ ] REGRESSION GUARD, passes today: `not` is still prefix and the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] A bracketed low-precedence word operator parses (`go test ./internal/parse/ -run '^TestBracketedWordInfixParses$' -count=1 -v | grep -q '^--- PASS: TestBracketedWordInfixParses'`)
- [ ] The comma case's SHAPE is asserted, not just its Unknown count (`go test ./internal/parse/ -run '^TestBareWordInfixStillRegroups$' -count=1 -v | grep -q '^--- PASS: TestBareWordInfixStillRegroups'`)
- [ ] The design's own measurement is a test (`grep -q '1, 2 and 3' internal/parse/lowlogical_test.go`)
- [ ] REGRESSION GUARD, passes today: the T2 shortfall ratchet still agrees in both directions (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] The suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
