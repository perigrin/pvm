// ABOUTME: Module installer for PVI
// ABOUTME: Provides the main module installation functionality

package modules

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/pm/deps"
	"tamarou.com/pvm/internal/project"
	"tamarou.com/pvm/internal/xdg"
)

// Error codes for module installation operations
const (
	ErrInstallationFailed = "PVI-4301" // General installation failure
	ErrModuleNotResolved  = "PVI-4302" // Module could not be resolved
	ErrModuleMissing      = "PVI-4303" // Module not found in registry
	ErrDependencyFailure  = "PVI-4304" // Dependency resolution failed
)

// InstallProgressStage represents a stage in the module installation process
type InstallProgressStage int

const (
	// Module installation stages
	StageResolving InstallProgressStage = iota
	StageDownloading
	StageExtracting
	StageBuilding
	StageTesting
	StageInstallingModule
	StageCleaningUp
	StageFinished
)

// String returns a string representation of the installation stage
func (s InstallProgressStage) String() string {
	switch s {
	case StageResolving:
		return "Resolving dependencies"
	case StageDownloading:
		return "Downloading module"
	case StageExtracting:
		return "Extracting module"
	case StageBuilding:
		return "Building module"
	case StageTesting:
		return "Testing module"
	case StageInstallingModule:
		return "Installing module"
	case StageCleaningUp:
		return "Cleaning up"
	case StageFinished:
		return "Finished"
	default:
		return "Unknown"
	}
}

// InstallProgressCallback is called to report progress during installation
type InstallProgressCallback func(stage InstallProgressStage, moduleName string, details string, progress float64)

// ModuleInstallOptions contains options for installing a module
type ModuleInstallOptions struct {
	// Module name to install
	ModuleName string

	// Version constraint for the module
	VersionConstraint string

	// Path to the Perl interpreter to use
	PerlPath string

	// Installation directory (optional - if empty, will use XDG data directory)
	InstallDir string

	// Run tests before installation
	RunTests bool

	// Skip tests completely
	NoTest bool

	// Force installation even if tests fail
	Force bool

	// Clean build directory after installation
	Cleanup bool

	// Include verbose output
	Verbose bool

	// Skip prerequisite installation (dependencies)
	SkipDependencies bool

	// Additional build arguments
	BuildArgs []string

	// CPAN provider for metadata
	Provider cpan.Provider

	// Dependency resolver
	DependencyResolver deps.DependencyResolver

	// Progress callback
	ProgressCallback InstallProgressCallback

	// Context for cancellation
	Context context.Context

	// Project context (nil means no project context)
	ProjectContext *project.ProjectContext

	// Force global installation even in project context
	ForceGlobal bool
}

// ModuleInstallResult contains information about the installation
type ModuleInstallResult struct {
	// Module name
	ModuleName string

	// Module version
	Version string

	// Whether the module was successfully installed
	Success bool

	// Warning messages from the installation process
	Warnings []string

	// Error messages from the installation process
	Errors []string

	// The path where the module was installed
	InstallPath string

	// List of dependencies that were resolved and installed
	Dependencies []*ModuleInstallResult

	// Total time taken for installation
	Duration time.Duration
}

// getVersionSpecificInstallDir returns the version-specific installation directory
func getVersionSpecificInstallDir(perlPath string, dirs *xdg.Dirs) (string, error) {
	// Ensure we have a Perl path
	if perlPath == "" {
		var err error
		perlPath, err = perl.GetCurrentPerlPath()
		if err != nil {
			return "", errors.NewSystemError("007",
				"Failed to determine current Perl path", err)
		}
	}

	// Get the version from the Perl executable
	versionStr, err := perl.GetSystemPerlVersion(perlPath)
	if err != nil {
		return "", errors.NewSystemError("005",
			"Failed to determine Perl version", err)
	}

	// Clean up the version string and handle special cases
	version := strings.TrimSpace(versionStr)

	// Handle system Perl detection - if the path contains /usr/bin/perl or similar system locations
	// we'll use "system" as the version identifier to avoid conflicts with managed versions
	if isSystemPerlPath(perlPath) {
		version = "system"
		log.Infof("Detected system Perl, using 'system' as version identifier")
	} else {
		log.Infof("Using Perl version %s for module directory", version)
	}

	// Create version-specific directory following the proposed structure:
	// ~/.local/share/pvm/library/{version}/
	installDir := filepath.Join(dirs.DataDir, "library", version)

	return installDir, nil
}

