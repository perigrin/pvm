// ABOUTME: Dependency management types and data structures
// ABOUTME: Defines types for cpanfile, dependency resolution, and bundle management

package dependencies

import (
	"time"
)

// Requirement represents a dependency requirement in a cpanfile
type Requirement struct {
	// Module is the module name
	Module string `json:"module"`

	// Version is the version constraint
	Version string `json:"version,omitempty"`

	// Phase indicates the dependency phase (runtime, build, test, develop)
	Phase string `json:"phase"`

	// Relationship indicates the dependency relationship (requires, recommends, suggests)
	Relationship string `json:"relationship"`

	// Optional indicates if the requirement is optional
	Optional bool `json:"optional,omitempty"`

	// Comment is any comment associated with this requirement
	Comment string `json:"comment,omitempty"`
}

// CPANFile represents a parsed cpanfile
type CPANFile struct {
	// Requirements lists all module requirements
	Requirements []Requirement `json:"requirements"`

	// Features maps feature names to their requirements
	Features map[string][]Requirement `json:"features,omitempty"`

	// Platforms maps platform names to their requirements
	Platforms map[string][]Requirement `json:"platforms,omitempty"`

	// PerlVersion specifies the minimum Perl version
	PerlVersion string `json:"perl_version,omitempty"`

	// Author information
	Author string `json:"author,omitempty"`

	// License information
	License string `json:"license,omitempty"`

	// Repository information
	Repository string `json:"repository,omitempty"`

	// BugTracker information
	BugTracker string `json:"bug_tracker,omitempty"`

	// Homepage information
	Homepage string `json:"homepage,omitempty"`
}

// Snapshot represents a cpanfile snapshot with locked versions
type Snapshot struct {
	// GeneratedAt indicates when the snapshot was created
	GeneratedAt time.Time `json:"generated_at"`

	// GeneratedBy indicates the tool that generated the snapshot
	GeneratedBy string `json:"generated_by"`

	// PerlVersion is the Perl version used to generate the snapshot
	PerlVersion string `json:"perl_version"`

	// Modules lists all modules with their exact versions
	Modules []*SnapshotModule `json:"modules"`

	// Hash is a checksum of the snapshot content
	Hash string `json:"hash,omitempty"`
}

// SnapshotModule represents a module in a snapshot with exact version
type SnapshotModule struct {
	// Name is the module name
	Name string `json:"name"`

	// Version is the exact installed version
	Version string `json:"version"`

	// Distribution is the distribution name
	Distribution string `json:"distribution,omitempty"`

	// Source is where the module was obtained from
	Source string `json:"source,omitempty"`

	// Checksum is the module checksum for verification
	Checksum string `json:"checksum,omitempty"`

	// Dependencies lists the module's dependencies
	Dependencies []string `json:"dependencies,omitempty"`
}

// DependencyGraph represents a dependency graph for modules
type DependencyGraph struct {
	// Nodes maps module names to their dependency information
	Nodes map[string]*DependencyNode `json:"nodes"`

	// RootNodes lists the initially requested modules
	RootNodes []*DependencyNode `json:"root_nodes"`

	// Edges represents dependency relationships
	Edges []*DependencyEdge `json:"edges"`

	// RootModules lists the initially requested modules (legacy)
	RootModules []string `json:"root_modules"`

	// ResolutionTime is when the graph was resolved
	ResolutionTime time.Time `json:"resolution_time"`
}

// DependencyNode represents a module node in the dependency graph
type DependencyNode struct {
	// Name is the module name
	Name string `json:"name"`

	// Version is the resolved version
	Version string `json:"version"`

	// Dependencies lists the direct dependencies of this node
	Dependencies []*DependencyNode `json:"dependencies"`

	// Constraints lists version constraints applied to this node
	Constraints []VersionConstraint `json:"constraints"`

	// VersionConstraint is the original version constraint
	VersionConstraint string `json:"version_constraint,omitempty"`

	// Phase indicates the dependency phase
	Phase string `json:"phase"`

	// Relationship indicates the dependency relationship
	Relationship string `json:"relationship"`

	// Depth is the depth in the dependency tree
	Depth int `json:"depth"`

	// Satisfied indicates if the dependency is satisfied
	Satisfied bool `json:"satisfied"`

	// CoreModule indicates if this is a Perl core module
	CoreModule bool `json:"core_module,omitempty"`

	// AlreadyInstalled indicates if the module is already installed
	AlreadyInstalled bool `json:"already_installed,omitempty"`
}

