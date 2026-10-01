# RFC 0001: Type libraries

- **Status:** Accepted in part. Sections marked *Implemented* have
  landed; *Decided* sections are agreed and not yet built; *Open*
  sections wait on the types paper.
- **Date:** 2026-10-01
- **Issue:** 01a0f450-9ded-7c8a-8c8b-c04734752b2d
- **Decided by:** perigrin

## Summary

A module can define names and syntax that no reading of its source
recovers. A **type library** states them in a declaration file, the way
TypeScript's `@types/Foo` describes JavaScript the compiler never reads.
The parser reads the declaration in place of the module's source. One
file, `CORE.pmt`, does the same for perl's builtins, as TypeScript's
`lib.d.ts` does for the language itself.

## Motivation

The parser resolves `use` by reading the module's own source: its
`@EXPORT`, its `sub NAME (PROTO)` lines and its XS. That covers most of
CPAN. It fails in three ways, each found in PerlOnJava's T1 corpus:

1. **The module's `import` builds its exports.** Moose and Moose::Role
   export through Moose::Exporter; Mojolicious::Lite monkey-patches its
   DSL into the caller. Every parenless call to `has`, `requires`, `get`
   or `plugin` refused.
2. **The module adds syntax.** Future::AsyncAwait's `async sub` and
   `await` are a grammar extension, not names.
3. **The interface lives in C.** An installed perl ships no `.xs`, so a
   core XS sub's prototype is invisible (Storable's `dclone` is `$`).

Case 3 is solved by reading `.xs` from the perl.git checkout (555b4ff8).
This RFC covers cases 1 and 2 and the builtin table. Source filters
(Filter::Simple, Devel::Declare) rewrite text arbitrarily; no
declaration can describe them, and they stay refusals (perigrin).

## Design

### Declaration files (*Implemented*, a70a294e, a3095b73)

A declaration file is named for its module with the `.pmt` extension:
`Moose::Role` lives at `Moose/Role.pmt`. The files are embedded in the
binary under `internal/parse/declarations/`. When a module has a
declaration, the resolver reads it **instead of** the module's source,
whether or not the module is installed.

A declaration is written in the module syntax the resolver already
reads:

```perl
package Mojolicious::Lite;

our @EXPORT = qw(any get post put ... group helper hook plugin under);

sub get;
sub group (&);
```

`our @EXPORT` lists what a bare `use` imports. Each `sub` line declares
one sub and its prototype. Every value is measured on perl 5.42.0 with
the module installed, and the file's header names the module version.

Moose, Moose::Role and Mojolicious::Lite ship today.

### `M->import` in BEGIN is `use` (*Implemented*, a70a294e)

`BEGIN { require M; M->import(LIST) }` means `use M LIST`, and test
files load optional modules that way inside `eval { }`. An `import`
called on a literal class name in BEGIN now applies the module's
imports exactly as `use` does. A computed invocant (`$c->import`) still
marks the sub table incomplete.

### Declared syntax (*Implemented*, 15fb2815)

Keyword plugins built on XS::Parse::Keyword and XS::Parse::Sublike
state their grammar as data. A declaration file restates that data in
three attributes:

```perl
sub async  :sublike(PREFIX);       # a prefix on `sub`: async sub f {...}
sub await  :keyword(TERMEXPR);     # an expression: await $f
sub CANCEL :statement(ANONSUB);    # a statement ending at its block
```

The pieces are XS::Parse::Keyword's, without the `XPK_` prefix. The
parser reads `TERMEXPR` (an expression down to assignment, stopping at a
comma), `BLOCK` and `ANONSUB`. A declaration that names any other piece
declares nothing: guessing at a grammar builds a wrong tree.

Importing the module brings its syntax into scope from that point to
the end of the file, as features are scoped in this parser; `use M ()`
calls no import and brings none. Perl scopes these lexically.

### CORE.pmt: the builtin table (*Implemented*, 36df6a2c)

`CORE.pmt` declares the 188 keywords perl 5.42.0 reports a prototype
for, one line each, and parsing it builds the default builtin table:

```perl
sub bless :prototype($;$);
sub push  :prototype(\@@);
```

A test asks perl for every keyword's prototype and holds the file to
the answer. `:prototype(...)` is read as a prototype wherever a sub is
declared, which also fixed two perl.git files whose own subs use it.

### Declaration order: Perl's (*Decided*)

