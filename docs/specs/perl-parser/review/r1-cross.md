<!-- ABOUTME: Cross-cutting ponytail review of the Perl parser specification (round 1). -->
<!-- ABOUTME: Hunts contradictions and duplication between chapters; proposes owners and a minimum reading path. -->

# Cross-cutting review, round 1

Scope: the shape of the whole document, not any one chapter. Every "perl says"
below was run on perl 5.42.0; every `toke.c`/`perly.y` cite was checked against
`~/dev/perl5` at `94e5086608`. Actual line counts: 13,560 (README says 12,700).

## 1. Contradictions

| # | Claim | Chapter A says | Chapter B says | Which is right |
|---|---|---|---|---|
| 1 | Empty `{}` at statement position | ch5 5.11 (`05:2354`): "empty `{}` at statement position is a **block**"; Go code returns `false` on `RBRACE` | ch3 3.3 (`03:224`): next char `}` → **anon hash** (`toke.c:6714`) | **ch3.** `perl -MO=Concise -e '{}'` → `emptyavhv ANONHASH`. ch5 is wrong. |
| 2 | `{ $x, 1 }` / `{ $x => 1 }` at statement position | ch5 5.11 (`05:2356-2357`, code `05:2373` `case ... SCALAR_VAR`): scalar followed by `,` or `=>` → **hash** | ch3 3.3 (`03:224-231`): only a quoted string, `q`-word or bareword is scanned; `$` is neither → **block** | **ch3.** `perl -MO=Deparse -e '{ $x => 1 }'` → `{ $x, '???'; }` (a block). ch5 wrong. |
| 3 | `{ foo, 1 }` (lowercase bareword + comma) | ch5 5.11 (`05:2356`): bareword followed by `,` → hash. ch4 4.4.4 (`04:562`): `word ,` → hash | ch3 3.3 (`03:229-235`): comma counts only if first char is `q` or **not lowercase** (`toke.c:6806`) | **ch3.** `perl -MO=Deparse -e '{ foo, 1 }'` → block; `{ Foo, 1 }` → `+{'Foo', 1}`. ch3 is the only copy that matches `toke.c` and perl. |
| 4 | Are `__END__`/`__DATA__` line-anchored? | ch5 5.1.2 (`05:127-133`, `05:147-148`): "on a line by itself... must be at the start of a line" | ch2 2.5.4 (`02:670-680`): "ordinary keywords, not line-anchored markers... **[verified]** mid-line" | **ch2.** `print "x\n";  __DATA__` prints `x` and stops. |
| 5 | Character after `=` that opens POD | ch5 5.1.3 (`05:180`): "an identifier character" | ch2 2.5.1 (`02:566`): `isALPHA` (`toke.c:9728`) | **ch2.** `=_x` and `=1` at column 0 are syntax errors. |
| 6 | Genuine syntax errors in the 620-file corpus on 5.42 | ch0 0.5 (`00:143-146`): **1** (`op/for-many.t`); 0.7 names only that file | ch7 7.3.4 (`07:652-660`): **2** — `op/signatures.t:927` (named params) and `for-many.t:474` | **ch7.** `perl -c t/op/signatures.t` → `syntax error at line 927, near "(:"`. ch0's classifier undercounted; its "one file in 620" line is wrong. |
| 7 | Corpus compile rates | ch7 7.3.4 table (`07:599-612`): `class/` 1/12, `io/` 17/44, `op/` 142/228, `mro/` 62/73, `re/` 55/80 — sums to 340/511; fix is "build the tree or strip the prelude" | ch0 0.5 (`00:110-125`): `class/` 10/12, `io/` 44/44, `op/` 223/228, `mro/` 73/73, `re/` 79/80, **584/620**; fix is the two-root shim | **ch0.** ch7's table is the pre-shim run that ch0 §0.5 explicitly retracts. ch7 M3 already quotes 584 (`07:1567`), so the chapter is half-updated. Not the 411/620 figure, but the same class of stale number. |
| 8 | Parse-relevant `BEGIN` | ch0 0.1 (`00:19-20`): boilerplate 466, relevant **21 (3.4%)**; README `:40` same | ch3 3.6.3 (`03:544-545`): boilerplate 448, relevant **8 (1.3%)**; ch3 3.0 (`03:31`) "1–8%"; ch3 3.11 (`03:957`) "1.3% + 0.3%" | Different definitions of "relevant", never reconciled. Pick one number and one definition; ch0 owns it. Also `use constant`: ch0 27 files, ch3 32 — measured **27**. |
| 9 | Number of precedence levels | ch4 4.0 (`04:47` "those 33 lines"), 4.1 (`04:77`), 4.16 (`04:1588`), README `:26`: **33** | ch4 4.13.7 (`04:1355`): **32**; the 4.1 table has 32 rows | **32.** `perly.y:150-182` holds 32 `%left/%right/%nonassoc` lines. |
| 10 | PerlOnJava's precedence table | A1.1.6 (`A1:243`): "cleanest transcription... copy it as data"; A1.5 (`A1:947`): `precedence.go 24-level table ← PerlOnJava verbatim`; A1.8.4 (`A1:1201`): "copy... verbatim" | ch4 4.13.1-4.13.4, 4.13.8 (`04:1244-1295`, `04:1359-1371`): four PerlOnJava precedence bugs, one **high** severity (file tests below `.`) | **ch4.** README itself says five precedence bugs were found in the references. A1 was never updated after that finding. |
| 11 | How many `PL_expect` states the Go lexer carries | ch2 2.2.1-2.2.2 (`02:266-308`) and ch3 3.1.2, 3.9.4 (`03:68-80`, `03:888-896`): **11**, "do not economise here" | ch5 5.0 (`05:82`): `ExpectKind // XSTATE, XTERM, XOPERATOR, XBLOCK, XATTRBLOCK, XATTRTERM` — **6**. ch6 6.3.1 (`06:394`): `Expect ExpectMode // term vs. operator` — **2**, plus perl-lsp's `AfterSub`/`AfterArrow`/`HashBraceDepth` bools; safe checkpoint is `Expect == ExpectTerm` (`06:418`). A1.8.1 (`A1:1188`): perl-lsp's 5 modes "solve the same problems structurally" | **11** (`perl.h:5972-5985`). ch6 reintroduces exactly the ad-hoc fields ch2 2.2.2 (`02:296-303`) calls "wrong for a compiler". |
| 12 | What a lexer checkpoint must contain | ch2 2.14.1 (`02:2321-2333`): 11 fields incl. `SublexStack`, `InPod`, no last-lop | ch3 3.10 (`03:909-915`): 5 fields incl. `LastLopOp`, no sublex/POD. ch6 6.3.1 (`06:394-406`): perl-lsp's field set | Three structs for one thing. ch3's `LastLopOp` is needed (`toke.c:6700` tests `PL_last_lop`); ch2's sublex stack is needed. Nobody has the union. |
| 13 | What the tree is | ch4 4.14 (`04:1381-1400`): typed Go structs, `Node interface{ Base() *nodeBase }`. ch5 5.17 (`05:3132-3170`): typed structs with `Span Span`, `Stmt` interface, `VarExpr` (ch4 calls it `Var`). A1.5 (`A1:940`): "typed structs + type switch, ~70 kinds" | ch6 6.4.1 (`06:531-556`): uniform CST `Node{kind KindID; children}`, string kinds `"expression_statement"`, `"variable_declaration"` (`06:768-776`), and "PSC does not change — not one of its 4,265 lines" (`06:527`). A1.8.7 (`A1:1213`): keep the tree-sitter API | Cannot all be true. PSC keys on tree-sitter kind names, so ch4/ch5's typed AST means rewriting PSC; ch6's CST means ch4 4.14 and ch5 5.17 are not the tree. **No chapter owns this decision.** |
| 14 | When to build incrementality, and at what grain | ch1 1.2 (`01:25`): "retrofitting incrementality onto a batch parser is a rewrite". ch7 M2 (`07:1460-1475`): incremental is milestone **2**, before the oracle (M3) and prototypes (M4) | ch5 5.14.5 (`05:2964-2976`): start full-reparse, add incremental "only when profiling says", **sub-body** level, "do not attempt statement-level". ch6 6.5.2 (`06:783-785`): "start with **statement-level** and move up". ch6 6.11 (`06:1799-1812`): incrementality is steps 6-7, after measuring | Four orderings. ch6 6.11 and ch5 5.14.5 agree on "full reparse first"; ch7's ladder and ch6's own §6.5.2 do not. |
| 15 | Panic-mode sync points | ch5 5.13.4 (`05:2636-2638`): perl-lsp "is *not* line-start sensitive. That is a simplification you should not copy"; column-0 tiers, `braceDepth = 0` reset (`05:2662-2684`) | ch6 6.8.3 (`06:1340-1378`): flat `isSyncToken` set, not line-start sensitive, local depth tracking — the design ch5 says not to copy. ch2 2.14.4 (`02:2385-2390`): a third, four-item list | ch5's tiered version is the one with a rationale. ch6 contradicts it in code. |
| 16 | HIR | ch6 6.1.4 (`06:207-227`): "build **neither** [HIR nor PIR], at first"; a side table keyed by byte offset | A1.2.3 (`A1:540-549`): "AST → HIR is a real separation and worth having. Build AST and a scope graph"; A1.5 layout has `internal/hir`; A1.6 (`A1:1040`) budgets 4,000 lines for it | ch6 is the architecture chapter; A1 should not carry a competing recommendation. |
| 17 | `B::Deparse`'s role | ch1 1.4 (`01:78`): "weaker than it looks... **never** for disambiguation" | ch3 3.9.1 (`03:773-776`): "Deparse is the workhorse" of the differential test; ch3 3.12 (`03:980`) harness compares against Deparse | ch7 7.1 has it right (Concise = shape, Deparse = round-trip). ch3 should say that. |
| 18 | Feature bundles | ch5 5.9.3 (`05:2131-2132`): `:5.36 ... -bareword_filehandles`; `:5.38 -smartmatch` | `feature.pm` on 5.42: `bareword_filehandles` is still in `:5.36`, removed in `:5.38`; `smartmatch` is in `:5.38` and `:5.40`, removed in `:5.42` | **feature.pm.** README says "a feature gate moved by two versions" — two more have. |
| 19 | `StatementBoundaries` predicate | ch6 6.5.2 (`06:739-741`): "Use chapter 5's `StatementBoundaries` predicate; do not re-derive it here" | ch5: no such predicate exists (0 grep hits) | Dangling cross-reference. |
| 20 | README chapter sizes | README `:22-30`: ch0 127, ch2 2,499, ch5 2,595, ch6 1,817, A1 1,215; "Roughly 12,700" | `wc -l`: 271, 2,511, 3,204, 1,841, 1,219; 13,560 | Stale by 860 lines. |

