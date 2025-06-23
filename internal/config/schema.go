// ABOUTME: Configuration schema validation for the PVM Ecosystem
// ABOUTME: Provides schema-based validation for templates and profiles

package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// SchemaField represents a configuration field schema
type SchemaField struct {
	Name        string                  `json:"name"`
	Type        string                  `json:"type"`
	Required    bool                    `json:"required"`
	Default     interface{}             `json:"default,omitempty"`
	Description string                  `json:"description,omitempty"`
	ValidValues []interface{}           `json:"valid_values,omitempty"`
	MinValue    *float64                `json:"min_value,omitempty"`
	MaxValue    *float64                `json:"max_value,omitempty"`
	Pattern     string                  `json:"pattern,omitempty"`
	Children    map[string]*SchemaField `json:"children,omitempty"`
}

// ConfigSchema represents the full configuration schema
type ConfigSchema struct {
	Version string                  `json:"version"`
	Fields  map[string]*SchemaField `json:"fields"`
}

// SchemaValidator provides advanced schema validation for configurations
type SchemaValidator struct {
	schema *ConfigSchema
}

// NewSchemaValidator creates a new schema validator with the built-in schema
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		schema: getBuiltinSchema(),
	}
}

// ValidateConfig validates a configuration against the schema
func (sv *SchemaValidator) ValidateConfig(config *Config) []error {
	var errors []error

	// Use reflection to validate against schema
	configValue := reflect.ValueOf(config).Elem()
	configType := configValue.Type()

	// Validate each section
	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Field(i)
		fieldType := configType.Field(i)

		// Get TOML tag name
		tagName := fieldType.Tag.Get("toml")
		if tagName == "" {
			tagName = strings.ToLower(fieldType.Name)
		}

		// Check if field exists in schema
		if schemaField, exists := sv.schema.Fields[tagName]; exists {
			if fieldErrors := sv.validateField(field, schemaField, tagName); len(fieldErrors) > 0 {
				errors = append(errors, fieldErrors...)
			}
		}
	}

	return errors
}

// validateField validates a single field against its schema
func (sv *SchemaValidator) validateField(value reflect.Value, schema *SchemaField, fieldPath string) []error {
	var errors []error

	// Handle nil pointers
	if value.Kind() == reflect.Ptr && value.IsNil() {
		if schema.Required {
			errors = append(errors, fmt.Errorf("field '%s' is required but not provided", fieldPath))
		}
		return errors
	}

	// Dereference pointers
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Validate based on schema type
	switch schema.Type {
	case "string":
		if fieldErrors := sv.validateString(value, schema, fieldPath); len(fieldErrors) > 0 {
			errors = append(errors, fieldErrors...)
		}
	case "int":
		if fieldErrors := sv.validateInt(value, schema, fieldPath); len(fieldErrors) > 0 {
			errors = append(errors, fieldErrors...)
		}
	case "bool":
		if fieldErrors := sv.validateBool(value, schema, fieldPath); len(fieldErrors) > 0 {
			errors = append(errors, fieldErrors...)
		}
	case "array":
		if fieldErrors := sv.validateArray(value, schema, fieldPath); len(fieldErrors) > 0 {
			errors = append(errors, fieldErrors...)
		}
	case "map":
		if fieldErrors := sv.validateMap(value, schema, fieldPath); len(fieldErrors) > 0 {
			errors = append(errors, fieldErrors...)
		}
	case "object":
		if fieldErrors := sv.validateObject(value, schema, fieldPath); len(fieldErrors) > 0 {
			errors = append(errors, fieldErrors...)
		}
	case "duration":
		if fieldErrors := sv.validateDuration(value, schema, fieldPath); len(fieldErrors) > 0 {
			errors = append(errors, fieldErrors...)
		}
	default:
		errors = append(errors, fmt.Errorf("unknown schema type '%s' for field '%s'", schema.Type, fieldPath))
	}

	return errors
}

// validateString validates string fields
func (sv *SchemaValidator) validateString(value reflect.Value, schema *SchemaField, fieldPath string) []error {
	var errors []error

	if value.Kind() != reflect.String {
		errors = append(errors, fmt.Errorf("field '%s' must be a string", fieldPath))
		return errors
	}

	str := value.String()

	// Check valid values
	if len(schema.ValidValues) > 0 {
		valid := false
		for _, validValue := range schema.ValidValues {
			if validStr, ok := validValue.(string); ok && str == validStr {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, fmt.Errorf("field '%s' must be one of %v, got '%s'", fieldPath, schema.ValidValues, str))
		}
	}

	// Check pattern (simplified regex check)
	if schema.Pattern != "" {
		if !sv.matchesPattern(str, schema.Pattern) {
			errors = append(errors, fmt.Errorf("field '%s' does not match required pattern '%s'", fieldPath, schema.Pattern))
		}
	}

	return errors
}

