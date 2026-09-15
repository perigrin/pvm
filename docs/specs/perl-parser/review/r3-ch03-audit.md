<!-- ABOUTME: Round 3 of the ponytail review: claim-by-claim audit of ch3's PL_expect state machine against toke.c and perl 5.42.0. -->
<!-- ABOUTME: Row-by-row verdicts, what was changed in ch3, what remains unresolved, and whether ch3 can be trusted as the single owner. -->

# Round 3 — ch3 state-machine audit

Ground truth: `~/dev/perl5` at `94e5086608` (`toke.c`, `perl.h`,
`perly.y`); behaviour: `/home/perigrin/.local/bin/perl`, v5.42.0, via
`-MO=Deparse,-p` and `-MO=Concise`. Every line number below was read, not
grepped-and-assumed. Verdict key: **C** confirmed (cite verified),
**W** wrong, **U** uncited (true, cite added), **M** miscited (line exists,
says something else).

## Totals

| Verdict | Count |
|---|---:|
| CONFIRMED | 61 |
| WRONG | 11 |
| UNCITED | 7 |
| MISCITED | 56 |

135 claims audited (tallied from the per-section tables below; a row marked
"C + U" counts as UNCITED). The MISCITED count is dominated by §3.4.2's
weight table (18 of its 19 line numbers were off by 15–35 lines), §3.8.2
(11 of 13 off by 2–8 lines) and §3.8.1 (7 of 10). The WRONG count is the
one that matters; see "Is ch3 now trustworthy".

## §3.1.1 — the enum

| Claim | Verdict | Evidence |
|---|---|---|
| `perl.h:5972-5985`, eleven members, the listed names and order | C | `perl.h:5972` `typedef enum {` … `5985` `} expectation;` — 11 members exactly as listed. Listing lacked the in-enum comment `/* update exp_name[] … */`; added. |
| `exp_name[]` at `toke.c:5465-5469` lists `"SIGVAR"` with no enum member | C | `toke.c:5465-5469`; 12 strings, `"SIGVAR"` at index 10. **Consequence added**: `exp_name[XTERMORDORDOR]` prints `"SIGVAR"` under `-DT`. (Not run — the installed perl is not a DEBUGGING build; this is from reading the array against the enum.) |

## §3.1.2 — the eleven states (set-by column)

