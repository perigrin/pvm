// ABOUTME: Advanced configuration merging for the PVM Ecosystem
// ABOUTME: Provides sophisticated merging strategies for nested configurations

package config

// MergeStrategy defines how different configuration values should be merged
type MergeStrategy int

const (
	// MergeReplace replaces the target value entirely with the source value
	MergeReplace MergeStrategy = iota

	// MergeAppend appends source arrays to target arrays
	MergeAppend

	// MergeDeep performs deep merging for nested objects
	MergeDeep

	// MergePriority respects explicit priority markers in configuration
	MergePriority
)

// MergerConfig controls advanced merging behavior
type MergerConfig struct {
	// ArrayStrategy specifies how to merge arrays
	ArrayStrategy MergeStrategy

	// MapStrategy specifies how to merge maps
	MapStrategy MergeStrategy

	// ConflictResolution specifies how to handle merge conflicts
	ConflictResolution MergeStrategy

	// PreserveOverrides maintains explicit override markers
	PreserveOverrides bool

	// AllowTypeCoercion enables type coercion during merging
	AllowTypeCoercion bool
}

// NewMergerConfig creates default merger configuration
func NewMergerConfig() *MergerConfig {
	return &MergerConfig{
		ArrayStrategy:      MergeReplace,
		MapStrategy:        MergeDeep,
		ConflictResolution: MergeReplace,
		PreserveOverrides:  true,
		AllowTypeCoercion:  false,
	}
}

// AdvancedMerger provides sophisticated configuration merging
type AdvancedMerger struct {
	config *MergerConfig
}

// NewAdvancedMerger creates a new advanced merger
func NewAdvancedMerger(config *MergerConfig) *AdvancedMerger {
	if config == nil {
		config = NewMergerConfig()
	}
	return &AdvancedMerger{config: config}
}

// MergeConfigs performs advanced merging of multiple configurations
func (m *AdvancedMerger) MergeConfigs(configs ...*Config) *Config {
	if len(configs) == 0 {
		return NewDefaultConfig()
	}

	result := NewDefaultConfig()

	for _, config := range configs {
		if config == nil {
			continue
		}

		// Merge each section with advanced strategies
		if config.PVM != nil {
			result.PVM = m.mergePVMConfigAdvanced(result.PVM, config.PVM)
		}
		if config.PVX != nil {
			result.PVX = m.mergePVXConfigAdvanced(result.PVX, config.PVX)
		}
		if config.PM != nil {
			result.PM = m.mergePMConfigAdvanced(result.PM, config.PM)
		}
		if config.PSC != nil {
			result.PSC = m.mergePSCConfigAdvanced(result.PSC, config.PSC)
		}
	}

	return result
}

// mergePVMConfigAdvanced performs advanced PVM configuration merging
func (m *AdvancedMerger) mergePVMConfigAdvanced(target, source *PVMConfig) *PVMConfig {
	if target == nil {
		target = &PVMConfig{}
	}
	if source == nil {
		return target
	}

	// Merge scalar fields
	m.mergeStringField(&target.DefaultPerl, source.DefaultPerl)
	m.mergeStringField(&target.DownloadMirror, source.DownloadMirror)
	m.mergeStringField(&target.PatchesDir, source.PatchesDir)
	m.mergeStringField(&target.Compiler, source.Compiler)

	// Merge numeric fields - only override if source has non-zero value
	if source.BuildJobs != 0 {
		target.BuildJobs = source.BuildJobs
	}
	// Boolean fields always merge (false can be a valid override)
	target.RunTests = source.RunTests

	// Merge map with advanced strategy
	target.VersionAliases = m.mergeStringMap(target.VersionAliases, source.VersionAliases)

	// Merge remotes: additive, same-name override wins from higher-precedence source
	target.Remotes = mergeRemoteConfigs(target.Remotes, source.Remotes)

	return target
}

// mergeRemoteConfigs merges two remote config slices. Source entries override
// target entries with the same name.
func mergeRemoteConfigs(target, source []PVMRemoteConfig) []PVMRemoteConfig {
	if len(source) == 0 {
		return target
	}
	if len(target) == 0 {
		result := make([]PVMRemoteConfig, len(source))
		copy(result, source)
		return result
	}

	merged := make([]PVMRemoteConfig, len(target))
	copy(merged, target)

	for _, s := range source {
		found := false
		for i, t := range merged {
			if t.Name == s.Name {
				merged[i] = s
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, s)
		}
	}
	return merged
}

