// ABOUTME: Tests for Perl script metadata parsing in the pvx package
// ABOUTME: Validates POD-block and comment-block metadata formats, dependency extraction, and format serialization

package pvx

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseScriptMetadata_PODFormat(t *testing.T) {
	// Create a temporary script with POD metadata
	content := `#!/usr/bin/perl

=begin pvm
dependencies = [
  "DBI",
  "JSON::PP >= 4.0",
  "File::Spec"
]
perl_version = "5.30"
description = "Test script with dependencies"
=end pvm

use strict;
use warnings;
use DBI;
use JSON::PP;

print "Hello, World!\n";
`

	tmpFile := createTempScript2(t, content)
	defer os.Remove(tmpFile)

	metadata, err := ParseScriptMetadata(tmpFile)
	if err != nil {
		t.Fatalf("ParseScriptMetadata failed: %v", err)
	}

	expectedDeps := []string{"DBI", "JSON::PP >= 4.0", "File::Spec"}
	if !reflect.DeepEqual(metadata.Dependencies, expectedDeps) {
		t.Errorf("Dependencies mismatch. Expected %v, got %v", expectedDeps, metadata.Dependencies)
	}

	if metadata.PerlVersion != "5.30" {
		t.Errorf("PerlVersion mismatch. Expected '5.30', got '%s'", metadata.PerlVersion)
	}

	if metadata.Description != "Test script with dependencies" {
		t.Errorf("Description mismatch. Expected 'Test script with dependencies', got '%s'", metadata.Description)
	}
}

func TestParseScriptMetadata_CommentFormat(t *testing.T) {
	// Create a temporary script with comment metadata
	content := `#!/usr/bin/perl

# /// pvm
# dependencies = [
#   "DBI",
#   "Moose",
#   "Test::More >= 1.3"
# ]
# perl_version = "5.32"
# ///

use strict;
use warnings;
use DBI;
use Moose;

print "Hello, World!\n";
`

	tmpFile := createTempScript2(t, content)
	defer os.Remove(tmpFile)

	metadata, err := ParseScriptMetadata(tmpFile)
	if err != nil {
		t.Fatalf("ParseScriptMetadata failed: %v", err)
	}

	expectedDeps := []string{"DBI", "Moose", "Test::More >= 1.3"}
	if !reflect.DeepEqual(metadata.Dependencies, expectedDeps) {
		t.Errorf("Dependencies mismatch. Expected %v, got %v", expectedDeps, metadata.Dependencies)
	}

	if metadata.PerlVersion != "5.32" {
		t.Errorf("PerlVersion mismatch. Expected '5.32', got '%s'", metadata.PerlVersion)
	}
}

func TestParseScriptMetadata_SingleLineDependencies(t *testing.T) {
	// Create a temporary script with single-line dependencies
	content := `#!/usr/bin/perl

=begin pvm
dependencies = ["DBI", "JSON::PP", "File::Spec"]
perl_version = "5.30"
=end pvm

use strict;
use warnings;

print "Hello, World!\n";
`

	tmpFile := createTempScript2(t, content)
	defer os.Remove(tmpFile)

	metadata, err := ParseScriptMetadata(tmpFile)
	if err != nil {
		t.Fatalf("ParseScriptMetadata failed: %v", err)
	}

	expectedDeps := []string{"DBI", "JSON::PP", "File::Spec"}
	if !reflect.DeepEqual(metadata.Dependencies, expectedDeps) {
		t.Errorf("Dependencies mismatch. Expected %v, got %v", expectedDeps, metadata.Dependencies)
	}
}

