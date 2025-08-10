// ABOUTME: Viper-based configuration management for proper profile inheritance
// ABOUTME: Solves boolean field merging issues with explicit field tracking

package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	"github.com/spf13/viper"
)

// ViperProfileManager uses Viper for advanced configuration management
type ViperProfileManager struct {
	profilesDir     string
	templateManager *TemplateManager
	profileCache    map[string]*viper.Viper
}

// NewViperProfileManager creates a new Viper-based profile manager
func NewViperProfileManager(profilesDir string, templateManager *TemplateManager) *ViperProfileManager {
	return &ViperProfileManager{
		profilesDir:     profilesDir,
		templateManager: templateManager,
		profileCache:    make(map[string]*viper.Viper),
	}
}

// LoadProfile loads a single profile using Viper
func (vpm *ViperProfileManager) LoadProfile(profileName string) (*viper.Viper, error) {
	// Check cache first
	if v, exists := vpm.profileCache[profileName]; exists {
		return v, nil
	}

	profilePath := filepath.Join(vpm.profilesDir, profileName+".profile.toml")
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile %s not found", profileName)
	}

	// Create new Viper instance for this profile
	v := viper.New()
	v.SetConfigFile(profilePath)
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read profile %s: %w", profileName, err)
	}

	// Cache the loaded profile
	vpm.profileCache[profileName] = v

	return v, nil
}

// ResolveProfileWithViper resolves a profile with proper inheritance using Viper
func (vpm *ViperProfileManager) ResolveProfileWithViper(profileName string, extraVariables map[string]string) (*Config, error) {
	// Load the profile
	profileViper, err := vpm.LoadProfile(profileName)
	if err != nil {
		return nil, err
	}

	// Check if profile uses a template and render it
	if templateName := profileViper.GetString("template"); templateName != "" {
		// Merge profile variables with extra variables
		variables := make(map[string]string)
		if profileViper.IsSet("variables") {
			profileVariables := profileViper.GetStringMapString("variables")
			for k, v := range profileVariables {
				variables[k] = v
			}
		}
		for k, v := range extraVariables {
			variables[k] = v
		}

		// Get the template and render it to TOML text
		template, err := vpm.templateManager.GetTemplate(templateName)
		if err != nil {
			return nil, fmt.Errorf("failed to get template %s for profile %s: %w", templateName, profileName, err)
		}

		// Render template content to TOML text
		renderedTOML, err := vpm.renderTemplateToTOML(template, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to render template %s for profile %s: %w", templateName, profileName, err)
		}

		// Create new Viper instance with rendered config
		templateViper := viper.New()
		templateViper.SetConfigType("toml")
		if err := templateViper.ReadConfig(strings.NewReader(renderedTOML)); err != nil {
			return nil, fmt.Errorf("failed to parse rendered template: %w", err)
		}

		// Use the rendered config as the base
		profileViper = templateViper
	}

	// Create a new Viper instance for the resolved configuration
	resolvedViper := viper.New()

	// Set defaults from our default configuration
	vpm.setDefaults(resolvedViper)

	// Resolve inheritance chain
	if err := vpm.resolveInheritanceChain(profileViper, resolvedViper, extraVariables); err != nil {
		return nil, fmt.Errorf("failed to resolve inheritance for profile %s: %w", profileName, err)
	}

	// Apply the current profile's configuration (this will only override explicitly set values)
	vpm.mergeProfileConfig(profileViper, resolvedViper)

	// Convert Viper config to our Config struct
	config, err := vpm.viperToConfig(resolvedViper)
	if err != nil {
		return nil, fmt.Errorf("failed to convert viper config to Config struct: %w", err)
	}

	return config, nil
}

