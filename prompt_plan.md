# PVM UX Redesign: Build Plan

## Overview

This plan transforms PVM from a collection of tools into a unified, context-aware Perl development ecosystem with TypeScript-quality tooling. The approach prioritizes safety, incremental progress, test-driven development, and maintains backward compatibility while adding modern workflow capabilities.

## Target Architecture

- **Unified command hierarchy** with `pvm` as primary interface
- **Context-aware defaults** (project vs global behavior)
- **uv-like workflow** for Perl module management
- **TypeScript-style build system** with type-check + compile
- **CPAN distribution ready** builds
- **Backward compatible** aliases (pvi, pvx, psc)

## Implementation Phases

### Phase 1: Foundation & Project Detection (5 steps) ✅ **COMPLETED**
- ✅ Fix command routing infrastructure
- ✅ Implement project context detection
- ✅ Add basic project initialization
- ✅ Context-aware configuration system
- ✅ Project status and health checks

### Phase 2: Module Management (4 steps)
- Context-aware module installation
- cpanfile management
- Lockfile generation

### Phase 3: Build System (6 steps)
- PSC integration with type checking
- .pmc compilation for development
- CPAN distribution builds
- Watch mode and continuous builds

### Phase 4: Development Workflow (4 steps)
- Development environment
- Test integration
- Status and diagnostics

### Phase 5: Polish & Documentation (3 steps)
- Enhanced help system
- Configuration management
- Documentation and examples

---

## Phase 1: Foundation & Project Detection

### Step 1.1: Fix Command Router Infrastructure ✅ **COMPLETED**

**Goal**: Ensure subcommands work properly (`pvm pvi install` should route correctly)

**Status**: ✅ **COMPLETED** - All routing functionality works correctly. Added comprehensive tests to verify:
- Direct symlink routing (./pvi, ./pvx, ./psc)
- Subcommand routing (./pvm pvi, ./pvm pvx, ./pvm psc)
- Help routing (./pvm help pvi, etc.)
- Version routing (./pvm pvi version, etc.)
- Backward compatibility with existing symlinks

```
Fix the command router in internal/cli/router.go to properly detect and route subcommands. The router should:

1. Detect the component (pvm, pvi, pvx, psc) from binary name OR subcommand
2. Route "pvm pvi install Module::Name" to PVI module installer
3. Route "pvm pvx script.pl" to PVX executor
4. Route "pvm psc check file.pl" to PSC type checker
5. Maintain backward compatibility with symlinks

Implementation requirements:
- Modify DetectComponent() to check for subcommands in os.Args
- Update CreateRootCommand() to properly register subcommands
- Ensure GlobalRegistry.Get() works for both direct and subcommand invocation
- Add comprehensive tests for all routing scenarios

Test cases to implement:
- Direct symlink: ./pvi install Module::Name
- Subcommand: ./pvm pvi install Module::Name
- Help: ./pvm help pvi (should show PVI help)
- Version: ./pvm pvi version (should show PVI version)

Verification:
- All existing tests continue to pass
- New routing tests verify correct component selection
- Manual testing: make pvm && ./build/pvm help pvi works
```

### Step 1.2: Implement Project Context Detection ✅ **COMPLETED**

**Goal**: Auto-detect project directories and set context-aware defaults

**Status**: ✅ **COMPLETED** - Full project detection system implemented in `internal/project/detector.go`:
- ProjectContext struct with all required fields
- DetectProject() function with directory tree walking
- Support for .perl-version, cpanfile, pvm.toml, .git markers
- Caching system for performance
- GetCurrentProject() helper function

```
Create internal/project/detector.go that implements project context detection:

1. Create ProjectContext struct with fields:
   - IsProject bool
   - RootDir string
   - PerlVersion string (from .perl-version)
   - HasCpanfile bool
   - LocalLibDir string
   - ConfigFile string (pvm.toml path)

2. Implement DetectProject(workingDir string) function that:
   - Walks up directory tree looking for project markers
   - Checks for .perl-version, cpanfile, pvm.toml, .git
   - Returns ProjectContext with detected information
   - Caches results for performance

3. Project detection logic (in priority order):
   - .perl-version file (definitive project root)
   - cpanfile present (Perl project with dependencies)
   - pvm.toml config file (PVM project)
   - .git directory (git repository)

4. Integration points:
   - Add GetCurrentProject() function to config package
   - Modify config.LoadEffectiveConfig() to use project context
   - Create project-aware path resolution (local lib vs global)

Test cases:
- Project detection in various directory structures
- Caching behavior and invalidation
- Integration with existing config system
- Performance with deep directory structures

The detector should be used by module installation, execution, and type checking to determine whether to use project-local or global behavior.
```

### Step 1.3: Create Project Initialization Command ✅ **COMPLETED**

**Goal**: Add `pvm init` command to scaffold new projects

**Status**: ✅ **COMPLETED** - Full project initialization system implemented in `internal/pvm/project.go`:
- `pvm project init` command with template support
- Creates .perl-version, cpanfile, pvm.toml, .gitignore
- Smart initialization (current dir vs new project dir)
- Template system with built-in and user-defined templates
- Project name validation and structure creation

