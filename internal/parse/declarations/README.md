# Declaration files

Each `.tpm` file here declares one Perl module's subs and their types. psc
reads it in place of the module's own source. The path follows the package
name: `Moose::Role` is `Moose/Role.tpm`. `CORE.tpm` is the builtin table.

The files are embedded in the binary (`../declarations.go`), so adding one
here ships it.

## The syntax

A `.tpm` file is written in **typed Perl**: Perl whose declarations carry
types. Typed Perl is accepted only in `.tpm` files.

```perl
package Carp;

sub croak :export (List[Str|Boolean|DualVar|Ref] @message) None;
sub carp  :export (List[Str|Boolean|DualVar|Ref] @message) Boolean;
```

- `:export` or `:export_ok` marks what a `use` imports.
- `:prototype(...)` gives the prototype perl reports.
- The signature gives each parameter's type, and the type after it is the
  return type. `None` means the sub never returns.
- `;` stands where a definition would have its body.

[RFC 0001](../../../docs/rfcs/0001-type-libraries.md) is the reference for
the syntax. Its Design sections cover:

- **Declared syntax** and **Typed Perl, in `.tpm` only**: the form of a
  declaration.
- **Type names**: the types a declaration can name, from `internal/types`.
- **What a parameter type means**: the type perl coerces the argument to.
- **Multi declarations** and **Context selects by return type**: one sub
  with several candidates.
- **Operator declarations** and **Precedence**: operators a module adds.
- **CORE.tpm: the builtin table**.
- **Core-module libraries**: how the core-module files here were ranked,
  written and held.

## Held to perl

Each core-module file states what perl 5.42 was measured to do, in comments
beside the declarations. Tests hold every file to that, running perl (psc
itself never does):

- **Prototypes:** each declared prototype equals `prototype()` for that sub
  (`Test*Prototypes` in `../heldtoperl_test.go`).
- **Exports:** the `:export` and `:export_ok` marks equal the module's
  `@EXPORT`, `@EXPORT_OK` and `%EXPORT_TAGS` (`Test*Exports` in the same
  file).
- **Return types:** the call cases in
  [`internal/infer/testdata/oracle/`](../../infer/testdata/oracle/README.md)
  are run under perl, and each observed type must be declared
  (`Test*ReturnTypes`).
- **Parameter types:** each case's arguments must fit the declared
  parameters (`TestOracleCasesFitDeclaredParameters`).
- **Every file reads cleanly:** `TestTpmDeclarationErrorsSurface`.

To change a declaration, measure the behaviour with `perl -w` first, then
record the measurement in a comment beside the line.
