# RFC 0001: Type libraries

- **Status:** Accepted in part. Sections marked *Implemented* have
  landed; *Decided* sections are agreed and not yet built; *Open*
  sections wait on the types paper. A section is *Implemented* when its
  rule is built and tested; the example declarations a section shows
  (`chomp`, `select`, `localtime`, ...) are illustrations, written into
  `CORE.pmt` and tested against perl by the issues that type the
  builtins (perigrin, 2026-10-05). A section that states a directive
  about specific lines, not an example, is built when those lines are.
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
sub group :prototype(&);
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
sub bless (Ref $ref, Str $class = __PACKAGE__) Object;
sub push (Array \@array, List @list) Int;
sub lock :prototype(\[$@%&*]);
```

A typed line's prototype is the one its types derive (`$;$` and `\@@`
here); a prototype-only line writes its own. A test asks perl for every
keyword's prototype and holds the file to the answer, and another holds
each line's derived prototype to it
(`TestCoreDerivedPrototypesArePerls`). `:prototype(...)` is read as a
prototype wherever a sub is declared, which also fixed two perl.git
files whose own subs use it.

Every line is typed but eight, which stay prototype-only. The six that
hold a glob slot -- `lock`, `pos`, `tie`, `tied`, `undef`, `untie` --
wait on its spelling (open question 11). `catch` and `method` are
keywords, like `try` and `sub`, not builtins to type (perigrin,
2026-10-07): no call reaches them, and their `()` lines are there only
because `prototype("CORE::catch")` answers. `isa`, which perl also
answers `()` for, is the infix operator, typed by its `:infix`
declaration ("Operator declarations").

The file also declares the builtins perl reports no prototype for --
`print`, `defined`, `grep`, `sort` and the rest ("Builtins with no
prototype", "Builtins that keep their own parse"). Their lines derive
none, so they stay out of the prototype table, where a name's presence
would read as a prototype it does not have; their typed signatures are
read beside the table's.

### Declaration order: Perl's (*Implemented*, 098211c6)

A declaration follows Perl's order: name, then attributes, then
signature. Measured on 5.42, `sub f :lvalue ($x) {}` compiles,
`sub f ($x) :lvalue {}` dies ("Subroutine attributes must come before
the signature"), and `sub :lvalue f {}` is no declaration: perl reads
the label `sub:` and the indirect call `'f'->lvalue({})`, which `use
v5.36` refuses as a syntax error. A `.pmt` reports any statement it
cannot read as "not a declaration". An operator
is therefore `sub + :infix (Num $x, Num $y) Num;`.

### Typed Perl, in `.pmt` only (*Implemented*, a3c0200c, 3247bd4a, 26f0f472, b695e2b4, 0cfefff4, acc10ef3, 5a22ba3b, d2094988)

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

`CORE.pmt` is the builtin table. `internal/infer` reads each builtin's
types from it, and the parser derives each builtin's parse shape from it
-- named unary, list operator or niladic -- from its prototype, its
`:unary` or its `:listop`; `TestCoreDerivedShapes` holds the derived
shapes to the sets perl was measured to give. The comments recording
those measurements are on `CORE.pmt`'s lines. No hand-kept table of
builtins remains beside it, of types or of parse shapes.

### Type names (*Implemented*, 02df00ad, cd485cd0, b2c809fc, 54784319)

A `.pmt` names types from the lattice in `internal/types`, and an
unknown name is an error. perigrin, 2026-10-02:

- **Unions are spelled `A|B`**, as `Str|Undef`. The lattice is a bitmask,
  so a union is an OR, and it is what Moose and Type::Tiny users already
  write. Spaces around `|` are allowed (`Str | Undef`), as Moose allows
  them. `None`, the bottom type, may be written: it is the return type
  of a sub that never returns, such as `exit`.
- **`IO` joins the lattice** (the paper has it), and **`FileHandle` is a
  named union, `Glob|GlobRef|IO`**, defined in the lattice as `Ref` and
  `Scalar` are, so a `.pmt` needs no alias syntax. Measured on 5.42: a
  bareword handle is a `Glob`, `open my $fh` gives a `GLOB` reference,
  and `IO` names what a glob's IO slot holds -- the slot type, at the
  top of the lattice beside `Code`, never a value. The value
  `*STDOUT{IO}` is a reference to it, of reftype `IO` blessed into
  `IO::File`, and measures as an `Object`, as `qr//` does (the lattice
  oracle witnesses it on its `Object <: Ref` edge). A string naming a handle is not a
  `FileHandle`: under `use strict`, `print {"STDOUT"} ...` dies ("Can't
  use string ("STDOUT") as a symbol ref"). A blessed handle such as an
  `IO::File` object is reftype `GLOB` but files under `Object`, so it is
  not a `FileHandle` until the paper says how blessing and reftype
  combine.
- **Names follow the paper.** The lattice's `Bool` is the paper's
  `Boolean`, and a `.pmt` writes `Boolean`.
- **`Void` joins the lattice** (perigrin, 2026-10-08), the paper's
  `Void := {()}`: arity 0, one inhabitant, the empty list, what returns
  and yields nothing. It sits under `List` beside `Scalar`, and is not
  `None`, which does not return. Scalar context coerces it to `Undef`:
  measured on 5.42, `f(())` under prototype `+` passes one undef. pvm's
  `List` is `Array|Hash|Scalar|Void`, so `Void|Scalar` is under `List`
  rather than equal to it as in the paper.
- **`Maybe[T]` is `Undef|T` and `Optional[T]` is `Void|T`** (perigrin,
  2026-10-08; the paper's "Absent and undefined"). They are names for
  those unions, not types of their own, and take exactly one type.
  Measured on 5.42, `srand(undef)` warns "Use of uninitialized value"
  and seeds 18446744073709551615, an `Int` slot holding undef,
  `Maybe[Int]`; `srand()` seeds itself, a slot holding nothing,
  `Optional[Int]`.
- **A parameter whose type includes `Void` may be absent, with no
  default** (perigrin, 2026-10-08). `sub srand (Optional[Int] $seed)
  Str;` says what perl does with no seed, which no expression states, and
  derives `;$`: the table puts a `;` before it. A written argument is
  never absent, so it is held to the type less `Void`: `srand("x")` is
  told `Int`, and `srand(@e)` is `srand(0)`, the array one scalar under
  `;$`. `Any` is the permissive top, as TypeScript's `any` is beside its
  `void` (`f(x: number | void)` allows `f()`, `f(x: any)` does not), not
  a union that happens to hold `Void`, so an `Any` parameter is required.
  `List` is such a union, so `List $` derives `;+`. A slurpy takes zero
  arguments already, and `Optional[Int] $x = die` is an error: `die`
  requires what `Void` says may be absent.
- **The paper's other types** (`VString`, `Format`, `LValueRef`) join
  the lattice when a declaration first needs one.

### One language for every `.pmt` (*Implemented*, 79191842, d1338cd2, b6bea9c6)

perigrin, 2026-10-02: everything `CORE.pmt` can say, any library's
`.pmt` can say too -- typed signatures, `multi` and its context forks,
the invocant colon, declared syntax. `CORE.pmt` is the declaration file for
the interpreter, not a dialect of its own. A construct no Perl-level sub
can have (the invocant colon) is still declarable for a library,
because a keyword plugin can build what a sub cannot.

A library's `.pmt` is held to `CORE.pmt`'s rules, and its errors name
the module. The one difference is the operator classes coined for
`CORE.pmt` (see "Operator declarations"), which a library may not name.
A sub a library exports is imported with its typed signatures.

Typed Perl is meant to outlive pvm's parser: Chalk is to read the same
`.pmt` files in time, so the syntax should stay something a second
implementation can parse from this RFC alone.

### Builtins with no prototype (*Implemented*, 3d340622, d1338cd2, 0cfefff4, e00fc501, 79a16715)

About 20 builtins have no prototype perl can report. `CORE.pmt`
declares their types, and `:unary` marks the eight that are named
unaries; a builtin with no prototype is otherwise a list operator, as an
ordinary sub is. Most of their forms turn out to be declarable; see
"Builtins that keep their own parse" for which, and how.

```perl
sub print (FileHandle $fh = select(): List[Str] @args = ($_)) Boolean|Undef;
```

The default handle is the selected one, not STDOUT, and `print` with no
arguments prints `$_`; to a closed handle it returns undef (measured).

**`:listop` types a list operator's positional parameters** (perigrin,
2026-10-07: "go with :listop for now"). It is `:unary`'s sibling: a line
stating it is a builtin with no prototype that parses as a list operator,
and its types derive no prototype. `split` needs it. `prototype("CORE::split")`
is `undef` and `split $a, $b` deparses as `split(/$a/, $b, 0)`, a list
operator, yet its parameters are positional and typed: the string is
taken in scalar context (`split /,/, @a` over two elements splits "2"),
a fourth argument is "Too many arguments for split", and with no
arguments it splits `$_` on whitespace. Without `:listop` its types would
derive `;$_$`, a prototype perl does not report, and `(List @args)`, which
derives `@`, says the string is flattened, which perl shows is false.

```perl
multi sub split :listop (Regex|Str $pattern = ' ', Str $string = $_, Int $limit = 0) Int;
multi sub split :listop (Regex|Str $pattern = ' ', Str $string = $_, Int $limit = 0) List;
```

A `:listop` line is held as a `:unary` one is. A `:prototype(...)`
beside it is an error, since it derives none; so is `:listop` with no
typed signature, which would leave it nothing to say, and `:listop`
beside `:unary`, two parses for one builtin. A multi's candidates derive
none when any of them states it. The `CORE.pmt` check, which reads
perl's `undef` as a derived `@`, has no derived prototype of a `:listop`
line to hold to perl's. A library `.pmt` states `:listop` as `CORE.pmt`
does, under the same rules ("One language for every `.pmt`"), for an
XS sub with no prototype but typed positional parameters.

### Builtins that keep their own parse (*Implemented*, eb839b7d, 2bccec6f, cb25b36c, 828544ec, c1107251, efed93fd, b8e22b93, db6c2483, 6556db7f)

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
multi sub grep (Code &block, List @list = die) List;    # &@
multi sub grep (Scalar $expr, List @list) List;         # $@
multi sub map  (Code &block, List @list = die) List;
multi sub map  (Scalar $expr, List @list) List;
multi sub sort (Code &block, List @list = die) List;
multi sub sort (List @list = die) List;
```

After a block the list must be written, though it may be empty:
`grep {1} ()` and `sort { $a <=> $b } ()` compile, while `grep {1};`,
`map({1})` and `sort { $a <=> $b };` are syntax errors, so the block
candidates' lists default to `die` ("A required argument defaults to
`die`"). After grep's or map's expression it may be left out:
`grep(1)` compiles.

**`map`'s and `grep`'s brace is variant selection.** Both guess whether
`{` opens a block or a hash constructor from its first tokens; a `&@`
sub never guesses. Wherever perl's guess disagrees with where the
first argument ends, perl rejects the program: `map { "\L$_" => 1 } @a`
(guessed hash, no comma) and `map { $_ => 1 }, @a` (guessed block,
then a comma) are both syntax errors. So the comma that ends the first
argument selects the expression variant and its absence the block
variant. That comma need not follow `}` directly: perl 5.42 accepts
`map { foo => 1 }->{foo}, @a`, `map { "a" => 1 } ? 1 : 2, @a` and
`map {}->{a}, @a`, a hash constructor continued before the comma, and
refuses `map { $_ => 1 }->{a}, @a`. Measured over 46 forms, perl
refuses wherever its guess and that comma disagree, so on every
program perl accepts the two pick the same variant. (It refuses more
besides: after a hash constructor's `}` perl expects a term, so
`map { "a" => 1 } + 1, @a` is a syntax error while `. "x"`, `|| 1`
and `->{a}` in the same place are not.) The first
argument's type selects the variant, as in Chalk, decided by that comma
rather than by the braces' contents: `map { $_ => 1 } @a` holds pairs
and is still a valid block. The parser keeps perl's first-tokens guess,
with an empty `{}` read as a hash as toke.c reads it, only to refuse
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
sub print (FileHandle $fh = select(): List[Str] @args = ($_)) Boolean|Undef;
multi sub sort (Code|Str $by: List @list = die) List;
sub exec (Str $program: List[Str] @args) Boolean;
```

