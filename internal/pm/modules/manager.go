// ABOUTME: Module manager for PVI
// ABOUTME: Provides functionality for managing installed modules

package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
)

// Error codes for module management operations
const (
	ErrListModulesFailed  = "PVI-4401" // Failed to list installed modules
	ErrUpdateModuleFailed = "PVI-4402" // Failed to update module
	ErrRemoveModuleFailed = "PVI-4403" // Failed to remove module
	ErrBundleExportFailed = "PVI-4404" // Failed to export module bundle
	ErrBundleImportFailed = "PVI-4405" // Failed to import module bundle
)

// InstalledModule represents an installed Perl module
type InstalledModule struct {
	Name             string    `json:"name"`
	Version          string    `json:"version"`
	Path             string    `json:"path"`
	InstallationTime time.Time `json:"installation_time,omitempty"`
	Description      string    `json:"description,omitempty"`
	PerlVersion      string    `json:"perl_version,omitempty"`
	CoreModule       bool      `json:"core_module,omitempty"`
}

// ModuleListOptions contains options for listing installed modules
type ModuleListOptions struct {
	// Path to the Perl interpreter to use
	PerlPath string

	// Pattern to filter modules by name
	Pattern string

	// Include core modules
	IncludeCore bool

	// Show only the latest version of each module
	LatestOnly bool

	// Context for cancellation
	Context context.Context
}

// ListInstalledModules lists all installed Perl modules
func ListInstalledModules(options *ModuleListOptions) ([]*InstalledModule, error) {
	if options == nil {
		return nil, errors.NewSystemError(
			ErrListModulesFailed,
			"No list options provided",
			nil)
	}

	// Set default context if not specified
	if options.Context == nil {
		options.Context = context.Background()
	}

	// Create a Perl script to list all installed modules with their versions and paths
	script := `
use strict;
use warnings;
use ExtUtils::Installed;
use ExtUtils::MakeMaker;
use JSON;

my $installer = ExtUtils::Installed->new();
my @modules = $installer->modules();

my $results = [];

foreach my $module (sort @modules) {
    next if $module eq 'Perl';  # Skip the Perl "module"

    # Skip modules that don't match the pattern if provided
    if ($ARGV[0] && $module !~ /$ARGV[0]/i) {
        next;
    }

    my $version = eval { $installer->version($module) } || "unknown";
    my $is_core = MM->is_core_only($module) ? 1 : 0;

    # Skip core modules if requested
    if (!$ARGV[1] && $is_core) {
        next;
    }

    # Get module files to determine path
    my @files = eval { $installer->files($module, 'all') };
    my $path = "";

    # Find the module's main .pm file
    foreach my $file (@files) {
        if ($file =~ /${module}(\.pm)?$/) {
            $path = $file;
            last;
        }
    }

    # If we didn't find an exact match, use the first file
    if (!$path && @files) {
        $path = $files[0];
    }

    # Get description from module documentation if possible
    my $description = "";
    if ($path && -f $path) {
        my $pod_section = '';
        eval {
            open my $fh, '<', $path or die "Cannot open $path: $!";
            while (my $line = <$fh>) {
                if ($line =~ /^=head1\s+NAME\s*$/) {
                    $pod_section = 'NAME';
                    next;
                }
                if ($pod_section eq 'NAME' && $line =~ /^\S/ && $line !~ /^=/) {
                    chomp $line;
                    $line =~ s/^$module\s+-\s+//;  # Remove "Module - " prefix
                    $description = $line;
                    last;
                }
                if ($pod_section eq 'NAME' && $line =~ /^=/) {
                    last;  # End of NAME section
                }
            }
            close $fh;
        };
    }

    push @$results, {
        name => $module,
        version => "$version",
        path => $path,
        core_module => $is_core,
        description => $description,
    };
}

print encode_json($results);
`

	// Execute the Perl script
	cmd := exec.CommandContext(options.Context, options.PerlPath, "-e", script, options.Pattern, fmt.Sprintf("%v", options.IncludeCore))
	output, err := cmd.Output()
	if err != nil {
		var stderr string
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		return nil, errors.NewSystemError(
			ErrListModulesFailed,
			fmt.Sprintf("Failed to list installed modules: %s", stderr),
			err)
	}

	// Parse the JSON output
	var modulesData []map[string]interface{}
	if err := json.Unmarshal(output, &modulesData); err != nil {
		return nil, errors.NewSystemError(
			ErrListModulesFailed,
			"Failed to parse module list output",
			err)
	}

	// Convert to InstalledModule objects
	modules := make([]*InstalledModule, 0, len(modulesData))
	for _, data := range modulesData {
		module := &InstalledModule{
			Name:        data["name"].(string),
			Version:     data["version"].(string),
			CoreModule:  data["core_module"].(bool),
			Description: data["description"].(string),
		}

		if path, ok := data["path"].(string); ok {
			module.Path = path
			// Try to get file info for installation time
			if info, err := os.Stat(path); err == nil {
				module.InstallationTime = info.ModTime()
			}
		}

		modules = append(modules, module)
	}

	// Filter to latest version only if requested
	if options.LatestOnly {
		moduleMap := make(map[string]*InstalledModule)
		for _, mod := range modules {
			if existing, ok := moduleMap[mod.Name]; !ok || compareVersions(mod.Version, existing.Version) > 0 {
				moduleMap[mod.Name] = mod
			}
		}

		modules = make([]*InstalledModule, 0, len(moduleMap))
		for _, mod := range moduleMap {
			modules = append(modules, mod)
		}
	}

	// Sort by name
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})

	return modules, nil
}