```
Implement pvm init command in internal/pvm/init.go:

1. Create InitCommand that:
   - Takes optional project name argument
   - Creates project directory structure if name provided
   - Initializes in current directory if no name given
   - Detects existing Perl files and preserves structure

2. Project scaffolding:
   - Create .perl-version with current/default Perl version
   - Generate cpanfile with basic structure
   - Create pvm.toml with project configuration
   - Set up directory structure (lib/, t/, script/)
   - Generate .gitignore with PVM-specific ignores

3. Smart initialization:
   - Detect existing .pm files and suggest project name
   - Preserve existing directory structure
   - Don't overwrite existing configuration files
   - Validate project name (no spaces, valid Perl module name)

4. Template system:
   - Support multiple project templates (minimal, standard, web)
   - Allow custom templates via configuration
   - Generate appropriate boilerplate based on template

Integration:
- Add init command to PVM command registry
- Integrate with project detector for validation
- Use existing config system for template management

Test thoroughly:
- New project creation in empty directory
- Initialization in existing Perl project
- Error handling for invalid project names
- Template generation and file creation
```

### Step 1.4: Implement Context-Aware Configuration ✅ **COMPLETED**

**Goal**: Make configuration system project-aware with proper precedence

**Status**: ✅ **COMPLETED** - Context-aware configuration system implemented:
- Full precedence order: Project > User > System configurations
- Project config loading in `internal/config/loader.go`
- Integration with project detector via `internal/project/config.go`
- Support for `.pvm/pvm.toml` and `pvm.toml` project configs
- LoadEffectiveConfig() merges all sources with proper precedence

```
Enhance the configuration system to support project context:

1. Modify internal/config/loader.go:
   - LoadProjectConfig(projectRoot string) function
   - Merge project and user configurations with proper precedence
   - Support for pvm.toml project configuration files

2. Configuration precedence (highest to lowest):
   - Command line flags
   - Project pvm.toml
   - User ~/.config/pvm/config.toml
   - System defaults

3. Project configuration schema (pvm.toml):
   ```toml
   [project]
   name = "MyApp"
   version = "1.0.0"
   perl_version = "5.40.0"

   [dependencies]
   cpanfile = "cpanfile"
   local_lib = "lib"

   [build]
   output_dir = "build"
   ```

4. Enhanced LoadEffectiveConfig():
   - Use project detector to find project root
   - Load and merge project configuration
   - Provide project-aware defaults (local lib paths, etc.)
   - Cache merged configuration for performance

Integration testing:
- Configuration precedence works correctly
- Project-specific settings override user defaults
- Command-line flags override everything
- Performance is acceptable with caching
```

### Step 1.5: Add Project Status Command ✅ **COMPLETED**

**Goal**: Implement `pvm status` to show project health and configuration

**Status**: ✅ **COMPLETED** - Project status system implemented in `internal/pvm/project.go`:
- `pvm project status` command with comprehensive health checks
- Shows project detection, Perl version, dependencies, build status
- JSON output option available
- Actionable next steps recommendations
- Integration with project detector

```
Create internal/pvm/status.go with comprehensive project status:

1. StatusCommand implementation:
   - Show project detection results
   - Display effective configuration
   - Check Perl version consistency
   - Validate project structure

2. Status information to display:
   - Project root directory and detection method
   - Perl version (project vs global vs actual)
   - Dependencies status (cpanfile vs installed)
   - Local lib status and module count
   - Build status (if build directory exists)
   - Git status (if in git repository)

3. Health checks:
   - Perl version matches .perl-version
   - Required modules are installed
   - No obvious configuration issues
   - Build artifacts are current

4. Output format:
   - Clean, readable status display
   - Color coding for status (green/yellow/red)
   - Actionable suggestions for issues
   - JSON output option for scripting

Test cases:
- Status in project vs non-project directories
- Various project configurations and states
- Error handling for incomplete projects
- Performance with large projects

This command serves as the foundation for project-aware behavior throughout the system.
```

---

## ✅ Phase 1 Complete: Foundation & Project Detection

**Status**: All 5 steps completed successfully!

Phase 1 provides the foundation for the unified PVM ecosystem:
- **Router Infrastructure**: Unified command routing with backward compatibility
- **Project Detection**: Automatic project context detection with caching
- **Project Initialization**: Smart project scaffolding with templates
- **Configuration System**: Project-aware config with proper precedence
- **Status Monitoring**: Comprehensive project health checks

Foundation is solid and ready for Phase 2 development.

---

## Phase 2: Module Management

### Step 2.1: Context-Aware Module Installation ✅ **COMPLETED**

**Goal**: Make `pvi install` project-aware with automatic local lib management

