# Graded conformance corpus: completeness audit

**Date:** 2026-09-22
**Auditor:** read-only audit agent
**Authority:** `docs/plans/2026-09-21-graded-conformance-corpus.md` (1050 lines)
**Milestone state at audit:** `m3-conformance-corpus` 35/35 done
**Tree state:** branch `feat/m1-parse-skeleton`, HEAD `2a2e9e55`, working tree clean
**Suite state:** `go test ./internal/conformance/ -count=1 -v` — PASS, 160.6s,
155 top-level PASS, 0 FAIL, 26 SKIP

## Verdict

The corpus is **substantially finished and of unusually high quality**. The
three ordering checks are real, mutation-tested, and gated over the live
corpus. Fourteen tiers exist with READMEs, adjacency files, and per-tier
tests. The format, glossary, ratchet, refusal codes and croak splitter all
exist and all run.

It is **not finished in substance** on three counts, one of which is the
exact defect class the brief asked me to hunt:

- **One tier is half-built against its own acceptance criterion** (Finding 1,
  BLOCKING). Tier 01 is chartered by the plan as "numbers, strings, quoting,
  `qw`". It has thirteen files, all numeric. Its acceptance criterion names
  string boundaries explicitly, and the gate that verifies it passes only
  because the tier's *adjacency* file happens to contain `'plain'`. This is
  the tier-06 `die`/`exit` bug shape, one layer up: a gate satisfied by
  incidental data rather than by the construct it names.
- **Three tier READMEs state refusal counts that are false** (Finding 2, GAP).
  Nothing machine-reads them, so they drifted silently as bugs were fixed.
- **One skip is stale** (Finding 3, GAP). `glossary_test.go:101` skips the
  leading-decimal boundary citing a lexer bug that commit `a68ac5f9` fixed.
  The corpus format fails loudly on exactly this; the glossary test does not.

Two smaller items are recorded as NOTE.

## Oracles used

Named per the audit protocol, because several questions here have no external
oracle and the distinction matters for how much weight each finding carries.

| # | Oracle | Type | Used for |
|---|--------|------|----------|
| O1 | `docs/plans/2026-09-21-graded-conformance-corpus.md` | Plan oracle | Every commitment-coverage claim |
| O2 | `git-zhi issue show <id>` acceptance criteria | Plan oracle | Finding 1 (tier 01 criterion names strings) |
| O3 | `conformance/GLOSSARY.md` | Documented-spec oracle | Finding 1 (string literal / quote-like operator are defined categories with measured boundaries) |
| O4 | `internal/conformance/testdata/corpus.ratchet` | Ground-truth artifact | Findings 2, 3 (which files actually refuse) |
| O5 | Anchored count of `^#\s*STATUS refuses` | Internal-invariant | Finding 2 (README prose vs file markers) |
| O6 | perl 5.42.0 at the resolved `perlPath` | External oracle | Ran under the suite; not re-derived by me |

**No oracle exists** for "is the corpus's *coverage* of Perl complete" — the
plan deliberately removes file counts and schedules, and says so at lines
59-74. Findings 1 and 4 are therefore argued against the plan's own tier
charter and against the issues' acceptance criteria (O1, O2), not against an
invented completeness standard.

**Anti-pattern 6 guard:** the suite passing is not evidence of correctness
anywhere below. Every finding is argued plan-vs-code or artifact-vs-artifact.

---

## Confirmed findings

### Finding 1: Tier 01 ships the numeric half of its charter; the string,
### quoting and `qw` half does not exist, and its gate passes on the
### adjacency file's incidental data — BLOCKING

**What the plan commits to.**

Plan line 124, "The tiers":

    01_literals/       numbers, strings, quoting, qw

The tier's own README repeats the charter verbatim at
`conformance/01_literals/README.md:3`: "Numbers, strings, quoting, `qw`."

The issue's acceptance criterion (O2, `git-zhi issue show 01a0c360`) is
sharper still:

> - [ ] Tier 01 covers the numeric **and string literal** boundaries the
>   glossary names, including v-strings and the -1-is-two-tokens case
>   (`go test ./internal/conformance/ -run '^TestTierLiteralsCoversGlossary$' ...`)

`conformance/GLOSSARY.md:86-107` (O3) defines `string literal` as its own
category with measured boundaries, and `:141-184` defines `quote-like
operator` as a *separate* category, stating the reason explicitly:

> **This is a separate category from `string literal` because the two make
> different claims.** A corpus file asserting `one string literal whose text
> is "hi"` must not be satisfied by `qw(hi)`, which is not a string at all.

**What actually exists.**

All thirteen tier-01 files are numeric. Every construct file declares
`# INTRODUCES numeric literal`:

```
$ grep -h '^# INTRODUCES' conformance/01_literals/*.t | sort | uniq -c
     12 # INTRODUCES numeric literal
      1 # INTRODUCES nothing of its own      <- 00_adjacency.t
```

No file in the corpus — any tier — is named for or introduces a string
literal, `q`, `qq`, `qw`, or an escape sequence:

```
$ ls conformance/*/*.t | grep -iE 'quot|string|escape|_q[qwx]?_|interpol'
conformance/03_context/02_interpolated_array.t
conformance/04_operators/02_string.t
conformance/04_operators/04_string_comparison.t
conformance/09_regex/05_interpolated.t
conformance/13_opaque/01_heredoc_interpolating.t
conformance/13_opaque/14_qq_delimiters.t
conformance/14_recursive/06_eval_string.t
```