// RemoveModuleOptions contains options for removing a module
type RemoveModuleOptions struct {
	// Module name to remove
	ModuleName string

	// Path to the Perl interpreter to use
	PerlPath string

	// Force removal even if there are dependencies
	Force bool

	// Include verbose output
	Verbose bool

	// Context for cancellation
	Context context.Context
}

// RemoveModuleResult contains the result of a module removal operation
type RemoveModuleResult struct {
	// Module name that was removed
	ModuleName string

	// Command output (for verbose mode)
	Output string

	// Success indicates if the removal was successful
	Success bool
}

// RemoveModule uninstalls a Perl module
func RemoveModule(options *RemoveModuleOptions) (*RemoveModuleResult, error) {
	if options == nil {
		return nil, errors.NewSystemError(
			ErrRemoveModuleFailed,
			"No removal options provided",
			nil)
	}

	// Set default context if not specified
	if options.Context == nil {
		options.Context = context.Background()
	}

	// First, check if the module is installed
	listOptions := &ModuleListOptions{
		PerlPath:    options.PerlPath,
		Pattern:     "^" + options.ModuleName + "$",
		IncludeCore: true,
		Context:     options.Context,
	}

	modules, err := ListInstalledModules(listOptions)
	if err != nil {
		return nil, errors.NewSystemError(
			ErrRemoveModuleFailed,
			fmt.Sprintf("Failed to check if module %s is installed", options.ModuleName),
			err)
	}

	if len(modules) == 0 {
		return nil, errors.NewSystemError(
			ErrRemoveModuleFailed,
			fmt.Sprintf("Module %s is not installed", options.ModuleName),
			nil)
	}

	// Check if it's a core module
	if modules[0].CoreModule && !options.Force {
		return nil, errors.NewSystemError(
			ErrRemoveModuleFailed,
			fmt.Sprintf("Cannot remove core module %s (use --force to override)", options.ModuleName),
			nil)
	}

	// If there's no specific path found, we can't remove it
	if modules[0].Path == "" {
		return nil, errors.NewSystemError(
			ErrRemoveModuleFailed,
			fmt.Sprintf("Cannot determine installation path for module %s", options.ModuleName),
			nil)
	}

	// Create a Perl script to uninstall the module
	script := `
use strict;
use warnings;
use ExtUtils::Packlist;
use ExtUtils::Installed;
use File::Path qw(remove_tree);
use File::Basename qw(dirname);

my $module = $ARGV[0];
my $verbose = $ARGV[1] eq 'true' ? 1 : 0;

my $installer = ExtUtils::Installed->new();

# Check if module is installed
if (!grep { $_ eq $module } $installer->modules()) {
    print "Module $module is not installed\n";
    exit 1;
}

my @module_files;
my $packlist_file;

eval {
    @module_files = $installer->files($module, 'all');

    # Try to find the .packlist file
    foreach my $file (@module_files) {
        if ($file =~ /\.packlist$/) {
            $packlist_file = $file;
            last;
        }
    }
};

if ($@) {
    print "Error: $@\n";
    exit 2;
}

if ($packlist_file) {
    print "Found packlist at $packlist_file\n" if $verbose;

    # Read the packlist to get a full list of files to remove
    my $packlist = ExtUtils::Packlist->new($packlist_file);
    my @files = sort keys %$packlist;

    # Add the packlist itself
    push @files, $packlist_file;

    # Remove all files
    my $removed = 0;
    foreach my $file (@files) {
        next unless -f $file;  # Skip if file doesn't exist

        print "Removing file: $file\n" if $verbose;
        my $success = unlink $file;
        $removed++ if $success;
    }

    # Try to remove any directories that might be empty now
    my %dirs;
    foreach my $file (@files) {
        my $dir = dirname($file);
        $dirs{$dir} = 1;
    }

    # Sort directories by depth (deepest first) to remove nested dirs
    my @dirs = sort {
        (my $a_count = $a) =~ s/[^\/]//g;
        (my $b_count = $b) =~ s/[^\/]//g;
        length($b_count) <=> length($a_count)
    } keys %dirs;

    foreach my $dir (@dirs) {
        next unless -d $dir;
        if (is_empty_dir($dir)) {
            print "Removing empty directory: $dir\n" if $verbose;
            rmdir $dir;
        }
    }

    print "Removed $removed files for module $module\n";
} else {
    # If no packlist, just remove the known files
    my $removed = 0;
    foreach my $file (@module_files) {
        next unless -f $file;
        print "Removing file: $file\n" if $verbose;
        my $success = unlink $file;
        $removed++ if $success;
    }
    print "Removed $removed files for module $module\n";
}

exit 0;

sub is_empty_dir {
    my $dir = shift;
    opendir(my $dh, $dir) or return 0;
    my @entries = grep { $_ ne '.' && $_ ne '..' } readdir($dh);
    closedir($dh);
    return scalar(@entries) == 0;
}
`

	// Execute the Perl script
	cmd := exec.CommandContext(options.Context, options.PerlPath, "-e", script, options.ModuleName, fmt.Sprintf("%v", options.Verbose))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.NewSystemError(
			ErrRemoveModuleFailed,
			fmt.Sprintf("Failed to remove module %s: %s", options.ModuleName, string(output)),
			err)
	}

	// Return result with output for the caller to handle formatting
	result := &RemoveModuleResult{
		ModuleName: options.ModuleName,
		Output:     string(output),
		Success:    true,
	}

	return result, nil
}

