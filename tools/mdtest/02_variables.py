"""Build tier 02's topic files. Prose is written here; claims are copied."""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check

TIER = '02_variables'

# The INTRODUCES/DEPENDS ON paragraph every topic in this tier carries.
# Repeated rather than shared because the op-budget lint reads it out of
# each file, and a topic that did not declare its tier would not be linted.
DECL = """**Tier 02 variables.** Introduces `aassign`, `aelem`, `aelemfast`,
`aelemfast_lex`, `aelemfastlex_store`, `aslice`, `av2arylen`, `delete`,
`each`, `gv`, `gvsv`, `helem`, `hslice`, `multideref`, `padav`, `padhv`,
`push`, `rv2av`, `rv2hv`, `sassign`, `unshift`, `values`. Depends on
01_literals."""

covered = set()

covered |= topic(TIER, 'arrays.md', 'Arrays', """
Naming an array, counting it, and reaching one element of it. Four ways
of writing a subscript, which perl compiles to four different ops.

""" + DECL + """

`scalar(@a)` throughout rather than `print "@a"`. Measured, the
interpolation emits `join` and a `gvsv` for `$"` -- the `gvsv` is `$"`,
which the interpolation reads -- and `print "$a[0]"` emits `stringify`.
None of those is naming a variable, and a case that printed `"@a"` would
drag two unclaimed ops into the tier.
""", [
    ('01_array.t', 'An array holds a list',
     """The tier's baseline, the way `01_binary.t` is tier 01's: every
other case here starts by filling an array or a hash, so a regression
here explains all of them at once rather than being diagnosed nine
times.

The ops are `padav` for the array itself and `aassign` for the list
assignment. There is no `scalar` op: `scalar(@a)` compiles to the padav
in scalar context, so the keyword leaves no trace of its own."""),

    ('02_array_element.t', 'A constant subscript reads as one op and writes as another',
     """`$a[0]` on the right of `=` is not the same op as `$a[0]` on the
left. This is the asymmetry the tier README calls out: read and write
are the same three characters in the source, and a parser that treats
them as one node is not wrong -- but perl does not, and a corpus that
never writes an element cannot see the second op at all.

The write is `aelemfastlex_store` and the read is `aelemfast_lex`. Both
are the `_lex` spellings because `@a` is a lexical; the package
spellings are in the names topic."""),

    ('03_array_last_index.t', '`$#a`, and the subscript the optimiser will not fold',
     """Two claims in one body because the second only exists in terms of
the first. `$#a` alone is `av2arylen`. `$a[$#a]` is the case where the
subscript is an EXPRESSION rather than a constant, so perl builds a
plain `aelem` instead of the `aelemfast_lex` above -- which makes
`aelem`, the op a reader would expect to be this tier's centre, an edge
case reachable only by writing the subscript this way."""),

    ('04_array_slice.t', 'An array slice: the sigil decides, not the name',
     """`@a[0, 2]` is the `@` sigil on an array name with a LIST
subscript, returning several elements rather than one. The sigil is the
whole point: `$a[0]` and `@a[0]` name the same array and differ only in
what they hand back, so a lexer that binds the sigil to the name and a
parser that binds it to the subscript form disagree here and nowhere
else in the tier.

`13`, not `1 3`: the slice's two elements arrive as separate arguments
to `print` and `$,` is unset, so nothing separates them -- the same
reason tier 01's adjacency file prints `qw(a b c)` as `abc`."""),
])

