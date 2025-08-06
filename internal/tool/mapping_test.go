// ABOUTME: Comprehensive tests for tool-to-module mapping functionality
// ABOUTME: Tests all mapping resolution strategies and validation logic
package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToolMapping(t *testing.T) {
	tm := NewToolMapping()

	assert.NotNil(t, tm)
	assert.NotNil(t, tm.builtinMappings)
	assert.NotNil(t, tm.configMappings)

	// Should have built-in mappings
	assert.Greater(t, len(tm.builtinMappings), 0)

	// Should include common tools
	assert.Contains(t, tm.builtinMappings, "ack")
	assert.Contains(t, tm.builtinMappings, "cpanm")
	assert.Contains(t, tm.builtinMappings, "prove")
}

func TestResolveToolToModule_BuiltinMappings(t *testing.T) {
	tm := NewToolMapping()

	testCases := []struct {
		toolName       string
		expectedModule string
		expectedSource string
	}{
		{"ack", "App::Ack", "builtin"},
		{"cpanm", "App::cpanminus", "builtin"},
		{"prove", "Test::Harness", "builtin"},
		{"perltidy", "Perl-Tidy", "builtin"},
		{"perlcritic", "Perl::Critic", "builtin"},
	}

	for _, tc := range testCases {
		t.Run(tc.toolName, func(t *testing.T) {
			resolution, err := tm.ResolveToolToModule(tc.toolName)

			require.NoError(t, err)
			assert.Equal(t, tc.toolName, resolution.ToolName)
			assert.Equal(t, tc.expectedModule, resolution.ModuleName)
			assert.Equal(t, tc.expectedSource, resolution.Source)
		})
	}
}

func TestResolveToolToModule_ExplicitModule(t *testing.T) {
	tm := NewToolMapping()

	testCases := []string{
		"App::Ack",
		"Test::Harness",
		"Perl-Tidy",
		"My::Custom::Module",
	}

	for _, moduleName := range testCases {
		t.Run(moduleName, func(t *testing.T) {
			resolution, err := tm.ResolveToolToModule(moduleName)

			require.NoError(t, err)
			assert.Equal(t, moduleName, resolution.ToolName)
			assert.Equal(t, moduleName, resolution.ModuleName)
			assert.Equal(t, "explicit", resolution.Source)
		})
	}
}

func TestResolveToolToModule_ConfigOverride(t *testing.T) {
	tm := NewToolMapping()

	// Add config mapping that overrides built-in
	err := tm.AddConfigMapping("ack", "My::Custom::Ack")
	require.NoError(t, err)

	resolution, err := tm.ResolveToolToModule("ack")
	require.NoError(t, err)

	assert.Equal(t, "ack", resolution.ToolName)
	assert.Equal(t, "My::Custom::Ack", resolution.ModuleName)
	assert.Equal(t, "config", resolution.Source)
}

func TestResolveToolToModule_UnknownTool(t *testing.T) {
	tm := NewToolMapping()

	_, err := tm.ResolveToolToModule("nonexistent-tool")
	assert.Error(t, err)

	var toolErr *ToolError
	require.ErrorAs(t, err, &toolErr)
	assert.Equal(t, ErrToolNotFound, toolErr.Code)
}

func TestResolveToolToModule_EmptyToolName(t *testing.T) {
	tm := NewToolMapping()

	_, err := tm.ResolveToolToModule("")
	assert.Error(t, err)

	var toolErr *ToolError
	require.ErrorAs(t, err, &toolErr)
	assert.Equal(t, ErrInvalidToolName, toolErr.Code)
}

func TestAddConfigMapping_Valid(t *testing.T) {
	tm := NewToolMapping()

	testCases := []struct {
		toolName   string
		moduleName string
	}{
		{"my-tool", "App::MyTool"},
		{"test_tool", "Test::Tool"},
		{"simple", "Simple"},
		{"complex-name", "Very::Complex::Module::Name"},
	}

	for _, tc := range testCases {
		t.Run(tc.toolName, func(t *testing.T) {
			err := tm.AddConfigMapping(tc.toolName, tc.moduleName)
			assert.NoError(t, err)

			// Verify the mapping was added
			resolution, err := tm.ResolveToolToModule(tc.toolName)
			require.NoError(t, err)
			assert.Equal(t, tc.moduleName, resolution.ModuleName)
			assert.Equal(t, "config", resolution.Source)
		})
	}
}