A signature with an invocant colon derives no prototype, matching perl.
No Perl-level sub can have the slot, but a keyword plugin can build it,
so any `.pmt` may declare one (see "One language for every `.pmt`").
Perl's `method` takes `$self` implicitly and never lists an
invocant, so the colon cannot collide with one. For `sort` the slot
names a comparison routine, which may be a string (`sort $n @x` with
`$n = "byname"`); with a comma the string is data instead
(`sort "byname", @x`). Its list must be written too: with none after
it the routine is the list, `sort byname` yielding "byname" while
`sort byname ()` yields nothing (measured).

Two cases are open questions (9 and 10 below), not decided: `defined
&f`, whose operand perl does not call, and how to type a parameter
evaluated per element, as `grep EXPR`'s is.

### A slurpy takes no bare element type (*Implemented*, 26084781, 54784319)

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
sub print (FileHandle $fh = select(): List[Str] @args = ($_)) Boolean|Undef;
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

### A typed signature and a prototype say the same thing (*Implemented*, b85929e1, d7a0face, 5ab639a7, 74f46579, 37153b30, 7a0ea048, ffaddd08, 56a13978, 521fb37e)

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
| `@`, `%` | the rest of the call, flattened | `List @`, `List %` |
| `\@`, `\%` | an actual array or hash, passed whole | `Array \@a`, `Hash \%h` |
| `\$` | any scalar lvalue, passed as a reference: `sref(1)` dies, "must be scalar (not constant item)" | `Scalar \$x` |
| `\[$@%]` | any one of those containers | their union |
| `+` | one array or hash, passed whole, or one scalar: `plus(%h)` sees a HASH, `plus(1,2)` is too many arguments | `Array\|Hash\|Scalar $`, exactly; `plus(())` passes one undef. That and `Void`, `List $`, may be absent, and is `;+`: under `(;+)` `g()` passes nothing. A wider type such as `Any $` is a `$` |
| `&` | a block or a code reference when first; elsewhere `sub {...}` or `\&name` | `Code &` |
| `*` | a bareword filehandle or any scalar: `star(STDOUT)`, `star($s)` | `Glob *` |
| `_` | a scalar, defaulting to `$_` | `Scalar $ = $_` |
| `;` | marks what follows as optional | parameters with defaults, or whose type includes `Void` |

