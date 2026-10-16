// ABOUTME: Tests for incRoots, the @INC psc check and the LSP parse a file against.
// ABOUTME: Fixture projects name fixture installs through a fixture registry; no perl is run.

package psc

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/platform"
)

// rootsInstall lays out a pvm-style install whose Config.pm names privlib
// as its one library directory, and returns its prefix. Its perl is
// bin/perl.exe on Windows, as pvm installs it there.
func rootsInstall(t *testing.T, privlib string) string {
	t.Helper()
	prefix := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(prefix, "bin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(prefix, "bin", platform.ExecutableName("perl")), []byte("not a program\n"), 0o644))
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

	roots, why := incRoots(target)
	assert.Equal(t, []string{"/target-perl/lib"}, roots)
	assert.Empty(t, why)
	roots, _ = incRoots(filepath.Dir(target))
	assert.Equal(t, []string{"/target-perl/lib"}, roots, "a directory target resolves from itself")
}

// TestIncRootsReportsUnresolvedPerl checks that a perl the project names
// but psc cannot find statically is reported, once per psc check run,
// and @INC falls back to PERL5LIB; a project naming no perl falls back
// the same way and reports nothing.
func TestIncRootsReportsUnresolvedPerl(t *testing.T) {
	rootsRegistry(t, map[string]string{})
	lib := t.TempDir()
	t.Setenv("PERL5LIB", lib)

	named := filepath.Dir(rootsProject(t, "dev"))
	require.NoError(t, os.WriteFile(filepath.Join(named, "b.pl"), []byte("1;\n"), 0o644))
	roots, why := incRoots(named)
	assert.Equal(t, []string{lib}, roots)
	assert.Contains(t, why, `"dev"`)

	out := runCheck(t, named)
	assert.Equal(t, 1, strings.Count(out, "psc: note: "), "reported once per run: %s", out)
	assert.Contains(t, out, `names "dev", which only MetaCPAN can resolve; @INC is PERL5LIB alone`)

	// The control: a project that names no perl.
	plain := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(plain, "a.pl"), []byte("1;\n"), 0o644))
	roots, why = incRoots(plain)
	assert.Equal(t, []string{lib}, roots)
	assert.Empty(t, why)
	assert.Empty(t, runCheck(t, plain))
}

// runCheck runs psc check on target and returns what it prints.
func runCheck(t *testing.T, target string) string {
	t.Helper()
	cmd := NewCommand()
	cmd.SetArgs([]string{"check", target})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	_ = cmd.Execute()
	return out.String()
}

// TestLSPLogsUnresolvedPerl checks that the LSP reports a perl the
// project names but psc cannot resolve as psc check does, as a note, sent
// as a window/logMessage once per directory rather than as a diagnostic:
// a diagnostic would put it in every file and differ between machines
// that do and do not have that perl. A project naming no perl, the
// control, logs nothing.
func TestLSPLogsUnresolvedPerl(t *testing.T) {
	rootsRegistry(t, map[string]string{})
	h, out := newTestHandler(t)
	h.initialized = true

	// logs opens path and returns the messages of the logMessages the open
	// sent; its diagnostics are published, and none of them is the note.
	logs := func(path string) []string {
		t.Helper()
		uri := FileURI(path)
		params, err := json.Marshal(lspDidOpenParams{TextDocument: lspTextDocumentItem{URI: uri, LanguageID: "perl", Version: 1, Text: "1;\n"}})
		require.NoError(t, err)
		require.NoError(t, h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: params}))
		var logged []string
		published := false
		// One transport reads them all: it buffers ahead of the message it
		// returns.
		tr := newTransport(out, io.Discard)
		for {
			msg, err := tr.readMessage()
			if err != nil {
				break
			}
			switch msg.Method {
			case "window/logMessage":
				var p lspLogMessageParams
				require.NoError(t, json.Unmarshal(msg.Params, &p))
				logged = append(logged, p.Message)
			case "textDocument/publishDiagnostics":
				published = true
				assert.NotContains(t, string(msg.Params), "psc: note", "the note is not a diagnostic")
			}
		}
		assert.True(t, published, "%s: diagnostics are published", path)
		return logged
	}

	dev := filepath.Dir(rootsProject(t, "dev"))
	first := logs(filepath.Join(dev, "a.pl"))
	if assert.Len(t, first, 1, "the first open in a directory logs the note") {
		assert.Contains(t, first[0], `psc: note: `)
		assert.Contains(t, first[0], `names "dev", which only MetaCPAN can resolve; @INC is PERL5LIB alone`)
	}
	assert.Empty(t, logs(filepath.Join(dev, "b.pl")), "a second file in the directory logs nothing")
	assert.Empty(t, logs(filepath.Join(dev, "a.pl")), "a reopen logs nothing")
	assert.Len(t, logs(filepath.Join(filepath.Dir(rootsProject(t, "dev")), "a.pl")), 1, "another directory logs its own note")

	plain := t.TempDir()
	assert.Empty(t, logs(filepath.Join(plain, "a.pl")), "control: a project naming no perl logs nothing")
}