| State | Old claim | Verdict | Truth / evidence |
|---|---|---|---|
| `XOPERATOR` | `TERM()` `toke.c:254`; "end of any variable, string, number" | C + U | `254` ✓. Variable ends at `5793` (`yyl_dollar`), `7052` (`yyl_snail`), `6419` (`yyl_percent`), `6360` (`yyl_star`); strings `2762`/`2821` (`S_sublex_done`); `qw` `6144`; readline/heredoc `7180`. Added. **Exception added**: `yyl_dollar` `5797-5845` flips back to `XTERM` after `print $fh` when a term follows (`print $fh -1` → `print($fh (-1))`, verified). |
| `XTERM` | `OPERATOR()` `249`; "after any infix operator, after `(`" | C + U | `249` ✓. `(` is conditional: `7146-7149` sets `XTERM` only when `(` does not directly follow a list/unary op (`print(STDOUT 1)`). Added `UNI` `296`, `LOP(f,XTERM)`, `HASHBRACK` `6655`, `use` `5443`, `sort` `9041`. |
| `XREF` | `PREREF()` `253`, `${`, `@{`, `intuit_method`; `map`/`grep` `8756`/`8585` | C + U | `253` ✓; `8756` `LOP(OP_MAPSTART, XREF)` ✓; `8585` `LOP(OP_GREPSTART, XREF)` ✓. `PREREF` callers cited: `5699` `$`, `5687` `$#`, `7034` `@`, `6410` `%`, `6951` `&`, `6363` `*`. Added `print`/`printf`/`say` `8837`/`8841`/`8952` — the chapter's own worked example used `print` → `XREF` without a cite. `intuit_method` `5106`; `scan_ident` `11231`. |
| `XREF` question column | "`{` after `@` is a deref block, not an anon hash; `map {` falls through to the §3.3 heuristic" | C (refined) | Both true but for different reasons: `6719-6721` skips the heuristic when `PL_expect == XREF && PL_oldoldbufptr != PL_last_lop` (bare sigil); after `map`/`print` the `{` *does* run the heuristic. Verified: `@{ "foo", "bar" }` is a block even though `"foo",` would say hash; `map { "$_" => 1 } @l` is a syntax error because the heuristic says hash. |
| `XSTATE` | "After `;`, after `{` opening a block (`toke.c:6684`)" | M | `6684` is `PL_lex_allbrackets++` inside the `XATTRTERM`/`XTERMBLOCK` arm. `XSTATE` is set in every block arm: `6685`, `6691`, `6696`, `6837`, `6841`; `;` at `9669`; stack pop `6862`. Fixed. |
| `XSTATE` question column | "bareword + `:` is a label" | C | `9389-9396`. **Added**: `{` in `XSTATE` goes through the §3.3 heuristic — `{ a => 1 }` at statement start is a hash (verified: `+{'a', 1}`). |
| `XBLOCK` | `PREBLOCK()` `251`, `PHASERBLOCK` `255` | C + U | Both ✓. `PREBLOCK` callers enumerated: `else` `8475`, `continue` `8393`, `try`/`catch`/`finally` `9127`/`8371`/`8550`, `defer` `8443`, `default` `8439`, `package NAME` `8862`, `format` `5919`, `)` followed by `{` `7164`, `&`-proto sub `6636`. Also `5632` (signature `)`) and `6471` (`XATTRBLOCK` after `:`). |
| `XATTRBLOCK` | "`sub NAME` with attributes pending" | W | Set **unconditionally** after the name in `sub`/`method`/`format NAME` (`5879`, inside `if (isIDFIRST…)`) and after `class NAME` (`8382`). Nothing is "pending"; the state means attributes *may* follow. Fixed with cites. |
| `XATTRTERM` | "`sub {` with attributes pending" | W | Set when `sub`/`method` is **not** followed by a name (`5907`, the `else` arm). Same mischaracterisation. Fixed. |
| `XTERMBLOCK` | `eval {` `8495`; "`do {` in a regex code block" `10114`; `XATTRTERM`→ `6474` | C, but incomplete | `8495` ✓, `10114` ✓ (a `(?{…})` code block; the lexer forces a synthetic `KW_DO`), `6474` ✓. **Missing the most common source**: `do {` itself, via `PRETERMBLOCK` (`252`, caller `7514`). Verified `my $x = do { 1 }` parses as a block. Added, plus `13343` (format). |
| `XBLOCKTERM` | "`METHCALL0` path (`toke.c:8213`)" | M | `8213` is blank/comment; the assignment is `8222`, in the `8215-8224` block: bareword, no CV, followed by `$` or `{`, `indirect` on. Fixed. |
| `XPOSTDEREF` | "`->` handler" | U | `yyl_hyphen` `6285-6293`: set only when `->` is followed by one of the postfix-deref forms. Cite added; description tightened. Verified `$r->@*` → `@$r`. |
| `XTERMORDORDOR` | `FTST()` `toke.c:256` | M + W (incomplete) | `256` is `POSTDEREF`; `FTST` is `261`. Also set by `UNIDOR` (`297`) for `getc`/`pop`/`pos`/`readline`/`readpipe`/`readlink`/`shift`/`undef`/`umask` (`8594`…`9163`). Verified `shift // 0` → `(shift(@ARGV) // …)`. Fixed. |
| `XTERMORDORDOR` paragraph | "resolves `//` toward the operator and everything else toward a term" | C (refined) | All consumers listed now: `7060` (`//`), `7129` (`~~`), `6652` (`{` as in `XTERM`), `9905-9907` (v-string check). |

## §3.1.3 — the macro table

