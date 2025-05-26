package pvx

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestAutoDetectDependencies(t *testing.T) {
	// Create a test script with various dependencies
	content := `#!/usr/bin/perl

use strict;
use warnings;
use DBI;
use JSON::PP;
use File::Spec; # Core module
use Moose;
use feature 'say';

require HTTP::Tiny;
require "Path/Tiny.pm";

print "Hello, World!\n";
`

	tmpFile := createTempScript(t, content)
	defer os.Remove(tmpFile)

	dependencies, err := AutoDetectDependencies(tmpFile)
	if err != nil {
		t.Fatalf("AutoDetectDependencies failed: %v", err)
	}

	// Expected dependencies (pragmas should be filtered out)
	expected := []string{"DBI", "JSON::PP", "File::Spec", "Moose", "HTTP::Tiny", "Path::Tiny"}
	sort.Strings(dependencies)
	sort.Strings(expected)

	if !reflect.DeepEqual(dependencies, expected) {
		t.Errorf("Dependencies mismatch. Expected %v, got %v", expected, dependencies)
	}
}

func TestExtractDependenciesFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "Basic use statements",
			content: `use strict;
use warnings;
use DBI;
use JSON::PP;`,
			expected: []string{"DBI", "JSON::PP"},
		},
		{
			name: "Use with versions",
			content: `use DBI 1.631;
use JSON::PP 4.0;
use Moose 2.2006;`,
			expected: []string{"DBI", "JSON::PP", "Moose"},
		},
		{
			name: "Require statements",
			content: `require DBI;
require "JSON/PP.pm";
require 'HTTP/Tiny.pm';`,
			expected: []string{"DBI", "JSON::PP", "HTTP::Tiny"},
		},
		{
			name: "Mixed statements with comments",
			content: `# This is a comment
use strict; # pragma
use DBI; # database module
# use Test::More; # commented out
require HTTP::Tiny;`,
			expected: []string{"DBI", "HTTP::Tiny"},
		},
		{
			name: "Pragmas should be filtered",
			content: `use strict;
use warnings;
use utf8;
use feature 'say';
use autodie;
use DBI; # This should be included`,
			expected: []string{"DBI"},
		},
		{
			name: "Complex module names",
			content: `use Test::More::UTF8;
use Moo::Role;
use Type::Tiny::XS;`,
			expected: []string{"Test::More::UTF8", "Moo::Role", "Type::Tiny::XS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDependenciesFromContent(tt.content)
			sort.Strings(got)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("extractDependenciesFromContent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsPragma(t *testing.T) {
	tests := []struct {
		module   string
		expected bool
	}{
		{"strict", true},
		{"warnings", true},
		{"DBI", false},
		{"JSON::PP", false},
		{"feature", true},
		{"utf8", true},
		{"Moose", false},
		{"autodie", true},
		{"constant", true},
		{"Test::More", false},
	}

	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			got := isPragma(tt.module)
			if got != tt.expected {
				t.Errorf("isPragma(%s) = %v, want %v", tt.module, got, tt.expected)
			}
		})
	}
}

func TestFilterCPANModules(t *testing.T) {
	dependencies := []string{
		"DBI",          // CPAN module
		"JSON::PP",     // CPAN module
		"File::Spec",   // Core module
		"Moose",        // CPAN module
		"Carp",         // Core module
		"Data::Dumper", // Core module
		"Test::More",   // CPAN module
		"List::Util",   // Core module
	}

	expected := []string{"DBI", "JSON::PP", "Moose", "Test::More"}
	got := FilterCPANModules(dependencies)

	sort.Strings(got)
	sort.Strings(expected)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("FilterCPANModules() = %v, want %v", got, expected)
	}
}

func TestAutoDetectDependenciesWithOptions(t *testing.T) {
	content := `#!/usr/bin/perl
use strict;
use warnings;
use DBI;
use File::Spec;
use Moose;
use Carp;
`

	tmpFile := createTempScript(t, content)
	defer os.Remove(tmpFile)

	// Test with core modules included
	depsWithCore, err := AutoDetectDependenciesWithOptions(tmpFile, true)
	if err != nil {
		t.Fatalf("AutoDetectDependenciesWithOptions (with core) failed: %v", err)
	}

	expectedWithCore := []string{"DBI", "File::Spec", "Moose", "Carp"}
	sort.Strings(depsWithCore)
	sort.Strings(expectedWithCore)

	if !reflect.DeepEqual(depsWithCore, expectedWithCore) {
		t.Errorf("Dependencies with core modules: expected %v, got %v", expectedWithCore, depsWithCore)
	}

	// Test with core modules filtered out
	depsWithoutCore, err := AutoDetectDependenciesWithOptions(tmpFile, false)
	if err != nil {
		t.Fatalf("AutoDetectDependenciesWithOptions (without core) failed: %v", err)
	}

	expectedWithoutCore := []string{"DBI", "Moose"}
	sort.Strings(depsWithoutCore)
	sort.Strings(expectedWithoutCore)

	if !reflect.DeepEqual(depsWithoutCore, expectedWithoutCore) {
		t.Errorf("Dependencies without core modules: expected %v, got %v", expectedWithoutCore, depsWithoutCore)
	}
}

func TestAutoDetectDependencies_EmptyFile(t *testing.T) {
	content := `#!/usr/bin/perl
# Just a comment
print "Hello, World!\n";
`

	tmpFile := createTempScript(t, content)
	defer os.Remove(tmpFile)

	dependencies, err := AutoDetectDependencies(tmpFile)
	if err != nil {
		t.Fatalf("AutoDetectDependencies failed: %v", err)
	}

	if len(dependencies) != 0 {
		t.Errorf("Expected no dependencies for script without use/require, got %v", dependencies)
	}
}

func TestAutoDetectDependencies_NonexistentFile(t *testing.T) {
	_, err := AutoDetectDependencies("/nonexistent/file.pl")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

// Helper function to create temporary scripts
func createTempScript(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_script.pl")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp script: %v", err)
	}

	return tmpFile
}
