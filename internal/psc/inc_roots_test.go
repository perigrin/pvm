// ABOUTME: Tests for incRoots, the @INC psc check and the LSP parse a file against.
// ABOUTME: Fixture projects name fixture installs through a fixture registry; no perl is run.

package psc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rootsInstall lays out a pvm-style install whose Config.pm names privlib
// as its one library directory, and returns its prefix.
func rootsInstall(t *testing.T, privlib string) string {
	t.Helper()
	prefix := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(prefix, "bin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(prefix, "bin", "perl"), []byte("not a program\n"), 0o644))
	arch := filepath.Join(prefix, "lib", "perl5", "5.42.0", "x86_64-linux")
	require.NoError(t, os.MkdirAll(arch, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(arch, "Config.pm"),
		[]byte("package Config;\nour %Config = (\n    privlibexp => '"+privlib+"',\n);\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(arch, "Config_heavy.pl"), []byte("1;\n"), 0o644))
	return prefix
}

// rootsRegistry points pvm's registry at a fresh file listing versions,
// each installed at its prefix, and leaves no other perl to find.
func rootsRegistry(t *testing.T, versions map[string]string) {
	t.Helper()
	home := t.TempDir()
	for _, v := range []string{"HOME", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME", "XDG_STATE_HOME"} {
		t.Setenv(v, home)
	}
	for _, v := range []string{"PERL5LIB", "PVM_PERL_VERSION", "PLENV_VERSION", "PERLBREW_PERL"} {
		t.Setenv(v, "")
	}
	entries := map[string]map[string]string{}
	for version, prefix := range versions {
		entries[version] = map[string]string{"version": version, "install_path": prefix, "source": "pvm"}
	}
	data, err := json.Marshal(map[string]any{"versions": entries})
	require.NoError(t, err)
	path := filepath.Join(home, "registry.json")
	require.NoError(t, os.WriteFile(path, data, 0o644))
	t.Setenv("PVM_PERL_REGISTRY", path)
}

// rootsProject is a fresh directory whose .pvm/pvm.toml names version,
// holding lib/a.pl, whose path it returns.
func rootsProject(t *testing.T, version string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".pvm"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".pvm", "pvm.toml"),
		[]byte("[pvm]\ndefault_perl = \""+version+"\"\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "lib"), 0o755))
	file := filepath.Join(dir, "lib", "a.pl")
	require.NoError(t, os.WriteFile(file, []byte("1;\n"), 0o644))
	return file
}

// TestIncRootsFromTargetProject checks that the perl incRoots resolves is
// the one the target's project names in its pvm.toml, not the one the
// process's working directory names, so psc check run from anywhere and
// the LSP started from anywhere parse a file alike.
func TestIncRootsFromTargetProject(t *testing.T) {
	rootsRegistry(t, map[string]string{
		"5.40.0": rootsInstall(t, "/cwd-perl/lib"),
		"5.42.0": rootsInstall(t, "/target-perl/lib"),
	})
	cwdProject := filepath.Dir(filepath.Dir(rootsProject(t, "5.40.0")))
	target := rootsProject(t, "5.42.0")
	t.Chdir(cwdProject)

	assert.Equal(t, []string{"/target-perl/lib"}, incRoots(target))
	assert.Equal(t, []string{"/target-perl/lib"}, incRoots(filepath.Dir(target)),
		"a directory target resolves from itself")
}
