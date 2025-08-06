// ABOUTME: Core tool-to-module mapping logic for resolving tool names to CPAN modules
// ABOUTME: Provides extensible mapping system with built-in mappings and configuration support
package tool

import (
	"fmt"
	"strings"
)

// ToolMapping represents the core mapping system for tool name resolution
type ToolMapping struct {
	builtinMappings map[string]string
	configMappings  map[string]string
	resolver        Resolver
}

// NewToolMapping creates a new tool mapping instance with built-in mappings
func NewToolMapping() *ToolMapping {
	tm := &ToolMapping{
		builtinMappings: make(map[string]string),
		configMappings:  make(map[string]string),
	}

	// Initialize built-in mappings
	tm.initBuiltinMappings()

	// Initialize MetaCPAN resolver for fallback searches
	if resolver, err := NewMetaCPANResolver(); err == nil {
		tm.SetResolver(resolver)
	}

	return tm
}

// initBuiltinMappings sets up the hardcoded mappings for common tools
func (tm *ToolMapping) initBuiltinMappings() {
	tm.builtinMappings = GetBuiltinMappings()
}

// ResolveToolToModule resolves a tool name to its CPAN module
func (tm *ToolMapping) ResolveToolToModule(toolName string) (*ToolResolution, error) {
	if toolName == "" {
		return nil, NewToolError(ErrInvalidToolName, fmt.Sprintf("tool name cannot be empty"))
	}

	// Check if it's already a module name (contains ::)
	if strings.Contains(toolName, "::") {
		return &ToolResolution{
			ToolName:   toolName,
			ModuleName: toolName,
			Source:     "explicit",
		}, nil
	}

	// Check config mappings first (user overrides)
	if module, exists := tm.configMappings[toolName]; exists {
		return &ToolResolution{
			ToolName:   toolName,
			ModuleName: module,
			Source:     "config",
		}, nil
	}

	// Check built-in mappings
	if module, exists := tm.builtinMappings[toolName]; exists {
		return &ToolResolution{
			ToolName:   toolName,
			ModuleName: module,
			Source:     "builtin",
		}, nil
	}

	// Try CPAN search as fallback
	if tm.resolver != nil {
		if resolution, err := tm.resolver.SearchTool(toolName); err == nil {
			return resolution, nil
		}
	}

	return nil, NewToolError(ErrToolNotFound, fmt.Sprintf("tool '%s' not found in mappings or CPAN search", toolName))
}

// AddConfigMapping adds a custom tool mapping from configuration
func (tm *ToolMapping) AddConfigMapping(toolName, moduleName string) error {
	if toolName == "" || moduleName == "" {
		return NewToolError(ErrInvalidMapping, "tool name and module name cannot be empty")
	}

	if err := tm.validateMapping(toolName, moduleName); err != nil {
		return err
	}

	tm.configMappings[toolName] = moduleName
	return nil
}

// SetResolver sets the CPAN resolver for fallback lookups
func (tm *ToolMapping) SetResolver(resolver Resolver) {
	tm.resolver = resolver
}

// ListMappings returns all available tool mappings
func (tm *ToolMapping) ListMappings() map[string]ToolResolution {
	result := make(map[string]ToolResolution)

	// Add built-in mappings
	for tool, module := range tm.builtinMappings {
		result[tool] = ToolResolution{
			ToolName:   tool,
			ModuleName: module,
			Source:     "builtin",
		}
	}

	// Add config mappings (overrides built-in)
	for tool, module := range tm.configMappings {
		result[tool] = ToolResolution{
			ToolName:   tool,
			ModuleName: module,
			Source:     "config",
		}
	}

	return result
}

// validateMapping validates a tool name and module name mapping
func (tm *ToolMapping) validateMapping(toolName, moduleName string) error {
	// Tool name validation - should be simple identifier
	if !isValidToolName(toolName) {
		return NewToolError(ErrInvalidMapping, fmt.Sprintf("invalid tool name '%s': must contain only letters, numbers, hyphens, and underscores", toolName))
	}

	// Module name validation - should follow Perl module naming
	if !isValidModuleName(moduleName) {
		return NewToolError(ErrInvalidMapping, fmt.Sprintf("invalid module name '%s': must follow Perl module naming conventions", moduleName))
	}

	return nil
}

// isValidToolName checks if a tool name follows valid patterns
func isValidToolName(name string) bool {
	if len(name) == 0 {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}

// isValidModuleName checks if a module name follows Perl conventions
func isValidModuleName(name string) bool {
	if len(name) == 0 {
		return false
	}

	parts := strings.Split(name, "::")
	for i, part := range parts {
		if len(part) == 0 {
			return false
		}

		// First part should start with an uppercase letter, other parts can start with any letter
		if i == 0 {
			if part[0] < 'A' || part[0] > 'Z' {
				return false
			}
		} else {
			if !((part[0] >= 'A' && part[0] <= 'Z') || (part[0] >= 'a' && part[0] <= 'z')) {
				return false
			}
		}

		for _, char := range part {
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == '_') {
				return false
			}
		}
	}

	return true
}

// GetBuiltinMappings returns the built-in tool mappings for use by other packages
func GetBuiltinMappings() map[string]string {
	mappings := map[string]string{
		"ack":         "App::Ack",
		"cpanm":       "App::cpanminus",
		"prove":       "Test::Harness",
		"perltidy":    "Perl-Tidy",
		"perlcritic":  "Perl::Critic",
		"fatpack":     "App::FatPacker",
		"plackup":     "Plack",
		"cpanfile":    "Module::CPANfile",
		"carton":      "Carton",
		"dzil":        "Dist::Zilla",
		"minil":       "Minilla",
		"pmversions":  "Perl::Version",
		"cpan-upload": "CPAN::Uploader",
		"cpan-audit":  "CPAN::Audit",
		"metacpan":    "MetaCPAN::Client",
	}
	return mappings
}