// isSystemPerlPath determines if the given Perl path is a system Perl installation
func isSystemPerlPath(perlPath string) bool {
	// Common system Perl locations
	systemPaths := []string{
		"/usr/bin/perl",
		"/usr/local/bin/perl",
		"/bin/perl",
		"/opt/perl/bin/perl",
	}

	for _, sysPath := range systemPaths {
		if perlPath == sysPath {
			return true
		}
	}

	// Also check if it's in typical system directories
	return strings.HasPrefix(perlPath, "/usr/") ||
		strings.HasPrefix(perlPath, "/bin/") ||
		strings.HasPrefix(perlPath, "/opt/")
}

// setupIsolationEnvironment creates a local::lib isolation environment for module installation
func setupIsolationEnvironment(options *ModuleInstallOptions) (string, map[string]string, error) {

	// Determine installation directory
	installDir := options.InstallDir
	if installDir == "" {
		// Check if we're in a project context and not forcing global installation
		if options.ProjectContext != nil && options.ProjectContext.IsProject && !options.ForceGlobal {
			// Install to project's local lib directory (respects configuration)
			installDir = options.ProjectContext.LocalLibDir
			log.Infof("Using project installation directory: %s", installDir)
		} else {
			// Use XDG data directory for version-specific installation
			dirs, err := xdg.GetDirs()
			if err != nil {
				return "", nil, errors.NewSystemError("001",
					"Failed to determine XDG directories", err)
			}

			// Create version-specific perl modules directory in XDG data directory
			// This follows the new structure: ~/.local/share/pvm/library/{version}/
			installDir, err = getVersionSpecificInstallDir(options.PerlPath, dirs)
			if err != nil {
				return "", nil, errors.NewSystemError("006",
					"Failed to determine version-specific installation directory", err)
			}
			log.Infof("Using version-specific installation directory: %s", installDir)
		}
	}

	// Ensure installation directory exists
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return "", nil, errors.NewSystemError("002",
			"Failed to create installation directory", err)
	}

	// Set up local::lib environment variables
	envVars := make(map[string]string)

	// Create lib and bin directories
	libDir := filepath.Join(installDir, "lib", "perl5")
	binDir := filepath.Join(installDir, "bin")

	if err := os.MkdirAll(libDir, 0755); err != nil {
		return "", nil, errors.NewSystemError("003",
			"Failed to create lib directory", err)
	}
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", nil, errors.NewSystemError("004",
			"Failed to create bin directory", err)
	}

	// Set up local::lib environment variables (based on PVX isolation logic)
	envVars["PERL_LOCAL_LIB_ROOT"] = installDir
	envVars["PERL_MB_OPT"] = fmt.Sprintf("--install_base '%s'", installDir)
	envVars["PERL_MM_OPT"] = fmt.Sprintf("INSTALL_BASE=%s", installDir)

	// Set up PERL5LIB with project context awareness
	// Use os.PathListSeparator for cross-platform path list joining
	sep := string(os.PathListSeparator)
	perl5lib := libDir

	// If we're in a project context, ensure project lib is also in the path
	if options.ProjectContext != nil && options.ProjectContext.IsProject {
		projectLib := filepath.Join(options.ProjectContext.RootDir, "lib", "perl5")
		if projectLib != libDir {
			// Add both the install lib and project lib to PERL5LIB
			perl5lib = libDir + sep + projectLib
		}

		// Also include the project's root lib directory for project modules
		projectRootLib := filepath.Join(options.ProjectContext.RootDir, "lib")
		if projectRootLib != libDir && projectRootLib != projectLib {
			perl5lib = perl5lib + sep + projectRootLib
		}
	}

	// Check if there's an existing PERL5LIB and preserve it
	if existingPerl5Lib := os.Getenv("PERL5LIB"); existingPerl5Lib != "" {
		perl5lib = perl5lib + sep + existingPerl5Lib
	}

	envVars["PERL5LIB"] = perl5lib

	// Update PATH to include bin directory
	currentPath := os.Getenv("PATH")
	if currentPath != "" {
		envVars["PATH"] = binDir + sep + currentPath
	} else {
		envVars["PATH"] = binDir
	}

	log.Infof("Set up isolation environment for module installation in: %s", installDir)
	if options.Verbose {
		log.Infof("PERL_LOCAL_LIB_ROOT=%s", envVars["PERL_LOCAL_LIB_ROOT"])
		log.Infof("PERL_MB_OPT=%s", envVars["PERL_MB_OPT"])
		log.Infof("PERL_MM_OPT=%s", envVars["PERL_MM_OPT"])
		log.Infof("PERL5LIB=%s", envVars["PERL5LIB"])
	}

	return installDir, envVars, nil
}

