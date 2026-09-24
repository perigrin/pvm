# Constructs that run while the file is parsed

`BEGIN`, `END`, `__PACKAGE__`, `__LINE__`, `__FILE__`, the `#line`
directive and `use constant`: the constructs where parsing and execution
INTERLEAVE, where something earlier in the file changes what the rest of
the parse sees or means.

**Tier 12 packages.** Introduces `require`. Depends on 07_subroutines.

`BEGIN` appears in 99 of T1's 986 files and in no corpus file before
these cases were written. Nor did `END`, `__PACKAGE__`, `__FILE__`,
`__LINE__` or `use constant` -- the corpus made no claim about any of
it.

NONE OF THESE EMIT AN OP OF ITS OWN. The phasers are compile-time
instructions and the optree records only what ends up running;
`__PACKAGE__` is a `const` by the time an optree exists; `use constant`
installs a sub which the optimiser then inlines. So the op lint cannot
see any of this area, and what each case claims instead is ORDERING, a
VALUE, or a TOKEN.

## BEGIN runs first, END runs last, whatever the source order

The source order is END, BEGIN, PRINT, and the output is the reverse of
neither. That is the whole case: a parser that treated the two phasers
as ordinary blocks would run them where they appear and print `end begin
main`, and one that ignored phasing entirely would print them in source
order too. Only a parser that knows `BEGIN` is hoisted to compile time
and `END` deferred to exit produces this sequence.

MEASURED perl 5.42.0, the ops say nothing:

    $ perl -MO=Concise,-exec -e 'BEGIN { print "B\n" } print "main\n"'
    ... const ... print ... const ... print ...

Two prints and two constants, exactly as two ordinary statements give.

The token facts count the two phaser names. They are barewords followed
by a block with no semicolon and no parens -- structurally identical to
a sub call with a hashref argument, which is what a lexer that did not
know them would produce.

```perl
END { print "end\n" }
BEGIN { print "begin\n" }
print "main\n";
```

```behavior
parses: yes
```

```output
begin
main
end
```

```tokens
one word whose text is "BEGIN"
one word whose text is "END"
```

## `__PACKAGE__` is a bareword that is not a bareword

The compiler replaces it with a string, and the string depends on where
it appears. Two occurrences of identical bytes, two different values,
decided by a `package` statement earlier in the file:

    $ perl -e 'package Foo; print __PACKAGE__, "\n";
               package main; print __PACKAGE__, "\n";'
    Foo
    main

That is the same shape as this corpus's version-gate findings -- a
declaration above changing what bytes below mean -- with the difference
that here BOTH readings are correct and neither fails.

MEASURED perl 5.42.0, the op is `const[PV "main"]` carrying
`s/TOKEN=PACKAGE`. The VALUE is a plain string by the time the optree
exists, so the op stream cannot tell this from `print "main"` by what it
computes, though the flag does record that a token was replaced. `const`
is tier 01's op either way. So the op lint cannot see this construct,
output can see its VALUE, and only a token fact can see that the source
said `__PACKAGE__` rather than the answer.

`__FILE__` is deliberately absent. Measured, it reports the path the
runner executed, which is a temporary file whose name changes every run,
so no output pin could survive. The `#line` case reaches it the only way
a corpus case can: by overriding it.

```perl
package Foo;
print __PACKAGE__, "\n";
package main;
print __PACKAGE__, "\n";
```

```behavior
parses: yes
```

```output
Foo
main
```

```tokens
one word whose text is "Foo"
```

## A `#` at column zero that is not a comment

The `#line` directive rewrites `__LINE__` and `__FILE__` for everything
after it. perlsyn gives the exact pattern in its "Plain Old Comments
(Not!)" section; the leading `#` must be at column zero, which is why
this case's directive is unindented while every other line could be.

MEASURED perl 5.42.0:

    $ cat lp.pl
    #line 200 "bzzzt"
    print __LINE__, " ", __FILE__, "\n";
    $ perl lp.pl
    200 bzzzt

    $ perl -e 'print __LINE__, " ", __FILE__, "\n"'   # no directive
    1 -e

The second line reports itself as line 200 of a file called `bzzzt`,
neither of which is true of the bytes on disk. A lexer that treats every
`#` as a comment skips the directive and reports `2` and the real path
-- silently, for every error location in the rest of the program.

AND PARSING IS NOT IMPLEMENTING. This case passes and the directive is
still UNIMPLEMENTED here: `scanComment` consumes `#line 200 "bzzzt"`
like any other `#`, so the lexer skips it as trivia and our `__LINE__`
and `__FILE__` would report the real position. That is precisely the
silent failure the paragraph above warns about, and it is ours. The case
cannot catch it: `output` is what PERL prints, and `parses: yes` is
satisfied by treating the directive as a comment, which is what makes it
parse. What would catch it is a token fact naming a `line directive`
category, and the glossary has none -- adding one is `01a0d0b0`.

THE DIRECTIVE IS ALSO WHAT MAKES `__FILE__` PINNABLE AT ALL, since an
ordinary `__FILE__` reports the runner's temporary path. Overriding it
supplies a name that does not depend on where the runner put the file.

This is a LEXER claim in a topic of compile-time constructs, and it
belongs with them because it is the same kind of thing: a line that
changes how the rest of the file is read.

```perl
#line 200 "bzzzt"
print __LINE__, " ", __FILE__, "\n";
```

```behavior
parses: yes
```

```output
200 bzzzt
```

```tokens
one word whose text is "__LINE__"
one word whose text is "__FILE__"
```

## `use constant` installs a bareword that parses as a call

A construct that changes the grammar for everything below it. `PI` below
the pragma is a CALL, not a string -- verified rather than assumed:

    $ perl -e 'use constant PI => 3; print defined(&PI) ? "yes" : "no"'
    yes

The same bareword above the pragma would be a string under `no strict`
and a syntax error under `use strict`. That is what makes this a parsing
claim rather than a fact about constants.

MEASURED perl 5.42.0, the op stream cannot see it:

    $ perl -MO=Concise,-exec -e 'use constant PI => 3; print PI, "\n"'
    ... const[IV 3] ... print ...

A `const`, exactly as `print 3` gives. The pragma runs at compile time,
installs a sub, and the optimiser inlines the call, so by the time the
optree exists there is no trace of either the pragma or the call.

`PI + 1` IS THE DISCRIMINATOR, and measured it is sharper than a first
draft claimed. That draft said a bareword reading would numify to 0 and
print 1. It does not:

    $ perl -e 'print PI + 1, "\n"'
    (no output, exit 0)
    $ perl -MO=Deparse -e 'print PI + 1, "\n"'
    print PI 1, "\n";

Without the pragma the bareword lands in print's FILEHANDLE slot -- the
same indirect-object shape `11_oo/06_indirect_new.t` measures -- so the
program prints nothing at all. The readings are `4` against NOTHING, not
4 against 1. The output pin discriminates either way, but the mechanism
is the filehandle slot rather than numification.

The token fact counts `constant`. The count of `PI` itself would be two
-- once in the pragma, once in the call -- which is a claim about the
case's shape rather than its construct. The fat comma is incidental and
already claimed by tier 02, but worth noting: `PI => 3` autoquotes the
left side, so the pragma receives the string `"PI"` and installs a sub
by that name.

```perl
use constant PI => 3;
print PI + 1, "\n";
```

```behavior
parses: yes
```

```output
4
```

```tokens
one word whose text is "constant"
```
