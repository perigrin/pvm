# chart-precedence-v2: rulings

perigrin's rulings on the V2 design, recorded on the day they were made. Only perigrin's words are rulings. Each entry gives the question as it was put to him, his answer verbatim, and what follows from it.

## 2026-11-25

- **Direction.** Asked how V1 lines up and what V2 is for, perigrin said: "so lets make v2 happen". He then chose approach 1 of three: "1". That approach is full reduce-side and shift-side judgement in the recognizer, with the production precedence pass retired and V1's `Constrain` and `ConstrainReference` kept as test referees.
- **The §6 correctness fixes are out of V2.** The orchestrator proposed this as a working assumption, and perigrin did not object to it. It is recorded here as an assumption, not a ruling.
- **Library-operator narrowing:** "Narrow before recognition (Recommended)". `libraryOperators`' dead (token, terminal) pairs are applied to the lattice before recognition. A scope refusal keeps its reason `scope`, but its byte may move. Per-file verdicts must not change.
- **Precedence refusals:** "Keep 'precedence' (Recommended)". The recognizer records when a set died only because of a precedence refusal, and reports reason `precedence` at the byte where it died, not V1's byte 0.
- **Rule 12:** "Lift rule 12 for refusals". Recognition records, beside the sets, which parent each completion was allowed for, and Build reads that record. README rule 12 is lifted for these refusal records only.
- **Guards:** "Regenerate everything". There is no precedence off-switch. Every guard runs with precedence on: the reference recognizer, the fuzzer, the set and forest digest goldens and the brute-force outside reference. Each is updated to match. `items-per-token` and `set-reuse` move by design.
