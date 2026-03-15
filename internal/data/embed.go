// ABOUTME: Central data embedding service for all PVM static content
// ABOUTME: Provides access to help templates and other data from internal/data/

package data

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed help pvm
var dataFS embed.FS

// GetHelpTemplate reads a help template from help/
func GetHelpTemplate(templateName string) (string, error) {
	templatePath := fmt.Sprintf("help/%s.md", templateName)
	content, err := fs.ReadFile(dataFS, templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read help template %s: %w", templateName, err)
	}
	return string(content), nil
}

// GetAvailableHelpTemplates returns a list of available help templates
func GetAvailableHelpTemplates() ([]string, error) {
	var templates []string

	err := fs.WalkDir(dataFS, "help", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, "README.md") {
			// Extract template name from path
			templateName := strings.TrimSuffix(strings.TrimPrefix(path, "help/"), ".md")
			templates = append(templates, templateName)
		}

		return nil
	})

	return templates, err
}

// GetDataFile reads any file from the embedded directories
func GetDataFile(filePath string) ([]byte, error) {
	return fs.ReadFile(dataFS, filePath)
}

// ListDataDirectory lists contents of a directory in embedded directories
func ListDataDirectory(dirPath string) ([]fs.DirEntry, error) {
	return fs.ReadDir(dataFS, dirPath)
}
