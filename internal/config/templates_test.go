// ABOUTME: Tests for configuration template system
// ABOUTME: Validates template rendering, inheritance, and error handling

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateManager(t *testing.T) {
	// Setup temporary directory
	tempDir := t.TempDir()
	templatesDir := filepath.Join(tempDir, "templates")

	// Create template manager
	tm := NewTemplateManager(templatesDir)

	t.Run("EmptyTemplatesDir", func(t *testing.T) {
		// Should not error when templates directory doesn't exist
		err := tm.LoadTemplates()
		if err != nil {
			t.Errorf("LoadTemplates should not error with missing directory: %v", err)
		}

		templates := tm.ListTemplates()
		if len(templates) != 0 {
			t.Errorf("Expected 0 templates, got %d", len(templates))
		}
	})

	t.Run("BasicTemplate", func(t *testing.T) {
		// Create templates directory
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to create templates directory: %v", err)
		}

		// Create a basic template
		templateContent := `name = "basic"
description = "Basic PVM configuration"
version = "1.0"

content = """
[pvm]
default_perl = "{{.perl_version}}"
build_jobs = {{.build_jobs}}
run_tests = true

[pvx]
isolation_level = "medium"
"""

[variables]
perl_version = "5.38.0"
build_jobs = "4"`

		templatePath := filepath.Join(templatesDir, "basic.template.toml")
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		// Load templates
		if err := tm.LoadTemplates(); err != nil {
			t.Fatalf("Failed to load templates: %v", err)
		}

		// Check template was loaded
		templates := tm.ListTemplates()
		if len(templates) != 1 {
			t.Errorf("Expected 1 template, got %d", len(templates))
		}
		if templates[0] != "basic" {
			t.Errorf("Expected template 'basic', got '%s'", templates[0])
		}

		// Get template
		template, err := tm.GetTemplate("basic")
		if err != nil {
			t.Errorf("Failed to get template: %v", err)
		}
		if template.Name != "basic" {
			t.Errorf("Expected template name 'basic', got '%s'", template.Name)
		}

		// Render template
		variables := map[string]string{
			"perl_version": "5.40.0",
			"build_jobs":   "8",
		}
		config, err := tm.RenderTemplate("basic", variables)
		if err != nil {
			t.Errorf("Failed to render template: %v", err)
		}

		// Verify rendered values
		if config.PVM.DefaultPerl != "5.40.0" {
			t.Errorf("Expected DefaultPerl '5.40.0', got '%s'", config.PVM.DefaultPerl)
		}
		if config.PVM.BuildJobs != 8 {
			t.Errorf("Expected BuildJobs 8, got %d", config.PVM.BuildJobs)
		}
	})

	t.Run("TemplateInheritance", func(t *testing.T) {
		// Create templates directory
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to create templates directory: %v", err)
		}

		// Create parent template
		parentContent := `name = "parent"
description = "Parent template"

content = """
[pvm]
default_perl = "{{.perl_version}}"
run_tests = true

[pvx]
isolation_level = "medium"
"""`

		parentPath := filepath.Join(templatesDir, "parent.template.toml")
		if err := os.WriteFile(parentPath, []byte(parentContent), 0644); err != nil {
			t.Fatalf("Failed to write parent template: %v", err)
		}

		// Create child template that inherits from parent
		childContent := `name = "child"
description = "Child template"
inherits = "parent"

content = """
[pvm]
build_jobs = {{.build_jobs}}

[pvi]
preferred_installer = "cpanm"
"""`

		childPath := filepath.Join(templatesDir, "child.template.toml")
		if err := os.WriteFile(childPath, []byte(childContent), 0644); err != nil {
			t.Fatalf("Failed to write child template: %v", err)
		}

		// Clean up any existing templates to avoid interference
		if err := os.RemoveAll(templatesDir); err != nil {
			t.Fatalf("Failed to clean templates directory: %v", err)
		}
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to recreate templates directory: %v", err)
		}

		// Re-write the parent and child templates
		if err := os.WriteFile(parentPath, []byte(parentContent), 0644); err != nil {
			t.Fatalf("Failed to re-write parent template: %v", err)
		}
		if err := os.WriteFile(childPath, []byte(childContent), 0644); err != nil {
			t.Fatalf("Failed to re-write child template: %v", err)
		}

		// Reload templates
		tm = NewTemplateManager(templatesDir)
		if err := tm.LoadTemplates(); err != nil {
			t.Fatalf("Failed to load templates: %v", err)
		}

		// Render child template
		variables := map[string]string{
			"perl_version": "5.38.2",
			"build_jobs":   "6",
		}
		config, err := tm.RenderTemplate("child", variables)
		if err != nil {
			t.Errorf("Failed to render child template: %v", err)
		}

		// Verify inheritance worked
		if config.PVM.DefaultPerl != "5.38.2" {
			t.Errorf("Expected DefaultPerl '5.38.2' (from parent), got '%s'", config.PVM.DefaultPerl)
		}
		if config.PVM.BuildJobs != 6 {
			t.Errorf("Expected BuildJobs 6 (from child), got %d", config.PVM.BuildJobs)
		}
		if config.PVM.RunTests != true {
			t.Errorf("Expected RunTests true (from parent), got %v", config.PVM.RunTests)
		}
		if config.PVI.PreferredInstaller != "cpanm" {
			t.Errorf("Expected PreferredInstaller 'cpanm' (from child), got '%s'", config.PVI.PreferredInstaller)
		}
	})

	t.Run("TemplateValidation", func(t *testing.T) {
		// Test validation with missing name
		invalidTemplate := &Template{
			Content: "[pvm]\ndefault_perl = \"5.38.0\"",
		}

		errors := tm.ValidateTemplate(invalidTemplate)
		if len(errors) == 0 {
			t.Error("Expected validation errors for template without name")
		}

		// Test validation with missing content
		invalidTemplate = &Template{
			Name: "test",
		}

		errors = tm.ValidateTemplate(invalidTemplate)
		if len(errors) == 0 {
			t.Error("Expected validation errors for template without content")
		}

		// Test validation with invalid TOML content
		invalidTemplate = &Template{
			Name:    "test",
			Content: "[pvm\ninvalid toml",
		}

		errors = tm.ValidateTemplate(invalidTemplate)
		if len(errors) == 0 {
			t.Error("Expected validation errors for template with invalid TOML")
		}
	})

	t.Run("TemplateFunctions", func(t *testing.T) {
		// Create templates directory
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to create templates directory: %v", err)
		}

		// Create template that uses template functions
		templateContent := `name = "functions"
description = "Template with functions"

content = """
[pvm]
default_perl = "{{upper .perl_version}}"
download_mirror = "{{default \"https://default.mirror.com\" .mirror_url}}"

[pvx]
isolation_level = "{{lower .isolation_level}}"
"""`

		templatePath := filepath.Join(templatesDir, "functions.template.toml")
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to write functions template: %v", err)
		}

		// Reload templates
		tm = NewTemplateManager(templatesDir)
		if err := tm.LoadTemplates(); err != nil {
			t.Fatalf("Failed to load templates: %v", err)
		}

		// Render template with functions
		variables := map[string]string{
			"perl_version":    "5.38.0",
			"isolation_level": "HIGH",
			// mirror_url intentionally not set to test default function
		}
		config, err := tm.RenderTemplate("functions", variables)
		if err != nil {
			t.Errorf("Failed to render functions template: %v", err)
		}

		// Verify function results
		if config.PVM.DefaultPerl != "5.38.0" {
			t.Errorf("Expected DefaultPerl '5.38.0' (upper function), got '%s'", config.PVM.DefaultPerl)
		}
		if config.PVM.DownloadMirror != "https://default.mirror.com" {
			t.Errorf("Expected DownloadMirror 'https://default.mirror.com' (default function), got '%s'", config.PVM.DownloadMirror)
		}
		if config.PVX.IsolationLevel != "high" {
			t.Errorf("Expected IsolationLevel 'high' (lower function), got '%s'", config.PVX.IsolationLevel)
		}
	})

	t.Run("SaveAndDeleteTemplate", func(t *testing.T) {
		// Create a new template
		newTemplate := &Template{
			Name:        "test-save",
			Description: "Test save template",
			Version:     "1.0",
			Variables: map[string]string{
				"test_var": "default_value",
			},
			Content: "[pvm]\ndefault_perl = \"{{.test_var}}\"",
		}

		// Save template
		if err := tm.SaveTemplate(newTemplate); err != nil {
			t.Errorf("Failed to save template: %v", err)
		}

		// Verify template was saved and loaded
		if err := tm.LoadTemplates(); err != nil {
			t.Errorf("Failed to reload templates: %v", err)
		}

		savedTemplate, err := tm.GetTemplate("test-save")
		if err != nil {
			t.Errorf("Failed to get saved template: %v", err)
		}
		if savedTemplate.Name != "test-save" {
			t.Errorf("Expected template name 'test-save', got '%s'", savedTemplate.Name)
		}

		// Delete template
		if err := tm.DeleteTemplate("test-save"); err != nil {
			t.Errorf("Failed to delete template: %v", err)
		}

		// Verify template was deleted
		if _, err := tm.GetTemplate("test-save"); err == nil {
			t.Error("Expected error when getting deleted template")
		}
	})

	t.Run("ErrorCases", func(t *testing.T) {
		// Create templates directory
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to create templates directory: %v", err)
		}

		// Test getting non-existent template
		if _, err := tm.GetTemplate("non-existent"); err == nil {
			t.Error("Expected error when getting non-existent template")
		}

		// Test rendering non-existent template
		if _, err := tm.RenderTemplate("non-existent", nil); err == nil {
			t.Error("Expected error when rendering non-existent template")
		}

		// Test circular inheritance
		circular1Content := `name = "circular1"
inherits = "circular2"
content = "[pvm]"`

		circular2Content := `name = "circular2"
inherits = "circular1"
content = "[pvx]"`

		circular1Path := filepath.Join(templatesDir, "circular1.template.toml")
		circular2Path := filepath.Join(templatesDir, "circular2.template.toml")

		if err := os.WriteFile(circular1Path, []byte(circular1Content), 0644); err != nil {
			t.Fatalf("Failed to write circular1 template: %v", err)
		}
		if err := os.WriteFile(circular2Path, []byte(circular2Content), 0644); err != nil {
			t.Fatalf("Failed to write circular2 template: %v", err)
		}

		// Reload templates
		tm = NewTemplateManager(templatesDir)
		if err := tm.LoadTemplates(); err != nil {
			t.Fatalf("Failed to load templates: %v", err)
		}

		// Try to render circular template (should eventually error)
		if _, err := tm.RenderTemplate("circular1", nil); err == nil {
			t.Error("Expected error when rendering template with circular inheritance")
		}
	})
}