**Derivation goes both directions** (perigrin):

- A declaration with only a prototype gets the coarse typed signature
  the table gives: `sub foo ($$)` reads as `(Scalar $, Scalar $)`.
- A typed declaration gets its prototype from its types, and needs no
  `:prototype(...)`.
- A declaration with both must have them agree; disagreement is an
  error in the declaration file.

The directions are not symmetric. Types are finer than prototypes
(`Str $` and `Int $` both give `$`), so prototype to types to prototype
round-trips, and types to prototype to types loses precision. The one
exception is a trailing `;`, which no type states: measured on 5.42,
`f 1, 2` under `($;)` is "Too many arguments" -- a list operator --
while under `($)` it reads `f(1), 2`, and `g 1` under `(;)` is "Too
many arguments" where `()` is a syntax error. So `$;` and `;` come back
from types as `$` and nothing; a declaration that needs them writes
`:prototype($;)`, which agrees with `(Scalar $x)` and is the prototype
kept (`not` and `getprotobynumber` in CORE.pmt). A test
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

The rule is there for list flattening, so the operators whose comma
does the flattening are its exceptions (perigrin, 2026-10-08): `,`,
`=>` (its quoting form) and `x`. "There is absolutely no way to
distinguish `(@a, @b)` from `(@a)` in Perl": the comma is the
flattening, so the rule about flattened arguments does not bind it.
Measured on 5.42, `(@a, 3)` with `@a = (1, 2)` is 3 elements, the left
operand flattening too; `(1, (2, 3))` is 3 elements; and `my $x = (4,
5)` is 5, with a "Useless use of a constant" warning. So `,` is

