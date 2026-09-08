// ABOUTME: Tests for system-Perl fallback when no system Perl has been imported into the registry
// ABOUTME: Reproduces the pvm use system / pvm current divergence reported in issue #448

package perl

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResolveFromSystemPerl_FallsBackToDetection reproduces issue #448.
// Scenario: user has at least one PVM-managed Perl installed (so the
// auto-import block at the top of ResolveVersion does not fire), but never
// ran `pvm import-system`, so the registry has no Source=="system" entry.
// Then `pvm use system` clears PVM_PERL_VERSION and `pvm current` falls
// through to resolveFromSystemPerl. Today that errors with "Run pvm
// import-system to register it" even though DetectSystemPerl() can find
// /usr/bin/perl on PATH. After the fix, resolveFromSystemPerl falls back
// to DetectSystemPerl when the registry has no system entry.
func TestResolveFromSystemPerl_FallsBackToDetection(t *testing.T) {
	origGetInstalledVersions := GetInstalledVersions
	origDetectSystemPerl := DetectSystemPerl
	defer func() {
		GetInstalledVersions = origGetInstalledVersions
		DetectSystemPerl = origDetectSystemPerl
	}()

	// Registry has only PVM-managed Perls; no Source=="system" entry
	// because the user never ran `pvm import-system`.
	GetInstalledVersions = func() ([]VersionInfo, error) {
		return []VersionInfo{
			{
				Version:     "5.42.0",
				InstallPath: "/home/user/.local/share/pvm/versions/5.42.0",
				Source:      "pvm",
			},
		}, nil
	}

	// DetectSystemPerl can find /usr/bin/perl regardless.
	DetectSystemPerl = func() (*SystemPerl, error) {
		path := "/usr/bin/perl"
		if runtime.GOOS == "windows" {
			path = `C:\Strawberry\perl\bin\perl.exe`
		}
		return &SystemPerl{
			Path:    path,
			Version: "5.34.1",
		}, nil
	}

	resolved, err := resolveFromSystemPerl()

	assert.NoError(t, err, "resolveFromSystemPerl must not error when DetectSystemPerl finds a Perl")
	if !assert.NotNil(t, resolved) {
		return
	}
	assert.Equal(t, "5.34.1", resolved.Version)
	assert.Equal(t, SystemPerlSource, resolved.Source)
}

// TestResolveFromSystemPerl_RegistryEntryStillPreferred verifies the
// existing happy path: when the registry has a Source=="system" entry,
// it is returned and DetectSystemPerl is not consulted. This locks in
// the prior behavior so the fallback only applies when registry has no
// system entry.
func TestResolveFromSystemPerl_RegistryEntryStillPreferred(t *testing.T) {
	origGetInstalledVersions := GetInstalledVersions
	origDetectSystemPerl := DetectSystemPerl
	defer func() {
		GetInstalledVersions = origGetInstalledVersions
		DetectSystemPerl = origDetectSystemPerl
	}()

	GetInstalledVersions = func() ([]VersionInfo, error) {
		return []VersionInfo{
			{
				Version:     "5.34.1",
				InstallPath: "/usr",
				Source:      "system",
			},
		}, nil
	}

	// If DetectSystemPerl is called, fail loudly — the registry entry
	// should be preferred and detection should not run.
	DetectSystemPerl = func() (*SystemPerl, error) {
		t.Fatal("DetectSystemPerl must not be called when a Source=='system' registry entry exists")
		return nil, nil
	}

	resolved, err := resolveFromSystemPerl()

	assert.NoError(t, err)
	assert.NotNil(t, resolved)
	assert.Equal(t, "5.34.1", resolved.Version)
	assert.Equal(t, SystemPerlSource, resolved.Source)
}

// TestResolveFromSystemPerl_NoSystemPerlAtAll verifies the genuinely-no-Perl
// case still errors. After the fix, the error message should mention what
// was tried (registry + detection), not just "Run pvm import-system".
func TestResolveFromSystemPerl_NoSystemPerlAtAll(t *testing.T) {
	origGetInstalledVersions := GetInstalledVersions
	origDetectSystemPerl := DetectSystemPerl
	defer func() {
		GetInstalledVersions = origGetInstalledVersions
		DetectSystemPerl = origDetectSystemPerl
	}()

	GetInstalledVersions = func() ([]VersionInfo, error) {
		return []VersionInfo{}, nil
	}
	DetectSystemPerl = func() (*SystemPerl, error) {
		return nil, assert.AnError
	}

	resolved, err := resolveFromSystemPerl()

	assert.Error(t, err)
	assert.Nil(t, resolved)
}
