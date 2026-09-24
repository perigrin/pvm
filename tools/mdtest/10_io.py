"""Port tier 10_io to mdtest topics.

Four topics: opening a handle and reading from it, writing to one and
choosing where "no handle" points, the symbol-table slice that reaches
the same glob from three sides, and the adjacency body.
"""
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check

TIER = '10_io'

covered = set()

covered |= topic(TIER, 'handles.md', 'Opening a handle and reading from it', """
`open`, `close`, `readline` and `eof`: a handle held in a lexical, and
the one op whose meaning is decided by what receives it.

**Tier 10 io.** Introduces `close`, `eof`, `open`, `readline`, `rv2gv`,
`say`, `select`, `sselect`. Depends on 03_context.

THIS IS WHERE THE TIER'S PLACEMENT COMES FROM. `<$fh>` is ONE op with
two behaviours: measured under 5.42.0 it compiles to `readline sKS/1`
when a scalar receives it and `readline lK/1` when an array does -- the
same op, a different flag, one line against every remaining line. A
parser that cannot say which context an expression is in cannot say
what `<$fh>` returns, and it cannot learn that from the readline
itself.

Every handle here is in-memory, opened on a scalar ref. These files
must be reproducible and a file that opens a real path is not --
`/etc/hostname` differs per machine. Measured, that costs nothing:
`\\"x\\n"` is constant-folded to `const[IV \\"x\\n"] s/FOLD` at compile
time, and the op stream of `open(my $fh, "<", \\"x\\n")` is identical op
for op to `open(my $fh, "<", "/etc/passwd")`.

No case checks its open. The idiomatic `open(...) or die` emits a `die`
op and no tier claims `die`, so a check would widen this tier's
declared set by an op belonging to a tier that does not exist yet. The
cases open unchecked and print a constant to show they ran.
""", [
    ('01_open_close.t', 'A three-argument open autovivifies the glob', """
`open(my $fh, "<", ...)` does not hand the fresh lexical a value: it
autovivifies a glob into the scalar's slot, through `rv2gv` with
`DREFSV` set. The handle IS the glob, which is the fact the rest of the
tier rests on.

Measured, the in-memory form against a real path -- `padsv rv2gv const
const open` for both. Five ops, the same five. The scalar ref is the
ordinary open as far as the op table can see.
"""),
    ('02_readline_scalar.t', '`<$fh>` in scalar context reads ONE line', """
The half of `readline` a scalar receives, and the baseline the list
case deviates from by one sigil. The handle holds two lines and this
case prints one, so the output is also the evidence that the scalar
form STOPPED -- a case reading a one-line handle could not tell the two
contexts apart by behaviour at all.

THE TOKEN FACTS ARE WHERE THE ANGLE SPLIT IS MADE. `<$fh>` and `<*.c>`
share a spelling and belong to different tiers: the first is this
tier's IO construct, the second is tier 13's delimiting problem.
Measured, they part at the optree -- `padsv readline` against `pushmark
const gv glob` -- and they are the same thing to a lexer, which cannot
tell a handle name from a pattern without knowing what the name means.

The behavioural half cannot make that claim. Measured, our parser reads
`<`, `$fh`, `>` as a perfectly good comparison chain and returns ZERO
Unknowns for it, so a lexer that split the angles would satisfy every
behavioural assertion here unnoticed. `no operator whose text is "<"`
is what falsifies it: a split lexer emits that operator, an unsplit one
never does.
"""),
    ('03_readline_list.t', '`<$fh>` in list context reads EVERY remaining line', """
The same `readline` op, flagged `lK/1` rather than `sKS/1`. One op, two
behaviours, selected by what receives it -- which is why this tier
declares `DEPENDS ON 03_context` and cannot precede it.

The count is printed rather than the lines, because the count is what
separates this from the scalar case. Printing the lines would show
`one` first in both and differ only in what followed.

The token facts repeat the scalar case's pair rather than sharing them,
because the spelling is what they assert and each case has its own.
`<$fh>` is ONE token in both, and the context that tells the two cases
apart is invisible to the lexer -- which is the point: everything
separating this tier's `<$fh>` from tier 13's `<*.c>` sits above the
lexer.
"""),
    ('06_eof.t', '`eof($fh)` inspects the glob without making it real', """
`eof` is its own op, and the `rv2gv` before it carries `FAKE`.
Measured: `close($fh)` is `padsv rv2gv sK*/1 close` and `eof($fh)` is
`padsv rv2gv sK*/FAKE,1 eof` -- the same `rv2gv` with a different flag,
because `eof` does not want the glob made real, only inspected. That is
the third job this tier's single claimed `rv2gv` does; the other two
are autovivifying a glob into `my $fh` at open and resolving a handle
operand at print.

The handle holds one line and the case reads it before asking, so `eof`
is true and the output is non-empty. Asking first would print nothing,
and a case whose output is empty cannot distinguish "eof was false"
from "the program did not run".
"""),
])

