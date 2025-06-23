// ABOUTME: Tests for built-in tool mappings and category functionality
// ABOUTME: Validates all built-in mappings and category organization
package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinMappings_Coverage(t *testing.T) {
	// Ensure all expected tools are present
	expectedTools := []string{
		"ack", "perlcritic", "perltidy", "cpanm", "cpan-upload", "cpan-audit",
		"metacpan", "prove", "fatpack", "dzil", "minil", "cpanfile", "carton",
		"plackup", "morbo", "hypnotoad", "pmversions", "reply", "re-pl",
		"pod2usage", "podchecker", "pod2html", "pod2man", "pod2text",
	}

	for _, tool := range expectedTools {
		t.Run(tool, func(t *testing.T) {
			_, exists := BuiltinMappings[tool]
			assert.True(t, exists, "Expected tool %s to exist in built-in mappings", tool)
		})
	}
}

func TestBuiltinMappings_Structure(t *testing.T) {
	for toolName, mapping := range BuiltinMappings {
		t.Run(toolName, func(t *testing.T) {
			// Every mapping should have required fields
			assert.NotEmpty(t, mapping.Module, "Module should not be empty for %s", toolName)
			assert.NotEmpty(t, mapping.Category, "Category should not be empty for %s", toolName)
			assert.NotEmpty(t, mapping.Executable, "Executable should not be empty for %s", toolName)

			// Module should be valid
			assert.True(t, isValidModuleName(mapping.Module), "Module name should be valid for %s: %s", toolName, mapping.Module)

			// Category should be one of expected categories
			validCategories := []string{
				"search", "quality", "formatting", "cpan", "security", "testing",
				"packaging", "dependencies", "web", "utility", "documentation",
			}
			assert.Contains(t, validCategories, mapping.Category, "Category should be valid for %s: %s", toolName, mapping.Category)
		})
	}
}

func TestGetBuiltinMapping(t *testing.T) {
	// Test existing mapping
	mapping, exists := GetBuiltinMapping("ack")
	assert.True(t, exists)
	assert.Equal(t, "App::Ack", mapping.Module)
	assert.Equal(t, "search", mapping.Category)
	assert.Equal(t, "ack", mapping.Executable)

	// Test non-existing mapping
	_, exists = GetBuiltinMapping("nonexistent-tool")
	assert.False(t, exists)
}

func TestListBuiltinTools(t *testing.T) {
	tools := ListBuiltinTools()

	assert.Greater(t, len(tools), 0)

	// Should contain expected tools
	assert.Contains(t, tools, "ack")
	assert.Contains(t, tools, "cpanm")
	assert.Contains(t, tools, "prove")

	// Should not contain duplicates
	toolSet := make(map[string]bool)
	for _, tool := range tools {
		assert.False(t, toolSet[tool], "Tool %s should not be duplicated", tool)
		toolSet[tool] = true
	}
}

func TestGetToolsByCategory(t *testing.T) {
	testCases := []struct {
		category      string
		expectedTools []string
	}{
		{"search", []string{"ack"}},
		{"cpan", []string{"cpanm", "cpan-upload", "metacpan"}},
		{"security", []string{"cpan-audit"}},
		{"testing", []string{"prove"}},
		{"quality", []string{"perlcritic"}},
		{"formatting", []string{"perltidy"}},
		{"web", []string{"plackup", "morbo", "hypnotoad"}},
		{"packaging", []string{"fatpack", "dzil", "minil"}},
		{"dependencies", []string{"cpanfile", "carton"}},
	}

	for _, tc := range testCases {
		t.Run(tc.category, func(t *testing.T) {
			tools := GetToolsByCategory(tc.category)

			for _, expectedTool := range tc.expectedTools {
				assert.Contains(t, tools, expectedTool, "Category %s should contain tool %s", tc.category, expectedTool)
			}

			// Verify all returned tools actually belong to this category
			for _, tool := range tools {
				mapping, exists := BuiltinMappings[tool]
				require.True(t, exists, "Tool %s should exist in built-in mappings", tool)
				assert.Equal(t, tc.category, mapping.Category, "Tool %s should belong to category %s", tool, tc.category)
			}
		})
	}
}

func TestGetToolsByCategory_NonexistentCategory(t *testing.T) {
	tools := GetToolsByCategory("nonexistent-category")
	assert.Empty(t, tools)
}

func TestGetAllCategories(t *testing.T) {
	categories := GetAllCategories()

	assert.Greater(t, len(categories), 0)

	expectedCategories := []string{
		"search", "quality", "formatting", "cpan", "security", "testing",
		"packaging", "dependencies", "web", "utility", "documentation",
	}

	for _, expectedCategory := range expectedCategories {
		assert.Contains(t, categories, expectedCategory, "Should contain category %s", expectedCategory)
	}

	// Should not contain duplicates
	categorySet := make(map[string]bool)
	for _, category := range categories {
		assert.False(t, categorySet[category], "Category %s should not be duplicated", category)
		categorySet[category] = true
	}
}

func TestBuiltinMappings_KnownGoodMappings(t *testing.T) {
	// Test specific mappings we know are correct
	knownMappings := map[string]string{
		"ack":         "App::Ack",
		"cpanm":       "App::cpanminus",
		"prove":       "Test::Harness",
		"perltidy":    "Perl::Tidy",
		"perlcritic":  "Perl::Critic",
		"fatpack":     "App::FatPacker",
		"plackup":     "Plack",
		"carton":      "Carton",
		"dzil":        "Dist::Zilla",
		"minil":       "Minilla",
		"cpan-upload": "CPAN::Uploader",
		"cpan-audit":  "CPAN::Audit",
		"metacpan":    "MetaCPAN::Client",
	}

	for tool, expectedModule := range knownMappings {
		t.Run(tool, func(t *testing.T) {
			mapping, exists := GetBuiltinMapping(tool)
			require.True(t, exists, "Tool %s should exist in built-in mappings", tool)
			assert.Equal(t, expectedModule, mapping.Module, "Tool %s should map to module %s", tool, expectedModule)
		})
	}
}

func TestBuiltinMappings_DescriptionsPresent(t *testing.T) {
	for toolName, mapping := range BuiltinMappings {
		t.Run(toolName, func(t *testing.T) {
			assert.NotEmpty(t, mapping.Description, "Description should not be empty for %s", toolName)
			assert.Greater(t, len(mapping.Description), 10, "Description should be descriptive for %s", toolName)
		})
	}
}
