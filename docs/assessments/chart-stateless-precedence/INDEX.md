# Evidence for the stateless precedence decision

This is the record behind `docs/plans/2026-11-23-chart-stateless-precedence.md`
and its minute, `docs/assessments/chart-stateless-precedence.md`. Everything
here was copied from the session scratchpad. Each file's original path is given
relative to that scratchpad.

Everything was taken on `pu` 8c3317079, on one 4-core machine running kernel
6.18.44-fc-v726 and Go 1.24.4. Ratios are over recursive descent's time on the
same files.

## Storage conventions

- **Gzipped diffs (`.diff.gz`).** Prototypes are gzipped so the repository's
  whitespace hooks cannot alter them.
- **Go sources (`.go.txt`).** Go sources carry `.txt` so Go tooling ignores
  them.

The executable the measurement ran (`ratchet.test`) and the raw CPU profile
(`constrain.pprof`) are not kept. Their text summaries are.

## Prototypes (`prototypes/`)

All three apply to 8c3317079 with `gzip -dc X.diff.gz | git apply`.

| File | Was | What it is |
|---|---|---|
| `stateless.diff.gz` | `stateless/stateless.diff` | The state-free filter: `ConstrainStateless`, and the parent-is-r variant |
| `combined.diff.gz` | `combined/combined.diff` | V0, V1 and V2 behind one switch, with the comparison harness |
| `no-map.diff.gz` | `assess-w3/fresh3/prototype.diff` | Today's `Constrain` with per-node lists in place of maps |

## Measurements (`measurement/`)

| File or directory | Was | What it is |
|---|---|---|
| `constrain-measurement.md` | `measure/report.md` | Stage times, `Constrain` on three forests, the profile and the context counts |
| `constrain-runs/` | `measure/*` | That measurement's raw environment, log and result files, plus `pprof-top.txt`, `pprof-cum.txt` and `ctxstats.log` (context counts) |
| `stateless-compare.txt`, `result-run*.txt`, `env-run*.txt` | `stateless/*` | The state-free filter's comparisons and timings |
| `combined-compare.txt` | `combined/result-compare.txt` | The combined prototype's comparisons and counts |
| `combined-runs/` | `combined/*` | The combined prototype's raw timing results (`result-run1..3.txt`), environments, logs, referee and toy outputs, and `run.sh` |

Two figures differ between files:

- **Whole-pipeline ratio for today's code.** It reads 5.49 in
  `constrain-measurement.md` and 5.62 in `combined-runs/`. They are different
  sessions and different harnesses. The decision quotes 5.62 because its V1
  and V2 ratios come from that same run.
- **Packed alternatives built.** `stateless-compare.txt` reports 2,566,149;
  every other file reports 2,565,144. The difference is not explained here,
  and no conclusion rests on it.

## Harnesses (`harness/`)

| Directory | Was | What it is |
|---|---|---|
| `measure/` | `measure/harness`, `measure/run.sh` | The stage-timing harness over the ratchet aggregate |
| `fresh3/` | `assess-w3/fresh3/exp_test.go.txt` | Round 3's fresh assessor's map experiment |
| `fresh4/` | `assess-w3/fresh4/` | Round 4's fresh assessor's tests and file list. Its runs covered 11,574 corpus files, 20,083 operator sentences (seed 4: 83 handcrafted plus 20,000 random) and 15,000 Leo-chain and nullable sentences (seed 44). They are summarised in `reports/round4-fresh.md`, and their generators are in these files. `list.txt.gz` holds 11,588 lines: 3,627 of them are repository paths under fresh4's since-deleted scratch worktree. Mapped onto an 8c3317079 checkout, the list resolves to the 11,574 files (round 6 confirmed this). |
| `fresh5/` | `assess-w3/fresh5/out/` | Round 5's fresh assessor's reproduction of the V1 timing |
| `codebase-ov4/` | `assess-w3/codebase/ov4/` | Round 4's codebase assessor's `go test -overlay` probes |

## Other

| File or directory | Was | What it is |
|---|---|---|
| `reports/` | `assess-w3/round*.md` | Every participant's report, verbatim |
| `c1/c1-report.md` | `chart-recovery/c1-report.md` | chart-recovery's C1 spike report, cited in the minute |
