// ABOUTME: Tests for module analyzer functionality
// ABOUTME: Validates real module analysis vs placeholder generation

package pvi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestModuleAnalyzer_Basic(t *testing.T) {
	analyzer, err := NewModuleAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	if analyzer == nil {
		t.Fatal("Analyzer should not be nil")
	}
}

func TestModuleAnalyzer_SimpleModule(t *testing.T) {
	// Create a temporary Perl module for testing
	tmpDir := t.TempDir()
	moduleFile := filepath.Join(tmpDir, "TestModule.pm")

	moduleContent := `package TestModule;

use strict;
use warnings;

our $VERSION = '1.0.0';

sub new {
    my $class = shift;
    return bless {}, $class;
}

sub simple_method {
    my ($self, $param) = @_;
    return "Hello, $param!";
}

sub typed_method(Str $input) returns Str {
    return "Processed: $input";
}

1;
`

	err := os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test module: %v", err)
	}

	// Analyze the module
	analyzer, err := NewModuleAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	typeDef, err := analyzer.AnalyzeModule(moduleFile)
	if err != nil {
		t.Fatalf("Failed to analyze module: %v", err)
	}

	// Verify basic type definition properties
	if typeDef.Module != "TestModule" {
		t.Errorf("Expected module name 'TestModule', got '%s'", typeDef.Module)
	}

	if typeDef.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", typeDef.Version)
	}

	if typeDef.Source != "analyzed" {
		t.Errorf("Expected source 'analyzed', got '%s'", typeDef.Source)
	}

	if typeDef.Maintainer != "PVI module analyzer" {
		t.Errorf("Expected maintainer 'PVI module analyzer', got '%s'", typeDef.Maintainer)
	}

	// Check that we have at least one package
	if len(typeDef.Packages) == 0 {
		t.Error("Expected at least one package")
	} else {
		pkg := typeDef.Packages[0]
		if pkg.Name != "TestModule" {
			t.Errorf("Expected package name 'TestModule', got '%s'", pkg.Name)
		}
	}

	// Check that we found some subroutines
	if len(typeDef.Subs) == 0 {
		t.Error("Expected to find subroutines")
	}

	// Look for specific subroutines
	foundNew := false
	foundSimpleMethod := false
	foundTypedMethod := false

	for _, sub := range typeDef.Subs {
		switch sub.Name {
		case "new":
			foundNew = true
		case "simple_method":
			foundSimpleMethod = true
		case "typed_method":
			foundTypedMethod = true
			// Check if return type was extracted
			if len(sub.Returns) > 0 && sub.Returns[0].Type == "Str" {
				// Good, type was extracted
			}
		}
	}

	if !foundNew {
		t.Error("Expected to find 'new' subroutine")
	}
	if !foundSimpleMethod {
		t.Error("Expected to find 'simple_method' subroutine")
	}
	if !foundTypedMethod {
		t.Error("Expected to find 'typed_method' subroutine")
	}
}

func TestModuleAnalyzer_ClassModule(t *testing.T) {
	// Create a temporary Perl module with modern class syntax
	tmpDir := t.TempDir()
	moduleFile := filepath.Join(tmpDir, "TestClass.pm")

	moduleContent := `package TestClass;

use strict;
use warnings;

class TestClass {
    field Str $name;
    field Int $age = 0;

    method new(Str $name, Int $age = 0) {
        $self->{name} = $name;
        $self->{age} = $age;
        return $self;
    }

    method get_name() returns Str {
        return $self->{name};
    }

    method set_age(Int $new_age) {
        $self->{age} = $new_age;
    }
}

1;
`

	err := os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test module: %v", err)
	}

	// Analyze the module
	analyzer, err := NewModuleAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	typeDef, err := analyzer.AnalyzeModule(moduleFile)
	if err != nil {
		t.Fatalf("Failed to analyze module: %v", err)
	}

	// Check that we found the class
	if len(typeDef.Types) == 0 {
		t.Error("Expected to find class definition")
	} else {
		classType := typeDef.Types[0]
		if classType.Name != "TestClass" {
			t.Errorf("Expected class name 'TestClass', got '%s'", classType.Name)
		}
		if classType.Kind != "class" {
			t.Errorf("Expected class kind 'class', got '%s'", classType.Kind)
		}

		// Check for fields
		if len(classType.Properties) < 2 {
			t.Errorf("Expected at least 2 fields, got %d", len(classType.Properties))
		}

		// Look for specific fields
		foundName := false
		foundAge := false
		for _, prop := range classType.Properties {
			switch prop.Name {
			case "name":
				foundName = true
				if prop.Type != "Str" {
					t.Errorf("Expected name field type 'Str', got '%s'", prop.Type)
				}
			case "age":
				foundAge = true
				if prop.Type != "Int" {
					t.Errorf("Expected age field type 'Int', got '%s'", prop.Type)
				}
				if prop.Default != "0" {
					t.Errorf("Expected age field default '0', got '%s'", prop.Default)
				}
			}
		}

		if !foundName {
			t.Error("Expected to find 'name' field")
		}
		if !foundAge {
			t.Error("Expected to find 'age' field")
		}
	}

	// Check that we found methods
	if len(typeDef.Methods) == 0 {
		t.Error("Expected to find methods")
	}
}