**Status**: ✅ **COMPLETED** - PVI module installer enhanced with full project-awareness:
- Automatic project context detection
- Project-local installation to `./lib/` directory when in project context
- Global installation to XDG directory when outside projects or with `--force-global`
- Enhanced PERL5LIB setup with project lib paths
- Preservation of existing PERL5LIB environment variables
- Comprehensive test coverage for all installation scenarios

```
Enhance internal/pvi/modules/installer.go for context-aware installation:

1. Modify InstallModule function to:
   - Use project detector to determine installation target
   - Install to project lib/ if in project, global otherwise
   - Automatically set up local::lib environment
   - Handle PERL5LIB and related environment variables

2. Installation logic:
   - In project: install to ./lib/ (create if needed)
   - Outside project: install to user or system location
   - Respect --global flag to force global installation
   - Set up proper @INC paths for module resolution

3. Local lib setup:
   - Create lib/ directory structure (lib/perl5, etc.)
   - Set PERL_LOCAL_LIB_ROOT, PERL_MB_OPT, PERL_MM_OPT
   - Update PERL5LIB to include project lib
   - Handle architecture-specific directories

4. Integration points:
   - Use ProjectContext from detector
   - Respect configuration for local lib preferences
   - Coordinate with existing parallel installation system
   - Maintain compatibility with existing PVI commands

Enhanced testing:
- Installation in project vs non-project directories
- Environment variable setup verification
- Module resolution after installation
- Parallel installation with project context
- Error handling for permission issues

This creates the foundation for the uv-like workflow where module installation "just works" based on project context.
```

### Step 2.2: Implement cpanfile Management ✅ **COMPLETED**

**Goal**: Add `pvm module add` command that updates cpanfile and installs modules

**Status**: ✅ **COMPLETED** - Full cpanfile management functionality implemented:
- Created CpanfileManager in `internal/pvi/cpanfile.go` for reading/writing cpanfile format
- Added `pvm module add` command with support for runtime and development dependencies
- Enhanced `pvm module install` to read from cpanfile when no modules specified
- Automatic rollback of cpanfile changes if installation fails
- Project-aware installation to local lib directory
- Preserves cpanfile formatting and comments when possible

```
Create internal/pvm/module.go with cpanfile-aware module management:

1. ModuleAddCommand implementation:
   - Add dependency to cpanfile
   - Install module to project lib/
   - Support dev dependencies (--dev flag)
   - Handle version constraints

2. cpanfile manipulation:
   - Parse existing cpanfile (use internal/cpan/carton.go)
   - Add new requirements in correct format
   - Preserve existing structure and comments
   - Write back formatted cpanfile

3. Module installation integration:
   - Use enhanced installer from Step 2.1
   - Install to project lib/ automatically
   - Update cpanfile before installation
   - Rollback cpanfile if installation fails

4. Command interface:
   ```bash
   pvm module add DBI                    # Add runtime dependency
   pvm module add Test::More --dev      # Add development dependency
   pvm module add DBI ">=1.643"         # Add with version constraint
   ```

5. cpanfile format:
   ```perl
   requires 'DBI', '>= 1.643';
   requires 'Test::More', '0';

   on 'develop' => sub {
       requires 'Test::More', '0';
   };
   ```

Integration testing:
- cpanfile parsing and writing
- Module installation success/failure handling
- Version constraint handling
- Development vs runtime dependencies
- Error recovery and rollback

This is the key command that enables the uv-like workflow: "pvm module add Module::Name" just works.
```

### Step 2.3: Module Installation from cpanfile ✅ **COMPLETED**

**Goal**: Enhance `pvm module install` to read from cpanfile when no modules specified

**Status**: ✅ **COMPLETED** - Module installation from cpanfile fully implemented:
- Enhanced `pvm module install` to automatically read from cpanfile when no modules specified
- Added `--dev` flag to include development dependencies
- Proper phase filtering (runtime vs development dependencies)
- Project-aware installation to local lib directory
- Clear error messages when cpanfile is missing or empty

```
Enhance module installation to support cpanfile-based installation:

1. Modify ModuleInstallCommand in internal/pvi/command.go:
   - If no modules specified, read from cpanfile
   - Install all dependencies from cpanfile
   - Support development dependencies (--dev flag)
   - Use project-aware installation from Step 2.1

2. cpanfile processing:
   - Parse cpanfile using existing carton parser
   - Extract runtime and development requirements
   - Resolve version constraints
   - Handle feature-based dependencies

3. Installation workflow:
   - Read cpanfile in project root
   - Determine which dependencies to install (runtime vs dev)
   - Use parallel installation when possible
   - Report installation progress and results

4. Command behavior:
   ```bash
   pvm module install                   # Install from cpanfile (runtime only)
   pvm module install --dev            # Install runtime + development deps
   pvm module install Module::Name     # Install specific module (existing behavior)
   ```

5. Error handling:
   - Clear messages when cpanfile missing or invalid
   - Dependency resolution failure handling
   - Partial installation recovery
   - Compatibility with existing installation options

Testing requirements:
- cpanfile parsing with various dependency types
- Installation success/failure scenarios
- Development vs runtime dependency handling
- Integration with parallel installation
- Error cases (missing cpanfile, invalid syntax)

This completes the input side of dependency management - adding and installing from cpanfile.
```

