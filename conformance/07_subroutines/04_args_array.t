#!perl
# Arguments arrive in `@_`, and a sub reads them as array elements:
# `$_[0]`, `$_[1]`, and `scalar @_` for the count.
#
# TIER 07 subroutines
# INTRODUCES the @_ argument protocol
# USES nothing from a later tier
#
# The construct this file is named for is spelled ENTIRELY in tier 02's
# syntax. `@_` is an array and `$_[0]` is element access; nothing about
# reading arguments needs an op this tier introduces. What makes it tier
# 07's is that `@_` is only populated by a call, which is `entersub` --
# the array exists, but outside a sub it holds nothing a caller put there.
#
# So the ops in the body are `aelemfast`, `rv2av` and friends, all tier
# 02's, and this tier claims none of them. That is recorded in the tier
# README under what is NOT claimed. What the main program contributes here
# is the call that fills the array.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'sub n { scalar @_ } print n(1,2,3), "\n"'
#   3
#
# The body's ops, which the main-program measurement does not reach:
#
#   $ perl -MO=Concise,-exec,n -e 'sub n { scalar @_ } print n(1,2,3), "\n"'
#   main::n:
#   1  <;> nextstate(main 1 -e:1) v:{
#   2  <#> gv[*_] s
#   3  <1> rv2av[t3] lK/1
#   4  <1> av2arylen sK/1
#   5  <1> leavesub[1 ref] K/REFC
#
# That `leavesub` is invisible to a lint reading the main program, which
# is the finding the tier README's last section records.

--- source
sub describe {
    return $_[0] . "-" . $_[1] . "-" . scalar(@_);
}
print describe("a", "b", "c"), "\n";

--- expect output
a-b-3

--- expect parses
