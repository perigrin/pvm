<!-- ABOUTME: Ponytail review round 1, findings from the coordinating reviewer. -->
<!-- ABOUTME: Covers README, ch0, ch1 and cross-file staleness checks. -->

# Round 1 — coordinator's own findings

Scope: `README.md`, `00-findings.md`, `01-scope.md`, plus staleness checks that
span every file. The four large chapters are with the Fable reviewers.

## Findings

| § | Finding | Rung | Action | Saves |
|---|---|---|---|---|
| README index table | The `Lines` column is stale in 5 of 9 rows, worst by 609 (ch5: says 2,595, is 3,204). It drifts on every edit and no reader acts on an exact line count. | 4 | SHORTEN — replace with a size bucket (`~1k`, `~3k`) or drop the column | ~0 lines, removes a maintenance burden |
| README total | "Roughly 12,700 lines" — actual is 13,560. Same drift. | 4 | Round harder ("~13k") or drop | 0 |
| ch1 §1.7 sources table | Repeats the three-source table already in `README.md`. Two copies to keep in sync. | 2 | CROSS-REF — README owns it | ~10 |

## Verified clean

Checks that could have found rot and did not:

- **No chapter still quotes the wrong corpus ceiling.** `66.3%` / `411/620`
  appears only in `00-findings.md`, where §0.5 documents the error on purpose.
- **No chapter still claims PSC is the latency bottleneck.** The one match in
  `06-incremental-lsp.md:1690` is the corrected negation.
- **ch1 carries no numbers at all**, so it is not a second copy of ch0's
  evidence. That separation is doing its job.

## The three things I would actually do

1. **Drop the `Lines` column from the README table.** It is wrong in 5 of 9
   rows already, one day after being written. Exact counts are metadata that
   rots; nobody chooses what to read based on 1,652 versus 1,647.
2. **Cut ch1's duplicate sources table**, leaving the README's. Ten lines, but
   it is the kind of duplication that later drifts into a contradiction.
3. **Nothing else at this level.** ch0 and ch1 are short and load-bearing. The
   real questions are chapter-scale and belong to the other reviewers.

---

# Verification of the ch2 reviewer's claims

The ch2 reviewer made four claims sharp enough to check. All four hold.

| Claim | Verdict | Evidence |
|---|---|---|
| Any `isGRAPH_A` char after a sigil is a variable; the 125-line table is the wrong shape | **HOLDS** | `@;` and `%;` are real variables — `@; = (1,2)` gives 2 elements, `%; = (a=>1)` gives 1 key |
| `$#` is fatal, not deprecated (the chapter says both) | **HOLDS** | `$# is no longer supported as of Perl 5.30` |
| ch2 is right and ch5 wrong on `__END__` not being line-anchored | **HOLDS** | An indented `  __END__` still terminates the file — code after it does not run |
| `strconv` subsumes the hand-rolled numeric scanner | **HOLDS** | `ParseInt(s,0,64)` takes `1_000`, `0b1010`, `017`, `0o17` and rejects `1__0`, `1_`; `ParseFloat` takes `1_000.5`, `0x1.8p1`, `5.`, `.5`, `10E10` |

One refinement to the last: `ParseInt("0x1p4")` fails, because a hex float is a
float. Perl gives `0x1p4 == 16` and `ParseFloat` agrees, so the rule is "try
`ParseInt` with base 0, fall back to `ParseFloat`" — two lines, not 57.

The `__END__` finding is the important one: it is a **cross-chapter
contradiction where one side was transcribed from PerlOnJava rather than
measured**. That is exactly the failure the spec's own "perl wins" rule exists
to catch, and it survived into the document anyway.

---

# Verification of the ch5 reviewer's claims

Three claims, all checkable, all holding.

| Claim | Verdict | Evidence |
|---|---|---|
| §5.11's `{` rule is backwards | **HOLDS** | `perl -w -MO=Deparse -e '{}'` warns `Useless use of anonymous hash ({})` and deparses `{};` — a HASH. `{ $x => 1 }` deparses as a BLOCK containing `$x, '???'`. ch5 says the reverse of both. |
| `__DATA__` is not line-anchored either | **HOLDS** | `print "x\n";  __DATA__` mid-line and indented still terminates — only `x` prints |
| ch6 references a ch5 predicate that does not exist | **HOLDS** | `06:740` says "Use chapter 5's `StatementBoundaries` predicate"; `grep -c StatementBoundaries 05-statements.md` = **0** |

## Why these three matter more than the line counts

The `{` finding is the most serious defect the review has produced. It is not
bloat or duplication — it is **a rule an implementer would encode directly, and
it is wrong**. The spec even ships Go code for it (`leftCurlyIsHash`,
`05:2361-2379`) implementing the wrong rule. ch3 §3.3, citing `toke.c:6714`,
has it right; the two chapters contradict and the wrong one carries the sample
code.

The `__END__`/`__DATA__` pair is the same error twice, and the ch2 reviewer
found the first half independently. Both times ch5 transcribed PerlOnJava
instead of measuring, and both times ch2 measured. The spec's "perl wins" rule
would have caught either if anyone had applied it across chapters rather than
within one.

The dangling `StatementBoundaries` reference is the parallel-authoring failure
in pure form: ch6 was written against an interface it assumed ch5 would define,
and nobody checked.

**Pattern across both reviews so far:** every cross-chapter contradiction found
resolves in favour of the chapter that ran perl, and against the chapter that
read another implementation. That is one rule, applied consistently, and it
would catch all of them.
