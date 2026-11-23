# crochet:assess round 3: codebase lens

Subject: `docs/plans/2026-11-23-chart-precedence-in-recognizer.md` at 3e4d0c43b, blob **a8960d05a84284fb0f9fcec803d101949fd9c405** (`git hash-object`). This is revision 3, which forbids hash maps on `forest.Constrain`'s evaluation path and requires identical output. I also read the minute's "Measurement after round 2" section. Read-only.

Code read: `forest/precedence.go` (`NewPrecedence`, `lookaheads`, `walk`, `resolve`, `reduces`, `shifts`, `Constrain`, `constrainer.eval/seq/mark`); `forest/forestdigest_test.go`; `resolve_test.go`'s `built`; `forest_test.go`'s `samplePaths`; `forest/precedence_test.go`; `chart/precedence.go` (the referee); README l.505-520 and l.636-665 (the timing protocol).

Measured with an instrumented copy of `precedence.go` swapped in through `go test -overlay` (the worktree is untouched; files in `.../assess-w3/codebase/ov3/`).

## 1. Disposition of open rows

| Id | Disposition |
|---|---|
| C1, C4, C5, C8, C9 (carried) | **Moot.** The value on items is retired (§5). They stay in the minute for any later design. |
| C7 (held) | **Moot.** The recognizer filter is retired, and `Constrain`'s state-keyed handling of indirob is unchanged (`TestPrecedenceToyResolutionNeedsConflict`). |
| C10 (held) | **Moot as framed**, because there is no filter to prove equivalent. Its core, an "unchanged on every corpus" claim with no exact test, recurs against §3.1 as **C17**. |
| C12 | **Moot.** `Build` is no longer changed, so `forest-digest.golden` stays unchanged trivially. That also means it pins nothing about `Constrain` (C17). |
| C13 | **Moot.** The reference contract is untouched. |
| C14 | **Released** against §1's ruling "Relative and ≤ 5" and §4's exit: the after-reading's median must be below the before-reading's minimum. The ≤ 5 clause no longer stands alone. |
| C15 | **Moot.** The 47.9% citation is gone from §2. |
| C16 | **Moot.** Recognition is unchanged, so the refusal reason cannot move. |

## 2. §2's facts against the code

| Claim | Result |
|---|---|
| Stage ratios, 43%, "more than lex + recognize + build" | They agree with the minute: 2.38/5.49 = 43.4%, and 0.66+0.94+0.73 = 2.33 < 2.38. The margin is thin: Constrain's minimum (2.33) equals that sum. |
| 0.88 and 0.89, and "about 5%" | Agree with the minute (0.877, 0.891). 0.109 × 2.38 / 5.49 = 4.7%. |
| 48% and 9% map shares | They match the minute's profile (`mapassign` 30% + eval's lookups 18%, and `walk` 9%) and the code. `firsts map[ctx][]int32` and `marked map[ctx]bool` are made with no size hint (precedence.go `Constrain`). `walk` reads `p.states[s].next`, a `map[int32]int32`. |
| 687 states over 237 symbols | **Confirmed:** `len(p.states)` = 687 and `len(tab.Symbols)` = 237. Note that `tab.EOF` = 237, one past the symbols, so a terminal-indexed table needs 238 columns (C20). |
| Contexts per node | Consistent in shape. On my smaller population (the 438-file digest sample plus perl5 `lib/*.pm`, 469 accepted files, 491 calls): 182,216 evaluated nodes; 76.0% in exactly one context; at most 8 eval contexts and 32 mark contexts per node. Histogram: {1: 138,407; 2: 42,273; 3: 388; 4: 1,034; 6: 52; 8: 62}. The aggregate's 86% and ≤ 12 are plausible. I did not reproduce them. |

The ceilings are measurements, not limits. The code allows up to 687 × 238 eval contexts per node, and as many again for each first terminal in mark. So §3.3 is load-bearing (C18).

## 3. Is "identical output" achievable and testable?

**Achievable: yes.**
- `eval(n, s, t)` is a pure function of the forest, which is not mutated until the unlink loop after `mark`, and of the context. Any correct memo gives the same values.
- `mark` is idempotent per (n, s, t, first), and `kept` is a monotone OR, so traversal order cannot change it.
- The kept-in-one-context shortcut is a property of `kept []bool` being per alternative, which a map-free memo preserves automatically.

**Testable as stated: no.** See C17.

## 4. New findings

### C17: Missing. §3.1: nothing pins which alternatives Constrain keeps, beyond a sample after resolve

§3.1 lists the evidence as the two digests, the ratchet count keys, and the referee. None of them pins `Constrain`'s kept set on every file:

- **`forest-digest.golden`** hashes `forest.Build`'s forest before the constraint (forestdigest_test.go `forestDigest` → resolve_test.go `built`). This change cannot move it.
- **`resolved-forest-digest.golden`** covers 438 sample files (`samplePaths`: std5k, the port snippets, one Guacamole case in 24) against a 2,211-file aggregate. It hashes the forest *after resolve*, whose `prune` can hide an over-kept alternative.
- **The referee and the toy referees** compare a verdict and one grouping drawn from at most 16 derivations, on generated sentences.
- **`readings-collapsed`** is a coarse count.

