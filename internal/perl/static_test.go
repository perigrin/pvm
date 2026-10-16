// ABOUTME: Tests for StaticPerl, which finds a project's perl from its version files and pvm's registry alone.
// ABOUTME: Fixture registries and installs live under t.TempDir; nothing is run and nothing is fetched.

package perl

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// staticRegistry points pvm's registry at a fresh file holding entries.
func staticRegistry(t *testing.T, entries ...VersionInfo) {
	t.Helper()
	reg := VersionRegistry{Versions: map[string]VersionInfo{}}
	for i, e := range entries {
		reg.Versions[string(rune('a'+i))] = e
	}
	data, err := json.Marshal(reg)
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "registry.json")
	require.NoError(t, os.WriteFile(path, data, 0o644))
	t.Setenv("PVM_PERL_REGISTRY", path)
}

// staticProject is a fresh directory with the given files written into it.
func staticProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
	}
	return dir
}

func TestStaticPerl(t *testing.T) {
	installed, installedPerl := fakeInstall(t)
	system, systemPerl := fakeInstall(t)
	fork, forkPerl := fakeInstall(t)
	staticRegistry(t,
		VersionInfo{Version: "5.42.0", InstallPath: installed, Source: "pvm"},
		VersionInfo{Version: "5.30.0", InstallPath: system, Source: "system"},
		VersionInfo{Version: "5.40.0", InstallPath: fork, Source: "pvm", Remote: "acme", BaseVersion: "5.40.0"},
	)

	cases := []struct {
		name     string
		files    map[string]string
		wantPath string
		wantWhy  string
	}{
		{"nothing named", map[string]string{}, "", ""},
		{"installed version", map[string]string{".perl-version": "5.42.0\n"}, installedPerl, ""},
		{"system is the imported system perl", map[string]string{".perl-version": "system\n"}, systemPerl, ""},
		{"fork by display name", map[string]string{".perl-version": "acme/5.40.0\n"}, forkPerl, ""},
		{"not installed", map[string]string{".perl-version": "5.8.9\n"}, "", `names "5.8.9", which is not installed`},
		{"dev needs MetaCPAN", map[string]string{".perl-version": "dev\n"}, "", `names "dev", which only MetaCPAN can resolve`},
		{"@latest needs MetaCPAN", map[string]string{".perl-version": "@latest\n"}, "", `names "@latest", which only MetaCPAN can resolve`},
		{"unknown alias", map[string]string{".perl-version": "@mine\n"}, "", `names the unknown alias "@mine"`},
		{"pvm.toml", map[string]string{".pvm/pvm.toml": "[pvm]\ndefault_perl = \"5.42.0\"\n"}, installedPerl, ""},
		{"pvm.toml alias", map[string]string{".pvm/pvm.toml": "[pvm]\ndefault_perl = \"mine\"\n[pvm.version_aliases]\nmine = \"5.42.0\"\n"}, installedPerl, ""},
		{".perl-version before pvm.toml", map[string]string{
			".perl-version": "system\n", ".pvm/pvm.toml": "[pvm]\ndefault_perl = \"5.42.0\"\n"}, systemPerl, ""},
		{"unreadable pvm.toml", map[string]string{".pvm/pvm.toml": "[pvm\n"}, "", "pvm.toml cannot be read"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := staticProject(t, c.files)
			// The project is found from a subdirectory, as a file's is.
			sub := filepath.Join(dir, "lib")
			require.NoError(t, os.MkdirAll(sub, 0o755))
			path, why := StaticPerl(sub)
			assert.Equal(t, c.wantPath, path)
			if c.wantWhy == "" {
				assert.Empty(t, why)
			} else {
				assert.Contains(t, why, c.wantWhy)
			}
		})
	}
}

// TestStaticPerlSystemNotImported checks that `system` with no system perl
// in the registry is unresolved, not looked for on disk.
func TestStaticPerlSystemNotImported(t *testing.T) {
	staticRegistry(t)
	path, why := StaticPerl(staticProject(t, map[string]string{".perl-version": "system\n"}))
	assert.Empty(t, path)
	assert.Contains(t, why, "not imported into pvm")
}
