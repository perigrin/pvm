// ABOUTME: Tests for configuration profiles system
// ABOUTME: Validates profile management, inheritance, and template integration

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileManager(t *testing.T) {
	// Setup temporary directories
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	templatesDir := filepath.Join(tempDir, "templates")

	// Create managers
	tm := NewTemplateManager(templatesDir)
	pm := NewProfileManager(profilesDir, tm)

	t.Run("EmptyProfilesDir", func(t *testing.T) {
		// Should not error when profiles directory doesn't exist
		err := pm.LoadProfiles()
		if err != nil {
			t.Errorf("LoadProfiles should not error with missing directory: %v", err)
		}

		profiles := pm.ListProfiles()
		if len(profiles) != 0 {
			t.Errorf("Expected 0 profiles, got %d", len(profiles))
		}
	})

	t.Run("BasicProfile", func(t *testing.T) {
		// Create profiles directory
		if err := os.MkdirAll(profilesDir, 0755); err != nil {
			t.Fatalf("Failed to create profiles directory: %v", err)
		}

		// Create a basic profile
		profileContent := `name = "development"
description = "Development environment profile"
environment = "dev"

[config.pvm]
default_perl = "5.38.0"
build_jobs = 2
run_tests = false

[config.pvx]
isolation_level = "low"
timeout = 600`

		profilePath := filepath.Join(profilesDir, "development.profile.toml")
		if err := os.WriteFile(profilePath, []byte(profileContent), 0644); err != nil {
			t.Fatalf("Failed to write profile file: %v", err)
		}

		// Load profiles
		if err := pm.LoadProfiles(); err != nil {
			t.Fatalf("Failed to load profiles: %v", err)
		}

		// Check profile was loaded
		profiles := pm.ListProfiles()
		if len(profiles) != 1 {
			t.Errorf("Expected 1 profile, got %d", len(profiles))
		}
		if profiles[0] != "development" {
			t.Errorf("Expected profile 'development', got '%s'", profiles[0])
		}

		// Get profile
		profile, err := pm.GetProfile("development")
		if err != nil {
			t.Errorf("Failed to get profile: %v", err)
		}
		if profile.Name != "development" {
			t.Errorf("Expected profile name 'development', got '%s'", profile.Name)
		}
		if profile.Environment != "dev" {
			t.Errorf("Expected environment 'dev', got '%s'", profile.Environment)
		}

		// Resolve profile
		config, err := pm.ResolveProfile("development", nil)
		if err != nil {
			t.Errorf("Failed to resolve profile: %v", err)
		}

		// Verify resolved values
		if config.PVM.DefaultPerl != "5.38.0" {
			t.Errorf("Expected DefaultPerl '5.38.0', got '%s'", config.PVM.DefaultPerl)
		}
		if config.PVM.BuildJobs != 2 {
			t.Errorf("Expected BuildJobs 2, got %d", config.PVM.BuildJobs)
		}
		if config.PVX.IsolationLevel != "low" {
			t.Errorf("Expected IsolationLevel 'low', got '%s'", config.PVX.IsolationLevel)
		}
	})

	t.Run("ProfileInheritance", func(t *testing.T) {
		// Create base profile
		baseContent := `name = "base"
description = "Base profile"
environment = "base"

[config.pvm]
run_tests = true
build_jobs = 4

[config.pvx]
cache_modules = true`

		basePath := filepath.Join(profilesDir, "base.profile.toml")
		if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
			t.Fatalf("Failed to write base profile: %v", err)
		}

		// Create production profile that inherits from base
		prodContent := `name = "production"
description = "Production environment"
environment = "prod"
inherits = ["base"]

[config.pvm]
default_perl = "5.36.0"
build_jobs = 8

[config.pvx]
isolation_level = "clean"`

		prodPath := filepath.Join(profilesDir, "production.profile.toml")
		if err := os.WriteFile(prodPath, []byte(prodContent), 0644); err != nil {
			t.Fatalf("Failed to write production profile: %v", err)
		}

		// Reload profiles
		pm = NewProfileManager(profilesDir, tm)
		if err := pm.LoadProfiles(); err != nil {
			t.Fatalf("Failed to load profiles: %v", err)
		}

		// Resolve production profile
		config, err := pm.ResolveProfile("production", nil)
		if err != nil {
			t.Errorf("Failed to resolve production profile: %v", err)
		}

		// Verify inheritance worked
		if config.PVM.RunTests != true {
			t.Errorf("Expected RunTests true (from base), got %v", config.PVM.RunTests)
		}
		if config.PVM.BuildJobs != 8 {
			t.Errorf("Expected BuildJobs 8 (overridden in production), got %d", config.PVM.BuildJobs)
		}
		if config.PVM.DefaultPerl != "5.36.0" {
			t.Errorf("Expected DefaultPerl '5.36.0' (from production), got '%s'", config.PVM.DefaultPerl)
		}
		if config.PVX.CacheModules != true {
			t.Errorf("Expected CacheModules true (from base), got %v", config.PVX.CacheModules)
		}
		if config.PVX.IsolationLevel != "clean" {
			t.Errorf("Expected IsolationLevel 'clean' (from production), got '%s'", config.PVX.IsolationLevel)
		}
	})

	t.Run("ProfileWithTemplate", func(t *testing.T) {
		// Create templates and profiles directories
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatalf("Failed to create templates directory: %v", err)
		}
		if err := os.MkdirAll(profilesDir, 0755); err != nil {
			t.Fatalf("Failed to create profiles directory: %v", err)
		}

		templateContent := `name = "web-app"
description = "Web application template"

content = '''
[config.pvm]
default_perl = "{{.perl_version}}"
build_jobs = {{.build_jobs}}

[config.pvx]
isolation_level = "{{.isolation_level}}"
timeout = {{.timeout}}
'''`

		templatePath := filepath.Join(templatesDir, "web-app.template.toml")
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		// Load templates
		if err := tm.LoadTemplates(); err != nil {
			t.Fatalf("Failed to load templates: %v", err)
		}

		// Create profile that uses template
		profileContent := `name = "web-dev"
description = "Web development profile"
environment = "dev"
template = "web-app"

[variables]
perl_version = "5.38.0"
build_jobs = "6"
isolation_level = "clean"
timeout = "900"`

		profilePath := filepath.Join(profilesDir, "web-dev.profile.toml")
		if err := os.WriteFile(profilePath, []byte(profileContent), 0644); err != nil {
			t.Fatalf("Failed to write web-dev profile: %v", err)
		}

		// Reload profiles
		pm = NewProfileManager(profilesDir, tm)
		if err := pm.LoadProfiles(); err != nil {
			t.Fatalf("Failed to load profiles: %v", err)
		}

		// Resolve profile with template
		config, err := pm.ResolveProfile("web-dev", nil)
		if err != nil {
			t.Errorf("Failed to resolve profile with template: %v", err)
		}

		// Verify template was rendered correctly
		if config.PVM.DefaultPerl != "5.38.0" {
			t.Errorf("Expected DefaultPerl '5.38.0', got '%s'", config.PVM.DefaultPerl)
		}
		if config.PVM.BuildJobs != 6 {
			t.Errorf("Expected BuildJobs 6, got %d", config.PVM.BuildJobs)
		}
		if config.PVX.IsolationLevel != "clean" {
			t.Errorf("Expected IsolationLevel 'clean', got '%s'", config.PVX.IsolationLevel)
		}
		if config.PVX.Timeout != 900 {
			t.Errorf("Expected Timeout 900, got %d", config.PVX.Timeout)
		}
	})

	t.Run("EnvironmentListing", func(t *testing.T) {
		// Check environments
		environments := pm.GetEnvironments()
		expectedEnvs := []string{"base", "dev", "prod"}
		if len(environments) != len(expectedEnvs) {
			t.Errorf("Expected %d environments, got %d", len(expectedEnvs), len(environments))
		}

		// Check profiles by environment
		devProfiles := pm.ListProfilesByEnvironment("dev")
		if len(devProfiles) < 1 {
			t.Error("Expected at least 1 dev profile")
		}

		prodProfiles := pm.ListProfilesByEnvironment("prod")
		if len(prodProfiles) < 1 {
			t.Error("Expected at least 1 prod profile")
		}
	})

	t.Run("ProfileValidation", func(t *testing.T) {
		// Test validation with missing name
		invalidProfile := &Profile{
			Environment: "test",
		}

		errors := pm.ValidateProfile(invalidProfile)
		if len(errors) == 0 {
			t.Error("Expected validation errors for profile without name")
		}

		// Test validation with missing environment
		invalidProfile = &Profile{
			Name: "test",
		}

		errors = pm.ValidateProfile(invalidProfile)
		if len(errors) == 0 {
			t.Error("Expected validation errors for profile without environment")
		}

		// Test validation with non-existent parent
		invalidProfile = &Profile{
			Name:        "test",
			Environment: "test",
			Inherits:    []string{"non-existent"},
		}

		errors = pm.ValidateProfile(invalidProfile)
		if len(errors) == 0 {
			t.Error("Expected validation errors for profile with non-existent parent")
		}
	})

	t.Run("CircularInheritance", func(t *testing.T) {
		// Create circular inheritance profiles
		circular1Content := `name = "circular1"
environment = "test"
inherits = ["circular2"]`

		circular2Content := `name = "circular2"
environment = "test"
inherits = ["circular1"]`

		circular1Path := filepath.Join(profilesDir, "circular1.profile.toml")
		circular2Path := filepath.Join(profilesDir, "circular2.profile.toml")

		if err := os.WriteFile(circular1Path, []byte(circular1Content), 0644); err != nil {
			t.Fatalf("Failed to write circular1 profile: %v", err)
		}
		if err := os.WriteFile(circular2Path, []byte(circular2Content), 0644); err != nil {
			t.Fatalf("Failed to write circular2 profile: %v", err)
		}

		// Reload profiles
		pm = NewProfileManager(profilesDir, tm)
		if err := pm.LoadProfiles(); err != nil {
			t.Fatalf("Failed to load profiles: %v", err)
		}

		// Test circular inheritance detection
		circular1, _ := pm.GetProfile("circular1")
		errors := pm.ValidateProfile(circular1)

		hasCircularError := false
		for _, err := range errors {
			if strings.Contains(err.Error(), "circular") {
				hasCircularError = true
				break
			}
		}
		if !hasCircularError {
			t.Error("Expected circular inheritance detection")
		}
	})

	t.Run("SaveAndDeleteProfile", func(t *testing.T) {
		// Create a new profile
		newProfile := &Profile{
			Name:        "test-save",
			Description: "Test save profile",
			Environment: "test",
			Variables: map[string]string{
				"test_var": "test_value",
			},
		}
		newProfile.Config = NewDefaultConfig()
		newProfile.Config.PVM.DefaultPerl = "5.40.0"

		// Save profile
		if err := pm.SaveProfile(newProfile); err != nil {
			t.Errorf("Failed to save profile: %v", err)
		}

		// Verify profile was saved and loaded
		if err := pm.LoadProfiles(); err != nil {
			t.Errorf("Failed to reload profiles: %v", err)
		}

		savedProfile, err := pm.GetProfile("test-save")
		if err != nil {
			t.Errorf("Failed to get saved profile: %v", err)
		}
		if savedProfile.Name != "test-save" {
			t.Errorf("Expected profile name 'test-save', got '%s'", savedProfile.Name)
		}

		// Delete profile
		if err := pm.DeleteProfile("test-save"); err != nil {
			t.Errorf("Failed to delete profile: %v", err)
		}

		// Verify profile was deleted
		if _, err := pm.GetProfile("test-save"); err == nil {
			t.Error("Expected error when getting deleted profile")
		}
	})

	t.Run("CreateProfileFromTemplate", func(t *testing.T) {
		// Create profile from template
		variables := map[string]string{
			"perl_version":    "5.38.0",
			"build_jobs":      "4",
			"isolation_level": "clean",
			"timeout":         "1200",
		}

		profile, err := pm.CreateProfileFromTemplate("test-from-template", "test", "web-app", variables)
		if err != nil {
			t.Errorf("Failed to create profile from template: %v", err)
		}

		if profile.Name != "test-from-template" {
			t.Errorf("Expected profile name 'test-from-template', got '%s'", profile.Name)
		}
		if profile.Environment != "test" {
			t.Errorf("Expected environment 'test', got '%s'", profile.Environment)
		}
		if profile.Template != "web-app" {
			t.Errorf("Expected template 'web-app', got '%s'", profile.Template)
		}
	})

	t.Run("ErrorCases", func(t *testing.T) {
		// Test getting non-existent profile
		if _, err := pm.GetProfile("non-existent"); err == nil {
			t.Error("Expected error when getting non-existent profile")
		}

		// Test resolving non-existent profile
		if _, err := pm.ResolveProfile("non-existent", nil); err == nil {
			t.Error("Expected error when resolving non-existent profile")
		}

		// Test deleting profile with dependents
		// Create dependent profile first
		dependentContent := `name = "dependent"
environment = "test"
inherits = ["base"]`

		dependentPath := filepath.Join(profilesDir, "dependent.profile.toml")
		if err := os.WriteFile(dependentPath, []byte(dependentContent), 0644); err != nil {
			t.Fatalf("Failed to write dependent profile: %v", err)
		}

		// Reload profiles
		pm = NewProfileManager(profilesDir, tm)
		if err := pm.LoadProfiles(); err != nil {
			t.Fatalf("Failed to load profiles: %v", err)
		}

		// Try to delete base profile (should fail due to dependent)
		if err := pm.DeleteProfile("base"); err == nil {
			t.Error("Expected error when deleting profile with dependents")
		}
	})
}