Because this is a pure refactor, an exact oracle is cheap. Keep today's `Constrain` as a test-only reference and compare, per file, the verdict and the full `kept` bitset (or the forest after `Constrain` only) between the two. Run it over the ratchet's aggregate, the toy and perly referee sentences, and a fuzz seed set.

This also guards the one realistic way an arena rewrite goes wrong: `eval` returns cached slices that callers share (`seq` returns `right` unchanged when m is 1). An arena that callers can append into would silently corrupt memoized results.

If it ships without this, a changed kept set on a file outside the sample passes every listed check.

### C18: Partial. §3.2 and §3.3 pull against each other, and §3.3 has no test

§3.2 says structures are "allocated once per call, not grown entry by entry". §3.3 says the design must stay correct when a node has many contexts. The number of contexts is unknown before the pass, with a theoretical bound of 687 × 238 per node. So either:
- the per-node storage has a fixed inline capacity plus an overflow that grows (amortized), or
- one pooled array grows during the pass.

Both are growth. The decision should say that amortized growth of a pooled array is allowed and that per-entry hash-map growth is not. Otherwise §3.2 forbids every design that meets §3.3.

§3.3 needs a test hook rather than a numeric bound. Pattern it on the interner's `hashMask` (sets.go, where a test forces every hash to collide): a test sets the inline capacity to 1 and reruns the referees, the digests and C17's oracle. Today's maximum of 12 eval contexts (32 for mark) means the overflow path never runs on the corpora. Untested, it is the place a wrong answer can hide.

### C19: Partial. §3.4 and §4: the precedence and resolve stage keys must not time a forest the stage already mutated

`Constrain` unlinks alternatives in place (the loop at the end of `Constrain`), and `resolve` prunes in place (resolve.go `prune`). The README's stage protocol prepares inputs "once outside the span" (l.638-641). Applied as written, every pass after the first times an already-constrained forest. §2 measured that at about 0.88 of the real cost.

`Forest` has no copy method. The key definitions must say that each pass constrains a fresh copy of `Nodes`/`Alts`, made outside the span. The same goes for `resolve`, which needs a constrained copy. Both §4 stops read the precedence key, so this belongs in the decision rather than in an issue.

### C20: Low. §3.2: say which maps stay

`NewPrecedence` legitimately uses maps at construction (`sym`, `level`, the closure's `in`, `id`, `kernels`). Only `lrState.next` is read per derivation (`walk`), and it becomes a dense goto table: 687 × 237, about 160 K entries, as int16 or int32.

The decision should say that construction-time maps are allowed and that only the lookup representation changes. It should also note that terminal-indexed tables need `tab.EOF+1` = 238 columns. `seq` walks a rule prefix on every call; a precomputed (state, item) table is an optional further step. Issue-level.

### C21: Low. §4: the exit names only "its first property"

"§3 built, its first property holding" can be read as not requiring §3.2. That is the decision's own subject. Name all four properties, or say that "built" includes §3.2.

### C22: Low, issue-level. Per-call scratch must stay on the constrainer

`Precedence` is shared via `grammar.Derived(tab, buildParsePrecedence)` (readings.go). Reusable buffers hung on it would race whenever `Parse` runs concurrently, as some tests do with charts. Keep scratch on `constrainer`, as §3.2's "once per call" already implies.

### Timing rules: no contradiction found

- The stage keys are "reported, never gated, never compared with the baseline". The stops compare before and after readings taken in one session, which the README does not forbid.
- §4's "quiet: load ≤ core count" is stricter than the README's checkpoint-3 rule ("a reading records its load rather than requiring a quiet one"). That rule governs checkpoint 3 only, so I see a difference of protocol, not a contradiction.
- `alloc-per-token-*` are timing-class keys (README key table), so their expected fall does not breach §3.1's "every ratchet count key unchanged". The issue should say so explicitly.

### Is "no map" the right property?

It is perigrin's ruling, and it targets 57% of the measured cost, so I do not contest it. As worded, it forbids two things it need not:
- a pre-sized map on a rare overflow path (C18);
- `NewPrecedence`'s construction-time maps, under a literal reading (C20).

It does not threaten the kept-in-one-context shortcut, which §5 rightly keeps out of scope and §3.1 preserves.

## 5. What I did not examine

- I did not reproduce the aggregate measurement or the profile; my context counts are a 469-file sample.
- `resolve` and `readings` internals, beyond the fact that `resolve` mutates.
- Whether forests can hold same-context cycles. `eval` has no in-progress guard; a map-free rewrite must keep whatever today's behaviour is.
- The ratchet code that will host the new stage keys.
- The v2 design and the recovery plan cited in the header.

## 6. Recommendation: **modify**

The retarget fits the code. The measured shape holds on my sample, 687 × 237 is confirmed, and identical output is achievable because `eval` and `mark` are pure and order-insensitive.

The decision text needs four changes:
- **C17:** an exact kept-set oracle against today's `Constrain` over the aggregate, because the listed checks cannot show identity.
- **C18:** reconcile §3.2 with §3.3, and add a forced-overflow test.
- **C19:** stage keys time a fresh forest each pass.
- **C21:** the exit names all properties.

Issue-level: C20 and C22.