covered |= topic(TIER, 'writing.md', 'Writing to a handle', """
`print` to a named destination, `say`, and the operator that changes
what "no handle" means.

**Tier 10 io.** Introduces `close`, `eof`, `open`, `readline`, `rv2gv`,
`say`, `select`, `sselect`. Depends on 03_context.

`print` IS NOT IN THAT LIST AND IS THIS TIER'S SUBJECT. Tier 01 claims
it deliberately -- a literal has to be observed somehow -- and a tier
claims an op once. That is not a filing technicality: measured, `print
"x"` and `print $fh "x"` emit the same `print vKS`, and what changes is
what precedes it. The difference between printing and printing
SOMEWHERE lives entirely in the operands, so a tier declared as a set
of op names cannot express it.

NO CASE WRITES TO STDERR, recorded because it reads as an omission. The
`output` block is compared against stdout, so a line sent to STDERR
lands in neither stream the runner reads and asserting it would assert
bytes nothing checks. The op stream does not distinguish them --
`gv[*STDERR] rv2gv print` against `gv[*STDOUT] rv2gv print` -- so the
bareword case measures its construct through STDOUT and stays
observable.
""", [
    ('04_print_lexical_handle.t', '`print $fh "x"` differs from `print "x"` only in operands', """
Measured under 5.42.0:

    print "x"        pushmark const      print vK
    print $fh "x"    pushmark padsv rv2gv const print vKS

The op name is `print` in both. This is why `print` stays tier 01's op
and is not reclaimed here.

A handle opened onto a scalar with `\\my $buf` makes the write
observable without a file: `print $out` fills `$buf`, and printing
`$buf` to stdout is what the case is checked against. `\\my $buf` emits
`srefgen` -- tier 08's op, and a legal use here -- because unlike the
open case's `\\"x\\n"` there is no constant to fold.
"""),
    ('05_print_bareword_handle.t', 'A bareword handle reaches `print` through `gv`', """
Measured under 5.42.0:

    print STDOUT "x"   gv[*STDOUT] rv2gv sKR/1 const print vKS
    print $fh    "x"   padsv[$fh]  rv2gv sKR/1 const print vKS

The bareword and the lexical converge one op later. A file using only
lexical handles therefore never emits `gv`, and this tier's declared
set is the union across its cases rather than a property of any one of
them. `gv` itself is tier 02's op, claimed there with the package
scalar; reaching it through a bareword handle is a use, not an
introduction.
"""),
    ('07_say.t', '`say` is feature-gated, and the featureless spelling compiles', """
The same bytes parse two different ways depending on a pragma written
earlier in the file, and WITHOUT THE FEATURE THE SPELLING IS NOT A
SYNTAX ERROR. Measured:

    my $x=1; say $x                        pushmark padsv method_named[PV "say"] entersub
    use feature "say"; my $x=1; say $x     pushmark padsv say vK

Without the pragma perl reads `say $x` as indirect-object method
dispatch and it compiles clean -- `perl -c` says `syntax OK` -- failing
only at RUNTIME with `Can't locate object method "say" via package "1"`
and exit code 255. That is `11_oo/10_isa_infix.t`'s finding met from
the worse side: there the featureless form is a syntax error and fails
loudly, here it is a clean compile of the WRONG PROGRAM.

THE CORPUS CANNOT PIN THAT PARSE FROM THIS TIER. The featureless form
emits `method_named`, which is tier 11's op, and 11 is LATER than 10; a
case may only emit ops its tier or an earlier one claims. (`entersub`,
the other op it emits, is tier 07's and would have been a legal use;
`method_named` alone is what blocks it.) The measurement is recorded
and asserted nowhere. A case that pins it belongs in tier 11 beside the
`isa` file whose finding it mirrors.

WHY `say` IS HERE AND NOT BESIDE `print` IN TIER 01: it takes a
filehandle through the SAME `rv2gv` -- `say $out $line` is `padsv[$out]
rv2gv sKR/1 padsv[$line] say vKS` against `print $out "x"`'s
`padsv[$out] rv2gv sKR/1 const print vKS`. The handle machinery is
identical; what `say` adds is a record separator, which is a property
of writing to a HANDLE.

THE NEWLINE BETWEEN `said` AND `printed` IS THE WHOLE ASSERTION. It is
in no string in this source, so it can only have come from `say`. A
parser that compiled `say` as `print` emits `[saidprinted]`; one that
appended a newline to both emits a third line. The brackets make the
trailing byte visible -- without them `[said\\nprinted]` and
`[said\\nprinted\\n]` would differ only in bytes the format cannot show.

`$ENV{X} // "said"` keeps the argument a RUNTIME value so nothing
folds; `X` is unset when the runner executes the case. `use feature
"say"` rather than `use v5.36`, which would also enable it: measured,
the version bundle sets `strict` as well -- ops carry `/STRICT`, the
nextstate flags read `fea=6` against `fea=15` -- so it turns on more
than this case is about.

THE TOKEN FACT CATCHES THE REWRITE, because behaviour cannot. A parser
that read the featureless spelling -- rewriting `say $out $line` into
`$out->say($line)`, which is what perl itself does without the pragma
-- produces an ARROW, and this source has none. The positive count says
the construct survived the lex as ONE word: it counts one and not two
even though `say` appears twice, because the other occurrence is inside
`use feature "say"` where it is a STRING and not a bare word.
"""),
    ('08_select.t', '`select` is two operators sharing a name, told apart by arity', """
One argument selects a default output handle; four arguments are the
`select(2)` syscall. Measured, they are not one operator with an
optional tail:

    select($out)                    rv2gv sK/DREFSV,1  select[t7] sK/1
    select(undef, undef, undef, 0)  undef undef undef const  sselect[t11] sK/4

A different op NAME, not a different flag. That is the inverse of
`rv2gv`, which does three unrelated jobs in this tier under one name.

So this is a PARSING fact. A parser must count the arguments before it
knows which operator it has read, and the arity is not recoverable from
the name, from the first argument, or from anything a lexer can see.

THERE IS NO TWO-ARGUMENT FORM. Measured, arity 2 and arity 3 are both
COMPILE-TIME errors carrying the same message -- `Not enough arguments
for select system call` -- which is perl deciding that anything past
one argument must be the four-argument form and then finding it short.
The wrong count never reaches runtime. Arity 0 is the one-argument
operator with nothing selected: measured, `select()` returns
`main::STDOUT`, the same value `select(STDOUT)` returns, which makes
the split 0-or-1 against 4. It is not written into the source because
it would add a third `select` to a case whose subject is that there are
two.

THE ONE-ARGUMENT FORM RETURNS THE PACKAGE-QUALIFIED PREVIOUS HANDLE --
measured, `main::STDOUT` and not `STDOUT` -- which is what makes the
round trip work: the value `select` hands back is a thing `select` will
take, so restoring needs no name of its own.

THE OUTPUT IS THE FALSIFYING HALF, and the `print` with NO HANDLE is
what makes it one. Line 3 is a bare `print "captured\\n"` with no
filehandle anywhere in the statement, and its bytes land in `$buf`
because line 2 changed where "no handle" points. A parser that compiled
`select` as a no-op prints `captured` to stdout and then `[][0]`: the
right bytes in the wrong stream, and an empty buffer where the capture
should be. Nothing about `print` itself changed -- it emits the same
`print vK` either way -- so this case measures `select` entirely
through what `print` DID NOT DO.

The `0` is the four-argument form's return, the number of handles
ready, zero because every handle argument is `undef`. A TIMEOUT OF ZERO
is what makes it usable: `select(undef,undef,undef,0.25)` is the
idiomatic sub-second sleep and would be the obvious thing to write, but
a corpus case cannot rest on a duration. With the timeout at 0 the call
returns immediately and deterministically; measured twice in
succession, both runs printed `[captured\\n][0]`.

NO TOKEN FACTS. `select` is a bare word in both spellings and the arity
split is a parsing question, not a lexical one -- the token stream is
`word(select) operator(() ...` for both, identically, and the
glossary's vocabulary admits `one` and `no` and nothing that would say
"four". This is the mirror of the scalar readline case: there the
behaviour could not see the split and the tokens could.
"""),
])

