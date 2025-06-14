// ABOUTME: Configuration template system for the PVM Ecosystem
// ABOUTME: Provides template rendering with variable substitution and inheritance

package config

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	texttemplate "text/template"

	toml "github.com/pelletier/go-toml/v2"
	"tamarou.com/pvm/internal/errors"
)

// Template represents a configuration template
type Template struct {
	Name        string            `toml:"name" json:"name"`
	Description string            `toml:"description,omitempty" json:"description,omitempty"`
	Author      string            `toml:"author,omitempty" json:"author,omitempty"`
	Version     string            `toml:"version,omitempty" json:"version,omitempty"`
	Variables   map[string]string `toml:"variables,omitempty" json:"variables,omitempty"`
	Content     string            `toml:"content" json:"content"`
	Inherits    string            `toml:"inherits,omitempty" json:"inherits,omitempty"`
}

// TemplateManager manages configuration templates
type TemplateManager struct {
	templatesDir string
	templates    map[string]*Template
	funcMap      texttemplate.FuncMap
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(templatesDir string) *TemplateManager {
	return &TemplateManager{
		templatesDir: templatesDir,
		templates:    make(map[string]*Template),
		funcMap: texttemplate.FuncMap{
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
			"env":       os.Getenv,
			"expandEnv": os.ExpandEnv,
		},
	}
}

// LoadTemplates loads all templates from the templates directory
func (tm *TemplateManager) LoadTemplates() error {
	if _, err := os.Stat(tm.templatesDir); os.IsNotExist(err) {
		// Templates directory doesn't exist, which is fine
		return nil
	}

	return filepath.WalkDir(tm.templatesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".template.toml") {
			return nil
		}

		template, err := tm.loadTemplate(path)
		if err != nil {
			return errors.NewConfigError("T001",
				"Failed to load template", err).
				WithLocation(path)
		}

		tm.templates[template.Name] = template
		return nil
	})
}

// loadTemplate loads a single template from a file
func (tm *TemplateManager) loadTemplate(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var template Template
	if err := toml.Unmarshal(data, &template); err != nil {
		return nil, err
	}

	// Validate template
	if template.Name == "" {
		return nil, fmt.Errorf("template name is required")
	}

	if template.Content == "" {
		return nil, fmt.Errorf("template content is required")
	}

	return &template, nil
}

// GetTemplate retrieves a template by name
func (tm *TemplateManager) GetTemplate(name string) (*Template, error) {
	template, exists := tm.templates[name]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", name)
	}
	return template, nil
}

