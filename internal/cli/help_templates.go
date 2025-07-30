// ABOUTME: Template system for PVM help content
// ABOUTME: Provides embedded help templates that can be easily updated

package cli

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
	
	"tamarou.com/pvm/internal/cli/ui"
)

//go:embed help
var helpTemplatesFS embed.FS

// HelpTemplateData contains data for help template rendering
type HelpTemplateData struct {
	// Add template variables as needed
	Version     string
	ProjectPath string
	// Add more fields as needed for template interpolation
}

// RenderHelpTemplate renders a help template with the given data
func RenderHelpTemplate(templateName string, data HelpTemplateData) (string, error) {
	// Read the template file
	templatePath := fmt.Sprintf("help/%s.md", templateName)
	templateContent, err := helpTemplatesFS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read help template %s: %w", templateName, err)
	}

	// Parse and execute the template
	tmpl, err := template.New(templateName).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse help template %s: %w", templateName, err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute help template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// GetAvailableHelpTemplates returns a list of available help templates
func GetAvailableHelpTemplates() ([]string, error) {
	var templates []string
	
	err := fs.WalkDir(helpTemplatesFS, "help", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			// Extract template name from path
			templateName := strings.TrimSuffix(strings.TrimPrefix(path, "help/"), ".md")
			templates = append(templates, templateName)
		}
		
		return nil
	})
	
	return templates, err
}

// RenderMarkdownAsHelp renders markdown content for help display
func RenderMarkdownAsHelp(markdown string, output *ui.Output) {
	lines := strings.Split(markdown, "\n")
	
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "# "):
			// Main header
			output.Header(strings.TrimPrefix(line, "# "))
		case strings.HasPrefix(line, "## "):
			// Sub header  
			output.SubHeader(strings.TrimPrefix(line, "## "))
		case strings.HasPrefix(line, "### "):
			// Sub section (treat as warning for emphasis)
			output.Warning("%s", strings.TrimPrefix(line, "### "))
		case strings.HasPrefix(line, "💡 "):
			// Info tips
			output.Info("%s", line)
		case strings.HasPrefix(line, "**Problem:**"):
			// Problem descriptions
			output.Info("%s", line)
		case strings.HasPrefix(line, "**Solution:**"):
			// Solutions
			output.Success("%s", line)
		case strings.HasPrefix(line, "**Command:**") || strings.HasPrefix(line, "**Commands:**"):
			// Commands
			output.Printf("%s\n", line)
		case strings.HasPrefix(line, "- "):
			// List items - collect and render as list
			continue // We'll handle list rendering separately
		case strings.HasPrefix(line, "`") && strings.HasSuffix(line, "`"):
			// Inline code
			output.Printf("  %s\n", line)
		case strings.HasPrefix(line, "   "):
			// Code blocks (indented)
			output.Printf("  %s\n", line)
		case line == "":
			// Empty line
			output.Println("")
		default:
			// Regular text
			output.Println(line)
		}
	}
}