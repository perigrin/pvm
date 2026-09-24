"""Port tier 11_oo to the mdtest topic format.

Five topics: the adjacency body, classic bless-and-dispatch, the
`class` feature, the tie protocol, and what the gated keywords mean
with their features off.
"""
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check  # noqa: E402

TIER = '11_oo'
covered = set()

covered |= topic(TIER, 'adjacency-11_oo.md', 'Every construct, each beside another', """
One body holding every construct this tier introduces, each adjacent to
another -- and the one file in the corpus written because the bug it
catches was already known.

**Tier 11 oo.** Introduces nothing of its own; the mixture is the
subject. Depends on 08_references, and `bless` is the whole of the
dependency: the blessed hash below is a tier 08 reference plus a
string. Pairing with tier 10 instead would assert nothing -- this tier
uses no file handle.

MEASURED against our parser, and this is the case the whole
adjacency-file design was written for:

    class Foo { ADJUST { 1 } }                   parses, 0 Unknowns
    class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both

`ADJUST` alone parses -- that is the ADJUST case in `class.md`, and it
is green. `ADJUST` followed by anything -- `method`, `field`, or a
second `ADJUST` -- does not. A corpus of one construct per file goes
green over this by CONSTRUCTION, because every construct in such a
corpus is measured alone and every construct alone passes.
""", [
    ('00_adjacency.t', 'The whole tier in one body', """
A `class` with a `field`, an `ADJUST` and a `method` in one body; a
classic `bless` of an empty anon hash; a compile-time method call and a
dynamic one on the result; an infix `isa` and a `can` chain; and a
`tie` with the `tied` that asks about it. The class side and the
classic side are adjacent to each other as well, because a parser that
switches modes on `class` has to switch back.

The refusal names WHICH SITE declines and the issue names which bug
somebody believed this was. They are different promises and the case
makes both. Measured, the Unknown starts at the `ADJUST` keyword and
runs to the end of the `method` after it -- the expression parser reads
the ADJUST block, then finds the `method` before the terminator, which
is `trailing_tokens`. Without the site, a refusal that drifted
elsewhere would leave this reading as though the ADJUST bug were still
what it measured.

`$yes` and `$code` are the argument-extent slice's two constructs
(issue 01a0c730), adjacent to the dispatch they are not. `$b isa Bar`
is an INFIX OPERATOR: measured, it emits `<2> isa` with no `entersub`
and no `method_named` at all, so it sits here beside four method calls
that emit both, and a parser that read it as a fifth is caught by the
mixture and by nothing else. `$b->can("hi")->($b)` is TWO calls through
one chain and only the first is a method call; measured, two `entersub`
under one `method_named`, standing next to the plain `$b->hi` and the
dynamic `$b->$name` it must not be confused with.

`tie` and `tied` are here for a reason the other constructs are not:
they need `Bar` to grow a second ROLE. The same package is a plain
blessed class -- `bless {}, "Bar"`, reached by `$b->hi` -- and a tie
implementation, because `TIESCALAR` and `FETCH` live in it. A parser
that treats a package as having one kind sees a body in which the same
name is both, which no one-construct file can present. `tie my $t,
"Bar", "arg"` declares `$t` IN the argument list, standing next to `my
$b = bless {}, "Bar"`, so the two spellings of "make a variable hold a
Bar" are adjacent and only one of them is an assignment. `tied($t)` is
a NAMED UNARY where `tie` is a list operator -- measured, `<1> tied
sK/1` against `<@> tie vK/2`: one child and no mark against a mark and
two.

`FETCH` returns `fetched` and not `tied`, which reads more naturally
and is the trap: the word would otherwise appear in the output where a
reader would take it for the keyword.

`Foo2` is `ref($c)` then `$c->m`, which ADJUST raised from 1 to 2;
`Barbarbar` is `ref($b)` then the same method reached two ways;
`fetched` is `$t` read through the tie, which is `FETCH` reached with
no call written anywhere, and `y` is `tied($t)` answering. Nothing
prints a raw object: `Bar=HASH(0x...)` carries an address that changes
every run, so `ref` is what this asserts on.
"""),
])