// mergePVXConfigAdvanced performs advanced PVX configuration merging
func (m *AdvancedMerger) mergePVXConfigAdvanced(target, source *PVXConfig) *PVXConfig {
	if target == nil {
		target = &PVXConfig{}
	}
	if source == nil {
		return target
	}

	// Merge scalar fields
	m.mergeStringField(&target.IsolationLevel, source.IsolationLevel)
	m.mergeStringField(&target.MaxMemory, source.MaxMemory)
	m.mergeStringField(&target.SaveOutputDir, source.SaveOutputDir)
	m.mergeStringField(&target.CustomModulePath, source.CustomModulePath)

	// Merge numeric and boolean fields
	target.Timeout = source.Timeout
	target.CacheModules = source.CacheModules
	target.CleanupAfter = source.CleanupAfter
	target.AlwaysInstallDeps = source.AlwaysInstallDeps
	target.IsolatedOutput = source.IsolatedOutput

	// Merge arrays with advanced strategy
	target.IsolationReadOnlyPaths = m.mergeStringSlice(target.IsolationReadOnlyPaths, source.IsolationReadOnlyPaths)
	target.IsolationReadWritePaths = m.mergeStringSlice(target.IsolationReadWritePaths, source.IsolationReadWritePaths)
	target.PreserveEnvVars = m.mergeStringSlice(target.PreserveEnvVars, source.PreserveEnvVars)
	target.AdditionalModulePaths = m.mergeStringSlice(target.AdditionalModulePaths, source.AdditionalModulePaths)

	return target
}

// mergePMConfigAdvanced performs advanced PVI configuration merging
func (m *AdvancedMerger) mergePMConfigAdvanced(target, source *PMConfig) *PMConfig {
	if target == nil {
		target = &PMConfig{}
	}
	if source == nil {
		return target
	}

	// Merge scalar fields
	m.mergeStringField(&target.PreferredInstaller, source.PreferredInstaller)
	m.mergeStringField(&target.DefaultMirror, source.DefaultMirror)
	m.mergeStringField(&target.MetadataSource, source.MetadataSource)
	m.mergeStringField(&target.MetadataURL, source.MetadataURL)
	m.mergeStringField(&target.CacheDir, source.CacheDir)

	// Merge numeric and boolean fields
	target.CacheTTL = source.CacheTTL
	target.TestDuringInstall = source.TestDuringInstall
	target.CacheModules = source.CacheModules
	target.ForceReinstall = source.ForceReinstall
	target.CheckSignatures = source.CheckSignatures
	target.DisableNetwork = source.DisableNetwork

	// Merge arrays with advanced strategy
	target.AdditionalMirrors = m.mergeStringSlice(target.AdditionalMirrors, source.AdditionalMirrors)

	return target
}

// mergePSCConfigAdvanced performs advanced PSC configuration merging
func (m *AdvancedMerger) mergePSCConfigAdvanced(target, source *PSCConfig) *PSCConfig {
	if target == nil {
		target = &PSCConfig{}
	}
	if source == nil {
		return target
	}

	// Merge scalar fields
	m.mergeStringField(&target.TypeDefinitionsPath, source.TypeDefinitionsPath)

	// Merge boolean fields
	target.StrictMode = source.StrictMode
	target.GenerateMissingTypes = source.GenerateMissingTypes
	target.CheckBeforeRun = source.CheckBeforeRun

	// Merge arrays with advanced strategy
	target.WatchExclude = m.mergeStringSlice(target.WatchExclude, source.WatchExclude)

	return target
}

// Helper methods for different field types

func (m *AdvancedMerger) mergeStringField(target *string, source string) {
	if source != "" {
		*target = source
	}
}

func (m *AdvancedMerger) mergeStringSlice(target, source []string) []string {
	if source == nil {
		return target
	}

	switch m.config.ArrayStrategy {
	case MergeAppend:
		// Append unique items from source to target
		result := make([]string, len(target))
		copy(result, target)

		for _, item := range source {
			if !containsString(result, item) {
				result = append(result, item)
			}
		}
		return result

	case MergeReplace:
		// Replace target with source
		result := make([]string, len(source))
		copy(result, source)
		return result

	default:
		// Default to replace strategy
		result := make([]string, len(source))
		copy(result, source)
		return result
	}
}

func (m *AdvancedMerger) mergeStringMap(target, source map[string]string) map[string]string {
	if source == nil {
		return target
	}

	if target == nil {
		target = make(map[string]string)
	}

	switch m.config.MapStrategy {
	case MergeDeep:
		// Deep merge - source values override target values
		for key, value := range source {
			target[key] = value
		}
		return target

	case MergeReplace:
		// Replace entire map
		result := make(map[string]string, len(source))
		for key, value := range source {
			result[key] = value
		}
		return result

	default:
		// Default to deep merge
		for key, value := range source {
			target[key] = value
		}
		return target
	}
}

