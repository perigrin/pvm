#!perl
# `grep` has the same BLOCK-or-EXPRESSION fork as `map`, and the same
# invisibility at the optree. This file too asserts in TOKEN FACTS.
#
# TIER 03 context
# INTRODUCES grepstart grepwhile
# USES nothing from a later tier
#
# THE CENTRAL FACT is `09_map.t`'s, measured again for the other keyword
# and holding: `my @b = grep { $_ } @a` and `my @b = grep($_, @a)` emit
# the same ops, in the same order, with the same flags --
#
#   pushmark, pushmark, padav[@a] lM, grepstart lK, grepwhile lK, gvsv[*_]
#
# The syntactic difference between the two forms is a COMMA, which the
# optree does not record, so the token stream is the only place the fork
# is observable. The file is written to hold exactly one comma -- `qw()`
# builds the list without any -- and the token fact below counts it.
#
# What is not identical is not the ops either: as with `map`, the block
# form advances each lexical's COP SEQUENCE RANGE by 2, because a block
# is a scope. That is pad metadata, not an op, and the op-name lint does
# not read it.
#
# WHERE grep IS SHARPER THAN map, and it is the reason both files exist
# rather than one. `map` in scalar context returns the count of what it
# would have returned in list context, so its two halves report the same
# NUMBER in different shapes -- 3 elements, or the number 3. `grep`
# FILTERS, so the count is not the input's: from three elements it keeps
# two, and the scalar half reports 2 while the list half produces the two
# surviving strings. Nothing here folds; the discrimination is between a
# count and the things counted, and they have different arities from the
# input.
#
# `$a[0]` is set from `$ENV{G}` and not from a literal for two reasons at
# once. A CONSTANT argument erases the construct -- a wholly-constant
# grep folds and leaves no `grepstart` to measure. And grep needs an
# element that is FALSE, or it filters nothing and the two contexts stop
# disagreeing about arity. `$ENV{G}` is unset when the runner executes,
# so the element is undef, which is both opaque at compile time and false
# at run time.
#
# Written as a bare `$a[0] = $ENV{G}` rather than `"$ENV{G}"`, which is
# what `09_map.t` uses. Measured, the interpolating form of a lone
# variable emits `stringify`, and `stringify` is claimed by NO tier -- so
# the quotes that read as harmless would fail the corpus lint for an op
# the READMEs do not yet describe. The plain assignment emits
# `aelemfastlex_store` and `multideref`, both tier 02's.
#
# No comparison appears in the block on purpose. `grep { $_ gt "a" } @a`
# is the obvious demonstration and emits `gt`, which is tier 04's, one
# tier forward. A bare `$_` tests the element's own truth, which is the
# smallest predicate there is.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = qw(x y z);
#              $a[0] = $ENV{G};
#              my @block = grep { $_ } @a;
#              my @expr  = grep($_, @a);
#              my $n = grep { $_ } @a;
#              print "@block @expr $n\n";'
#   y z y z 2
#
#   $ perl -MO=Concise,-exec <that program> | grep grepstart
#   h  <@> grepstart lK
#   r  <@> grepstart lK
#   10 <@> grepstart sK

--- source
my @a = qw(x y z);
$a[0] = $ENV{G};
my @block = grep { $_ } @a;
my @expr  = grep($_, @a);
my $n = grep { $_ } @a;
print "@block @expr $n\n";

--- expect output
y z y z 2

--- expect parses

--- expect tokens
one operator whose text is ","
one variable whose text is "@block"
one variable whose text is "@expr"
