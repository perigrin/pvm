#!perl
# `select` is TWO OPERATORS SHARING A NAME, and the ARITY is what decides
# which one. One argument selects a default output handle; four arguments
# are the `select(2)` syscall. They are not one operator with an optional
# tail -- perl emits DIFFERENT OPS for them.
#
# TIER 10 io
# INTRODUCES select and sselect
# USES my, open, close, print, undef
# MEASURED perl 5.42.0
#
# THE ARITY SPLIT, measured. Both spellings are in this file's source,
# four lines apart:
#
#   select($out)                     rv2gv sK/DREFSV,1  select[t7] sK/1
#   select(undef, undef, undef, 0)   undef undef undef const  sselect[t11] sK/4
#
# `select` against `sselect` -- a different op name, not a different flag
# on one op. That is a stronger split than anything else in this tier:
# `rv2gv` does THREE jobs under one name here (autovivify at open,
# resolve at print, inspect at eof) and the op table records one op for
# all three. `select` goes the other way -- one spelling, two ops -- and
# it is the only construct in this tier that does.
#
# So this is a PARSING fact and not a runtime one. A parser must count
# the arguments before it knows which operator it has read, and the
# arity is not recoverable from the name, from the first argument, or
# from anything a lexer can see. `07_subroutines`'s finding is that a
# call site does not carry what decides its parse; here the call site
# carries it, but only in the count.
#
# WHY NOT A TWO-ARGUMENT FORM. There is none. Measured, arity 2 and
# arity 3 are both COMPILE-TIME errors, and the message is the same one:
#
#   $ perl -c -e 'select(STDOUT, STDERR)'
#   Not enough arguments for select system call at -e line 1, near "STDERR)"
#   -e had compilation errors.
#
#   $ perl -c -e 'select(undef,undef,undef)'
#   Not enough arguments for select system call at -e line 1, near "undef)"
#   -e had compilation errors.
#
# The message names the SYSCALL in both, which is perl deciding that
# anything past one argument must be the four-argument form and then
# finding it short. So the arity is resolved at COMPILE time and the
# wrong count does not reach runtime: the two operators are not
# neighbours on a spectrum, they are two entries in the keyword table
# reached by different counts, and 2 and 3 reach neither.
#
# ARITY 0 IS THE ONE-ARGUMENT OPERATOR WITH NOTHING SELECTED. Measured,
# `select()` returns the current handle without changing it:
#
#   $ perl -e 'my $o = select(); print "[$o]\n"'
#   [main::STDOUT]
#
# Same value as `select(STDOUT)` returns below, which is the reading that
# makes the split 0-or-1 against 4 rather than 1 against 4. It is
# measured here and not written into the source: it would add a third
# `select` to a file whose subject is that there are two.
#
# THE ONE-ARGUMENT FORM RETURNS THE PACKAGE-QUALIFIED PREVIOUS HANDLE.
# Measured:
#
#   $ perl -e 'my $o = select(STDOUT); print "[$o]"'
#   [main::STDOUT]
#
# `main::STDOUT` and not `STDOUT`, which is what makes the round trip in
# this file work: the value `select` hands back is a thing `select` will
# take, so restoring the previous handle needs no name of its own.
#
# THE OUTPUT IS THE FALSIFYING HALF, and the `print` with NO HANDLE is
# what makes it one. Line 3 of the source is a bare `print "captured\n"`
# -- no filehandle anywhere in the statement -- and its bytes land in
# `$buf` rather than on stdout, because line 2 changed where "no handle"
# means. Measured:
#
#   $ perl conformance/10_io/08_select.t
#   [captured
#   ][0]
#
# A parser that compiled `select` as a no-op prints `captured` to stdout
# and then `[][0]`: the right bytes in the wrong stream, and an empty
# buffer where the capture should be. That is the assertion. Nothing
# about `print` itself changed -- it is tier 01's op, emitting the same
# `print vK` either way -- so this file measures `select` entirely
# through what `print` DID NOT DO.
#
# The `0` is the four-argument form's return: the number of handles ready,
# which is zero because every handle argument is `undef`. A TIMEOUT OF
# ZERO is what makes it usable here. `select(undef,undef,undef,0.25)` is
# the idiomatic sub-second sleep and would be the obvious thing to write,
# but a corpus file cannot rest on a duration -- the return is still 0 and
# the wall clock is not this tier's subject. With the timeout at 0 the
# call returns immediately and deterministically; measured twice in
# succession, both runs printed `[captured\n][0]`.
#
# `undef` is tier 04's op, claimed there, and reaching it as a syscall
# argument is a use rather than an introduction. `srefgen` from
# `\my $buf` is tier 08's on the same terms, as `00_adjacency.t` records.
#
# `or die` is deliberately absent on the open, for the reason
# `01_open_close.t` gives: no tier claims `die`, so a check on the return
# value would widen this tier's declared set by an op belonging to a tier
# that does not exist yet.
#
# NO TOKEN FACTS. `select` is a bare word in both spellings and the
# arity split is a PARSING question, not a lexical one -- the token
# stream is `word(select) operator(() ...` for both, identically, and no
# fact in the glossary's vocabulary can count arguments. The grammar
# admits `one` and `no` and nothing that would say "four". This is the
# mirror of `02_readline_scalar.t`'s situation: there the behaviour could
# not see the split and the tokens could, and here the tokens cannot see
# the split and the behaviour can. Both halves of this file's claim are
# made by the output above.
#
# `expect output` is written before `expect parses` rather than last: the
# blank line after it carries the output's trailing newline, and a blank
# line at END of file is what `end-of-file-fixer` strips.

--- source
open(my $out, ">", \my $buf);
my $prev = select($out);
print "captured\n";
select($prev);
close($out);
my $ready = select(undef, undef, undef, 0);
print "[$buf][$ready]\n";

--- expect output
[captured
][0]

--- expect parses
