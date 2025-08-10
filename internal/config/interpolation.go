// ABOUTME: Environment variable interpolation for configuration values
// ABOUTME: Supports ${VAR} and ${VAR:-default} syntax with cycle detection

package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/errors"
)

// InterpolationEngine handles environment variable interpolation in configuration values
type InterpolationEngine struct {
	// visitedVars tracks variables being interpolated to detect cycles
	visitedVars map[string]bool
	// maxDepth prevents infinite recursion
	maxDepth int
	// sensitiveVars contains patterns for sensitive environment variables
	sensitiveVars []string
}

// NewInterpolationEngine creates a new interpolation engine
func NewInterpolationEngine() *InterpolationEngine {
	return &InterpolationEngine{
		visitedVars: make(map[string]bool),
		maxDepth:    10,
		sensitiveVars: []string{
			"*PASSWORD*", "*SECRET*", "*KEY*", "*TOKEN*", "*API_KEY*",
		},
	}
}

// InterpolateConfig performs environment variable interpolation on the entire configuration
func (ie *InterpolationEngine) InterpolateConfig(config *Config) (*Config, error) {
	// Create a copy of the config to avoid modifying the original
	configCopy := *config

	// Interpolate each section
	if config.PVM != nil {
		pvmCopy := *config.PVM
		if err := ie.interpolateStruct(&pvmCopy); err != nil {
			return nil, fmt.Errorf("PVM config interpolation failed: %w", err)
		}
		configCopy.PVM = &pvmCopy
	}

	if config.PVX != nil {
		pvxCopy := *config.PVX
		if err := ie.interpolateStruct(&pvxCopy); err != nil {
			return nil, fmt.Errorf("PVX config interpolation failed: %w", err)
		}
		configCopy.PVX = &pvxCopy
	}

	if config.PM != nil {
		pviCopy := *config.PM
		if err := ie.interpolateStruct(&pviCopy); err != nil {
			return nil, fmt.Errorf("PM config interpolation failed: %w", err)
		}
		configCopy.PM = &pviCopy
	}

	if config.PSC != nil {
		pscCopy := *config.PSC
		if err := ie.interpolateStruct(&pscCopy); err != nil {
			return nil, fmt.Errorf("PSC config interpolation failed: %w", err)
		}
		configCopy.PSC = &pscCopy
	}

	if config.MCP != nil {
		mcpCopy := *config.MCP
		if err := ie.interpolateStruct(&mcpCopy); err != nil {
			return nil, fmt.Errorf("MCP config interpolation failed: %w", err)
		}
		configCopy.MCP = &mcpCopy
	}

	return &configCopy, nil
}