### Step 2.4: Lockfile Generation and Sync ✅ **COMPLETED**

**Goal**: Add `pvm module sync` command to generate/update cpanfile.snapshot

**Status**: ✅ **COMPLETED** - Full lockfile management functionality implemented:
- Created `pvm module sync` command with support for generating and installing from cpanfile.snapshot
- Added CpanfileManager methods for snapshot generation, reading, and writing
- Implemented carton-compatible snapshot format with exact version locking
- Added `--from-snapshot` flag to install exact versions from lockfile
- Supports proper distribution tracking with pathname and version information
- Project-aware operation that requires project context

```
Implement lockfile management for reproducible builds:

1. Create ModuleSyncCommand in internal/pvm/module.go:
   - Generate cpanfile.snapshot from installed modules
   - Lock exact versions of all dependencies
   - Include transitive dependencies
   - Support for different environments (runtime, test, develop)

2. Snapshot generation:
   - Query installed modules in project lib/
   - Determine exact versions and distributions
   - Build dependency tree with transitives
   - Format as cpanfile.snapshot

3. cpanfile.snapshot format:
   ```
   # carton snapshot format v1.0
   DISTRIBUTIONS
     DBI-1.643
       pathname: T/TI/TIMB/DBI-1.643.tar.gz
       provides:
         DBI 1.643
       requirements:
         ExtUtils::MakeMaker 6.48
   ```

4. Integration with installation:
   - Auto-sync after successful installations
   - Validate snapshot against cpanfile
   - Support for snapshot-based installation
   - Handle version conflicts and resolution

5. Command interface:
   ```bash
   pvm module sync                     # Generate/update lockfile
   pvm module install --from-snapshot  # Install from exact lockfile
   ```

Testing:
- Snapshot generation accuracy
- Transitive dependency capture
- Version locking verification
- Integration with module add/install commands
- Performance with large dependency trees

This provides reproducible builds and completes the dependency management workflow.
```

---

## Phase 3: Build System

### Step 3.1: PSC Integration Infrastructure ✅ **COMPLETED**

**Goal**: Create foundation for PSC integration with type checking and compilation

**Status**: ✅ **COMPLETED** - PSC integration infrastructure implemented in `internal/build/psc.go`:
- PSCBuilder struct with TypeCheck(), Compile(), and Watch() methods
- Structured error reporting with TypeError type
- File discovery and filtering for Perl files (.pl, .pm, .t)
- Integration with existing typechecker and compiler packages
- Basic polling-based file watching (can be enhanced with fsnotify later)
- Context-aware operations with cancellation support

```
Create internal/build/psc.go for PSC integration:

1. PSCBuilder struct with methods:
   - TypeCheck(files []string) error
   - Compile(inputDir, outputDir string) error
   - Watch(dirs []string, callback func()) error
   - GetTypeErrors() []TypeError

2. Type checking integration:
   - Execute PSC type checker on project files
   - Parse and return type errors in structured format
   - Support for incremental type checking
   - Handle PSC exit codes and error messages

3. File discovery:
   - Find all .pm and .pl files in project
   - Respect .gitignore and build directories
   - Support for custom include/exclude patterns
   - Watch for file system changes

4. Error handling and reporting:
   - Structured type error representation
   - Source location mapping (file, line, column)
   - Error categorization (error, warning, info)
   - Integration with existing error system

5. Configuration:
   - PSC executable location and flags
   - Type checking strictness levels
   - Custom include paths and options
   - Integration with project configuration

Testing:
- PSC process execution and error handling
- File discovery and filtering
- Type error parsing and formatting
- Integration with project detector
- Performance with large codebases

This creates the foundation for all build system functionality.
```

### Step 3.2: Development Build (.pmc generation) ✅ **COMPLETED**

**Goal**: Implement `pvm build --inline` for .pmc file generation

**Status**: ✅ **COMPLETED** - Development build system implemented in `internal/build/inline.go`:
- InlineBuilder struct with Build(), Clean(), and IsStale() methods
- Type checking integration before .pmc generation
- File discovery for .pm files in target directories
- .pmc file generation with proper directory structure
- Clean functionality to remove generated .pmc files
- Staleness detection for incremental builds

```
Create internal/build/inline.go for development builds:

1. InlineBuilder implementation:
   - Type check all project files first
   - Strip type annotations from .pm files
   - Generate .pmc files alongside originals
   - Preserve code formatting and comments

2. Type annotation stripping:
   - Use PSC --strip functionality
   - Remove type declarations and annotations
   - Keep valid Perl code
   - Maintain line numbers for debugging

3. .pmc file handling:
   - Generate .pmc alongside .pm files
   - Perl automatically prefers .pmc over .pm
   - Handle architecture-specific locations
   - Clean up stale .pmc files

4. Build process:
   ```bash
   pvm build --inline
   # 1. Type check all .pm files
   # 2. For each .pm file: generate corresponding .pmc
   # 3. Report build results
   ```

5. Integration:
   - Use PSCBuilder from Step 3.1
   - Respect project configuration
   - Handle build errors gracefully
   - Support for watch mode

6. File management:
   - Track generated .pmc files
   - Clean command to remove .pmc files
   - .gitignore integration for .pmc files
   - Handle file permissions and timestamps

Testing:
- Type annotation stripping accuracy
- .pmc file generation and loading
- Error handling for type check failures
- Integration with Perl module loading
- Performance with large projects

This enables fast development builds where developers get type checking but runtime uses compiled .pmc files.
```

