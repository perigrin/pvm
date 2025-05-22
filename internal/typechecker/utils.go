// ABOUTME: Utility functions for type checking
// ABOUTME: Helper functions used across the typechecker package

package typechecker

import (
	"strings"

	"tamarou.com/pvm/internal/parser"
)

// getNodeText is a helper function to extract text from a node
func getNodeText(node parser.Node) string {
	// Use the Text() method from the Node interface
	return node.Text()
}

// ExtractTypeAndParams extracts the base type and parameters from a parameterized type
// e.g., "ArrayRef[Int]" -> "ArrayRef", ["Int"]
func ExtractTypeAndParams(paramType string) (string, []string) {
	idx := strings.Index(paramType, "[")
	if idx < 0 {
		return paramType, nil
	}

	baseType := paramType[:idx]
	paramStr := paramType[idx+1 : len(paramType)-1] // Remove outer brackets

	// Split parameters by comma, handling nested brackets
	var params []string
	bracketCount := 0
	start := 0

	for i, c := range paramStr {
		switch {
		case c == '[':
			bracketCount++
		case c == ']':
			bracketCount--
		case c == ',' && bracketCount == 0:
			params = append(params, strings.TrimSpace(paramStr[start:i]))
			start = i + 1
		}
	}

	// Add the last parameter
	if start < len(paramStr) {
		params = append(params, strings.TrimSpace(paramStr[start:]))
	}

	return baseType, params
}

// extractModuleNameFromPath extracts a module name from a file path
func extractModuleNameFromPath(path string) string {
	// This is a simplified implementation that would need to be enhanced in a real system
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}

	filename := parts[len(parts)-1]
	moduleName := strings.TrimSuffix(filename, ".pm")
	moduleName = strings.TrimSuffix(moduleName, ".pl")

	// Handle lib/Module/Name.pm style paths
	if len(parts) >= 3 {
		if parts[len(parts)-3] == "lib" {
			// This might be a module in a standard lib directory
			// Try to construct a module name from the path
			libIndex := -1
			for i, part := range parts {
				if part == "lib" {
					libIndex = i
					break
				}
			}

			if libIndex >= 0 && libIndex < len(parts)-1 {
				// Reconstruct the module name from the path components after "lib"
				moduleComponents := parts[libIndex+1 : len(parts)-1]
				moduleComponents = append(moduleComponents, moduleName)
				moduleName = strings.Join(moduleComponents, "::")
			}
		}
	}

	return moduleName
}