// interpolateStruct uses reflection to interpolate all string fields in a struct
func (ie *InterpolationEngine) interpolateStruct(v interface{}) error {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", value.Kind())
	}

	structType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		if err := ie.interpolateField(field, fieldType.Name); err != nil {
			return fmt.Errorf("failed to interpolate field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// interpolateField interpolates a single field based on its type
func (ie *InterpolationEngine) interpolateField(field reflect.Value, fieldName string) error {
	switch field.Kind() {
	case reflect.String:
		original := field.String()
		interpolated, err := ie.InterpolateString(original)
		if err != nil {
			return fmt.Errorf("string interpolation failed: %w", err)
		}
		field.SetString(interpolated)

	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			// Handle []string fields
			for i := 0; i < field.Len(); i++ {
				elem := field.Index(i)
				original := elem.String()
				interpolated, err := ie.InterpolateString(original)
				if err != nil {
					return fmt.Errorf("slice element interpolation failed: %w", err)
				}
				elem.SetString(interpolated)
			}
		}

	case reflect.Map:
		if field.Type().Key().Kind() == reflect.String && field.Type().Elem().Kind() == reflect.String {
			// Handle map[string]string fields
			for _, key := range field.MapKeys() {
				value := field.MapIndex(key)
				original := value.String()
				interpolated, err := ie.InterpolateString(original)
				if err != nil {
					return fmt.Errorf("map value interpolation failed: %w", err)
				}
				field.SetMapIndex(key, reflect.ValueOf(interpolated))
			}
		}

	case reflect.Int:
		// Try to interpolate as string first, then convert to int
		if fieldName != "" {
			// For int fields, we don't interpolate directly but could support env vars that contain numbers
		}
	}

	return nil
}

// InterpolateString performs environment variable interpolation on a single string
func (ie *InterpolationEngine) InterpolateString(input string) (string, error) {
	// Reset visited variables for each top-level interpolation
	ie.visitedVars = make(map[string]bool)

	return ie.interpolateStringRecursive(input, 0)
}

// interpolateStringRecursive handles recursive interpolation with cycle detection
func (ie *InterpolationEngine) interpolateStringRecursive(input string, depth int) (string, error) {
	if depth > ie.maxDepth {
		return "", errors.NewConfigError("010",
			"Interpolation depth exceeded (possible cycle)", nil).
			WithDetail(fmt.Sprintf("Max depth: %d", ie.maxDepth))
	}

	result := input
	changed := true

	// Keep looping until no more substitutions are possible
	for changed {

		// Find the first ${...} pattern
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		// Find the matching closing brace, handling nested braces
		braceCount := 0
		end := -1
		for i := start + 2; i < len(result); i++ {
			if result[i] == '{' {
				braceCount++
			} else if result[i] == '}' {
				if braceCount == 0 {
					end = i
					break
				}
				braceCount--
			}
		}

		if end == -1 {
			// No matching closing brace found, skip this
			break
		}

		varExpression := result[start+2 : end] // VAR or VAR:-default

		// Parse variable name and default value
		varName, defaultValue := ie.parseVarExpression(varExpression)

		// Check for cycles
		if ie.visitedVars[varName] {
			return "", errors.NewConfigError("011",
				"Circular reference detected in environment variable interpolation", nil).
				WithDetail(fmt.Sprintf("Variable: %s", varName))
		}

		// Mark variable as visited
		ie.visitedVars[varName] = true

		// Get environment variable value
		envValue := os.Getenv(varName)

		var replacementValue string
		switch {
		case envValue != "":
			// Recursively interpolate the environment variable value
			var err error
			replacementValue, err = ie.interpolateStringRecursive(envValue, depth+1)
			if err != nil {
				return "", fmt.Errorf("recursive interpolation of %s failed: %w", varName, err)
			}
		case defaultValue != "":
			// Use default value and interpolate it recursively
			var err error
			replacementValue, err = ie.interpolateStringRecursive(defaultValue, depth+1)
			if err != nil {
				return "", fmt.Errorf("recursive interpolation of default value for %s failed: %w", varName, err)
			}
		default:
			// No value and no default - leave as empty string
			replacementValue = ""
		}

		// Replace the pattern with the interpolated value
		result = result[:start] + replacementValue + result[end+1:]
		changed = true

		// Unmark variable as visited for other interpolations
		delete(ie.visitedVars, varName)
	}

	return result, nil
}

// parseVarExpression parses VAR or VAR:-default expressions
func (ie *InterpolationEngine) parseVarExpression(expr string) (varName, defaultValue string) {
	// Check for default value syntax: VAR:-default
	if strings.Contains(expr, ":-") {
		parts := strings.SplitN(expr, ":-", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), parts[1]
		}
	}

	// No default value
	return strings.TrimSpace(expr), ""
}