covered |= topic(TIER, 'hashes.md', 'Hashes', """
Naming a hash, subscripting it with braces, and the two words -- `exists`
and `delete` -- that mostly leave no op behind.

""" + DECL + """

Hash order is not guaranteed, so nothing below prints a hash's contents.
`scalar(keys %h)` asks how many rather than which, and that is the count
every case here pins.
""", [
    ('07_hash.t', 'A hash, and a bareword key',
     """A hash is named with `%` and subscripted with braces; `$h{a}` is
one element of it, and the bareword key needs no quotes.

`$h{a}` emits no `helem`. It compiles to `multideref($h{"a"})`, the op
that swallows most element access in this tier; `helem` survives only
where the subscript is an expression, which is the next case. `keys`
likewise emits no op of its own here -- it becomes a FLAG on the
`padhv`, `sM/KEYS`. That flag is a property of SCALAR context, not of
the keyword; in list context `keys` emits an op, which is why the
aggregate-operators topic writes `values` and no `keys` at all."""),

    ('08_hash_element_expr.t', 'A computed subscript is the only route to `helem`',
     """`$h{a}` is `multideref`. `$h{$k[0]}` is `helem` over an
`aelemfast_lex` -- same construct in the source, different op, decided
entirely by what is inside the braces. This case exists because it is
the only way to reach `helem` at all, and a tier that claimed `helem`
without it would be claiming an op no file emits."""),

    ('09_hash_exists_delete.t', '`exists` and `delete` on an element leave no op of their own',
     """Measured, `exists $h{a}` compiles to `multideref($h{"a"})
sK/EXISTS` and `delete $h{a}` to `multideref($h{"a"}) sK/DELETE`: both
survive only as a FLAG on an op named after something else. A tier
derived from op names alone would contain no notion of `exists` at all,
and the slice case below would be the only evidence `delete` exists --
for the wrong reason, since the slice is the spelling where the
optimiser DECLINES to fold. This case is the common path; that one is
the exception.

BOTH TRUTH VALUES OF `exists` ARE HERE because the false one is a
different claim. Perl's false is the empty string, so `print exists
$h{z}` prints NOTHING and the pinned output carries a blank line -- a
case that tested only the true value could not tell `exists` from a
construct that always yields 1.

`print exists $h{a}` rather than `print exists $h{a} ? 1 : 0`: measured,
the conditional adds `cond_expr` and `0+(...)` adds `add`, and both are
ops later tiers introduce. The plain print is the only spelling of this
construct that stays inside the tier.

`delete` in scalar context returns the value removed, which is what
makes it observable without printing the hash."""),

    ('10_hash_slice.t', 'A hash slice, where `delete` becomes a real op',
     """`@h{...}` is a hash slice. `delete $h{a}` emits NO delete op, but
deleting a SLICE is different: the optimiser declines to build a
multideref for it, so `delete vK/SLICE` appears as a real op. That is
why the tier claims `delete` for the slice and not for the element, and
why `exists` is not claimed at all -- it has no op in any spelling."""),
])

covered |= topic(TIER, 'names.md', 'Names: braces and packages', """
Where the variable's NAME is the question rather than what it holds.
Braces around a name are punctuation; `$::` is a package-qualified name,
the scanner row measured at 0.0% clean over fourteen files.

""" + DECL + """

The package spellings are also where the tier's op list stops matching
the source a reader would write first. `sassign` -- the plain scalar
assignment -- is introduced by `$::x = 1`, not by `my $x = 1`, and the
four ops for a constant array subscript are only all reachable once both
a lexical and a package array are in the corpus.
""", [
    ('05_braced_name.t', '`${x}` is one variable, not a dereference',
     """GLOSSARY.md says this outright under `variable` -- "`${name}` is
one variable: the braces are punctuation around a name. But `${ $ref }`
is NOT one token, because its contents are an expression requiring a
parser." The two spellings differ by what is inside the braces and by
nothing else, which makes this the tier's sharpest lexing question: a
lexer that sees `${` and commits to a dereference is wrong here, and a
lexer that sees `${` and commits to a name is wrong at tier 08. This
case takes the half that belongs to this tier; `${ $ref }` is tier 08's,
and writing it here would be reaching forward.

BOTH SIGILS ARE BRACED because the brace rule is about the NAME rather
than about scalars: `@{a}` is the same array as `@a`, and a lexer that
special-cased `${` would pass a case that only wrote the scalar form.

The ops are `padsv` and `padav`, identical to the unbraced spellings:
perl resolves the braces away entirely, so the optree cannot tell this
case from one without them. The claim is a LEXICAL one and only the
source records it -- the same situation tier 01 recorded for `-1`, whose
two tokens fold to one constant."""),

    ('11_package_array.t', 'A package array subscripts through `rv2av`',
     """`$a[0]` on a lexical is `aelemfast_lex`. `$::a[0]` is `aelemfast`
behind an `rv2av` over a `gv`. Constant subscript, lexical or package,
read or write -- four ops for what reads as one construct, which is why
the tier's adjacency case carries all four rather than a representative
one."""),

    ('12_package_hash.t', 'A package hash reaches storage through `rv2hv`',
     """The element access is `multideref` either way -- the optimiser
folds the glob lookup into the deref chain just as it folds the pad
lookup -- so the difference this case pins is in naming the hash itself,
not in subscripting it. `scalar(keys %::h)` is what forces the bare name
into the optree, and that is `gv` then `rv2hv`."""),

    ('13_package_scalar.t', 'A package scalar is where `sassign` enters the corpus',
     """Tier 01's `my $x = 0.5` emits `padsv_store` and no `sassign` at
all -- the lexical store is one op, not an assignment over a variable.
The package scalar is the first place the two halves separate: `gvsv`
fetches the glob's scalar slot and `sassign` puts the value in it. So
the op a reader would look for in tier 01 is introduced four constructs
into tier 02, by the spelling nobody writes first.

`$::x` rather than `$main::x`: `::` with an empty package name IS
`main`, and the short spelling is the one that makes the lexing question
visible -- whether `$::` is a sigil plus a name that begins with a
separator."""),
])

