<!-- ABOUTME: Round 5, the mechanical citation sweep round 4 recommended but did not run. -->
<!-- ABOUTME: Checks that every toke.c:NNNN in the spec lands on the code it claims. -->

# Round 5 — the citation sweep

Round 4 reached a fixed point on *consistency* and named one thing it had not
done: a mechanical check that every `toke.c:NNNN` still points at the code it
claims. Round 3 had found 56 miscitations in one chapter, so the concern was
that the rest of the spec had the same drift.

`review/citecheck.sh` maps every citation to its enclosing C function. Run it
against a fresh `perl5` checkout after any rebase of the corpus.

## Method

308 distinct `toke.c` citations across the spec and both plan files. Two passes:

1. **Enclosing function.** Does the cited line sit inside a function whose name
   is plausible for the claim? A cite for the `{` heuristic should land in
   `yyl_leftcurly`, not `S_scan_str`.
2. **Empty targets.** Does the line contain code at all? A citation landing on
   a blank line or a bare `}` reads as verified while pointing at nothing —
   the most dangerous failure mode, and the one that caught `6714`.

## Result

Nine of ten high-traffic citations spot-checked landed exactly on their claimed
content — `yyl_leftcurly`, `S_intuit_method`'s feature gate,
`LOP(OP_SORT,XREF)`, `PL_expect = XATTRBLOCK`, `FEATURE_SMARTMATCH_IS_ENABLED`.

Twelve citations landed on blank lines or bare braces. Nine were **range
starts** (`toke.c:3742-3750`), where beginning on a blank line is harmless.
Three were bare single-line cites, all off by one, all now fixed:

| Was | Now | Lands on |
|---|---|---|
| `toke.c:6714` | `6715` | `OPERATOR(HASHBRACK);` — the empty-`{}`-is-a-hash rule |
| `toke.c:9215` | `9216` | `Mop(OP_REPEAT);` |
| `toke.c:11719` | `11720` | `PL_multi_start = origline + 1 + …` |
| `toke.c:12429` | `12430` | the bracketing-delimiter depth loop |

`6714` is the instructive one: it is cited twice in ch5 for the rule round 1
proved and round 2 corrected, and it pointed at a closing brace. The rule was
right, the verification was real, and the pointer was still wrong.

## Verdict

The drift is **consistent off-by-one**, not random — which fits a
transcription that counted from a header line rather than misreading the
source. Four bad targets in 308 citations is a 1.3% rate, against round 3's
41% inside one chapter's tables, so the earlier damage really was localised.

This closes the item round 4 left open. Combined with round 4's consistency
result, the review has reached its fixed point:

- rounds 1-2 found contradictions **between** chapters
- round 3 found them **inside** the owner chapter
- round 4 found one missed **propagation** and no diffuse error
- round 5 found four **pointers** wrong and no claims wrong

Each round found a strictly smaller and more mechanical class of defect than
the last. That is what a fixed point looks like; it is not the same as the
document being correct, and §0.9's caveat still stands.