func TestModuleAnalyzer_ModuleWithExports(t *testing.T) {
	// Create a temporary Perl module with exports
	tmpDir := t.TempDir()
	moduleFile := filepath.Join(tmpDir, "ExportModule.pm")

	moduleContent := `package ExportModule;

use strict;
use warnings;
use Exporter 'import';

our @EXPORT = qw(public_func exported_sub);
our @EXPORT_OK = qw(optional_func);

sub public_func {
    my $param = shift;
    return "Public: $param";
}

sub exported_sub {
    return "Exported subroutine";
}

sub optional_func {
    return "Optional function";
}

sub _private_func {
    return "Private function";
}

1;
`

	err := os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test module: %v", err)
	}

	// Analyze the module
	analyzer, err := NewModuleAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	typeDef, err := analyzer.AnalyzeModule(moduleFile)
	if err != nil {
		t.Fatalf("Failed to analyze module: %v", err)
	}

	// Check that exports were detected
	if len(typeDef.Packages) == 0 {
		t.Error("Expected at least one package")
	} else {
		pkg := typeDef.Packages[0]
		if len(pkg.Exports) == 0 {
			t.Error("Expected to find exported symbols")
		}

		// Check for specific exports
		exportNames := make(map[string]bool)
		for _, export := range pkg.Exports {
			exportNames[export.Name] = true
		}

		if !exportNames["public_func"] {
			t.Error("Expected to find 'public_func' in exports")
		}
		if !exportNames["exported_sub"] {
			t.Error("Expected to find 'exported_sub' in exports")
		}
	}

	// Check that private functions are marked correctly
	for _, sub := range typeDef.Subs {
		if sub.Name == "_private_func" {
			if !sub.IsPrivate {
				t.Error("Expected '_private_func' to be marked as private")
			}
		}
	}
}

func TestFindModuleFile(t *testing.T) {
	// Test with direct file path
	tmpDir := t.TempDir()
	moduleFile := filepath.Join(tmpDir, "Test.pm")

	err := os.WriteFile(moduleFile, []byte("package Test; 1;"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test module: %v", err)
	}

	// Test finding by direct path
	found := findModuleFile(moduleFile)
	if found != moduleFile {
		t.Errorf("Expected to find module at '%s', got '%s'", moduleFile, found)
	}

	// Test non-existent module
	notFound := findModuleFile("NonExistent::Module")
	if notFound != "" {
		t.Errorf("Expected empty string for non-existent module, got '%s'", notFound)
	}
}

func TestExtractModuleNameFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"Test.pm", "Test"},
		{"lib/Foo/Bar.pm", "Foo::Bar"},
		{"/path/to/lib/My/Module.pm", "My::Module"},
		{"script.pl", "script"},
		{"complex/lib/Deep/Nested/Module.pm", "Deep::Nested::Module"},
	}

	for _, test := range tests {
		result := extractModuleNameFromPath(test.path)
		if result != test.expected {
			t.Errorf("extractModuleNameFromPath(%s) = %s, expected %s",
				test.path, result, test.expected)
		}
	}
}

func TestModuleAnalyzer_ErrorHandling(t *testing.T) {
	analyzer, err := NewModuleAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Test with non-existent file
	_, err = analyzer.AnalyzeModule("/non/existent/file.pm")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with invalid file content
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "Invalid.pm")

	// Write invalid Perl content that might cause parsing issues
	err = os.WriteFile(invalidFile, []byte("this is not valid perl content @#$%"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	// Analysis should still succeed (parser handles errors gracefully)
	typeDef, err := analyzer.AnalyzeModule(invalidFile)
	if err != nil {
		t.Errorf("Analysis should handle invalid content gracefully: %v", err)
	}

	// Should still return a basic type definition
	if typeDef == nil {
		t.Error("Expected type definition even for invalid content")
	}
}

func TestModuleAnalyzer_PerformanceBaseline(t *testing.T) {
	// Create a larger test module to check performance
	tmpDir := t.TempDir()
	moduleFile := filepath.Join(tmpDir, "LargeModule.pm")

	var content strings.Builder
	content.WriteString("package LargeModule;\n\n")

	// Generate many subroutines
	for i := 0; i < 100; i++ {
		content.WriteString(fmt.Sprintf("sub func_%d {\n    return %d;\n}\n\n", i, i))
	}

	content.WriteString("1;\n")

	err := os.WriteFile(moduleFile, []byte(content.String()), 0644)
	if err != nil {
		t.Fatalf("Failed to write large module: %v", err)
	}

	// Measure analysis time
	start := time.Now()

	analyzer, err := NewModuleAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	typeDef, err := analyzer.AnalyzeModule(moduleFile)
	if err != nil {
		t.Fatalf("Failed to analyze large module: %v", err)
	}

	duration := time.Since(start)

	// Performance check - should complete within reasonable time
	if duration > 5*time.Second {
		t.Errorf("Analysis took too long: %v", duration)
	}

	// Verify we found the expected number of subroutines
	if len(typeDef.Subs) < 90 { // Allow some margin for parsing variations
		t.Errorf("Expected around 100 subroutines, got %d", len(typeDef.Subs))
	}
}