// ListTemplates returns a list of available template names
func (tm *TemplateManager) ListTemplates() []string {
	names := make([]string, 0, len(tm.templates))
	for name := range tm.templates {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RenderTemplate renders a template with given variables
func (tm *TemplateManager) RenderTemplate(templateName string, variables map[string]string) (*Config, error) {
	template, err := tm.GetTemplate(templateName)
	if err != nil {
		return nil, err
	}

	// Resolve template inheritance with cycle detection
	visited := make(map[string]bool)
	content, err := tm.resolveInheritanceWithCycleDetection(template, variables, visited)
	if err != nil {
		return nil, errors.NewConfigError("T002",
			"Failed to resolve template inheritance", err).
			WithLocation("template:" + templateName)
	}

	// Merge template variables from inheritance chain with provided variables
	mergedVars := tm.mergeInheritanceVariables(template, variables)

	// Render the template
	tmpl, err := texttemplate.New(templateName).Funcs(tm.funcMap).Parse(content)
	if err != nil {
		return nil, errors.NewConfigError("T003",
			"Failed to parse template", err).
			WithLocation("template:" + templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, mergedVars); err != nil {
		return nil, errors.NewConfigError("T004",
			"Failed to execute template", err).
			WithLocation("template:" + templateName)
	}

	// Parse the rendered configuration
	config, err := ParseBytes(buf.Bytes(), "template:"+templateName)
	if err != nil {
		return nil, errors.NewConfigError("T005",
			"Failed to parse rendered template", err).
			WithLocation("template:" + templateName)
	}

	return config, nil
}

// resolveInheritance resolves template inheritance chain (legacy wrapper)
func (tm *TemplateManager) resolveInheritance(template *Template, variables map[string]string) (string, error) {
	visited := make(map[string]bool)
	return tm.resolveInheritanceWithCycleDetection(template, variables, visited)
}

// resolveInheritanceWithCycleDetection resolves template inheritance chain with cycle detection
func (tm *TemplateManager) resolveInheritanceWithCycleDetection(template *Template, variables map[string]string, visited map[string]bool) (string, error) {
	// Check for cycles
	if visited[template.Name] {
		return "", fmt.Errorf("circular template inheritance detected involving template '%s'", template.Name)
	}

	if template.Inherits == "" {
		return template.Content, nil
	}

	// Mark this template as visited
	visited[template.Name] = true
	defer func() {
		// Clean up after processing this branch
		delete(visited, template.Name)
	}()

	// Get parent template
	parent, err := tm.GetTemplate(template.Inherits)
	if err != nil {
		return "", fmt.Errorf("parent template '%s' not found", template.Inherits)
	}

	// Recursively resolve parent inheritance
	parentContent, err := tm.resolveInheritanceWithCycleDetection(parent, variables, visited)
	if err != nil {
		return "", err
	}

	// Merge parent and child content
	// Child content takes precedence over parent content
	return tm.mergeTemplateContent(parentContent, template.Content)
}

// mergeTemplateContent merges parent and child template content
func (tm *TemplateManager) mergeTemplateContent(parent, child string) (string, error) {
	if parent == "" {
		return child, nil
	}
	if child == "" {
		return parent, nil
	}

	// Parse both templates to extract sections
	parentSections, err := tm.parseTemplateSections(parent)
	if err != nil {
		return "", fmt.Errorf("failed to parse parent template: %w", err)
	}

	childSections, err := tm.parseTemplateSections(child)
	if err != nil {
		return "", fmt.Errorf("failed to parse child template: %w", err)
	}

	// Merge sections - child sections are merged with parent sections at field level
	merged := make(map[string]string)
	for section, content := range parentSections {
		merged[section] = content
	}

	// For each child section, merge with parent section if it exists
	for section, childContent := range childSections {
		if parentContent, exists := merged[section]; exists {
			// Merge section content at field level
			mergedContent, err := tm.mergeSectionContent(parentContent, childContent)
			if err != nil {
				return "", fmt.Errorf("failed to merge section %s: %w", section, err)
			}
			merged[section] = mergedContent
		} else {
			// No parent section, use child content as-is
			merged[section] = childContent
		}
	}

	// Reconstruct template content
	return tm.reconstructTemplate(merged), nil
}

// mergeSectionContent merges content within the same TOML section
func (tm *TemplateManager) mergeSectionContent(parent, child string) (string, error) {
	// Extract the section header and content
	parentLines := strings.Split(parent, "\n")
	childLines := strings.Split(child, "\n")

	var sectionHeader string
	var parentFields []string
	var childFields []string

	// Parse parent section
	for i, line := range parentLines {
		if i == 0 && strings.HasPrefix(strings.TrimSpace(line), "[") {
			sectionHeader = line
		} else if strings.TrimSpace(line) != "" {
			parentFields = append(parentFields, line)
		}
	}

	// Parse child section and collect fields
	for i, line := range childLines {
		if i == 0 && strings.HasPrefix(strings.TrimSpace(line), "[") {
			// Skip section header from child
		} else if strings.TrimSpace(line) != "" {
			childFields = append(childFields, line)
		}
	}

	// Build merged section: header + parent fields + child fields
	var result []string
	if sectionHeader != "" {
		result = append(result, sectionHeader)
	}

	// Add parent fields first
	for _, field := range parentFields {
		result = append(result, field)
	}

	// Add child fields (which will override parent fields with same key in TOML)
	for _, field := range childFields {
		result = append(result, field)
	}

	return strings.Join(result, "\n"), nil
}

// mergeConfigMaps recursively merges two configuration maps
func (tm *TemplateManager) mergeConfigMaps(parent, child map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Start with parent values
	for key, value := range parent {
		merged[key] = value
	}

	// Override/merge with child values
	for key, childValue := range child {
		if parentValue, exists := merged[key]; exists {
			// If both are maps, merge recursively
			if parentMap, ok := parentValue.(map[string]interface{}); ok {
				if childMap, ok := childValue.(map[string]interface{}); ok {
					merged[key] = tm.mergeConfigMaps(parentMap, childMap)
					continue
				}
			}
		}
		// Otherwise, child value overrides parent value
		merged[key] = childValue
	}

	return merged
}

// parseTemplateSections parses template content into sections
func (tm *TemplateManager) parseTemplateSections(content string) (map[string]string, error) {
	sections := make(map[string]string)

	lines := strings.Split(content, "\n")
	var currentSection string
	var currentContent []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this is a section header
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			// Save previous section
			if currentSection != "" {
				sections[currentSection] = strings.Join(currentContent, "\n")
			}

			// Start new section
			currentSection = trimmed
			currentContent = []string{line}
		} else {
			// Add to current section
			currentContent = append(currentContent, line)
		}
	}

	// Save final section
	if currentSection != "" {
		sections[currentSection] = strings.Join(currentContent, "\n")
	}

	return sections, nil
}

// reconstructTemplate reconstructs template content from sections
func (tm *TemplateManager) reconstructTemplate(sections map[string]string) string {
	var result []string

	// Sort sections for consistent output
	sectionNames := make([]string, 0, len(sections))
	for name := range sections {
		sectionNames = append(sectionNames, name)
	}
	sort.Strings(sectionNames)

	for _, name := range sectionNames {
		if content := sections[name]; content != "" {
			result = append(result, content)
		}
	}

	return strings.Join(result, "\n\n")
}

// mergeVariables merges template variables with provided variables
func (tm *TemplateManager) mergeVariables(templateVars, providedVars map[string]string) map[string]string {
	merged := make(map[string]string)

	// Start with template variables (defaults)
	for key, value := range templateVars {
		merged[key] = value
	}

	// Override with provided variables
	for key, value := range providedVars {
		merged[key] = value
	}

	return merged
}

// mergeInheritanceVariables merges variables from the entire inheritance chain
func (tm *TemplateManager) mergeInheritanceVariables(template *Template, providedVars map[string]string) map[string]string {
	visited := make(map[string]bool)
	return tm.mergeInheritanceVariablesWithCycleDetection(template, providedVars, visited)
}

// mergeInheritanceVariablesWithCycleDetection merges variables with cycle detection
func (tm *TemplateManager) mergeInheritanceVariablesWithCycleDetection(template *Template, providedVars map[string]string, visited map[string]bool) map[string]string {
	merged := make(map[string]string)

	// Check for cycles
	if visited[template.Name] {
		// Return empty map to avoid infinite recursion
		return merged
	}

	// Mark this template as visited
	visited[template.Name] = true
	defer func() {
		delete(visited, template.Name)
	}()

	// Recursively collect variables from parent templates first
	if template.Inherits != "" {
		if parent, err := tm.GetTemplate(template.Inherits); err == nil {
			parentVars := tm.mergeInheritanceVariablesWithCycleDetection(parent, nil, visited)
			for key, value := range parentVars {
				merged[key] = value
			}
		}
	}

	// Add current template variables (child overrides parent)
	for key, value := range template.Variables {
		merged[key] = value
	}

	// Finally, override with provided variables (highest priority)
	for key, value := range providedVars {
		merged[key] = value
	}

	return merged
}

// ValidateTemplate validates a template for correctness
func (tm *TemplateManager) ValidateTemplate(template *Template) []error {
	var errors []error

	// Check required fields
	if template.Name == "" {
		errors = append(errors, fmt.Errorf("template name is required"))
	}

	if template.Content == "" {
		errors = append(errors, fmt.Errorf("template content is required"))
	}

	// Check inheritance chain
	if template.Inherits != "" {
		if _, err := tm.GetTemplate(template.Inherits); err != nil {
			errors = append(errors, fmt.Errorf("parent template '%s' not found", template.Inherits))
		}
	}

	// Try to parse template content
	if template.Content != "" {
		tmpl, err := texttemplate.New(template.Name).Funcs(tm.funcMap).Parse(template.Content)
		if err != nil {
			errors = append(errors, fmt.Errorf("template parsing error: %w", err))
		} else {
			// Test rendering with default variables
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, template.Variables); err != nil {
				errors = append(errors, fmt.Errorf("template execution error: %w", err))
			} else {
				// Try to parse the rendered result
				if _, err := ParseBytes(buf.Bytes(), "validation"); err != nil {
					errors = append(errors, fmt.Errorf("rendered template is not valid TOML: %w", err))
				}
			}
		}
	}

	return errors
}

// SaveTemplate saves a template to the templates directory
func (tm *TemplateManager) SaveTemplate(template *Template) error {
	// Validate template first
	if errs := tm.ValidateTemplate(template); len(errs) > 0 {
		return fmt.Errorf("template validation failed: %v", errs)
	}

	// Ensure templates directory exists
	if err := os.MkdirAll(tm.templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Serialize template
	data, err := toml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	// Write to file
	path := filepath.Join(tm.templatesDir, template.Name+".template.toml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	// Update in-memory cache
	tm.templates[template.Name] = template

	return nil
}

// DeleteTemplate removes a template
func (tm *TemplateManager) DeleteTemplate(name string) error {
	// Check if template exists
	if _, exists := tm.templates[name]; !exists {
		return fmt.Errorf("template '%s' not found", name)
	}

	// Delete file
	path := filepath.Join(tm.templatesDir, name+".template.toml")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete template file: %w", err)
	}

	// Remove from cache
	delete(tm.templates, name)

	return nil
}
