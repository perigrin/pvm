<!-- ABOUTME: The parse oracle — measures whether a parser agrees with perl, not just whether it errors. -->
<!-- ABOUTME: Extracts perl's own parse decisions (optree, prototypes) as JSON ground truth. -->

# Parse Oracle

Perl reports how it parsed something. That makes parser *fidelity* measurable
instead of a matter of opinion.

```
$ perl -MO=Concise,-exec -e 'sub f(\@){} my @a; f(@a)'   # srefgen PRESENT
$ perl -MO=Concise,-exec -e 'sub f{}    my @a; f(@a)'    # srefgen ABSENT
```

Identical source shape, different parse, and the optree says which. Verified —
this also holds when the prototype is installed by a string `eval` inside
`BEGIN`, which is the undecidability case.

## Why this exists

Nothing has ever checked whether our tree is the tree perl builds — only
whether it contains ERROR nodes. Those are different questions:

| Metric | Question | Status |
|---|---|---|
| Coverage | Does it parse without error? | Measured today |
| **Fidelity** | **Does it parse the way perl does?** | **Unmeasured** |

A file can score perfectly on coverage and still be parsed wrong. `f @a` read
as a list where perl took a reference produces a clean tree and a false one.

## Usage

```sh
perl testdata/parse_facts.pl FILE            # JSON on stdout
ORACLE_CHDIR=/path/to/t perl testdata/parse_facts.pl FILE
```

Emits `ok`, the linear op sequence (`-exec` order, so it diffs cleanly),
`srefgen`/`entersub` counts, and every prototype in scope.

## What the optree can and cannot answer

The optree is captured **after** the peephole optimiser, so it is not a
faithful record of the parse tree:

```
$ perl -MO=Concise,-exec -e 'my $x = 1 + 2'
3  <$> const[IV 3] s/FOLD
```

The `add` op is gone — constant folding happened before Concise saw it. Two
rules follow.

**Compare for the presence of a marker op, never for op counts or an exact
sequence.** `srefgen` works as a signal because nothing folds it away. A
whole-tree diff would report differences that are the optimiser's, not the
parser's.

**The `s/FOLD` flag is a gift.** Perl marks folded constants rather than
silently rewriting them, so folding is detectable when it matters.

Concise output is byte-identical across runs (verified by md5), so results are
safe to cache on a content hash. `perl -c` costs ~9 ms and Concise ~25 ms per
file, which is the reason to cache at corpus scale.

## Running against perl's own test suite

perl's tests are **not** runnable in place. `t/test.pl:119` does `@INC = ()`
and then unshifts `../lib`, so they only compile inside a tree where `../lib`
is populated — `PERL5LIB` cannot help, since `@INC` is cleared after it is
read. 498 of the 620 files follow this convention.

Populate the shim from **both** library roots. `Config.pm` and every XS
module's `.pm` half live in the architecture-specific directory; leaving it out
costs ~170 files and makes the corpus look far harder than it is.

```sh
mkdir -p shim/lib && cp -r perl5/t shim/t

perl -e 'for (@INC) { print "$_\n" if -d && !/site_perl|vendor_perl/ }' \
  | while read d; do cp -rn "$d"/* shim/lib/ 2>/dev/null; done

cd shim/t && perl -c op/sub.t          # syntax OK
```

## Ground-truth ceiling

Measured with perl 5.42.0 and the shim above: **584/620 (94.2%)** compile.
Nine of fourteen directories are at 100%; `porting/` (43.2%) tests perl's own
source tree rather than the language.

Classifying by stderr rather than exit status — measured against the earlier
*incomplete* shim, when 185 files were failing — shows how little of it was
ever about Perl:

| Cause | Files |
|---|---:|
| Missing module / `@INC` | 167 |
| Version or feature skew | 1 |
| **Genuine syntax error** | **1** (`op/for-many.t`) |
| Other (exit status, timeouts) | 16 |

The completed shim recovers almost all of the first row, which is how 411/620
became 584/620.

**One file in 620 fails for a reason about the language** — `op/for-many.t`,
which uses `foreach my ( \@array ) (...)`, a blead-only refaliasing form that
5.42 cannot parse. The corpus here is blead 5.45 while the interpreter is
5.42.0, so pin both and treat a mismatch as a reason to re-baseline.

The harness must classify stderr. An exit status cannot tell "your parser is
wrong" from "this machine lacks `Config.pm`", and a ratchet built on exit
status encodes the second as though it were the first. A pass rate well under
94% is evidence the setup is broken, not that Perl is hard.
