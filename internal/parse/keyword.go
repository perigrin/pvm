// ABOUTME: The one keyword shape CORE.pmt cannot yet derive: split's, a list operator.
// ABOUTME: Every other builtin's shape is derived from its CORE.pmt line (coreShapes).

package parse

// listOperator holds split, the one builtin CORE.pmt has no line for: it has
// no prototype but typed positional parameters, and typed Perl has no
// spelling for that yet (01a1113c). It goes when CORE.pmt can declare it.
//
// `split` is a list operator, which the keyword sweep missed because perl
// rewrites `split /$x/, $y` into a three-argument form, so it did not
// deparse to the shape the classifier looked for. Verified by hand on 5.42.0:
//
//	perl -MO=Deparse,-p -e 'our ($a,$c); my @z = split /,/, $a < 5, 2;'
//	(my @z = split(/,/, ($a < 5), 2));
var listOperator = map[string]bool{"split": true}