// setDefaults sets default values in Viper from our default configuration
func (vpm *ViperProfileManager) setDefaults(v *viper.Viper) {
	defaultConfig := NewDefaultConfig()

	// PVM defaults
	v.SetDefault("config.pvm.default_perl", defaultConfig.PVM.DefaultPerl)
	v.SetDefault("config.pvm.build_jobs", defaultConfig.PVM.BuildJobs)
	v.SetDefault("config.pvm.download_mirror", defaultConfig.PVM.DownloadMirror)
	v.SetDefault("config.pvm.run_tests", defaultConfig.PVM.RunTests)
	v.SetDefault("config.pvm.patches_dir", defaultConfig.PVM.PatchesDir)
	v.SetDefault("config.pvm.compiler", defaultConfig.PVM.Compiler)
	for k, val := range defaultConfig.PVM.VersionAliases {
		v.SetDefault(fmt.Sprintf("config.pvm.version_aliases.%s", k), val)
	}

	// PVX defaults
	v.SetDefault("config.pvx.cache_modules", defaultConfig.PVX.CacheModules)
	v.SetDefault("config.pvx.cleanup_after", defaultConfig.PVX.CleanupAfter)
	v.SetDefault("config.pvx.isolation_level", defaultConfig.PVX.IsolationLevel)
	v.SetDefault("config.pvx.always_install_deps", defaultConfig.PVX.AlwaysInstallDeps)
	v.SetDefault("config.pvx.timeout", defaultConfig.PVX.Timeout)
	v.SetDefault("config.pvx.max_memory", defaultConfig.PVX.MaxMemory)
	v.SetDefault("config.pvx.isolated_output", defaultConfig.PVX.IsolatedOutput)
	v.SetDefault("config.pvx.save_output_dir", defaultConfig.PVX.SaveOutputDir)
	v.SetDefault("config.pvx.custom_module_path", defaultConfig.PVX.CustomModulePath)
	v.SetDefault("config.pvx.isolation_read_only_paths", defaultConfig.PVX.IsolationReadOnlyPaths)
	v.SetDefault("config.pvx.isolation_read_write_paths", defaultConfig.PVX.IsolationReadWritePaths)
	v.SetDefault("config.pvx.preserve_env_vars", defaultConfig.PVX.PreserveEnvVars)
	v.SetDefault("config.pvx.additional_module_paths", defaultConfig.PVX.AdditionalModulePaths)

	// PVI defaults
	v.SetDefault("config.pm.preferred_installer", defaultConfig.PM.PreferredInstaller)
	v.SetDefault("config.pm.default_mirror", defaultConfig.PM.DefaultMirror)
	v.SetDefault("config.pm.additional_mirrors", defaultConfig.PM.AdditionalMirrors)
	v.SetDefault("config.pm.metadata_source", defaultConfig.PM.MetadataSource)
	v.SetDefault("config.pm.metadata_url", defaultConfig.PM.MetadataURL)
	v.SetDefault("config.pm.cache_dir", defaultConfig.PM.CacheDir)
	v.SetDefault("config.pm.cache_ttl", defaultConfig.PM.CacheTTL)
	v.SetDefault("config.pm.test_during_install", defaultConfig.PM.TestDuringInstall)
	v.SetDefault("config.pm.cache_modules", defaultConfig.PM.CacheModules)
	v.SetDefault("config.pm.force_reinstall", defaultConfig.PM.ForceReinstall)
	v.SetDefault("config.pm.check_signatures", defaultConfig.PM.CheckSignatures)
	v.SetDefault("config.pm.disable_network", defaultConfig.PM.DisableNetwork)

	// PSC defaults
	v.SetDefault("config.psc.type_definitions_path", defaultConfig.PSC.TypeDefinitionsPath)
	v.SetDefault("config.psc.strict_mode", defaultConfig.PSC.StrictMode)
	v.SetDefault("config.psc.watch_exclude", defaultConfig.PSC.WatchExclude)
	v.SetDefault("config.psc.generate_missing_types", defaultConfig.PSC.GenerateMissingTypes)
	v.SetDefault("config.psc.check_before_run", defaultConfig.PSC.CheckBeforeRun)

	// MCP defaults
	v.SetDefault("config.mcp.port", defaultConfig.MCP.Port)
	v.SetDefault("config.mcp.host", defaultConfig.MCP.Host)
	v.SetDefault("config.mcp.auto_discover_projects", defaultConfig.MCP.AutoDiscoverProjects)
	v.SetDefault("config.mcp.auto_fix_errors", defaultConfig.MCP.AutoFixErrors)
}

