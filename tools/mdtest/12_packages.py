"""Build tier 12's topic files from corpus.json.

Four topics: the package statement in both spellings, the loading
constructs (`use`, `require`, `import`), the compile-phase constructs
that change what the rest of the parse sees, and the adjacency body.
"""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check

TIER = '12_packages'

covered = set()

covered |= topic(TIER, 'packages.md', 'The package statement', """
`package NAME;` and `package NAME { ... }` -- the two spellings of the
construct that moves the compiler's notion of the current package.

**Tier 12 packages.** Introduces `require`. Depends on 07_subroutines.

NEITHER SPELLING EMITS AN OP. `package Foo;` contributes no runtime op
at all; its only trace is the package name inside the `nextstate` that
follows, which is a field of an op tier 01 already claims. The block
form emits `enterloop`, `stub` and `leaveloop` -- a bare block, which is
tier 05's construct, plus tier 11's placeholder. Nothing in either
stream says `package`.

So both cases pin the construct by a FULLY QUALIFIED CALL instead.
Delete the `package` line and the sub lands in `main::`, the call is to
an undefined subroutine, and perl dies. That is what a tier of no-op
constructs has in place of an op check: every pinned line is falsifiable
by deleting a statement.
""", [
    ('01_package_statement.t', '`package Foo;` switches package for the rest of the file', """
`package Greet;` switches the current package for the rest of the
enclosing block or file. The whole effect is on the compiler, so the
call `Greet::hello()` is the only evidence the statement took effect --
a parser that dropped it would put `hello` in `main::` and this program
would die on an undefined subroutine.

MEASURED perl 5.42.0:

    $ perl -e 'package Greet; sub hello { "hello" } package main; print Greet::hello(), "\\n"'
    hello
"""),
    ('02_package_block.t', '`package Foo { ... }` scopes the switch to a block', """
The same construct as the statement form with a scope attached, and the
spelling `class Foo { ... }` also uses -- which is why tier 11 could not
be ordered after this tier. Its three ops belong to tiers 05 and 11;
nothing in the stream says `package`.

MEASURED perl 5.42.0:

    $ perl -e 'package Louder { sub shout { "HELLO" } } print Louder::shout(), "\\n"'
    HELLO
"""),
])