// validateInt validates integer fields
func (sv *SchemaValidator) validateInt(value reflect.Value, schema *SchemaField, fieldPath string) []error {
	var errors []error

	if value.Kind() != reflect.Int {
		errors = append(errors, fmt.Errorf("field '%s' must be an integer", fieldPath))
		return errors
	}

	intVal := float64(value.Int())

	// Check min value
	if schema.MinValue != nil && intVal < *schema.MinValue {
		errors = append(errors, fmt.Errorf("field '%s' must be at least %v, got %v", fieldPath, *schema.MinValue, intVal))
	}

	// Check max value
	if schema.MaxValue != nil && intVal > *schema.MaxValue {
		errors = append(errors, fmt.Errorf("field '%s' must be at most %v, got %v", fieldPath, *schema.MaxValue, intVal))
	}

	// Check valid values
	if len(schema.ValidValues) > 0 {
		valid := false
		for _, validValue := range schema.ValidValues {
			if validInt, ok := validValue.(float64); ok && intVal == validInt {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, fmt.Errorf("field '%s' must be one of %v, got %v", fieldPath, schema.ValidValues, intVal))
		}
	}

	return errors
}

// validateBool validates boolean fields
func (sv *SchemaValidator) validateBool(value reflect.Value, schema *SchemaField, fieldPath string) []error {
	var errors []error

	if value.Kind() != reflect.Bool {
		errors = append(errors, fmt.Errorf("field '%s' must be a boolean", fieldPath))
		return errors
	}

	return errors
}

// validateDuration validates time.Duration fields
func (sv *SchemaValidator) validateDuration(value reflect.Value, schema *SchemaField, fieldPath string) []error {
	var errors []error

	// Check if it's a duration type
	if value.Type().String() != "time.Duration" {
		errors = append(errors, fmt.Errorf("field '%s' must be a duration", fieldPath))
		return errors
	}

	// Convert duration to seconds for validation
	durationVal := value.Int() / int64(time.Second)

	// Check min value (must be positive for timeout)
	if schema.MinValue != nil && float64(durationVal) < *schema.MinValue {
		errors = append(errors, fmt.Errorf("field '%s' must be a positive duration, got %v seconds", fieldPath, durationVal))
	}

	// Check max value
	if schema.MaxValue != nil && float64(durationVal) > *schema.MaxValue {
		errors = append(errors, fmt.Errorf("field '%s' duration cannot exceed %v seconds, got %v", fieldPath, *schema.MaxValue, durationVal))
	}

	return errors
}

// validateArray validates array/slice fields
func (sv *SchemaValidator) validateArray(value reflect.Value, schema *SchemaField, fieldPath string) []error {
	var errors []error

	if value.Kind() != reflect.Slice {
		errors = append(errors, fmt.Errorf("field '%s' must be an array", fieldPath))
		return errors
	}

	// Validate each element if children schema is provided
	if schema.Children != nil {
		if elementSchema, exists := schema.Children["element"]; exists {
			for i := 0; i < value.Len(); i++ {
				element := value.Index(i)
				elementPath := fmt.Sprintf("%s[%d]", fieldPath, i)
				if elementErrors := sv.validateField(element, elementSchema, elementPath); len(elementErrors) > 0 {
					errors = append(errors, elementErrors...)
				}
			}
		}
	}

	return errors
}

// validateMap validates map fields
func (sv *SchemaValidator) validateMap(value reflect.Value, schema *SchemaField, fieldPath string) []error {
	var errors []error

	if value.Kind() != reflect.Map {
		errors = append(errors, fmt.Errorf("field '%s' must be a map", fieldPath))
		return errors
	}

	// Validate each value if children schema is provided
	if schema.Children != nil {
		if valueSchema, exists := schema.Children["value"]; exists {
			for _, key := range value.MapKeys() {
				mapValue := value.MapIndex(key)
				valuePath := fmt.Sprintf("%s.%s", fieldPath, key.String())
				if valueErrors := sv.validateField(mapValue, valueSchema, valuePath); len(valueErrors) > 0 {
					errors = append(errors, valueErrors...)
				}
			}
		}
	}

	return errors
}

// validateObject validates object/struct fields
func (sv *SchemaValidator) validateObject(value reflect.Value, schema *SchemaField, fieldPath string) []error {
	var errors []error

	if value.Kind() != reflect.Struct {
		errors = append(errors, fmt.Errorf("field '%s' must be an object", fieldPath))
		return errors
	}

	// Validate each field if children schema is provided
	if schema.Children != nil {
		valueType := value.Type()
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			fieldType := valueType.Field(i)

			// Get TOML tag name
			tagName := fieldType.Tag.Get("toml")
			if tagName == "" {
				tagName = strings.ToLower(fieldType.Name)
			}

			if childSchema, exists := schema.Children[tagName]; exists {
				childPath := fmt.Sprintf("%s.%s", fieldPath, tagName)
				if childErrors := sv.validateField(field, childSchema, childPath); len(childErrors) > 0 {
					errors = append(errors, childErrors...)
				}
			}
		}
	}

	return errors
}