Smaller, verified, chapter-local (for the chapter reviewers):

- ch6 6.10.1 (`06:1656`) "5,000 lines is the *median* real file": perl5/lib `.pm` median is **164** lines (n=73, p90=802, max 7,690). The latency budget is built on this.
- ch6 (`06:220`, `06:527`) "4,265 lines" of PSC: `internal/infer` non-test is 5,012.
- ch3 3.1.3 macro table: `FTST` is `toke.c:261` not 256; `LOOPX` is `toke.c:257` not 252. `XTERMORDORDOR` is also set by `UNIDOR` (`toke.c:297`, `shift // 0`), which ch3 omits and ch2 (`02:277`) gestures at.
- ch0 0.5's per-directory table lists 591 of the 620 files (missing `test_pl bigmem win32 japh benchmark`), so its rows sum to 559 compiled, not the 584 it totals.
- ch0 0.6 "core" is 44 files (`base comp cmd opbasic`); ch7 T2 "core" is 56 (adds `class`). The 65.9% starting line and the M0/M1 gates are over different sets.
- ch6 6.4.1's method-count contract holds on recount (`Kind` 112, `Child` 119, `ChildCount` 109, `IsNamed` 50, `Text` 47, `StartByte` 29, `EndByte` 13, `Parent` 6); `IsError`/`HasError` are 0 in non-test code.

