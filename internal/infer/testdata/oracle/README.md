# Oracle call cases

Each file here holds the calls that check one `.tpm` declaration in
`internal/parse/declarations/` against real perl 5.42. Tests run perl;
psc itself never does.

There is one file per declared module, named for the module with `::` as `-`.
So `List::Util` is `List-Util.pl`, and its declaration is
`internal/parse/declarations/List/Util.tpm`.

## The format

A file is a Perl list that pairs each sub's name with the argument lists to
call it with, one call per list:

```perl
(
    carp     => [ ['warned'], [ 'two', ' parts' ], [ { c => 1 } ] ],
    longmess => [ ['long'], [], [ { x => 1 }, 'b' ] ],
);
```

A method that needs a real object rather than the class name uses the reserved
key `INSTANCE`. Its `ctor` and `args` name one constructor call, and every sub
in `subs` is called on the object that call builds:

```perl
INSTANCE => { ctor => 'new', args => [ 'src.tar', 0 ], subs => [qw( add_data list_files ... )] },
```

Code before the list runs first, in a scratch directory, with HOME and
TMPDIR also pointing there. `Cwd.pl` uses this to remove its own working
directory so that `getcwd` can be seen returning undef.

## What reads them

- **The return-type oracle** (`../return_oracle.pl`, from
  `returntypes_test.go`):
  - It makes every call in scalar and in list context.
  - `../observe_types.pl` sorts each value it returns into a type: Int, Str,
    Undef, DualVar, Glob, a reference type, and so on.
  - A type the `.tpm` does not declare for that call fails the test. This
    holds each library's return types to perl with WRONG = 0.
- **The parameter oracle** (`../parameter_oracle.pl`, from
  `parametertypes_test.go`):
  - It calls nothing. It observes each call's arguments.
  - It checks that every argument fits the declared parameter of the
    candidate psc would select for that call.

## When a sub cannot be called here

- **`oracleExempt`** (`returntypes_test.go`): a sub that exits, forks, uses
  the network, writes outside the scratch directory, or needs a terminal is
  listed here with its reason. Its prototype and export membership are still
  held to perl.
- **`platformExempt`** (`returntypes_test.go`): a name that differs on macOS
  or Windows is listed per OS, with the CI job that measured it.
  Zero-argument constants
  are measured live on the running perl instead.
- **`parameterExempt`** (`parametertypes_test.go`): a flag argument,
  written as a literal 0 or 1, that a declared Boolean parameter admits only
  through perl's truth coercion.
- **`pendingAudit`** (`returntypes_test.go`): a known misfit, keyed by the
  issue that fixes it. It is empty now.

`TestPendingAuditEntriesStillMisfit` fails a `pendingAudit` entry that no
longer misfits, so a fix cannot land without removing its entry.

## Keeping coverage complete

- `TestEveryMultiCandidateIsExercised`: every candidate of a multi must be
  reached by some call here.
- `TestNoModuleIsAllExempt`: a module whose every sub is exempt fails unless
  it is allowlisted with a reason.
- `TestOracleCasesFitDeclaredParameters`: the parameter oracle over every
  held module.

To add a case, measure the call with `perl -w` first, then add its argument
list under the sub's name.