Every one of those is a *later* tier using a string as a fixture for a
different subject. None introduces the string literal as a construct.

**Trigger / minimal failing case — the gate is self-satisfying.**

`internal/conformance/tier01_test.go:82-100` lists fifteen boundaries. I
mapped each to the file(s) supplying it:

```
= 42;                  -> 00_adjacency.t                        <- ONLY
= 0.5;                 -> 00_adjacency.t 02_decimal.t
= .5;                  -> 00_adjacency.t 04_leading_decimal.t
= 5e-1;                -> 00_adjacency.t 08_signed_exponent.t
= 4_294_967_296;       -> 00_adjacency.t 10_underscore_separators.t
= 0xff;                -> 00_adjacency.t 03_hexadecimal.t
= 0b1010;              -> 00_adjacency.t 01_binary.t
= 0377;                -> 00_adjacency.t 06_octal_leading_zero.t
= 0o377;               -> 00_adjacency.t 07_octal_prefix.t
= 1.;                  -> 00_adjacency.t 09_trailing_decimal.t
= -1;                  -> 00_adjacency.t 05_negative.t
= v65.66.67;           -> 00_adjacency.t 12_vstring_v.t
= 65.66.67;            -> 00_adjacency.t 11_vstring_bare.t
= 'plain';             -> 00_adjacency.t                        <- ONLY
= "                    -> 00_adjacency.t                        <- ONLY
```

Twelve boundaries have a dedicated construct file. Three do not, and two of
those three are exactly the string boundaries the acceptance criterion names.
They are satisfied by two lines of `00_adjacency.t`:

```perl
my $sq = 'plain';
my $dq = "$int-$dec";
```

The check is `strings.Contains` over the *joined* sources of the whole tier
(`tier01_test.go:105-112`), so the adjacency file — whose declared role is
`# INTRODUCES nothing of its own` — satisfies a criterion about what the tier
introduces. `qw` is in the same position: it appears in the corpus only inside
`00_adjacency.t`'s print statement.

**Why no probe mutation is attached.** I attempted to delete the two string
bindings from `00_adjacency.t` to demonstrate `TestTierLiteralsCoversGlossary`
failing, and the sandbox correctly refused the write — this is a read-only
audit and `conformance/` is off limits. The static evidence above is
sufficient and does not depend on the mutation: the gate is a substring search
over a set of files, and exactly one file in that set contains the two
substrings. Remove that file's two lines and the two boundaries have no other
supplier. A remediation phase can confirm by mutation in one command.

**Consequences that are not just bookkeeping.**

1. **No token facts for strings or quoting exist in tier 01.**
   `00_adjacency.t` has no `--- expect tokens` section at all. The plan's
   whole argument for the token layer (lines 341-404, 706-740) is that a
   lexical fact is the *only* check that sees grouping. Tier 01 therefore
   makes no lexical claim about `'plain'`, `"..."`, or `qw(a b c)`.
2. **The third source layer is weakened where the plan leans hardest on it.**
   Plan lines 880-886 assign the token stream the job of classifying tiers
   01-03 "whose constructs are lexical and where a declaration would be
   checking itself." A tier-01 string construct with no token fact is
   precisely the case that layer exists for.
3. **The glossary's sharpest string boundaries are unexercised.** `q{a{b}c}`
   is documented at `GLOSSARY.md:176-179` as nesting to the five-character
   string `a{b}c`; nothing in the corpus asserts it.

**Not a parser bug.** `go test ./internal/lexer/ -count=1` passes, and the
lexer has dedicated quote/delimiter tests (`internal/lexer/quote_test.go`,
`delim_test.go`). This is a *corpus* coverage gap, not a broken lexer. That
makes it cheaper to fix and does not reduce its severity against the plan.

**Site count and locations:**
- `conformance/01_literals/` — 13 files, 0 covering strings/quoting/`qw`
- `internal/conformance/tier01_test.go:97-99` — the three self-satisfied boundaries
- `conformance/01_literals/README.md:3` vs `:36` — charter says four subjects, body says "the two literal categories this tier owns"
- `conformance/01_literals/00_adjacency.t` — sole supplier of `'plain'`, `"`, `qw`

**Layer responsible:** corpus content + the tier-01 coverage gate.

**Suggested remediation shape:** add construct files for the string and
quoting half — single-quoted, double-quoted with an escape, `q`, `qq`, `qw`,
and the nesting-delimiter boundary `q{a{b}c}` — each carrying `--- expect
tokens` in the `string literal` / `quote-like operator` vocabulary. Separately,
make `TestTierLiteralsCoversGlossary` exclude `00_adjacency.t` from the joined
source, so a boundary must be supplied by a file that introduces it. That
second change is the durable one: it converts the gate from "appears somewhere
in the tier" to "some construct file introduces it", which is what the criterion
means.

**Side effects of the fix:**
- Renumbering. The plan makes tier-01 numbering derived
  (`TestTierNumberingRegenerates`, `TestDerivedTierNumberingRegenerates`), so
  inserting `q`/`qq`/`qw` files renumbers the tier alphabetically. That is the
  regeneration working, but it moves ~8 file names and the ratchet is
  file-name-keyed with prefixes stripped, so entries should survive.
- `00_adjacency.t` must gain the new constructs to keep
  `TestTierLiteralsAdjacency` passing, and its `--- expect output` re-pinned.
- Tier 01's `## INTRODUCES` op set may need `const` only, or may gain ops if a
  `qw` file emits differently; `TestCorpusLints`' reverse direction (a README
  claiming an op no file emits) will catch an over-claim.