### Step 3.3: Distribution Build System ✅ **COMPLETED**

**Goal**: Implement `pvm build` for CPAN-ready distribution generation

**Status**: ✅ **COMPLETED** - Full distribution build system implemented in `internal/build/distribution.go`:
- DistributionBuilder with Build(), Clean(), and Validate() methods
- Complete CPAN distribution structure generation (lib/, t/, script/, META files)
- Metadata generation: META.json, META.yml, Makefile.PL, MANIFEST
- File processing with type stripping from .pm files
- Project metadata extraction from modules and cpanfile
- Comprehensive validation and error handling
- Full test coverage with all scenarios

```
Create internal/build/distribution.go for production builds:

1. DistributionBuilder implementation:
   - Type check entire project
   - Strip types and copy to build directory
   - Generate CPAN metadata files
   - Create proper distribution structure

2. Build directory structure:
   ```
   build/
   ├── lib/              # Clean Perl modules (types stripped)
   ├── t/                # Tests
   ├── script/           # Scripts
   ├── META.json         # CPAN metadata
   ├── META.yml          # Legacy metadata
   ├── Makefile.PL       # Installer
   ├── MANIFEST          # File list
   └── cpanfile          # Dependencies
   ```

3. Metadata generation:
   - Generate META.json from project config and cpanfile
   - Create appropriate installer (Makefile.PL or Build.PL)
   - Generate MANIFEST with all distribution files
   - Handle author, license, version information

4. File processing:
   - Copy and strip types from all .pm files
   - Copy tests, scripts, and documentation
   - Exclude build artifacts and local files
   - Handle file permissions and timestamps

5. Validation:
   - Validate distribution structure
   - Check metadata consistency
   - Verify module compilation
   - Test installability

6. Command integration:
   ```bash
   pvm build                    # Full distribution build
   pvm build --clean           # Clean build directory first
   pvm build --no-meta         # Skip metadata generation
   ```

Testing:
- Distribution structure validation
- Metadata generation accuracy
- Type stripping in distribution context
- CPAN compatibility verification
- Build reproducibility

This creates production-ready distributions that can be uploaded to CPAN.
```

### Step 3.4: Watch Mode and Continuous Builds ✅ **COMPLETED**

**Goal**: Add `pvm build --watch` for continuous development builds

**Status**: ✅ **COMPLETED** - Continuous build system implemented in `internal/build/watcher.go`:
- BuildWatcher with file monitoring and event processing
- Support for different build types: typecheck, inline, distribution, full
- Debounced file change detection with configurable patterns
- Build queue with smart event categorization
- Integration with PSC, inline, and distribution builders
- Comprehensive test coverage for all watcher functionality

```
Create internal/build/watcher.go for continuous build system:

1. BuildWatcher implementation:
   - Monitor file system for changes
   - Trigger appropriate builds on changes
   - Debounce rapid changes
   - Handle build errors gracefully

2. File watching:
   - Watch lib/, script/, t/ directories
   - Filter for relevant file extensions (.pm, .pl, .t)
   - Respect .gitignore patterns
   - Handle file renames and deletions

3. Build triggering:
   - Incremental builds when possible
   - Full rebuild for configuration changes
   - Type check only for fast feedback
   - Clean builds for distribution

4. Event handling:
   - File change detection and classification
   - Build queue management with prioritization
   - Progress reporting and status updates
   - Error recovery and continuation

5. Integration modes:
   ```bash
   pvm build --watch            # Continuous distribution builds
   pvm build --inline --watch   # Continuous .pmc generation
   pvm dev                      # Combined watch + test + serve
   ```

6. Performance optimization:
   - Incremental type checking
   - Smart dependency analysis
   - Build caching and invalidation
   - Resource usage monitoring

Testing:
- File system event handling
- Build triggering accuracy
- Performance with large projects
- Error recovery scenarios
- Integration with different build modes

This enables modern development workflows with instant feedback on code changes.
```

### Step 3.5: Build Configuration System ✅ **COMPLETED**

**Goal**: Add comprehensive build configuration to pvm.toml

**Status**: ✅ **COMPLETED** - Comprehensive build configuration system implemented:
- BuildConfig types with full TOML support and validation
- Project and build configuration structures with proper merging
- Helper functions for project-aware configuration access
- Support for all build modes (distribution, inline, both)
- Extensive test coverage for configuration parsing, validation, and helpers
- Foundation ready for Step 3.6 build command integration