covered |= topic(TIER, 'symbol-table.md', 'Globs and the symbol table', """
`*name` writes a name, `\\*name` reads one, and `AUTOLOAD` is what
happens when the lookup finds nothing.

**Tier 10 io.** Introduces `close`, `eof`, `open`, `readline`, `rv2gv`,
`say`, `select`, `sselect`. Depends on 03_context.

THESE SIT IN 10_io RATHER THAN 08_references BECAUSE OF WHAT THEY EMIT.
Measured, `*alias = sub {...}` emits `rv2gv`, which this tier claims for
its filehandles -- a glob and a filehandle are the same thing to perl,
which is the whole reason `open(my $fh, ...)` and `*STDOUT` live in one
namespace. Tier 08 owns references and could not have them. The three
cases reach one symbol table from three sides: one installs a name, one
references a name, and one fails to find a name.

The typeglob did not appear in this corpus at all before these cases.
Measured across 1,141 lines of source, `*` occurred three times and
every one was a `sprintf` width specifier -- `"%0*d"`, `"%*d"`,
`"%-*d"`. Zero typeglobs, so nothing here could tell a lexer that
treats `*` as always multiplication from one that gets it right.
""", [
    ('09_glob_assign.t', '`*` is a sigil and an operator, resolved by position', """
`*name` introduces a fourth namespace using the same character as
multiplication. The fork is a one-character lexical one and it is
resolved by POSITION -- a sigil where a term is expected, an operator
where one just ended, which is the expect-state mechanism the spec
describes for `%`, `<`, `&` and `/`.

BOTH READINGS APPEAR IN ONE STATEMENT PAIR. The sigil installs `sq`
into the symbol table; the two multiplications inside and after it are
ordinary arithmetic. A lexer that resolved `*` by looking only at the
character produces either a syntax error or a multiplication where the
installation belongs, and `9 6` is unreachable either way.

The installed sub is called by NAME on the next line, which is what
makes the installation observable: a parser could accept the assignment
and do nothing, and `sq($n)` would then be a call to an undefined sub.

THE TOKEN FACT CANNOT BE ABOUT THE SIGIL, and that is itself the
measurement. Our lexer emits `Operator("*")` for all three occurrences
-- the glob sigil and both multiplications -- so no count of `*`
separates them and a fact naming it would be a claim about the source's
punctuation. The glossary has no typeglob category to assert instead.
What the case can claim is the shape around it: `sub` appears exactly
once, in the anonymous constructor.
"""),
    ('10_glob_ref.t', '`\\*STDOUT` is a reference to a GLOB, its own type', """
Not SCALAR, not CODE, not the filehandle it names. Measured, `ref(\\*STDOUT)`
is `GLOB` -- a type this corpus never named before, since tier 08
covers `SCALAR`, `ARRAY`, `HASH` and `CODE` and stops there.

THE BACKSLASH IS THE PART A PARSER CAN GET WRONG. `\\*STDOUT` is a
reference-to-glob; `*STDOUT` alone is the glob itself, and `\\*` is not a
compound operator -- it is tier 08's `\\` applied to a term that happens
to start with `*`. A lexer that read `\\*` as one token, or that read
`*STDOUT` as multiplication by a bareword, produces something `ref`
would not call `GLOB`.

`STDOUT` rather than a glob this case creates, because a bareword
filehandle is the one glob guaranteed to exist without installing
anything -- and it is this tier's own subject.
"""),
    ('11_autoload.t', '`AUTOLOAD` catches a call to a sub that does not exist', """
A sub name special to the LANGUAGE rather than to the program: perl
calls it when a named sub cannot be found. The same symbol table seen
from the other side -- here the lookup fails and perl falls back.

`missing` is never defined. The call finds nothing, perl dispatches to
`AUTOLOAD`, and `$AUTOLOAD` holds the fully qualified name that was
sought -- `main::missing`, which the substitution trims to `missing`.

WHAT A PARSER GETS WRONG HERE IS THE CALL ITSELF. `missing()` has no
declaration anywhere in the source, so a parser that resolves calls at
parse time has nothing to resolve; one that treats an unknown bareword
followed by parens as a call gets it right and defers the question to
runtime, which is what perl does. `07_subroutines/10_undeclared_callee.t`
makes the same claim for an ordinary sub; the difference here is that
the sub genuinely does not exist and the program still works.

THE TOKEN FACT COUNTS ONE `AUTOLOAD` IN A SOURCE THAT SPELLS IT THREE
TIMES, and that is the claim rather than an oversight. Two of the three
are `$AUTOLOAD`, a Variable, and only the sub name is a bare Word. The
fact separates the special sub name from the special variable that
carries its argument -- different things wearing one spelling. A lexer
that let the sigil fall off, or that read the bare name as a variable,
fails it.

The `s///` is tier 09's op, reached rather than introduced: `$AUTOLOAD`
arrives package-qualified and the unqualified name is what makes the
output readable as a claim.
"""),
])

