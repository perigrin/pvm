<!-- ABOUTME: Round 4 of the ponytail review: the consumers of ch3's corrected state machine (ch2, ch4, ch5, ch6, ch7, findings, plans) diffed against ch3 as rewritten in round 3. -->
<!-- ABOUTME: What disagreed, what changed, what was checked and already consistent, and whether the specification has reached its fixed point. -->

# Round 4 — consumers vs the corrected ch3

Brief: round 3 rewrote ch3 §3.1–§3.3 and §3.5.4; round 2 had aligned ch4/5/6
to the *pre*-rewrite text. This round diffs every consumer against the
current ch3. Ground truth is the text of `03-lexer-feedback.md`, not
`toke.c`. The exceptions: two cite reconciliations (`toke.c:9034-9043`,
`perly.y:350-362`, both against `94e5086608`) and five one-line `Deparse`
runs on 5.42.0 for claims ch3 does not cover. Both are listed below.

## Tally

| | Count |
|---|---:|
| Claims checked against ch3 (state names, `{` sites, checkpoint conditions, prototype/forward-call prose, cites into ch3's regions) | 52 |
| Behavioural contradictions with ch3 (fixed) | 3 |
| Structural omissions — a struct or list missing a field ch3 now requires (fixed) | 3 |
| Misleading-but-not-wrong prose (fixed) | 4 |
| Stale cites into ch3-owned regions (fixed) | 7 cells |
| Already consistent | 35 |

Nine files edited, +33/−24 lines. No section moved or cut; no `[verified]`
marker touched; no measured fact removed.

## What disagreed, and what changed

### Behavioural contradictions

1. **`00-findings.md` §0.3** said perl's not-yet-seen-sub behaviour is "the
   conservative default" and a single-pass parser is "already
   bug-compatible with perl on forward calls". ch3 §3.5.4 (round 3): `f(@a)`
   is an unprototyped call, `f $x, 1` is an indirect method call, `f @a` is
   a compile error. **Edited 00-findings.md** — the brief allowed this only
   for an outright contradiction, and this is one. Rewritten to the §3.5.4
   rule with a cross-reference; the srefgen measurements above it are
   untouched.
2. **`07-conformance.md` §7.7.3** said a prototype edit "changes how calls
   parse — possibly earlier in the file … propagates *backwards*", and that
   perl "allows a prototype to affect calls textually before the declaration
   in some arrangements". ch3 §3.5.4/§3.5.5: a prototype applies only to
   calls textually after the definition; forward calls have no
   list-operator reading. Row and paragraph rewritten: the edit's effect is
   *forward* but escapes the enclosing re-parse region, which is the point
   the test exists to prove. The test itself is unchanged.
3. **`05-statements.md` §5.5.5** prototype table: `*` → "glob". ch3 §3.5.2
   (round 3, verified `ref \$_[0]` is `SCALAR`): a bareword arrives as the
   plain string. Cell rewritten with the cross-reference.

### Structural omissions

4. **`06-incremental-lsp.md` §6.3.1 `LexState`** claims to be "everything
   the lexer needs to resume". The corrected ch3 reads *whether the previous
   token was a list operator* in three places it did not before: §3.1.4's
   default row (`PL_oldoldbufptr == PL_last_lop`), §3.3 step 3 (the `XREF`
   skip), and §3.1.2's `print $fh -1` exception; and `PL_last_uni` for `(`.
   Round 2 added `LastLopOp` (an opcode, for §3.4.4) but that does not
   answer "did a list op *just* precede this token". Added `AfterLop` and
   `AfterUni bool` — edit-invariant, unlike a byte offset, so `Equal` still
   works across an edit. The safe-checkpoint condition (`XSTATE`, all stacks
   empty) is unchanged: after `;` nothing "directly follows" a list op.
5. **`02-lexical-structure.md` §2.14.1** checkpoint field list: same gap;
   extended the `PL_last_lop_op` bullet by one clause.
6. **`05-statements.md` §5.0 `ParseState`** carried `Expect` and said in
   prose "carry [the bracket stack] unchanged", but had no field for it.
   Added `BrackStack []Expect`.

### Misleading prose

7. **`04-expressions.md` §4.4.4** example comment `map { $_ => 1 } @list
   # ambiguous! perl guesses BLOCK, this is a bug`. Per ch3 §3.3 step 6 the
   `$` leaves the scan pointer in place, step 7 sees `$`, and the answer is
   *block* — which is what the programmer wanted; the guess is right, not a
   bug. The failing case is `map { "$_" => 1 }` (ch3 §3.2.1, verified).
   Comment rewritten to say exactly that.
8. **`04-expressions.md` §4.3** precedence row 16 listed `~~` with no
   qualifier. ch3 §3.2: lexes only while the `smartmatch` feature is on,
   off under `use v5.42`. One clause added to the note column. (`~~` also
   appears in the operator table at §4.3's `NCEQOP` row and the Go map at
   line 263 — left, since they describe the token when it does lex.)
9. **`05-statements.md` §5.3.4** "the grammar sets `parser->expect =
   XTERM` after each semicolon" read as contradicting ch3 §3.1.2's "`;` →
   `XSTATE` (`toke.c:9669`)". Both are true: the lexer sets `XSTATE`, the
   C-style-`for` mid-rule actions at `perly.y:354` and `:359` (cites read)
   overwrite it. Added "overriding the `XSTATE` … (chapter 3 §3.1.2)".
   Confirmed on 5.42.0: `for (;/x/;) { last }` deparses as `while (/x/)`.
10. **`plans/2026-09-05-parser-prior-art.md` A1.4** `{` row recommended
    "perl-lsp's" with no statement of whose *rule* to implement, while ch3
    §3.3 says copy `toke.c`'s nine steps verbatim. Clarified: perl-lsp's
    *shape* (lexer-fed, real string boundaries), `toke.c`'s rule, naming
    the two steps round 3 added.

### The one owner-side edit — `sort`

ch3 §3.1.2 listed `sort` under `XTERM` (`9041`); ch4 §4.9 says `sort` is
`LOP(OP_SORT, XREF)` (`9043`) and §4.9.2 runs the `{` heuristic after it. A
reader of ch3 alone would conclude `sort {` arrives in `XTERM` and is an
unconditional `HASHBRACK`, which is absurd. I read `toke.c:9038-9043`:
`PL_expect = XTERM` at 9041, `force_word` at 9042, `LOP(OP_SORT,XREF)` at
9043. By ch3's own §3.1.3 `LOP` caveat, `XREF` wins unless `force_word`
queued a comparator name. Both cites were real; the ch3 row was true but
incomplete in a way that contradicted ch4. **Edited ch3 §3.1.2**: the
`XTERM` cell now says what overwrites it, and `sort (9043)` joins the
`XREF` row's `LOP` list. Verified on 5.42.0, since ch3 had no `sort {`
example: `sort { $a <=> $b } @l` → block; `sort { "a", 1 } @l` and
`sort { a => 1 } @l` → *syntax error near "} @l"* (heuristic said hash).
ch4 §4.9.2 was right.

### Stale cites into ch3-owned regions

| File | Was | Now (per ch3) |
|---|---|---|
| ch4 §4.4.4 | `yyl_leftcurly` `6719-6730` | `6643`; heuristic `6698-6842` |
| ch4 §4.16 table | heuristic `6719-6730` | `6698-6842` |
| ch4 §4.9, §4.16 table | `LOP(OP_GREPSTART)` `8584`, `LOP(OP_MAPSTART)` `8755` | `8585`, `8756` (off by one — the `case` line, not the `LOP`) |
| ch4 §4.4.6, §4.16 table | `XPOSTDEREF` `6285-6296` / `6285-6300` | `6285-6293` |
| ch5 §5.15 | comma rule `toke.c:6806` | `6810-6814` |

## Checked and already consistent

The brief's two named worries were both clean:

- **ch5 §5.11 stub and §5.12 skeleton.** `case LBRACE: if
  !p.looksLikeAnonHash()` at statement start is exactly what the corrected
  §3.3 says (`XSTATE` runs the heuristic). Had round 2 encoded the old
  "`XSTATE`: block" row, this would have been `return finishBlockStmt`
  unconditionally; it did not. Both `[verified]` facts (`{}` hash,
  `{ $x => 1 }` block) match ch3 §3.3's verified list; `toke.c:6714` lies
  inside ch3's `6706-6715`; `perly.y:269`, `1536-1537` unchanged.
- **ch6 §6.3.1 `Expect`/`BrackStack`.** Eleven states, stack of `Expect`,
  cross-refs to §3.1/§3.1.4; the safe-checkpoint condition is `Expect ==
  XSTATE` with `BrackStack`, `SublexStack`, `HeredocQueue`,
  `DelimiterStack` empty — correct against §3.1.4 (a `}` at depth 0 pops
  back to `XSTATE` only if its entry was `XSTATE`, which the empty-stack
  condition guarantees). §6.5.3's `st.Equal` rationale names no state.
  `initialLexState()` is not defined in text; perl starts in `XSTATE` and
  nothing in ch6 says otherwise.

Also consistent: ch4 §4.9.2 prose and code (`XREF` after `map`, heuristic
runs, one shared predicate); ch4 §4.4.6 `XPOSTDEREF` trigger list matches
§3.1.2; ch4 §4.8.1 paren cliff (`S_lop`/`UNI3`) matches §3.1.3; ch4 §4.16
checklist. ch5 §5.0 `BEGIN`-at-`XSTATE` table row, §5.5.3 (`perly.y`
sets `XATTRBLOCK` after a signature — a grammar action, a different layer
from ch3's `toke.c:5632` `XBLOCK` at the `)`; both cited, not in
conflict), §5.5.5 "only already-declared subs", §5.5.8 `AUTOLOAD` guard,
§5.7.2 `class` → `XATTRBLOCK` (`8382`, same line as ch3), §5.9.1 `use` →
`XTERM` (`5443`), §5.14 (names no state), §5.18. ch2 §2.2.1, the five POD
`XSTATE` guards (§2.5.1–§2.5.3, §2.15 item 12, §2.16 row 10, §2.17
item 2), §2.13's `XTERMORDORDOR` mention, §2.14.2. Prior-art A1.4 regex
row, A1.5, A1.8.1 (all "eleven states, ch3 §3.1"). The conformance plan
names no expectation state and makes no `{` or forward-call claim; its
prototype milestone (M4) and oracle test agree with §3.5.5. README. ch7
elsewhere.

Not audited: `toke.c` line numbers in ch2/4/5/6 that point *outside* ch3's
corrected sections (§3.4.2, §3.8 and the lexer proper). Round 3 warned
those may be off by 2–35 lines from a different `toke.c` revision; the
`8584`/`8755` off-by-ones above are consistent with that warning. That is
a mechanical cite sweep, not a consistency question.

## Is the specification internally consistent?

**On the topics the review set out to converge — the expectation machine,
the `{` heuristic, checkpoint conditions, and now forward calls — yes,
with this round's edits, and I believe this is the fixed point.** Evidence:

- Every file that names an expectation state, calls the `{` predicate, or
  states a checkpoint condition was read at each hit (grep for all eleven
  state names, `looksLikeAnonHash`, `BrackStack`, `Expect`, `checkpoint`,
  `~~`, `forward`, `predeclare`, `bug-compat`, `prototype`), and 35 of 52
  hits needed nothing.
- The three behavioural contradictions all trace to **one** round-3
  rewrite — §3.5.4 forward calls — whose consumers (findings §0.3, ch7
  §7.7.3, and by extension ch5's `*` row from §3.5.2's companion fix) round
  3 did not touch. That is a single missed propagation, not a diffuse error
  rate. The state-machine consumers, which the brief expected to be broken,
  were not: round 2 had aligned them to the predicate's *name*, not to the
  wrong row's *content*.
- The one owner-side problem (`sort`) was a true-but-incomplete cell, found
  because two real cites disagreed, and is now reconciled with a perl run.

What would change my mind: a round-3-style line-by-line audit of ch4 or
ch5 against `toke.c` finding a *rule* stated wrongly with correct examples,
the failure mode round 3 named. Nothing in this round's reading suggested
one, but this round read for agreement with ch3, not for truth against
perl. If a round 5 is run, it should be that cite sweep — scriptable: for
every `toke.c:NNNN` in ch2/4/5/6, print the enclosing function name and
compare with the prose — rather than another consistency pass.
