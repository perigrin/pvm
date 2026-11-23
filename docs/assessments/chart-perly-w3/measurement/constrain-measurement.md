# Constrain measurement, pu 8c3317079 (2026-11-23)

Taken by the measure-constrain agent under perigrin's "Measure first, then decide".
Saved here by the coordinator from the agent's message, since the agent could not write it.
Raw files are in this directory: env-*, log-*, result-*.txt, aggregate-*.txt, constrain.pprof,
pprof-top.txt, pprof-cum.txt, run.sh and harness/.

## Population

The harness is a test in cmd/ratchet's package main. It calls the ratchet's own code:

- `corpus()`, with T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit;
- `perlGit(/opt/corpora/perl5 @94e50866)`;
- `realInputs`;
- `aggregateFiles(productionProtocol, rdRunner)`.

The population is 2214 files, matching the baseline's `superset-population`. The aggregate is
2211 files (T1 827, guacamole 822, perlgit 562), and its list was identical in all three runs.
Three files are excluded: std5k.pl, edit5k.pl and T1/code_too_large.t.

## Protocol

- Machine: 4 cores, kernel 6.18.44-fc-v726, go1.24.4. The 1-minute load at each run's start
  was between 0.27 and 1.39, against a gate of 4.
- Each run had 2 warm-up rounds and 9 timed rounds, long-lived, with the stages interleaved
  inside each round.
- There were three processes. Each figure is the median of the three run medians, with the
  smallest and largest run median in brackets.

## Stage time, in units of RD's time (546 ms)

| Stage | Time |
|---|---|
| ParseWith (whole pipeline) | 5.4878 (5.4821–5.6433) |
| lex | 0.6608 |
| recognize | 0.9422 |
| build | 0.7311 |
| libops (libraryOperators) | 0.0922 |
| constrain | 2.3826 (2.3285–2.4536) |
| resolve | 0.1980 |
| readings | 0.1457 |

Constrain is 43.4% (42.5–44.2%) of ParseWith.

Whole ParseWith reads 5.49 here, against the committed 4.79. The two were taken on different
machines, and this harness's heap is larger. Compare stages with each other, not with the
ratchet's keys.

## Constrain on three forests, as a fraction of A's time

| Forest | Fraction of A | Time |
|---|---|---|
| A, as built | 1 | 1303 ms |
| B, reduce-side pruned first | 0.8912 (0.8611–0.9173) | 1170 ms |
| C, everything Constrain drops removed first | 0.8770 (0.8472–0.9205) | 1164 ms |

## Packed alternatives

| Forest | Count | Removed |
|---|---|---|
| A | 2,565,144 | |
| B | 2,421,356 | 5.61% |
| C | 2,150,267 | 16.17% |

## Exactness

- Alternatives the filter removes that Constrain keeps: 0.
- Verdict flips: 0.
- Files whose root the filter kills: 0.

## Profile of Constrain (30 passes)

Map operations keyed by ctx are about 57% of Constrain's own CPU:

- mapassign: 30%, of which 7.4 s of the 12.4 s is rehashing or growing maps that were made
  with no size hint;
- mapaccess2 for eval's firsts lookup: 18%;
- the goto map in Precedence.walk: 9%.

The rest:

- Precedence.reduces: 11%;
- mallocgc: 13%.