| Row | Old | Verdict | Truth |
|---|---|---|---|
| `TERM` 254 | | C | |
| `OPERATOR` 249 | | C | |
| `PREBLOCK` 251 | | C | |
| `PREREF` 253 | | C | |
| `Aop/Mop/BAop/SHop` 266-274 | | M | Operator macros span `265-278` (`BOop` at 265, `ChEop`…`NCRop` at 275-278). Fixed. |
| `FUN0` 262 | | C | `FUN0OP` 263, `FUN1` 264 added. |
| `UNI` 296 | | C | `UNI3` at 285 added. |
| `FTST` 256 | | M | `261`. Fixed. |
| `LOOPX` 252 | | M | `257` (252 is `PRETERMBLOCK`). Fixed. |
| `LOP` 2152 | | C | `S_lop` at 2174; **added** that `x` is applied only when `PL_nexttoke == 0` (`2186-2188`). |
| `POSTDEREF` 256, `S_postderef` 2227 | | C / M | Macro 256 ✓; `S_postderef` is at `2234` (2227 is its docblock); assignments `2244`, `2258`. Fixed. |
| "80 assignment sites, ~10 targets" | | W | `grep -c 'PL_expect *=[^=]'` = 98 (24 macro definitions + 74 sites); all 11 members are targets. Fixed. |
| Missing rows | | U | Added `PRETERMBLOCK` (252), `UNIDOR` (297), `UNIBRACK` (303 — **leaves `PL_expect` unchanged**, which is why `KEY_eval` sets it by hand at `8495-8505`), `UNIPROTO` (298), `OLDLOP` (382, `return` at `8893`). |

## §3.1.4 — the bracket stack

| Claim | Verdict | Truth |
|---|---|---|
| pushed in `yyl_leftcurly` `6643` | C | function start. |
| popped in `yyl_rightcurly` `6858` | M | function at `6853`; the pop is `6862`. Fixed. |
| context table `6653-6697` | C | `6650-6703` more precisely; fixed. |
| `XTERM`/`XTERMORDORDOR` → push `XOPERATOR`, `HASHBRACK` | C | `6651-6655`. |
| `XOPERATOR` → push `XOPERATOR` | C | falls through to `6682-6686`. |
| `XATTRTERM`/`XTERMBLOCK` → `XOPERATOR` | C | `6682-6686`. |
| `XATTRBLOCK`/`XBLOCK` → `XSTATE` | C | `6688-6692`. |
| `XBLOCKTERM` → `XTERM` | C | `6694-6697`. |
| default → `XTERM` if after list op else `XOPERATOR` | C but incomplete | `6700-6703` ✓; **the entry is overwritten with `XSTATE` at `6840`** when the heuristic decides "block" in a non-`XREF` state. Added, plus the inside-the-braces expectation for each arm. |

## §3.2 — the disambiguation table

| Row | Verdict | Notes |
|---|---|---|
| `/` `yyl_slash` 7058 | C | `7060` for the `//` test. |
| `?` 9810 | C | `OPERATOR(PERLY_QUESTION_MARK)` regardless of state. |
| `{` "`XBLOCK`/`XSTATE`: block; `XREF`: deref block" | W | `XSTATE` is not a block state; it is a `default` state and runs the heuristic (`{ a => 1 }` at statement start is a hash, verified). Fixed to list the five block states and the three heuristic states. |
| `}` 6853 | C | |
| `<` 7169, heredoc condition | C | `7176-7177`. |
| `>` 7221 | C | |
| `*` 6353 | C | |
| `%` 6391 | C | |
| `&` 6901 | C | |
| `+` 6325 | C | |
| `-` 6200, filetest condition | C | `6202`; added the `=>` escape at `6213`. |
| `~` "`~~` smartmatch if `XOPERATOR`" | W (for 5.42) | `7128-7129`: also `XTERMORDORDOR`, and gated on `FEATURE_SMARTMATCH_IS_ENABLED`. Verified: `$x ~~ 1` lexes by default, is a syntax error under `use v5.42`. Fixed. |
| `(` 7144 | C | `7146-7149`. |
| `:` 6456 | C | |
| bareword 8023 | C | |

