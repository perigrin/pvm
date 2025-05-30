// ABOUTME: Configuration profiles system for the PVM Ecosystem
// ABOUTME: Provides environment-specific configuration profiles with inheritance

package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"tamarou.com/pvm/internal/errors"
)

// Profile represents a configuration profile for specific environments
type Profile struct {
	Name        string            `toml:"name" json:"name"`
	Description string            `toml:"description,omitempty" json:"description,omitempty"`
	Environment string            `toml:"environment" json:"environment"`
	Inherits    []string          `toml:"inherits,omitempty" json:"inherits,omitempty"`
	Config      *Config           `toml:"config" json:"config"`
	Variables   map[string]string `toml:"variables,omitempty" json:"variables,omitempty"`
	Template    string            `toml:"template,omitempty" json:"template,omitempty"`
}

// ProfileManager manages configuration profiles
type ProfileManager struct {
	profilesDir     string
	profiles        map[string]*Profile
	templateManager *TemplateManager
}

// NewProfileManager creates a new profile manager
func NewProfileManager(profilesDir string, templateManager *TemplateManager) *ProfileManager {
	return &ProfileManager{
		profilesDir:     profilesDir,
		profiles:        make(map[string]*Profile),
		templateManager: templateManager,
	}
}

// LoadProfiles loads all profiles from the profiles directory
func (pm *ProfileManager) LoadProfiles() error {
	if _, err := os.Stat(pm.profilesDir); os.IsNotExist(err) {
		// Profiles directory doesn't exist, which is fine
		return nil
	}

	return filepath.WalkDir(pm.profilesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".profile.toml") {
			return nil
		}

		profile, err := pm.loadProfile(path)
		if err != nil {
			return errors.NewConfigError("P001",
				"Failed to load profile", err).
				WithLocation(path)
		}

		pm.profiles[profile.Name] = profile
		return nil
	})
}

// loadProfile loads a single profile from a file
func (pm *ProfileManager) loadProfile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var profile Profile
	if err := toml.Unmarshal(data, &profile); err != nil {
		return nil, err
	}

	// Validate profile
	if profile.Name == "" {
		return nil, fmt.Errorf("profile name is required")
	}

	if profile.Environment == "" {
		return nil, fmt.Errorf("profile environment is required")
	}

	return &profile, nil
}

// GetProfile retrieves a profile by name
func (pm *ProfileManager) GetProfile(name string) (*Profile, error) {
	profile, exists := pm.profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}
	return profile, nil
}