Things I checked that are **consistent**: 584/620 = 94.2% (the 411/620 figure appears only as ch0's documented retraction); the latency direction (parse 20-50× `Analyze`) in ch0, ch6, README; `indirect` off at `use v5.36` in ch0, ch3, ch5; context propagation lives only in ch4 4.12; heredocs live only in ch2 2.9 (66 mentions elsewhere are all references, not re-explanations).

## 2. Duplication

| Topic | Locations (lines) | Recommended owner | Lines saved |
|---|---|---|---|
| `{` block vs anon-hash heuristic | ch3 3.3 (42), ch4 4.4.4 (15) + 4.9.2 (33), ch5 5.11 (55), ch6 6.1.2 row A6 | **ch3 3.3** — the only copy that matches `toke.c` and perl. ch4 keeps one sentence + the `BlockGuessed` field; ch5 keeps one sentence | ~95, and two wrong Go functions (`leftCurlyIsHash`, the `looksLikeAnonHash` comment) |
| `PL_expect` state table | ch2 2.2.1-2.2.2 (56), ch3 3.1.2-3.1.4 (80), ch3 3.9.4 five-state critique (30, repeats ch2 2.2.2) | **ch3 3.1** for the states and transitions; ch2 keeps `PL_lex_state` (2.2.3) and a pointer | ~60 |
| Error recovery | ch5 5.13 (361), ch6 6.8 (163), ch2 2.14.3-2.14.4 (25) | **ch5 5.13** (it has the rationale and the tiers). ch6 6.8 keeps only the diag filter + PSC gating (6.8.4, ~50 lines) and the "never reuse an error subtree" rule (6.8.5) | ~120 |
| Incremental re-parse: boundaries, checkpoints, invalidation | ch2 2.14 (79), ch3 3.10 (38), ch5 5.14 (133), ch6 6.1-6.5 (~900), ch7 7.7 (104) | **ch6.** ch2/ch3/ch5 each shrink to "the state this layer must expose to a checkpoint" (≤15 lines each). ch5 5.14.1's perl-lsp cautionary tale is already ch6 6.1.3 | ~200 |
| Incremental == fresh-parse property test | ch6 6.5.4 (33), ch7 7.7.1-7.7.2 (55) | **ch7 7.7**; ch6 states the property in one sentence and points | ~30 |
| Oracle probes (`-c`, `prototype`, Concise, Deparse) | ch1 1.4 (25), ch3 3.9.1 (30), ch7 7.1.1-7.1.2 (110), ch0 0.3 (15) | **ch7 7.1**; ch1 and ch3 keep the srefgen one-liner and point | ~45 |
| Prototype character table | ch3 3.5.2 (18), ch5 5.5.5 (16) | **ch3 3.5.2** | ~16 |
| POD, `__END__`/`__DATA__`, shebang | ch2 2.3.4 + 2.5 (~175), ch5 5.1.2-5.1.4 (~97) | **ch2**; ch5 keeps the `File.Data` / `Pod` structs (~20). Deletes contradictions 4 and 5 for free | ~75 |
| Parse-vs-Analyze latency table | ch0 0.8 (40), ch6 6.10.2 "Measured today" (30), README (10) | **ch0 0.8**; ch6 and README keep the headline number and cite | ~30 |
| `use constant` / known-module prototype tables | ch3 3.6.4 tiers 0-1 (12), ch5 5.9.3 "use constant" + "Anything else" (15) | **ch3 3.6.4** for resolution; ch5 keeps the feature-bit table | ~15 |
| Corpus inventory and measurement | ch0 0.5-0.7 (110), ch7 7.3.1 (42), 7.3.4 (103, stale), 7.3.6 (16), ch2 2.15.1, ch5 5.16 | **ch0** for numbers, **ch7 7.3.6** for tiers. ch7 7.3.4 becomes "see §0.5" plus the version-pin paragraph | ~80 |
| Reference-implementation bug lists | ch2 2.16 (58), ch4 4.13 (134), ch5 5.15 (127), A1.1.7 (97), A1.2.6 (88) | Per-chapter divergence notes are fine. A1 must stop re-asserting what ch4 refuted (contradiction 10) | ~40 in A1 |
| HIR/PIR assessment | ch6 6.1.4 (45), A1.2.3 (55) | **A1** describes, **ch6** decides — and they must agree | ~20 |
| "What to build first" | ch2 2.15 (19 steps), ch3 3.12 (14), ch4 4.16 (15), ch5 5.16 + 5.18, ch6 6.11 (8) + 6.12, ch7 7.8 ladder + 7.9 + 7.10 (13), A1.6 milestones (5), A1.8 (7) | **Nine lists.** ch7 7.8 owns milestones; README owns the reading path. Per-chapter checklists may stay as acceptance lists but must stop each naming a different "first" (byte reader / Expect enum / Document+line table / oracle test) | ~60, plus one decision |

Total: roughly 900 lines of duplication, of which the `{` heuristic and the
incremental sections carry the contradictions.

## 3. Chapter-level verdicts

| File | Verdict | Reasoning |
|---|---|---|
| `README.md` | **KEEP, FIX** | Right shape. Fix the line counts and "33 levels"; add the minimum path (§4). |
| `00-findings.md` | **KEEP AS IS** | The evidential spine. Fix "one file" → two (contradiction 6) and make the table sum to its total. |
| `01-scope.md` | **KEEP AS IS** | 133 lines, does its job. Probe table → pointer to ch7 7.1. |
| `02-lexical-structure.md` | **SHORTEN** | The lexer of Perl is legitimately 2,300 lines. Hand 2.14 to ch6 and the perl-lsp critique in 2.2.2 to ch3; ~150 lines out. |
| `03-lexer-feedback.md` | **KEEP AS IS** | The best chapter and the one that is right where others are wrong. Make it the declared owner of `PL_expect`, `{`, and prototypes. Reconcile 8-vs-21 with ch0. |
| `04-expressions.md` | **KEEP AS IS** | Fix 33→32; replace 4.9.2's code with a pointer to ch3 3.3. |
| `05-statements.md` | **SHORTEN** | 3,204 lines, the longest, and where the perl-verified errors cluster. 5.1.2-5.1.4 → ch2, 5.11 → ch3, 5.14 → ch6, 5.15 → divergence notes. ~400 lines out; what remains is good. |
| `06-incremental-lsp.md` | **SHORTEN** | The architecture is sound but half of it is not a Perl spec: 6.6 position tracking (163), 6.7 Go concurrency (174), 6.9.2 JSON-RPC framing by hand (116) are generic LSP-server design. Move those to a server design doc; keep 6.1, 6.3-6.5, 6.8.4-6.8.5, 6.10 (fix the median). Align its `Expect`, checkpoint, sync set and AST with ch2/ch3/ch5. ~500 lines out. |
| `07-conformance.md` | **SPLIT: half stays, half MOVES OUT** | 7.1 (oracle), 7.2 (metrics), 7.5-7.7 (differential, fuzz, property) define what "conforms" means — that is spec, ~800 lines, keep. 7.3.1/7.3.7 corpus inventory and vendoring, 7.4 ratchet + CI YAML (217), 7.8 milestone ladder (170), 7.9-7.10 are a project plan → `docs/plans/`. Replace 7.3.4 with a pointer to §0.5. |
| `A1-prior-art.md` | **MOVE OUT** | Excellent research notes; not this document. Kept inside the spec it asserts four refuted recommendations with the spec's authority (contradictions 10, 11, 13, 16). Move to `docs/plans/` or `docs/research/`; if anything stays, it is A1.3 (three parsers measured, ~100 lines) as an appendix. |

Post-cuts the spec is roughly 11,000 lines, 8 files, with one owner per topic.

## 4. The minimum path

The README says "read in this order" and lists all nine as equally required.
Nothing tells an implementer what the smallest working thing is, and ch7's own
M0 ("lexer round-trips the core, no parser yet") is the answer buried on line
1418 of the seventh chapter.

**To M0 (a lexer that round-trips `t/base t/comp t/cmd t/opbasic`), ~3,100 lines:**

1. `README.md` — the four findings (95)
2. `00-findings.md` (271)
3. `01-scope.md` (133)
4. ch3 3.0-3.4 — the feedback model and `PL_expect` (~250 of 987)
5. ch2 2.1-2.13 — the tokens (~2,260; skip 2.14-2.16)
6. ch7 7.2 + M0 gate (~100)

**To M1 (parses the core without error, feeds PSC), ~7,100 lines — about half:**

7. ch3 3.5-3.9, 3.11 — prototypes, `BEGIN`, barewords, the approximation table (+~400)
8. ch4 4.1-4.2, 4.4-4.9, 4.12, 4.14 (~1,000 of 1,652)
9. ch5 5.0-5.12 (~2,480; skip 5.13-5.15)
10. ch7 7.1, 7.9 — the oracle and the first test (~250)

**Deferred until M1 is green:** ch5 5.13 (recovery), ch6 (incrementality and
the server), ch7 7.5-7.7 (differential, fuzz, property). A1 is background
reading, not path.

Put that list in the README under "Read this first". It costs 20 lines and it
is the single change most likely to get a parser written.

## The three things I would actually do

1. **Delete the wrong copies.** Remove ch5 5.11, 5.1.2-5.1.4, 5.14 and ch4 4.9.2's code, pointing at ch3 3.3, ch2 2.5, ch6 — one commit removes three perl-verified false claims and ~350 lines.
2. **Decide what the tree is.** One page: CST with tree-sitter kind names (PSC unchanged) or typed AST (PSC rewritten). Make ch4 4.14, ch5 5.17, ch6 6.4.1 and A1.5 point at it; today they describe three incompatible trees and ch6's "PSC does not change" promise hangs on the answer.
3. **Move A1 and ch7's plan half to `docs/plans/`, fix ch7 7.3.4, and add the 7,100-line minimum path to the README.** The spec gets ~2,500 lines shorter and stops contradicting itself about precedence tables, expect states, HIR and build order.