```perl
multi sub , :infix :looser(=) :assoc(left) (List @l, List @r) List;   # list context: append
multi sub , :infix (Scalar $l, Scalar $r) Scalar;                     # scalar context: the right
```

`x` is one for the same reason: its list candidate's left operand is a
parenthesised list its comma flattens (see "Operator declarations").
The exceptions are named, not operators in general: any other operator,
and every sub, with a non-final `List` parameter is a declaration error
(TestCommaIsTheFinalListException).

### The scalar container (*Implemented*, 1a6f3e93, 7a0ea048, 3c6a73bb, e6fba74e, 352d0eca)

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
sub any (Code \&block, List @list) Boolean;  # prototype \&@
```

`\(@args)` is perlref's list form, a reference to each element, so
`chomp`'s declaration says it writes through every argument: measured,
`chomp(@l, $x)` chomps both, and bare `chomp` chomps `$_`. `Array @a`
(a container type with a flattening sigil) is not valid. A code slot
is `Code \&c`: perlref lists `\&sub` among the forms refaliasing
accepts (`\&foo = \&bar` makes `foo()` mean `bar()`, measured). A glob
slot (`\*`, inside `\[$@%*]` for `tie`, `tied`, `untie`, `lock`,
`undef` and `pos`) has no spelling yet (perigrin, 2026-10-03): perl
cannot alias a glob this way -- `\*G = \*STDOUT` is a compile error,
"Can't modify reference to ref-to-glob cast" -- so `Glob \*g` would be
an extension rather than Perl's syntax. Until it is decided those six
builtins keep prototype-only lines in `CORE.pmt`, untyped. This replaces
an earlier `:lvalue` parameter attribute for the same job.

Measured on 5.42, `\$` accepts any scalar lvalue, not only a variable:
`$x`, `$h{k}`, `$a[0]`, `f()->[0]` and `$x = 7` (the assignment runs,
then its target is passed) each arrive as a `SCALAR` reference, and
`substr($x, 0, 1)` arrives as an `LVALUE` reference, the paper's
`LValueRef`. A constant (`1`, `"str"`) or a sub's result (`f()`) dies:
"must be scalar (not constant item)", "(not subroutine entry)". So the
property is lvalue-ness, not "unevaluated": `$x = 7` is evaluated.

What an aliased list is as a type stays the paper's question.

### Operator declarations (*Implemented*, d76d94b1, b5b1f905, 94d4c5c2, 31cf13bc, edbd7955, 0a726b85, a6fcb6e8, 1776c37f)

perigrin, 2026-10-02, **provisional**: these spellings stand for now and
may change as operators are declared.

**No result rule.** `sub + :infix(ADD) (Num $x, Num $y) Num|Inf;` is
complete. `Int <: Num|Inf` in the lattice, so `Num|Inf` is an upper bound
that `Int + Int = Int` satisfies; recovering `Int` for two `Int`s is
inference narrowing within that bound, not something the declaration
states. (Settled earlier with bson; see perl5-son's
`docs/plans/2026-09-16-a-declaration-syntax-for-signatures.md`,
"Problem 1 (WITHDRAWN)".)

**A result is what perl returns.** perigrin, 2026-10-07. An operator's
operands are the type it coerces them to, and its result is what perl
returns for operands of those types, which may leave them. The paper's
`[Plus]` rule says the same: its premises `v ⇓^Num n` type the operands
through `Num`, and its conclusion is only the number `n₁ + n₂`, which
`Num` need not hold, since `Num` excludes `Inf` and `NaN` (Theorem 3).
Measured on 5.42 with finite operands, `1e308 + 1e308` is `Inf`,
`-1e308 * 10` is `-Inf` (the lattice's `Inf` is either sign) and
`(-1) ** 0.5` is `NaN`, so `+`, `-`, `*` and `/` are declared
`(Num $x, Num $y) Num|Inf` and `**` `(Num $x, Num $y) Num|NaN|Inf`;
`%` stays `Num`. The builtins follow the same rule: `exp` is `Num|Inf`,
`hex` and `oct` `Int|Inf`. A case that dies, `1/0` or `sqrt(-1)`, is no
value and widens nothing. `TestCoreArithmeticReturnsWhatPerlReturns`
holds each declaration to its measurement.

A result of `Num|Inf` passed where `Num` is wanted is a lossy coercion,
and `psc check` reports it: `my $c = $a * $b; $c + 1` warns
`left operand of "+": expected Num, got Int|Num|Inf`.

**Fixity and precedence.** `:infix(CLASS)`, `:prefix` and `:postfix`.
The class names perl's precedence levels in XS::Parse::Infix's
vocabulary, which already classifies user-defined infix operators that
way (`XPI_CLS_ADD_MISC`, `MUL_MISC`, `POW_MISC`, `LOGICAL_AND_MISC`,
`LOGICAL_OR_MISC`, `LOGICAL_AND_LOW_MISC`, `LOGICAL_OR_LOW_MISC`,
`ASSIGN_MISC`, `LOW_MISC`, `HIGH_MISC`, and the predicate classes
`RELATION`, `EQUALITY`, `ORDERING`, `MATCHRE`, `ISA`), written without
the `XPI_CLS_` prefix and `_MISC` suffix: `and` is `LOGICAL_AND_LOW`,
`or` and `xor` are `LOGICAL_OR_LOW`. A class is a relation, `:equiv` of
the operator it stands for the level of, and the parser's precedence is
derived from the relations (see "Precedence is a relation between
operators"). A library declaring an XS::Parse::Infix operator uses the
same spelling.

XS::Parse::Infix classes no operator at the levels of `&`, of `|` and
`^`, or of `<<` and `>>` (perigrin, 2026-10-06), nor at that of `..`
and `...`. Those levels have no class: they are named by their
operators, and their lines state their relations (see "Precedence is a
relation between operators") -- `sub & :infix :tighter(|)
:assoc(left)`, `sub ^ :infix :equiv(|)` -- so only XS::Parse::Infix's
own class names are spellable in `:infix(CLASS)`, and a library naming
another is in error. XS::Parse::Infix is an influence, not a limitation
(perigrin, 2026-10-08): a library's infix operator may take any level
its relations give, `:equiv` of any operator, those levels included
(`sub ⊕ :infix :equiv(&)` binds as `&`), or a level of its own between
two (`:tighter(+) :looser(*)`). A library's bare `:infix`, with no class
and no relation, places its operator at no level and is a declaration
error naming it.

**Operators that fork** are multis:

- `..` is a range in list context and a flip-flop in scalar context: a
  multi whose return types fork on context.
- `x` repeats a list only when its left operand is parenthesised and
  it is evaluated in list context: `my $x = (1,2) x 2` gives `22`.
  Measured on 5.42: `(1,2) x 2` and `(@a) x 2` give `1 2 1 2`, while
  `"ab" x 2` gives `abab` and `@a x 2` gives `22` (the count, in scalar
  context, repeated as a string). The parenthesis selects the variant,
  a parse fact, as the comma does for `map`:

```perl
multi sub x :infix(MUL) (List @l, Int $n) List;   # (LIST) x N
multi sub x :infix(MUL) (Str $s, Int $n) Str;     # EXPR x N
```

  The call site states each operand's shape, `@` for a parenthesised
  list and `$` otherwise. A `@` parameter takes only a `@` operand, and
  a candidate taking each `@` operand as a list is selected before one
  taking it as a scalar; without that, `Str <: List` would make the
  `Str` candidate the most specific. Context comes first: in scalar
  context the `Str` candidate's return answers it, so `my $x = (1,2) x
  2` is the `Str` one's. The `@` parameter comes before another:
  `x` is one of the operators the final-List rule names as exceptions
  (see "A typed signature and a prototype say the same thing").
- `=~` forks on context: in list context a match is its captures and
  `s///g` its count; in scalar context a match is a boolean, `s///` and
  `tr///` a count, and `s///r` and `tr///r` the new string.
