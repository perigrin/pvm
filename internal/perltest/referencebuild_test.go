// ABOUTME: Tests that IsReferenceBuild recognizes only the build its exact counts were measured on:
// ABOUTME: a static libperl whose undefined %Config keys and privlib .pm files hash to that build's.

package perltest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsStaticLibperl(t *testing.T) {
	assert.True(t, isStaticLibperl("false"), "useshrplib false is pvm's own static build")
	assert.True(t, isStaticLibperl("false\n"), "perl's own print leaves no trailing newline trimmed")
	assert.False(t, isStaticLibperl("true"), "useshrplib true is a shared-libperl build, not the reference one")
	assert.False(t, isStaticLibperl(""), "no answer at all is not the reference build")
}

// TestUndefKeysDigestIsOrderFreeAndExact: the digest ignores the order perl
// lists keys in, and one key more or less is another build.
func TestUndefKeysDigestIsOrderFreeAndExact(t *testing.T) {
	a := undefKeysDigest([]string{"i_gdbm", "d_ndbm", "i_db"})
	assert.Equal(t, a, undefKeysDigest([]string{"i_db", "i_gdbm", "d_ndbm"}))
	assert.NotEqual(t, a, undefKeysDigest([]string{"i_db", "i_gdbm"}))
	assert.NotEqual(t, a, undefKeysDigest([]string{"i_db", "i_gdbm", "d_ndbm", "i_ndbm"}))
}

// TestPrivlibDigestMatchesShellPipeline: privlibModules and privlibDigest
// hash a tree as `find . -name '*.pm' | sed 's|^\./||' | LC_ALL=C sort`
// piped to sha256 does, which is how the reference digest was taken.
func TestPrivlibDigestMatchesShellPipeline(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skipf("no sh to run the pipeline: %v", err)
	}
	if _, err := exec.LookPath("sha256sum"); err != nil {
		t.Skipf("no sha256sum to run the pipeline: %v", err)
	}
	dir := t.TempDir()
	for _, f := range []string{"Carp.pm", "App/Cpan.pm", "App/Prove.pm", "a_lower.pm", "Z.pm", "notes.txt"} {
		path := filepath.Join(dir, filepath.FromSlash(f))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, nil, 0o644))
	}
	paths, err := privlibModules(dir)
	require.NoError(t, err)
	assert.Len(t, paths, 5, "only the .pm files")

	cmd := exec.Command("sh", "-c", `find . -name '*.pm' | sed 's|^\./||' | LC_ALL=C sort | sha256sum`)
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, strings.Fields(string(out))[0], privlibDigest(paths))
}

// TestIsReferenceBuildMatchesItsDigests: on pvm's own perl 5.42.0,
// IsReferenceBuild is exactly the conjunction of its three measurements, so
// a build that matches the pinned counts' build is recognized, and any
// other -- another machine's pvm build included -- is not.
func TestIsReferenceBuildMatchesItsDigests(t *testing.T) {
	p := Perl542(t)
	if !p.FromPvm {
		t.Skip("perl came from PATH, not pvm's own registry")
	}
	shr, err := p.Command("-MConfig", "-e", `print $Config{useshrplib}`).Output()
	require.NoError(t, err)
	keys, err := p.Command("-MConfig", "-e", `print "$_\n" for grep { !defined $Config{$_} } keys %Config`).Output()
	require.NoError(t, err)
	paths, err := privlibModules(p.Privlib)
	require.NoError(t, err)
	want := isStaticLibperl(string(shr)) &&
		undefKeysDigest(strings.Fields(string(keys))) == referenceUndefKeysDigest &&
		privlibDigest(paths) == referencePrivlibDigest
	assert.Equal(t, want, p.IsReferenceBuild(t))
	t.Logf("%s: reference build %v", p.Path, want)
}