// ModuleBundleInfo represents a bundle of modules for export/import
type ModuleBundleInfo struct {
	// Name of the bundle
	Name string `json:"name"`

	// Description of the bundle
	Description string `json:"description"`

	// Created timestamp
	Created time.Time `json:"created"`

	// Perl version used to create the bundle
	PerlVersion string `json:"perl_version"`

	// Modules included in the bundle
	Modules []*ModuleBundleEntry `json:"modules"`
}

// ModuleBundleEntry represents a module in a bundle
type ModuleBundleEntry struct {
	// Module name
	Name string `json:"name"`

	// Module version constraint (e.g., ">=2.0.0")
	VersionConstraint string `json:"version_constraint,omitempty"`

	// Development dependency
	IsDev bool `json:"is_dev,omitempty"`

	// Optional dependency
	IsOptional bool `json:"is_optional,omitempty"`
}

// ExportBundleOptions contains options for exporting a module bundle
type ExportBundleOptions struct {
	// Path to the output file
	OutputPath string

	// Bundle name
	Name string

	// Bundle description
	Description string

	// Path to the Perl interpreter to use
	PerlPath string

	// Pattern to filter modules by name
	Pattern string

	// Include core modules
	IncludeCore bool

	// Include version constraints
	IncludeVersions bool

	// Context for cancellation
	Context context.Context
}

