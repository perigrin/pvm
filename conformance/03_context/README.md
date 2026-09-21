# 03_context

Scalar versus list context, and the pairs that discriminate them.

## Why this tier sits here

Context needs a thing to be in it. Tier 01 supplies literals and tier 02
supplies the aggregates -- `@a`, `%h` -- whose behaviour differs between the
two contexts, and without an aggregate there is nothing to discriminate:
`my $x = 1` is the same program in either context. So this tier cannot come
before 02.

Could it move earlier? No further than immediately after 02, and it cannot
move later either. Tier 04 is the first tier that cannot proceed without
context: `@a + 1` is arithmetic on a count and `"$x" . @a` is not, and the
operator is what decides. An operator tier written without context has to
describe `+` and `.` as taking "values" and then say nothing about which
value an array is -- which is the question. Context sits between the tier
that produces aggregates and the tier that consumes them, and there is no
slack on either side.

## DEPENDS ON

    02_variables

The dependency is on the aggregates, not on element access. `$a[0]` is
already a scalar and tells you nothing; `@a` in two contexts is the whole
subject.

## INTRODUCES

    join list localtime reverse sort wantarray

## Why those ops, and not the ones the source implies

Measured with `perl -MO=Concise,-exec` under 5.42.0. Three of these are
surprises and one absence is the largest fact in the tier.

**`scalar` is not an op.** `my $n = scalar(@a)` and `my $n = @a` compile to
byte-identical optrees: `padav[@a] s` then `padsv_store`. The keyword that
names this tier's subject leaves no trace at all. What distinguishes the two
contexts is the `s` versus `l` flag that B::Concise prints on `padav`, and
`s` versus `l` on the `aassign` in the count idiom `my $c = () = @a`. Context
is a FLAG on an op, not an op.

That is worth saying plainly rather than working around: the op set
under-describes this tier more than any other. A tier lint built only on op
names cannot see the difference between the two halves of every
discriminating pair, because the two halves emit the same op. The ops listed
here are the ops the tier's files emit; they are not a characterisation of
context, and nothing at the op-name level is.

**`reverse`, `sort` and `localtime` are claimed because the pairs need
them.** Each appears twice per file, once with `s` and once with `l`, and
the pair is the point: `my $x = reverse @a` emits `reverse[t4] sK/1` and
`my @b = reverse @a` emits `reverse[t6] lK/1` -- same op, same operand,
different answer. `localtime` is the sharpest of the three because the two
results have unrelated types, a string and a nine-element list.

**`join`, which no reader of the source would predict.** `"@a"` contains no
function call, but it compiles to `gvsv[*"]` -- fetching the list separator
`$"` -- followed by `join[t5] sK/2`. String interpolation of an array IS a
join, decided at compile time. This is the third context after scalar and
list in everything but name, and it is only visible in the optree. The
`gvsv` is tier 02's, claimed there with the package scalar; only `join` is
new here.

**`list`, which survives only where the comma operator is erased.**
`my $last = (4,5,6)` emits `pushmark v`, `const[IV 6]`, `list sKP` -- the
constants `4` and `5` are gone, folded away, because in scalar context the
comma operator discards its left operands. The optimiser erased two-thirds of
the construct the file is about. Same lesson tier 01 records for
`my $x = 1+2`: the ops LINT a declared tier and cannot derive one, and a file
demonstrating a construct may emit fewer ops than the construct has parts.

**Not claimed, though these files emit them.** `padav`, `padhv`, `aassign`
and `gvsv` are tier 02's -- every file here uses them, none introduces them.
`and` and `cond_expr` are what a boolean-context file WOULD emit: `if (@a)`
puts the array in boolean context and emits `padav[@a] s/BOOL`, where the
flag is this tier's and the branch op is tier 06's. That is why no such file
is here -- the flag cannot be reached without the branch, so boolean context
waits for 06. `av2arylen` comes from `$#a`, which is a sigil form and belongs to tier
02 even though it competes with `scalar @a` for the same job.

**`wantarray` is reachable here, and the earlier claim that it was not is
withdrawn.** This README previously said `entersub`, `gv` and `leavesub`
surround `wantarray` and put it out of reach of a tier that stops at 03.
Measured, that is false: a sub is where `wantarray` is USEFUL, not where it
is legal. `perl -MO=Concise,-exec -e 'my $w = wantarray'` emits a bare
`wantarray s` followed by `padsv_store`, with no call machinery at all, and
returns undef because a file body is void context. The op stays claimed and
`07_wantarray.t` reaches it without leaving the tier.

What the sub-free form costs is the obvious way to OBSERVE the answer.
`defined $w ? ... : ...` emits `cond_expr` and `if (defined $w)` emits
`and` -- both tier 06's. So the file counts a one-element array instead,
which asserts that the op ran and produced a value without branching on
what the value is.

As always the list is a UNION across the tier's files, and `00_adjacency.t`
happening to emit all six is a fact about that program rather than a rule
the format requires: `padrange` fuses consecutive `my` declarations and the
comma operator in scalar context erases its own left operands, so a file
demonstrating more constructs can emit fewer ops.

## Four discriminating pairs, which is the number

An op set cannot describe this tier, but it is not true that nothing can.
The one thing the optree does record about context is the FLAG, and a pair
is an op measured with both values of it. Counting those is what turns the
inference gap from a note into a quantity.

Measured, the tier holds four:

| op | scalar half | list half | file |
|---|---|---|---|
| `padav` | `padav[@a] s` | `padav[@a] l` | `01_scalar_of_array.t` |
| `reverse` | `reverse[t4] sK/1` | `reverse[t6] lK/1` | `04_reverse.t` |
| `aassign` | `aassign[t6] sKS` | `aassign[t8] lKPS` | `05_sort.t` |
| `localtime` | `localtime[t5] s` | `localtime[t2] l` | `06_localtime.t` |

Four and not six. `join`, `list` and `wantarray` are introduced here and
have no second half, because neither interpolation nor the scalar-context
comma HAS a list-context form -- `"@a"` is a join whatever receives it, and
in list context the comma is not a `list` op at all -- while `wantarray`
reports the enclosing context rather than being placed in one. A pair is a
property of the ops that answer two questions, and three of the tier's six
answer only one.

`padav` is tier 02's op, and the pair is still this tier's. Using an op is
not introducing it; that `01_scalar_of_array.t` introduces no op of its own
while carrying the tier's baseline pair is the clearest statement of what
this tier is.

All four pairs also appear in `00_adjacency.t`, in one body, which is the
adjacency claim stated in the only terms the optree can check. A file
holding one half of each pair would satisfy every source-coverage check
this tier has and measure none of its subject -- which the adjacency file
did until it was measured.

## Files

| file | construct | new op |
|---|---|---|
| `01_scalar_of_array.t` | `my $n = @a`, `scalar(@a)` | none -- the point |
| `02_interpolated_array.t` | `"@a"` | `join` |
| `03_comma_in_scalar_context.t` | `my $last = (4,5,6)` | `list` |
| `04_reverse.t` | `reverse` in both contexts | `reverse` |
| `05_sort.t` | `sort`, and the `() =` count idiom in both contexts | `sort` |
| `06_localtime.t` | `localtime` in both contexts | `localtime` |
| `07_wantarray.t` | `wantarray` at file scope | `wantarray` |
| `00_adjacency.t` | all of the above, consecutively, each pair in both contexts | none |

`06_localtime.t` pins shape, not time: nine elements in list context, one
element in scalar. Pinning the formatted string would make the file fail
tomorrow and in another timezone, which is the flakiness the corpus cannot
afford.