covered |= topic(TIER, 'aggregate-operators.md', 'The aggregate-argument operators', """
`push`, `unshift`, `values` and `each` share an ARGUMENT RULE nothing
else in the corpus has: the first argument is the aggregate ITSELF, not
an expression to be flattened.

""" + DECL + """

Measured, `push @a, @tail` emits `padav[@a] lRM` and `padav[@tail] l` --
the SAME op with different flags, the container slot against the
flattened slot. A parser that flattens the first slot builds a tree perl
does not build while printing something plausible, which is why these
cases pin an ELEMENT as well as a count.

They also emit ops of their own, which is the opposite of what the
`exists`/`delete` cases would lead a reader to expect, so the contrast
is worth stating. `keys` is the trap: `scalar(keys %h)` compiles `keys`
away to the flag `sM/KEYS`, but measured, `my @k = keys %h` in LIST
context emits a `keys` op just as `values` does. The flag is a property
of the CONTEXT, not of the keyword, so nothing below writes `keys` at
all rather than quietly adding an op the tier has never claimed.
""", [
    ('14_push.t', '`push`: the first argument is an array, not an expression',
     """Every other list operator in the corpus takes expressions and
flattens them. `push` does not: its first argument slot holds the array
ITSELF, and only the arguments after the comma are flattened into it. A
parser that treats `push @a, @tail` as both arrays flattened builds a
tree perl does not build, while producing output plausible enough that
nothing behavioural would notice.

`lRM` is the aggregate slot -- lvalue, ref-modify, the array as a
container. `l` is the flattened slot. The flags are the whole
distinction, and this case pins THREE consequences of them, because each
alone is reachable by a wrong tree. `@tail` still holds TWO elements:
the flattened slot is read, not consumed, so a parser that MOVED the
elements rather than copying them fails here and nowhere else. `@a`
holds THREE, which is the count a flattening parser would also reach --
so the count alone proves nothing, and it is printed only to make the
third line legible. `$a[2]` is 30, the LAST element of `@tail`, which
pins the order the flattening put them in.

There is no constant-folding trap here, and the tier's `$ENV{X}` idiom
would be wrong to copy: measured, `push @a, 3` with everything constant
still emits `push`, because an array is a runtime container and the
optimiser has nothing to fold it into. (`//` is tier 04's `dor`, so the
idiom is out of budget here regardless.)"""),

    ('15_unshift.t', '`unshift`: the same rule at the other end',
     """This case exists because the rule and the OP are separable
claims. A parser can hard-code `push`'s aggregate first slot as a
special case for the one keyword it was tested on, and `unshift` is the
second keyword that rule has to cover. T1 uses it in 15 files.

Measured, the flags are `push`'s exactly over a different op: `lRM` on
the container, `l` on the flattened list, and `unshift` where `push` had
`push`. So the tier claims a second op and no new rule.

The ELEMENT is what separates the two behaviourally. `@head` still holds
ONE element -- the flattened slot read, not consumed, which is the
`push` case's first line mirrored. `@a` holds three, the count a
flattening parser also reaches. `$a[0]` is 10, the value that arrived
from `@head`, and that is the line separating this operator from `push`:
under `push` the same three-element result would have 20 there. A parser
that got the operator right and the END wrong passes a count check and
fails this.

The negative token fact is the falsifying half: NO word spelled `shift`.
A lexer that read `unshift` as `un` followed by `shift` -- or that
longest-matched the keyword table wrongly -- would produce a `shift`
here and fail. `shift` is tier 11's op, so it is also the spelling this
case must not accidentally contain."""),

    ('16_values.t', '`values` takes the container, not a flattened list',
     """`values %h` is not `values(%h)` with the hash flattened into a
list of key-value pairs. The hash is the argument, whole, and the
operator reads its values out of the container. A parser that flattens
`%h` first hands `values` a flat list of scalars where perl hands it one
hash -- and for a one-key hash the printed answer would still look
right, which is why this case pins the op and a token fact rather than
only the output.

Measured, the hash reaches the operator as `padhv`, the container op,
not as an aassign'd list: `padhv[%h:1,3] lRM` then `values[t4] lK/1`.

ONE KEY, DELIBERATELY. Hash order is not guaranteed, so a case printing
the values of a two-key hash would pin an expectation perl does not
promise. With one key the list has one member and its value is
determined."""),

    ('06_each.t', '`each` returns a PAIR where its siblings return a flat list',
     """`values` and `keys` take the same argument and return a flat list
of one thing per entry. `each` takes the same argument and returns TWO
-- one key and one value -- so the list-assignment on its left has two
scalars to fill. A parser that models `each` on its siblings gives the
construct the wrong arity and the pair silently collapses to the key.

Measured, the hash arrives as `padhv` and `each` is a real op, not a
flag: `padhv[%h:1,3] lRM`, `each lK/1`, then `padrange[$k:2,3; $v:2,3]
RM/LVINTRO,range=2` and `aassign[t5] vKS/COM_RC1`. The `range=2` on the
padrange is the arity, visible: two pad slots are introduced as one
range because the assignment's left side is a two-element list. A
one-element return would have had a padsv there.

NO LOOP, WHICH IS THE POINT OF THE SPELLING. T1 writes `while (my ($k,
$v) = each %h)`, and measured, that form drags in `enterloop`,
`leaveloop`, `and` and `unstack` -- ops belonging to tiers 05, 04 and
06. A bare list assignment reaches the same `each` with nothing this
tier does not claim, and the iterator's LOOPING is a control-flow claim
rather than a claim about how `each` parses.

ONE KEY, for the `values` case's reason: hash order is not guaranteed,
so a two-key hash would make the pair unpredictable and the pinned
output a guess. With one key the first iteration is determined."""),
])

