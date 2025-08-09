// ABOUTME: Unit tests for PVM Perl subcommands functionality
// ABOUTME: Tests command structure, flag parsing, and validation logic for perl build and tarball commands

package pvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli/ui"
)

func TestNewPerlCommand(t *testing.T) {
	cmd := newPerlCommand()

	// Test command metadata
	if cmd.Use != "perl" {
		t.Errorf("Expected command use to be 'perl', got %q", cmd.Use)
	}

	if !strings.Contains(cmd.Short, "Manage Perl") {
		t.Errorf("Expected command short description to mention Perl management, got %q", cmd.Short)
	}

	// Test subcommands exist
	subCmdNames := []string{}
	for _, subCmd := range cmd.Commands() {
		subCmdNames = append(subCmdNames, subCmd.Name())
	}

	expectedSubCmds := []string{"system", "build", "tarball"}
	for _, expected := range expectedSubCmds {
		found := false
		for _, actual := range subCmdNames {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q to be present, but it was not found in %v", expected, subCmdNames)
		}
	}
}

func TestPerlBuildCommandStructure(t *testing.T) {
	cmd := newPerlBuildCommand()

	// Verify the command structure
	if cmd.Use != "build [version|URL]" {
		t.Errorf("Expected command use to be 'build [version|URL]', got %q", cmd.Use)
	}

	if !strings.Contains(cmd.Short, "Build Perl") {
		t.Errorf("Expected command short description to mention Build Perl, got %q", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "URL") {
		t.Errorf("Expected command long description to mention URL support, got %q", cmd.Long)
	}

	// Verify required flags exist
	requiredFlags := []string{
		"source", "prefix", "output-dir", "jobs", "test", "cleanup",
		"build-only", "configure-options", "relocatable", "shared-lib",
		"upload", "platforms", "mirror", "github-token", "github-repo",
		"release-tag", "draft-release", "prerelease",
	}

	for _, flagName := range requiredFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %q to exist", flagName)
		}
	}
}

func TestPerlBuildCommandFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "basic build with version",
			args:        []string{"5.38.0"},
			expectError: false,
		},
		{
			name:        "build with URL",
			args:        []string{"https://example.com/perl-5.38.0.tar.gz"},
			expectError: false,
		},
		{
			name:        "build with source flag",
			args:        []string{"5.38.0", "--source", "/path/to/source.tar.gz"},
			expectError: false,
		},
		{
			name:        "build with prefix flag",
			args:        []string{"5.38.0", "--prefix", "/custom/install/path"},
			expectError: false,
		},
		{
			name:        "build with test flag",
			args:        []string{"5.38.0", "--test"},
			expectError: false,
		},
		{
			name:        "build with relocatable flag",
			args:        []string{"5.38.0", "--relocatable"},
			expectError: false,
		},
		{
			name:        "build with upload flags",
			args:        []string{"5.38.0", "--upload", "--github-token", "token", "--github-repo", "owner/repo"},
			expectError: false,
		},
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 0",
		},
		{
			name:        "too many arguments",
			args:        []string{"5.38.0", "extra"},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newPerlBuildCommand()

			// Mock RunE to avoid actual execution
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				return nil
			}

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestPerlTarballCommandStructure(t *testing.T) {
	cmd := newPerlTarballCommand()

	// Verify the command structure
	if cmd.Use != "tarball [version]" {
		t.Errorf("Expected command use to be 'tarball [version]', got %q", cmd.Use)
	}

	if !strings.Contains(cmd.Short, "Create tarball") {
		t.Errorf("Expected command short description to mention Create tarball, got %q", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "Build Perl Binaries workflow") {
		t.Errorf("Expected command long description to mention workflow issues, got %q", cmd.Long)
	}

	// Verify required flags exist
	requiredFlags := []string{"output", "compression-level", "exclude", "verify"}
	for _, flagName := range requiredFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %q to exist", flagName)
		}
	}

	// Verify default values
	compressionLevel, _ := cmd.Flags().GetInt("compression-level")
	if compressionLevel != 6 {
		t.Errorf("Expected default compression-level to be 6, got %d", compressionLevel)
	}

	verify, _ := cmd.Flags().GetBool("verify")
	if !verify {
		t.Errorf("Expected default verify to be true, got %v", verify)
	}

	excludePatterns, _ := cmd.Flags().GetStringArray("exclude")
	expectedPatterns := []string{"*.log", "*.tmp", ".pvm-*"}
	if len(excludePatterns) != len(expectedPatterns) {
		t.Errorf("Expected %d default exclude patterns, got %d", len(expectedPatterns), len(excludePatterns))
	}
}

func TestPerlTarballCommandFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "basic tarball creation",
			args:        []string{"5.38.0"},
			expectError: false,
		},
		{
			name:        "tarball with custom output",
			args:        []string{"5.38.0", "--output", "custom-perl.tar.gz"},
			expectError: false,
		},
		{
			name:        "tarball with compression level",
			args:        []string{"5.38.0", "--compression-level", "9"},
			expectError: false,
		},
		{
			name:        "tarball with custom exclude patterns",
			args:        []string{"5.38.0", "--exclude", "*.tmp", "--exclude", "*.log"},
			expectError: false,
		},
		{
			name:        "tarball without verification",
			args:        []string{"5.38.0", "--verify=false"},
			expectError: false,
		},
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 0",
		},
		{
			name:        "too many arguments",
			args:        []string{"5.38.0", "extra"},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newPerlTarballCommand()

			// Mock RunE to avoid actual execution
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				return nil
			}

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestCreateEnhancedTarGzArchive(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "pvm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]string{
		"bin/perl":          "#!/usr/bin/perl\nprint 'hello';\n",
		"lib/perl5/Test.pm": "package Test;\n1;\n",
		"man/man1/perl.1":   ".TH PERL 1\nperl manual\n",
		"test.log":          "log file content", // Should be excluded
		"cache/temp.tmp":    "temporary file",   // Should be excluded
		".pvm-metadata":     "metadata",         // Should be excluded
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0o755)
		if err != nil {
			t.Fatalf("Failed to create directory for %s: %v", filePath, err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0o644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filePath, err)
		}
	}

	// Create output tarball
	outputPath := filepath.Join(tempDir, "test.tar.gz")
	excludePatterns := []string{"*.log", "*.tmp", ".pvm-*"}

	// Create a real UI for testing, but discard output
	ctx := &ui.UIContext{
		Writer: os.Stdout,
		Quiet:  true, // Suppress output during testing
	}
	mockUI := ui.NewOutput(ctx)

	err = createEnhancedTarGzArchive(tempDir, outputPath, 6, excludePatterns, mockUI)
	if err != nil {
		t.Fatalf("Failed to create tarball: %v", err)
	}

	// Verify the tarball was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Tarball was not created at %s", outputPath)
	}

	// Verify tarball is not empty
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat tarball: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("Tarball is empty")
	}

	t.Logf("Successfully created tarball of size %d bytes", info.Size())
}

// mockUIOutput functions removed - using real UI for testing

// containsString function is already defined in command_test.go

