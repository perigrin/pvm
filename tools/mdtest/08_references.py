"""Port tier 08_references to the mdtest topic format.

Four topics: taking a reference, dereferencing it, the code-reference
pair, and the adjacency body. Prose is written here; every source,
pin, token fact and refusal record is copied by `gen.case`.
"""
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check

TIER = '08_references'
covered = set()

covered |= topic(TIER, 'taking-references.md', 'Taking a reference', """
The constructs that PRODUCE a reference -- the backslash in both
contexts, the anonymous array constructor -- and the builtin that reads
one back.

**Tier 08 references.** Introduces `anonlist`, `prototype`, `ref`,
`refgen`, `rv2cv`, `rv2sv`, `srefgen`. Depends on 07_subroutines.

Every case takes a reference to a VARIABLE, never to a literal. `my $r
= \\1` arrives as `const[IV \\1] s/FOLD` and emits no `srefgen` at all:
the optimiser erases the very construct the tier is about. The array
and hash fixtures are what defeat that, and it is the same lesson tier
01 learned from `1+2`.

Nor does any case print a reference. A reference stringifies as
`ARRAY(0x5606f0a12345)` and the address changes every run, so `ref`
supplies the stable category name instead.
""", [
    ('01_backslash_scalar.t', 'The backslash in scalar context', """
`\\@a` in scalar context is one op, `srefgen`, and the array it points
at stays reachable through the scalar that holds it.

The optree is `padav[@a] lRM`, then `srefgen sK/1`, then
`padsv_store[$s]` -- the array is loaded first and the reference taken
of it, which is why this tier cannot precede tier 02.
"""),
    ('02_backslash_list.t', 'The same backslash in list context', """
The SAME backslash in list context is a DIFFERENT op: `\\(@a)` emits
`refgen`, not `srefgen`, and distributes over the array's elements.

This case and the scalar one above differ in source only by the
parentheses and the assignment target, and perl compiles them to two
different ops. That is why the tier claims both: one spelling in the
source is two ops in the optree, and a corpus that wrote only the
scalar form would claim an op it never exercised.

The distribution is the second surprise. `\\(@a)` is not a reference to
the array -- it is a LIST of references, one per element, so `$r[0]` is
a SCALAR reference and `${$r[0]}` is the element behind it. Writing
`$r[0]->[0]` instead dies with "Not an ARRAY reference", which is how
this case arrived at its current form.
"""),
    ('03_anonymous_array.t', 'The anonymous array constructor', """
`[10, 20, 30]` builds an array and yields a reference to it in one op,
`anonlist`, with no named array anywhere in the program.

The anonymous HASH constructor is deliberately NOT here. `{}` and
`{ %a }` compile to `emptyavhv` and `anonhash`, which tier 11 claims
where objects are built; writing one in this tier would use an op a
LATER tier owns and the lint would refuse it. `[...]` alone is tier
08's.

`ref` is what makes the assertion deterministic -- it prints the stable
category name rather than an address that changes every run.
"""),
    ('09_ref_builtin.t', '`ref` over all three referent kinds', """
`ref` reports what a reference points at, and it is the one operation
in this tier that has no home in an earlier one. The tier description
does not mention it; the op measurement is where it came from. Its
operand is a reference and its result is a plain string, so no earlier
tier can claim it -- there is nothing for it to be about before this
tier.

It is also what every other case here leans on for determinism.

The three categories are measured together because the op is the same
for all of them: `ref` does not discriminate by sigil, the referent
does.
"""),
])

