<!-- ABOUTME: How to write and test a .tpm declaration for a core Perl module. -->
<!-- ABOUTME: States the procedure; docs/rfcs/0001-type-libraries.md states the design it follows. -->

---
stability: 1
covers: []
---

This is the procedure for declaring a core module: picking it, writing its
`.tpm`, holding it to perl, and verifying the result. The design it follows —
why declarations look this way — is `docs/rfcs/0001-type-libraries.md`. This
page does not restate that; it only says what to do.

## Choosing the module

`TestCoreModuleRanking` (`internal/parse/ranking_test.go`) ranks core modules
by how often they are imported from outside their own namespace, over an
installed perl's `privlib`. A core module is any `.pm` under `privlib`; an
import is a `use` or `require` line in any `.pm` or `.pl` under `privlib`,
except one in a file under the module's own top-level namespace. Ties break
by module name, ascending.

The top 50 are pinned in `internal/parse/testdata/core_module_ranking.txt`,
one `Name Count` line per module. `TestCoreModuleRanking` only compares
against this file; it has no `-update` flag to regenerate it. To
regenerate against a newer perl, run the test, read the ranking it computed
from the failure diff, and write the new top 50 to the file by hand. Module
51 is whichever name the new ranking adds past the current last line.

## Writing the `.tpm`

A module's declaration lives at `internal/parse/declarations/<Module/Path>.tpm`,
mirroring the module's package name (`::` becomes `/`): `IO::Select` is
`internal/parse/declarations/IO/Select.tpm`. Start every `.tpm` with two
`# ABOUTME:` lines, as every file in the tree does.

**No export lists.** The parser refuses `@EXPORT`, `@EXPORT_OK`, and
`%EXPORT_TAGS` in a `.tpm` — not because they're redundant, but because the
list and the subs it names can drift. Instead, every name a `use` can install
is marked on the thing itself:

- `:export` — installed by a bare `use Module;`
- `:export_ok` — installed only when asked for by name
- `:export_ok(tag1,tag2)` — also installed by `use Module ':tag1';`

A name a module's own `import` installs (not Exporter's) is marked the same
way. Locale::Maketext::Simple's `import` installs `loc`, so
`internal/parse/declarations/Locale/Maketext/Simple.tpm` marks it
`sub loc :export`, even though the install comes from `import`'s own logic,
not an `@EXPORT` array.

**`&name` entries are sub names.** The skeleton generator's `installs` mode
reports an `&name` for a name in `@EXPORT`/`@EXPORT_OK` that is a sub; write
it as the sub's bare `name`, not `&name`.

**Tags as attribute arguments.** `internal/parse/declarations/Errno.tpm`
marks most constants `:export_ok(POSIX)` — in `%EXPORT_TAGS`' `POSIX` group —
and the rest plain `:export_ok`:

```perl
sub EACCES :export_ok(POSIX) () Int;
sub EADV   :export_ok        () Int;
```

**`:prototype` and `:listop`.** When perl reports a prototype for the sub
(`prototype(\&Module::name)` is defined), write it as `:prototype(...)`:

```perl
sub GetOptions :export :prototype(@) (List @args) Boolean;
```

When perl reports *no* prototype but the sub still needs to parse like a
builtin with declared parameters (so a typed signature's absence of
parentheses isn't mistaken for perl's explicit empty `()` prototype), mark
it `:listop` instead. `internal/parse/declarations/Cwd.tpm`:

```perl
sub cwd  :export :listop () Str;              # no params, no perl prototype
sub chdir :export_ok :listop (Str $dir) Int;  # one param, no perl prototype
```

**`None` for subs that die.** A sub whose every path ends in `die` or
`exit` — never returns to its caller — is declared `None`, the same way
`CORE.tpm` declares `die` and `exit` themselves:

```perl
sub croak :export (List[Str|Boolean|DualVar|Ref] @message) None;
```

**Typed `our` variables.** A module-level variable a `use` can install is
declared with its type before the variable, and the export mark as an
attribute after it:

```perl
our Hash[Str|Undef] %Config :export;
our Str $CRLF :export_ok(crlf,all);
```

A typed `our` is legal only in a `.tpm`; perl itself refuses `our Hash[Str]
%x` at compile time, same as the parser does outside a declaration file.

**Multis, selected by argument type or arity.** A sub whose return type (or
parameter shape) depends on calling context or argument count is written as
several `multi sub` candidates with the same name.
`internal/parse/declarations/IO/Select.tpm` has one argument-count form and
several context-dependent ones:

```perl
multi sub handles (Object $self:) List[FileHandle];   # list context
multi sub handles (Object $self:) Int;                # scalar context

multi sub can_read (Object $self: Optional[Num] $timeout) List[FileHandle];
multi sub can_read (Object $self: Optional[Num] $timeout) Int|Undef;
```

**Methods take an invocant.** A sub that is only ever called as a method —
measured: a bare `use Module;` installs nothing — takes its invocant before
a colon in the signature, and carries no `:export` mark at all:

```perl
sub new (Str $class: List[FileHandle] @handles) Object;
sub add (Object $self: List[FileHandle] @handles) Int;
```

**A declared `import` with a signature.** When a module's own `import`
decides what to install from its arguments — not from `@EXPORT`/`@EXPORT_OK`
membership — declare `import` itself, with the invocant before the colon,
and name the one parameter whose *presence* changes what gets installed
after the key for it in the signature. `internal/parse/declarations/Test/More.tpm`:

```perl
sub import (Str $class: Str $import = undef, List @plan) Void;
```

Measured on 5.42.0: `use Test::More import => ['ok'], tests => 1;` installs
only `ok`; a use-list carrying that key installs nothing else, one without
it installs the module's full default set. The parameter is named for the
key, not given a generic name, so the signature states the rule.

## Holding it to perl

A batch of modules measured together adds three tests, each named for the
batch:

- `Test<Batch>Prototypes` — every declared prototype equals
  `prototype(\&Module::name)`, via `held_to_perl.pl prototypes`.
- `Test<Batch>Exports` — every `:export`/`:export_ok` mark equals what
  `use Module;`, `use Module qw(name);`, and `use Module ':tag';` each
  actually install in a fresh caller package, via
  `held_to_perl.pl installs`.
- `Test<Batch>ReturnTypes` — every declared return type admits what the
  return-type oracle observes when perl calls the sub.

`internal/parse/testdata/held_to_perl.pl` is the driver all three run
against a real perl; psc itself never runs perl. Its `skeleton` mode prints
a starting `.tpm` for a module — `package M;`, then one
`sub NAME :export|:export_ok(tags) (PARAMS) TODO;` or
`our TODO NAME :export...;` per name a bare and tagged `use` would install,
params derived from the sub's own prototype — so writing a new module's
declaration starts from `held_to_perl.pl skeleton Module` and fills in the
`TODO`s rather than transcribing names by hand.

**AUTOLOAD subs aren't defined until primed.** A module like POSIX
forward-declares subs through `AUTOLOAD`: `exists &POSIX::stdin` is false
right after `require POSIX`, true only once some `use` has run. The export
test checks `exists &name` rather than calling the sub directly, and primes
an AUTOLOAD module with one `use` before checking.

**Oracle call cases.** Each module measured for return types gets a case
file at `internal/infer/testdata/oracle/<Module>.pl` (module name,
`::` replaced with `-`), a Perl hash literal mapping each sub name to a list
of argument lists — one call per list, each checked against the declared
type:

```perl
(
    carp      => [ ['warned'], [ 'two', ' parts' ] ],
    croak     => [ ['croaked'], [ { error => 1 } ] ],
    longmess  => [ ['long'], [] ],
);
```

A method needs an instance first. The `INSTANCE` key names a constructor the
driver calls once, and every other case in the file runs against that one
shared instance, in alphabetical order by name —
`internal/infer/testdata/oracle/IO-Select.pl`:

```perl
(
    new           => [ [] ],
    add           => [ [ \*STDIN ] ],
    remove        => [ [ \*STDIN ] ],
    INSTANCE      => {
        ctor => 'new',
        args => [ \*STDIN ],
        subs => [qw(add remove exists handles can_read can_write has_exception count)],
    },
);
```

**`oracleExempt`.** Some subs are unsafe to call in the oracle's scratch
directory — network, fork/exec, signals, `_exit`, mutating a file outside
scratch, blocking on a terminal. These go in the `oracleExempt` map in
`internal/infer/returntypes_test.go`, keyed `"Module::sub"`, with a one-line
reason:

```go
"POSIX::fork": "creates a real child process",
```

An exempt sub is held to its prototype and export membership only; the
return-type oracle never calls it. `TestReturnOracleNeverCallsAnExemptSub`
proves the mechanism works at all, and `TestNoModuleIsAllExempt` guards
against the degenerate case: a module whose *every* declared sub is exempt
would pass its return-type test without ever calling perl. A module that
genuinely has nothing safe to call — `XSLoader`, `Net::Ping` — is named in
the `allExemptAllowlist` map instead of silently passing; any other
all-exempt module fails that test until it gets a real, callable case.

## Modules other modules depend on

`Exporter`, `XSLoader`, and `DynaLoader` are inherited from or called by
many other modules' own imports. A declaration for one of these three could
change how those other modules resolve, so each is written and then
measured against the full corpus: it must leave every module that uses it
resolving exactly as before, and hold the m1 gates, or it does not ship.
These three are the only modules of the 50 that may not ship, and only
perigrin rules one out — a measurement that would exclude one is put to him
first, and the recorded reason is his.

A module excluded this way is recorded in
`internal/parse/testdata/core_module_not_shipped.txt`, one line per module:

```
Module<TAB>full-uuid<TAB>reason
```

`full-uuid` is the issue (or ruling) that excluded it, so the reason traces
back to its source.

## Verifying

Run the declaration's own package tests, and the parser and inference
suites generally:

```sh
go test ./internal/parse/...
go test ./internal/infer/...
```

These tests call out to a real perl 5.42.0 for `Test<Batch>Prototypes`,
`Test<Batch>Exports`, and `Test<Batch>ReturnTypes`, and for
`TestCoreModuleRanking`. `internal/perltest.Perl542` finds the perl 5.42.0
pvm itself has installed and reads its `privlib`/`archlib`; where that perl
isn't installed, it calls `t.Skip` rather than failing the test. A clean
`go test` run on a machine without it is not evidence a declaration holds —
check that these tests actually ran, not skipped, before trusting the
result.

## Lib baseline

`internal/psc/testdata/lib_true_positives.txt` is what `psc check` should
still flag in perl 5.42's own `lib` after all narrowing is applied — a
`privlib`-relative path, a tab, and the diagnostic line `perl -w` prints
*running the real lib file*. An entry built from a `-e` one-liner that
re-creates the construct inline, rather than evidence that the actual lib
file triggers it, does not belong here: the file's own code is what
`lib_true_positives.txt` claims is still a true positive, and a minimal
repro proves a different, smaller program, not that claim.

`--warn-possible-undef` (`psc check`'s flag, `infer.Options.WarnPossibleUndef`)
gates one specific class of diagnostic: a declared call whose `Undef` result
is one of the outcomes the declaration itself states, not one the call's own
arguments prove — a single declared candidate (`Scalar::Util::blessed`), a
multi's catch-all case, or a provably-sound arity. These are silent by
default; `psc check` reports them only when the flag is passed. This is
perigrin's ruling: don't warn on a call shape the declaration says is
genuinely ambiguous unless asked to.

## See also

- `docs/rfcs/0001-type-libraries.md` — the design.
- `internal/parse/testdata/held_to_perl.pl` — the perl-held measurement
  driver (`inc`, `prototypes`, `installs`, `skeleton` modes).
- `internal/infer/returntypes_test.go` — `oracleExempt`,
  `allExemptAllowlist`, `exemptMapFor`.