func TestAddConfigMapping_Invalid(t *testing.T) {
	tm := NewToolMapping()

	testCases := []struct {
		name          string
		toolName      string
		moduleName    string
		expectedError string
	}{
		{"empty tool name", "", "App::Test", "tool name and module name cannot be empty"},
		{"empty module name", "test", "", "tool name and module name cannot be empty"},
		{"invalid tool name", "tool/with/slashes", "App::Test", "invalid tool name"},
		{"invalid module name", "test", "Invalid::Module::123abc", "invalid module name"},
		{"tool with spaces", "tool with spaces", "App::Test", "invalid tool name"},
		{"module with spaces", "test", "App::Test Space", "invalid module name"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tm.AddConfigMapping(tc.toolName, tc.moduleName)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestListMappings(t *testing.T) {
	tm := NewToolMapping()

	// Add some config mappings
	err := tm.AddConfigMapping("custom-tool", "App::CustomTool")
	require.NoError(t, err)

	err = tm.AddConfigMapping("ack", "My::Custom::Ack") // Override built-in
	require.NoError(t, err)

	mappings := tm.ListMappings()

	// Should contain built-in mappings
	assert.Contains(t, mappings, "cpanm")
	assert.Equal(t, "App::cpanminus", mappings["cpanm"].ModuleName)
	assert.Equal(t, "builtin", mappings["cpanm"].Source)

	// Should contain custom mapping
	assert.Contains(t, mappings, "custom-tool")
	assert.Equal(t, "App::CustomTool", mappings["custom-tool"].ModuleName)
	assert.Equal(t, "config", mappings["custom-tool"].Source)

	// Config should override built-in
	assert.Contains(t, mappings, "ack")
	assert.Equal(t, "My::Custom::Ack", mappings["ack"].ModuleName)
	assert.Equal(t, "config", mappings["ack"].Source)
}

func TestIsValidToolName(t *testing.T) {
	validNames := []string{
		"ack",
		"test-tool",
		"my_tool",
		"Tool123",
		"a",
		"very-long-tool-name-with-numbers-123",
	}

	for _, name := range validNames {
		t.Run("valid_"+name, func(t *testing.T) {
			assert.True(t, isValidToolName(name))
		})
	}

	invalidNames := []string{
		"",
		"tool with spaces",
		"tool/with/slashes",
		"tool\\with\\backslashes",
		"tool@with@symbols",
		"tool.with.dots",
		"tool:with:colons",
	}

	for _, name := range invalidNames {
		t.Run("invalid_"+name, func(t *testing.T) {
			assert.False(t, isValidToolName(name))
		})
	}
}

func TestIsValidModuleName(t *testing.T) {
	validNames := []string{
		"Module",
		"App::Module",
		"Very::Complex::Module::Name",
		"Test_Module",
		"Module123",
		"App::Module_With_Underscores",
	}

	for _, name := range validNames {
		t.Run("valid_"+name, func(t *testing.T) {
			assert.True(t, isValidModuleName(name))
		})
	}

	invalidNames := []string{
		"",
		"::Module",         // starts with ::
		"Module::",         // ends with ::
		"Module::::Double", // double ::
		"module",           // starts with lowercase
		"123Module",        // starts with number
		"App::123Module",   // part starts with number
		"Module With Spaces",
		"Module-With-Dashes",
		"Module@With@Symbols",
	}

	for _, name := range invalidNames {
		t.Run("invalid_"+name, func(t *testing.T) {
			assert.False(t, isValidModuleName(name))
		})
	}
}

func TestToolMappingWithResolver(t *testing.T) {
	tm := NewToolMapping()

	// Create a mock resolver that returns a specific result
	mockResolver := NewMockCPANResolver()
	tm.SetResolver(mockResolver)

	// Test that resolver is set
	assert.Equal(t, mockResolver, tm.resolver)

	// Test resolving unknown tool without actual CPAN call
	_, err := tm.ResolveToolToModule("definitely-nonexistent-tool")
	assert.Error(t, err)
}

// MockCPANResolver for testing without external dependencies
type MockCPANResolver struct {
	responses map[string]*ToolResolution
	errors    map[string]error
}

func NewMockCPANResolver() *MockCPANResolver {
	return &MockCPANResolver{
		responses: make(map[string]*ToolResolution),
		errors:    make(map[string]error),
	}
}

func (m *MockCPANResolver) SearchTool(toolName string) (*ToolResolution, error) {
	if err, exists := m.errors[toolName]; exists {
		return nil, err
	}

	if resolution, exists := m.responses[toolName]; exists {
		return resolution, nil
	}

	return nil, NewToolError(ErrToolNotFound, "tool not found in mock")
}