// resolveInheritanceChain resolves the inheritance chain for a profile
func (vpm *ViperProfileManager) resolveInheritanceChain(profileViper, resolvedViper *viper.Viper, extraVariables map[string]string) error {
	// Get the list of profiles this profile inherits from
	inherits := profileViper.GetStringSlice("inherits")
	if len(inherits) == 0 {
		return nil
	}

	// Resolve each parent profile and merge its config
	for _, parentName := range inherits {
		parentConfig, err := vpm.ResolveProfileWithViper(parentName, extraVariables)
		if err != nil {
			return fmt.Errorf("failed to resolve parent profile %s: %w", parentName, err)
		}

		// Convert parent config back to Viper and merge
		parentViper := vpm.configToViper(parentConfig)
		vpm.mergeViperConfigs(parentViper, resolvedViper)
	}

	return nil
}

// mergeProfileConfig merges a profile's config section into the resolved config
func (vpm *ViperProfileManager) mergeProfileConfig(profileViper, resolvedViper *viper.Viper) {
	// Get all keys that are explicitly set in the profile (under config.*)
	for _, key := range profileViper.AllKeys() {
		if strings.HasPrefix(key, "config.") {
			value := profileViper.Get(key)
			resolvedViper.Set(key, value)
		}
	}
}

// mergeViperConfigs merges sourceViper into targetViper
func (vpm *ViperProfileManager) mergeViperConfigs(sourceViper, targetViper *viper.Viper) {
	// Merge all explicitly set config keys from source into target
	for _, key := range sourceViper.AllKeys() {
		if strings.HasPrefix(key, "config.") {
			// Only set if the key is explicitly set in source (not just a default)
			if sourceViper.IsSet(key) {
				value := sourceViper.Get(key)
				targetViper.Set(key, value)
			}
		}
	}
}

