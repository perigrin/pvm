# Early V2 speed re-reading

This is the re-reading taken after issue `01a21db4-4792` merged, on perigrin's ruling of 2026-11-27, "A, then re-read". That issue rebuilt Build's walk by context without a per-node hash map.

It used the same driver and protocol as `../early-reading/`, interleaved B A B A B A in one session.

| | |
|---|---|
| **Before** | `e3432c0f` |
| **After** | `6a555065`: pu with 4792, d218 and d1eb merged; d247, d278 and d2a2 not yet merged |

`T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit make test` passed on the after commit before the runs, with 0 FAIL. load1 at the start of each run was between 0.80 and 0.98, on 4 cores, and every run exited 0.

This reading is not the d2d1 reading, which is taken once the chain is complete.

## speed-ratio-aggregate-longlived

| Run | State | Value |
|---|---|---|
| 1 | before | 3.2481 |
| 2 | after | 3.0461 |
| 3 | before | 3.2591 |
| 4 | after | 3.1869 |
| 5 | before | 3.3143 |
| 6 | after | 3.0685 |

- **Before-minimum:** 3.2481
- **After-median:** 3.0685

The after-median is below the before-minimum.

All 27 answer keys are identical across the six runs.

## The forest stage

| Key | Before | After |
|---|---|---|
| `speed-stage-forest-aggregate-longlived` | 0.6710, 0.6689, 0.6888 | 1.1232, 1.1235, 1.0961 |
| `alloc-per-token-forest` | 0.063 | 0.078 |

On the first reading the forest stage was 1.45–1.51 after.

exit: met