## §3.2.1 — worked examples

| Example comment | Verdict | Fix |
|---|---|---|
| `sub f { 1 }  # sub NAME -> XBLOCK` | W | `sub NAME` → `XATTRBLOCK` (`5879`), and `yyl_leftcurly` treats `XATTRBLOCK` like `XBLOCK`. Fixed. |
| `print {$fh} "x"  # print -> XREF -> deref block` | C (refined) | It is `XREF` *after a list op*, so the heuristic runs and says block. Reworded. |
| `map { $_ } @l`, `{; $_ }`, `+{ a=>1 }` | C | All verified. |
| Added | — | `{ a => 1 }` (hash), `{ $x => 1 }` (block) at statement start; `my $h = { $x => 1 }` (hash, no heuristic in `XTERM`); `map { "$_" => 1 } @l` (syntax error); `@{ "foo", "bar" }` (block). All verified. |

## §3.3 — the `{` heuristic

| Step | Old cite | Verdict | Truth |
|---|---|---|---|
| range 6698-6840 | | C | `6698-6842`. |
| 2: `}` → hash, 6714 | | C | `6706-6715`; `6707` is the `${}`-in-string error. |
| (missing) `XREF` not after list op → skip heuristic | | W (omission) | `6719-6721`. **This changes the stated rule**: as written, steps 3–6 applied to `@{ "a", "b" }` and would call it a hash. Perl calls it a block (verified). Inserted as step 3. |
| 3: quotes 6742-6748 | | M | `6740-6746`. |
| 4: `q`-forms 6750-6786 | | M | `6747-6797`, incl. `q =>` immediate hash at `6760-6762` and the plain-`q`-word branch `6789-6797`, neither of which the chapter mentioned. |
| 5: "else scan a bareword" | | W (imprecise) | Only if the first char is a word char (`6799-6806`); otherwise the pointer does not move. This is *the* reason `{ $x => 1 }` is a block, and the old text did not make the rule produce that outcome. Fixed. |
| 6: comma/`=>` 6806-6808 | | M | `6810-6814`. Condition text ✓. |
| 7: `XREF` term/statement 6810-6829 | | M | `6815-6838`; term at `6824`/`6833`, statement `6837`. |
| 8: block | | U | `6839-6841`, incl. the stack-entry rewrite. |
| comment quote 6722-6733 | | M | `6723-6737`. |
| lowercase asymmetry | | C | Verified: `{ foo, 1 }` block, `{ Foo, 1 }` hash. |

Does the stated rule now produce Round 1's outcomes? `{}` → step 2 → hash ✓.
`{ $x => 1 }` (statement start) → step 6 pointer stays on `$` → step 7 sees
`$`, not `,`/`=>` → step 9 block ✓. In `XTERM` it never reaches the
heuristic and is a hash — both verified.

## §3.4 — `intuit_more` (audited because its cites were in scope of "in ch3 only")

| Claim | Verdict | Truth |
|---|---|---|
| `S_intuit_more` 4553-4913 | M | docblock `4553`, function `4576-5041`. |
| called from `yyl_snail` 7038, `yyl_percent` 6410 | M + W | `7039` ✓; `6410` is `PREREF(PERLY_PERCENT_SIGN)` — the call is `6413`. Also called from `yyl_dollar` `5707` (omitted) and `10129`, `11103`. Fixed. |
| decision steps 4598-4650 | C | `4598-4649`. |
| symbol-table branch 4653-4688 | C | `4653-4695`. |
| weight table, 19 rows | M ×18 | Every line but `4948` was wrong. Re-cited from the source: 4698, 4711, 4728, 4732, 4762, 4807, 4832, 4844, 4846, 4863, 4869 (new `+1` row the chapter lacked), 4896, 4899, 4918, 4928, 4934, 4943, 4948, 4987, 5003, 5016, verdict `5039`. Values (Δweight) were all correct. |
| "−100 at 4770" | M | `4807`. |
| §3.4.3 `%` rule | C (+`XPOSTDEREF` added) | `6391-6423`. |
| §3.4.4 `sort` short-circuit 8153-8156 | M | `8156-8161`; and the condition also excludes `map`/`grep` (`8160-8161`). Fixed in §3.8.2 row 5. |