- Excluding the adjacency file from the coverage gate may surface the same
  latent gap in `TestTierVariablesCoversGlossary`
  (`tier02_test.go:53-67`), which uses the identical joined-source pattern.

---

### Finding 2: Three tier READMEs state refusal counts that contradict the
### files and the ratchet — GAP

**What the plan commits to.** Plan lines 645-648: "The measured behaviour is
in the comment", justified by B::SoN's burn that "two of their three burns
were not loose assertions but tests passing where nobody could tell whether
that was still meaningful." A stated measurement that has silently gone false
is that failure mode in documentation form.

**Oracle:** O4 (`corpus.ratchet`) and O5 (anchored `^#\s*STATUS refuses`
count). Both agree; the ratchet records 21 refusing files corpus-wide and the
anchored marker count is also 21 once the one prose false-positive is excluded
(see NOTE 1).

**Evidence, per tier:**

| README claim | Stated | Actual | Delta |
|---|---|---|---|
| `01_literals/README.md:42` "Four of the tier's construct files refuse" | 4 | **3** | over by 1 |
| `04_operators/README.md:231` "Five files here are `STATUS refuses`" | 5 | **6** | under by 1 |
| `13_opaque/README.md:140` "Six of the fifteen files refuse" | 6 | **8** | under by 2 |

Measured:

```
$ grep -lP '^#\s*STATUS refuses\b' conformance/01_literals/*.t
08_signed_exponent.t  11_vstring_bare.t  12_vstring_v.t                      (3)

$ grep -lP '^#\s*STATUS refuses\b' conformance/04_operators/*.t
00_adjacency.t  02_string.t  04_string_comparison.t
05_logical.t  06_and_cliff.t  10_undef_arity.t                               (6)

$ grep -lP '^#\s*STATUS refuses\b' conformance/13_opaque/*.t
00_adjacency.t  01_heredoc_interpolating.t  02_heredoc_literal.t
03_heredoc_indented.t  06_format_write.t  08_data_section.t
09_format_picture.t  11_data_not_perl.t                                      (8)
```

**Tier 01 is stale in the direction that matters most.** Commit `a68ac5f9`
("fix(lexer): a leading-decimal literal is a number") fixed the `.5` split,
`04_leading_decimal.t` lost its `STATUS refuses` marker, and the ratchet now
records `- leading_decimal.t`. The README was not updated in that commit and
still claims four refusals. A count that over-states refusals makes the tier
look less finished than it is, and — worse for this plan — it is the artifact
a reader consults to know what the tier measured.

**Tier 13 contradicts itself internally.** `README.md:140` says "Six of the
fifteen files refuse", and `:147-155` then enumerates by refusal code:
"`unimplemented_statement` for the **two** format declarations ...
`trailing_tokens` for the **three** heredocs **and the adjacency file** ...
`not_a_term` for the **two** data sections". That is 2+3+1+2 = **8**, which is
the correct number. The headline sentence and its own breakdown disagree, and
`:165` compounds it with "would report the same **five** refusals". Three
different counts in one section.

**Tier 04's arithmetic survives, its total does not.** `:231` says "Five files
... and four of the five refuse on exactly that [Word-vs-operator]". From the
ratchet, four files do refuse with code `tokens`
(`and_cliff`, `logical`, `string`, `string_comparison`), so the "four" is
right. But there are six refusing files, not five: `10_undef_arity.t`
(`trailing_tokens`) and `00_adjacency.t` (`not_a_term,trailing_tokens`) are
both present, and the README only accounts for the adjacency one at `:248`.

**Why nothing caught it.** Only two README sections are machine-read:

```
internal/conformance/lint.go:43   func readTierOps(corpus string)   -> ## INTRODUCES
internal/conformance/lint.go:345  func readTierDeps(corpus string)  -> ## DEPENDS ON
```

Everything in the "What the tier measured" narrative sections is ungated
prose. The corpus has a strong mechanism for exactly this class of drift — a
file-level ratchet with a stale-marker failure — and the READMEs sit outside it.

**Layer responsible:** documentation, unguarded by tooling.

**Suggested remediation shape:** correct the three counts. Then, to stop
recurrence, consider a check that derives the per-tier refusal count from the
files and compares it against a machine-readable line in the README (e.g. a
`## REFUSALS` block in the same indented-code form the other two blocks use),
so the count joins the two sections that cannot drift. A prose sentence will
drift again; the plan's own argument at lines 1008-1014 ("in a Go table it is
a second list; in the tier directory the drift shows up in the same diff")
applies verbatim.

**Side effects of the fix:** adding a third machine-read block means
`readTierOps`' "first indented code block is the value" convention
(`TestReadTierOpsIgnoresTrailingProse`, `lint_test.go:571`) must be extended
carefully — that test exists because an earlier regex swallowed prose into
`DEPENDS ON`. Any new block parser needs the same anchoring and its own
malformed-input test.

---

### Finding 3: A glossary boundary is skipped citing a bug that has been
### fixed; the skip cannot go stale — GAP

**What the plan commits to.** Plan lines 686-695:

> The reverse is an ERROR, loudly:
>
>     file is marked `STATUS refuses` (01a0c13f-97f5) but now PASSES.
>     Remove the STATUS line -- a stale marker hides a regression.
>
> A marker that outlives its bug is worse than no marker, because it silences
> a file that has started failing again for an unrelated reason.

**What exists.** For *corpus files* this is fully implemented and tested:
`verdict()` returns `staleMarker`, and `TestStaleRefusalMarkerFails`
(`conformance_test.go:593-614`) asserts a passing file carrying a marker
fails. That is correct and I confirmed it runs.

