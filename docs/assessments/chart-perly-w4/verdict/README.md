# V2 verdict reading

This is design §7's formal verdict reading, issue 01a21447-d2d1. `drive.sh` ran it once, in one quiet session. It ran the speed protocol interleaved B A B A B A. This directory keeps every run's raw stdout, stderr and meta, which `drive.sh` wrote. Every figure below comes from those files.

| | |
|---|---|
| **Before** | `e3432c0f48b96cb4abede6210e1751bf215ffff3`, pu with V1 after the baseline retake (`forest.go` calls `Constrain`) |
| **After** | `1477098790ad6668b402eb0c1d0ee67586e18fd6`, pu with every V2 code issue merged (`forest.go` no longer calls `Constrain`) |

Before each run, the driver waited for `load1` to drop below 1.0. Each side's ratchet binary was built once, before the runs. The runs used 4 cores on kernel 6.18.44-fc-v880.

| Run | State | `load1` | Waited (s) | Start | End | Exit |
|---|---|---|---|---|---|---|
| 1 | before | 0.89 | 15 | 2026-11-27T20:08:23Z | 2026-11-27T20:19:03Z | 0 |
| 2 | after | 0.97 | 15 | 2026-11-27T20:19:18Z | 2026-11-27T20:29:39Z | 0 |
| 3 | before | 0.83 | 15 | 2026-11-27T20:29:54Z | 2026-11-27T20:40:46Z | 0 |
| 4 | after | 0.95 | 30 | 2026-11-27T20:41:16Z | 2026-11-27T20:52:16Z | 0 |
| 5 | before | 0.86 | 15 | 2026-11-27T20:52:31Z | 2026-11-27T21:03:22Z | 0 |
| 6 | after | 0.88 | 30 | 2026-11-27T21:03:52Z | 2026-11-27T21:14:15Z | 0 |

## The exit

The exit is met when `speed-ratio-aggregate-longlived`'s after-median is below its before-minimum.

| Run | State | `speed-ratio-aggregate-longlived` |
|---|---|---|
| 1 | before | 3.3716 |
| 2 | after | 3.0753 |
| 3 | before | 3.3255 |
| 4 | after | 3.1492 |
| 5 | before | 3.3345 |
| 6 | after | 3.0205 |

The before-minimum is 3.3255 and the before-median is 3.3345. The after-median is 3.0753. Because 3.0753 is below 3.3255, the exit is met. The 27 answer keys are identical across all six runs (below), and every run exited 0.

exit: met

## The 27 answer keys

The answer keys are every key except the `speed-*`, `lex-overhead*`, `alloc-*` and `gc-*` keys, `items-per-token` and `set-reuse`. Each key has the same value in all six runs, so one value serves both sides.

| Key | Before and after |
|---|---|
| `tables-drift` | 0 |
| `leo-disagree` | 0 |
| `ledger-gaps` | 0 |
| `rd-accepted-chart-refused` | 0 |
| `chart-accepted-both-refused` | 0 |
| `lex-unmapped` | 9 |
| `tdt-aligned` | 0.9943 |
| `tdt-mismatch` | 0 |
| `tdt-unseen` | 2967 |
| `terminals-per-token` | 1.243 |
| `tdt-unseen-unlisted` | 0 |
| `tdt-unaligned-unlisted` | 0 |
| `ledger-excluded` | 18 |
| `superset-population` | 2214 |
| `superset-recognized` | 1.000 |
| `rd-accepted-perl-refused` | 230 |
| `chart-accepted-perl-refused` | 225 |
| `oracle-wrong` | 0 |
| `oracle-exact-chart` | 8524 |
| `oracle-exact-rd` | 8345 |
| `precedence-disagree` | 0 |
| `precedence-covered` | 1304 |
| `precedence-uncovered-unlisted` | 0 |
| `readings-collapsed` | 835 |
| `collapse-over-limit` | 0 |
| `superset-population-perlgit` | 562 |
| `chart-beats-rd` | 3 |

## Timing, stage, allocation and collector keys

These are reported, not gated (P8 c); only `speed-ratio-aggregate-longlived` decides the exit. The before values are from runs 1, 3 and 5. The after values are from runs 2, 4 and 6.

