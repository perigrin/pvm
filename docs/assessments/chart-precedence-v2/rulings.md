# chart-precedence-v2: rulings

These are perigrin's rulings on the V2 design, recorded on the day each was made. Only perigrin's words are rulings. Each entry gives:

- the question as it was put to perigrin;
- perigrin's answer, verbatim;
- what follows from the answer.

## 2026-11-25, while the design was being written

- **Direction.**
  - Asked how V1 lines up and what V2 is for, perigrin said: "so lets make v2 happen".
  - perigrin then chose approach 1 of three: "1".
  - Approach 1 is full reduce-side and shift-side judgement in the recognizer. The production precedence pass is retired. V1's `Constrain` and `ConstrainReference` are kept as test referees.
- **The §6 correctness fixes are out of V2.** The orchestrator proposed this as a working assumption, and perigrin did not object. It is recorded as an assumption, not a ruling.
- **Library-operator narrowing:** "Narrow before recognition (Recommended)".
  - `libraryOperators`' dead (token, terminal) pairs are applied to the lattice before recognition.
  - A scope refusal keeps its reason `scope`, but its byte may move.
  - Per-file verdicts must not change.
- **Precedence refusals:** "Keep 'precedence' (Recommended)". The recognizer records when a set died only because of a precedence refusal. It reports reason `precedence` at the byte where the set died, not at V1's byte 0.
- **Rule 12:** "Lift rule 12 for refusals". Recognition records, beside the sets, which parent each completion was allowed for, and Build reads that record. The rule is lifted for these refusal records only.
- **Guards:** "Regenerate everything".
  - There is no precedence off-switch.
  - Every guard runs with precedence on: the reference recognizer, the fuzzer, the set and forest digest goldens, and the brute-force outside reference. Each is updated to match.
  - `items-per-token` and `set-reuse` move by design.

## 2026-11-25, after assess round 1

The questions came from round-1 findings A1, A3, A9, A10, C1, P1, C9 and P6 (see the minute).

- **Superseding the governing design:** "Supersede, bounded (Recommended)".
  - V2 explicitly supersedes these parts of `docs/plans/2026-10-30-chart-parser-v2-design.md`, for precedence:
    - §8, "precedence leaves recognition (perigrin, 2026-10-30)";
    - rule 6's "after the chart";
    - rule 12's "items stay (state, distance)".
  - Items may carry the canonical spine top and nothing else.
  - Adding any other fact to an item is a stop. It replaces the §10 kill criterion "Facts in items".
- **What "a file's verdict" means:** "Accept/refuse only (Recommended)".
  - The verdict is accepted or refused.
  - A refused file's reason and byte may differ from V1's: they become positional.
  - `refused-where.txt` is regenerated.
  - This narrows the earlier "Precedence refusals" ruling: what must match V1 is accept or refuse, not the reason.
- **The exit:** "Relative only (Recommended)".
  - V2 must beat V1 past noise.
  - A correct V2 that is no faster is reported to perigrin as a stop.
- **Guards and keys:** "Hold golden, retire key (Recommended)".
  - `resolved-forest-digest.golden` stays byte-identical, as a V2 guard. This narrows "Regenerate everything".
  - `speed-stage-precedence-aggregate-longlived` is retired.

## 2026-11-25, after assess round 2

The question came from round-2 findings A20, C16 and P14 (see the minute).

- **The refusal reason:** "Diagnose by rerun (Recommended)".
  - When a file is refused, and only then, recognition is rerun without precedence, then rerun with narrowing. This reproduces V1's order exactly: grammar, then scope, then precedence.
  - `lex-unmapped` and resolve's `scope` are unchanged.
  - Reasons match V1's. Only a `precedence` refusal's byte moves, to where the chart died.
  - The cost falls on refused files only.
  - Production keeps a no-precedence mode, used only to classify refusals. This narrows "Regenerate everything"'s "no precedence off-switch". The mode classifies refusals and is not a way to parse.

## 2026-11-25, during assess round 3

- **Process:** "The goal is alignment and consensus". perigrin said this while round 3's reports were arriving, after the orchestrator suggested that a contradiction three participants had raised (C18, P17 and A27) might need perigrin's call.
  - A finding the subject can answer goes back to its raisers to settle.
  - Only views that do not reconcile go to perigrin.

**Note on the round-2 entry.** Two clauses in the round-2 entry quote the option perigrin chose; perigrin did not say them. Round 3 measured both:

- **"Only a `precedence` refusal's byte moves."** Finding A26 shows this is false for a `scope` refusal from narrowing. Measured: `foo plus 2 plus 3;` V1 byte 4, rerun byte 11. The earlier ruling "Narrow before recognition" already says that byte may move, and the subject follows that ruling.
- **"Used only to classify refusals."** The referees also use the mode, in tests. The subject now says so, and the raisers judge whether that fits.

## 2026-11-25, during assess round 6

- **Acceptance:** "I think if 7 of 8 participants agree and there are no blockers we could move forward yeah".
  - perigrin said this while round 6's last report, from the new fresh reader assess-fresh6, was still out.
  - At that point:
    - seven participants recommended accept: assess-codebase, assess-architecture, assess-pushback, assess-fresh2, assess-fresh3, assess-fresh4 and assess-fresh5;
    - no blocking finding was held.
  - Revision 7 is accepted.
  - assess-fresh6's report is recorded in the minute when it arrives. If it raises a blocking finding, the finding goes to perigrin and is not answered silently.

## 2026-11-25, after assess round 6

- **The late findings:** "Revise to rev 8, run round 7 (Recommended)". assess-fresh6's report arrived after the acceptance ruling. It recommended modify and raised three blocking findings: F6-C1, F6-C2 and F6-C3. F6-C3 is a defect for perly.y itself.
  - Revision 8 answers all three.
  - Round 7 runs within the bound of eight, to confirm consensus.
  - The acceptance of revision 7 is held until round 7 is done.

## 2026-11-25, close of assessment

There was no new ruling. The hold set by "Revise to rev 8, run round 7" ends because the assessment reached a fixed point at round 8, on revision 9:

- all ten participants recommended accept;
- no blocking finding was held or raised.