```
Enhance project configuration for build system control:

1. Build configuration schema:
   ```toml
   [build]
   mode = "distribution"        # "distribution", "inline", "both"
   output_dir = "build"        # Build output directory
   clean_before_build = true   # Clean output before building

   [build.typecheck]
   strict = false              # Strict type checking
   experimental = false        # Experimental type features
   target_perl = "5.36"       # Target Perl version

   [build.files]
   include = ["lib/**/*.pm"]   # Files to include
   exclude = ["local/**"]      # Files to exclude
   watch_dirs = ["lib", "script", "t"]

   [build.distribution]
   include_tests = true        # Include tests in distribution
   include_scripts = true      # Include scripts
   installer = "ExtUtils::MakeMaker"
   ```

2. Configuration integration:
   - Load build config from pvm.toml
   - Merge with user defaults
   - Override with command-line flags
   - Validate configuration values

3. Build customization:
   - Custom file inclusion/exclusion patterns
   - Build hook system for custom steps
   - Plugin architecture for extensions
   - Environment-specific configurations

4. Validation and defaults:
   - Validate configuration on load
   - Provide sensible defaults
   - Error reporting for invalid configs
   - Migration for config format changes

Testing:
- Configuration loading and merging
- Build behavior customization
- Validation and error handling
- Integration with all build modes
- Performance impact assessment

This provides flexible build system configuration while maintaining simple defaults.
```

### Step 3.6: Build Command Integration ✅ **COMPLETED**

**Goal**: Wire all build functionality into unified `pvm build` command

**Status**: ✅ **COMPLETED** - Unified build command fully implemented in `internal/pvm/build.go`:
- Comprehensive build orchestration with type checking, compilation, and metadata generation
- Support for all build modes: distribution, inline, type-check-only, and watch
- Command-line flag integration with project-aware configuration merging
- Error handling and progress reporting for all build operations
- Integration with PSC, inline, and distribution builders from Steps 3.1-3.5
- Full test coverage for command structure, flag parsing, and option handling

```
Create comprehensive build command in internal/pvm/build.go:

1. BuildCommand implementation:
   - Support all build modes (distribution, inline, watch)
   - Integrate type checking, compilation, metadata generation
   - Provide clear progress reporting and error handling
   - Handle build configuration and flags

2. Command interface:
   ```bash
   pvm build                    # Default: distribution build
   pvm build --inline          # Development build (.pmc files)
   pvm build --watch           # Continuous build
   pvm build --check-only      # Type check without compilation
   pvm build --clean           # Clean build
   ```

3. Build orchestration:
   - Coordinate PSC integration, file processing, metadata generation
   - Handle build dependencies and ordering
   - Provide incremental builds when possible
   - Report build status and timing

4. Error handling:
   - Clear error messages for build failures
   - Recovery suggestions for common issues
   - Build artifact cleanup on failure
   - Integration with existing error system

5. Progress reporting:
   - Build phase indication (type check, compile, metadata)
   - File processing progress
   - Error and warning summaries
   - Build timing and performance metrics

6. Integration testing:
   - End-to-end build workflows
   - Error scenarios and recovery
   - Performance with various project sizes
   - Integration with project detection and configuration

This creates the main user interface for the build system, providing a unified experience for all build operations.
```

---

## ✅ Phase 3 Complete: Build System

**Status**: All 6 steps completed successfully!

Phase 3 provides a comprehensive build system for Perl projects:
- **PSC Integration**: Type checking and compilation infrastructure
- **Inline Builds**: Development .pmc file generation for fast iteration
- **Distribution Builds**: CPAN-ready package creation with metadata
- **Watch Mode**: Continuous builds with file monitoring and debouncing
- **Build Configuration**: Flexible TOML-based configuration system
- **Unified Command**: Single `pvm build` interface for all build operations

The build system is production-ready and provides TypeScript-quality tooling for Perl development.

---

## Phase 4: Development Workflow

### Step 4.1: Development Environment Command ✅ **COMPLETED**

**Goal**: Implement `pvm dev` for integrated development environment

**Status**: ✅ **COMPLETED** - Development environment command fully implemented in `internal/pvm/dev.go`:
- DevEnvironment with service coordination for build watching, testing, and type checking
- BuildWatcherService, TestRunnerService, TypeCheckerService with proper interfaces
- Configurable service intervals and selective enabling/disabling
- Graceful service startup, shutdown, and status monitoring
- Integration with existing build system and project detection
- Comprehensive test coverage for all development environment functionality