covered |= topic(TIER, 'dereferencing.md', 'Dereferencing', """
Five spellings and one hash form, all reaching through a scalar that
holds a reference -- and the reason the op stream cannot tell most of
them apart.

**Tier 08 references.** Introduces `anonlist`, `prototype`, `ref`,
`refgen`, `rv2cv`, `rv2sv`, `srefgen`. Depends on 07_subroutines.

THE OP STREAM COLLAPSES THESE SPELLINGS. `@{$r}`, `@$r` and `$r->@*`
all emit `rv2av` and are indistinguishable in it; `$r->[2]` and
`${$r}[2]` both use `multideref` and differ only in op COUNT and a
private flag, the brace form's extra op being tier 01's `padsv`. A
tier lint that compares op-name sets alone sees these cases as
identical. The `tokens` claims are the only check that reads the
spelling rather than the result, which is why every case below carries
one.

Both of the tier's hard markers live here: `deref-brace` (`${`) and
`deref-at` (`@{`).
""", [
    ('04_arrow_deref.t', 'The arrow dereference', """
`$r->[2]` is ONE op: a `multideref` that has absorbed the pad lookup of
`$r` along with the subscript.

MEASURED perl 5.42.0, the whole print statement:

    a  <0> pushmark s
    b  <+> multideref($r->[2]) sK
    c  <$> const[PV "\\n"] s
    d  <@> print vK

The subscript chain is fused. `$r->[2]` does not emit `padsv` then
`rv2av` then `aelem`; it emits one op naming the whole path, base
included. This is the case the brace form below is compared against,
and the pair is this tier's version of tier 01's `.5` problem: two
spellings that mean the same thing and must still be told apart. This
one has an arrow and no `${`; the next asserts the reverse.
"""),
    ('05_brace_deref.t', 'The brace dereference of an array reference', """
`${$r}[2]` means what `$r->[2]` means and does NOT compile to the same
ops: the brace form leaves the pad lookup as a separate `padsv`.

MEASURED perl 5.42.0, the whole print statement:

    a  <0> pushmark s
    b  <0> padsv[$r:1,4] sM/DREFAV
    c  <+> multideref(->[2]) sK
    d  <$> const[PV "\\n"] s
    e  <@> print vK

Against the arrow form's single `multideref($r->[2])`: same result,
same op NAMES plus one that tier 01 already owns. The difference is
visible only in the op COUNT and the private flag, so an op-level lint
cannot distinguish the two. This case has no arrow at all, which no
op-level check can see.
"""),
    ('06_deref_at.t', 'The `@{ }` whole-array dereference', """
`@{$r}` dereferences a scalar as a whole array, and the op it emits --
`rv2av` -- is tier 02's, reached here by a route tier 02 does not have.
This is the hard marker `deref-at`, and the spelling tier 11's method
bodies are written in.

THE DEREF IS WRITTEN TWICE, and that is the point. The interpolated
`"@{$r}\\n"` is what makes the output deterministic: printing the list
bare would run the three elements together as `102030`, which is true
but reads as a bug. But inside a double-quoted string our lexer
produces ONE `Quote` token and the dereference is not separately
tokenised at all -- so a case with only the interpolated form asserts
nothing about its own construct, and its negative fact would be
satisfied entirely by the `\\@a` two lines above. The bare `my @c =
@{$r};` is what gives the lexer the construct to tokenise. Measured,
our lexer emits `DerefSigil(@) Operator({) Variable($r) CloseBracket(})`,
and a lexer that emitted the single token `Variable("@{$r}")` instead
would satisfy every behavioural pin in this tier.

The copy also keeps the deref in the optree in a second place: the
interpolation compiles through `join`, tier 03's, while the assignment
is a plain `rv2av` feeding an `aassign`.
"""),
    ('10_deref_at_sigil.t', 'The sigil-only `@$r` dereference', """
`@$r` is the SAME op as `@{$r}` -- `rv2av` -- and the op stream cannot
tell which was written. Only the tokens can, and the interpolated case
above cannot make that claim, because there the whole string is one
token.

Measured, our lexer produces two tokens, `DerefSigil(@) Variable($r)`,
and a lexer that produced the single token `Variable("@$r")` instead
would satisfy every behavioural pin in this tier: perl prints the same
bytes and compiles the same ops.

MEASURED perl 5.42.0, the assignment statement -- identical to the
postfix case's except for the subscript constant its last statement
reads:

    f  <0> pushmark s
    g  <0> padsv[$r:2,4] s
    h  <1> rv2av[t5] lK/1
    i  <0> pushmark s
    j  <0> padav[@c:3,4] lRM*/LVINTRO
    k  <2> aassign[t6] vKS/COM_AGG

Copying into a named array rather than printing the list keeps `join`
in tier 03 where it belongs: printing `@$r` bare would run the three
elements together as `102030`, which is true and reads as a bug.

This case once also carried `no operator whose text is "->"`, and an
earlier note said that fact was "satisfied by the `\\@a` on the line
above". That was wrong about the mechanism: a negative fact is
satisfied by ABSENCE, not by some other token standing in for it. The
fact was vacuous because the source contains no `->` bytes at all, so
no lexing of it could produce one. It was deleted along with eight
siblings by `01a0cfb2`.
"""),
    ('11_deref_postfix.t', 'The postfix `$r->@*` dereference', """
`$r->@*` is the THIRD spelling of the same `rv2av`, and the only one of
the three that puts an arrow in front of a sigil. This case is the
proof rather than the assertion: its optree and the sigil case's are
identical op for op, differing only in the subscript constant the last
statement reads.

MEASURED perl 5.42.0, the assignment statement:

    f  <0> pushmark s
    g  <0> padsv[$r:2,4] s
    h  <1> rv2av[t5] lK/1
    i  <0> pushmark s
    j  <0> padav[@c:3,4] lRM*/LVINTRO
    k  <2> aassign[t6] vKS/COM_AGG

So the tokens carry the whole distinction, and they carry it in a
direction the other two spellings do not go. `@*` is a SIGIL AND A STAR
after an arrow, which a lexer may reasonably read as the glob `*` --
measured, ours emits `Operator(->) Variable(@*)`, and a lexer that
emitted `Operator(->) Operator(@) Operator(*)` would parse as a
multiplication and return no Unknown at all. The `one operator whose
text is "->"` claim pairs with the arrow case's identical one to say
the arrow survived; the negative says the lexer did not swallow the
construct whole.

Postfix dereference has been stable since 5.24 and is not experimental;
`perlref` documents it under "Postfix Dereference Syntax".
"""),
    ('07_deref_brace_hash.t', 'The `%{ }` and `${ }{ }` hash dereferences', """
`%{$r}` dereferences a scalar as a whole hash, and `${$r}{a}` reaches
one of its values -- both without any anonymous hash constructor. This
is the tier's other hard marker, `deref-brace`.

The hash is a NAMED one taken a reference to, not `{ a => 1 }`. The
anonymous hash constructor compiles to `emptyavhv`/`anonhash`, which
tier 11 claims where objects are built, and a tier-08 case emitting a
tier-11 op is exactly what the lint refuses. `\\%h` reaches the same
place through `srefgen`, which this tier owns.

`scalar(keys %copy)` emits no `keys` op: measured, the optimiser fuses
it into `padhv[%copy] sM/KEYS`. Another instance of more construct,
fewer ops.
"""),
])