covered |= topic(TIER, 'bless-dispatch.md', 'bless and method dispatch', """
The classic object system: a reference plus a string, and the four ops
perl uses to find a method on the result.

**Tier 11 oo.** Introduces `anonhash`, `bless`, `emptyavhv`, `isa`,
`method`, `method_named`, `method_super`, `methstart`, `shift`, `stub`,
`tie`, `tied`. Depends on 08_references.

What this tier adds is how the callee is found; the `entersub` and
`leavesub` that surround every call below are tier 07's. A bare
`package` statement emits NO runtime op, so these cases can name
packages without borrowing tier 12's `use`.

Cases print `ref $o` rather than the object itself: printing an object
gives `Foo=HASH(0x55d3...)`, whose address changes every run, where
`ref` gives the stable string the blessing installed. `ref` is tier
08's op.
""", [
    ('01_bless_empty.t', 'bless of an EMPTY anonymous hash', """
`bless {}, $c` blesses an empty anonymous hash, and the optimiser has a
dedicated op for that case: `emptyavhv`, not `anonhash`. This is the
case that looks simplest and carries the op nobody expects.

Measured, the hash never becomes an `anonhash`:

    6  <0> emptyavhv[t3] s/ANONHASH
    7  <0> padsv[$c:1,3] s
    8  <@> bless sK/2

The populated case below is the same construct with contents in the
braces and emits three different ops. A parser that learns one form
learns nothing about the other.
"""),
    ('02_bless_populated.t', 'bless of a POPULATED anonymous hash', """
`bless { %a }, $c` builds a real anonymous hash first. Measured, with
no `emptyavhv` anywhere:

    d  <0> pushmark s
    e  <0> padhv[%a:1,4] l
    f  <@> anonhash sK*/1
    g  <0> padsv[$c:2,4] s
    h  <@> bless sK/2

The pair with the empty case is the point: one construct in the source
-- an anonymous hash handed to `bless` -- compiles to two unrelated op
streams depending on whether the braces are empty, so the corpus has to
carry both or it measures half of `bless`. `%a` is a tier 02 hash and
`aassign`, `padhv` are tier 02's ops; what this case introduces is
`anonhash` and the `bless` around it.
"""),
    ('03_method_named.t', 'A method name written out', """
`$o->hi`, with the method name written out, resolves the name at
compile time:

    b  <.> method_named[PV "hi"] l
    c  <1> entersub[t3] lKRS/TARG

This is the spelling everyone measures, and measuring only it would
miss the dynamic and SUPER cases, which are different ops for what a
reader calls the same construct.

THE TOKEN FACT, and why the direct spelling needs one as much as the
indirect one does. The indirect case below declares that `new Foo`
holds NO arrow, which is the only place indirect object notation is
visible at all -- both spellings emit the same four ops. But a negative
alone is satisfied VACUOUSLY by a lexer that never emits `->`: one that
folded the arrow into the word beside it, or dropped it as trivia,
passes the indirect claim perfectly while getting every direct call in
this tier wrong. This is the matching positive, and the pair is what
makes either falsifiable. ONE arrow, not "at least one", which is why
this case carries the claim rather than the adjacency body: that body
makes four arrow calls and could only say something vaguer.
"""),
    ('04_method_dynamic.t', 'A method name in a scalar', """
`$o->$m`, with the name in a variable, is a DIFFERENT op: `method`,
resolved at run time, not `method_named`. The source differs from the
named case by one sigil and the op changes entirely. `method_named`
carries the name as a constant in the op itself; `method` takes it off
the stack, so the `padsv` that supplies it is part of the call rather
than an argument to it:

    d  <0> padsv[$o:3,5] sM
    e  <0> padsv[$m:4,5] s
    f  <.> method lK/1
    g  <1> entersub[t4] lKRS/TARG
"""),
    ('05_method_super.t', 'SUPER:: dispatch', """
`$o->SUPER::hi()` is a third dispatch op again: `method_super`, which
starts its search in the CURRENT package's `@ISA` rather than in the
invocant's class:

    i  <.> method_super[PV "hi"] l
    j  <1> entersub[t7] lKRS/TARG

`SUPER::` resolves against the package the call is COMPILED in, not the
one the object is blessed into, which is why this case makes the call
at file scope inside `package Derived;` rather than from inside a
method. Written the usual way -- `sub hi { $_[0]->SUPER::hi() }` -- the
op would sit in the sub's own optree, and `perl -MO=Concise,-exec` with
no sub named dumps the main program alone, so nothing would measure it.
The case is shaped by what can be observed, and says so. `our @ISA =
("Base")` is tier 02's `gv`, `rv2av` and `aassign`.
"""),
    ('06_indirect_new.t', 'Indirect object notation: `new Foo`', """
`new Foo` and `Foo->new` compile to the SAME op stream. Indirect object
notation is a lexing problem, not a compilation one. Both spellings
emit:

    3  <0> pushmark s
    4  <$> const[PV "Foo"] sM/BARE
    5  <.> method_named[PV "new"] s
    6  <1> entersub[t2] sKRS/TARG

Nothing downstream of the lexer can tell them apart, so a corpus that
only checks behaviour or the optree cannot see the difference at all --
the same shape as the `5e-1` case in tier 01, and the reason the format
carries token facts. This is the tier's hard marker, `indirect-new`.
"""),
    ('10_isa_infix.t', '`isa` is an infix operator, not dispatch', """
`$o isa Foo` emits a binary `isa` op with no call machinery at all --
no `entersub`, no `method_named` -- where `$o->isa("Foo")` emits both:

    my $infix  = $o isa Foo;      a  <2> isa sK/2
    my $method = $o->isa("Foo");  h  <.> method_named[PV "isa"] s
                                  i  <1> entersub[t5] sKRS/TARG,STRICT

One word, two unrelated parses, and only one of them is this tier's
dispatch machinery. 29 of T1's 986 files use the word and the corpus
named neither spelling; a tier that named only the method spelling
would have named the half that is NOT distinctive.

IT IS FEATURE-GATED, which is the second half of its parse and the
reason this case opens with `use v5.36`. Measured, without the feature
the infix spelling is not a weaker parse but a SYNTAX ERROR, so a lexer
cannot decide what `isa` is from the token alone. Measured under 5.42.0
it is no longer experimental: `use v5.36` with `$o isa Foo` emits no
warning, where `class` still needs a `no warnings` line.

THE FALSE CASE IS WHAT MAKES THE OUTPUT FALSIFYING. `$yes` is 1 and
`$no` is the EMPTY STRING -- perl's false, which interpolates to
nothing -- so the two run together as `1` and the `|` is what proves a
second value was printed at all. A parser that read `isa` as always
true prints `11|`; one that read it as always false prints `|`. Both
wrong readings are one byte from the right one.

`bless {}, $ENV{X} // "Foo"` rather than `Foo->new`, for two reasons:
the `//` keeps the class name a RUNTIME value so nothing folds, and
`Foo->new` would put an ARROW in a source whose whole point is that it
has none. What the facts cannot say is how many `isa`s there are --
there are TWO, and the grammar admits only `one` and `no`.
"""),
    ('11_can_chain.t', 'The `->can(...)->()` chain is two calls', """
`$o->can("hi")->($o)` is TWO CALLS THROUGH ONE CHAIN, and only the
first is a method call; the second arrow dereferences a code ref.
Measured:

    b  <0> pushmark s
    c  <0> padsv[$o:3,4] sM      <- the ARGUMENT of the second call
    d  <0> pushmark s
    e  <0> padsv[$o:3,4] sM      <- the INVOCANT of the first
    f  <$> const[PV "hi"] sM
    g  <.> method_named[PV "can"] s
    h  <1> entersub[t4] sKRS/TARG    <- the method call
    i  <1> entersub[t5] lKS/TARG     <- the CODE DEREF

TWO `entersub`, ONE `method_named`. That asymmetry is the whole
construct. A parser that read BOTH arrows as method calls emits two
`method_named`; one that read the second arrow as part of the first
call emits one `entersub`. Both are ordinary-looking op streams and
both are wrong. The flags say the same thing again -- `h` is `sKRS`, a
resolved method call, and `i` is `lKS` with no `R`. 132 of T1's 986
files -- 13% -- use `can`, and the construct introduces no op of its
own, which is why it is worth a case rather than a list entry.

THE INVOCANT IS PASSED EXPLICITLY, and that is not stylistic. `can`
returns a bare CODE ref and calling a code ref passes no receiver, so
`$o->can("hi")->()` calls `hi` with an EMPTY `@_`. That is the semantic
content of "only the first is a method call".

The body returns `"called"` and not `"hi"`, which is what makes the
count of ONE string true: written the obvious way the source holds two
strings spelled `"hi"` and the fact reads `got 2, want 1`. Found by the
runner, and recorded because the obvious spelling is the broken one.
What the facts cannot say is that there are exactly TWO arrows, the
construct's defining count; the grammar admits only `one` and `no`.
"""),
])