```
Create internal/pvm/dev.go for comprehensive development mode:

1. DevCommand implementation:
   - Start multiple watchers (build, test, lint)
   - Coordinate different development tools
   - Provide unified status dashboard
   - Handle graceful shutdown and cleanup

2. Development services:
   - Build watcher (continuous .pmc generation)
   - Test runner (execute tests on changes)
   - Type checker (continuous validation)
   - Optional: LSP server, documentation server

3. Service coordination:
   - Start services in correct order
   - Handle service dependencies
   - Restart failed services
   - Aggregate logs and status

4. User interface:
   - Real-time status display
   - Color-coded service status
   - Error highlighting and navigation
   - Keyboard shortcuts for common actions

5. Configuration:
   ```toml
   [development]
   auto_test = true            # Run tests on changes
   auto_format = false         # Format code on save
   show_coverage = true        # Display test coverage
   services = ["build", "test"] # Services to start
   ```

6. Integration:
   - Use build watcher from Step 3.4
   - Integrate with test runner from Step 4.2
   - Coordinate with project configuration
   - Handle different project types

Testing:
- Service startup and coordination
- Error handling and recovery
- Performance and resource usage
- Integration with various project configurations
- Graceful shutdown scenarios

This provides a modern development experience similar to tools like cargo watch or npm run dev.
```

### Step 4.2: Test Integration ✅ **COMPLETED**

**Goal**: Implement `pvm test` with project-aware test running

**Status**: ✅ **COMPLETED** - Test command implemented in `internal/pvm/test.go`:
- Comprehensive test discovery for .t and test .pl files
- Project-aware test execution with proper environment setup
- Integration with project detector for context-aware behavior
- Support for test patterns, verbose output, and result reporting
- Proper test result parsing and summary statistics
- Full test coverage for all test command functionality

```
Create internal/pvm/test.go for comprehensive test runner:

1. TestCommand implementation:
   - Discover tests in project structure
   - Set up proper test environment (PERL5LIB, etc.)
   - Execute tests with appropriate Perl version
   - Report results in structured format

2. Test discovery:
   - Find .t files in t/ directory
   - Support for subdirectories and patterns
   - Custom test patterns via configuration
   - Integration test identification

3. Environment setup:
   - Use project Perl version
   - Include project lib/ in @INC
   - Set up test-specific environment variables
   - Handle module dependencies

4. Test execution:
   - Parallel test execution when possible
   - Progress reporting during execution
   - Capture and format test output
   - Handle test failures and errors

5. Result reporting:
   - Summary of passed/failed tests
   - Detailed failure information
   - Coverage reporting (if available)
   - Integration with CI/CD formats (TAP, JUnit)

6. Build integration:
   - Ensure build is current before testing
   - Use .pmc files if available
   - Handle build failures gracefully
   - Support for testing against distribution

Testing:
- Test discovery in various project structures
- Environment setup verification
- Test execution and result parsing
- Integration with build system
- Performance with large test suites

This provides a reliable test runner that works with the project environment and build system.
```

### Step 4.3: Project Status and Health Checks ✅ **COMPLETED**

**Goal**: Enhance `pvm status` with comprehensive project health monitoring

**Status**: ✅ **COMPLETED** - Comprehensive project health check system implemented:
- `pvm project doctor` command with extensive health checks
- Auto-fix functionality (--fix flag) for common issues
- JSON output support for programmatic access
- Enhanced project status command with JSON support
- Color-coded health status indicators (healthy, warning, critical)
- Comprehensive health check categories covering all project aspects

```
Enhance internal/pvm/status.go with advanced health checks:

1. Health monitoring:
   - Perl version consistency (project vs installed)
   - Dependency status (missing, outdated, conflicts)
   - Build status and freshness
   - Test results and coverage
   - Configuration validation

2. Status categories:
   - Project structure and configuration
   - Dependencies and modules
   - Build artifacts and currency
   - Development environment setup
   - Integration with external tools

3. Health checks:
   - Check .perl-version matches installed version
   - Verify all cpanfile dependencies are installed
   - Validate build artifacts are current
   - Check for common configuration issues
   - Verify test environment setup

4. Diagnostic reporting:
   - Color-coded status indicators
   - Detailed issue descriptions
   - Actionable resolution suggestions
   - Links to relevant documentation

5. Doctor mode:
   ```bash
   pvm status                  # Quick status overview
   pvm doctor                  # Comprehensive health check
   pvm doctor --fix           # Attempt automatic fixes
   ```

6. Integration:
   - Use project detector and configuration
   - Check build system status
   - Validate test environment
   - Monitor dependency status

Testing:
- Health check accuracy
- Issue detection and reporting
- Auto-fix functionality
- Performance with large projects
- Integration with all project components

This provides visibility into project health and helps developers identify and resolve issues quickly.
```

### Step 4.4: Enhanced Help and Discovery ✅ **COMPLETED**

**Goal**: Implement context-aware help system and command suggestions

**Status**: ✅ **COMPLETED** - Enhanced help system fully implemented in `internal/cli/help.go`:
- Context-aware help with project detection and workflow suggestions
- Multiple help topics: workflows, getting-started, troubleshooting, next steps
- Command suggestion system for typos using similarity matching
- Project-aware help content that adapts to current project state
- Comprehensive test coverage for all help functionality
- Integration with PVM command structure