covered |= topic(TIER, 'code-references.md', 'Code references', """
`\\&foo` on a named sub, the call through the result, and the builtin
that reads the prototype back out of one.

**Tier 08 references.** Introduces `anonlist`, `prototype`, `ref`,
`refgen`, `rv2cv`, `rv2sv`, `srefgen`. Depends on 07_subroutines.

THIS IS WHAT PINS THE TIER BELOW 07 rather than at 03. Everything else
here needs only tier 02's aggregates; `\\&foo` needs a named sub to
exist, and a tier introducing `rv2cv` before subroutines existed would
be claiming an op for a construct it could not write.

`rv2cv` is narrower than it looks. It appears ONLY for `\\&foo` on a
named sub: `&{$r}()` and `&$r()` both compile to `entersub` with no
`rv2cv` at all. So these two cases are the tier's only source for that
op.
""", [
    ('08_code_ref.t', 'The code reference and the call through it', """
`\\&twice` on a NAMED sub is the only construct in this tier that emits
`rv2cv`, and `$c->(21)` calls the result through tier 07's `entersub`
-- the same "different route to an earlier tier's op" this tier does
with `rv2av` and `multideref`. Removing this case would leave the
README claiming an op nothing emits.
"""),
    ('12_prototype_builtin.t', '`prototype` reads the parser\'s own input back', """
`prototype \\&f` hands back, as a runtime STRING, the very text that
changed how calls to `f` parse. The parser's own input is readable as
data.

WHY THIS TIER AND NOT 07. `prototype` takes a CODE REFERENCE, and a
code reference is this tier's construct. Measured, `prototype \\&f`
emits `rv2cv` and `srefgen`, and both are THIS tier's. Tier 07 declares
a prototype and measures what it does to a call site, but tier 07
cannot WRITE this construct, because `\\&f` is not available to it. Tier
08 is the earliest tier in which `prototype` is spellable at all, so
the op lint and the placement argument agree rather than merely not
conflicting.

WHAT TIER 07 ALREADY CLAIMS, AND WHAT THIS ADDS.
`07_subroutines/09_prototype_extent.t` measures the `($)` prototype's
effect on PARSING: `print g 1, 2` cuts the argument extent to one and
the second constant falls through to the enclosing `print`. That claim
is about a CALL SITE, where the prototype is invisible and only the
output reveals it. This case makes the complementary claim and repeats
none of it: there is no parenless call here, and the prototyped sub is
never called at all. What is here is the prototype coming back OUT as
the two-character string `($)`.

THE WRONG PARSE THIS RULES OUT. `prototype \\&f` is a word immediately
followed by a reference operator applied to a sub sigil, with no
parentheses to bracket the argument. A lexer that read `\\&` as ONE
token -- a plausible reading, since the two characters only ever occur
together in this tier -- produces a program whose optree still contains
`srefgen` and whose output is unchanged, because perl's ops and perl's
bytes cannot see the difference. The token facts are the only place
that reading dies: exactly one `\\` operator, and NOT one whose text is
`\\&`.

`prototype` is the one word in the source, which is also a fact worth
pinning: a lexer that folded `prototype \\&f` into a single term the way
it might a quote-like operator would leave no separate `prototype` word
behind. Its own text appears NOWHERE ELSE in the source, so the count
is a real claim rather than a coincidence of the fixture.

MEASURED perl 5.42.0, the whole program:

    3  <#> gv[IV \\"$"] s
    4  <1> rv2cv[t3] lKRM/AMPER,TARG
    5  <1> srefgen sK/1
    6  <1> prototype sK/1

`gv[IV \\"$"]` rather than `gv[IV \\&main::f]`: with a prototype in force
perl has folded the sub's identity into its prototype string at compile
time, the same substitution `07_subroutines/09_prototype_extent.t`
records at its `entersub`. So the op stream does not even carry the
sub's NAME here, which is a further reason the token facts must.
"""),
])