func TestParseScriptMetadata_NoMetadata(t *testing.T) {
	// Create a temporary script without metadata
	content := `#!/usr/bin/perl

use strict;
use warnings;
use DBI;
use JSON::PP;

print "Hello, World!\n";
`

	tmpFile := createTempScript2(t, content)
	defer os.Remove(tmpFile)

	metadata, err := ParseScriptMetadata(tmpFile)
	if err != nil {
		t.Fatalf("ParseScriptMetadata failed: %v", err)
	}

	if len(metadata.Dependencies) != 0 {
		t.Errorf("Expected no dependencies, got %v", metadata.Dependencies)
	}

	if metadata.PerlVersion != "" {
		t.Errorf("Expected no perl version, got '%s'", metadata.PerlVersion)
	}
}

func TestHasMetadata(t *testing.T) {
	// Script with metadata
	contentWithMetadata := `#!/usr/bin/perl

=begin pvm
dependencies = ["DBI"]
=end pvm

use strict;
print "Hello!\n";
`

	tmpFileWithMetadata := createTempScript2(t, contentWithMetadata)
	defer os.Remove(tmpFileWithMetadata)

	if !HasMetadata(tmpFileWithMetadata) {
		t.Error("Expected HasMetadata to return true for script with metadata")
	}

	// Script without metadata
	contentWithoutMetadata := `#!/usr/bin/perl
use strict;
print "Hello!\n";
`

	tmpFileWithoutMetadata := createTempScript2(t, contentWithoutMetadata)
	defer os.Remove(tmpFileWithoutMetadata)

	if HasMetadata(tmpFileWithoutMetadata) {
		t.Error("Expected HasMetadata to return false for script without metadata")
	}
}

func TestFormatMetadataAsPOD(t *testing.T) {
	metadata := &ScriptMetadata{
		Dependencies: []string{"DBI", "JSON::PP >= 4.0"},
		PerlVersion:  "5.30",
		Description:  "Test script",
	}

	formatted := FormatMetadataAsPOD(metadata)

	expectedSubstrings := []string{
		"=begin pvm",
		"=end pvm",
		`"DBI",`,
		`"JSON::PP >= 4.0",`,
		`perl_version = "5.30"`,
		`description = "Test script"`,
	}

	for _, substr := range expectedSubstrings {
		if !containsString(formatted, substr) {
			t.Errorf("Expected formatted POD to contain '%s', but it didn't. Got:\n%s", substr, formatted)
		}
	}
}

func TestFormatMetadataAsComments(t *testing.T) {
	metadata := &ScriptMetadata{
		Dependencies: []string{"DBI", "Moose"},
		PerlVersion:  "5.32",
	}

	formatted := FormatMetadataAsComments(metadata)

	expectedSubstrings := []string{
		"# /// pvm",
		"# ///",
		`#   "DBI",`,
		`#   "Moose",`,
		`# perl_version = "5.32"`,
	}

	for _, substr := range expectedSubstrings {
		if !containsString(formatted, substr) {
			t.Errorf("Expected formatted comments to contain '%s', but it didn't. Got:\n%s", substr, formatted)
		}
	}
}

func TestParseScriptMetadata_ComplexVersionConstraints(t *testing.T) {
	// Test various version constraint formats
	content := `#!/usr/bin/perl

=begin pvm
dependencies = [
  "DBI >= 1.6",
  "JSON::PP < 5.0, >= 4.0",
  "Moose ~> 2.2",
  "File::Spec"
]
=end pvm

use strict;
`

	tmpFile := createTempScript2(t, content)
	defer os.Remove(tmpFile)

	metadata, err := ParseScriptMetadata(tmpFile)
	if err != nil {
		t.Fatalf("ParseScriptMetadata failed: %v", err)
	}

	expectedDeps := []string{
		"DBI >= 1.6",
		"JSON::PP < 5.0, >= 4.0",
		"Moose ~> 2.2",
		"File::Spec",
	}

	if !reflect.DeepEqual(metadata.Dependencies, expectedDeps) {
		t.Errorf("Dependencies mismatch. Expected %v, got %v", expectedDeps, metadata.Dependencies)
	}
}

// Helper functions

// Helper functions
func createTempScript2(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_script.pl")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp script: %v", err)
	}

	return tmpFile
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