// InterpolateAndConvert performs interpolation and type conversion for specific types
func (ie *InterpolationEngine) InterpolateAndConvert(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return ie.InterpolateString(v)

	case int:
		// For int values, we don't interpolate directly
		return v, nil

	case bool:
		// For bool values, we don't interpolate directly
		return v, nil

	case []string:
		result := make([]string, len(v))
		for i, str := range v {
			interpolated, err := ie.InterpolateString(str)
			if err != nil {
				return nil, fmt.Errorf("failed to interpolate slice element %d: %w", i, err)
			}
			result[i] = interpolated
		}
		return result, nil

	case map[string]string:
		result := make(map[string]string)
		for key, val := range v {
			interpolatedVal, err := ie.InterpolateString(val)
			if err != nil {
				return nil, fmt.Errorf("failed to interpolate map value for key %s: %w", key, err)
			}
			result[key] = interpolatedVal
		}
		return result, nil

	default:
		// For other types, return as-is
		return value, nil
	}
}

// ConvertInterpolatedValue converts an interpolated string to the appropriate type
func (ie *InterpolationEngine) ConvertInterpolatedValue(interpolated string, targetType reflect.Type) (interface{}, error) {
	switch targetType.Kind() {
	case reflect.String:
		return interpolated, nil

	case reflect.Int:
		val, err := strconv.Atoi(interpolated)
		if err != nil {
			return nil, fmt.Errorf("failed to convert '%s' to int: %w", interpolated, err)
		}
		return val, nil

	case reflect.Bool:
		val, err := strconv.ParseBool(interpolated)
		if err != nil {
			return nil, fmt.Errorf("failed to convert '%s' to bool: %w", interpolated, err)
		}
		return val, nil

	default:
		return interpolated, nil
	}
}

// IsSensitiveVariable checks if a variable name matches sensitive patterns
func (ie *InterpolationEngine) IsSensitiveVariable(varName string) bool {
	upperName := strings.ToUpper(varName)
	for _, pattern := range ie.sensitiveVars {
		// Simple pattern matching - * means any characters
		pattern = strings.ToUpper(pattern)
		if strings.Contains(pattern, "*") {
			// Convert glob-like pattern to simple substring matching
			pattern = strings.ReplaceAll(pattern, "*", "")
			if strings.Contains(upperName, pattern) {
				return true
			}
		} else if upperName == pattern {
			return true
		}
	}
	return false
}

// MaskSensitiveValue masks a sensitive environment variable value for logging
func (ie *InterpolationEngine) MaskSensitiveValue(varName, value string) string {
	if ie.IsSensitiveVariable(varName) {
		if len(value) <= 8 {
			return "***"
		}
		return value[:2] + "***" + value[len(value)-2:]
	}
	return value
}

// ValidateInterpolatedConfig validates the configuration after interpolation
func (ie *InterpolationEngine) ValidateInterpolatedConfig(config *Config) error {
	// Use existing validation methods
	validationErrors := config.Validate()
	if len(validationErrors) > 0 {
		messages := make([]string, len(validationErrors))
		for i, err := range validationErrors {
			messages[i] = err.Error()
		}

		return errors.NewConfigError("012",
			"Configuration validation failed after interpolation", nil).
			WithDetail(strings.Join(messages, "; "))
	}

	return nil
}

// InterpolationResult holds the result of interpolation with metadata
type InterpolationResult struct {
	Config           *Config
	InterpolatedVars map[string]string
	SensitiveVars    []string
	ValidationErrors []error
}

// InterpolateConfigWithResult performs interpolation and returns detailed results
func (ie *InterpolationEngine) InterpolateConfigWithResult(config *Config) (*InterpolationResult, error) {
	// Track interpolated variables
	interpolatedVars := make(map[string]string)
	var sensitiveVars []string

	// Perform interpolation
	interpolatedConfig, err := ie.InterpolateConfig(config)
	if err != nil {
		return nil, err
	}

	// Validate the interpolated configuration
	validationErrors := interpolatedConfig.Validate()

	return &InterpolationResult{
		Config:           interpolatedConfig,
		InterpolatedVars: interpolatedVars,
		SensitiveVars:    sensitiveVars,
		ValidationErrors: validationErrors,
	}, nil
}