covered |= topic(TIER, 'adjacency-08_references.md', 'Every reference construct, each beside another', """
One body holding every construct this tier introduces, each adjacent to
another -- and adjacent to tier 07's subroutines, which is what this
tier depends on.

**Tier 08 references.** Introduces nothing of its own; it is the
mixture that is the subject. Depends on 07_subroutines.

WHY A MIXTURE NEEDS ITS OWN CASE. The tier's other cases are one
construct each, which is what makes them diagnosable: when the brace
dereference case refuses, the construct that refused is the only one
present. That same property is why such a corpus cannot reach an
ADJACENCY bug -- a parser that handles every construct alone and
mis-handles a pair goes green over the pair.
""", [
    ('00_adjacency.t', 'The whole tier in one body', """
The constructs here are the backslash in both contexts, the anonymous
array, the arrow, all three whole-aggregate dereferences, the brace
element derefs, the code reference, `prototype` and `ref` -- adjacent
within one statement where the construct allows it, and on consecutive
statements where it does not. The two print statements put ten of them
side by side, which is where a parser that handles `$$s[0]` and
`${$l}[1]` separately but not next to each other would show it.

THE THREE SPELLINGS OF ONE OP are all here on consecutive lines:
`@{$s}`, `@$s` and `$s->@*` each emit `rv2av` and nothing in the op
stream separates them, so a parser that reads two of the three and
guesses at the third produces an identical optree for the wrong
program. Three lines apart is where that guess has to hold.

The tier's declared prerequisite is 07_subroutines, so the adjacency
pairs with it: `sub twice` is tier 07's, `\\&twice` is this tier's, and
`$c->(21)` is the arrow applied to the result -- the two tiers touching
in one expression rather than in two separate files.

`\\(@a)` is the one that needs care. It distributes: `@r` holds a list
of SCALAR references, one per element of `@a`, so `${$r[1]}` is 20 and
`$r[1]->[0]` dies with "Not an ARRAY reference". That failure is what
this body's first draft printed.

`prototype \\&one` is here for the reason every other construct is:
adjacency. It sits between the code reference it needs and the
dereferences around it, so a parser that reads `prototype \\&one` alone
but loses the following `@{$s}` shows it here and nowhere else. `sub
one ($)` carries the prototype the call never uses -- the sub is
declared to be REFLECTED, not to be called.

The hash is a NAMED one taken a reference to, not `{ k => 1 }`. The
anonymous hash constructor compiles to `emptyavhv`/`anonhash`, which
tier 11 claims, and a tier-08 case emitting a tier-11 op is what the
dependency lint refuses.
"""),
])

check(TIER, covered)
