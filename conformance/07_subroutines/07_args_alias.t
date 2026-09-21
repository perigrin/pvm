#!perl
# `@_` ALIASES. `$_[0]++` inside a sub increments the caller's variable,
# because the elements of `@_` are not copies of the arguments -- they ARE
# the arguments, the same SVs under different names.
#
# TIER 07 subroutines
# INTRODUCES the aliasing property of @_
# USES nothing from a later tier
#
# This is the file `04_args_array.t` is not. That one READS `$_[0]`, which
# every array in the language supports and which therefore measures nothing
# about `@_` in particular. Writing through it is the whole difference: an
# ordinary array's element is its own storage and a write to it is local,
# and `@_`'s is not.
#
# The contrast is in the source rather than in prose about it. `bump_alias`
# writes through `$_[0]` and the caller sees the change; `bump_copy` takes
# the conventional `my ($n) = @_` copy and writes to that, and the caller
# does not. Same increment, same op, two different variables -- which is
# why `$untouched` stays 1 and is the second half of the pinned output.
#
# MEASURED perl 5.42.0, the behaviour:
#
#   $ perl -e 'sub b { $_[0]++ } my $x = 1; b($x); print "$x\n"'
#   2
#
# MEASURED perl 5.42.0, the two bodies. THE OPS ARE THE SAME, and that is
# this file's finding:
#
#   $ perl -MO=Concise,-exec,bump_alias ...
#   main::bump_alias:
#   1  <;> nextstate(main 2 -:2) v
#   2  <#> aelemfast[*_] sM
#   3  <1> postinc[t2] sK/1
#   4  <1> leavesub[1 ref] K/REFC,1
#
#   $ perl -MO=Concise,-exec,bump_copy ...
#   main::bump_copy:
#   1  <;> nextstate(main 4 -:5) v
#   2  <0> padrange[$n:4,5] */LVINTRO,range=1
#   3  <2> aassign[t4] vKS
#   4  <;> nextstate(main 5 -:6) v
#   5  <0> padsv[$n:4,5] sRM
#   6  <1> postinc[t5] sK/1
#   7  <1> leavesub[1 ref] K/REFC,1
#
# One `postinc` each. The aliasing is not an op and not a flag on one: it
# is a property of what `aelemfast[*_]` fetches, established when
# `entersub` filled `@_` with the caller's SVs rather than with copies of
# them. So no op claim can express it and no optree comparison can find
# it -- only running the program can, which is what `--- expect output`
# does and why this file leans on it entirely.
#
# Both those bodies are also invisible to the dependency lint, which reads
# the main program alone; the tier README's last section is the record of
# that. The main optree here is `aassign`, two `entersub`s and a
# `multiconcat`, all of them tier 04 or earlier except the calls.
#
# `expect output` is written before `expect parses` rather than last,
# because the blank line after it carries the output's own trailing
# newline and a blank line at END of file is what `end-of-file-fixer`
# strips.

--- source
sub bump_alias {
    $_[0]++;
}
sub bump_copy {
    my ($n) = @_;
    $n++;
}
my ($aliased, $untouched) = (1, 1);
bump_alias($aliased);
bump_copy($untouched);
print "$aliased $untouched\n";

--- expect output
2 1

--- expect parses
