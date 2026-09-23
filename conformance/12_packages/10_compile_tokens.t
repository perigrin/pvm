#!perl
# `__PACKAGE__` is a bareword that is not a bareword: the compiler
# replaces it with a string, and the string depends on where it appears.
#
# TIER 12 packages
# INTRODUCES nothing of its own
# USES the package statement from 01_package_statement.t
# STATUS parses as of `01a0cf3e`. It refused until then, and the
# refusal was not about the compile phase at all: `print
# __PACKAGE__, "\n"` put the token in `print`'s FILEHANDLE slot,
# because `isBarewordHandle` took any all-caps word and a
# compile-time token is all-caps. The comma after it then had
# nothing to attach to.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'print __PACKAGE__, "\n"'
#   ... const[PV "main"] ... print ...
#
# A `const` carrying `s/TOKEN=PACKAGE`. The VALUE is a plain string by
# the time the optree exists, so the op stream cannot tell this from
# `print "main"` by what it computes -- though the flag does record
# that a token was replaced, which an earlier draft of this header
# denied. What matters for the lint is that the op is `const` either
# way, and `const` is tier 01's. So
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
# literal` here and the file would fail.
#
# THE FACT SPELLS ITS DELIMITERS, and a first draft did not. A Quote
# token's text INCLUDES the quotes -- `internal/lexer/quote.go` takes
# `start` before consuming the opener -- so a fact naming a bare
# `__PACKAGE__` can never match any Quote and is unfalsifiable. Every
# other string-literal fact in the corpus spells them; the draft was
# the only one that did not. `11_oo/10_isa_infix.t` is the same shape
# and writes `"\"Bar\""`.

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