// ExportModuleBundle exports installed modules to a bundle file
func ExportModuleBundle(options *ExportBundleOptions) error {
	if options == nil {
		return errors.NewSystemError(
			ErrBundleExportFailed,
			"No export options provided",
			nil)
	}

	// Set default context if not specified
	if options.Context == nil {
		options.Context = context.Background()
	}

	// Check if output path is provided
	if options.OutputPath == "" {
		return errors.NewSystemError(
			ErrBundleExportFailed,
			"No output path provided",
			nil)
	}

	// List installed modules
	listOptions := &ModuleListOptions{
		PerlPath:    options.PerlPath,
		Pattern:     options.Pattern,
		IncludeCore: options.IncludeCore,
		LatestOnly:  true,
		Context:     options.Context,
	}

	modules, err := ListInstalledModules(listOptions)
	if err != nil {
		return errors.NewSystemError(
			ErrBundleExportFailed,
			"Failed to list installed modules for bundle export",
			err)
	}

	// Get Perl version
	cmd := exec.CommandContext(options.Context, options.PerlPath, "-e", "print $^V")
	perlVersionBytes, err := cmd.Output()
	perlVersion := "unknown"
	if err == nil {
		perlVersion = strings.TrimPrefix(string(perlVersionBytes), "v")
	}

	// Create bundle info
	bundle := &ModuleBundleInfo{
		Name:        options.Name,
		Description: options.Description,
		Created:     time.Now(),
		PerlVersion: perlVersion,
		Modules:     make([]*ModuleBundleEntry, 0, len(modules)),
	}

	// Add modules to bundle
	for _, mod := range modules {
		entry := &ModuleBundleEntry{
			Name: mod.Name,
		}

		if options.IncludeVersions {
			entry.VersionConstraint = ">=" + mod.Version
		}

		bundle.Modules = append(bundle.Modules, entry)
	}

	// Write to file
	bundleJSON, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return errors.NewSystemError(
			ErrBundleExportFailed,
			"Failed to marshal bundle JSON",
			err)
	}

	if err := os.WriteFile(options.OutputPath, bundleJSON, 0644); err != nil {
		return errors.NewSystemError(
			ErrBundleExportFailed,
			fmt.Sprintf("Failed to write bundle to %s", options.OutputPath),
			err)
	}

	return nil
}

// ImportBundleOptions contains options for importing a module bundle
type ImportBundleOptions struct {
	// Path to the input file
	InputPath string

	// Path to the Perl interpreter to use
	PerlPath string

	// Installation directory (usually site_perl)
	InstallDir string

	// Skip tests during installation
	SkipTests bool

	// Force installation even if tests fail
	Force bool

	// Include verbose output
	Verbose bool

	// Skip prerequisite installation (dependencies)
	SkipDependencies bool

	// CPAN provider for metadata
	Provider interface{} // This should be cpan.Provider but avoiding circular imports

	// Dependency resolver
	DependencyResolver interface{} // This should be deps.DependencyResolver

	// Progress callback
	ProgressCallback func(module string, current, total int, details string)

	// Context for cancellation
	Context context.Context
}

// ImportModuleBundle imports modules from a bundle file
func ImportModuleBundle(options *ImportBundleOptions) error {
	if options == nil {
		return errors.NewSystemError(
			ErrBundleImportFailed,
			"No import options provided",
			nil)
	}

	// Set default context if not specified
	if options.Context == nil {
		options.Context = context.Background()
	}

	// Check if input path is provided
	if options.InputPath == "" {
		return errors.NewSystemError(
			ErrBundleImportFailed,
			"No input path provided",
			nil)
	}

	// Read bundle file
	bundleData, err := os.ReadFile(options.InputPath)
	if err != nil {
		return errors.NewSystemError(
			ErrBundleImportFailed,
			fmt.Sprintf("Failed to read bundle from %s", options.InputPath),
			err)
	}

	// Parse bundle
	var bundle ModuleBundleInfo
	if err := json.Unmarshal(bundleData, &bundle); err != nil {
		return errors.NewSystemError(
			ErrBundleImportFailed,
			"Failed to parse bundle JSON",
			err)
	}

	// Import modules
	total := len(bundle.Modules)
	for i, mod := range bundle.Modules {
		if options.ProgressCallback != nil {
			options.ProgressCallback(mod.Name, i+1, total, fmt.Sprintf("Installing module %d of %d", i+1, total))
		}

		// Skip optional modules for now (would need to add a flag to options)
		if mod.IsOptional {
			continue
		}

		// Here we would call InstallModule, but we're avoiding circular imports
		// The actual implementation in the command.go file will do this
	}

	return nil
}

