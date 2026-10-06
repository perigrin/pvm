// ABOUTME: Tests that CORE.pmt types its user, network, socket, IPC and control-flow builtins.
// ABOUTME: Each type is measured on perl 5.42, and every prototyped line is typed but the prototype-only ones.
package parse_test

import (
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// systemFlowBuiltins is perlfunc's "Fetching user and group info",
// "Fetching network info", "Low-level socket functions", "System V
// interprocess communication functions" and "Keywords related to the
// control flow of your Perl program" among CORE.pmt's prototyped lines,
// without the names the earlier batches hold.
func systemFlowBuiltins(t *testing.T) []string {
	t.Helper()
	others := perlFunctionKinds(t, "String", "Regexp", "Math", "ARRAY", "LIST", "HASH", "I/O", "Binary", "File",
		"Process", "Time", "Modules", "Objects", "Namespace", "Misc")
	names := slices.DeleteFunc(perlFunctionKinds(t, "User", "Network", "Socket", "SysV", "Flow"),
		func(name string) bool { return slices.Contains(others, name) })
	// Measured on 5.42.0: an empty or short answer is perl not being asked.
	if len(names) != 65 {
		t.Fatalf("Pod::Functions gives %d prototyped user, network, socket, IPC and control-flow builtins, %v; measured 65", len(names), names)
	}
	return names
}

// TestCoreSystemFlowBuiltinsTyped: every user, network, socket, IPC and
// control-flow builtin has a typed CORE.pmt line.
func TestCoreSystemFlowBuiltinsTyped(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range untypedLines(src, systemFlowBuiltins(t)) {
		t.Errorf("%s: CORE.pmt declares it untyped", name)
	}
}

// TestCoreNiladicFlowKeywordsTypedEmpty: a control-flow keyword perl
// prototypes `()` -- `__FILE__`, `wantarray`, `break` -- is typed with no
// parameters, so its derived prototype is the empty one, present in the
// prototype table.
func TestCoreNiladicFlowKeywordsTypedEmpty(t *testing.T) {
	perl := perlPrototypes(t)
	core := parse.CoreSignatures()
	table := parse.CoreTable()
	for _, name := range []string{"__FILE__", "wantarray", "break"} {
		if proto, ok := perl[name]; !ok || proto != "" {
			t.Errorf("%s: perl reports (%s), present %v; measured ()", name, proto, ok)
		}
		sigs := core[name]
		if len(sigs) != 1 || len(sigs[0].Params) != 0 || sigs[0].Returns == types.Unknown {
			t.Errorf("%s: CORE.pmt has %+v; want one typed line with no parameters", name, sigs)
		}
		if proto, ok := table[name]; !ok || proto != "" {
			t.Errorf("%s: the prototype table has (%s), present %v; want ()", name, proto, ok)
		}
	}
}

// prototypeOnlyLines are the prototyped CORE.pmt lines that carry no
// types, each with what holds it so. lock, pos, tie, tied, undef and
// untie hold a glob slot, which has no spelling yet (RFC 0001, open
// question 11, perigrin 2026-10-03; "The scalar container" keeps them
// prototype-only). catch, isa and method are keywords no call reaches,
// so there is no value to measure (01a111b5-1a31-70f3-9a21-023067e90a6f).
var prototypeOnlyLines = map[string]string{
	"lock":   "a glob slot, RFC 0001 open question 11",
	"pos":    "a glob slot, RFC 0001 open question 11",
	"tie":    "a glob slot, RFC 0001 open question 11",
	"tied":   "a glob slot, RFC 0001 open question 11",
	"undef":  "a glob slot, RFC 0001 open question 11",
	"untie":  "a glob slot, RFC 0001 open question 11",
	"catch":  "no call reaches it, 01a111b5-1a31-70f3-9a21-023067e90a6f",
	"isa":    "no call reaches it, 01a111b5-1a31-70f3-9a21-023067e90a6f",
	"method": "no call reaches it, 01a111b5-1a31-70f3-9a21-023067e90a6f",
}

// untypedPrototypedLines reports each builtin perl prototypes whose
// CORE.pmt line is untyped, and each of prototypeOnlyLines that is typed
// all the same. A file in error reads as nothing, so its error is the
// report.
func untypedPrototypedLines(src []byte, perl map[string]string) []string {
	if _, err := parse.CoreSignaturesOf(src); err != nil {
		return []string{err.Error()}
	}
	names := slices.Sorted(maps.Keys(perl))
	untyped := untypedLines(src, names)
	var bad []string
	for _, name := range names {
		why, exempt := prototypeOnlyLines[name]
		switch isUntyped := slices.Contains(untyped, name); {
		case exempt && !isUntyped:
			bad = append(bad, fmt.Sprintf("%s: typed, though it stays prototype-only (%s)", name, why))
		case !exempt && isUntyped:
			bad = append(bad, name+": CORE.pmt declares it untyped")
		}
	}
	return bad
}

// TestCoreEveryPrototypedLineIsTyped: every one of the 188 builtins perl
// prototypes has a typed CORE.pmt line, but for prototypeOnlyLines, which
// carry none, and the file reads with no declaration errors.
func TestCoreEveryPrototypedLineIsTyped(t *testing.T) {
	perl := perlPrototypes(t)
	if len(perl) != 188 {
		t.Fatalf("perl prototypes %d builtins; measured 188", len(perl))
	}
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range untypedPrototypedLines(src, perl) {
		t.Error(b)
	}
}

// TestCoreEveryLineCheckExemptsOnlyPrototypeOnlyLines: the exemptions are exact,
// not a loophole. Over a copy of CORE.pmt whose socket line is
// prototype-only again the check reports socket, and over copies with pos
// or catch typed it reports that line.
func TestCoreEveryLineCheckExemptsOnlyPrototypeOnlyLines(t *testing.T) {
	perl := perlPrototypes(t)
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ name, line, want string }{
		{"socket", `sub socket :prototype(*$$$);`, "socket: CORE.pmt declares it untyped"},
		{"pos", `sub pos (Scalar $x = $_) Int|Undef;`, "pos: typed, though it stays prototype-only (a glob slot, RFC 0001 open question 11)"},
		{"catch", `sub catch () None;`, "catch: typed, though it stays prototype-only (no call reaches it, 01a111b5-1a31-70f3-9a21-023067e90a6f)"},
	} {
		re := regexp.MustCompile(`(?m)^sub ` + c.name + ` .*;$`)
		if len(re.FindAll(src, -1)) != 1 {
			t.Fatalf("CORE.pmt has no one %s line to rewrite", c.name)
		}
		got := untypedPrototypedLines(re.ReplaceAllLiteral(src, []byte(c.line)), perl)
		if !slices.Equal(got, []string{c.want}) {
			t.Errorf("with %s: reported %q; want %q", c.line, got, c.want)
		}
	}
}