// configToViper converts a Config struct to a Viper instance
func (vpm *ViperProfileManager) configToViper(config *Config) *viper.Viper {
	v := viper.New()

	// PVM config
	if config.PVM != nil {
		v.Set("config.pvm.default_perl", config.PVM.DefaultPerl)
		v.Set("config.pvm.build_jobs", config.PVM.BuildJobs)
		v.Set("config.pvm.download_mirror", config.PVM.DownloadMirror)
		v.Set("config.pvm.run_tests", config.PVM.RunTests)
		v.Set("config.pvm.patches_dir", config.PVM.PatchesDir)
		v.Set("config.pvm.compiler", config.PVM.Compiler)
		if config.PVM.VersionAliases != nil {
			for k, val := range config.PVM.VersionAliases {
				v.Set(fmt.Sprintf("config.pvm.version_aliases.%s", k), val)
			}
		}
	}

	// PVX config
	if config.PVX != nil {
		v.Set("config.pvx.cache_modules", config.PVX.CacheModules)
		v.Set("config.pvx.cleanup_after", config.PVX.CleanupAfter)
		v.Set("config.pvx.isolation_level", config.PVX.IsolationLevel)
		v.Set("config.pvx.always_install_deps", config.PVX.AlwaysInstallDeps)
		v.Set("config.pvx.timeout", config.PVX.Timeout)
		v.Set("config.pvx.max_memory", config.PVX.MaxMemory)
		v.Set("config.pvx.isolated_output", config.PVX.IsolatedOutput)
		v.Set("config.pvx.save_output_dir", config.PVX.SaveOutputDir)
		v.Set("config.pvx.custom_module_path", config.PVX.CustomModulePath)
		v.Set("config.pvx.isolation_read_only_paths", config.PVX.IsolationReadOnlyPaths)
		v.Set("config.pvx.isolation_read_write_paths", config.PVX.IsolationReadWritePaths)
		v.Set("config.pvx.preserve_env_vars", config.PVX.PreserveEnvVars)
		v.Set("config.pvx.additional_module_paths", config.PVX.AdditionalModulePaths)
	}

	// PVI config
	if config.PM != nil {
		v.Set("config.pm.preferred_installer", config.PM.PreferredInstaller)
		v.Set("config.pm.default_mirror", config.PM.DefaultMirror)
		v.Set("config.pm.additional_mirrors", config.PM.AdditionalMirrors)
		v.Set("config.pm.metadata_source", config.PM.MetadataSource)
		v.Set("config.pm.metadata_url", config.PM.MetadataURL)
		v.Set("config.pm.cache_dir", config.PM.CacheDir)
		v.Set("config.pm.cache_ttl", config.PM.CacheTTL)
		v.Set("config.pm.test_during_install", config.PM.TestDuringInstall)
		v.Set("config.pm.cache_modules", config.PM.CacheModules)
		v.Set("config.pm.force_reinstall", config.PM.ForceReinstall)
		v.Set("config.pm.check_signatures", config.PM.CheckSignatures)
		v.Set("config.pm.disable_network", config.PM.DisableNetwork)
	}

	// PSC config
	if config.PSC != nil {
		v.Set("config.psc.type_definitions_path", config.PSC.TypeDefinitionsPath)
		v.Set("config.psc.strict_mode", config.PSC.StrictMode)
		v.Set("config.psc.watch_exclude", config.PSC.WatchExclude)
		v.Set("config.psc.generate_missing_types", config.PSC.GenerateMissingTypes)
		v.Set("config.psc.check_before_run", config.PSC.CheckBeforeRun)
	}

	// MCP config
	if config.MCP != nil {
		v.Set("config.mcp.port", config.MCP.Port)
		v.Set("config.mcp.host", config.MCP.Host)
		v.Set("config.mcp.auto_discover_projects", config.MCP.AutoDiscoverProjects)
		v.Set("config.mcp.auto_fix_errors", config.MCP.AutoFixErrors)
	}

	return v
}

// expandStringSlice expands environment variables in each string in a slice
func expandStringSlice(slice []string) []string {
	expanded := make([]string, len(slice))
	for i, s := range slice {
		expanded[i] = expandEnvironmentVariables(s)
	}
	return expanded
}