A declaration follows Perl's order: name, then attributes, then
signature. Measured on 5.42, `sub f :lvalue ($x) {}` compiles,
`sub f ($x) :lvalue {}` dies ("Subroutine attributes must come before
the signature") and `sub :lvalue f {}` is a syntax error. An operator
is therefore `sub + :infix (Num $x, Num $y) Num;`.

### Typed Perl, in `.pmt` only (*Decided*)

The syntax of `.pmt` files is **typed Perl** (perigrin): Perl whose
declarations carry types. The full form of a declaration is

```perl
sub NAME :attributes (Type $var, ...) Type;
```

with `;` where a definition would have its body. Typed Perl is accepted
**only in `.pmt` files** (perigrin); ordinary `.pm` and `.pl` source is
parsed as plain Perl. Accepting it elsewhere would also misread valid
programs, as measured: without the signatures feature,
`sub f (Ref $x) {...}` is valid Perl whose parentheses are a prototype.

When a signature is present, the prototype moves into `:prototype(...)`,
as Perl itself requires:

```perl
sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__) Object;
```

Type names come from the lattice in `internal/types` (`Int`, `Num`,
`Str`, `Undef`, `Ref`, `Object`, `List`, ...). The lattice is a bitmask,
so a union such as `Str|Undef` is proposed as the spelling for "a
string or undef"; the paper may choose another.

Once `CORE.pmt` carries types, `internal/types/signatures.go` (27 typed
builtins) and `internal/parse/keyword.go` (parse shapes) fold into it,
each after a test shows the file agrees with the table it replaces. The
comments recording their measurements move with them.

### Builtins with no prototype (*Decided*)

About 20 builtins have no prototype because their parse is their own:
`print`'s filehandle slot, `grep`/`map`/`sort`'s leading block,
`split`'s pattern, `defined &f`. Their special forms stay in the
parser. `CORE.pmt` declares their types, and `:unary` marks the eight
that are named unaries; a builtin with no prototype is otherwise a list
operator, as an ordinary sub is.

```perl
sub print (FileHandle $fh = select(), Str @args = ($_)) Bool;
```

The default handle is the selected one, not STDOUT, and `print` with no
arguments prints `$_` (measured).

### Multi declarations (*Decided*)

Some builtins return different types depending on how they are called.
A declaration may be `multi sub`, giving several signatures for one
name. A call site selects one by arity, by argument types that are not
subtypes of each other, or by context:

```perl
multi sub each (Hash %h)  List;                    # (Str, value)
multi sub each (Array @a) List;                    # (Int, value)
multi sub select (FileHandle $fh) Str;             # the previous handle
multi sub select ($r, $w, $e, Num $timeout) Int;   # a count
```

When the call site cannot decide, the consumer joins the candidates.

Bounded polymorphism needs no `multi`. `abs (Num $x) Num` already
accepts an `Int`, and the result is `meet(join(arguments), declared)`,
so `abs` of an `Int` is an `Int`.

### `:context(...)` (*Decided*)

A function whose type depends on its calling context (one that
effectively branches on `wantarray`) says so with `:context`, spelled
in prototype sigils:

| attribute | answers for |
|---|---|
| none | every context |
| `:context($)` | scalar |
| `:context(@)` | list |
| `:context()` | void |
| `:context($@)` | scalar or list |

As with `:prototype`, absent and empty differ. A call in a context no
declaration answers for has no type:

```perl
multi sub localtime :prototype(;$) :context($) (Int $time = time) Str;
multi sub localtime :prototype(;$) :context(@) (Int $time = time) List[Int];
sub sort :context(@) (...) List;     # scalar sort is undefined
```

Context selects a declaration; narrowing cannot stand in for it.
`my $t = localtime(0)` is a date string, while the first element of the
list result is `0`, an `Int`.

Boolean context counts as scalar. Measured on 5.42, `wantarray` reports
scalar inside `if (f())`, `!f()` and `f() and ...`, so no Perl-level
sub can tell them apart.

### Corpus runs share module reads (*Implemented*, 5374dda4)

Resolving imports by reading module source made the T1 ratchet re-read
Test::More's imports for every file, and it outgrew `go test`'s
ten-minute timeout. A `parse.Session` reads each module once per set of
search directories. The T1 ratchet went from 241 seconds to 4.

## Open questions

These wait on the types paper (`~/dev/types-paper`), so that pvm and
perl5-son extend one lattice:

1. **Parametric containers.** The lattice is flat. `List[T]` (values in
   flight, covariant) and `Array[T]` (a container, invariant) differ;
   `Hash[T]` flattens to `List[Str|T]`; `ArrayRef[T]` should follow from
   `Array[T]`. Until the paper decides, a type on a slurpy in a `.pmt`
   means the element type of the caller's flattened list: `Str @args`
   reads as `List[Str]`.
2. **Lvalue lists.** `chomp` and `chop` modify their arguments in
   place, which neither `List[Str]` nor `Array[Str]` says.
3. **Tuples.** `each` returns `(Str, T)` for a hash, which wants a type
   like `List[Str, T]`.
4. **Multi and context-indexed function types** have no counterpart in
   the paper's function types yet.

Separate from the paper:

5. A search path for user-written `.pmt` files beside the embedded set.
6. The remaining XS::Parse pieces, and sublikes other than `PREFIX`.
7. Lexical rather than file-wide scope for declared syntax.

## Rejected alternatives

- **Generating declarations by running perl.** Reading a module's
  symbol table means perl has already parsed it; a Perl parser that
  must run Perl is circular. It also fails for uninstalled dependencies
  and pins data to one machine's module versions. `CORE.pmt` is the
  exception: the builtins have no source to read, and its generation
  is checked against perl by a test.
- **Attribute before the name** (`sub :infix + (...)`). Perl rejects it.
- **Typed Perl in ordinary source.** It is accepted only in `.pmt`
  files; elsewhere it would misread valid Perl (see above).
- **Narrowing a union return instead of context declarations.** The
  scalar-context value is not one of the list-context values.
- **A separate grammar for the special-form builtins.** Their parse code
  exists and is measured; a declaration needs only their types.
