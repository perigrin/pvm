# What the keywords mean with the feature off

Every version-gated construct in this corpus carries its pragma at the
top, and not one pinned what the same bytes mean with the feature OFF
-- which is half of what "version-gated" means, and the half where two
previously-found bugs lived. These four cases pin the other half.

**Tier 11 oo.** Introduces nothing of its own; each case uses `eval`
from 06_control and method dispatch from this tier. Depends on
08_references.

The shape is the same every time: ungated, the keyword is read as a
METHOD NAME or a FILEHANDLE, the bytes compile clean, and the program
means something else. The `eval` traps the death so STDOUT stays
pinnable -- without it a file would print nothing and exit 255, which
says nothing about where it failed.

THE TOKEN LAYER CANNOT SEPARATE THE TWO READINGS, and saying so is the
honest version of a claim these cases got wrong twice. Each positive
fact below is true under BOTH readings -- the keyword lexes as one word
whether it is a keyword or a method name -- so on its own it says
nothing about which reading applies. An earlier draft added `no
operator whose text is "->"` to each, reasoning that a method call is
written with an arrow and a lexer producing one here would have
manufactured it. A LEXER CANNOT MANUFACTURE BYTES THAT ARE NOT THERE:
`scanOperator` matches its table against `l.src` at each position, so a
token whose text is `->` requires those two bytes in the source. The
fact could never fail. The arrow was read off perl's DEPARSE of the
ungated reading and written as if it were a claim about tokens; the
reparse is the PARSER reinterpreting the same tokens, not the lexer
emitting different ones, which is exactly why the two readings are
dangerous. So each positive stays and does the work it can -- the
keyword lexes as ONE word, not split and not swallowed -- and THE
OUTPUT IS THE DISCRIMINATOR, because identical bytes lex identically.

## `state` without its feature is a method call

Without `use feature "state"`, `state $n = 0` is not a syntax error --
it is a METHOD CALL on an undeclared scalar, used as an lvalue.
Measured:

    $ perl -MO=Deparse -e 'sub c { state $n = 0; $n = $n + 1; return $n }'
    -e syntax OK
    sub c { $n->state = 0; $n = $n + 1; return $n; }

`-c` reports SYNTAX OK. The bytes compile clean and mean something
else: `state` is read as a method name, `$n` as the invocant, and the
whole thing as an lvalue. It fails only when the sub is called:

    $ perl -e 'sub c { state $n = 0; $n = $n + 1; return $n } print c()'
    Can't call method "state" on an undefined value

THE FAILURE IS AT RUNTIME, which is what makes this worth a case: a
parser that ignores the pragma produces a program perl accepts, and
nothing at compile time says otherwise. `defined $r` is the
discriminator -- under the feature `c()` returns 1 and this prints
`ran`; ungated it dies inside the eval and prints `died`. The enabled
half is `05_scoping/03_state.t`.

```perl
sub c { state $n = 0; $n = $n + 1; return $n }
my $r = eval { c() };
print defined $r ? "ran" : "died", "\n";
```

```behavior
parses: yes
```

```output
died
```

```tokens
one word whose text is "state"
```

## `field` without the class feature is a method call

Without the `class` feature, `field $x` is a METHOD CALL on `$x` and
`class Foo { }` is an indirect method call whose block is an anonymous
hash. Measured:

    $ perl -MO=Deparse -e 'my $x; field $x;'
    my $x;
    $x->field;

    $ perl -MO=Deparse -e 'class Foo { }'
    'Foo'->class({});

Both compile clean, and the brace group is read as a hashref
constructor -- the same indirect-object shape the `new Foo` case
measures.

THE ASYMMETRY IS THE PART WORTH RECORDING, because it decides what a
case can claim. Measured:

    class Foo { }                        compiles -- indirect method call
    class Foo { field $x; method m {} }  SYNTAX ERROR near "; method "

So the silent reparse applies only to the EMPTY form. A populated class
body is a hard error without the feature, which means the
field-and-method case has no silent-reparse partner to write and the
empty-class case sits exactly in the dangerous window.

