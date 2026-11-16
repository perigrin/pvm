// ABOUTME: A library operator's level to the PLUGIN_* terminal toke.c's tokentype_for_plugop gives it (ruling A).
// ABOUTME: Perl's buckets: from its declared class where that is one of perl's, else by perl's <= thresholds over recursive descent's placement.

package grammar

import "tamarou.com/pvm/internal/parse"

// plugopLevels are perl.h's INFIX_PREC_* levels in order (perl.h:6117-6127 at 94e50866), each with the terminal
// tokentype_for_plugop gives a precedence at or below it and above the one before (toke.c:9264-9288). class is the
// level's name, which :infix(CLASS) gives it; anchor the core operator perl.h places it as, whose binding power in
// recursive descent stands for the threshold (none for LOW, below every core infix, or HIGH, above).
var plugopLevels = []struct{ class, anchor, terminal string }{
	{"LOW", "", "PLUGIN_LOW_OP"},                            // 10
	{"LOGICAL_OR_LOW", "or", "PLUGIN_LOGICAL_OR_LOW_OP"},    // 30, as or
	{"LOGICAL_AND_LOW", "and", "PLUGIN_LOGICAL_AND_LOW_OP"}, // 40, as and
	{"ASSIGN", "=", "PLUGIN_ASSIGN_OP"},                     // 50, as =
	{"LOGICAL_OR", "||", "PLUGIN_LOGICAL_OR_OP"},            // 70, as ||
	{"LOGICAL_AND", "&&", "PLUGIN_LOGICAL_AND_OP"},          // 80, as &&
	{"REL", "==", "PLUGIN_REL_OP"},                          // 90, just below ==
	{"ADD", "+", "PLUGIN_ADD_OP"},                           // 110, as +
	{"MUL", "*", "PLUGIN_MUL_OP"},                           // 130, as *
	{"POW", "**", "PLUGIN_POW_OP"},                          // 150, as **
	{"HIGH", "", "PLUGIN_HIGH_OP"},                          // 170
}

// LibraryTerminal is the terminal perl reads op as: its class's where that is one of perl's, otherwise the first level
// whose anchor recursive descent places at or above op, by tokentype_for_plugop's <=, and HIGH above them all. An
// operator placed strictly between two anchors takes the upper one's bucket, where recursive descent groups it between:
// the parity limit.
// shortcut: a class XS::Parse::Infix names but perl.h does not (EQUALITY, RELATION, ...) is placed by its anchor's
// level too, which puts RELATION, ISA and MATCHRE above REL; upgrade if perl is shown to read one at REL.
func LibraryTerminal(op parse.LibraryOperator) string {
	for _, l := range plugopLevels {
		if op.Class == l.class {
			return l.terminal
		}
	}
	for _, l := range plugopLevels {
		if info, ok := parse.InfixInfo(l.anchor); ok && op.Info.BP <= info.BP {
			return l.terminal
		}
	}
	return "PLUGIN_HIGH_OP"
}
