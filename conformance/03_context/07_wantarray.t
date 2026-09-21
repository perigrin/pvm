#!perl
# `wantarray` asks which context the current code was called in. At file
# scope the answer is undef, and asking it costs one op and nothing else.
#
# TIER 03 context
# INTRODUCES wantarray
# USES nothing from a later tier
#
# `wantarray` is the only op in perl whose entire meaning is this tier:
# it returns true in list context, false in scalar context, and undef in
# void context. Everything else here shows context from the outside;
# this shows it from the inside.
#
# The tier README previously warned that reaching it meant reaching tier
# 07 -- that `wantarray` lives inside a sub and so drags in `entersub`,
# `leavesub` and `gv`. Measured, that is false. A sub is where `wantarray`
# is USEFUL, not where it is legal: at file scope it compiles to a bare
# `wantarray s` with no call machinery at all, and returns undef because
# a file body is void context.
#
# Observing the result is the part that needs care. `defined $w ? ... : ...`
# emits `cond_expr`, which is tier 06's; `if (defined $w)` the same. So
# the file counts instead: `my @seen = ($w)` holds exactly one element
# whether `$w` is undef or not, and printing that count is an assertion
# about shape that needs no branch. That `1` says the op ran and produced
# a value; that the value is undef is what the comment records and what
# the one-liner below shows.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $w = wantarray; print defined($w) ? "defined\n" : "undef\n"'
#   undef
#
#   $ perl -e 'my $w = wantarray; my @seen = ($w); print scalar(@seen), "\n"'
#   1

--- source
my $w = wantarray;
my @seen = ($w);
print scalar(@seen), "\n";

--- expect output
1

--- expect parses

--- expect tokens
one word whose text is "wantarray"
