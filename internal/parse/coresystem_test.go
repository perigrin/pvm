// ABOUTME: Tests that CORE.pmt types its user, network, socket, IPC and control-flow builtins.
// ABOUTME: Each type is the one measured on perl 5.42; niladic keywords derive an empty prototype.
package parse_test

import (
	"os"
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
