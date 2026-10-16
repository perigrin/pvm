// ABOUTME: Tests that a required file's path is resolved only inside the directories the loader searches.
// ABOUTME: `require '../outside/x.pl'` names a file outside the project, which psc does not read.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parse"
)

// TestRequireStaysInRoots checks the loader against perl 5.42's rule
// for a required path: `./x.pl` and `../x.pl` are relative to the working
// directory only and never searched for in @INC (`perl -I../inc -e
// 'require "../x.pl"'` with ../inc/../x.pl present dies "Can't locate
// ../x.pl"), and any other relative path is searched for in @INC. psc
// stands the file's directory in for the working directory, and reads
// nothing outside the directories it searches: a path that climbs out of
// them, or an absolute one, is not found.
func TestRequireStaysInRoots(t *testing.T) {
	top := t.TempDir()
	proj := filepath.Join(top, "proj")
	for _, f := range []string{"outside/x.pl", "proj/x.pl", "proj/sub/y.pl"} {
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(top, f)), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(top, f), []byte("sub "+filepath.Base(filepath.Dir(f))+" { 1 }\n1;\n"), 0o644))
	}
	load := parse.DirLoader(proj)

	for _, name := range []string{"../outside/x.pl", "./../outside/x.pl", "sub/../../outside/x.pl", filepath.Join(top, "outside", "x.pl")} {
		_, found := load(name)
		assert.False(t, found, "%s is outside the roots", name)
	}
	// The controls: paths that stay inside are found.
	for _, name := range []string{"./x.pl", "x.pl", "sub/y.pl", "sub/../x.pl"} {
		_, found := load(name)
		assert.True(t, found, "%s is inside the roots", name)
	}
}
