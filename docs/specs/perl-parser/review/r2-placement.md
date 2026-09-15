<!-- ABOUTME: Round 2 placement record: what left the Perl parser specification for docs/plans/, and why. -->
<!-- ABOUTME: Lists rescued facts, fixed references, and the file inventory after the split. -->

# Round 2 — placement of chapter 7 and appendix A1

Round 1 (`r1-cross.md` §3) said ch7 should split and A1 should leave. This
round did it. Nothing was deleted; every measured fact is in one of the two
trees, and the table below says which.

## What moved where

| From | To | Lines |
|---|---|---:|
| `A1-prior-art.md` (whole file, `git mv`) | `docs/plans/2026-09-05-parser-prior-art.md` | 1,239 → 1,246 (provenance note added) |
| ch7 §7.1.4 Go oracle harness (code) | plan §1 | 145 |
| ch7 §7.1.6 CI speed techniques + `TestOracleCorpus` | plan §2 | 54 |
| ch7 §7.3.4 stderr classifier rule, `-T` shebang note | plan §3.1 | 16 (rewritten) |
| ch7 §7.3.7 vendoring options, `corpusRoot`, `corpus.pin` | plan §3.2 | 51 |
| ch7 §7.4 ratchet: idea, baseline format, `TestRatchet`, CI YAML | plan §4 | 215 |
| ch7 §7.8 milestone ladder M0–M6 | plan §5 | 170 |
| ch7 §7.9 first test, §7.10 checklist | plan §6, §7 | 73 |

Chapter 7 went from 1,647 lines to 894. What stays: §7.0 the problem, §7.1
the oracle and its limits, §7.1.5 the exact/wider/WRONG/no-answer buckets,
§7.2 the four metrics, §7.3 the corpus and its tiers, §7.5 differential
testing, §7.6 fuzz invariants, §7.7 the incremental property. Section
numbers were **not** renumbered; each moved section leaves a heading with a
one-paragraph statement of what the spec still requires plus a pointer to
the plan section (§7.1.4, §7.1.6, §7.3.4, §7.3.7, §7.4, §7.8). Round-1
reviews cite ch7 by number and those citations still resolve.

## Facts checked before moving

Grepped `0[0-6]-*.md` and `README.md` for `§7.`, `chapter 7`, `M0`–`M6`,
`A1.`, `appendix`, `prior art`, licence terms, the 25k-line budget, and the
perl-lsp "reparses fully" quote.

- **No spec chapter cites A1 by section.** ch1 §1.6 listed it in the
  chapter table (now a pointer); ch6 §6.1.3 quotes `text_sync.rs:1-9`
  directly; ch1 §1.7 carries its own source line counts; ch2 §2.1.2 has its
  own copy of the PerlOnJava lexer quote. Nothing in the spec depends on
  A1.3 (three parsers measured), A1.6 (budget) or A1.7 (licensing).
- **The ~25k-line budget** (24,003 vs ~28,000, A1.6) is already in
  `docs/plans/2026-09-04-perl-parser-decision.md` "The cost". It is a
  project-cost estimate, not a fact about Perl, so it was not added to
  `00-findings.md`.
- **Licensing** lived in A1.7 and ch7 §7.3.7. The one-sentence fact (perl5
  and PerlOnJava tests are Artistic/GPL) stays in the §7.3.7 stub; the
  options analysis is plan §3.2 and the licence texts are in the prior-art
  notes.
- **Timings** (8.8 ms `perl -c`, 25 ms Concise) were already in §7.1.1's
  table; the raw "100 runs = 0.88 s" measurement is plan §2 only.
- **The stale §7.3.4 compile table** (`class/` 1/12, `io/` 17/44 …) is gone.
  It was not copied to the plan either — `00-findings.md` §0.5 owns those
  numbers and §7.3.4 now points there. The two genuine syntax errors and the
  version-pin requirement are stated in prose with pointers to §0.5/§0.7.
- **Rescued into `00-findings.md`: nothing.** No moved fact was one the spec
  depended on and lacked.

## References fixed