covered |= topic(TIER, 'loading.md', 'use, require and import', """
`use Foo LIST` locates a file, loads it, and then calls
`Foo->import(LIST)`. `require` is the same construct with the import
step removed, and `import` is an ordinary method call. Those three
statements are this whole area.

**Tier 12 packages.** Introduces `require`. Depends on 07_subroutines.

`use` EMITS NOTHING -- it is a `BEGIN` block that has finished before
there is an optree to dump. `require` is the one construct here that
reaches run time and the one op this tier introduces: `require strict;`
emits `const[PV "strict.pm"] s/BARE` and then `require`, `require $m`
emits `padsv` and the same `require`. One op, several sources. `import`
emits `pushmark`, two `const`, `method_named` and `entersub` -- tiers
01, 07 and 11 between them, with nothing left over. There is no import
op; the reason `use` is heavier than `require` is a call the op stream
cannot distinguish from any other call.

Because both halves of `use` are compile-time, the cases observe them
through `%INC` and the symbol table rather than through anything the
loaded module emits. The `04`/`05` pair is where the tier's claim
actually lives: neither case alone separates loading from importing.
""", [
    ('03_use_pragma.t', '`use strict;` lexes, and that is all it establishes', """
A pragma is a `BEGIN` block, finished before there is a runtime optree.
The op stream for this case is the stream for `my $x = "ok"; print
"$x\\n"` alone, differing only in the feature bits printed on
`nextstate`, which are a field rather than an op.

WHAT THIS CASE DOES NOT REACH, stated plainly because the pin does not
say it. Measured: delete both `use` lines and this program still prints
`ok`. So the pin establishes that the lines LEX and the statement is
delimited, and it establishes nothing about the load or the import. A
pragma cannot do better -- its effect is lexical and compile-time, and
`$^H` and `${^WARNING_BITS}` consulted at run time report the CALLER's
scope, measured OFF inside the very file that turned them on.

It is kept because `use` with a pragma is the overwhelmingly common
spelling and a lexer that mis-delimited it would fail here. The case
that reaches the rest is the next one.
"""),
    ('04_use_import.t', '`use POSIX;` loads the file AND calls import', """
This is the case that makes `use` FALSIFIABLE, through a module with an
exporter and the symbol table. `use POSIX;` puts `floor` into `main::`,
and deleting the `use` line changes both pinned lines -- so the output
measures the statement rather than its absence of syntax errors.

POSIX is core, its import list is not versioned in any way this case
observes, and nothing here depends on what it prints, because it prints
nothing.

MEASURED perl 5.42.0:

    $ perl -e 'use POSIX; print +($INC{"POSIX.pm"} ? "yes" : "no"), ($main::{"floor"} ? "yes" : "no"), "\\n"'
    yesyes
"""),
    ('05_use_empty_list.t', '`use POSIX ();` loads the file and suppresses import', """
The empty parentheses are not an empty argument list but the absence of
one. Same module and the same two observations as the case above, one
statement apart:

    use POSIX;      loaded yes    imported yes
    use POSIX ();   loaded yes    imported no

The first column is `require`'s half of `use`, the second is `import`'s.
The two cases' op streams are byte-identical to each other and to the
two `print` statements alone, which is why the symbol table is where the
difference is visible at all. That is the tier's defining difficulty
stated as a measurement: `use` is real, and no optree records it.

A pragma could not stand in here, because a pragma's whole effect IS its
import and there would be nothing left to observe.

MEASURED perl 5.42.0:

    $ perl -e 'use POSIX (); print +($INC{"POSIX.pm"} ? "yes" : "no"), ($main::{"floor"} ? "yes" : "no"), "\\n"'
    yesno
"""),
    ('06_require_bareword.t', '`require strict;` -- the tier\'s one runtime op', """
The bareword-to-filename rewrite is the compiler's: by the time the op
runs, `strict` has become the string `"strict.pm"`, emitted as
`const[PV "strict.pm"] s/BARE` and consumed by `require`. So `require`
takes a filename, never a package name, and the bareword spelling is
sugar the op stream cannot see.

`strict` is chosen because it is core, prints nothing, and its version
does not reach the output. The load is observed through `%INC` rather
than through anything the module itself emits.

MEASURED perl 5.42.0:

    $ perl -e 'require strict; print +($INC{"strict.pm"} ? "yes" : "no"), "\\n"'
    yes
"""),
    ('07_require_expression.t', '`require $m;` takes the filename from a variable', """
This is the case that shows the bareword form was sugar. `require
strict` emits `const[PV "strict.pm"] s/BARE` then `require`; this emits
`padsv` then `require`. One op, two sources, and only the operand
differs -- which is also why `require Foo::Bar` and `require
"Foo/Bar.pm"` are the same statement and `require $m` with `$m` holding
`"Foo::Bar"` is not.

MEASURED perl 5.42.0:

    $ perl -e 'my $m = "warnings.pm"; require $m; print +($INC{$m} ? "yes" : "no"), "\\n"'
    yes
"""),
    ('08_import_call.t', '`import` is an ordinary method call', """
`Marker->import("tag")` emits `pushmark`, `const[PV "Marker"] sM/BARE`,
`const[PV "tag"] sM`, `method_named[PV "import"]`, `entersub` -- tiers
01, 07 and 11 between them and nothing left over. The claim is that
`import` is not special, and the way to demonstrate it is to call it
directly and get the same stream `use` would have produced.

Called at run time rather than from `BEGIN` so the output order is the
source order. Under `BEGIN` the same call runs during compilation and
prints first, which is true of any BEGIN block and is tier 05's
business, not this tier's.

MEASURED perl 5.42.0:

    $ perl -e 'package Marker; sub import { print "import $_[0] $_[1]\\n" } package main; Marker->import("tag"); print "after\\n"'
    import Marker tag
    after
"""),
])

covered |= topic(TIER, 'compile-phase.md', 'Constructs that run while the file is parsed', """
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
""", [
    ('09_begin_end.t', 'BEGIN runs first, END runs last, whatever the source order', """
The source order is END, BEGIN, PRINT, and the output is the reverse of
neither. That is the whole case: a parser that treated the two phasers
as ordinary blocks would run them where they appear and print `end begin
main`, and one that ignored phasing entirely would print them in source
order too. Only a parser that knows `BEGIN` is hoisted to compile time
and `END` deferred to exit produces this sequence.

MEASURED perl 5.42.0, the ops say nothing:

    $ perl -MO=Concise,-exec -e 'BEGIN { print "B\\n" } print "main\\n"'
    ... const ... print ... const ... print ...

Two prints and two constants, exactly as two ordinary statements give.

The token facts count the two phaser names. They are barewords followed
by a block with no semicolon and no parens -- structurally identical to
a sub call with a hashref argument, which is what a lexer that did not
know them would produce.
"""),
    ('10_compile_tokens.t', '`__PACKAGE__` is a bareword that is not a bareword', """
The compiler replaces it with a string, and the string depends on where
it appears. Two occurrences of identical bytes, two different values,
decided by a `package` statement earlier in the file:

    $ perl -e 'package Foo; print __PACKAGE__, "\\n";
               package main; print __PACKAGE__, "\\n";'
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
"""),
    ('11_line_directive.t', 'A `#` at column zero that is not a comment', """
The `#line` directive rewrites `__LINE__` and `__FILE__` for everything
after it. perlsyn gives the exact pattern in its "Plain Old Comments
(Not!)" section; the leading `#` must be at column zero, which is why
this case's directive is unindented while every other line could be.

MEASURED perl 5.42.0:

    $ cat lp.pl
    #line 200 "bzzzt"
    print __LINE__, " ", __FILE__, "\\n";
    $ perl lp.pl
    200 bzzzt

    $ perl -e 'print __LINE__, " ", __FILE__, "\\n"'   # no directive
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
"""),
    ('12_use_constant.t', '`use constant` installs a bareword that parses as a call', """
A construct that changes the grammar for everything below it. `PI` below
the pragma is a CALL, not a string -- verified rather than assumed:

    $ perl -e 'use constant PI => 3; print defined(&PI) ? "yes" : "no"'
    yes

The same bareword above the pragma would be a string under `no strict`
and a syntax error under `use strict`. That is what makes this a parsing
claim rather than a fact about constants.

MEASURED perl 5.42.0, the op stream cannot see it:

    $ perl -MO=Concise,-exec -e 'use constant PI => 3; print PI, "\\n"'
    ... const[IV 3] ... print ...

A `const`, exactly as `print 3` gives. The pragma runs at compile time,
installs a sub, and the optimiser inlines the call, so by the time the
optree exists there is no trace of either the pragma or the call.

`PI + 1` IS THE DISCRIMINATOR, and measured it is sharper than a first
draft claimed. That draft said a bareword reading would numify to 0 and
print 1. It does not:

    $ perl -e 'print PI + 1, "\\n"'
    (no output, exit 0)
    $ perl -MO=Deparse -e 'print PI + 1, "\\n"'
    print PI 1, "\\n";

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
"""),
])

