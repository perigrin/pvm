#!perl
# `__PACKAGE__` is a bareword that is not a bareword: the compiler
# replaces it with a string, and the string depends on where it appears.
#
# TIER 12 packages
# INTRODUCES nothing of its own
# USES the package statement from 01_package_statement.t
# STATUS refuses as of this file. Issue 01a0cc04-33c8-7a0f-ac01-8d9c3ec9df85. Refusal trailing_tokens.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'print __PACKAGE__, "\n"'
#   ... const[PV "main"] ... print ...
#
# A `const`. By the time the optree exists the token is gone and a plain
# string sits in its place -- indistinguishable from `print "main"`. So
# the op lint cannot see this construct, output can see its VALUE, and
# only a token fact can see that the source said `__PACKAGE__` rather
# than the answer.
#
# THE VALUE IS POSITION-DEPENDENT, which is what makes it worth pinning
# rather than asserting:
#
#   $ perl -e 'package Foo; print __PACKAGE__, "\n";
#              package main; print __PACKAGE__, "\n";'
#   Foo
#   main
#
# Two occurrences of identical bytes, two different strings, decided by a
# `package` statement earlier in the file. That is the same shape as this
# corpus's version-gate findings -- a declaration above changing what
# bytes below mean -- with the difference that here BOTH readings are
# correct and neither fails.
#
# `__FILE__` is deliberately absent. Measured, it reports the path the
# runner executed, which is a temporary file whose name changes every
# run, so no output pin could survive. `11_line_directive.t` reaches it
# the only way a corpus file can: by overriding it.
#
# The negative token fact is the load-bearing one. `__PACKAGE__` must
# arrive as a WORD, not as a string literal: a lexer that read the
# double underscores as quoting -- the way it must for `__END__` and
# `__DATA__`, which really do delimit -- would produce a `string
# literal` here and the file would fail. That text appears nowhere else
# in the source, so the negative is a claim rather than an accident.

--- source
package Foo;
print __PACKAGE__, "\n";
package main;
print __PACKAGE__, "\n";

--- expect output
Foo
main

--- expect parses

--- expect tokens
one word whose text is "Foo"
no string literal whose text is "__PACKAGE__"