// DependencyEdge represents a dependency relationship between modules
type DependencyEdge struct {
	// From is the dependent module
	From string `json:"from"`

	// To is the dependency module
	To string `json:"to"`

	// Relationship is the type of dependency
	Relationship string `json:"relationship"`

	// Phase is the dependency phase
	Phase string `json:"phase"`

	// VersionConstraint is the version constraint
	VersionConstraint string `json:"version_constraint,omitempty"`
}

// Conflict represents a dependency conflict
type Conflict struct {
	// Module is the conflicting module name
	Module string `json:"module"`

	// Type indicates the type of conflict
	Type ConflictType `json:"type"`

	// Versions lists the conflicting versions
	Versions []string `json:"versions"`

	// Dependencies lists the dependencies that contribute to this conflict
	Dependencies []*ConflictDependency `json:"dependencies"`

	// RequiredVersions lists the conflicting version requirements (legacy)
	RequiredVersions []string `json:"required_versions"`

	// ConflictingModules lists the modules that have conflicting requirements (legacy)
	ConflictingModules []string `json:"conflicting_modules"`

	// Severity indicates the conflict severity
	Severity ConflictSeverity `json:"severity"`

	// Resolvable indicates if the conflict can be automatically resolved
	Resolvable bool `json:"resolvable"`

	// SuggestedResolution provides a suggested resolution
	SuggestedResolution string `json:"suggested_resolution,omitempty"`
}

// ConflictSeverity represents the severity of a dependency conflict
type ConflictSeverity int

const (
	// ConflictInfo represents an informational conflict
	ConflictInfo ConflictSeverity = iota

	// ConflictWarning represents a warning-level conflict
	ConflictWarning

	// ConflictError represents an error-level conflict
	ConflictError

	// ConflictFatal represents a fatal conflict that prevents installation
	ConflictFatal
)