// ListProfiles returns a list of available profile names
func (pm *ProfileManager) ListProfiles() []string {
	names := make([]string, 0, len(pm.profiles))
	for name := range pm.profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListProfilesByEnvironment returns profiles for a specific environment
func (pm *ProfileManager) ListProfilesByEnvironment(environment string) []string {
	var names []string
	for _, profile := range pm.profiles {
		if profile.Environment == environment {
			names = append(names, profile.Name)
		}
	}
	sort.Strings(names)
	return names
}

// GetEnvironments returns a list of unique environments
func (pm *ProfileManager) GetEnvironments() []string {
	envs := make(map[string]bool)
	for _, profile := range pm.profiles {
		envs[profile.Environment] = true
	}

	result := make([]string, 0, len(envs))
	for env := range envs {
		result = append(result, env)
	}
	sort.Strings(result)
	return result
}

// ResolveProfile resolves a profile with inheritance and template rendering
func (pm *ProfileManager) ResolveProfile(profileName string, extraVariables map[string]string) (*Config, error) {
	// Use Viper-based implementation for better inheritance handling
	vpm := NewViperProfileManager(pm.profilesDir, pm.templateManager)
	config, err := vpm.ResolveProfileWithViper(profileName, extraVariables)
	if err != nil {
		return nil, err
	}

	// Apply environment variable interpolation
	ie := NewInterpolationEngine()
	interpolatedConfig, err := ie.InterpolateConfig(config)
	if err != nil {
		return nil, errors.NewConfigError("P004",
			"Failed to interpolate configuration", err).
			WithLocation("profile:" + profileName)
	}

	return interpolatedConfig, nil
}

// mergeVariables merges profile variables with extra variables
func (pm *ProfileManager) mergeVariables(profileVars, extraVars map[string]string) map[string]string {
	merged := make(map[string]string)

	// Start with profile variables
	for key, value := range profileVars {
		merged[key] = value
	}

	// Override with extra variables
	for key, value := range extraVars {
		merged[key] = value
	}

	return merged
}

// ValidateProfile validates a profile for correctness
func (pm *ProfileManager) ValidateProfile(profile *Profile) []error {
	var errors []error

	// Check required fields
	if profile.Name == "" {
		errors = append(errors, fmt.Errorf("profile name is required"))
	}

	if profile.Environment == "" {
		errors = append(errors, fmt.Errorf("profile environment is required"))
	}

	// Check inheritance chain
	for _, parentName := range profile.Inherits {
		if _, err := pm.GetProfile(parentName); err != nil {
			errors = append(errors, fmt.Errorf("parent profile '%s' not found", parentName))
		}
	}

	// Check template exists if specified
	if profile.Template != "" && pm.templateManager != nil {
		if _, err := pm.templateManager.GetTemplate(profile.Template); err != nil {
			errors = append(errors, fmt.Errorf("template '%s' not found", profile.Template))
		}
	}

	// Validate configuration if present
	if profile.Config != nil {
		if configErrors := profile.Config.Validate(); len(configErrors) > 0 {
			for _, configErr := range configErrors {
				errors = append(errors, fmt.Errorf("profile config validation error: %w", configErr))
			}
		}
	}

	// Check for circular inheritance
	if pm.hasCircularInheritance(profile.Name, profile.Inherits, make(map[string]bool)) {
		errors = append(errors, fmt.Errorf("circular inheritance detected"))
	}

	return errors
}

// hasCircularInheritance checks for circular inheritance in profiles
func (pm *ProfileManager) hasCircularInheritance(profileName string, inherits []string, visited map[string]bool) bool {
	if visited[profileName] {
		return true
	}

	visited[profileName] = true

	for _, parent := range inherits {
		if parentProfile, exists := pm.profiles[parent]; exists {
			if pm.hasCircularInheritance(parent, parentProfile.Inherits, visited) {
				return true
			}
		}
	}

	delete(visited, profileName)
	return false
}

// SaveProfile saves a profile to the profiles directory
func (pm *ProfileManager) SaveProfile(profile *Profile) error {
	// Validate profile first
	if errs := pm.ValidateProfile(profile); len(errs) > 0 {
		return fmt.Errorf("profile validation failed: %v", errs)
	}

	// Ensure profiles directory exists
	if err := os.MkdirAll(pm.profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	// Serialize profile
	data, err := toml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	// Write to file
	path := filepath.Join(pm.profilesDir, profile.Name+".profile.toml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	// Update in-memory cache
	pm.profiles[profile.Name] = profile

	return nil
}

// DeleteProfile removes a profile
func (pm *ProfileManager) DeleteProfile(name string) error {
	// Check if profile exists
	if _, exists := pm.profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	// Check if any other profiles inherit from this one
	for _, profile := range pm.profiles {
		for _, parent := range profile.Inherits {
			if parent == name {
				return fmt.Errorf("profile '%s' is inherited by '%s', cannot delete", name, profile.Name)
			}
		}
	}

	// Delete file
	path := filepath.Join(pm.profilesDir, name+".profile.toml")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete profile file: %w", err)
	}

	// Remove from cache
	delete(pm.profiles, name)

	return nil
}

// CreateProfileFromTemplate creates a new profile from a template
func (pm *ProfileManager) CreateProfileFromTemplate(name, environment, templateName string, variables map[string]string) (*Profile, error) {
	if pm.templateManager == nil {
		return nil, fmt.Errorf("template manager not available")
	}

	// Validate template exists
	if _, err := pm.templateManager.GetTemplate(templateName); err != nil {
		return nil, fmt.Errorf("template '%s' not found: %w", templateName, err)
	}

	profile := &Profile{
		Name:        name,
		Environment: environment,
		Template:    templateName,
		Variables:   variables,
	}

	return profile, nil
}

// GetProfilesInheritingFrom returns profiles that inherit from the given profile
func (pm *ProfileManager) GetProfilesInheritingFrom(profileName string) []string {
	var result []string

	for _, profile := range pm.profiles {
		for _, parent := range profile.Inherits {
			if parent == profileName {
				result = append(result, profile.Name)
				break
			}
		}
	}

	sort.Strings(result)
	return result
}