This case pins the `field` half, because it is observable without the
indirect-object spelling: the method call dies at runtime on an
undefined invocant, which the eval traps. The `class` half is recorded
above rather than written, since `'Foo'->class({})` would need a
`class` sub in scope to produce output and that sub would then be the
subject rather than the reparse.

```perl
my $x;
my $r = eval { field $x; 1 };
print defined $r ? "ran" : "died", "\n";
```

```behavior
parses: yes
```

```output
died
```

```tokens
one word whose text is "field"
```

## `__CLASS__` without the feature is a filehandle

Without the `class` feature, `__CLASS__` is a FILEHANDLE, and `print
__CLASS__;` prints `$_` to it. Measured:

    $ perl -MO=Deparse -e 'print __CLASS__;'
    print __CLASS__ $_;

    $ perl -MO=Concise -e 'print __CLASS__;'
    ... rv2gv sKR/1 ...
        gv[*__CLASS__] s ...

`rv2gv` over `*__CLASS__` -- a GLOB. Perl reads the bareword as a
filehandle name, `$_` as the thing to print, and the statement as a
print to a handle nobody opened. It compiles clean, prints nothing, and
exits 0.

THAT SILENCE IS THE HAZARD. `state` and `field` die at runtime, loudly
enough that a test notices. This one SUCCEEDS and produces no output,
so a program that meant to print a class name prints nothing and says
nothing about why. The sentinel after it is what makes the silence
observable: the `after` arrives, the class name does not, and a parser
that read `__CLASS__` as a term would have printed something before it.

This case is also where the copied arrow fact was sharpest wrong: the
ungated reading here is not a method call at all, so there was never an
arrow in the deparse to read the claim off. Our parser has since
accepted the file -- `print __CLASS__;` was refused because `__CLASS__`
is all-caps and was taken into `print`'s filehandle slot, the same bug
`10_compile_tokens.t` records reached through a different keyword.

```perl
print __CLASS__;
print "after\n";
```

```behavior
parses: yes
```

```output
after
```

```tokens
one word whose text is "__CLASS__"
```

## `defer` without its feature runs the block FIRST

THE WORST OF THE SIX, and the reason is temporal rather than
structural. The corpus's other version-gate findings change what a
construct IS: `say` becomes a method call, `isa` becomes a filehandle
print, `state` becomes a method call on undef, and each fails loudly or
at runtime. This one changes WHEN the code runs. Measured:

    $ perl -e 'sub f { defer { print "D\n" } print "body\n" } f();'
    D
    body
    Can't locate object method "defer" via package "1"

    $ perl -e 'use feature "defer"; sub f { defer { print "D\n" }
               print "body\n" } f();'
    body
    D

Ungated, `defer { ... }` is an indirect method call whose brace group
is an anonymous hash constructor -- so the BLOCK IS EVALUATED EAGERLY
to build the hash, printing `D`, and only then does perl look for a
`defer` method and fail. Gated, the block runs at scope exit, after
`body`. The two readings produce the same two lines in the OPPOSITE
ORDER, and the failure arrives after the damage. A parser that always
treats `defer BLOCK` as a compound statement is wrong on pre-5.36 code;
one that never does is wrong on modern code, and nothing at compile
time distinguishes them.

What this claims is the ORDER: `D` before `body`, which is the ungated
reading, where the gated one gives `body` before `D`. Perl's deparse of
the ungated reading keeps the block form and shows no arrow at all:

    $ perl -MO=Deparse -e 'defer { print "D\n" } print "body\n";'
    defer {
        print "D\n"
    } print("body\n");

```perl
sub f { defer { print "D\n" } print "body\n" }
eval { f() };
print "end\n";
```

```behavior
parses: yes
refuses: 01a0d087-28dd-711f-a0d5-54cb8515910c
refusal: unimplemented_statement
```

```output
D
body
end
```

```tokens
one word whose text is "defer"
```