covered |= topic(TIER, 'class-feature.md', 'The `class` feature', """
`class`, `field`, `method` and `ADJUST`: four keywords that a reader
would call the subject of this tier and that produce, between them,
tier 05's `enterloop` and `leaveloop` plus `stub` and `methstart`.

**Tier 11 oo.** Introduces `anonhash`, `bless`, `emptyavhv`, `isa`,
`method`, `method_named`, `method_super`, `methstart`, `shift`, `stub`,
`tie`, `tied`. Depends on 08_references.

This is the sharpest case in the corpus of a construct the optimiser
erases. Measured, neither `use feature 'class'` nor `no warnings
'experimental::class'` emits a runtime op, so the pragmas the syntax
requires cost this tier nothing and do not borrow tier 12's `use`. `use
v5.42;` is not enough on its own: it does not enable `feature 'class'`
in 5.42.0 and `class Foo` under it is a syntax error, so each case
names the feature. The constructor perl generates is XS code with no
Perl optree, so nothing this tier introduces can be measured through
it; `ref` on the result is what shows the class exists.
""", [
    ('07_class_empty.t', 'An empty class body', """
`class Empty {}` is a complete, working class -- `Empty->new` returns
an object -- and it compiles to a bare block containing `stub`. There
is no `class` op:

    1  <0> enter v
    2  <;> nextstate(main 3 d.pl:3) v:%,{,fea=15
    3  <{> enterloop(next->5 last->5 redo->4) v
    4  <0> stub v
    5  <2> leaveloop vK/2
"""),
    ('08_class_field_method.t', '`field` and `method` in a class body', """
`field` and `method` emit no op of their own. A class body holding both
compiles to a bare block of `nextstate`s and nothing else:

    3  <{> enterloop(next->5 last->5 redo->4) v
    4  <;> nextstate(Foo 7 a.pl:4) v:%,us,*,&,{,$,fea=15,0x10
    5  <2> leaveloop vK/2

`field $x = 5;` contributes one `nextstate` and NO store: the
initialiser is compiled into a field-init tree that is not reachable
from the main optree. The method body is likewise its own CV, so the
`methstart` that opens it is not in the main stream -- naming it takes
`-exec,Foo::m`:

    Foo::m:
    1  <+> methstart() v
    2  <;> nextstate(Foo 10 a.pl:4) v:%,us,*,&,$,fea=15,0x10
    3  <0> padsv[$x:FAKE:] s
    4  <1> leavesub[1 ref] K/REFC,1

That is the finding, not an omission: the op stream of a file using
`class` says almost nothing about the class. `methstart` has no classic
counterpart -- a `sub` doing the same job opens with `shift`. Two
spellings of one object system, two disjoint prefixes.
"""),
    ('09_class_adjust.t', '`ADJUST` alone in a class body', """
`ADJUST` runs after construction and emits no op of its own. It becomes
an anonymous sub in the class stash with no CODE slot, so B::Concise
cannot name it and its body is unreachable from any dump; what the main
optree shows is a `nextstate` where the block was, and the whole class
body is three ops:

    3  <{> enterloop(next->5 last->5 redo->4) v
    4  <;> nextstate(Foo 5 f.pl:4) v:%,{,fea=15
    5  <2> leaveloop vK/2

ADJUST sits here BY ITSELF on purpose, and it parses under our parser.
The adjacency body is the same construct with a `method` after it, and
the pair is the whole demonstration: one construct per file goes green
over an adjacency bug by construction, because every construct in such
a corpus is measured alone.
"""),
])

