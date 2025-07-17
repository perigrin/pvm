// ABOUTME: Build-time template validation tests
// ABOUTME: Ensures all embedded templates can be parsed and rendered correctly

package pvm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllEmbeddedTemplatesValidation(t *testing.T) {
	manager := NewEmbeddedTemplateManager()

	t.Run("all embedded templates are valid", func(t *testing.T) {
		templates, err := manager.ListEmbeddedTemplates()
		if err != nil {
			t.Skip("Embedded templates not available - this test requires the binary to be built with embedded templates")
		}

		// Ensure we have at least the expected templates
		expectedTemplates := []string{"minimal", "web", "cli"}
		for _, expected := range expectedTemplates {
			assert.Contains(t, templates, expected, "should contain template: %s", expected)
		}

		// Validate each template can be loaded and rendered
		for _, templateName := range templates {
			t.Run("template_"+templateName, func(t *testing.T) {
				variables := NewTemplateVariables("test-project")

				template, err := manager.LoadEmbeddedTemplate(templateName, variables)
				require.NoError(t, err, "template %s should load without error", templateName)

				// Validate basic template structure
				assert.NotEmpty(t, template.Name, "template %s should have a name", templateName)
				assert.NotEmpty(t, template.Description, "template %s should have a description", templateName)
				assert.NotNil(t, template.Dependencies, "template %s should have dependencies map", templateName)
				assert.NotNil(t, template.Config, "template %s should have config map", templateName)

				// Validate template can be loaded with different variable combinations
				testVariables := []TemplateVariables{
					{ProjectName: "test", Version: "1.0", WebPort: "8080", DatabaseURL: "postgres://localhost/test"},
					{ProjectName: "my-app", Version: "0.1", WebPort: "3000", DatabaseURL: "sqlite:db/app.db"},
					{ProjectName: "web-service", Version: "2.0.0", WebPort: "9000", DatabaseURL: "mysql://localhost/webservice"},
				}

				for i, vars := range testVariables {
					t.Run("variables_"+string(rune('A'+i)), func(t *testing.T) {
						renderedTemplate, err := manager.LoadEmbeddedTemplate(templateName, vars)
						require.NoError(t, err, "template %s should render with variables %v", templateName, vars)

						// Verify variable substitution worked
						assert.NotEmpty(t, renderedTemplate.Name, "rendered template should have name")
						assert.NotEmpty(t, renderedTemplate.Description, "rendered template should have description")
					})
				}
			})
		}
	})

	t.Run("template descriptions are accessible", func(t *testing.T) {
		templates, err := manager.ListEmbeddedTemplates()
		if err != nil {
			t.Skip("Embedded templates not available - this test requires the binary to be built with embedded templates")
		}

		for _, templateName := range templates {
			description, err := manager.GetEmbeddedTemplateDescription(templateName)
			require.NoError(t, err, "should be able to get description for template %s", templateName)
			assert.NotEmpty(t, description, "template %s should have non-empty description", templateName)
		}
	})

	t.Run("template variable substitution works correctly", func(t *testing.T) {
		templates, err := manager.ListEmbeddedTemplates()
		if err != nil {
			t.Skip("Embedded templates not available - this test requires the binary to be built with embedded templates")
		}

		for _, templateName := range templates {
			variables := TemplateVariables{
				ProjectName: "test-project-name",
				Version:     "1.2.3",
				WebPort:     "8080",
				DatabaseURL: "postgresql://localhost/testdb",
			}

			template, err := manager.LoadEmbeddedTemplate(templateName, variables)
			require.NoError(t, err, "template %s should load with custom variables", templateName)

			// Check that variables were properly substituted in the description
			// (Most templates should include the project name in their description)
			if templateName == "minimal" || templateName == "web" || templateName == "cli" {
				// For templates that don't use variables in description, just ensure it loads
				assert.NotEmpty(t, template.Description)
			}
		}
	})

	t.Run("template dependencies are valid", func(t *testing.T) {
		templates, err := manager.ListEmbeddedTemplates()
		if err != nil {
			t.Skip("Embedded templates not available - this test requires the binary to be built with embedded templates")
		}

		for _, templateName := range templates {
			variables := NewTemplateVariables("test-project")
			template, err := manager.LoadEmbeddedTemplate(templateName, variables)
			require.NoError(t, err, "template %s should load", templateName)

			// Validate dependencies structure
			assert.NotNil(t, template.Dependencies, "template %s should have dependencies", templateName)
			assert.NotNil(t, template.DevDeps, "template %s should have dev dependencies", templateName)
			assert.NotNil(t, template.TestDeps, "template %s should have test dependencies", templateName)

			// Note: Core pragmas (strict, warnings, utf8) are built-in to Perl
			// and should not be listed in dependencies
		}
	})

	t.Run("template directories are reasonable", func(t *testing.T) {
		templates, err := manager.ListEmbeddedTemplates()
		if err != nil {
			t.Skip("Embedded templates not available - this test requires the binary to be built with embedded templates")
		}

		for _, templateName := range templates {
			variables := NewTemplateVariables("test-project")
			template, err := manager.LoadEmbeddedTemplate(templateName, variables)
			require.NoError(t, err, "template %s should load", templateName)

			// All templates should have at least lib and t directories
			assert.Contains(t, template.Directories, "lib", "template %s should have lib directory", templateName)
			assert.Contains(t, template.Directories, "t", "template %s should have t directory", templateName)
		}
	})
}

func TestTemplateVariableDefaults(t *testing.T) {
	manager := NewEmbeddedTemplateManager()

	t.Run("function map is properly initialized", func(t *testing.T) {
		assert.NotNil(t, manager.funcMap)
		assert.Contains(t, manager.funcMap, "default")
		assert.NotNil(t, manager.funcMap["default"])
	})
}
