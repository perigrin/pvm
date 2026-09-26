# Pod blocks and data sections

Two regions the lexer eats whole and the compiler never sees. A pod block
emits NOTHING -- not an op, not a `nextstate`; the line numbers in the
surrounding `nextstate`s move and that is the only trace. `__END__` and
`__DATA__` emit nothing either: reading the section emits `readline` and
`gv`, which describe the read and not the marker.

**Tier 13 opaque.** Introduces `backtick`, `enterwrite`, `glob`, `pack`,
`unpack`. Depends on 10_io. The dependency is here and it is specific:
the only way to observe that a data section exists is to read the `DATA`
filehandle it creates, and that is tier 10's `readline`.

Each construct appears twice, and the pairing is the tier's thesis split
in half. "Delimit without lexing the contents" is a claim about EXTENT
and a claim about CONTENTS, and a region whose contents are inert makes
only the first -- prose and two bare words are delimited the same way
whether or not they were lexed. The second case of each pair puts
hostile bytes inside the region so the two readings can be told apart.

All four cases PASS. The data cases refused with `not_a_term` until the
parser agreed with the lexer that a section is trivia: the lexer had
always ended the program text at the marker and read `<DATA>` as one
angle-bracket term, and the parser had been handing the section token to
the expression parser, which has no term for it.

Both data sections are read in LIST context deliberately. The usual
spelling, `while (my $line = <DATA>)`, emits `defined` -- perl inserts it
around the readline so an empty last line does not end the loop -- and
`defined` belongs to no tier yet.

Every body here is written clear of any line beginning `--- `, which the
corpus format reads as a section marker with no awareness of Perl's
nesting. Tier 13 is the first tier able to hit that.

## A pod block compiles to nothing

Measured, this source and the same source with the pod deleted emit the
IDENTICAL op sequence -- const, padsv_store, pushmark, padsv,
multiconcat, print -- differing only in the line numbers recorded on the
nextstates. So the pinned output cannot fail for the reason this case
exists: a lexer that read the pod as prose, as code, or as nothing at all
would print `1` either way.

That is why this case asserts the pod token. It is the whole assertion;
the rest is scaffolding proving the program still ran.

The block runs from a `=` in column 1 followed by an identifier through
the `=cut` line, and the TERMINATOR IS PART OF THE TOKEN, for the same
reason a heredoc body's is: every byte belongs to the block.

```perl
my $x = 1;

=pod

Prose perl never compiles.

=cut

print "$x\n";
```

```behavior
parses: yes
```

```output
1
```

```tokens
one pod block whose text is "=pod\n\nProse perl never compiles.\n\n=cut\n"
```

## A pod block holding text that is not a program

The block holds an unclosed `(`, a `$this` that is not a declared
variable and two stray semicolons. Measured, perl compiles and runs the
program without noticing any of it: the pod is eaten by the lexer before
the parser exists. Measured against our lexer, the same bytes arrive as
ONE pod token with nothing lexed inside.

`=head1` rather than `=pod`, because the block's opener is any `=` in
column 1 followed by an identifier and the case above already covers
`=pod`.

This case PASSES, in a tier where eight of the fifteen files refuse, and
that is the useful shape: the pod block is the construct this tier gets
entirely right, opaque contents and all, so the tier's refusals cannot be
read as "the lexer cannot handle opaque regions".

```perl
my $x = 1;

=head1 NOT PERL

$this is not( a variable; and this ; is not a statement

=cut

print "$x\n";
```

```behavior
parses: yes
```

```output
1
```

```tokens
one pod block whose text is "=head1 NOT PERL\n\n$this is not( a variable; and this ; is not a statement\n\n=cut\n"
```

## A data section read back through `DATA`

`__DATA__` ends the program text and opens a filehandle onto what
follows. This is the one construct in the tier with an observable
consequence, and the consequence is not the marker's: `my @lines =
<DATA>` emits `gv`, `readline`, `padav` and `aassign` -- tiers 02 and 10
-- and would emit exactly those for a handle opened with `open`. Nothing
in the op stream says `__DATA__` created the handle.

Measured, the two statements before the marker parse cleanly and the
section is trivia, the same route a pod block takes. The glob case passes
with the same token category in the same position, so the gap was never
the angle brackets -- the SECTION was the discriminator, and `<DATA>`
without one always parsed.

THE SECTION TOKEN CARRIES ONE TRAILING NEWLINE, NOT TWO. It runs from the
marker to end of file, and end of file is where the ```perl block ends:
the reader gives a case's program exactly one trailing newline. So the
body here is `one\ntwo\n` and the token is `"__DATA__\none\ntwo\n"`.

That boundary is worth stating because a trailing blank line inside a
section is not nothing. Measured, a section written `one\ntwo\n\n` makes
perl's `DATA` yield a third, empty line -- inside a data section trailing
whitespace is DATA, where anywhere else in a Perl program it is
whitespace. There simply is no such line to carry here, and the pinned
output agrees: two lines, no third.

`__END__` is the same construct under another name: measured, both end
the program text and both open DATA in package main. The adjacency case
uses the other spelling.

```perl
my @lines = <DATA>;
print @lines;
__DATA__
one
two
```

```behavior
parses: yes
```

```output
one
two
```

```tokens
one readline operator whose text is "<DATA>"
one data section whose text is "__DATA__\none\ntwo\n"
```

## A data section holding bytes that are not a program

The section holds `sub not_compiled { $x <=> }`, which is a syntax error
as Perl -- `<=>` with no right operand -- and `q{ unbalanced`, whose
quote never closes. Measured, perl compiles the program and prints both
lines back: neither is a program, and neither was ever offered to the
parser. Measured against our lexer, the whole section arrives as ONE data
section token with nothing lexed inside.

The unbalanced `q{` is the sharper case for the lexer/parser split: a
lexer that had read the section would still be inside an unterminated
quote at end of file.

The section token ends with one newline, per the case above: it runs to
end of file, and the program gets exactly one. The second line is not
newline-terminated in the section's own text either -- `q{ unbalanced\n`
is the last line and its newline is the file's.

```perl
my @lines = <DATA>;
print @lines;
__DATA__
sub not_compiled { $x <=> }
q{ unbalanced
```

```behavior
parses: yes
```

```output
sub not_compiled { $x <=> }
q{ unbalanced
```

```tokens
one readline operator whose text is "<DATA>"
one data section whose text is "__DATA__\nsub not_compiled { $x <=> }\nq{ unbalanced\n"
```