For the *glossary boundary table* it is not. `glossary_test.go:194-196`:

```go
t.Run(tc.name, func(t *testing.T) {
        if tc.known != "" {
                t.Skipf("known refusal: %s", tc.known)
        }
        ...
```

The skip is unconditional on the presence of a `known` string. Nothing
verifies the refusal is still real, so a fixed bug leaves a permanently
skipped boundary.

**Trigger — one entry is already stale.** `glossary_test.go:99-101`:

```go
name: "a leading decimal point is part of the literal",
src:  "my $x = .5;", category: "numeric literal", text: ".5", want: 1,
known: "the lexer splits .5 into Operator(.) Number(5); M1 issue 01a0c13f-97f5",
```

Commit `a68ac5f9` fixed exactly this, in its own words:

> `.5` lexed as `Operator(.) Number(5)`, so a leading-decimal literal was
> never a term ... `scanNumber` now starts on a `.` that precedes a digit
> WHERE A TERM IS EXPECTED

Corroborating: `conformance/01_literals/04_leading_decimal.t` no longer
carries a `STATUS refuses` line (anchored count 0), and `corpus.ratchet`
records `- leading_decimal.t` (clean). The corpus half of the system has
recorded the fix; the glossary half still skips.

Observed in the run:

```
--- SKIP: TestCategoryBoundaries/a_leading_decimal_point_is_part_of_the_literal (0.00s)
```

**The other two skips in that table are believed live**, not stale:
`an_exponent_sign_is_part_of_the_literal` (`5e-1`) and
`a_v-string_is_not_a_numeric_literal` — both still have refusing corpus files
(`08_signed_exponent.t`, `11_vstring_bare.t`, `12_vstring_v.t` all carry
anchored markers and all appear in the ratchet). So this is one stale entry,
not three.

**Layer responsible:** `internal/conformance/glossary_test.go`.

**Suggested remediation shape:** invert the skip into an
expected-failure assertion — run the check, and *fail* if a `known` case now
passes, with the message the plan prescribes ("remove the `known` line"). That
makes the glossary table obey the same rule the corpus files already obey, and
it would have self-reported this finding. Removing the `known` line on the
leading-decimal entry is then a one-line follow-on.

**Side effects of the fix:** converting to expected-failure will immediately
fail on the leading-decimal entry (correctly), so the two changes land
together. The `5e-1` and v-string entries must keep working as
expected-failures, so the new helper needs to distinguish "still failing as
recorded" from "now passing".

---

### Finding 4: The negative-file commitment is one file; the croak splitter
### measures a split it never places — GAP

**What the plan commits to.** The plan devotes a substantial section to
negatives (lines 490-551) and returns to them at 952-967. Specific
commitments:

- "**Negative files, in-tier where they have a home.**" (line 490)
- "The first draft had NO negative coverage at all, and a suite that only
  tests what should parse cannot distinguish a real parser from one that
  accepts everything." (lines 492-494)
- "They place by construct, and they are better in-tier than gathered" with
  three worked placements — unterminated heredoc/`qw`/`q` to `13_opaque`,
  "Missing operator before `@foo`" to `04_operators`, "foo found where
  operator expected" to `07_subroutines` (lines 532-537)
- "201 of 340 are compile-time failures and become `parsent`" (line 959)
- The tooling list names "the croak extraction runner — with its `perl -c`
  split" as a blocking prerequisite (line 278)

**What exists.** The splitter is built, tested, and genuinely measured —
`internal/conformance/croak.go` exports `ExtractCroakCases` (`:204`) and
`MeasureCroakSplit` (`:292`), and the tests run for real against a perl
checkout rather than skipping (verified: `TestRuntimeCroakIsNotANegative`
PASS 8.72s, `TestCroakCountsReported` PASS 5.71s, `TestCroakSplitIsMeasured`
PASS 4.65s). `TestRuntimeCroakIsNotANegative` is a strong test: it asserts the
four wholly-runtime files classify as positives and fails if the set is empty.
The issue `01a0c35f` "The croak extraction runner, split by `perl -c` because
the filename lies" is closed, and against its own scope it is genuinely done.

What does not exist is the placement. The corpus contains exactly **one**
`--- expect parsent` file out of 175:

```
$ grep -lP '^--- expect parsent' conformance/*/*.t
conformance/07_subroutines/10_undeclared_callee.t
```

Nothing routes croak cases into tiers — no reference to croak in `lint.go`,
`run.go`, or any tier README. The ~201 compile-time failures the plan says
"become a `parsent` case" are measured on each run and discarded.

**Why this is a GAP and not BLOCKING.** The plan's tooling list (line 278)
asks for the *runner*, and the runner exists. The placement is implied by
"they place by construct, and they are better in-tier than gathered" but no
m3 issue owns it — I checked all 35 and none names negative placement.
Reasonable readers could scope `01a0c35f` either way. What makes it a finding
rather than a non-issue is the plan's own stated stakes: a suite that only
tests what should parse "cannot distinguish a real parser from one that
accepts everything", and at 1 negative in 175 files that discrimination is
approximately absent.

**Anti-pattern 7 guard — naming where this is deferred.** It is deferred
*nowhere*. The plan's closing section (lines 1033-1050) names exactly two
deliberate deferrals — the `t/` sweep and adjacency beyond the adjacent pair —
and states "Everything else in this document is owned by an issue in
`m3-conformance-corpus`." Negative placement is neither deferred there nor
owned by an m3 issue. That gap between the plan's closing claim and the issue
list is itself the finding.

