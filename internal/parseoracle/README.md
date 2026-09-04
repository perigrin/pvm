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

## Running against perl's own test suite

perl's tests are **not** runnable in place. `t/test.pl:119` does `@INC = ()`
and then unshifts `../lib`, so they only compile inside a built perl tree —
`PERL5LIB` cannot help, since `@INC` is cleared after it is read. 498 of the
620 files follow this convention.

The shim: a directory holding a copy of `t/` beside a populated `lib/`.

```sh
mkdir -p shim/lib && cp -r perl5/t shim/t
cp -r "$(perl -e 'print $INC[-1]')"/* shim/lib/
cd shim/t && perl -c op/sub.t          # syntax OK
```

## Ground-truth ceiling

Measured with perl 5.42.0. This is what perl itself can compile here, and
therefore the ceiling any conformance target must be stated against — a goal of
"parse all 620" would be measuring against files perl cannot compile either.

| Directory | Compiles | Rate |
|---|---|---|
| base | 9/9 | 100% |
| comp | 25/25 | 100% |
| cmd | 5/5 | 100% |
| opbasic | 5/5 | 100% |
| lib | 9/10 | 90% |
| mro | 62/73 | 84.9% |
| op | 160/228 | 70.2% |
| uni | 21/30 | 70% |
| re | 55/80 | 68.8% |
| run | 18/28 | 64.3% |
| io | 22/44 | 50% |
| porting | 10/37 | 27% |
| class | 1/12 | 8.3% |
| **total** | **411/620** | **66.3%** |

The failures are environmental, not syntactic: unbuilt XS, `Config`,
platform-specific tests. `porting/` tests perl's source tree rather than the
language. `class/` is 5.38 syntax this perl mostly rejects — it becomes
available on a newer toolchain rather than being permanently out of reach.

`base`, `comp`, `cmd` and `opbasic` at 100% are the natural first milestones:
49 files that perl compiles cleanly, so any failure there is ours.