covered |= topic(TIER, 'tie.md', 'The tie protocol', """
`tie` binds a variable to a class and `tied` asks whether it is bound.
Two words one lexer would call one family, and two different argument
grammars.

**Tier 11 oo.** Introduces `anonhash`, `bless`, `emptyavhv`, `isa`,
`method`, `method_named`, `method_super`, `methstart`, `shift`, `stub`,
`tie`, `tied`. Depends on 08_references.

Both words were UNCLAIMED by every tier of this corpus, and the
placement was measured rather than argued. `07_subroutines` was the
first proposal, on the ground that `TIESCALAR` and `FETCH` are ordinary
named subs -- true, and not sufficient. `tie` REQUIRES the constructor
to return a BLESSED object, and a constructor that does not fails
SILENTLY:

    $ perl -e 'package C; sub TIESCALAR { my $s = "x"; return \\$s }
        sub FETCH { return 42 }
        package main; tie my $c,"C"; print "[", $c, "]\\n";'
    []

No error, no warning, an empty value. So the blessed constructor is not
optional, and the minimal one emits `emptyavhv` and `bless` -- both of
them this tier's, and both counted against these files because `opsOf`
reads inside a file's own CVs. Tier 11 is the EARLIEST tier that can
hold either word.

Both cases spell the class `$ENV{X} // "Counter"` to keep it a RUNTIME
value so the `tie` cannot fold; `X` is unset when the runner executes
the file. `bless {}, "Counter"` inside the constructor rather than
`$_[0]`: the class arrives as a string these cases already name, and
`$_[0]` would add tier 02's `aelemfast` for nothing the construct is
about.

THE SUBSTRING WORRY IS NOT REAL, and it cost a draft. `tie` is a prefix
of `tied`, so the obvious fear is that a source containing both cannot
count either. Measured against the checker: `checkTokenFact`
(`internal/conformance/fact.go:50`) compares `tokenText == text` on the
token's FULL text and never as a substring, so `one word whose text is
"tie"` counts exactly one even where `tied` also appears.
""", [
    ('12_tie_variable.t', '`tie` is a list operator taking a VARIABLE', """
`tie VARIABLE, CLASS, LIST` takes a VARIABLE where every other call in
this tier takes an expression, and hands the trailing LIST to a
constructor it names by string. Three things about the argument list
are true of no other construct here: the first argument is a VARIABLE
and `tie my $x, ...` DECLARES it in the same breath; the second is a
CLASS NAMED BY STRING, resolved at runtime; and the trailing LIST is
passed to the constructor, so the extent of the argument list decides
what the class receives. Measured, main program only:

    b  <0> pushmark s
    c  <0> padsv[$x:41,42] sRM/LVINTRO   <- the VARIABLE, declared here
    d  <+> multideref($ENV{"X"}) sK
    e  <|> dor(other->f) sK/1
    f      <$> const[PV "Counter"] s     <- the CLASS, a runtime string
    g  <$> const[PV "a"] s               <- the LIST, two more elements
    h  <$> const[PV "b"] s
    i  <@> tie vK/3

`vK/3` is the arity: three children under the mark, which is variable,
class and the two-element list flattened. A parser that stopped the
argument extent at the class name emits `vK/2` and the program still
runs, printing the same thing, which is why the arity is recorded here
and the output cannot carry this half of the claim. A parser that read
`tie` as an ordinary named operator gets the second and third right and
the first wrong: `my $x` in argument position is a declaration, and
nothing in the call site says so.

THE OUTPUT IS THE DISPATCH. `42` is `FETCH`'s return value reached by
reading `$x`, and the read looks like an ordinary scalar read in the
source -- `print $x` -- which is the whole point: the call is INVISIBLE
at the call site. A parser that compiled `print $x` as a plain pad read
prints the empty string, because an untied `my $x` is undef.

The token fact pins `tie` as ONE word, the count a lexer folding the
keyword into the declaration beside it -- `tie my` read as one thing --
gets wrong. A negative `no word whose text is "tied"` was drafted and
WITHDRAWN: with full-text comparison there was never a way for `tied`
to appear in a source that does not write it, so the fact could not
fail and asserted nothing.
"""),
    ('13_tied_boolean.t', '`tied` is a named unary, in boolean position', """
`tied` asks a variable whether it is tied and answers with the OBJECT
the tie is bound to, or with undef. Measured, main program only:

    m  <0> padsv[$x:41,43] sRM
    n  <1> tied sK/1            <- arity ONE, a named unary
    o  <|> cond_expr(other->p) lK/1
    q  <0> padsv[$plain:42,43] sRM
    r  <1> tied sK/1
    s  <|> cond_expr(other->t) lK/1

`<1>` is the arity and it is the parse: one child, no pushmark, which
is a named unary and not a list operator, where the `tie` two lines
earlier emits `<@> tie vK/2` -- a list op with a mark. Two words a
lexer would call one family, two different argument grammars, and the
op stream says so.

NO `ref` AND NO `defined`, which an earlier draft assumed were both
required. Measured, `tied $x` in the condition of a ternary is enough:
an untied variable gives undef, which is false, and a tied one gives a
blessed reference, which is true. `ref(tied $x)` would add tier 08's
`ref` and `defined(tied $y)` tier 04's `defined`; both are earlier
tiers and would be legal, and neither is needed.

BOTH CASES ARE WHAT MAKE THE OUTPUT FALSIFYING. `$x` is tied and
`$plain` is not, so the line carries a true answer and a false one. A
parser that read `tied` as always true prints `tt`; one that read it as
always false prints `uu`. `$plain = 1` rather than an undef variable,
because the question is whether the VARIABLE is tied and not whether
its value is defined: an untied undef would let a parser that compiled
`tied` as `defined` pass.

The fact asserted is `tie` and not `tied`, and that is the honest half:
there are TWO `tied`s here -- the true case and the false one -- and
the grammar admits only `one` and `no`, so `one word whose text is
"tied"` would be FALSE of a correct lex. `tie` IS one, and pinning it
says the binding and the question reached the token stream as different
words.
"""),
])

covered |= topic(TIER, 'ungated.md', 'What the keywords mean with the feature off', """
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
""", [
    ('14_state_ungated.t', '`state` without its feature is a method call', """
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
"""),
    ('15_class_ungated.t', '`field` without the class feature is a method call', """
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
"""),
    ('16_classname_ungated.t', '`__CLASS__` without the feature is a filehandle', """
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
"""),
    ('17_defer_ungated.t', '`defer` without its feature runs the block FIRST', """
THE WORST OF THE SIX, and the reason is temporal rather than
structural. The corpus's other version-gate findings change what a
construct IS: `say` becomes a method call, `isa` becomes a filehandle
print, `state` becomes a method call on undef, and each fails loudly or
at runtime. This one changes WHEN the code runs. Measured:

    $ perl -e 'sub f { defer { print "D\\n" } print "body\\n" } f();'
    D
    body
    Can't locate object method "defer" via package "1"

    $ perl -e 'use feature "defer"; sub f { defer { print "D\\n" }
               print "body\\n" } f();'
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

    $ perl -MO=Deparse -e 'defer { print "D\\n" } print "body\\n";'
    defer {
        print "D\\n"
    } print("body\\n");
"""),
])

check(TIER, covered)
