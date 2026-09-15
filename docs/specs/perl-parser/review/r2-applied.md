<!-- ABOUTME: Round 2 of the ponytail review, applied: one owner per topic, the losing copies deleted. -->
<!-- ABOUTME: Records what was cut from where, one new contradiction found in an owner chapter, and what was deliberately kept. -->

# Round 2 — applied

Round 1 found the contradictions; this round deleted the losing copies and
left one-line cross-references. Every `toke.c` cite and every `[verified]`
measurement in a deleted section was moved to the owner first. Perl claims
below were re-run on 5.42.0; `toke.c` cites are against `94e5086608`.

## Ownership, applied

| Topic | Owner | Deleted from | Lines out |
|---|---|---|---:|
| Expectation-state machine (11 states, perl's names) | **ch3 §3.1** (not §3.2 — the brief's number was off by one; §3.2 is the disambiguation table) | ch2 §2.2.1's 11-row table and §2.2.2's perl-lsp 5-mode critique → two-line stubs; the `mode.rs:39-63`, `mode.rs:19-28`, `checkpoint_impl.rs:29-34` cites and the "POD rule unreachable" point moved into ch3 §3.9.4. ch5 §5.0's 6-state `ExpectKind` → `Expect`, "the eleven-state enum, ch3 §3.1". ch6 §6.3.1's 2-state `ExpectMode` plus `AfterSub`/`AfterArrow`/`AfterVarSub`/`HashBraceDepth`/`InPrototype` → the 11-state `Expect` + `BrackStack`. ch3 §3.10's third `LexState` struct → prose; its `LastLopOp` moved into ch6's struct. Prior-art notes (`docs/plans/…prior-art.md` A1.4, A1.5, A1.8.1): "five states, verbatim" → eleven, ch3 §3.1 | ch2 −40, ch3 −10, ch6 −4 |
| `{` block-vs-hash | **ch3 §3.3** | ch5 §5.11 entirely, including `leftCurlyIsHash` (which returned *block* for `{}` and *hash* for `$x =>` — both backwards); replaced by a 12-line stub with the two `[verified]` facts and the `perly.y:269` / `perly.y:1536-1537` cites. ch5 §5.12's call renamed to `looksLikeAnonHash` (ch4's name). ch5 §5.15's `isHashLiteral` paragraph, which said perl's statement-position answer for `{}` *differs* from PerlOnJava's (it does not: both say hash), rewritten to eight lines. ch4 §4.4.4's lossy paraphrase ("`word =>` or `word ,`") → cross-ref; ch4 §4.9.2 comment → "ch3 §3.3, verbatim" | ch5 −50, ch4 −2 |
| `__END__` / `__DATA__` / POD / shebang | **ch2 §2.5, §2.3.4** | ch5 §5.1.2's "must be at the start of a line" rule and the C excerpt in §5.1.3; §5.1.4 → one line. Kept in ch5: `File`/`DataSection`/`Pod` structs, the `toke.c:8295-8299` `yyl_fake_eof` cite (ch2 lacked it), the raw-bytes rule, and a one-line PerlOnJava divergence carrying the `DataSection.java:261-268` cite. Moved to ch2: `^D`/`^Z` fake EOF (`toke.c:9504-9508`, `DataSection.java:217`), the shebang `-M` note | ch5 −45 |
| Error recovery | **ch5 §5.13** | ch6 §6.8.3's `isSyncToken` + `recover` (no cap, no column-0 tier) → 8-line stub; its one rule ch5 lacked ("nearest sync point, never further") moved into ch5 §5.13.4. ch2 §2.14.4's four-item recovery list → one line | ch6 −38, ch2 −8 |
| Incremental machinery | **ch6** | ch5 §5.14 (133 lines: the perl-lsp cautionary tale, the invalidation list, `ParseCache`/`SafeRegion`/`Reparse`, the build-order recommendation) → 40 lines: the safe-boundary table only. Moved to ch6: `RecoveryDiagnosticsUnstable` (`incremental.rs:249-252`, `:5-7`, `CHECKPOINT_INTERVAL`, `MAX_INCREMENTAL_EDIT_BYTES`) into §6.1.3; `ContextSensitiveFormat` (`incremental.rs:164`) into §6.1.2 row A10. ch2 §2.14.1's `Checkpoint` struct → a bullet list of lexer-owned fields; §2.14.2 → two lines; §2.14 retitled "What the lexer must expose to a checkpoint". ch2 §2.14.3 (budgets, `Error`/`UnknownRest`) kept — lexer behaviour, not machinery | ch5 −93, ch2 −20 |
| Precedence | **ch4** | Nothing left to delete: the prior-art notes already say "32-level table ← `perly.y:150-182` (NOT PerlOnJava)". Confirmed by grep; ch2 §2.1.1 and ch5 §5.4.2 mention precedence only as token inventory and modifier precedence | 0 |
| "Build this first" lists | **ch6 §6.11** | ch2 §2.15 "ordered by a suggested build sequence" and item 3's "get this in early" → "acceptance list; build order is ch6 §6.11". Same one-line deferral added to ch3 §3.12, ch4 §4.16, ch5 §5.16, and the plan's checklist. ch6 §6.11 now names itself the only build order, with the oracle self-check test as step 0. Plan M2 gained one paragraph: it is the gate that flips §6.11 step 7's flag, not a reason to build incrementality earlier. Prior-art A1.6's five-row milestone table → pointer to the ladder | +12 net |

Net per file (spec): ch2 2,511 → 2,455; ch3 987 → 985; ch4 1,656 → 1,660;
ch5 3,204 → 3,028; ch6 1,845 → 1,838. ch7 was reshaped by the coordinator's
placement commit (`261c8aa9`), not by this pass.