covered |= topic(TIER, 'adjacency-10_io.md', 'Every construct, each beside another', """
One body holding every construct this tier introduces, each adjacent to
another.

**Tier 10 io.** Introduces nothing of its own; the mixture is the
subject. Depends on 03_context.

The tier's other cases are one construct each, which is what makes them
diagnosable. That same property is why such a corpus cannot reach an
ADJACENCY bug: a parser handling every construct alone and mishandling
a pair goes green over the pair.
""", [
    ('00_adjacency.t', 'The whole tier in one body', """
The pairs that matter are the ones a one-construct case cannot make.

A scalar readline followed by a LIST readline on the SAME handle is the
whole of this tier's context dependency in two adjacent statements --
the second reads what the first left, so the printed `2` is only
correct if both contexts were resolved and resolved DIFFERENTLY. The
two readline cases each open a fresh handle and so can never disagree
about position.

The second pairing is `eof` inside a print LIST that also holds an
array, which puts the tier's own op in an argument position rather than
alone in a statement.

THE THIRD PAIRING IS `select` WITH `say`, and it is the one that needs
two constructs most. `say "round trip"` names NO HANDLE, and its bytes
land in `$buf` rather than on stdout, because the `select($out)` above
it changed where "no handle" points. Neither case alone can make that
claim: the `say` case passes its handle explicitly and the `select`
case redirects a `print`. Here the two constructs are the same
assertion -- a parser that dropped either one puts `round trip` on
stdout and leaves the buffer empty.

The four-argument `select` follows, the other operator sharing that
name: measured, `select($out)` emits `select` and
`select(undef,undef,undef,0)` emits `sselect`, a different op reached
by a different argument count. Both spellings are here because the
tier's declared set holds both and an adjacency case must reach every
op the tier introduces.

The pairing with 03_context is the scalar-versus-list readline itself:
context is not a separate construct to place beside this one, it is the
thing selecting which readline happens.

The ops this reaches beyond the tier's own, all claimed earlier and
none new: `gv`, `padav` and `aassign` (02), `cond_expr` and `goto` from
the ternary (06), `undef` from the syscall arguments (04), `srefgen`
from `\\my $buf` (08).

`use feature "say"` is required and is not decoration: without it `say`
is not this tier's op at all but a method call. Measured, the pragma
changes no op in this body beyond enabling `say` itself.
"""),
])

check(TIER, covered)