covered |= topic(TIER, 'adjacency-12_packages.md', 'Every construct, each beside another', """
One body holding every construct this tier covers, each adjacent to
another, and paired with the declared prerequisite.

**Tier 12 packages.** Introduces `require`. Depends on 07_subroutines.

WHY A MIXTURE NEEDS ITS OWN CASE. The tier's other cases are one
construct each, which is what makes them diagnosable: when the bareword
`require` case fails, `require` is the only construct present. That same
property is why a corpus of such cases cannot reach an ADJACENCY bug --
a parser that handles every construct alone and mis-handles a pair goes
green over the pair. Tier 11 has the case the design was written for:
`class Foo { ADJUST { 1 } }` parses and `class Foo { ADJUST { 1 } method
m { 2 } }` does not.
""", [
    ('00_adjacency.t', 'The whole tier in one body', """
The pairs this case puts next to each other:

    an importing `use` immediately followed by an empty-list `use`
    a `package NAME;` immediately followed by a `package NAME { }`
    a bareword `require` immediately followed by an expression `require`
    a `package main;` immediately followed by a statement
    an `import` sub DEFINED in one package and CALLED from another

The last is the pairing with `07_subroutines`, this tier's declared
prerequisite. `import` is the reason `use` is heavier than `require`,
and it is a plain sub call; a case that declared the dependency without
crossing the package boundary would assert nothing about it. Pairing
with tier 11 instead would assert nothing at all -- this tier needs
nothing from it.

`Greet::hello` and `Louder::shout` are called by their fully qualified
names, the third spelling of the package boundary here and the only one
that appears in the op stream: `gv[*Greet::hello]`. The two `package`
statements that put them there leave no trace.

EVERY PINNED LINE IS FALSIFIABLE BY DELETING A STATEMENT, which is what
a tier of no-op constructs has instead of an op check. Measured, each
deletion changes the output:

    drop `package Greet;`    `Greet::hello` is undefined and the program dies
    drop `package Louder {`  likewise for `Louder::shout`
    drop `use POSIX;`        `use:` reads `no` where it read `yes`
    drop either `require`    `inc:` loses a `yes`
    drop `Greet->import`     the first line of output goes

`use Fcntl ();` is the one whose deletion does NOT change the output,
because its whole claim is that it imports nothing: it is pinned by
CONTRAST with the `use POSIX;` above it. `LOCK_EX` is chosen over
`O_RDONLY` for exactly that reason -- POSIX exports `O_RDONLY` too, so
pinning it would have read `yes` whether or not `Fcntl` was ever loaded,
and the pair would have measured one statement twice. Measured on
5.42.0: POSIX exports `O_RDONLY` and does not export `LOCK_EX`.

THE COMPILE-PHASE SLICE WRAPS THE WHOLE BODY, which is the adjacency
claim it makes. `BEGIN` is written AFTER `END` in the source and runs
first; `END` runs after the last statement. Neither emits an op, so the
ordering is the only evidence either exists, and a parser that treated
them as ordinary blocks would print them where they appear.

`pkg main line 300 file adj` is four compile-time constructs in one
statement. The `#line` directive above it is a `#` at column zero that
is NOT a comment: it rewrites `__LINE__` and `__FILE__` for everything
below, which is why the line reports 300 and `adj` rather than its real
position. `__FILE__` is only pinnable at all because of that -- without
the directive it reports the runner's temp path. `TAG` is the bareword
`use constant` installed, and it prints `c` rather than `TAG`, which is
what separates an installed sub from a bareword string.
"""),
])

check(TIER, covered)
