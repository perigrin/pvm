#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another -- plus the deferred form, which appears only here.
#
# TIER 14 recursive
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other files are one construct each, which is what makes them
# diagnosable. That same property is why they cannot reach the bug this
# tier is most exposed to. A lexer that handles re-entry by special-casing
# ONE level -- remember you are inside a replacement, lex to the closing
# delimiter, hand it back -- passes `01_subst_eval.t`, `02_subst_double_eval.t`,
# `03_code_block.t` and `04_qr_with_code.t` individually and still has no
# stack. Four re-entrant constructs in one statement sequence is what asks
# whether the mechanism nests or merely remembers.
#
# The pairing with the earlier tier is with `09_regex`, this tier's
# declared dependency: `qr/c/` on the second-to-last line is tier 09's
# plain compiled pattern, and the `(??{ $i })` beside it is this tier's
# deferred re-entry interpolating that object mid-match. Adjacent, in one
# pattern, which is where a lexer that delimits `(?{` by counting braces
# rather than by parsing would find the extra `?` and stop.
#
# `(??{ })` -- the DEFERRED form -- appears in this file and nowhere else,
# because it introduces no op and no token category the other files do not
# already cover. Measured, `/a(??{ $inner })/` is one `match` op whose
# pattern string holds the block, the same shape as `(?{ })`: the
# difference between the two is WHEN the engine runs the code and what it
# does with the result -- `(?{ })` runs it for effect, `(??{ })` uses its
# return value as a pattern to match right there. That distinction is
# entirely inside the regex engine and invisible to both the optree and
# the token stream, so it earns a place in the adjacency file rather than
# a file of its own.
#
# No `use re 'eval'` and no warnings suppression anywhere here. Measured,
# LITERAL code blocks compile and run clean under `use strict; use
# warnings`; the pragma is required only when the pattern itself is
# interpolated from a variable. `(??{ $i })` interpolates a variable INTO
# the block, not the block into the pattern, which is why it too is exempt.
#
# MEASURED perl 5.42.0, run as written:
#
#   $ perl conformance/14_recursive/00_adjacency.t
#   hit
#   212c 5
#
# The string walks `abc` -> `2bc` (the `/e` replacement evaluates `$n+1`)
# -> `212c` (the `/ee` replacement evaluates `$c` to `3*4`, then evaluates
# THAT to `12`). `$k` is 5 because the `(?{ })` block ran when the engine
# reached it, and `hit` prints because the deferred block returned the
# compiled `qr/c/` and the engine matched it.

--- source
my $s = "abc";
my $n = 1;
$s =~ s/a/$n+1/e;
my $c = q{3*4};
$s =~ s/b/$c/ee;
my $k = 0;
$s =~ /2(?{ $k = 5 })/;
my $i = qr/c/;
print "hit\n" if $s =~ /(??{ $i })/;
print "$s $k\n";

--- expect output
hit
212c 5

--- expect parses

--- expect tokens
one quote-like operator whose text is "s/a/$n+1/e"
one quote-like operator whose text is "s/b/$c/ee"
one quote-like operator whose text is "qr/c/"
no operator whose text is "+"
no numeric literal whose text is "5"