| Where | Was | Now |
|---|---|---|
| ch7 §7.1.3 | "§7.7 covers sandboxing" — §7.7 never did; pre-existing false pointer | "runs in CI against a pinned perl (plan §4.4)" |
| ch7 §7.1.3 | "See §7.3.4 — this bit users already" | "See §0.7" |
| ch7 §7.3.2 | `signatures.t`/`for-many.t` "see §7.3.4" | "see §0.5 and §0.7" |
| ch7 §7.3.6 | "covers the ladder through M4. Do not vendor T5 until M5" | points at §7.8 stub / plan |
| ch7 §7.2 | "ratchet" used undefined | one-clause definition + plan §4 |
| ch7 ABOUTME | named the ratchet and the ladder | names what the chapter now holds |
| `00-findings.md` §0.6 | "the number M1 has to move" — M1 no longer defined in the spec | names the plan and §5. One-line reference fix; no fact changed |
| `01-scope.md` §1.6 | table row for A1 | two-sentence pointer to both plan documents |
| `README.md` | nine-row table, exact `Lines` column, "12,700 lines" | eight rows, size buckets, "roughly 11,600"; new "Moved out of the specification" section with both pointers |
| plan (11 sites) | bare `§7.x` that now cross the file boundary | `spec §7.x`, or the plan's own section where the target moved too |
| plan §3.2 code comment | `see docs/specs/perl-parser/07-conformance.md §7.3.7` | `see docs/plans/2026-09-05-parser-conformance-plan.md §3.2` |
| prior-art notes | "Chapter 4 §4.13" etc. with no anchor | header paragraph defines "Chapter N" as `docs/specs/perl-parser/0N-*.md` and states the spec wins on disagreement |

Post-edit grep of the spec directory for `A1`, `appendix`, `§7.4`, `§7.8`,
`§7.9`, `§7.10`, `§7.1.4`, `§7.1.6` finds only the stubs and the README
pointer (ch6's "Class A1" trigger labels are unrelated).

## File inventory

```
docs/specs/perl-parser/
   296  00-findings.md      (+1 line: M1 pointer)
   136  01-scope.md         (+3: A1 row → pointer paragraph)
  2511  02-lexical-structure.md
   987  03-lexer-feedback.md
  1656  04-expressions.md
  3204  05-statements.md
  1845  06-incremental-lsp.md
   894  07-conformance.md   (was 1,647)
   105  README.md
 11634  total (was 13,560 with A1)

docs/plans/
   758  2026-09-05-parser-conformance-plan.md   (new)
  1246  2026-09-05-parser-prior-art.md          (moved; was A1-prior-art.md)
```

## Was the split the right call?

**A1: yes, without reservation.** Nothing in the spec cited it, its two
load-bearing facts already live in the decision memo and ch6, and inside
the spec it carried four recommendations the spec's own chapters had
refuted (r1-cross contradictions 10, 11, 13, 16) with the spec's authority.

**Ch7: right, with three costs worth stating.**

1. **Six stub headings.** Keeping section numbers stable means a reader of
   ch7 meets six short sections that say "the rest is in the plan". That
   is the price of not breaking round-1 citations and the plan's own
   `spec §7.x` references. Renumbering would have been cleaner to read and
   would have broken every existing citation; I chose stability.
2. **The boundary is a judgment, not a rule.** I kept §7.3 (corpus
   inventory, pathological files, tiers) in the spec on the grounds that
   the metrics are measured *over* that corpus, so naming it is part of
   defining conformance. §7.3.5 (a tour of PerlOnJava's corpus) and §7.3.6
   (tiers) exist mostly to serve the ladder, and a stricter reading would
   move them too. Likewise §7.5.4, §7.6.2, §7.7.2 keep ~250 lines of Go
   because there the code *is* the statement of the invariant, while
   §7.1.4's harness code moved because it is glue. Someone who wants "no
   code in the spec" has three more sections to move.
3. **A two-way dependency now exists.** The plan cites `spec §7.x`; the
   spec stubs cite `plan §N`. Renumber either side and the other dangles.
   That is inherent in a split and is the reason for cost 1.

What the split bought: chapter 7 no longer carries a schedule that will be
wrong the week M0 slips, or CI YAML pinned to `go-version: '1.23'`, or a
retracted compile table. Those rot on a different clock from "what `srefgen`
in an optree means", and now they rot in a file whose name has a date on it.