**Layer responsible:** corpus content; milestone decomposition.

**Suggested remediation shape:** either (a) file an issue in the milestone
that follows m3 to place compile-fail croak cases in-tier, and amend the
plan's closing section to name it as a third deliberate deferral; or (b) if
placement was consciously judged out of scope, record that decision in the
plan so the "everything else is owned by an issue" sentence stays true. Option
(b) is cheap and restores the document's accuracy immediately.

**Side effects of the fix:** placing ~201 negatives in-tier will interact with
the dependency lint, which cannot lint a `parsent` file — perl builds no
optree for a program it will not compile. That path exists
(`tier07_callforms_test.go:296` skips lint for `parsent`, and
`TestLintSkipsAParsentFile` covers it), but it has been exercised by one file;
201 will find its edges. Tier file counts and derived numbering would also
move substantially.

---

## Notes

### NOTE 1: `grep 'STATUS refuses'` over-counts by one; the parser is correct

A naive count of files mentioning `STATUS refuses` gives 22 against the
ratchet's 21. The extra is
`conformance/07_subroutines/10_undeclared_callee.t`, which discusses the
marker in prose while explaining it deliberately carries none:

> It carries no `STATUS refuses` line either -- the runner consults our parser
> only for a file claiming `--- expect parses`, so there is no refusal of ours
> for a marker to go stale about.

The production parser is correctly anchored —
`internal/conformance/file.go:206`:

