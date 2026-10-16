// ABOUTME: Runs CORE.tpm tests each alone, in a fresh process, as git zhi verify does.
// ABOUTME: A CORE.tpm read must derive the parser's binding powers itself, not inherit them.
package parse_test

import (
	"os"
	"os/exec"
	"testing"
)

// TestCoreReadsAloneDeriveTheirPowers: a test that reads CORE.tpm must pass
// when it is the only test the process runs. The binding powers are derived
// from CORE.tpm once per process (RFC 0001, "Precedence is a relation
// between operators"), and a CORE.tpm read that left the deriving to some
// earlier parse misreads defaults such as chdir's `$dir = $_` when nothing
// parsed before it.
func TestCoreReadsAloneDeriveTheirPowers(t *testing.T) {
	if os.Getenv("PVM_ISOLATION_CHILD") != "" {
		t.Skip("running as a child of TestCoreReadsAloneDeriveTheirPowers")
	}
	for _, name := range []string{
		"TestCoreDerivedPrototypesArePerls",
		"TestCoreKeysValuesMatchMeasuredSignatures",
		"TestCoreAllAnyDeriveRefAmpAt",
		"TestCoreGrepMapSortOutOfCoreTable",
	} {
		cmd := exec.Command(os.Args[0], "-test.run=^"+name+"$", "-test.count=1")
		cmd.Env = append(os.Environ(), "PVM_ISOLATION_CHILD=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("%s alone: %v\n%s", name, err, out)
		}
	}
}
