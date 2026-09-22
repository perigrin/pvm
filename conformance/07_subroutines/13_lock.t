#!perl
# `lock` is a named unary that takes a VARIABLE, emits its own op, and --
# absent actual threads -- does nothing observable at all.
#
# TIER 07 subroutines
# INTRODUCES the lock statement
# USES nothing from a later tier
#
# WHY IT IS IN THIS TIER, said plainly: `lock` is not about subroutines.
# It declares no sub, calls no sub, and has no argument protocol. It sits
# here for PROXIMITY to `13_tie_scalar.t` and `14_tied.t`, which are this
# tier's on a real argument, and because its own op budget is so small
# that almost any tier could hold it -- `const padsv padsv_store print
# pushmark` are all tier 01's, so nothing about the ops places it. Given
# a free choice it would sit with the unary operators in `04_operators`.
# It is recorded here rather than argued into a theme it does not fit.
#
# WHAT IT ACTUALLY MEASURES is the parse, which is the part that is real.
# `lock $x` is a named unary whose operand is a variable and whose value
# is never used, in statement position -- the shape a parser is most
# likely to mis-handle by treating the keyword as an ordinary bareword
# function call. Measured, it is not one: it has its own op.
#
#   $ perl -MO=Concise,-exec -e 'my $n = $ENV{N} // 7; lock($n); print "locked ", $n, "\n";'
#   ops: const dor enter lock multideref nextstate padsv padsv_store print pushmark
#
# `lock` is in that stream as an op of its own. A parser that compiled it
# to a subroutine call would emit `entersub` and a `gv` naming a sub that
# does not exist, and the program would die at run time.
#
# THREADS, AND A CLAIM THIS FILE DOES NOT MAKE. It is tempting to say
# `lock` works "without threads". That is not what was measured, and the
# distinction matters. MEASURED perl 5.42.0:
#
#   $ perl -MConfig -e 'print "useithreads=", ($Config{useithreads} // "undef"), "\n"'
#   useithreads=define
#
# The pinned interpreter is a THREADED build -- `x86_64-linux-thread-multi`
# -- so this file establishes that `lock` compiles and runs on a threaded
# perl whether or not `threads.pm` has been loaded, and nothing about an
# unthreaded one:
#
#   $ perl -e 'my $n = 7; lock($n); print "no threads.pm: ok\n"'
#   no threads.pm: ok
#   $ perl -Mthreads -e 'my $n = 7; lock($n); print "with threads.pm: ok\n"'
#   with threads.pm: ok
#
# Without other threads to contend with, the lock is uncontended and the
# statement is a no-op with respect to output. That is precisely why the
# file pins `$n` rather than anything about locking: there IS no
# observable effect of the lock, so the only honest assertion is that the
# statement compiled, ran, and left the program otherwise unchanged.
#
# `$ENV{N}` is unset when the runner executes, so `$n` is 7 by the `//`
# default. The corpus idiom for a runtime value, used here for this
# tier's own reason as well as the general one: a constant operand can
# erase a construct, and a `lock` on a folded constant would be a weaker
# measurement of the operand's arity.
#
# THE TOKEN FACT is the discrimination against the other spelling. There
# is no `unlock` in perl -- a lock is released when its scope exits --
# and a parser inventing one as the obvious counterpart would introduce a
# word this file says is absent, checked against the whole source
# including these comments.

--- source
my $n = $ENV{N} // 7;
lock($n);
print "locked ", $n, "\n";

--- expect output
locked 7

--- expect parses

--- expect tokens
one word whose text is "lock"
no word whose text is "unlock"