```go
var reRefusal = regexp.MustCompile(`(?m)^#\s*STATUS refuses\b`)
```

so the prose is excluded and marker count reconciles to 21 = 21 against the
ratchet. **No finding.** Recorded because it is a trap for the next auditor,
and because the file's reasoning for carrying no marker is correct and worth
preserving.

### NOTE 2: `TestEveryNamedConstructAppears` has the borrowed-word shape, and
### currently reaches the right answer anyway

`internal/conformance/namedconstructs_test.go:53` checks 35 constructs by
`\b<construct>\b` over all corpus sources. Three of the 35 have no file named
for them, and I traced what satisfies each:

| Construct | Satisfied by | Genuine? |
|---|---|---|
| `defined` | `04_operators/09_named_unary.t`: `my $loose = defined $x + 1;` | yes |
| `warn` | `07_subroutines/12_builtin_extent.t`: `my @greedy = (warn "a", "b");` | yes |
| `tr` | `09_regex/00_adjacency.t`: `my $tr = "a.c";` | **no — a variable named `$tr`** |

`\btr\b` matches inside `$tr` because `$` is a non-word character (confirmed
by probe). So one of the 35 is satisfiable by a variable name. It does not
currently produce a wrong answer, because
`09_regex/11_transliteration.t` genuinely contains `$t =~ tr/./Z/` and the
tier claims the `trans` op, which `TestTierRegexOpsAreEarned` verifies is
emitted.

The test's own comment (`:47-52`) anticipates this for `time` and argues the
tier lint covers it. That argument holds *here* because tier 09 claims
`trans`. It would not hold for a construct whose op the optimiser erases —
which is the plan's own repeated caveat (lines 866-870, "the optimiser can
erase the construct a tier is about"). **Not a finding today; a latent one.**
Tightening the pattern to reject a preceding sigil is a two-character change
if anyone touches this file.

---

## Commitment coverage, section by section

Walked per the brief. "Exists" means I located the artifact and, where it is a
check, confirmed from the verbose run that it executes and asserts.

### "The proposal" (lines 43-92)

| Commitment | State |
|---|---|
| Ordered by what a parser must understand first | EXISTS — 14 numbered tiers, `DEPENDS ON` gated by `TestRealCorpusNecessityHolds` |
| One construct per file | EXISTS — with the documented exception of adjacency files |
| Descriptively named, numerically prefixed | EXISTS |
| Every file validated against perl | EXISTS — `TestTier*PerlValidated` for all 14 tiers, all ran |
| Numbering regenerable, not hand-maintained | EXISTS — `TestDerivedTierNumberingRegenerates`, `TestAccidentalTierNumberingIsNotDerivable`; `TestEveryTierDeclaresFileOrder` records which convention each tier uses |
| No file count, no schedule | HONOURED — no count anywhere in the tooling |

### "The ordering criterion" (lines 94-120)

| Commitment | State |
|---|---|
| Partial order, not a chain | EXISTS — `specPrerequisites` (`prereq_test.go:59-74`) encodes `09_regex -> 01_literals` and `12_packages -> 07_subroutines`, consistent with plan lines 105-107 |
| Numbering is one topological sort | EXISTS — `checkTierNecessity` requires prerequisite < tier, not adjacency |
| oo before packages | EXISTS — `11_oo`, `12_packages`, and `11_oo` declares `08_references` |

### "The tiers" (lines 122-206)

All fourteen exist. No tier named in the plan is missing; no directory exists
that the plan does not name. `13_base` is correctly absent.

| Tier | Dir | README | Files | Adjacency | Per-tier test |
|---|---|---|---|---|---|
| 01_literals | yes | yes | 13 | yes | yes |
| 02_variables | yes | yes | 17 | yes | yes |
| 03_context | yes | yes | 11 | yes | yes |
| 04_operators | yes | yes | 16 | yes | yes |
| 05_scoping | yes | yes | 6 | yes | yes |
| 06_control | yes | yes | 18 | yes | yes |
| 07_subroutines | yes | yes | 14 | yes | yes (+ callforms) |
| 08_references | yes | yes | 13 | yes | yes |
| 09_regex | yes | yes | 13 | yes | yes |
| 10_io | yes | yes | 9 | yes | yes |
| 11_oo | yes | yes | 14 | yes | yes |
| 12_packages | yes | yes | 9 | yes | yes |
| 13_opaque | yes | yes | 15 | yes | yes |
| 14_recursive | yes | yes | 7 | yes | yes |

175 files total, matching the ratchet header's "175 files, 154 clean, 21
refusing" — which I independently confirmed.

**Content gap against the tier charter:** tier 01 only (Finding 1). The other
thirteen charters are covered as written, spot-checked: `02_variables`
includes `delete/exists` and `$::`; `06_control` includes `goto`
(`12_goto.t`); `11_oo` includes both `bless` and `class`/`field`/`method`/
`ADJUST` plus indirect `new`; `13_opaque` includes format, qx, `<*>`,
heredocs, POD, `__DATA__`.

**`hardMarkers` all place** — `TestEveryHardMarkerPlaced`
(`hardmarkers_test.go:151`) ran and passed, and it reads the claims from tier
READMEs rather than a Go table, which is the plan's stated preference.

### "Where the files come from" (lines 208-284)

| Commitment | State |
|---|---|
| Everything authored, nothing extracted | HONOURED — no T1 file content in the corpus |
| T1 a source of cases, cited never copied | EXISTS — `namedConstructs` is a measured T1 debt list (35 items), gated |
| Negatives extracted by a runner, not authored | PARTIAL — Finding 4 |
| Croak extraction splits by `perl -c` | EXISTS and genuinely runs |
| Tooling list, all 8 items | 7 of 8 complete; the 8th (croak) complete as a runner, not as placement |

### "What a corpus file contains" (lines 286-848)

| Subsection | Commitment | State |
|---|---|---|
| Structural first, behavioural backstop | Both layers present | EXISTS — token facts + `--- expect output` |
| The failure mode to design against | Emitter must not import from producer | OUT OF SCOPE for corpus; concerns `canon.go` |
| — | 31 adjacent precedence pairs under perl | EXISTS — `TestTierOperatorsAdjacentPairs` (`tier04_test.go:811`) ran and passed |
| Lexical facts on the token stream, not the CST | No corpus file asserts CST shape | HONOURED — verified: zero CST assertions |
| — | Declared fact, not a token dump | EXISTS — `fact.go:23` two-form grammar |
| Three kinds of file | Construct files | EXISTS |
| — | Adjacency file per tier, every construct in one body | EXISTS — `TestEveryTierHasAnAdjacencyFile` walks `specTiers` |
| — | Tier 11 adjacency catches ADJUST | EXISTS — `TestTierOoAdjacencyCatchesAdjust` (`tier11_test.go:458`); log confirms it catches `trailing_tokens at [97,134)` |
| — | Negative files in-tier | PARTIAL — Finding 4 |
| The file format | `--- ` opens a section | EXISTS |
| — | Repeated section is an error | EXISTS — `file.go:109` `duplicate section %q` |
| — | Exactly one of parses/parsent required | EXISTS — `TestParseFileBothExpectations` |
| — | One trailing newline stripped | EXISTS — `TestParseFileExpectOutputNewline` |
| — | Empty pinned output is an assertion | EXISTS — `TestEmptyExpectedOutputIsChecked` + `TestAbsentExpectedOutputSkipsCheck` (both halves) |
| Perl adjudicates before we do | CORPUS BUG reported, not a refusal | EXISTS — `TestPerlAdjudicatesFirst` |
| A known refusal skips; a stale marker fails | Skip on marker | EXISTS |
| — | Stale marker fails loudly | EXISTS for corpus files; ABSENT for the glossary table (Finding 3) |
| — | One passing file per tier "should not be optional" | TRUE IN FACT, UNGATED — every tier has ≥6 passing files; no test enforces it |
| Refusals carry a code | Codes not messages, assert on code | EXISTS — 10 sites retrofitted; `TestRefusalCodeMismatchFails`, `TestRefusalCodeParsesFromHeader` |
| The glossary is a prerequisite | Categories defined, growing an entry at a time | EXISTS — 12 categories, `TestGlossaryMatchesCategories` keeps `GLOSSARY.md` and `categories.go` in step |
| — | Coupling in one file | EXISTS — `categories.go`, 12 entries, sole reader of the token enum |
| The SoN IR is not a third layer | Not present | HONOURED |

### "The checks that make the ordering a claim" (lines 850-889)

**Check 1 — Dependency.** `TestCorpusLints` (`lint_test.go:208`). Real and
non-vacuous:
- Subtests per file, 175 of them, all ran.
- Guards against a *silently disabled* tier: a directory holding `.t` files
  that is not a numbered tier fails (`:222-236`). This closes the
  "renumbering typo disables a whole tier" hole explicitly.
- Runs the reverse direction too: a README claiming an op no corpus file emits
  fails (`:291-299`). That is the direction most lints omit.
- `TestLintRejectsUndeclaredConstruct`, `TestLintToleratesFoldedOps`,
  `TestUnclaimedOpReported`, `TestOpTableDerivedFromReadmes` cover the lint
  itself. `subops_test.go` proves `opsOf` sees inside named subs, methods, and
  other packages, with mutation tests.

**Check 2 — Tier necessity (REVISED form).** `checkTierNecessity`, gated over
the live corpus by `TestRealCorpusNecessityHolds` (`lint_test.go:543`) and
proven by six table cases in `TestTierNecessityCheck` (`:488`) including
self-dependency, a later prerequisite, a non-existent tier, and a
non-tier-name. Implements the REVISED criterion (prerequisite among earlier
tiers), not the superseded one.

Additionally — and this is the strongest single piece of work I found —
`prereq_test.go` documents that the per-tier op-pairing check was **measurably
vacuous** and replaces it:

> Three tiers proved it independently by mutation -- 12 repointed to 11_oo, 11
> to 07_subroutines, 04 to 02_variables -- and each stayed GREEN.

and then records that all three stronger op-based rules are dead, with counts
(103 ops, none claimed twice; "an op absent from the tier's own construct
files" fails 10 of 14). `TestTierPrerequisiteMutationsAreCaught` replays those
three real mutations against fixtures. This is the team finding and fixing its
own self-satisfying gate, which is the defect class this audit was
commissioned to hunt.

**Check 3 — perl validation.** `TestTier*PerlValidated` for all fourteen
tiers; every file a subtest; all ran (visible in the log). Backed by
`TestPerlIsTheMeasuredVersion` (asserts `$] == 5.042000` on the *resolved*
perl, not `PATH`'s) and `TestRunnerUsesPinnedPerl` (catches a call site
drifting back to bare `perl`). Two separate tests for two distinct failure
modes.

### Three sources layer (lines 872-889)

| Layer | State |
|---|---|
| The file DECLARES (`# TIER` header) | EXISTS — every corpus file carries `# TIER NN name` and `# INTRODUCES` |
| ops LINT it (`ops(file) ⊆ ∪ ops(tiers ≤ N)`) | EXISTS — `lintFile`, gated by `TestCorpusLints` over 175 files, both directions |
| The token stream classifies tiers 01-03 | **PARTIAL — real, but thin in tier 02** |