## The dangling interface

ch6 §6.5.2 cited a `StatementBoundaries` predicate ch5 never defined. Chosen
fix: **name what exists.** ch5 §5.14 is now titled "Safe re-parse boundaries"
and holds only the table; ch6 §6.5.2 cites it by number and states that
`isReparseAnchor` is the node-kind projection of that table. `isReparseAnchor`
itself was cut from twelve kinds (including `expression_statement` and
`package_statement`, the latter unsafe in its `;` form per the table) to
`subroutine_declaration` and `source_file`.

## The statement-level call

ch5:2973 "Do not attempt statement-level incremental parsing" vs ch6:786
"Start with statement-level and move up on failure".

**ch5's position won, on the merits**, and the coordinator's note
(`r2-coordinator-note.md`) ruled the same way independently: parse costs
6-500 ms, `Analyze` 0.5-2 ms (§0.8), so per-statement bookkeeping buys a
fraction of the cheaper number; ch6 §6.11 itself says full-reparse first and
measure before optimising; and ch6's own preceding sentence argued for
sub-level anchoring. The ch6 sentence is deleted, not softened. ch6 §6.5.2
now anchors at the enclosing sub, falls back to a full reparse at top level,
and narrows to blocks only if §6.11 step 4 shows sub-level re-parse missing
the budget. ch5 keeps the position in the one place it survives — the table
row for "Single statement" — because ch5 §5.14.5 (a second build-order list)
was cut as a whole. ch6 §6.12's summary row changed to match.

## A contradiction round 1 missed — in an owner chapter

**ch3 §3.1.2 said `map {` reaches the lexer in `XTERMBLOCK`; ch3 §3.2.1's
worked example said the same. ch4 §4.9.2 said `XREF`.** `toke.c:8756` is
`LOP(OP_MAPSTART, XREF)` and `toke.c:8585` is `LOP(OP_GREPSTART, XREF)`;
`XTERMBLOCK` is set by `eval {` (`toke.c:8495`), `do {` inside a regex code
block (`toke.c:10114`), and the `XATTRTERM → XTERMBLOCK` path for an anon
sub's attribute list (`toke.c:6474`). In `yyl_leftcurly`, `XREF` falls to the
`default:` heuristic — which is why `map { $_ => 1 }` needs `+{` at all,
and is consistent with ch3 §3.3 step 7 mentioning `XREF`. ch4 was right;
the owner was wrong. Fixed in ch3 §3.1.2 (both the `XREF` and `XTERMBLOCK`
rows, with cites) and §3.2.1.

This is the same failure shape as round 1's: the copy that named a `toke.c`
line was right; the copy that did not was wrong. Since it was found and
fixed inside this round, it does not by itself force a round 3 — but it was
in the chapter every other chapter now defers to, so a round-3 reader should
spot-check ch3 §3.1.2's "set by" column line by line against `toke.c`.

## Not deleted, and why

- **ch2 §2.14.3 (byte budgets, `Error`/`UnknownRest`).** Lexer behaviour
  perl lacks, cited to `mode.rs:33`, `token.rs:69-70`, `token.rs:126`, and
  referenced by ch2 §2.15 and ch5 §5.13.7. Not incremental machinery.
- **ch2 §2.14.1's perl-lsp checkpoint divergence note** (`checkpoint_impl.rs:23-36`,
  `:14`, `:18`). ch2 §2.16 rows 14-16 cite it; it stays under the field list.
- **ch6 §6.8.2's missing-token insertion table.** Overlaps ch5 §5.13.3 but
  the brief named only §6.8.3, and the `IsMissing`/`HasError`-as-flag
  paragraph is incremental-specific.
- **ch5 §5.14's table row for "Single statement".** It is the language-side
  fact; ch6 decides what to anchor on.
- **ch6 §6.3.2's per-token `States` array** and the r1 stdlib findings
  (`slices.Replace`, `textproto`, timer `Reset`). Out of this round's brief.
- **Prior-art A1.6's Go line budget.** Measured; only its milestone table
  went. The prior-art file now lives in `docs/plans/` and its milestone
  table would have duplicated the conformance plan's ladder in the same
  directory — that duplication is the plan's problem, noted here, not fixed.
- **README chapter sizes.** Already rewritten to buckets by the placement
  commit; ch5's "~3.2k" and ch2's "~2.5k" still round correctly.

## Working-tree note

The placement commit `261c8aa9` landed while this pass was running and
captured the ch5 and ch6 edits above (HEAD's ch5 is 3,027 lines, ch6 1,837,
though `r2-placement.md`'s inventory still lists 3,204 / 1,845). The ch2,
ch3, ch4 edits, the follow-up ch5/ch6 lines, and the two plan-file edits are
uncommitted. Nothing here was committed by this pass.

## Is the document consistent on the six topics?

| Topic | Copies before | Copies after | Agree? |
|---|---:|---:|---|
| Expect states | 5 (ch2, ch3, ch5, ch6, A1) | 1 + 4 pointers | yes — every checkpoint condition now says `XSTATE` |
| `{` heuristic | 4 (ch3, ch4 ×2, ch5) | 1 + 3 pointers | yes — `{}` hash, `{ $x => 1 }` block, `{ foo, 1 }` block, everywhere |
| `__END__`/POD | 2 | 1 + pointer | yes — neither marker line-anchored |
| Error recovery | 3 (ch5, ch6, ch2) | 1 + 2 pointers | yes |
| Incremental | 4 (ch2, ch3, ch5, ch6) | 1 + table + 2 pointers | yes — sub-body first, full reparse before any of it |
| Precedence | 1 | 1 | yes |