func TestPerlExecAllCommandStructure(t *testing.T) {
	cmd := newPerlExecAllCommand()

	// Verify the command structure
	if cmd.Use != "exec-all [command...]" {
		t.Errorf("Expected command use to be 'exec-all [command...]', got %q", cmd.Use)
	}

	if !strings.Contains(cmd.Short, "Execute a command with all installed Perl versions") {
		t.Errorf("Expected command short description to mention exec-all functionality, got %q", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "sequentially") {
		t.Errorf("Expected command long description to mention sequential execution, got %q", cmd.Long)
	}

	if !strings.Contains(cmd.Long, "Examples:") {
		t.Errorf("Expected command long description to include examples, got %q", cmd.Long)
	}
}

func TestPerlExecAllCommandFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "basic command execution",
			args:        []string{"perl", "-v"},
			expectError: false,
		},
		{
			name:        "prove command",
			args:        []string{"prove", "t/basic.t"},
			expectError: false,
		},
		{
			name:        "perl one-liner",
			args:        []string{"perl", "-e", "print 'test'"},
			expectError: false,
		},
		{
			name:        "complex command with flags",
			args:        []string{"perl", "-MData::Dumper", "-e", "print $Data::Dumper::VERSION"},
			expectError: false,
		},
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
			errorMsg:    "no command specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newPerlExecAllCommand()

			// Mock the execution to avoid actually running commands
			originalRunE := cmd.RunE
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				if len(args) == 0 {
					return fmt.Errorf("no command specified")
				}
				return nil
			}
			defer func() { cmd.RunE = originalRunE }()

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestExecAllResultStruct(t *testing.T) {
	// Test the ExecAllResult struct creation and usage
	result := ExecAllResult{
		Version:  "5.38.0",
		ExitCode: 0,
		Output:   "test output",
		Error:    "",
		Duration: 100000000, // 100ms in nanoseconds
	}

	if result.Version != "5.38.0" {
		t.Errorf("Expected version '5.38.0', got %q", result.Version)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if result.Output != "test output" {
		t.Errorf("Expected output 'test output', got %q", result.Output)
	}

	if result.Duration != 100000000 {
		t.Errorf("Expected duration 100000000, got %d", result.Duration)
	}
}

func TestDisplayExecAllSummary(t *testing.T) {
	// Create test results
	results := []ExecAllResult{
		{Version: "5.42.0", ExitCode: 0, Output: "success"},
		{Version: "5.40.0", ExitCode: 0, Output: "success"},
		{Version: "5.38.0", ExitCode: 1, Output: "failure", Error: "test error"},
	}

	// Create a UI context for testing
	ctx := &ui.UIContext{
		Writer: os.Stdout,
		Quiet:  true, // Suppress output during testing
	}
	mockUI := ui.NewOutput(ctx)

	// This should not panic or error
	displayExecAllSummary(results, mockUI)

	// Test with all successes
	successResults := []ExecAllResult{
		{Version: "5.42.0", ExitCode: 0, Output: "success"},
		{Version: "5.40.0", ExitCode: 0, Output: "success"},
	}

	displayExecAllSummary(successResults, mockUI)

	// Test with all failures
	failureResults := []ExecAllResult{
		{Version: "5.42.0", ExitCode: 1, Output: "", Error: "failed"},
		{Version: "5.40.0", ExitCode: 2, Output: "", Error: "failed"},
	}

	displayExecAllSummary(failureResults, mockUI)
}

func TestDetermineExitCode(t *testing.T) {
	// Test all successes - should not exit
	successResults := []ExecAllResult{
		{Version: "5.42.0", ExitCode: 0},
		{Version: "5.40.0", ExitCode: 0},
	}

	err := determineExitCode(successResults)
	if err != nil {
		t.Errorf("Expected no error for all successes, got %v", err)
	}

	// Note: Testing actual exit behavior would terminate the test process,
	// so we can only test the success case here. The exit behavior is
	// verified through integration tests.
}

func TestIsLikelyVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Version patterns
		{"5.38.0", true},
		{"5.40.1", true},
		{"5.42", true},
		{"@latest", true},
		{"@stable", true},
		{"system", true},
		{"latest", true},

		// Command patterns
		{"perl", false},
		{"cpan", false},
		{"prove", false},
		{"perldoc", false},
		{"cpanm", false},

		// Ambiguous cases
		{"test", false}, // Not in common commands, but also not version-like
		{"5", false},    // Too short to be a clear version
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isLikelyVersion(tt.input)
			if result != tt.expected {
				t.Errorf("isLikelyVersion(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