The third layer is **not absent** — this was the brief's specific suspicion
and the answer is no. It is implemented (`fact.go`, `categories.go`), gated
(`TestUnknownCategoryFails`, `TestGlossaryMatchesCategories`,
`TestCategoryBoundaries`), and used across the corpus:

```
tier:          files with --- expect tokens
01_literals:   12 of 13
02_variables:   4 of 17     <- thin
03_context:    10 of 11
```

Tier 02 is the outlier. Thirteen of its seventeen files carry no token fact,
including every file whose subject is a sigil or element-access spelling
(`01_array.t`, `02_array_element.t`, `03_array_last_index.t`, `04_array_slice.t`,
`05_braced_name.t`, `07_hash.t`, `08_hash_element_expr.t`,
`09_hash_exists_delete.t`, `10_hash_slice.t`, and the three `package_*` files).
`GLOSSARY.md:119-129` defines `variable` with four measured boundaries —
`$#x` is one token, `${name}` is one, `${ $ref }` is not, `$$` is one and
`$$ref` is two — and **no corpus file asserts any of them as a token fact**:

```
$ grep -rn 'whose text is "\$#' conformance/          -> (none)
$ grep -rn 'variable whose text is "\$\$"' conformance/ -> (none)
```

`TestTierVariablesCoversGlossary` (`tier02_test.go:49`) does cover these
boundaries, but by **source-text substring**, not by token fact — the same
mechanism as Finding 1 and with the same blind spot: it asks whether the
characters appear, not whether any file makes a lexical claim about them. The
test documents the choice ("a token-based check would ask our lexer whether
the tier covers a case our lexer gets wrong, and answer no"), which is a sound
argument for *coverage* but leaves the *assertion* unmade.

One boundary is properly deferred rather than dropped:
`02_variables/05_braced_name.t:17` states `${ $ref }` "is tier 08's, and
writing it here would be the file reaching forward", and
`08_references/05_brace_deref.t` exists. That deferral is honoured.

**Severity: GAP, recorded here rather than as a numbered finding** because the
layer exists and the boundaries are covered by *some* check. What is missing is
token-level assertion in the tier the plan says the token stream is supposed to
classify.

### "Sequencing" (lines 968-1000)

| Commitment | State |
|---|---|
| Corpus-tied gates die, incl. `easy_test.go` whole | EXISTS — issue `01a0c436` closed; no `easy_test.go` in the tree |
| `hardMarkers` mined for placements then deleted | PARTIAL — mined and gated by `TestEveryHardMarkerPlaced`; `hardmarkers_test.go` retained as the placement gate, which is a reasonable reading |
| Smallest useful corpus = tiers 01-04 + tier 07 call-form slice | EXISTS — `TestSmallestUsefulCorpus` (`tier07_callforms_test.go:545`) ran and passed |

### "Questions this document has closed" (lines 1002-1024)

| Question | Plan's answer | State |
|---|---|---|
| Does each tier get a README? | Yes, machine-readable INTRODUCES | EXISTS, 14 of 14 |
| Does `04_operators` need internal grouping? | Deferred to the tier-04 issue | HONOURED — tier 04 has 16 flat files; deferral intact |
| Construct set from declaration or CST? | Declaration + ops lint + token stream; CST out | EXISTS as specified; no CST assertions anywhere |
| `conformance/` at repo root? | Yes | EXISTS |
| What does a file look like? | "The file format" | EXISTS |

### Deliberate deferrals (lines 1033-1050)

| Deferral | Honoured? |
|---|---|
| The `t/` sweep — belongs to the milestone after m3 | YES — nothing in m3 runs it |
| Adjacency beyond the adjacent pair (N with N-3) | YES — every adjacency test pairs with the declared prerequisite only |

The plan's closing sentence — "Everything else in this document is owned by an
issue in `m3-conformance-corpus`" — is **false for one item**: negative-file
placement (Finding 4).

---

## Vacuity survey

The brief asked specifically for checks that would pass if the thing they
check were broken. Findings against each shape I searched for:

| Shape | Found |
|---|---|
| Assertion that can never fail | None found. Spot-checked every `TestTier*Adjacency`; all have a failure path with a non-empty expectation set |
| Gate satisfied by incidental data | **Finding 1** (tier 01 strings/`qw` via the adjacency file). Also NOTE 2 (`tr` via `$tr`), latent |
| Test greping for something in its own fixture | None found. Corpus-wide checks drive from `specTiers`, a list deliberately kept separate from the directory — `lint_test.go:455-459` documents why |
| Skipped test | 26 skips, all classified. 25 legitimate (documented refusals, env-dependent with named reason). **1 stale** — Finding 3 |
| `t.Skip` without recorded reason | None. All 10 `t.Skip` sites carry a reason; three carry a comment explaining why skipping beats passing vacuously |
| Empty loop body | None found |
| Env-dependent test silently not running | None. Checked the two risky ones: croak tests genuinely ran (8.72s/5.71s/4.65s wall time) and `git-zhi` is installed at `/home/perigrin/.local/bin/git-zhi`, so `TestRefusalCitationMustResolve` genuinely ran |
| A check whose expectation list is empty | Guarded explicitly in several places — e.g. `TestRealCorpusNecessityHolds` fails on `len(deps) == 0`, `TestTierPrerequisitesMatchTheSpec` fails on `len(declared) == 0` |

The `tierPending` skip (`lint_test.go:605-624`) deserves specific mention: it
skips a tier's subtests when the tier has no README. With all fourteen READMEs
present it is currently **inert** — I confirmed no `TestEveryTierHasAReadme`
or `TestEveryTierHasAnAdjacencyFile` subtest skipped in the run. It was a
live vacuity risk during construction and is now dead code with a documented
rationale. Not a finding; worth deleting when someone is nearby.

---

## What convinced me the rest is real

Recording this because "it passes" is not evidence and the brief asked for the
basis of any completeness claim.

1. **The team has already found and fixed this exact defect class twice**, and
   left the evidence in the code. `tier06_test.go:100-116` documents the
   `die`/`exit` borrowed-word bug the brief mentioned, names why it was
   invisible ("both files spell `eval` as their observation frame"), and states
   the rule: "A gate that passes on incidental vocabulary is worse than no
   gate." `prereq_test.go:23-56` documents three mutations that left the suite
   green and replaces the check that missed them.
2. **Checks are driven from lists deliberately held apart from the thing they
   check.** `specTiers` and `specPrerequisites` are both justified in comments
   with the same argument — "a list read from the thing it checks cannot report
   an absence". This is the single most common source of vacuity and it has
   been reasoned about explicitly.
3. **Both directions are checked where one would be easier.** The op lint
   catches under-declaration *and* over-declaration. `checkDeclaredPrerequisites`
   compares in both directions with a stated reason.
4. **Mutation testing appears repeatedly**, not just as a comment:
   `TestTierPrerequisiteMutationsAreCaught`, `TestOpsOfMutations`,
   `TestAccidentalTierNumberingIsNotDerivable`,
   `TestTierPrerequisiteCheckAcceptsTheUnmutatedFixture` (the control case).
5. **The environment-dependent tests genuinely ran here.** I verified by wall
   time and by locating `git-zhi` on disk, rather than trusting the absence of
   a SKIP line.

---

## Cross-references

- `docs/plans/2026-09-21-graded-conformance-corpus.md` — closing section needs
  a third deferral named, or an issue filed (Finding 4)
- `docs/plans/2026-09-21-cst-is-the-source-of-truth.md` — referenced by the
  plan as the reason the CST is not a fourth source; not audited here
- `internal/conformance/testdata/corpus.ratchet` — ground truth for Findings
  2 and 3; currently accurate at 175/154/21
- Milestone `m3-conformance-corpus` — reports 35/35; Findings 1 and 4 are work
  the milestone's issues either under-delivered (01a0c360) or never owned
  (negative placement)
- Issue `01a0c360` — its first acceptance criterion names string literal
  boundaries and is satisfied vacuously (Finding 1)
- Commit `a68ac5f9` — fixed the leading-decimal lexer bug; left
  `01_literals/README.md:42` and `glossary_test.go:101` stale (Findings 2, 3)

---

## Punch list

| # | Finding | Severity |
|---|---|---|
| 1 | Tier 01 has no string/quoting/`qw` construct files; its glossary gate is satisfied by the adjacency file | BLOCKING |
| 2 | Three tier READMEs state false refusal counts (01: 4→3, 04: 5→6, 13: 6→8); tier 13 contradicts itself internally | GAP |
| 3 | `glossary_test.go:101` skips a boundary whose bug `a68ac5f9` fixed; the skip cannot go stale | GAP |
| 4 | Negative placement is 1 file in 175; the croak splitter measures a split it never places, and no issue owns it | GAP |
| — | Tier 02: 13 of 17 files carry no token fact; no file asserts the `variable` glossary boundaries | GAP (in "Three sources layer") |
| — | `TestEveryNamedConstructAppears`: `\btr\b` matches `$tr`; right answer today, latent | NOTE |
| — | `grep 'STATUS refuses'` over-counts by 1; production regex is correctly anchored | NOTE |
| — | `tierPending` is now inert with all 14 READMEs present | NOTE |