- `\` forks on a parenthesis as `x` does: `my @r = \(1,2,3)` is three
  references, `my @r = \@a` one.

Which operands are `@`-shaped is the operator's, for "Call sites" to
apply. Measured on 5.42, `qw(a b) x 2` is `a b a b`, a `qw` list shaped
as a parenthesised one; and to `\` a sub call is a list too, `my @r =
\f()` being a reference to each value `f` returns, where `f() x 2`
repeats a string.

**`:bareword`** is a parameter attribute, and the only one a `.pmt`
parameter takes (perigrin, 2026-10-08). It is a parser hint, not a type
coercion: the operand is read as a word, not an expression, and the
parameter's type says what the word becomes. `=>` is the comma whose
left operand is one:

```perl
multi sub => :infix :equiv(,) (Str $lhs :bareword, List @rhs) List;
multi sub => :infix :equiv(,) (Str $lhs :bareword, Scalar $rhs) Scalar;
```

The type alone does not do it, because the quoting decides what the
operand is before any type applies. Measured on 5.42 under `use
strict`, with `sub foo { "CALLED" }`: `(foo => 1)` gives "foo" where
`(foo, 1)` gives "CALLED"; `(time => 1)` gives "time" where `(time, 1)`
gives the time; and `(nosuch => 1)` gives "nosuch" where `(nosuch, 1)`
is "Bareword not allowed while strict subs". The lexer reads which
operators have a `:bareword` left operand from `CORE.pmt`: a program
is lexed with them, so `s => 1` is the word `s` and not a substitution
and `-e => 1` the word `-e` and not a file test, and `CORE.pmt` itself
is lexed with the lexer's own list, which a test holds to `CORE.pmt`'s
(TestFatCommaAutoquotes, TestLexerBarewordOperatorsAreCores). Elsewhere
the parser treats `=>` as the comma it is `:equiv` to.

### Precedence is a relation between operators (*Decided*)

The goal (perigrin, 2026-10-08): every builtin and operator is defined
with a signature in `CORE.pmt` that lets it be parsed, and the parser
is derived from `CORE.pmt` -- as its keyword shapes already are. An
operator's precedence is part of that definition.

pvm and Chalk each kept a precedence table by hand -- pvm's in
`internal/parse/precedence.go` with a class map in
`internal/parse/operators.go`, Chalk's
`Chalk::Grammar::Perl::PrecedenceTable` feeding its Precedence
semiring -- and the two disagreed: Chalk gives `isa` its own level
tighter than the relational operators, as perlop does, while pvm put
it beside them, as perly.y does (toke.c lexes `isa` as an NCRELOP).
pvm's is now a reading of `CORE.pmt`: its binding powers are derived
from the relations, with `isa` in perlop's row. Measured on 5.42, the
two placements differ only in what perl refuses: `$x isa Foo < 1` and
`$a < $b isa Foo` are both syntax errors, so an unparenthesised `isa`
and a relational operator never stand together, and the parser refuses
both orders. The relations are plain Perl any other implementation can
read, and whether one does is that implementation's own work.

**Precedence is a partial order, stated on each operator** (perigrin,
2026-10-08), as Raku states it with `is tighter`, `is looser` and
`is equiv`, rather than a numbered table: a library that declares an
infix operator must be able to place it between two existing levels
without renumbering anything. An operator line relates itself to an
operator it can see:

```perl
sub * :infix :tighter(+) :assoc(left) (Num $x, Num $y) Num|Inf;
sub + :infix :assoc(left) (Num $x, Num $y) Num|Inf;
sub - :infix :equiv(+) (Num $x, Num $y) Num|Inf;
sub . :infix :equiv(+) (Str $x, Str $y) Str;
sub ⊕ :infix :tighter(+) :looser(*) :assoc(left) (Num $x, Num $y) Num;
```

- `:tighter(OP)` and `:looser(OP)` order the operator's level against
  another operator's; `:equiv(OP)` puts it in that operator's level,
  whose associativity it takes.
- `:assoc(...)` is perlop's: `left`, `right`, `nonassoc`, `chained`
  (`<` `<=` ...) or `chain_na` (`==` `!=` ...). A level states it once.
- XS::Parse::Infix's classes remain a spelling of a level:
  `:infix(ADD)` is `:infix :equiv(+)`, so a library writes the class it
  registers with. Levels XS::Parse::Infix does not class (`&`, `|`
  and `^`, `<<` and `>>`, `..` and `...`) need no name of their own:
  they are named by their operators.
- A `:prefix` operator relates the same way: `not` states that it is
  tighter than `and`, and `!` that it is tighter than `=~`.
- **A named operator's level is derived, not stated** (perigrin,
  2026-10-08), from its shape, as the shape is from its prototype. This
  is perl's own rule, measured on 5.42 with user subs: a `($)`, `(_)` or
  `(;$)` prototype makes `u 1 < 2, 3` parse as `(u(1) < 2), 3`, a named
  unary tighter than `<`; `(@)` makes it `u(1 < 2, 3)`, a list operator;
  `()` makes `u 1` a syntax error, a term. The builtins agree: `defined
  $x < 2`, `ref $x < 2`, `chdir $x < 2` and `sleep $x < 2` are each
  `(op($x) < 2)`. So a builtin's line carries no relation: its `:unary`
  or `:listop`, or failing those the prototype its signature derives,
  puts it in perlop's row -- a named unary between `<<`/`>>` and `isa`,
  a rightward list operator below `,`, a term above everything -- and
  `:tighter`, `:looser` and `:equiv` are for the symbolic and
  word-shaped operators, the `:infix` and `:prefix` lines. The control
  words are the stated exception: `goto` and `dump` have no prototype
  that places them, and perl parses their operand at the assignment
  level (`goto $x = 1` is `goto ($x = 1)`, `goto $x, 1` is `(goto $x),
  1`), so their lines state `:equiv(=)`. A builtin that is also an
  operator, `not`, has the operator's level.

**The parser derives a total order** by sorting the relations
topologically. A `.pmt` whose relations form a cycle, or leave two
levels unordered where a parse needs an answer, is a declaration error
naming the operators. `CORE.pmt`'s relations reproduce perlop's table
exactly, and a test holds the derived order and associativity to
perlop's own table. Every operator perlop's table names that is an
infix, prefix or postfix operator has a line, including `^^`, `&.`,
`|.`, `^.`, `~.`, the compound assignments, `,` and `=>`, and
`++`/`--`. One with no line yet is placed by the same relations, kept
beside the parser (`undeclaredOperators` in
`internal/parse/operators.go`) until its line is written. A library's own operator joins the same order, so
`sub ⊕ :infix :tighter(+) :looser(*)` parses between `+` and `*`
wherever the library is in scope: from its import to the end of the
file, as its declared syntax is, and `use M ()` brings none. A library
whose relations form a cycle with `CORE.pmt`'s, or two libraries in
scope whose operators are left unordered, are a declaration error
naming the operators. An operator may be spelled with a non-ASCII
symbol, `⊕`, as an operator plugin may register one.

### Multi declarations (*Implemented*, d7ff54be, e0927971, 597f89f2, 56fdf0e2, 2a00759b, 1f2f170c, 263ec15e, d44fb222, 7341cfec, 7e129204)

Some builtins return different types depending on how they are called.
A declaration may be `multi sub`, giving several signatures for one
name. The selection rules pick one by arity, by argument types, or by
context; applying them at a real call in `infer` is "Call sites" below:

```perl
multi sub each (Hash \%h)  List;                   # (Str, value)
multi sub each (Array \@a) List;                   # (Int, value)
multi sub select () Str;                           # the selected handle
multi sub select (FileHandle $fh) Str;             # the previous handle
multi sub select ($r, $w, $e, Num $timeout) List;  # (nfound, timeleft)
```

When the call site cannot decide, the consumer joins the candidates.

**A name declared both as `sub` and as `multi sub` is a multi, with a
warning** (perigrin, 2026-10-07). Every candidate is kept, in the order
declared, whichever line comes first, and the declaration file carries
the warning `sub f: declared both as sub and as multi sub; read as a
multi`; it is no error, and drops nothing. A plain `sub f` still
replaces an earlier plain `sub f` when the file declares no `multi sub
f`. A declaration file's warnings reach the parse root beside its
errors, each naming its module.

**The most specific candidate wins** (perigrin, 2026-10-02). When more
than one candidate accepts a call because their parameters are related
by subtyping -- `multi sub f (Int $x) Str;` beside `multi sub f (Num $x)
Int;`, called with an `Int` -- the candidate whose parameters are all
subtypes of the other's is selected, as multi dispatch does in Raku,
CLOS and Julia; the lattice's `Int <: Num` makes "more specific"
precise. Candidates with no single most specific one for some argument
types -- `(Int $a, Num $b)` beside `(Num $a, Int $b)`, called with two
`Int`s -- are ambiguous, and the declaration is an error. A narrower
candidate whose result is only what narrowing within the wider one's
bound would give (`(Int $x) Int` beside `(Num $x) Num`) is redundant,
and `psc` may warn about it.

**A call whose arity no candidate accepts fails** (perigrin,
2026-10-02, stricter than perl where perl only finds out at run time).
The selection rules report it as a failure, never a join. The parser's
half is "Refusing what perl refuses"; the `psc check` half is "Call
sites".

**A call's type is the selected candidate's declared return type**
(perigrin, 2026-10-06). An `Int` argument reaches a `Num` parameter
through `Int <: Num`, so `sqrt (Num $x) Num` and `abs (Num $x) Num` of
an `Int` are each a `Num`, which is sound. An earlier draft typed the
result `meet(join(arguments), declared)`, which would have made
`sqrt`, `atan2` and `chr` of `Int`s `Int`s. Measured on 5.42.0 with
`B::svref_2object`, `sqrt(2)`, `sqrt(4)` and `atan2(1,1)` are floats
(NOK only), and `chr(65)` is a string (POK only).

### Context selects by return type (*Implemented*)

A function whose type depends on its calling context (one that
effectively branches on `wantarray`) is a multi whose candidates return
different types (perigrin, 2026-10-08). The paper calls this
return-type dispatch: the call site's context is a demand on the
result, and the candidate whose return type meets it is selected. Each
return type answers one context:

| return type | answers |
|---|---|
| `Void` | void |
| a list: one holding `Array` or `Hash`, as `List`, `List[T]`, `Array` and `Hash` do | list |
| any other | scalar |

"A list" is not a subtype test: `Scalar <: List`, so every scalar
return type is under `List` too. Context is still decided at the call
site, syntactically; what says which candidate answers it is the
return type. Each of the 37 `CORE.pmt` candidates that stated a
`:context` attribute returns a type answering the context it stated.

```perl
multi sub localtime :prototype(;$) (Int $time = time) Str|Undef;
multi sub localtime :prototype(;$) (Int $time = time) List[Int];
```

Context selects a declaration; narrowing cannot stand in for it.
`my $t = localtime(0)` is a date string, while the first element of the
list result is `0`, an `Int`.

Of the candidates that take a call, those whose return type answers its
context are selected among, before the most specific is sought. When
none does, every one is, and the context coerces the result as it
coerces any value: `grep` returns only a list, so a scalar `grep` is its
`List`, coerced. Where perl's scalar answer is no coercion of the list,
a scalar candidate states it: measured on 5.42, every form of a scalar
`sort` is undef, so each of sort's list candidates has a twin with the
same parameters returning `Undef`. Void context is a form of scalar context
(perlglossary, "void context"), so a call in void context with no `Void`
candidate is the scalar candidate's. Two candidates with the same
parameters whose return types answer the same context share every call,
so they are ambiguous, a declaration error; a scalar return beside a list
one is a fork.

Boolean context counts as scalar. Measured on 5.42, `wantarray` reports
scalar inside `if (f())`, `!f()` and `f() and ...`, so no Perl-level
sub can tell them apart.

A `:context(...)` attribute is a declaration error naming it: the
return type already says which context a candidate answers, and an
attribute beside it could only repeat it or contradict it.

### Refusing what perl refuses (*Implemented*, 1f6ae965, 98a376b3, 48c2da55, 7ff4e757, 335cccdf, 1ef08a0a, 4488a1ae)

perigrin, 2026-10-02. Where perl refuses a call at compile time, the
parser refuses it too, for prototypes read from module source as well
as for `.pmt` declarations. Measured on 5.42:

- A call whose arity no declaration accepts: `select(1, 2)` dies "Not
  enough arguments for select system call", `localtime(1, 2)` "Too many
  arguments for localtime", and `each()`, `sort()` and `grep()` "Not
  enough arguments".
- A constant or a sub's result passed to a `\$` slot: `sref(1)` dies
  "Type of arg 1 to main::sref must be scalar (not constant item)",
  `sref(f())` "(not subroutine entry)".

Some builtins parse more loosely than their declarations say, and the
declaration cannot carry both: `not()` compiles and is true, though
`sub not :prototype($;) (Scalar $x)` requires the argument, because
perl's `$;` means one thing for a user sub and another for `not`.
Measured by calling every CORE.pmt builtin with none to six arguments,
such a builtin is refused only on the side perl refuses: `close($x, $x)`
compiles and `closedir()` does not; `system()` compiles and `do($x, $x)`
does not; and `fc($x, $x)` without its feature is a user's sub.

Where perl would compile the call -- a candidate set with no prototype,
or a library dispatching at run time -- the parse stays as perl reads it
and `psc check` reports an error ("Call sites").

**A required argument defaults to `die`** (perigrin, 2026-10-05). In a
perl signature a default runs only when its argument is omitted, so
`$x = die "..."` makes `$x` required (measured: `f(1)` is 1, `f()`
dies). A declaration says the same: `multi sub sort (List @list = die)
List;` -- `sort()` is refused, while `sort(())` and `sort @e` write an
argument and pass, which is perl's rule (measured: only `sort()` among
`push(@a)`, `die()`, `reverse()`, `unlink()`, `return()` and the like
fails to compile). A `List` parameter with no default accepts zero
arguments, as perl's own slurpy does (`sub h (@l)`, `h()` is 0), so no
other declaration needs anything beyond the block forms of grep, map
and sort and sort's invocant form, whose list perl requires too
("Builtins that keep their own parse"). No special case for `sort`
remains.

A default on a slurpy parameter is a typed-Perl extension: perl's
signatures reject it ("A slurpy parameter may not have a default
value"). `.pmt` declarations use it for `= die` here and for `print`'s
`= ($_)`.

**An argument perl computes when it is absent has no default**
(perigrin, 2026-10-08). `srand` with none calls perl's own `Perl_seed()`
(util.c), which reads entropy -- `getentropy`, else `/dev/urandom`, else
a hash of the time, pid and stack -- `sleep` with none sleeps for ever,
`caller` with none gives its three-field frame. No expression states
those, so the parameter is `Optional[T]`, which may be absent with no
default: `sub srand (Optional[Int] $seed) Str;`, likewise `umask`,
`sleep`, `caller`, `reset`, `send`'s address and `substr`'s
replacement. A default may also name a parameter declared after it:
`exec`'s and `system`'s program slot is optional, the program taken from
the list (`exec "ls"` compiles), so `sub exec (Str $program = $args[0]:
List[Str] @args) Boolean;`.

### Call sites (*Decided*)

perigrin, 2026-10-02. "Multi declarations" and "Context selects by
return type" cover
declaring candidates and the rules that select among them, built and
tested on their own. This section is the consumer: `infer` types a call
to a declared builtin or library sub by applying those rules at the call
site -- the arguments' inferred types, the call's arity and its context
-- and takes the selected candidate's return type. When the call site
cannot decide between candidates it joins them; when the declarations
rule the call out, it fails as "Multi declarations" says. It waits on
`infer` running on `internal/parse` (milestone m2-lower-the-tree), which
is why it is a section of its own.

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
   `Array[T]`. Container types such as `List[Str]` already parse in `.pmt` files
   (see "A slurpy takes no bare element type"); what they mean waits on
   the paper.
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
8. Lexical rather than file-wide scope for declared syntax and a
   library's operators.
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
11. How to spell a glob slot (`\*` in a prototype) in a declaration.
    Perl's refaliasing has no glob form, so the backslash rule does not
    carry over; `lock`, `undef`, `tie`, `tied`, `untie` and `pos` stay
    untyped until this is settled.

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
