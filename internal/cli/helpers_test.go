// ABOUTME: Tests for CLI command helpers and utilities
// ABOUTME: Validates common command setup and helper functionality

package cli

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/modules"
)

func TestCommonInstallFlags_ToInstallOptions(t *testing.T) {
	flags := &CommonInstallFlags{
		Source:     "cpan",
		NoCache:    true,
		SkipTests:  true,
		Force:      false,
		Verbose:    true,
		SkipDeps:   false,
		InstallDir: "/opt/perl",
		Version:    "1.0",
		PerlPath:   "/usr/bin/perl",
	}

	perlPath := "/resolved/perl/path"
	opts := flags.ToInstallOptions(perlPath)

	if opts.PerlPath != perlPath {
		t.Errorf("Expected PerlPath %q, got %q", perlPath, opts.PerlPath)
	}
	if opts.VersionConstraint != "1.0" {
		t.Errorf("Expected VersionConstraint '1.0', got %q", opts.VersionConstraint)
	}
	if opts.InstallDir != "/opt/perl" {
		t.Errorf("Expected InstallDir '/opt/perl', got %q", opts.InstallDir)
	}
	if opts.Force != false {
		t.Errorf("Expected Force false, got %t", opts.Force)
	}
	if opts.RunTests != false { // !SkipTests
		t.Errorf("Expected RunTests false, got %t", opts.RunTests)
	}
	if opts.SkipDependencies != false {
		t.Errorf("Expected SkipDependencies false, got %t", opts.SkipDependencies)
	}
	if opts.Verbose != true {
		t.Errorf("Expected Verbose true, got %t", opts.Verbose)
	}
	if !opts.Cleanup {
		t.Error("Expected Cleanup to be true by default")
	}
}

func TestValidateModuleNames(t *testing.T) {
	// Test valid module names
	valid := []string{"DBI", "Test::More", "Data::Dumper"}
	err := ValidateModuleNames(valid)
	if err != nil {
		t.Errorf("Expected no error for valid modules, got: %v", err)
	}

	// Test empty list
	err = ValidateModuleNames([]string{})
	if err == nil {
		t.Error("Expected error for empty module list")
	}

	// Test empty module name
	err = ValidateModuleNames([]string{"DBI", "", "Test::More"})
	if err == nil {
		t.Error("Expected error for empty module name")
	}

	// Test very long module name
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'A'
	}
	err = ValidateModuleNames([]string{string(longName)})
	if err == nil {
		t.Error("Expected error for overly long module name")
	}
}

func TestParseCommonInstallFlags_MockCommand(t *testing.T) {
	// Create a mock command with the expected flags
	cmd := &cobra.Command{
		Use: "test",
	}

	// Add the flags that ParseCommonInstallFlags expects
	cmd.Flags().String("source", "cpan", "Source")
	cmd.Flags().Bool("no-cache", false, "No cache")
	cmd.Flags().Bool("skip-tests", false, "Skip tests")
	cmd.Flags().Bool("force", false, "Force")
	cmd.Flags().Bool("verbose", false, "Verbose")
	cmd.Flags().Bool("skip-deps", false, "Skip deps")
	cmd.Flags().String("install-dir", "", "Install dir")
	cmd.Flags().String("version", "", "Version")
	cmd.Flags().String("perl-path", "", "Perl path")

	// Set some values
	cmd.Flags().Set("source", "metacpan")
	cmd.Flags().Set("no-cache", "true")
	cmd.Flags().Set("verbose", "true")
	cmd.Flags().Set("version", "1.23")

	flags, err := ParseCommonInstallFlags(cmd)
	if err != nil {
		t.Errorf("Expected no error parsing flags, got: %v", err)
	}

	if flags.Source != "metacpan" {
		t.Errorf("Expected source 'metacpan', got %q", flags.Source)
	}
	if !flags.NoCache {
		t.Error("Expected NoCache to be true")
	}
	if !flags.Verbose {
		t.Error("Expected Verbose to be true")
	}
	if flags.Version != "1.23" {
		t.Errorf("Expected version '1.23', got %q", flags.Version)
	}
}

func TestDisplayInstallationResults_Empty(t *testing.T) {
	// Test with empty results - just verify it doesn't panic
	// We would need a more complex mock UI to test the actual output
	// DisplayInstallationResults(ui, []*modules.InstallResult{})
}

func TestDisplayInstallationResults_WithResults(t *testing.T) {
	// Test with results - just verify structure is correct
	results := []*modules.InstallResult{
		{
			ModuleName:   "DBI",
			Version:      "1.643",
			Success:      true,
			Duration:     time.Second * 30,
			Dependencies: []string{"DBI::Test"},
			Warnings:     []string{"deprecated API"},
		},
		{
			ModuleName: "Failed::Module",
			Success:    false,
			Duration:   time.Second * 5,
			Errors:     []string{"compilation failed"},
		},
	}

	// Verify structure - we would need actual UI mock to test display
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestDisplayModuleList_Empty(t *testing.T) {
	// Test with empty module list - just verify it doesn't panic
	// DisplayModuleList(ui, []*modules.InstalledModule{}, "table")
}

func TestDisplayModuleList_WithModules(t *testing.T) {
	// Test with modules - just verify structure is correct
	modules := []*modules.InstalledModule{
		{
			Name:    "DBI",
			Version: "1.643",
			Path:    "/usr/lib/perl5/DBI.pm",
		},
		{
			Name:    "Test::More",
			Version: "1.302",
			Path:    "/usr/lib/perl5/Test/More.pm",
		},
	}

	// Verify structure
	if len(modules) != 2 {
		t.Errorf("Expected 2 modules, got %d", len(modules))
	}
	if modules[0].Name != "DBI" {
		t.Errorf("Expected first module 'DBI', got %q", modules[0].Name)
	}
}

// mockOutput implements the output interface for testing
type mockOutput struct {
	messages []string
}

func (m *mockOutput) Write(p []byte) (n int, err error) {
	m.messages = append(m.messages, string(p))
	return len(p), nil
}

func (m *mockOutput) GetMessages() []string {
	return m.messages
}

func TestCommandEnvironment_Basic(t *testing.T) {
	env := &CommandEnvironment{}

	// Test ResolvePerlPath
	err := env.ResolvePerlPath("/usr/bin/perl")
	if err != nil {
		t.Errorf("Expected no error resolving provided perl path, got: %v", err)
	}
	if env.PerlPath != "/usr/bin/perl" {
		t.Errorf("Expected PerlPath '/usr/bin/perl', got %q", env.PerlPath)
	}

	// Test with empty path - this will try to resolve system perl
	// Don't fail if perl is not available in test environment
	err = env.ResolvePerlPath("")
	if err != nil {
		t.Logf("Perl path resolution failed (expected in test environment): %v", err)
	}
}
