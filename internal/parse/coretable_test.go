// ABOUTME: Tests that the builtin table parsed from declarations/CORE.pmt is perl's own.
// ABOUTME: Every keyword's prototype is asked of perl 5.42, and the file must agree.
package parse_test

import (
	"os/exec"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/conformance"
	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

// TestCoreDeclarationsArePerls: CORE.pmt declares each builtin that has a
// prototype, `sub bless :prototype($;$);`, and parsing it builds the default
// table. For every keyword, the table must hold exactly what
// `prototype("CORE::name")` reports on perl 5.42.0 -- an entry when perl
// reports one, none when it reports undef or dies.
func TestCoreDeclarationsArePerls(t *testing.T) {
	want := perlPrototypes(t)
	got := parse.CoreTable()
	for name, proto := range want {
		if g, ok := got[name]; !ok || g != proto {
			t.Errorf("%s: perl says (%s); CORE.pmt has (%s), declared=%v", name, proto, g, ok)
		}
	}
	for name := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("%s: declared in CORE.pmt; perl reports no prototype", name)
		}
	}
	if len(want) != 188 {
		t.Errorf("perl reports %d prototyped keywords; measured 188", len(want))
	}
}

// perlPrototypes asks perl 5.42 for every keyword's prototype, without its
// parentheses, keeping those it reports; the test skips when there is no
// perl to ask.
func perlPrototypes(t *testing.T) map[string]string {
	t.Helper()
	perl, err := conformance.PerlPath()
	if err != nil {
		t.Skipf("no perl 5.42 to ask: %v", err)
	}
	// binmode: on Windows perl's text-mode STDOUT writes CRLF, which would
	// leave a \r on every prototype.
	cmd := exec.Command(perl, "-e", `binmode STDOUT; while (my $k = <STDIN>) { chomp $k;
	    my $p = eval { prototype("CORE::$k") };
	    print "$k\t$p\n" if defined $p }`)
	cmd.Stdin = strings.NewReader(strings.Join(lexer.Keywords(), "\n") + "\n")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("asking perl: %v", err)
	}
	want := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name, proto, _ := strings.Cut(line, "\t")
		want[name] = proto
	}
	return want
}
