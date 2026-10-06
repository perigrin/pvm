# PVM Types

Perl has types even though it never asks you to write them: `"42" + 1`
works because `"42"` is a number, and under `use warnings` `"hello" + 1`
warns because it is not. PVM finds those types in ordinary Perl, so your
code needs no annotations.

## How PVM Finds Types

PVM reads what each value is from how it is written and used, and orders
the types the way Perl's own coercions do:

```
Int <: Num <: Str <: Scalar <: List
```

Every integer is a number, every number stringifies, and every scalar is
a one-element list. Alongside these sit `Undef`, `Boolean`, the reference
types (`ScalarRef`, `ArrayRef`, `HashRef`, `CodeRef`, `GlobRef`,
`Object`), the containers `Array` and `Hash`, and `IO`, what a
filehandle's IO slot holds. A filehandle is a `FileHandle`: a bareword
handle (`Glob`), a lexical one (`GlobRef`) or an `IO`.

A check such as `defined($x)` or `ref($x)` narrows a value's type inside
the branch it guards.

## Checking a File

`psc check` reports where a value reaches an operation that cannot use
it:

```perl
my $count = 42;
my $name  = "example";
my $sum   = $count + $name;
```

```
$ psc check script.pl
script.pl:3:22: warning: right operand of "+": expected Num, got Str [coercion-mismatch]
```

`psc check` accepts a file or a directory. With `--strict`, it also
reports values whose type it could not work out, instead of accepting
them.

## Declaration Files

Some modules define subs that their source never shows. Moose builds
`has` and `extends` when it is imported, and Mojolicious::Lite installs
`get` and `post` into your package when you import it. Some add syntax:
Future::AsyncAwait's `async sub` and `await`.

For these, PVM ships **declaration files**, which state what a module
defines. A declaration file is named for its module with a `.pmt`
extension (`Moose/Role.pmt` for Moose::Role), and PVM reads it in place
of the module's source, whether or not the module is installed. PVM
ships declarations for:

- Moose and Moose::Role
- Mojolicious::Lite
- Future::AsyncAwait
- perl's own builtins, in `CORE.pmt`

Declaration files are written in **typed Perl**, a superset of Perl
whose declarations can carry types. The files PVM ships today use only
the plain-Perl part:

```perl
package Mojolicious::Lite;

our @EXPORT = qw(
    any get options patch post put query websocket
    new app del group helper hook plugin under
);

sub get;
sub group :prototype(&);
```

Typed Perl is valid only in `.pmt` files, and you cannot yet add your
own. The design, including the typed declarations still to come, is in
`docs/rfcs/0001-type-libraries.md`.