// OutdatedModuleInfo represents information about an outdated module
type OutdatedModuleInfo struct {
	// Name of the module
	Name string `json:"name"`

	// Currently installed version
	InstalledVersion string `json:"installed_version"`

	// Latest available version
	LatestVersion string `json:"latest_version"`

	// Upgrade available flag
	UpgradeAvailable bool `json:"upgrade_available"`
}

// CheckOutdatedOptions contains options for checking outdated modules
type CheckOutdatedOptions struct {
	// Path to the Perl interpreter to use
	PerlPath string

	// Pattern to filter modules by name
	Pattern string

	// Include core modules
	IncludeCore bool

	// CPAN provider for metadata
	Provider interface{} // This should be cpan.Provider but avoiding circular imports

	// Context for cancellation
	Context context.Context
}

// CheckOutdatedModules checks for modules that have newer versions available
func CheckOutdatedModules(options *CheckOutdatedOptions, checkLatest func(string) (string, error)) ([]*OutdatedModuleInfo, error) {
	if options == nil {
		return nil, errors.NewSystemError(
			ErrListModulesFailed,
			"No check options provided",
			nil)
	}

	// Set default context if not specified
	if options.Context == nil {
		options.Context = context.Background()
	}

	// List installed modules
	listOptions := &ModuleListOptions{
		PerlPath:    options.PerlPath,
		Pattern:     options.Pattern,
		IncludeCore: options.IncludeCore,
		LatestOnly:  true,
		Context:     options.Context,
	}

	modules, err := ListInstalledModules(listOptions)
	if err != nil {
		return nil, errors.NewSystemError(
			ErrListModulesFailed,
			"Failed to list installed modules for outdated check",
			err)
	}

	// Check for updates
	result := make([]*OutdatedModuleInfo, 0, len(modules))
	for _, mod := range modules {
		info := &OutdatedModuleInfo{
			Name:             mod.Name,
			InstalledVersion: mod.Version,
			LatestVersion:    mod.Version,
			UpgradeAvailable: false,
		}

		// Get latest version from provider
		if checkLatest != nil {
			latestVersion, err := checkLatest(mod.Name)
			if err == nil && latestVersion != "" {
				info.LatestVersion = latestVersion
				info.UpgradeAvailable = compareVersions(latestVersion, mod.Version) > 0
			}
		}

		if info.UpgradeAvailable {
			result = append(result, info)
		}
	}

	// Sort by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// MirrorSettings represents the mirror configuration
type MirrorSettings struct {
	// Default mirror
	DefaultMirror string `json:"default_mirror"`

	// Additional mirrors
	AdditionalMirrors []string `json:"additional_mirrors"`
}

// compareVersions compares two version strings.
// Returns:
//
//	-1 if v1 < v2
//	 0 if v1 == v2
//	 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Split versions by dots
	v1Parts := strings.Split(strings.TrimPrefix(v1, "v"), ".")
	v2Parts := strings.Split(strings.TrimPrefix(v2, "v"), ".")

	// Compare each part numerically
	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 string
		if i < len(v1Parts) {
			p1 = v1Parts[i]
		} else {
			p1 = "0"
		}

		if i < len(v2Parts) {
			p2 = v2Parts[i]
		} else {
			p2 = "0"
		}

		// Extract numeric portion of parts
		p1Numeric := strings.Split(p1, "_")[0]
		p2Numeric := strings.Split(p2, "_")[0]

		var n1, n2 int
		_, _ = fmt.Sscanf(p1Numeric, "%d", &n1)
		_, _ = fmt.Sscanf(p2Numeric, "%d", &n2)

		if n1 < n2 {
			return -1
		} else if n1 > n2 {
			return 1
		}
	}

	return 0
}
