#!perl
# `$a++` returns the OLD value and `++$a` the new one, and the corpus
# emitted `postinc` only as a side effect of a file about something else.
#
# TIER 04 operators
# INTRODUCES the prefix increment and both decrements
# USES postinc from 07_subroutines
#
# perlop's level 3. Before this file the corpus had no file whose SUBJECT
# was inc/dec: `postinc` appeared in `07_args_alias.t` and an adjacency
# body, both using it to mutate something the file was really about, and
# `preinc`, `postdec` and `predec` appeared nowhere at all.
#
# So nothing claimed the return value, which is the whole distinction
# between the two fixings.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $a = 5; my $b = $a++; my $c = ++$a; print "$a $b $c"'
#   7 5 7
#
# `$b` is 5 -- the value BEFORE the increment -- and `$c` is 7, the value
# after. `$a` ends at 7 either way, so a parser that confused the two
# fixings gets the variable right and the captured values wrong. Only
# binding both to separate variables shows it.
#
# THE OPS DISTINGUISH THEM TOO, which is unusual for this tier: `postinc`
# and `preinc` are separate ops rather than one op with a flag. That is
# why this file INTRODUCES three of the four -- `postinc` is already
# tier 07's, claimed there before any file made it a subject.
#
# THE TOKEN FACT IS A NEGATIVE because counting cannot work here: the
# source spells `++` twice and `--` twice, so any count is a claim about
# the file's shape. What it rules out is the mis-lex -- `$a++ + 1` would
# produce a `+++` run if the lexer were greedy about the third character,
# and that spelling appears nowhere in this source.

--- source
my $a = $ENV{X} // 5;
my $b = $a++;
my $c = ++$a;
my $d = $a--;
my $e = --$a;
print "$a $b $c $d $e\n";

--- expect output
5 5 7 7 5

--- expect parses

--- expect tokens
no operator whose text is "+++"
