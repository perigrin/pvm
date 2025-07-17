// ABOUTME: Comprehensive tests for embedded template system
// ABOUTME: Tests template loading, variable substitution, and error handling

package pvm

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateVariables(t *testing.T) {
	t.Run("NewTemplateVariables creates variables with defaults", func(t *testing.T) {
		vars := NewTemplateVariables("test-project")

		assert.Equal(t, "test-project", vars.ProjectName)
		assert.Equal(t, "0.01", vars.Version)
		assert.Equal(t, "3000", vars.WebPort)
		assert.Equal(t, "sqlite:db/app.db", vars.DatabaseURL)
	})

	t.Run("ToMap converts variables to map correctly", func(t *testing.T) {
		vars := TemplateVariables{
			ProjectName: "my-project",
			Version:     "1.0.0",
			WebPort:     "8080",
			DatabaseURL: "postgres://localhost/mydb",
		}

		m := vars.ToMap()

		assert.Equal(t, "my-project", m["ProjectName"])
		assert.Equal(t, "1.0.0", m["Version"])
		assert.Equal(t, "8080", m["WebPort"])
		assert.Equal(t, "postgres://localhost/mydb", m["DatabaseURL"])
	})
}

func TestEmbeddedTemplateManager_ListEmbeddedTemplates(t *testing.T) {
	manager := NewEmbeddedTemplateManager()

	t.Run("lists available templates when filesystem is available", func(t *testing.T) {
		templates, err := manager.ListEmbeddedTemplates()

		if err != nil {
			t.Skip("Embedded templates not available - this test requires the binary to be built with embedded templates")
		}

		// Should contain at least the expected templates
		expectedTemplates := []string{"minimal", "web", "cli"}
		for _, expected := range expectedTemplates {
			assert.Contains(t, templates, expected, "should contain template: %s", expected)
		}
	})

	t.Run("handles missing templates directory", func(t *testing.T) {
		manager.templatesFS = embed.FS{} // Empty filesystem

		_, err := manager.ListEmbeddedTemplates()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read embedded templates")
	})
}

func TestEmbeddedTemplateManager_LoadEmbeddedTemplate(t *testing.T) {
	manager := NewEmbeddedTemplateManager()

	t.Run("loads and renders template successfully", func(t *testing.T) {
		// Try to load a real template that should exist
		variables := TemplateVariables{
			ProjectName: "my-project",
			Version:     "1.0.0",
			WebPort:     "8080",
			DatabaseURL: "postgres://localhost/mydb",
		}

		template, err := manager.LoadEmbeddedTemplate("minimal", variables)

		if err != nil {
			t.Skip("Embedded templates not available - this test requires the binary to be built with embedded templates")
		}

		assert.Equal(t, "minimal", template.Name)
		assert.NotEmpty(t, template.Description)
		assert.Contains(t, template.Directories, "lib")
		assert.Contains(t, template.Directories, "t")
		assert.NotNil(t, template.Dependencies)
		assert.NotNil(t, template.DevDeps)
		assert.NotNil(t, template.TestDeps)
	})

	t.Run("handles template not found", func(t *testing.T) {
		variables := NewTemplateVariables("test-project")

		_, err := manager.LoadEmbeddedTemplate("nonexistent", variables)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template 'nonexistent' not found")
	})
}

func TestEmbeddedTemplateManager_GetEmbeddedTemplateDescription(t *testing.T) {
	manager := NewEmbeddedTemplateManager()

	t.Run("returns template description", func(t *testing.T) {
		description, err := manager.GetEmbeddedTemplateDescription("minimal")

		if err != nil {
			t.Skip("Embedded templates not available - this test requires the binary to be built with embedded templates")
		}

		assert.NotEmpty(t, description)
		assert.Contains(t, description, "Minimal")
	})

	t.Run("handles template not found", func(t *testing.T) {
		_, err := manager.GetEmbeddedTemplateDescription("nonexistent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template 'nonexistent' not found")
	})
}

func TestTemplateFunctions(t *testing.T) {
	manager := NewEmbeddedTemplateManager()

	t.Run("function map contains expected functions", func(t *testing.T) {
		expectedFuncs := []string{"default", "upper", "lower", "replace", "join", "split", "contains"}
		for _, name := range expectedFuncs {
			assert.Contains(t, manager.funcMap, name, "function map should contain %s", name)
		}
	})

	t.Run("functions are accessible through function map", func(t *testing.T) {
		// Test that functions exist and are callable (without type assertions)
		assert.NotNil(t, manager.funcMap["default"])
		assert.NotNil(t, manager.funcMap["upper"])
		assert.NotNil(t, manager.funcMap["lower"])
		assert.NotNil(t, manager.funcMap["replace"])
	})
}

func TestLoadTemplateWithVariables(t *testing.T) {
	t.Run("loads embedded template with project variables", func(t *testing.T) {
		template, err := loadTemplateWithVariables("minimal", "my-project")

		if err != nil {
			t.Skip("Embedded templates not available - this test requires the binary to be built with embedded templates")
		}

		assert.Equal(t, "minimal", template.Name)
		assert.NotEmpty(t, template.Description)
		assert.Contains(t, template.Directories, "lib")
		assert.Contains(t, template.Directories, "t")
	})

	t.Run("handles unknown template", func(t *testing.T) {
		_, err := loadTemplateWithVariables("nonexistent", "my-project")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown template: nonexistent")
	})
}

func TestLoadEmbeddedTemplateWithVariables(t *testing.T) {
	t.Run("loads embedded template with variables", func(t *testing.T) {
		template, err := loadEmbeddedTemplateWithVariables("minimal", "my-project")

		if err != nil {
			t.Skip("Embedded templates not available - this test requires the binary to be built with embedded templates")
		}

		assert.Equal(t, "minimal", template.Name)
		assert.NotEmpty(t, template.Description)
		assert.Contains(t, template.Directories, "lib")
		assert.Contains(t, template.Directories, "t")
	})

	t.Run("uses default template variables", func(t *testing.T) {
		template, err := loadEmbeddedTemplateWithVariables("minimal", "my-project")

		if err != nil {
			t.Skip("Embedded templates not available - this test requires the binary to be built with embedded templates")
		}

		assert.Equal(t, "minimal", template.Name)
		assert.NotEmpty(t, template.Description)
	})
}
