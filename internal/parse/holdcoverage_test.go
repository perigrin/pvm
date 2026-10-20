// ABOUTME: Checks that internal/perltest's shared hold lists name exactly the ranking's shipped top 50.
// ABOUTME: A module the ranking adds or drops, or core_module_not_shipped.txt excuses, is caught here.

package parse_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/perltest"
)

// rankedModuleNames reads testdata/core_module_ranking.txt's Name column:
// the top 50 core modules TestCoreModuleRanking pins, one "Name Count"
// line each.
func rankedModuleNames(t *testing.T) []string {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "core_module_ranking.txt"))
	require.NoError(t, err)
	defer f.Close()
	var names []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		name, _, ok := strings.Cut(line, " ")
		require.True(t, ok, "malformed ranking line %q", line)
		names = append(names, name)
	}
	require.NoError(t, sc.Err())
	return names
}

// notShippedModuleNames reads testdata/core_module_not_shipped.txt's
// Module column, skipping blank and comment lines.
func notShippedModuleNames(t *testing.T) map[string]bool {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "core_module_not_shipped.txt"))
	require.NoError(t, err)
	defer f.Close()
	out := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, _, ok := strings.Cut(line, "\t")
		require.True(t, ok, "malformed not-shipped line %q", line)
		out[name] = true
	}
	require.NoError(t, sc.Err())
	return out
}

// TestHoldListsCoverEveryDeclaration: internal/perltest.AllHeld names
// exactly the ranked top-50 core modules that core_module_not_shipped.txt
// does not excuse -- the hold lists internal/parse and internal/infer
// measure against perl are derived from the ranking, not hand-kept
// separately from it.
func TestHoldListsCoverEveryDeclaration(t *testing.T) {
	notShipped := notShippedModuleNames(t)
	var want []string
	for _, name := range rankedModuleNames(t) {
		if !notShipped[name] {
			want = append(want, name)
		}
	}
	assert.ElementsMatch(t, want, perltest.AllHeld(),
		"perltest.AllHeld must name exactly the shipped top-50 core modules")
}