## §3.5 — prototypes

| Claim | Verdict | Truth |
|---|---|---|
| §3.5.1 Concise demonstration | C | Re-run: `f(@a)` forward → `padav lM`, no `srefgen`; `f @a` backward → `srefgen`. |
| `$` forces scalar | C | `sub f($){} f @a` → Concise `padav … sM`. |
| `*` "bareword is **not** stringified" | W | `sub f(*){…} f STDIN` under `strict`: `ref \$_[0]` is `SCALAR`, value `STDIN`. The bareword *is* a plain string; what `*` buys is that it is accepted under `strict subs` and not resolved as a call. Fixed. |
| `&` first → block without comma | C | `f { 1 } @l` → `&f(sub {…}, @l)`. |
| `_` defaults to `$_` | C | `f($_)`. |
| `+` | C | `f(@a)`, `f($s)` both compile. |
| `yyl_subproto` 6588-6640 | C | Added its only caller `7973` (`yyl_constant_op`) and `TERM(FUNC0SUB)` `6596` for the empty prototype. |
| UNIPROTO condition 6603-6615 | C | `6603-6614`. |
| `perly.y:169` `%nonassoc UNIOP UNIOPSUB`, between comparison and shift | C | `167` CHRELOP, `169` UNIOP, `171` SHIFTOP. Verified `f $a + 1` → `f($a+1)`, `f $a > 1` → `f($a) > 1`, `f $x, $y` → `(f($x), $y)`. |
| `S_intuit_method` proto check 5079-5087 | M | `5087-5094`. |
| §3.5.4 "falls through to the default list-operator parse … no error … bug-compatible for every forward call" | **W** | Verified: `f $x, 1` with unknown `f` → `($x->f, 1)` (indirect method call); `f bar` → `'bar'->f`; `f @a` and `f 1, 2` → *"found where operator expected (Do you need to predeclare "f"?)"*, a compile error; only `f(@a)` is an unprototyped call. Under `use v5.36` the method rows become syntax errors too. Rewritten as a table with cites `8195-8212`, `8177-8193`, `8214-8224`, `8227-8250`, `8261-8271`, and the static-parser consequence restated: paren form is exactly perl; paren-less form to an unknown name has *no* list-operator reading in perl and must carry a diagnostic. |
| `c.cv` lookup 8107-8117 | M | `8113-8119`. |

## §3.6 – §3.8 (cite spot-checks)

| Claim | Verdict |
|---|---|
| keyword plugin 9341-9362 / infix 9366-9385 | M → `9338-9361`, `9363-9387`. |
| `S_new_constant` `call_sv` 10568 | C |
| `use constant PI - 1` "a parser may read `PI(-1)`" | W (imprecise): perl with `PI` unknown reads `'PI' - 1` (bareword then `XOPERATOR`, `8174`); only a parser that treats unknown barewords as list ops reads `PI(-1)`. Fixed, with `6596`. |
| `Perl_filter_add` 5164-5230, `filter_read` 5268, `filter_gets` 5387-5395 | C |
| §3.8.1 rows 1-10 | C ×3, M ×7 (9336→9333-9335, 9341→9338-9361, 9366→9363-9387, 9430→9436, 9436→9441-9453, 9455→9456; row 10 goes to `yyl_word_or_keyword`, not directly to `yyl_just_a_word`). Fixed. |
| §3.8.2 rows 1-13 | C ×2, M ×11 (all within 2–8 lines; re-cited as ranges). Row 4 gained the "directly after the list/unary op" precondition (`8122-8126`); row 5 gained the `map`/`grep` exclusion. |
| §3.8.3 5046-5142, gate 5071, `isUPPER` 5100 | C (function ends `5144`). |
| §3.9.4 `6603-6615`, `5472` | C |

