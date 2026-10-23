# Op count fixtures

These fixtures come from perl's `t/perf/opcount.t`. Each one is a
`test_opcount` call from that file: a snippet sub and the number of each op
perl compiles it to. `opcount_test.go` checks every count against the pinned
perl, 5.44.0, through `perltest.PinnedPerl`. The tests run perl; psc itself
never does.

They were mined from perl5 at 94e5086608bcf1d01cd2d836f11ca8378a3320b6, the
revision in `../corpus.pin`.

## The format

`opcount.md` holds one fixture per `##` heading:

````markdown
## L679: try/catch: catch block is optimized

```perl prelude
use feature 'try';
```

```perl
sub { my @a; try {} catch($e) { $a[0] } }
```

| op | count |
| --- | --- |
| aelem | 0 |
| aelemfast_lex | 1 |
````

- **The heading** gives the opcount.t line where the `test_opcount(` call
  starts, then the call's description.
- **The `perl prelude` fence** is optional. It holds what the snippet needs
  from its enclosing scope in opcount.t: a `use feature`, a `no warnings`,
  the `use builtin` import list, or a lexical such as `my @a;`. It is copied
  from opcount.t.
- **The `perl` fence** holds the snippet exactly as the call writes it.
- **The table** lists each op the call names, with the count opcount.t
  expects. A null op is named `ex-` plus the op it replaced, as opcount.t
  names it.

## How the counts are checked

- **`TestOpcountFixturesHeldToPerl`** compiles each snippet on its own:
  `use warnings; use strict;`, then the prelude, then the snippet. It counts
  the ops under the sub's `ROOT` the way opcount.t's `test_opcount` does. A
  snippet that fails to compile, or makes perl warn, is a failure.
- **`TestOpcountFixturesMatchTheirSourceLines`** checks the snippet against
  the call at the line the heading names. It also checks that the call's
  op=>count pairs are exactly the table.

## What is not mined

A zero count passes for any snippet that lacks the op, so a fixture must
hold at least one nonzero count. These calls are left out:

- **Calls with only zero counts:** lines 262, 559, 583, 1286, 1294, 1300,
  1306 and 1312.
- **Calls built in a loop by `eval`:** lines 183, 341, 383, 534 and 611.
  These have no literal snippet to quote.
