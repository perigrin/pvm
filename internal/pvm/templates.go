// ABOUTME: Embedded template system for workspace initialization
// ABOUTME: Provides text/template based project templates using embed.FS

package pvm

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pelletier/go-toml/v2"
)

// Global embedded filesystem - will be initialized from main package
var GlobalTemplatesFS embed.FS

// TemplateVariables represents the structured set of variables available to templates
type TemplateVariables struct {
	ProjectName string
	Version     string
	WebPort     string
	DatabaseURL string
}

// ToMap converts TemplateVariables to a map for template execution
func (tv TemplateVariables) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"ProjectName": tv.ProjectName,
		"Version":     tv.Version,
		"WebPort":     tv.WebPort,
		"DatabaseURL": tv.DatabaseURL,
	}
}

// NewTemplateVariables creates a new TemplateVariables with sensible defaults
func NewTemplateVariables(projectName string) TemplateVariables {
	return TemplateVariables{
		ProjectName: projectName,
		Version:     "0.01",
		WebPort:     "3000",
		DatabaseURL: "sqlite:db/app.db",
	}
}

// EmbeddedTemplateManager manages embedded project templates
type EmbeddedTemplateManager struct {
	templatesFS embed.FS
	funcMap     template.FuncMap
}

// NewEmbeddedTemplateManager creates a new embedded template manager
func NewEmbeddedTemplateManager() *EmbeddedTemplateManager {
	return &EmbeddedTemplateManager{
		templatesFS: GlobalTemplatesFS,
		funcMap: template.FuncMap{
			"upper":      strings.ToUpper,
			"lower":      strings.ToLower,
			"title":      strings.ToTitle,
			"join":       strings.Join,
			"split":      strings.Split,
			"contains":   strings.Contains,
			"hasPrefix":  strings.HasPrefix,
			"hasSuffix":  strings.HasSuffix,
			"trimSpace":  strings.TrimSpace,
			"trimPrefix": strings.TrimPrefix,
			"trimSuffix": strings.TrimSuffix,
			"replace":    strings.ReplaceAll,
			"default": func(def string, val interface{}) string {
				if val == nil {
					return def
				}
				if str, ok := val.(string); ok && str != "" {
					return str
				}
				return def
			},
		},
	}
}

// ListEmbeddedTemplates returns all available embedded template names
func (etm *EmbeddedTemplateManager) ListEmbeddedTemplates() ([]string, error) {
	entries, err := fs.ReadDir(etm.templatesFS, "assets/templates")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded templates: %w", err)
	}

	var templates []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".toml.tmpl") {
			name := strings.TrimSuffix(entry.Name(), ".toml.tmpl")
			templates = append(templates, name)
		}
	}

	return templates, nil
}

// LoadEmbeddedTemplate loads and renders an embedded template
func (etm *EmbeddedTemplateManager) LoadEmbeddedTemplate(name string, variables TemplateVariables) (ProjectTemplate, error) {
	templatePath := filepath.Join("assets/templates", name+".toml.tmpl")

	// Read template content
	content, err := fs.ReadFile(etm.templatesFS, templatePath)
	if err != nil {
		return ProjectTemplate{}, fmt.Errorf("template '%s' not found: %w", name, err)
	}

	// Parse and execute template
	tmpl, err := template.New(name).Funcs(etm.funcMap).Parse(string(content))
	if err != nil {
		return ProjectTemplate{}, fmt.Errorf("failed to parse template '%s': %w", name, err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, variables.ToMap()); err != nil {
		return ProjectTemplate{}, fmt.Errorf("failed to execute template '%s': %w", name, err)
	}

	// Parse the rendered TOML
	var templateData struct {
		Name        string `toml:"name"`
		Description string `toml:"description"`
		Directories struct {
			Create []string `toml:"create"`
		} `toml:"directories"`
		Dependencies     map[string]string `toml:"dependencies"`
		DevDependencies  map[string]string `toml:"dev_dependencies"`
		TestDependencies map[string]string `toml:"test_dependencies"`
		Config           map[string]any    `toml:"config"`
		GitIgnore        struct {
			Entries []string `toml:"entries"`
		} `toml:"gitignore_additions"`
	}

	if err := toml.Unmarshal([]byte(buf.String()), &templateData); err != nil {
		return ProjectTemplate{}, fmt.Errorf("failed to parse rendered template '%s': %w", name, err)
	}

	// Convert to ProjectTemplate
	return ProjectTemplate{
		Name:         templateData.Name,
		Description:  templateData.Description,
		Directories:  templateData.Directories.Create,
		Dependencies: templateData.Dependencies,
		DevDeps:      templateData.DevDependencies,
		TestDeps:     templateData.TestDependencies,
		Config:       templateData.Config,
		GitIgnore:    templateData.GitIgnore.Entries,
	}, nil
}

// GetEmbeddedTemplateDescription returns the description of an embedded template
func (etm *EmbeddedTemplateManager) GetEmbeddedTemplateDescription(name string) (string, error) {
	// Load template with minimal variables to get description
	variables := NewTemplateVariables("example")
	template, err := etm.LoadEmbeddedTemplate(name, variables)
	if err != nil {
		return "", err
	}
	return template.Description, nil
}