```
Create internal/cli/help.go for enhanced help system:

1. Context-aware help:
   - Different help content based on project context
   - Workflow-oriented help sections
   - Command suggestions based on current state
   - Integration with project status

2. Help categories:
   - Getting started workflows
   - Project-specific commands
   - Build and deployment guides
   - Troubleshooting and diagnostics
   - Configuration references

3. Interactive help:
   - Command suggestion based on context
   - "Did you mean?" for typos
   - Next steps recommendations
   - Integration with status information

4. Help content:
   ```bash
   pvm help                    # General help with context awareness
   pvm help workflows          # Common development workflows
   pvm help getting-started    # New user onboarding
   pvm help troubleshooting    # Common issues and solutions
   ```

5. Command suggestions:
   - Analyze current project state
   - Suggest next logical commands
   - Provide workflow guidance
   - Integration with error messages

6. Documentation integration:
   - Link to online documentation
   - Provide command examples
   - Show configuration options
   - Include best practices

Testing:
- Context detection accuracy
- Help content relevance
- Command suggestion quality
- Integration with project state
- User experience evaluation

This improves discoverability and reduces the learning curve for new users.
```

---

## ✅ Phase 4 Complete: Development Workflow

**Status**: All 4 steps completed successfully!

Phase 4 provides comprehensive development workflow support:
- **Development Environment**: Integrated development mode with service coordination
- **Test Integration**: Project-aware test running with comprehensive discovery
- **Status Monitoring**: Enhanced project health checks with auto-fix functionality
- **Enhanced Help**: Context-aware help system with workflow guidance and discovery

The development workflow is production-ready and provides modern IDE-like capabilities for Perl development.

---

## Phase 5: Polish & Documentation

### Step 5.1: Configuration Management Enhancement

**Goal**: Provide comprehensive configuration management commands

```
Create internal/pvm/config.go for configuration management:

1. Configuration commands:
   - pvm config get/set for individual settings
   - pvm config list for all configurations
   - pvm config reset for defaults
   - pvm config validate for validation

2. Configuration scopes:
   - Global user configuration
   - Project-specific configuration
   - Environment-specific overrides
   - Command-line flag integration

3. Interactive configuration:
   - Guided setup for new users
   - Configuration wizard for projects
   - Validation with helpful error messages
   - Migration for configuration changes

4. Configuration editing:
   - Edit configuration files directly
   - Validate before saving
   - Backup and restore functionality
   - Template-based configuration

Testing:
- Configuration CRUD operations
- Scope resolution and precedence
- Validation and error handling
- Migration and compatibility
- User experience evaluation

This provides comprehensive configuration management to support all PVM functionality.
```

### Step 5.2: Performance Optimization and Monitoring

**Goal**: Optimize performance and add monitoring capabilities

```
Implement performance optimization across the system:

1. Performance monitoring:
   - Command execution timing
   - Build performance metrics
   - Memory usage tracking
   - File system operation monitoring

2. Optimization areas:
   - Project detection caching
   - Configuration loading optimization
   - Build system incremental updates
   - Test execution parallelization

3. Monitoring integration:
   - Optional telemetry collection
   - Performance regression detection
   - Resource usage reporting
   - Bottleneck identification

4. Caching system:
   - Project context caching
   - Configuration caching
   - Build artifact caching
   - Dependency resolution caching

Testing:
- Performance benchmark suite
- Memory usage verification
- Cache effectiveness measurement
- Resource usage monitoring
- Regression detection

This ensures PVM performs well even with large projects and complex configurations.
```

### Step 5.3: Documentation and Examples

**Goal**: Create comprehensive documentation and example projects

```
Create comprehensive documentation system:

1. Documentation structure:
   - Getting started guide
   - Command reference
   - Configuration reference
   - Workflow examples
   - Best practices guide

2. Example projects:
   - Minimal Perl project
   - Web application project
   - CPAN module project
   - Legacy migration example

3. Integration guides:
   - CI/CD integration examples
   - Editor integration setup
   - Docker integration
   - Deployment workflows

4. Interactive documentation:
   - Built-in tutorials
   - Command examples with execution
   - Configuration validators
   - Troubleshooting guides

Testing:
- Documentation accuracy
- Example project functionality
- Integration guide validation
- User experience testing
- Accessibility verification

This ensures users can effectively adopt and use PVM in their workflows.
```

---

## Implementation Guidelines

### Test-Driven Development
- Write tests before implementation
- Maintain 100% test coverage for new code
- Use table-driven tests for comprehensive coverage
- Include integration tests for end-to-end workflows

### Safety Measures
- Implement each step incrementally
- Maintain backward compatibility throughout
- Use feature flags for experimental functionality
- Provide clear rollback procedures

### Quality Assurance
- Run full test suite after each step
- Verify no regression in existing functionality
- Test performance impact of changes
- Validate user experience improvements

### Integration Testing
- Test command routing and subcommand functionality
- Verify project detection in various scenarios
- Validate build system with different project types
- Test configuration management across scopes

This plan provides a comprehensive roadmap for transforming PVM into a modern, unified Perl development ecosystem while maintaining safety, testability, and backward compatibility throughout the implementation process.
