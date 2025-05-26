// ABOUTME: Test suite for Carton integration functionality
// ABOUTME: Tests cpanfile and cpanfile.snapshot parsing along with module resolution

package cpan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCPANFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected *CPANFile
	}{
		{
			name: "Basic requirements",
			content: `requires 'Moose', '2.0';
requires 'DBI';
requires 'JSON', '>= 4.0';`,
			expected: &CPANFile{
				Requirements: []Requirement{
					{Module: "Moose", Version: "2.0", Relationship: "requires", Phase: "runtime"},
					{Module: "DBI", Relationship: "requires", Phase: "runtime"},
					{Module: "JSON", Version: ">= 4.0", Relationship: "requires", Phase: "runtime"},
				},
				Features:  make(map[string][]Requirement),
				Platforms: make(map[string][]Requirement),
			},
		},
		{
			name: "Test and build requirements",
			content: `requires 'Moose';
test_requires 'Test::More', '0.88';
build_requires 'Module::Build';`,
			expected: &CPANFile{
				Requirements: []Requirement{
					{Module: "Moose", Relationship: "requires", Phase: "runtime"},
					{Module: "Test::More", Version: "0.88", Relationship: "test_requires", Phase: "test"},
					{Module: "Module::Build", Relationship: "build_requires", Phase: "build"},
				},
				Features:  make(map[string][]Requirement),
				Platforms: make(map[string][]Requirement),
			},
		},
		{
			name: "Feature requirements",
			content: `requires 'Moose';
feature 'mysql', 'MySQL support' => sub {
    requires 'DBI';
    requires 'DBD::mysql';
};`,
			expected: &CPANFile{
				Requirements: []Requirement{
					{Module: "Moose", Relationship: "requires", Phase: "runtime"},
				},
				Features: map[string][]Requirement{
					"mysql": {
						{Module: "DBI", Relationship: "requires", Phase: "runtime"},
						{Module: "DBD::mysql", Relationship: "requires", Phase: "runtime"},
					},
				},
				Platforms: make(map[string][]Requirement),
			},
		},
		{
			name: "Platform requirements",
			content: `requires 'Moose';
on 'MSWin32' => sub {
    requires 'Win32::API';
};`,
			expected: &CPANFile{
				Requirements: []Requirement{
					{Module: "Moose", Relationship: "requires", Phase: "runtime"},
				},
				Features: make(map[string][]Requirement),
				Platforms: map[string][]Requirement{
					"MSWin32": {
						{Module: "Win32::API", Relationship: "requires", Phase: "runtime"},
					},
				},
			},
		},
		{
			name: "Phase-specific requirements",
			content: `requires 'Moose';
on 'test' => sub {
    requires 'Test::More';
    requires 'Test::Exception';
};`,
			expected: &CPANFile{
				Requirements: []Requirement{
					{Module: "Moose", Relationship: "requires", Phase: "runtime"},
					{Module: "Test::More", Relationship: "requires", Phase: "test"},
					{Module: "Test::Exception", Relationship: "requires", Phase: "test"},
				},
				Features:  make(map[string][]Requirement),
				Platforms: make(map[string][]Requirement),
			},
		},
		{
			name: "Recommends and suggests",
			content: `requires 'Moose';
recommends 'JSON::XS';
suggests 'YAML::Syck';`,
			expected: &CPANFile{
				Requirements: []Requirement{
					{Module: "Moose", Relationship: "requires", Phase: "runtime"},
					{Module: "JSON::XS", Relationship: "recommends", Phase: "runtime"},
					{Module: "YAML::Syck", Relationship: "suggests", Phase: "runtime"},
				},
				Features:  make(map[string][]Requirement),
				Platforms: make(map[string][]Requirement),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCPANFile(tt.content)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseCPANSnapshot(t *testing.T) {
	content := `# carton snapshot format: version 1.0
DISTRIBUTIONS
  App-cpanminus-1.7046
    pathname: M/MI/MIYAGAWA/App-cpanminus-1.7046.tar.gz
    provides:
      App::cpanminus 1.7046
      App::cpanminus::fatscript undef
  DBI-1.643
    pathname: T/TI/TIMB/DBI-1.643.tar.gz
    provides:
      DBI 1.643
      DBI::DBD 12.015924`

	expected := &CPANSnapshot{
		Modules: map[string]SnapshotModule{
			"App::cpanminus": {
				Version:      "1.7046",
				Distribution: "App-cpanminus-1.7046",
				Dependencies: []Dependency{},
			},
			"App::cpanminus::fatscript": {
				Version:      "undef",
				Distribution: "App-cpanminus-1.7046",
				Dependencies: []Dependency{},
			},
			"DBI": {
				Version:      "1.643",
				Distribution: "DBI-1.643",
				Dependencies: []Dependency{},
			},
			"DBI::DBD": {
				Version:      "12.015924",
				Distribution: "DBI-1.643",
				Dependencies: []Dependency{},
			},
		},
	}

	result, err := ParseCPANSnapshot(content)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestCartonIntegration(t *testing.T) {
	// Create temporary project directory
	tempDir := t.TempDir()

	// Create cpanfile
	cpanfileContent := `requires 'Moose', '2.0';
requires 'DBI';
test_requires 'Test::More', '0.88';`

	err := os.WriteFile(filepath.Join(tempDir, "cpanfile"), []byte(cpanfileContent), 0644)
	require.NoError(t, err)

	// Create cpanfile.snapshot
	snapshotContent := `# carton snapshot format: version 1.0
DISTRIBUTIONS
  Moose-2.2015
    pathname: E/ET/ETHER/Moose-2.2015.tar.gz
    provides:
      Moose 2.2015
      Moose::Meta::Class 2.2015
  DBI-1.643
    pathname: T/TI/TIMB/DBI-1.643.tar.gz
    provides:
      DBI 1.643
      DBI::DBD 12.015924`

	err = os.WriteFile(filepath.Join(tempDir, "cpanfile.snapshot"), []byte(snapshotContent), 0644)
	require.NoError(t, err)

	// Create local lib structure
	localLib := filepath.Join(tempDir, "local", "lib", "perl5")
	err = os.MkdirAll(localLib, 0755)
	require.NoError(t, err)

	// Create a test module file
	mooseDir := filepath.Join(localLib, "Moose")
	err = os.MkdirAll(mooseDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(mooseDir, "Meta.pm"), []byte("package Moose::Meta; 1;"), 0644)
	require.NoError(t, err)

	// Create Carton instance
	carton := &Carton{
		CartonPath:  "/usr/bin/carton", // Mock path
		ProjectRoot: tempDir,
	}

	t.Run("ParseCPANFile", func(t *testing.T) {
		cpanfile, err := carton.ParseCPANFile()
		require.NoError(t, err)

		assert.Len(t, cpanfile.Requirements, 3)
		assert.Equal(t, "Moose", cpanfile.Requirements[0].Module)
		assert.Equal(t, "2.0", cpanfile.Requirements[0].Version)
		assert.Equal(t, "requires", cpanfile.Requirements[0].Relationship)
		assert.Equal(t, "runtime", cpanfile.Requirements[0].Phase)

		assert.Equal(t, "Test::More", cpanfile.Requirements[2].Module)
		assert.Equal(t, "test", cpanfile.Requirements[2].Phase)
	})

	t.Run("ParseSnapshot", func(t *testing.T) {
		snapshotPath := filepath.Join(tempDir, "cpanfile.snapshot")
		snapshot, err := carton.ParseSnapshot(snapshotPath)
		require.NoError(t, err)

		assert.Contains(t, snapshot.Modules, "Moose")
		assert.Equal(t, "2.2015", snapshot.Modules["Moose"].Version)
		assert.Equal(t, "Moose-2.2015", snapshot.Modules["Moose"].Distribution)

		assert.Contains(t, snapshot.Modules, "DBI")
		assert.Equal(t, "1.643", snapshot.Modules["DBI"].Version)
	})

	t.Run("GetInstalledModules", func(t *testing.T) {
		modules, err := carton.GetInstalledModules()
		require.NoError(t, err)

		assert.Len(t, modules, 4) // Moose, Moose::Meta::Class, DBI, DBI::DBD

		// Find Moose module
		var mooseModule *ModuleInfo
		for _, mod := range modules {
			if mod.Name == "Moose" {
				mooseModule = &mod
				break
			}
		}

		require.NotNil(t, mooseModule)
		assert.Equal(t, "2.2015", mooseModule.Version)
		assert.Equal(t, "Moose-2.2015", mooseModule.Distribution)
	})

	t.Run("GetModuleInfo", func(t *testing.T) {
		info, err := carton.GetModuleInfo("Moose")
		require.NoError(t, err)

		assert.Equal(t, "Moose", info.Name)
		assert.Equal(t, "2.2015", info.Version)
		assert.Equal(t, "Moose-2.2015", info.Distribution)
	})

	t.Run("GetModulePath", func(t *testing.T) {
		path, err := carton.GetModulePath("Moose::Meta")
		require.NoError(t, err)

		expectedPath := filepath.Join(localLib, "Moose", "Meta.pm")
		assert.Equal(t, expectedPath, path)
	})

	t.Run("GetDependencies", func(t *testing.T) {
		deps, err := carton.GetDependencies("")
		require.NoError(t, err)

		assert.Len(t, deps, 3)

		// Check for Moose dependency
		var mooseDep *Dependency
		for _, dep := range deps {
			if dep.Name == "Moose" {
				mooseDep = &dep
				break
			}
		}

		require.NotNil(t, mooseDep)
		assert.Equal(t, "2.0", mooseDep.Version)
		assert.Equal(t, "requires", mooseDep.Type)
		assert.Equal(t, "runtime", mooseDep.Phase)
	})
}

func TestCartonDetection(t *testing.T) {
	t.Run("FindProjectRoot", func(t *testing.T) {
		// Create temporary directory structure
		tempDir := t.TempDir()
		subDir := filepath.Join(tempDir, "sub", "dir")
		err := os.MkdirAll(subDir, 0755)
		require.NoError(t, err)

		// Create cpanfile in temp directory
		err = os.WriteFile(filepath.Join(tempDir, "cpanfile"), []byte("requires 'Moose';"), 0644)
		require.NoError(t, err)

		// Change to subdirectory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		err = os.Chdir(subDir)
		require.NoError(t, err)

		// Test findProjectRoot
		root := findProjectRoot()
		// Handle symlinks in temp directories (macOS has /private prefix)
		expected, _ := filepath.EvalSymlinks(tempDir)
		actual, _ := filepath.EvalSymlinks(root)
		assert.Equal(t, expected, actual)
	})

	t.Run("NewCarton", func(t *testing.T) {
		carton := NewCarton()
		assert.NotNil(t, carton)
		assert.Equal(t, "carton", carton.Name())
	})
}

func TestCartonModulePathResolution(t *testing.T) {
	tempDir := t.TempDir()

	// Create complex local lib structure with architecture-specific paths
	localLib := filepath.Join(tempDir, "local", "lib", "perl5")
	archDir := filepath.Join(localLib, "x86_64-linux-gnu-thread-multi")

	err := os.MkdirAll(archDir, 0755)
	require.NoError(t, err)

	// Create test modules in different locations
	testModules := map[string]string{
		"Simple::Module":     filepath.Join(localLib, "Simple", "Module.pm"),
		"Arch::Specific":     filepath.Join(archDir, "Arch", "Specific.pm"),
		"Deep::Nested::Path": filepath.Join(localLib, "Deep", "Nested", "Path.pm"),
	}

	for module, path := range testModules {
		err := os.MkdirAll(filepath.Dir(path), 0755)
		require.NoError(t, err)

		err = os.WriteFile(path, []byte("package "+module+"; 1;"), 0644)
		require.NoError(t, err)
	}

	carton := &Carton{
		CartonPath:  "/usr/bin/carton",
		ProjectRoot: tempDir,
	}

	for module, expectedPath := range testModules {
		t.Run("Module_"+module, func(t *testing.T) {
			path, err := carton.GetModulePath(module)
			require.NoError(t, err)
			assert.Equal(t, expectedPath, path)
		})
	}

	t.Run("ModuleNotFound", func(t *testing.T) {
		_, err := carton.GetModulePath("NonExistent::Module")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in carton local lib")
	})
}

func TestCartonErrors(t *testing.T) {
	tempDir := t.TempDir()

	carton := &Carton{
		CartonPath:  "/usr/bin/carton",
		ProjectRoot: tempDir,
	}

	t.Run("MissingCPANFile", func(t *testing.T) {
		_, err := carton.ParseCPANFile()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cpanfile not found")
	})

	t.Run("MissingSnapshot", func(t *testing.T) {
		_, err := carton.GetInstalledModules()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cpanfile.snapshot not found")
	})

	t.Run("MissingLocalLib", func(t *testing.T) {
		_, err := carton.GetModulePath("Some::Module")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "carton local lib not found")
	})
}