covered |= topic(TIER, 'adjacency-02_variables.md', 'Every construct, each beside another', """
One body holding every construct this tier introduces, each adjacent to
another, and adjacent to tier 01's literals.

""" + DECL + """ The mixture itself is the subject; it introduces
nothing of its own.

The tier's other sixteen cases are one construct each, which is what
makes them diagnosable: when the computed-subscript case refuses, the
construct that refused is the only one present. That same property is
why a corpus of such cases cannot reach an ADJACENCY bug -- a parser
that handles every construct alone and mis-handles a pair goes green
over the pair.

THIS TIER'S ADJACENCY HAS A SHAPE THE ONE-CONSTRUCT CASES CANNOT HAVE.
Four of the tier's ops are the SAME construct spelled four ways -- a
constant subscript, lexical or package, read or written -- and the cases
that isolate them each see one. Here all four sit in one body, so a
parser that collapses `$a[0]` read onto `$a[0]` written, or a lexical
array onto a package one, is visible.

DEPENDS ON 01_literals, so the pairing is real and not internal: the
literals are what the variables hold. `qw(x y)` binds into an array,
`"$w[0]-$w[1]"` interpolates two elements back out, and the numbers and
strings of tier 01 are the values every subscript here returns. That
pairing is checked on the SOURCE rather than the ops -- measured, a
program with no tier-01 construct in it still emits six of tier 01's
nine ops, because `print` and a statement are themselves tier 01's, so
an op-based pairing check would pass over a body holding nothing of the
prerequisite at all.
""", [
    ('00_adjacency.t', 'The whole tier in one body',
     """`delete` appears twice because it is two constructs. On the slice
it is a real `delete` op; on the element it is a FLAG on a `multideref`,
and `exists` is the same flag position with no op in any spelling. The
hash starts with four keys so the slice can remove one, the element
delete another, and two remain to be looked up.

The runs are unseparated because `$,` is unset and each `print` gets its
arguments as a list, the same reason tier 01's adjacency file prints
`qw(a b c)` as `abc`. Line by line: `$a[0]` `$a[$#a]` `$#a`
`scalar(@a)` `$tag`; then `@a[0,1]` `$h{a}` `$h{$k[0]}` `@h{"a"}`
`scalar(keys %h)`; then `exists` true, `exists` false printing nothing,
the deleted value and a braced array name; then the three package
variables; then the four aggregate-argument operators.

THE AGGREGATE OPERATORS COME AFTER THE PRINTS, which is not the layout a
reader would choose, and the reason is that they MUTATE `@a`. Placed
before, they would shift every subscript the earlier lines print and
turn a case about adjacency into a case about arithmetic on indices.
Placed after, they are still in the same compiled body -- which is all
adjacency claims -- while the four established lines stay exactly what
they were.

They are adjacent to the constructs that matter to them: `push` and
`unshift` take the `@a` the element and slice lines have been reading,
`unshift`'s argument is the `$#a` of the last-index case, and `values`
takes the `%h` the deletes have already thinned to two keys. A parser
that flattened an aggregate first slot would push the values of `%h`
into a `@a` it had also flattened, and `scalar(@a)` would not be 5.

`%one` is a SEPARATE, one-key hash, and `each` needs it. Hash order is
not guaranteed, so `each %h` over the four-key hash would return an
unpredictable pair and the pinned line would be a guess; over a one-key
hash the first iteration is determined. `values %h` is safe on the big
hash only because it is wrapped in `scalar`, which asks how many rather
than which.

`print $a[0]` rather than `print "$a[0]"` throughout: the interpolated
subscript emits `stringify` and an interpolated `@a` emits `join` plus a
`gvsv` for `$"`, none of which this tier claims. `$tag` is built by a
separate statement so the interpolation is tier 01's `multiconcat` over
two already-fetched elements rather than a fetch inside the print."""),
])

check(TIER, covered)