// String returns a string representation of the conflict severity
func (s ConflictSeverity) String() string {
	switch s {
	case ConflictInfo:
		return "info"
	case ConflictWarning:
		return "warning"
	case ConflictError:
		return "error"
	case ConflictFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// Resolution represents a suggested resolution for a dependency conflict
type Resolution struct {
	// Module is the module this resolution applies to
	Module string `json:"module"`

	// Conflict is the conflict being resolved
	Conflict *Conflict `json:"conflict"`

	// Suggested lists the suggested resolution options
	Suggested []*ResolutionOption `json:"suggested"`

	// Action is the suggested action
	Action ResolutionAction `json:"action"`

	// TargetVersion is the version to use for resolution
	TargetVersion string `json:"target_version,omitempty"`

	// ModulesToUpdate lists modules that need to be updated
	ModulesToUpdate []string `json:"modules_to_update,omitempty"`

	// ModulesToRemove lists modules that need to be removed
	ModulesToRemove []string `json:"modules_to_remove,omitempty"`

	// Description explains the resolution
	Description string `json:"description"`

	// Automatic indicates if this resolution can be applied automatically
	Automatic bool `json:"automatic"`
}

// ResolutionAction represents the type of action to resolve a conflict
type ResolutionAction int

const (
	// ActionUpgrade suggests upgrading to a newer version
	ActionUpgrade ResolutionAction = iota

	// ActionDowngrade suggests downgrading to an older version
	ActionDowngrade

	// ActionRemove suggests removing conflicting modules
	ActionRemove

	// ActionIgnore suggests ignoring the conflict
	ActionIgnore

	// ActionManualResolve suggests manual resolution is required
	ActionManualResolve
)

// String returns a string representation of the resolution action
func (a ResolutionAction) String() string {
	switch a {
	case ActionUpgrade:
		return "upgrade"
	case ActionDowngrade:
		return "downgrade"
	case ActionRemove:
		return "remove"
	case ActionIgnore:
		return "ignore"
	case ActionManualResolve:
		return "manual_resolve"
	default:
		return "unknown"
	}
}

// InstallPlan represents a plan for installing modules with dependencies
type InstallPlan struct {
	// Modules lists modules to install in dependency order
	Modules []*InstallPlanModule `json:"modules"`

	// Dependencies maps module names to their dependencies
	Dependencies map[string][]string `json:"dependencies"`

	// Levels represents installation levels for parallel installation
	Levels [][]string `json:"levels"`

	// TotalModules is the total number of modules to install
	TotalModules int `json:"total_modules"`

	// EstimatedDuration is the estimated installation time
	EstimatedDuration time.Duration `json:"estimated_duration,omitempty"`

	// Conflicts lists any unresolved conflicts
	Conflicts []*Conflict `json:"conflicts,omitempty"`

	// CreatedAt indicates when the plan was created
	CreatedAt time.Time `json:"created_at"`

	// Valid indicates if the plan is valid for execution
	Valid bool `json:"valid"`
}

// PlannedInstallation represents a planned module installation
type PlannedInstallation struct {
	// Module is the module name
	Module string `json:"module"`

	// Version is the version to install
	Version string `json:"version"`

	// Order is the installation order (lower numbers first)
	Order int `json:"order"`

	// Dependencies lists modules this depends on
	Dependencies []string `json:"dependencies,omitempty"`

	// Phase is the dependency phase
	Phase string `json:"phase"`

	// Required indicates if the module is required
	Required bool `json:"required"`

	// AlreadyInstalled indicates if the module is already installed
	AlreadyInstalled bool `json:"already_installed,omitempty"`

	// SkipInstallation indicates if installation should be skipped
	SkipInstallation bool `json:"skip_installation,omitempty"`
}

// BundleExportOptions contains options for exporting module bundles
type BundleExportOptions struct {
	// IncludeDev includes development dependencies
	IncludeDev bool

	// IncludeVersions includes exact version constraints
	IncludeVersions bool

	// IncludeCore includes core modules
	IncludeCore bool

	// Format specifies the export format (json, yaml, cpanfile)
	Format string

	// CompressionLevel specifies compression level (0-9)
	CompressionLevel int

	// OutputPath is the target file path
	OutputPath string

	// ModuleFilter allows filtering modules by pattern
	ModuleFilter string
}

// BundleImportOptions contains options for importing module bundles
type BundleImportOptions struct {
	// SkipInstalled skips modules that are already installed
	SkipInstalled bool

	// UpdateExisting updates existing modules to bundle versions
	UpdateExisting bool

	// SkipDependencies skips dependency resolution
	SkipDependencies bool

	// SkipTests skips running tests during installation
	SkipTests bool

	// Force forces installation even if tests fail
	Force bool

	// DryRun shows what would be installed without actually installing
	DryRun bool

	// Parallel enables parallel installation
	Parallel bool

	// Workers specifies the number of parallel workers
	Workers int
}

// ValidationError represents an error during bundle or cpanfile validation
type ValidationError struct {
	// Field is the field that failed validation
	Field string `json:"field"`

	// Value is the invalid value
	Value string `json:"value"`

	// Message describes the validation error
	Message string `json:"message"`

	// Severity indicates the error severity
	Severity ValidationSeverity `json:"severity"`
}

// ValidationSeverity represents the severity of a validation error
type ValidationSeverity int

const (
	// ValidationInfo represents an informational validation message
	ValidationInfo ValidationSeverity = iota

	// ValidationWarning represents a validation warning
	ValidationWarning

	// ValidationErrorSeverity represents a validation error
	ValidationErrorSeverity

	// ValidationFatal represents a fatal validation error
	ValidationFatal
)

// String returns a string representation of the validation severity
func (s ValidationSeverity) String() string {
	switch s {
	case ValidationInfo:
		return "info"
	case ValidationWarning:
		return "warning"
	case ValidationErrorSeverity:
		return "error"
	case ValidationFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// Additional types for dependency resolution

// VersionConstraint represents a version constraint for dependency resolution
type VersionConstraint struct {
	// Module is the module this constraint applies to
	Module string `json:"module"`

	// Constraint is the version constraint (e.g., ">= 2.0", "== 1.0")
	Constraint string `json:"constraint"`

	// Source is the module that requires this constraint
	Source string `json:"source"`
}

// ConflictType represents the type of dependency conflict
type ConflictType int

const (
	// ConflictTypeVersion represents a version conflict
	ConflictTypeVersion ConflictType = iota

	// ConflictTypeCircular represents a circular dependency
	ConflictTypeCircular

	// ConflictTypeMissing represents a missing dependency
	ConflictTypeMissing
)

// ConflictDependency represents a dependency that contributes to a conflict
type ConflictDependency struct {
	// Dependant is the module that requires the dependency
	Dependant string `json:"dependant"`

	// Required is the version required
	Required string `json:"required"`
}

// ResolutionOption represents a potential resolution for a conflict
type ResolutionOption struct {
	// Version is the suggested version
	Version string `json:"version"`

	// Description explains what this option does
	Description string `json:"description"`

	// Impact describes the impact of choosing this option
	Impact ResolutionImpact `json:"impact"`
}

// ResolutionImpact describes the impact of a resolution choice
type ResolutionImpact struct {
	// AffectedModules is the number of modules affected by this resolution
	AffectedModules int `json:"affected_modules"`

	// BreakingChanges indicates if this resolution may introduce breaking changes
	BreakingChanges bool `json:"breaking_changes"`
}

// InstallPlanModule represents a module in an install plan
type InstallPlanModule struct {
	// Name is the module name
	Name string `json:"name"`

	// Version is the version to install
	Version string `json:"version"`

	// Dependencies lists the module's dependencies
	Dependencies []string `json:"dependencies"`
}