// InstallModule installs a Perl module and its dependencies
func InstallModule(options *ModuleInstallOptions) (*ModuleInstallResult, error) {
	startTime := time.Now()

	// Use default options if nil
	if options == nil {
		return &ModuleInstallResult{
				ModuleName:   "",
				Success:      false,
				Warnings:     []string{"No installation options provided"},
				Errors:       []string{"No installation options provided"},
				Dependencies: []*ModuleInstallResult{},
			}, errors.NewSystemError(
				ErrInstallationFailed,
				"No installation options provided",
				nil)
	}

	// Initialize result
	result := &ModuleInstallResult{
		ModuleName:   options.ModuleName,
		Success:      false,
		Warnings:     []string{},
		Errors:       []string{},
		Dependencies: []*ModuleInstallResult{},
	}

	// Ensure module name is provided
	if options.ModuleName == "" {
		return result, errors.NewSystemError(
			ErrModuleMissing,
			"No module name specified",
			nil)
	}

	// Set default context if not specified
	if options.Context == nil {
		options.Context = context.Background()
	}

	// Auto-detect project context if not provided
	if options.ProjectContext == nil {
		if projectCtx, err := project.GetCurrentProject(); err == nil {
			options.ProjectContext = projectCtx
			if projectCtx.IsProject {
				log.Infof("Detected project context: %s (detected via %s)", projectCtx.RootDir, projectCtx.DetectionInfo)
			}
		} else {
			log.Debugf("Failed to detect project context: %v", err)
		}
	}

	// Ensure we have a provider
	if options.Provider == nil {
		return result, errors.NewSystemError(
			ErrModuleMissing,
			"No CPAN provider specified",
			nil)
	}

	// Ensure we have a perl path
	if options.PerlPath == "" {
		perlPath, err := perl.GetCurrentPerlPath()
		if err != nil {
			return result, errors.NewSystemError(
				ErrInstallationFailed,
				"Failed to determine Perl path",
				err)
		}
		options.PerlPath = perlPath
	}

	// Set up isolation environment for local::lib installation
	actualInstallDir, isolationEnv, err := setupIsolationEnvironment(options)
	if err != nil {
		return result, errors.NewSystemError(
			ErrInstallationFailed,
			"Failed to set up isolation environment",
			err)
	}

	// Update options to use the actual install directory
	options.InstallDir = actualInstallDir

	// Get XDG directories for cache/build paths
	dirs, err := xdg.GetDirs()
	if err != nil {
		return result, errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Ensure directories exist
	err = dirs.EnsureDirs()
	if err != nil {
		return result, errors.NewSystemError("002",
			"Failed to create required directories", err)
	}

	// Create module build directories
	modulesBuildDir := filepath.Join(dirs.BuildDir, "modules")
	if err := os.MkdirAll(modulesBuildDir, 0755); err != nil {
		return result, errors.NewSystemError("003",
			"Failed to create modules build directory", err)
	}

	// Create timestamp-based build directory for this module
	// Sanitize module name for filesystem use: Perl module names contain "::"
	// which is illegal in Windows paths
	sanitizedName := strings.ReplaceAll(options.ModuleName, "::", "-")
	timestamp := time.Now().Format("20060102-150405")
	buildDir := filepath.Join(modulesBuildDir, fmt.Sprintf("%s-%s", sanitizedName, timestamp))
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return result, errors.NewSystemError("004",
			"Failed to create module build directory", err)
	}

	// Update progress
	updateProgress := func(stage InstallProgressStage, details string, progress float64) {
		if options.ProgressCallback != nil {
			options.ProgressCallback(stage, options.ModuleName, details, progress)
		}
	}

	// Start resolving dependencies if not skipped
	if !options.SkipDependencies && options.DependencyResolver != nil {
		updateProgress(StageResolving, "Resolving dependencies", 0.0)

		// Create dependency resolution options
		resolutionOptions := &deps.DependencyResolutionOptions{
			Provider:     options.Provider,
			IncludeCore:  false,
			IncludeTest:  options.RunTests,
			IncludeBuild: true,
			IncludeDev:   false,
			MaxDepth:     0, // No limit
			Verbose:      options.Verbose,
		}

		// Resolve dependencies
		log.Infof("Resolving dependencies for module %s", options.ModuleName)
		depResult, err := options.DependencyResolver.ResolveDependencies(
			options.Context, options.ModuleName, resolutionOptions)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to resolve dependencies: %v", err))
			return result, errors.NewSystemError(
				ErrDependencyFailure,
				fmt.Sprintf("Failed to resolve dependencies for %s", options.ModuleName),
				err)
		}

		// Get flattened dependencies
		flatDeps := options.DependencyResolver.GetFlattenedDependencies(depResult)
		log.Infof("Resolved %d dependencies for %s", len(flatDeps)-1, options.ModuleName) // -1 to exclude root module

		// Process dependencies in correct order (leaves first)
		// Skip the root module (the one we're installing directly)
		for _, dep := range flatDeps {
			if dep.IsRoot || dep.IsCore {
				continue
			}

			log.Infof("Installing dependency: %s", dep.Name)
			updateProgress(StageResolving, fmt.Sprintf("Installing dependency: %s", dep.Name), 0.5)

			// Create options for installing the dependency
			depOptions := &ModuleInstallOptions{
				ModuleName:         dep.Name,
				VersionConstraint:  dep.VersionConstraint,
				PerlPath:           options.PerlPath,
				InstallDir:         options.InstallDir,
				RunTests:           options.RunTests,
				NoTest:             options.NoTest,
				Force:              options.Force,
				Cleanup:            options.Cleanup,
				Verbose:            options.Verbose,
				SkipDependencies:   true, // Skip recursive dependency resolution
				Provider:           options.Provider,
				DependencyResolver: nil, // Avoid further dependency resolution
				Context:            options.Context,
				ProjectContext:     options.ProjectContext, // Pass along project context
				ForceGlobal:        options.ForceGlobal,    // Pass along global preference
			}

			// Install the dependency
			depInstallResult, err := InstallModule(depOptions)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to install dependency %s: %v", dep.Name, err))

				// Don't fail the main installation for optional dependencies
				if dep.Type == "recommends" || dep.Type == "suggests" {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Optional dependency %s failed to install: %v", dep.Name, err))
					continue
				}

				// Only fail for required dependencies
				if !options.Force {
					return result, errors.NewSystemError(
						ErrDependencyFailure,
						fmt.Sprintf("Failed to install dependency %s for %s", dep.Name, options.ModuleName),
						err)
				}

				// Force continue despite dependency failure
				result.Warnings = append(result.Warnings, fmt.Sprintf("Dependency %s failed to install, continuing anyway (--force)", dep.Name))
			}

			// Add to dependencies list
			result.Dependencies = append(result.Dependencies, depInstallResult)
		}
	}

	// Start downloading the module
	updateProgress(StageDownloading, "Downloading module", 0.0)

	// Create download options
	downloadOptions := &DownloadOptions{
		ModuleName:        options.ModuleName,
		VersionConstraint: options.VersionConstraint,
		Provider:          options.Provider,
		Context:           options.Context,
		SkipCache:         false,
		ProgressCallback: func(total, transferred int64, done bool) {
			progress := float64(0)
			if total > 0 {
				progress = float64(transferred) / float64(total)
			}
			updateProgress(StageDownloading, fmt.Sprintf("Downloading %s (%d%%)", options.ModuleName, int(progress*100)), progress)
		},
	}

	// Download the module
	log.Infof("Downloading module: %s", options.ModuleName)
	downloadResult, err := DownloadModule(downloadOptions)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to download module: %v", err))
		return result, errors.NewSystemError(
			ErrInstallationFailed,
			fmt.Sprintf("Failed to download module %s", options.ModuleName),
			err)
	}

	// Update result with version from download
	result.Version = downloadResult.Version

	// Extract the module
	updateProgress(StageExtracting, "Extracting module archive", 0.0)

	log.Infof("Extracting module: %s", options.ModuleName)
	extractResult, err := ExtractModuleArchive(downloadResult.Path, buildDir, options.Context)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to extract module: %v", err))
		return result, errors.NewSystemError(
			ErrInstallationFailed,
			fmt.Sprintf("Failed to extract module %s", options.ModuleName),
			err)
	}

	// Update progress
	updateProgress(StageExtracting, "Module extracted", 1.0)

	// Build and install the module
	updateProgress(StageBuilding, "Building module", 0.0)

	// Create build options
	buildOptions := &ModuleBuildOptions{
		ModuleDir:    extractResult.ExtractedDir,
		ModuleName:   options.ModuleName,
		Distribution: extractResult.Distribution,
		PerlPath:     options.PerlPath,
		InstallDir:   options.InstallDir,
		BuildDir:     buildDir,
		RunTests:     options.RunTests && !options.NoTest,
		NoTest:       options.NoTest,
		Force:        options.Force,
		Cleanup:      options.Cleanup,
		Verbose:      options.Verbose,
		BuildArgs:    options.BuildArgs,
		SkipPrereqs:  options.SkipDependencies,
		Environment:  isolationEnv, // Use PVX isolation environment for local::lib
		Context:      options.Context,
		ProgressCallback: func(stage BuildProgressStage, details string, progress float64) {
			// Map build stages to install stages
			installStage := StageBuilding
			switch stage {
			case StagePrepare, StageCreateBuildScript:
				installStage = StageBuilding
			case StageBuild:
				installStage = StageBuilding
			case StageTest:
				installStage = StageTesting
			case StageInstall:
				installStage = StageInstallingModule
			case StageCleanup:
				installStage = StageCleaningUp
			case StageDone:
				installStage = StageFinished
			}

			updateProgress(installStage, details, progress)
		},
	}

	// Build and install the module
	log.Infof("Building and installing module: %s", options.ModuleName)
	buildResult, err := BuildAndInstallModule(buildOptions)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to build and install module: %v", err))
		return result, errors.NewSystemError(
			ErrInstallationFailed,
			fmt.Sprintf("Failed to build and install module %s", options.ModuleName),
			err)
	}

	// Update result with build warnings and errors
	result.Warnings = append(result.Warnings, buildResult.Warnings...)
	result.Errors = append(result.Errors, buildResult.Errors...)

	// Set success and install path
	result.Success = buildResult.Installed
	result.InstallPath = options.InstallDir // This is approximate

	// Calculate total duration
	result.Duration = time.Since(startTime)

	// Final progress update
	updateProgress(StageFinished, "Module installation completed", 1.0)

	return result, nil
}
