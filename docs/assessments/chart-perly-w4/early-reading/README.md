# Early V2 speed reading

This reading was taken early, on perigrin's ruling of 2026-11-27, "Take a quiet reading now". It runs design §7's protocol, interleaved B A B A B A in one session, by `drive.sh`. This directory keeps every run's raw output and meta. It is not the d2d1 reading: that one still runs once the chain is complete.

| | |
|---|---|
| **Before** | `e3432c0f`, pu with V1 after the baseline retake |
| **After** | `5b890890`, pu with every V2 code issue merged except d247, d278 and d2a2 |

`T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit make test` passed on the after commit before the runs, with 0 FAIL.

Each run waited for `load1` to drop below 1.0. The `load1` values at start were 0.84 to 0.95, on 4 cores, and every run exited 0. Each side's ratchet binary was built once, before the runs.

## speed-ratio-aggregate-longlived

| Run | State | Value |
|---|---|---|
| 1 | before | 3.2143 |
| 2 | after | 3.3983 |
| 3 | before | 3.2489 |
| 4 | after | 3.4617 |
| 5 | before | 3.3005 |
| 6 | after | 3.3991 |

The before-minimum is 3.2143 and the after-median is 3.3991. The exit needs the after-median below the before-minimum, so it is unmet. That is design §7's stop, "the checks pass but the exit is unmet", and it was reported to perigrin.

## The 27 answer keys

All 27 answer keys are identical across the six runs.

## The stage keys (aggregate-longlived)

| Stage | Before | After |
|---|---|---|
| forest | 0.6612, 0.6704, 0.6748 | 1.4490, 1.5116, 1.4562 |
| recognize | 0.7937, 0.7670, 0.8013 | 0.8599, 0.8269, 0.8363 |
| precedence (still times Constrain on a copy) | 0.6907, 0.7175, 0.6305 | 0.6452, 0.6291, 0.6260 |
| lex | 0.6448, 0.6453, 0.6427 | 0.6546, 0.6595, 0.7138 |
| resolve | 0.1786, 0.1797, 0.1682 | 0.1782, 0.1791, 0.1772 |
| readings | 0.1259, 0.1193, 0.1189 | 0.1241, 0.1235, 0.1247 |

## Allocation and GC keys

These are reported, not gated.

| Key | Before | After |
|---|---|---|
| alloc-per-token-chart | 8.738 | 2.984 |
| alloc-per-token-forest | 0.063 | 0.113 |
| gc-cpu-share-chart | 0.1372, 0.1232, 0.1354 | 0.1232, 0.1332, 0.1276 |

exit: unmet