// matchesPattern performs simple pattern matching (simplified regex)
func (sv *SchemaValidator) matchesPattern(value, pattern string) bool {
	// Simple pattern matching for common cases
	switch pattern {
	case "url":
		return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
	case "memory":
		return strings.HasSuffix(value, "MB") || strings.HasSuffix(value, "GB") ||
			strings.HasSuffix(value, "KB") || strings.HasSuffix(value, "TB")
	case "version":
		// Allow special version aliases
		if value == "latest" || value == "stable" {
			return true
		}
		// Check numeric version format
		parts := strings.Split(value, ".")
		if len(parts) < 2 || len(parts) > 3 {
			return false
		}
		for _, part := range parts {
			if _, err := strconv.Atoi(part); err != nil {
				return false
			}
		}
		return true
	case "path":
		return len(value) > 0 && (strings.HasPrefix(value, "/") || strings.Contains(value, "$"))
	default:
		// For now, assume other patterns are valid
		return true
	}
}

// getBuiltinSchema returns the built-in configuration schema
func getBuiltinSchema() *ConfigSchema {
	minZero := 0.0
	minPositive := 0.000001 // minimum positive value for durations
	maxPort := 65535.0
	maxTimeout := 3600.0
	maxBuildJobs := 64.0
	maxConcurrentRequests := 100.0

	return &ConfigSchema{
		Version: "1.0",
		Fields: map[string]*SchemaField{
			"pvm": {
				Name: "pvm",
				Type: "object",
				Children: map[string]*SchemaField{
					"default_perl": {
						Name:        "default_perl",
						Type:        "string",
						Required:    false,
						Description: "Default Perl version to use",
						Pattern:     "version",
					},
					"build_jobs": {
						Name:        "build_jobs",
						Type:        "int",
						Required:    false,
						Description: "Number of parallel build jobs",
						MinValue:    &minZero,
						MaxValue:    &maxBuildJobs,
					},
					"download_mirror": {
						Name:        "download_mirror",
						Type:        "string",
						Required:    false,
						Description: "Perl download mirror URL",
						Pattern:     "url",
					},
					"run_tests": {
						Name:        "run_tests",
						Type:        "bool",
						Required:    false,
						Description: "Run tests during Perl installation",
					},
				},
			},
			"pvx": {
				Name: "pvx",
				Type: "object",
				Children: map[string]*SchemaField{
					"isolation_level": {
						Name:        "isolation_level",
						Type:        "string",
						Required:    false,
						Description: "Script execution isolation level",
						ValidValues: []interface{}{"global", "local", "clean"},
					},
					"timeout": {
						Name:        "timeout",
						Type:        "int",
						Required:    false,
						Description: "Maximum execution time in seconds",
						MinValue:    &minZero,
						MaxValue:    &maxTimeout,
					},
					"max_memory": {
						Name:        "max_memory",
						Type:        "string",
						Required:    false,
						Description: "Maximum memory usage",
						Pattern:     "memory",
					},
				},
			},
			"pvi": {
				Name: "pvi",
				Type: "object",
				Children: map[string]*SchemaField{
					"preferred_installer": {
						Name:        "preferred_installer",
						Type:        "string",
						Required:    false,
						Description: "Preferred module installer",
						ValidValues: []interface{}{"auto", "cpanm", "cpan", "cpm"},
					},
					"default_mirror": {
						Name:        "default_mirror",
						Type:        "string",
						Required:    false,
						Description: "Default CPAN mirror URL",
						Pattern:     "url",
					},
				},
			},
			"psc": {
				Name: "psc",
				Type: "object",
				Children: map[string]*SchemaField{
					"type_definitions_path": {
						Name:        "type_definitions_path",
						Type:        "string",
						Required:    false,
						Description: "Path to type definitions",
						Pattern:     "path",
					},
					"strict_mode": {
						Name:        "strict_mode",
						Type:        "bool",
						Required:    false,
						Description: "Enable strict type checking",
					},
				},
			},
			"mcp_server": {
				Name: "mcp_server",
				Type: "object",
				Children: map[string]*SchemaField{
					"port": {
						Name:        "port",
						Type:        "int",
						Required:    false,
						Description: "MCP server port",
						MinValue:    &minZero,
						MaxValue:    &maxPort,
					},
					"max_concurrent_requests": {
						Name:        "max_concurrent_requests",
						Type:        "int",
						Required:    false,
						Description: "Maximum concurrent requests",
						MinValue:    &minZero,
						MaxValue:    &maxConcurrentRequests,
					},
					"embedding_provider": {
						Name:        "embedding_provider",
						Type:        "string",
						Required:    false,
						Description: "Embedding provider",
						ValidValues: []interface{}{"openai", "voyageai", "huggingface"},
					},
					"request_timeout": {
						Name:        "request_timeout",
						Type:        "duration",
						Required:    false,
						Description: "Request timeout duration",
						MinValue:    &minPositive,
						MaxValue:    &maxTimeout,
					},
				},
			},
		},
	}
}