// Utility functions

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ConflictDetector identifies configuration conflicts during merging
type ConflictDetector struct {
	conflicts []MergeConflict
}

// MergeConflict represents a detected configuration conflict
type MergeConflict struct {
	Path        string      // Configuration path (e.g., "pvm.default_perl")
	TargetValue interface{} // Value in target configuration
	SourceValue interface{} // Value in source configuration
	Resolution  string      // How the conflict was resolved
}

// NewConflictDetector creates a new conflict detector
func NewConflictDetector() *ConflictDetector {
	return &ConflictDetector{
		conflicts: make([]MergeConflict, 0),
	}
}

// DetectConflicts analyzes two configurations and identifies conflicts
func (cd *ConflictDetector) DetectConflicts(target, source *Config) []MergeConflict {
	cd.conflicts = cd.conflicts[:0] // Reset conflicts

	// Compare PVM configurations
	if target.PVM != nil && source.PVM != nil {
		cd.comparePVMConfigs("pvm", target.PVM, source.PVM)
	}

	// Compare PVX configurations
	if target.PVX != nil && source.PVX != nil {
		cd.comparePVXConfigs("pvx", target.PVX, source.PVX)
	}

	// Compare PVI configurations
	if target.PM != nil && source.PM != nil {
		cd.comparePMConfigs("pvi", target.PM, source.PM)
	}

	// Compare PSC configurations
	if target.PSC != nil && source.PSC != nil {
		cd.comparePSCConfigs("psc", target.PSC, source.PSC)
	}

	return cd.conflicts
}

func (cd *ConflictDetector) comparePVMConfigs(prefix string, target, source *PVMConfig) {
	cd.compareStringField(prefix+".default_perl", target.DefaultPerl, source.DefaultPerl)
	cd.compareStringField(prefix+".download_mirror", target.DownloadMirror, source.DownloadMirror)
	cd.compareIntField(prefix+".build_jobs", target.BuildJobs, source.BuildJobs)
	cd.compareBoolField(prefix+".run_tests", target.RunTests, source.RunTests)
}

func (cd *ConflictDetector) comparePVXConfigs(prefix string, target, source *PVXConfig) {
	cd.compareStringField(prefix+".isolation_level", target.IsolationLevel, source.IsolationLevel)
	cd.compareStringField(prefix+".max_memory", target.MaxMemory, source.MaxMemory)
	cd.compareIntField(prefix+".timeout", target.Timeout, source.Timeout)
}

func (cd *ConflictDetector) comparePMConfigs(prefix string, target, source *PMConfig) {
	cd.compareStringField(prefix+".preferred_installer", target.PreferredInstaller, source.PreferredInstaller)
	cd.compareStringField(prefix+".default_mirror", target.DefaultMirror, source.DefaultMirror)
	cd.compareIntField(prefix+".cache_ttl", target.CacheTTL, source.CacheTTL)
}

func (cd *ConflictDetector) comparePSCConfigs(prefix string, target, source *PSCConfig) {
	cd.compareStringField(prefix+".type_definitions_path", target.TypeDefinitionsPath, source.TypeDefinitionsPath)
	cd.compareBoolField(prefix+".strict_mode", target.StrictMode, source.StrictMode)
}

func (cd *ConflictDetector) compareStringField(path, target, source string) {
	if target != "" && source != "" && target != source {
		cd.conflicts = append(cd.conflicts, MergeConflict{
			Path:        path,
			TargetValue: target,
			SourceValue: source,
			Resolution:  "source overrides target",
		})
	}
}

func (cd *ConflictDetector) compareIntField(path string, target, source int) {
	if target != 0 && source != 0 && target != source {
		cd.conflicts = append(cd.conflicts, MergeConflict{
			Path:        path,
			TargetValue: target,
			SourceValue: source,
			Resolution:  "source overrides target",
		})
	}
}

func (cd *ConflictDetector) compareBoolField(path string, target, source bool) {
	if target != source {
		cd.conflicts = append(cd.conflicts, MergeConflict{
			Path:        path,
			TargetValue: target,
			SourceValue: source,
			Resolution:  "source overrides target",
		})
	}
}

// GetConflicts returns all detected conflicts
func (cd *ConflictDetector) GetConflicts() []MergeConflict {
	return cd.conflicts
}

// HasConflicts returns true if any conflicts were detected
func (cd *ConflictDetector) HasConflicts() bool {
	return len(cd.conflicts) > 0
}
