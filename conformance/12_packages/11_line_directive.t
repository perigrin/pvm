#!perl
# A `#` at column zero matching perl's line-directive pattern is NOT a
# comment: it rewrites `__LINE__` and `__FILE__` for everything after it.
#
# TIER 12 packages
# INTRODUCES nothing of its own
# USES nothing from a later tier
# STATUS parses as of `01a0cf3e`, and what refused was never the
# directive. Measured, `#line 200 "bzzzt"` followed by `print 1;`
# parsed all along; it was `print __LINE__, " ", __FILE__` that
# refused, for the filehandle-slot reason
# `10_compile_tokens.t` records.
#
# AND PARSING IS NOT IMPLEMENTING. This file passes and the directive
# is still UNIMPLEMENTED: `scanComment` consumes `#line 200 "bzzzt"`
# like any other `#`, so the lexer skips it as trivia and our
# `__LINE__` and `__FILE__` would report the real position. That is
# precisely the silent failure the paragraph below warns about, and it
# is ours.
#
# The file cannot catch it. `--- expect output` is what PERL prints, so
# it tests perl; `--- expect parses` is satisfied by treating the
# directive as a comment, which is what makes it parse. What would
# catch it is a token fact naming a `line directive` category, and the
# glossary has none -- adding one is `01a0d0b0`.
#
# Recorded here rather than left to be rediscovered: a file that passes
# while its subject is unimplemented is worth exactly one sentence of
# warning, and this is it.
#
# This is a LEXER claim in a tier of compile-time constructs, and it
# belongs with them because it is the same kind of thing: a line that
# changes how the rest of the file is read.
#
# MEASURED perl 5.42.0:
#
#   $ cat lp.pl
#   #line 200 "bzzzt"
#   print __LINE__, " ", __FILE__, "\n";
#   $ perl lp.pl
#   200 bzzzt
#
#   $ perl -e 'print __LINE__, " ", __FILE__, "\n"'   # no directive
#   1 -e
#
# The file's second line reports itself as line 200 of a file called
# `bzzzt`, neither of which is true of the bytes on disk. A lexer that
# treats every `#` as a comment skips the directive and reports `2` and
# the real path -- silently, for every error location in the rest of the
# program.
#
# THE DIRECTIVE IS ALSO WHAT MAKES `__FILE__` PINNABLE AT ALL. Measured,
# an ordinary `__FILE__` reports the path the runner executed, which is a
# temporary file whose name changes every run, so `10_compile_tokens.t`
# leaves it out. Overriding it is the only way a corpus file can claim
# anything about it -- the directive supplies a name that does not depend
# on where the runner put the file.
#
# perlsyn gives the exact pattern in its "Plain Old Comments (Not!)"
# section. The leading `#` must be at column zero, which is why this
# file's directive is unindented while every other line could be.

--- source
#line 200 "bzzzt"
print __LINE__, " ", __FILE__, "\n";

--- expect output
200 bzzzt

--- expect parses

--- expect tokens
one word whose text is "__LINE__"
one word whose text is "__FILE__"
