# Every opaque region, each beside another

One body holding every construct this tier introduces, each adjacent to
another.

**Tier 13 opaque.** Introduces nothing of its own; it is the mixture that
is the subject. Depends on 10_io, and the dependency is live here: the
`__END__` section is read through the DATA filehandle, which is tier 10's
`readline`.

ADJACENCY MATTERS MORE HERE THAN ANYWHERE EARLIER, because every
construct in this tier CHANGES THE LEXER'S MODE and then has to hand it
back. A format body, a pod block and a heredoc body each suspend ordinary
lexing, consume lines by a rule of their own, and resume. A case holding
one of them tests that the mode was entered and left once. This one tests
that leaving one mode leaves the lexer able to enter the next.

THE ORDER IS NOT ARBITRARY. The format body sits FIRST, before any
statement, because its picture lines are the region least like Perl -- if
resuming after them is wrong, everything downstream is wrong for a reason
the single-construct cases cannot name. The heredoc body and the pod
block follow, so the heredoc's line-oriented terminator is immediately
followed by another line-oriented terminator of a different kind. The
data section is last because it has to be: it ends the program text, so
nothing can follow it.

A HAZARD THIS CASE HAD TO AVOID, and it is a real limitation of the
corpus format rather than a quirk of this body: `--- ` at the start of a
line is recognised as a SECTION MARKER, with no awareness of Perl's own
nesting. A heredoc body, a pod block or a data section whose content held
such a line would be cut in two. The bodies here are written clear of it.
Tier 13 is the first tier able to hit this, because it is the first with
multi-line opaque regions.

## The whole tier in one body

THREE CODES, ONE NAMED. Measured, the three Unknowns carry
`unimplemented_statement` (the format declaration), `trailing_tokens`
(the heredoc body) and `not_a_term` (the data section) -- the same three
the format, heredoc and data-section cases carry alone, and nothing else.
`hasCode` is membership, so naming any one is a true claim; the middle
one is named because it is the only one whose SPAN says something the
single-construct cases cannot.

THAT SPAN IS THE NEGATIVE THIS CASE EXISTS FOR. Alone, the heredoc's
`trailing_tokens` Unknown covers the body and the statement after it.
Here it covers the body, the pod block, AND the statement after those --
measured, from `heredoc line` through `my @none = <*.nonexistent-xyz>;`.
The pod block sits inside an Unknown and contributes none of its own,
which is the adjacency claim stated as a measurement: the lexer left the
heredoc's opaque region and entered the pod block's cleanly, and the
parser's failure to resume did not compound into a second failure at the
join.

Three Unknowns for three distinct constructs and nothing more. `qx`, the
glob and the pod block pass here as they pass alone. So no Unknown
appears at a JOIN between two constructs, which is the negative worth
having: the lexer leaves each opaque region cleanly and the failures do
not compound.

Every token fact below passes. The lexer handles all six constructs
adjacent to one another; the refusals are the parser's, and the token
section is what locates them there.

The spelling is `__END__` rather than `__DATA__` to cover the other half
of the data-section construct -- measured, both end the program text and
both open DATA in package main.

`write` prints between the two `print`s rather than at the end, so format
output and print output share one ordered stream on STDOUT. The output
ends with a BLANK LINE, and it is the corpus format's own.

```perl
format STDOUT =
a report line
.

my $shell = qx{echo hi};
my $here = <<"EOT";
heredoc line
EOT

=pod

Prose the lexer eats and the optree never sees.

=cut

my @none = <*.nonexistent-xyz>;
my @data = <DATA>;
print $shell, $here, "glob ", scalar(@none), "\n";
write;
print @data;
__END__
from the data section
```

```behavior
parses: yes
refuses: unfiled
refusal: trailing_tokens
```

```output
hi
heredoc line
glob 0
a report line
from the data section
```

```tokens
one format body whose text is "a report line\n.\n"
one heredoc opener whose text is "<<\"EOT\""
one heredoc body whose text is "heredoc line\nEOT\n"
one pod block whose text is "=pod\n\nProse the lexer eats and the optree never sees.\n\n=cut\n"
one quote-like operator whose text is "qx{echo hi}"
one readline operator whose text is "<*.nonexistent-xyz>"
one readline operator whose text is "<DATA>"
one data section whose text is "__END__\nfrom the data section\n\n"
```