| Key | Before runs | Before median | Before min | After runs | After median |
|---|---|---|---|---|---|
| `lex-overhead` | 1.159, 1.180, 1.202 | 1.180 | 1.159 | 1.179, 1.175, 1.148 | 1.175 |
| `lex-overhead-std5k` | 1.159, 1.168, 1.202 | 1.168 | 1.159 | 1.179, 1.175, 1.133 | 1.175 |
| `lex-overhead-aggregate` | 1.258, 1.271, 1.229 | 1.258 | 1.229 | 1.248, 1.199, 1.256 | 1.248 |
| `speed-ratio-std5k-longlived` | 3.2512, 3.0444, 3.0645 | 3.0645 | 3.0444 | 2.9312, 3.0522, 3.0172 | 3.0172 |
| `speed-ratio-aggregate-longlived` | 3.3716, 3.3255, 3.3345 | 3.3345 | 3.3255 | 3.0753, 3.1492, 3.0205 | 3.0753 |
| `speed-ratio-aggregate-perfile` | 3.3000, 3.5150, 3.4550 | 3.4550 | 3.3000 | 3.1514, 3.3022, 3.2749 | 3.2749 |
| `speed-stage-lex-aggregate-longlived` | 0.6529, 0.6350, 0.6260 | 0.6350 | 0.6260 | 0.6600, 0.6508, 0.6520 | 0.6520 |
| `speed-stage-recognize-aggregate-longlived` | 0.7938, 0.7724, 0.7946 | 0.7938 | 0.7724 | 0.8369, 0.8408, 0.8424 | 0.8408 |
| `speed-stage-forest-aggregate-longlived` | 0.6739, 0.6789, 0.6710 | 0.6739 | 0.6710 | 1.1311, 1.1262, 1.1039 | 1.1262 |
| `speed-stage-precedence-aggregate-longlived` | 0.6714, 0.6837, 0.6586 | 0.6714 | 0.6586 | absent | — |
| `speed-stage-resolve-aggregate-longlived` | 0.1821, 0.1849, 0.1794 | 0.1821 | 0.1794 | 0.1780, 0.1856, 0.1703 | 0.1780 |
| `speed-stage-readings-aggregate-longlived` | 0.1238, 0.1258, 0.1172 | 0.1238 | 0.1172 | 0.1290, 0.1237, 0.1272 | 0.1272 |
| `alloc-per-token-chart` | 8.738, 8.738, 8.738 | 8.738 | 8.738 | 2.949, 2.949, 2.949 | 2.949 |
| `alloc-per-token-rd` | 2.077, 2.077, 2.077 | 2.077 | 2.077 | 2.077, 2.077, 2.077 | 2.077 |
| `alloc-per-token-recognize` | 0.881, 0.881, 0.881 | 0.881 | 0.881 | 0.946, 0.946, 0.946 | 0.946 |
| `alloc-per-token-forest` | 0.063, 0.063, 0.063 | 0.063 | 0.063 | 0.078, 0.078, 0.078 | 0.078 |
| `gc-cpu-share-chart` | 0.1329, 0.1406, 0.1333 | 0.1333 | 0.1329 | 0.1283, 0.1182, 0.1287 | 0.1283 |
| `gc-cpu-share-rd` | 0.1083, 0.1028, 0.0984 | 0.1028 | 0.0984 | 0.1090, 0.1130, 0.1145 | 0.1130 |
| `items-per-token` | 4.862, 4.862, 4.862 | 4.862 | 4.862 | 4.408, 4.408, 4.408 | 4.408 |
| `set-reuse` | 0.704, 0.704, 0.704 | 0.704 | 0.704 | 0.694, 0.694, 0.694 | 0.694 |

`speed-stage-precedence-aggregate-longlived` does not appear in the after runs. Each before run prints 47 `chart-ratchet:` lines and each after run prints 46, and that key is the only difference. At the after commit, `forest.ParseWith` no longer runs `Constrain`, and the key left the ratchet with it. `speed-stage-forest-aggregate-longlived` rose, from a before-median of 0.6739 to an after-median of 1.1262, while the whole-pipeline ratio fell.
