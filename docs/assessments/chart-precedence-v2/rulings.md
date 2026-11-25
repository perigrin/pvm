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
