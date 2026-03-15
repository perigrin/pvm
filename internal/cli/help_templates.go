// ABOUTME: Template system for PVM help content
// ABOUTME: Provides embedded help templates that can be easily updated

package cli

import (
	"fmt"
	"strings"
	"text/template"

	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/data"
)

// HelpTemplateData contains data for help template rendering
type HelpTemplateData struct {
	// Add template variables as needed
	Version     string
	ProjectPath string
	// Add more fields as needed for template interpolation
}

// RenderHelpTemplate renders a help template with the given data
func RenderHelpTemplate(templateName string, templateData HelpTemplateData) (string, error) {
	// Read the template file from central data service
	templateContent, err := data.GetHelpTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to read help template %s: %w", templateName, err)
	}

	// Parse and execute the template
	tmpl, err := template.New(templateName).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse help template %s: %w", templateName, err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute help template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// GetAvailableHelpTemplates returns a list of available help templates
func GetAvailableHelpTemplates() ([]string, error) {
	return data.GetAvailableHelpTemplates()
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
