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

### Type names (*Decided*)

A `.pmt` names types from the lattice in `internal/types`, and an
unknown name is an error. perigrin, 2026-10-02:

- **Unions are spelled `A|B`**, as `Str|Undef`. The lattice is a bitmask,
  so a union is an OR, and it is what Moose and Type::Tiny users already
  write.
- **`IO` joins the lattice** (the paper has it), and **`FileHandle` is a
  named union, `Glob|GlobRef|IO`**, defined in the lattice as `Ref` and
  `Scalar` are, so a `.pmt` needs no alias syntax. Measured on 5.42: a
  bareword handle is a `Glob`, `open my $fh` gives a `GLOB` reference,
  and `*STDOUT{IO}` is an IO object. A string naming a handle is not a
  `FileHandle`: under `use strict`, `print {"STDOUT"} ...` dies ("Can't
  use string ("STDOUT") as a symbol ref"). A blessed handle such as an
  `IO::File` object is reftype `GLOB` but files under `Object`, so it is
  not a `FileHandle` until the paper says how blessing and reftype
  combine.
- **Names follow the paper.** The lattice's `Bool` is the paper's
  `Boolean`, and a `.pmt` writes `Boolean`.
- **The paper's other types** (`VString`, `Format`, `Void`, `LValueRef`)
  join the lattice when a declaration first needs one.

### One language for every `.pmt` (*Decided*)

perigrin, 2026-10-02: everything `CORE.pmt` can say, any library's
`.pmt` can say too -- typed signatures, `multi`, `:context`, the
invocant colon, declared syntax. `CORE.pmt` is the declaration file for
the interpreter, not a dialect of its own. A construct no Perl-level sub
can have (the invocant colon) is still declarable for a library,
because a keyword plugin can build what a sub cannot.

Typed Perl is meant to outlive pvm's parser: Chalk is to read the same
`.pmt` files in time, so the syntax should stay something a second
implementation can parse from this RFC alone.

### Builtins with no prototype (*Decided*)

About 20 builtins have no prototype perl can report. `CORE.pmt`
declares their types, and `:unary` marks the eight that are named
unaries; a builtin with no prototype is otherwise a list operator, as an
ordinary sub is. Most of their forms turn out to be declarable; see
"Builtins that keep their own parse" for which, and how.

```perl
sub print (FileHandle $fh = select(): List[Str] @args = ($_)) Boolean;
```

The default handle is the selected one, not STDOUT, and `print` with no
arguments prints `$_` (measured).

### Builtins that keep their own parse (*Decided*)

perigrin, 2026-10-02. Measured on 5.42 throughout.

**No prototype is the default prototype.** A sub with no prototype and
one with `(@)` parse identically in every form tried (`f $a, $b`,
`f @a, $b`, `f { 1 }`, `f`, `f + 1`, `f->[0]`, `f @a ? 1 : 2`); only
`prototype()` tells them apart, `undef` against `'@'`. So a declaration
that derives `@` is a sub with no prototype, and a one-argument CPAN sub
with none is declared from the caller's side, `(List @args)`, not
`(Str $x)`: the latter claims a `$` prototype the sub lacks. The
`CORE.pmt` check treats perl's `undef` as a derived `@`.

**Most "special forms" are multis.** perlsub reimplements `grep` as
`sub mygrep (&@)`, and that covers the block form exactly. The
expression form is not a block: `grep /a/ || 1, @l` deparses with the
same shape as a `($@)` sub's `g((/a/ || 1), @l)`. What differs is
evaluation, not parse (`grep $n++ >= 0, 1 .. 3` evaluates the
expression 3 times). So:

```perl
multi sub grep (Code &block, List @list) List;    # &@
multi sub grep (Scalar $expr, List @list) List;   # $@
multi sub map  (Code &block, List @list) List;
multi sub map  (Scalar $expr, List @list) List;
multi sub sort (Code &block, List @list) List;
multi sub sort (List @list) List;
```

**`map`'s and `grep`'s brace is variant selection.** Both guess whether
`{` opens a block or a hash constructor from its first tokens; a `&@`
sub never guesses. Wherever perl's guess disagrees with what follows
the closing brace, perl rejects the program: `map { "\L$_" => 1 } @a`
(guessed hash, no comma) and `map { $_ => 1 }, @a` (guessed block,
then a comma) are both syntax errors. So a comma after `}` selects the
expression variant and its absence the block variant, which agrees with
perl on every program perl accepts. The first argument's type selects
the variant, as in Chalk, decided by that comma rather than by the
braces' contents: `map { $_ => 1 } @a` holds pairs and is still a
valid block. The parser keeps perl's first-tokens guess only to refuse
what perl refuses.

**A leading slot with no comma is written with the invocant colon.**
`print STDERR LIST`, `sort byname @x` and `exec {'/bin/sh'} @args` take
their first argument with no comma after it, which no prototype
character expresses; perl reports `undef` for the prototype of `print`,
`printf`, `say`, `sort`, `exec` and `system`, and silently ignores a
`CORE::GLOBAL::sort` override. perldoc calls `print`'s and `exec`'s
slot an indirect object, but it is not the `indirect` feature: under
`use v5.36` `new Foo` fails while all three still work. A declaration
spells the slot with Raku's invocant colon:

```perl
sub print (FileHandle $fh = select(): List[Str] @args = ($_)) Boolean;
multi sub sort (Code|Str $by: List @list) List;
sub exec (Str $program: List[Str] @args) Boolean;
```

A signature with an invocant colon derives no prototype, matching perl.
No Perl-level sub can have the slot, but a keyword plugin can build it,
so any `.pmt` may declare one (see "One language for every `.pmt`").
Perl's `method` takes `$self` implicitly and never lists an
invocant, so the colon cannot collide with one. For `sort` the slot
names a comparison routine, which may be a string (`sort $n @x` with
`$n = "byname"`); with a comma the string is data instead
(`sort "byname", @x`).

Two cases are open questions (9 and 10 below), not decided: `defined
&f`, whose operand perl does not call, and how to type a parameter
evaluated per element, as `grep EXPR`'s is.

### A slurpy takes no bare element type (*Decided*)

`Str @args` is not allowed (perigrin). It is ambiguous twice over:

- **Which container.** `List[Str]`, the flattened values the caller
  passes, or `Array[Str]`, the copy a body would hold.
- **Which elements.** Whether `Str` applies to each flattened value or
  to the parameter as a whole.

A slurpy (`@` or `%`) is therefore either untyped or typed with an
explicit container type. Container types such as `List[Str]` parse now,
and like all typed Perl only in `.pmt` files (perigrin, 2026-10-02);
what they mean -- variance, flattening, `Array[T]` against `List[T]` --
is provisional until the paper defines it. `print`, with its container
explicit:

```perl
sub print (FileHandle $fh = select(): List[Str] @args = ($_)) Boolean;
```

See "What a parameter type means" for what `Str` asserts there.

### What a parameter type means (*Implemented*)

perigrin, 2026-10-02. A parameter type is the type the operation
coerces its argument to; there is no second, "membership" reading. The
paper defines membership as a lossless round trip (interpreting `v`
through `T` loses nothing, and `v` meets `T`'s operation contracts) and
subtyping as containment of those memberships plus the contracts, so
`A <: B` exactly when an `A` survives `B`'s round trip. An argument of
type `A` fits a parameter of type `T` when `A <: T`; otherwise the
coercion is lossy, which is what `psc check` already reports as
`coercion-mismatch` (`expected Num, got Str`).

What perl then does is the operation's contract, not the type's
meaning. `print $hashref` is a lossy `Ref`-through-`Str` round trip and
perl prints `HASH(0x...)` and continues; `bless "x"` has no coercion to
`Ref` and perl dies. Same reading, two contracts. (Marked *Implemented*
because it states the meaning `psc check` already applies; there is
nothing to build.)

### A typed signature and a prototype say the same thing (*Decided*)

A prototype's characters are parameter types seen from the caller.
`\@` means the caller writes an actual array, passed whole; `@` means
the rest of the call, flattened. That is the `Array`/`List`
distinction (perigrin), so

```perl
sub push (Array \@a, List @list) Int;   # prototype \@@; returns the new length
```

carries `push`'s prototype in its types. Measured on 5.42.0, each
prototype character corresponds to a parameter:

| prototype | the caller writes (measured) | parameter |
|---|---|---|
| `$` | any expression, in scalar context: with `@a = (5, 6, 7)`, `one(@a)` passes 3 | `Scalar $` |
| `@`, `%` | the rest of the call, flattened | `List @` |
| `\@`, `\%` | an actual array or hash, passed whole | `Array \@a`, `Hash \%h` |
| `\$` | any scalar lvalue, passed as a reference: `sref(1)` dies, "must be scalar (not constant item)" | `Scalar \$x` |
| `\[$@%]` | any one of those containers | their union |
| `+` | one array or hash, passed whole, or one scalar: `plus(%h)` sees a HASH, `plus(1,2)` is too many arguments | `Array\|Hash\|Scalar` |
| `&` (first) | a block or a code reference | `Code &` |
| `*` | a bareword filehandle or any scalar: `star(STDOUT)`, `star($s)` | `Glob *` |
| `_` | a scalar, defaulting to `$_` | `Scalar $ = $_` |
| `;` | marks what follows as optional | parameters with defaults |

**Derivation goes both directions** (perigrin):

- A declaration with only a prototype gets the coarse typed signature
  the table gives: `sub foo ($$)` reads as `(Scalar $, Scalar $)`.
- A typed declaration gets its prototype from its types, and needs no
  `:prototype(...)`.
- A declaration with both must have them agree; disagreement is an
  error in the declaration file.

The directions are not symmetric. Types are finer than prototypes
(`Str $` and `Int $` both give `$`), so prototype to types to prototype
round-trips, and types to prototype to types loses precision. A test
can hold the derived prototypes to `prototype("CORE::name")` for all
188 builtins.

In a `.pmt`, a parameter's sigil is the caller's view, as a prototype's
characters are. A backslashed parameter aliases the caller's container
(see "The scalar container"); an unbackslashed `List @` is the rest of
the call, flattened. So in `(Array \@a, List @list)` only the final
parameter is slurpy, where perl's own signatures reject two aggregates
("Multiple slurpy parameters not allowed").

A `List` parameter must be last (perigrin, 2026-10-02). Measured on
5.42, a signature `(@a, $x)` dies "Slurpy parameter not last", and a
prototype `(@$)` warns "Prototype after '@'": its `$` can never be
filled, since the `@` takes every argument. A declaration with
anything after a `List` parameter is an error, which points to
`Array \@a` for a single array followed by more parameters (`\@$`).

### The scalar container (*Decided*)

perigrin, 2026-10-02: a parameter that aliases the caller's container
is written with a backslash, as Perl writes aliasing. perlref,
"Assigning to References" (5.22+): assigning to a reference "performs
an aliasing operation, so that the variable name referenced on the
left-hand side becomes an alias for the thing referenced on the
right-hand side". Measured on 5.42, after `\my @a = \@orig; push @a,
3`, `@orig` holds the 3; after `\my $t = \$s; $t = "y"`, `$s` is `"y"`.
A backslashed parameter says the same: it is an alias for what the
caller wrote, which is exactly what `\$`, `\@` and `\%` deliver.

```perl
sub sref (Scalar \$x);                       # prototype \$
sub push (Array \@a, List @list) Int;        # prototype \@@
sub chomp (List[Str] \(@args = ($_))) Int;   # each element aliased
```

`\(@args)` is perlref's list form, a reference to each element, so
`chomp`'s declaration says it writes through every argument: measured,
`chomp(@l, $x)` chomps both, and bare `chomp` chomps `$_`. `Array @a`
(a container type with a flattening sigil) is not valid. This replaces
an earlier `:lvalue` parameter attribute for the same job.

Measured on 5.42, `\$` accepts any scalar lvalue, not only a variable:
`$x`, `$h{k}`, `$a[0]`, `f()->[0]` and `$x = 7` (the assignment runs,
then its target is passed) each arrive as a `SCALAR` reference, and
`substr($x, 0, 1)` arrives as an `LVALUE` reference, the paper's
`LValueRef`. A constant (`1`, `"str"`) or a sub's result (`f()`) dies:
"must be scalar (not constant item)", "(not subroutine entry)". So the
property is lvalue-ness, not "unevaluated": `$x = 7` is evaluated.

What an aliased list is as a type stays the paper's question.

### Operator declarations (*Decided*)

perigrin, 2026-10-02, **provisional**: these spellings stand for now and
may change as operators are declared.

**No result rule.** `sub + :infix(ADD) (Num $x, Num $y) Num;` is
complete. `Int <: Num` in the lattice, so `Num` is an upper bound that
`Int + Int = Int` satisfies; recovering `Int` for two `Int`s is
inference narrowing within that bound, not something the declaration
states. (Settled earlier with bson; see perl5-son's
`docs/plans/2026-09-16-a-declaration-syntax-for-signatures.md`,
"Problem 1 (WITHDRAWN)".)

**Fixity and precedence.** `:infix(CLASS)`, `:prefix` and `:postfix`.
The class names perl's precedence levels in XS::Parse::Infix's
vocabulary, which already classifies user-defined infix operators that
way (`XPI_CLS_ADD_MISC`, `MUL_MISC`, `POW_MISC`, `LOGICAL_AND_MISC`,
`LOGICAL_OR_MISC`, `ASSIGN_MISC`, `LOW_MISC`, `HIGH_MISC`, and the
predicate classes `RELATION`, `EQUALITY`, `ORDERING`, `MATCHRE`, `ISA`),
written without the `XPI_CLS_` prefix and `_MISC` suffix. The parser's
precedence table stays authoritative, and a test holds each `CORE.pmt`
operator's class to its level there. A library declaring an
XS::Parse::Infix operator uses the same spelling.

**Operators that fork** are multis:

- `..` is a range in list context and a flip-flop in scalar context: a
  `:context` multi.
- `x` repeats a list only when its left operand is parenthesised and
  it is evaluated in list context: `my $x = (1,2) x 2` gives `22`.
  Measured on 5.42: `(1,2) x 2` and `(@a) x 2` give `1 2 1 2`, while
  `"ab" x 2` gives `abab` and `@a x 2` gives `22` (the count, in scalar
  context, repeated as a string). The parenthesis selects the variant,
  a parse fact, as the comma does for `map`:

```perl
multi sub x :infix(MUL) :context(@) (List @l, Int $n) List;   # (LIST) x N
multi sub x :infix(MUL) (Str $s, Int $n) Str;     # EXPR x N
```

### Multi declarations (*Decided*)

Some builtins return different types depending on how they are called.
A declaration may be `multi sub`, giving several signatures for one
name. A call site selects one by arity, by argument types that are not
subtypes of each other, or by context:

```perl
multi sub each (Hash \%h)  List;                   # (Str, value)
multi sub each (Array \@a) List;                   # (Int, value)
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
   `Array[T]`. Until the paper decides, a slurpy in a `.pmt` is
   untyped (see "A slurpy takes no bare element type").
2. *(Resolved: see "What a parameter type means".)*
3. **Lvalue lists.** `chomp` and `chop` modify their arguments in
   place, which neither `List[Str]` nor `Array[Str]` says. A declaration
   spells it `\(@args)` (see "The scalar container"); what such a list
   is as a type is the paper's.
4. **Tuples.** `each` returns `(Str, T)` for a hash, which wants a type
   like `List[Str, T]`.
5. **Multi and context-indexed function types** have no counterpart in
   the paper's function types yet.

Separate from the paper:

6. A search path for user-written `.pmt` files beside the embedded set.
7. The remaining XS::Parse pieces, and sublikes other than `PREFIX`.
8. Lexical rather than file-wide scope for declared syntax.
9. How to type a parameter evaluated once per element, as `grep EXPR`'s
   and `map EXPR`'s first argument is, where every other parameter is
   evaluated once per call.
10. How to declare `defined &f` (and `exists &f`), whose operand names a
    sub that perl does not call (measured: neither calls `f`). Today the
    parser handles the form. perigrin suggests a parameter attribute,
    something like `$x :no_eval`, which has a Perl precedent in the
    class feature's `field $x :param` (attributes after the variable).
    The same attribute family may answer 9 (deferred, repeated
    evaluation is a different property from never).

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