## What was changed in ch3 (39 edits)

- §3.1.1: enum listing completed; `SIGVAR` consequence.
- §3.1.2: all eleven set-by cells rewritten with verified cites; two states
  (`XATTRBLOCK`, `XATTRTERM`) re-described; `do {` added to `XTERMBLOCK`;
  `UNIDOR` added to `XTERMORDORDOR`; `XSTATE` question column now says `{`
  goes through the heuristic.
- §3.1.3: 4 wrong line numbers fixed; 5 missing macros added; the `LOP`
  `PL_nexttoke` caveat; the 80/10 count replaced with the measured 98/24/11.
- §3.1.4: pop line; default-row rewrite to `XSTATE`; inside-brace expectation.
- §3.2: `{` row corrected (`XSTATE` is a heuristic state); `-` and `~` rows.
- §3.2.1: `sub NAME` comment; five new verified examples.
- §3.3: renumbered to nine steps with the `XREF` skip inserted and the
  "pointer does not move" rule made explicit; all cites replaced; verified
  outcomes listed.
- §3.4: caller list, all 19 weight cites, one missing weight row.
- §3.5: `*` row; `yyl_subproto` caller and `FUNC0SUB`; §3.5.4 rewritten.
- §3.6, §3.8: cites.

No restructuring; no cuts. Net +72 lines.

## Unresolved

- **`SIGVAR` under `-DT`**: stated from reading `exp_name[]` against the
  enum; not run, because the installed 5.42.0 is not a DEBUGGING build.
- **§3.4.2 weights** were checked for line numbers and Δ values, not
  behaviourally (that is the regex-charclass heuristic; a corpus diff is the
  right test and is already on the §3.12 checklist).
- **Why were the §3.4.2 and §3.8.2 cites systematically off** by 15–35 and
  2–8 lines while §3.1/§3.2's function-start cites were exact? The pattern
  is consistent with the tables having been transcribed against a slightly
  different `toke.c` revision than the one Round 2 pinned (`94e5086608`).
  Every chapter-4/5/6 cite into these regions should be assumed stale until
  checked — out of scope for this round.

## Is ch3 now trustworthy as the single owner?

**For the state machine itself — §3.1.2, §3.1.3, §3.1.4, §3.2, §3.3 — yes,
with this round's edits.** Every set-by cell and every transition now cites
a line I read, and the behavioural claims that distinguish the states
(`{` in `XSTATE` vs `XTERM` vs `XREF`, `//` after `shift`, `sub NAME` →
`XATTRBLOCK`) were run against 5.42.0.

**But the error rate before this round was not low.** Nine claims were
wrong, not merely miscited, and three of them were the kind that propagate:

1. §3.2's "`XSTATE`: block" — a Go parser built on that would turn every
   statement-level `{ a => 1 }` into a block, the opposite of perl.
2. §3.3's algorithm, as stated, called `@{ "a", "b" }` a hash.
3. §3.5.4's "forward calls fall through to a list-operator parse" — the
   chapter's headline "single most important fact" was false for every
   paren-less form.

None of those was caught by Round 1's example-level verification, because
the *examples* were right and the *rules* were wrong. That is exactly the
failure mode the brief warned about.

**Recommendation: one more bounded round, not on ch3.** Ch3's state
machine has now been read line-by-line. What has *not* been is the
downstream consumers: ch4/ch5/ch6 code that was rewritten in Round 2 to
"defer to ch3 §3.1/§3.3" — in particular ch5's `looksLikeAnonHash` stub and
ch6's `Expect`/`BrackStack` struct — should be re-checked against the
*corrected* §3.3 (nine steps, `XREF` skip, pointer-does-not-move) and the
corrected `XSTATE` row, because they were aligned to the version that was
wrong. That is a diff-against-ch3 audit, not another read of toke.c.