// viperToConfig converts a Viper instance to a Config struct
func (vpm *ViperProfileManager) viperToConfig(v *viper.Viper) (*Config, error) {
	config := &Config{
		PVM: &PVMConfig{
			DefaultPerl:    v.GetString("config.pvm.default_perl"),
			BuildJobs:      v.GetInt("config.pvm.build_jobs"),
			DownloadMirror: v.GetString("config.pvm.download_mirror"),
			RunTests:       v.GetBool("config.pvm.run_tests"),
			PatchesDir:     expandEnvironmentVariables(v.GetString("config.pvm.patches_dir")),
			Compiler:       v.GetString("config.pvm.compiler"),
			VersionAliases: v.GetStringMapString("config.pvm.version_aliases"),
		},
		PVX: &PVXConfig{
			CacheModules:            v.GetBool("config.pvx.cache_modules"),
			CleanupAfter:            v.GetBool("config.pvx.cleanup_after"),
			IsolationLevel:          v.GetString("config.pvx.isolation_level"),
			AlwaysInstallDeps:       v.GetBool("config.pvx.always_install_deps"),
			Timeout:                 v.GetInt("config.pvx.timeout"),
			MaxMemory:               v.GetString("config.pvx.max_memory"),
			IsolatedOutput:          v.GetBool("config.pvx.isolated_output"),
			SaveOutputDir:           expandEnvironmentVariables(v.GetString("config.pvx.save_output_dir")),
			CustomModulePath:        expandEnvironmentVariables(v.GetString("config.pvx.custom_module_path")),
			IsolationReadOnlyPaths:  expandStringSlice(v.GetStringSlice("config.pvx.isolation_read_only_paths")),
			IsolationReadWritePaths: expandStringSlice(v.GetStringSlice("config.pvx.isolation_read_write_paths")),
			PreserveEnvVars:         v.GetStringSlice("config.pvx.preserve_env_vars"),
			AdditionalModulePaths:   expandStringSlice(v.GetStringSlice("config.pvx.additional_module_paths")),
		},
		PM: &PMConfig{
			PreferredInstaller: v.GetString("config.pm.preferred_installer"),
			DefaultMirror:      v.GetString("config.pm.default_mirror"),
			AdditionalMirrors:  v.GetStringSlice("config.pm.additional_mirrors"),
			MetadataSource:     v.GetString("config.pm.metadata_source"),
			MetadataURL:        v.GetString("config.pm.metadata_url"),
			CacheDir:           expandEnvironmentVariables(v.GetString("config.pm.cache_dir")),
			CacheTTL:           v.GetInt("config.pm.cache_ttl"),
			TestDuringInstall:  v.GetBool("config.pm.test_during_install"),
			CacheModules:       v.GetBool("config.pm.cache_modules"),
			ForceReinstall:     v.GetBool("config.pm.force_reinstall"),
			CheckSignatures:    v.GetBool("config.pm.check_signatures"),
			DisableNetwork:     v.GetBool("config.pm.disable_network"),
		},
		PSC: &PSCConfig{
			TypeDefinitionsPath:  expandEnvironmentVariables(v.GetString("config.psc.type_definitions_path")),
			StrictMode:           v.GetBool("config.psc.strict_mode"),
			WatchExclude:         v.GetStringSlice("config.psc.watch_exclude"),
			GenerateMissingTypes: v.GetBool("config.psc.generate_missing_types"),
			CheckBeforeRun:       v.GetBool("config.psc.check_before_run"),
		},
		MCP: &MCPConfig{
			Port:                 v.GetInt("config.mcp.port"),
			Host:                 v.GetString("config.mcp.host"),
			AutoDiscoverProjects: v.GetBool("config.mcp.auto_discover_projects"),
			AutoFixErrors:        v.GetBool("config.mcp.auto_fix_errors"),
		},
	}

	return config, nil
}

// renderTemplateToTOML renders a template to TOML text
func (vpm *ViperProfileManager) renderTemplateToTOML(template *Template, variables map[string]string) (string, error) {
	// Since we can't access private methods, we'll duplicate the template rendering logic
	// This is a simplified version that handles the basic case without inheritance

	// Create template with function map (copied from TemplateManager)
	funcMap := texttemplate.FuncMap{
		"upper":      strings.ToUpper,
		"lower":      strings.ToLower,
		"title":      strings.ToTitle,
		"join":       strings.Join,
		"split":      strings.Split,
		"contains":   strings.Contains,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"trimSpace":  strings.TrimSpace,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
		"replace":    strings.ReplaceAll,
		"default": func(def string, val interface{}) string {
			if val == nil {
				return def
			}
			if str, ok := val.(string); ok && str != "" {
				return str
			}
			return def
		},
	}

	// Merge template variables with provided variables
	mergedVars := make(map[string]string)
	// Start with template variables (defaults)
	if template.Variables != nil {
		for key, value := range template.Variables {
			mergedVars[key] = value
		}
	}
	// Override with provided variables
	for key, value := range variables {
		mergedVars[key] = value
	}

	// Render the template
	tmpl, err := texttemplate.New(template.Name).Funcs(funcMap).Parse(template.Content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, mergedVars); err != nil {
		return "", err
	}

	return buf.String(), nil
}
