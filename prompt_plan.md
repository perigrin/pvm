# Detailed Implementation Blueprint for PVM Ecosystem

Based on my analysis of the specification, I've created a comprehensive blueprint for implementing the PVM Ecosystem. This plan breaks down the project into manageable steps, ensuring test-driven development and incremental progress.

## High-Level Implementation Strategy

The implementation follows these major phases, with each broken down into smaller, testable steps:

1. **Foundation Layer** - Core infrastructure and basic functionality
2. **Environment Management** - Perl installation and execution environment
3. **Module Management** - CPAN integration and module operations
4. **Type System** - Type checking and advanced integrations

## Detailed Implementation Steps

Below are the code generation prompts for each implementation step. Each prompt is designed to build upon previous work while maintaining a test-driven approach.

### Phase 1: Foundation Layer

#### Prompt 1: Project Structure Setup ✅

```
# Implement Basic Project Structure for PVM Ecosystem

## Context
You're starting the implementation of the PVM Ecosystem, a comprehensive suite of tools for Perl development environment management built as a single Go binary with multiple entry points. The system consists of four main components: PVM (Perl Version Manager), PVX (Perl Version eXecutor), PVI (Perl Version Installer), and PSC (Perl Script Compiler).

## Current Task
Establish the basic project structure following Go best practices. This includes setting up the Go module, directory structure, and initial documentation.

## Test Requirements
1. Create a simple test to verify the Go module configuration is correct
2. Add a test to ensure the project can be built

## Implementation Requirements
1. Initialize a Go module with appropriate module name (e.g., github.com/username/pvm)
2. Set up the following directory structure:
   - cmd/ (for command entry points: pvm, pvx, pvi, psc)
   - internal/ (for internal packages)
   - pkg/ (for public packages)
   - docs/ (for documentation)
   - test/ (for test fixtures and helpers)
3. Create a simple README.md with project overview
4. Add LICENSE file with appropriate open-source license
5. Create a basic .gitignore file for Go projects
6. Set up a minimal Go build script or Makefile

## Documentation Requirements
1. Document the project structure in README.md
2. Include basic instructions for building the project
3. Add comments explaining the purpose of key directories

## Considerations
- The structure should support multiple entry points from a single codebase
- The organization should facilitate clean separation of concerns
- Consider cross-platform compatibility (Linux, macOS, Windows)
```

#### Prompt 2: CLI Framework Implementation ✅

```
# Implement CLI Framework for PVM Ecosystem

## Context
Building on the project structure you've established, you now need to implement a CLI framework that will serve as the foundation for all command-line interactions across the four main components (PVM, PVX, PVI, PSC).

## Previous Implementation
- Basic project structure with directories and documentation
- Go module setup with initial build capability

## Current Task
Implement a flexible CLI framework that can handle commands, subcommands, flags, and arguments. This framework should be reusable across all components of the system.

## Test Requirements
1. Test command-line argument parsing for basic commands
2. Test help text generation and formatting
3. Test version flag functionality
4. Test error handling for invalid commands/arguments

## Implementation Requirements
1. Select and integrate an appropriate CLI library (e.g., cobra, urfave/cli, etc.)
2. Create a base command structure that can be extended for specific commands
3. Implement version flag displaying version information
4. Implement help functionality that displays usage information
5. Create a basic command registration mechanism
6. Implement proper error output for command failures

## Integration Points
- The CLI framework should be usable by all four entry points (PVM, PVX, PVI, PSC)
- The command structure should allow for future extension

## Documentation Requirements
1. Add code documentation for public functions and types
2. Update README.md with information about the CLI framework
3. Document how to add new commands to the system

## Considerations
- The CLI framework should be flexible enough to accommodate all planned commands
- Consider user experience and consistent help/error messaging
- Ensure cross-platform compatibility for terminal interactions
```

#### Prompt 3: Command Router Implementation ✅

```
# Implement Command Router for Multiple Entry Points

## Context
The PVM Ecosystem is designed as a single binary with multiple entry points (pvm, pvx, pvi, psc). You need to implement a command router that can identify which entry point was used and direct to the appropriate command set.

## Previous Implementation
- Basic project structure
- CLI framework with command handling capabilities
- Command registration mechanism

## Current Task
Create a command router that detects how the binary was invoked (by name) and routes to the appropriate command set. This will allow the single binary to function as four different tools.

## Test Requirements
1. Test detection of binary name from different invocation methods
2. Test correct routing to PVM commands when invoked as "pvm"
3. Test correct routing to PVX commands when invoked as "pvx"
4. Test correct routing to PVI commands when invoked as "pvi"
5. Test correct routing to PSC commands when invoked as "psc"
6. Test fallback behavior when invoked with an unknown name

## Implementation Requirements
1. Implement detection of binary name from os.Args[0] or equivalent
2. Create a router mechanism that selects the appropriate command set
3. Set up placeholders for the four main command sets
4. Implement a small set of dummy commands for each entry point to validate routing
5. Ensure proper error messages when invalid commands are provided

## Integration Points
- Integrate with the CLI framework from the previous step
- Set up the structure for the four command sets that will be implemented later

## Documentation Requirements
1. Document how the command routing works
2. Update the build documentation to explain how to create the necessary symlinks or copies
3. Add code comments explaining the routing logic

## Considerations
- The router should work on all target platforms (Linux, macOS, Windows)
- Consider how symlinks or file copies work on different platforms
- The design should allow for easy extension with new commands
```

#### Prompt 4: Logging and Error Framework ✅

```
# Implement Logging and Error Framework

## Context
The PVM Ecosystem needs a consistent approach to logging and error handling across all components. The specification outlines detailed requirements for error categories, formatting, and inter-component communication.

## Previous Implementation
- Project structure and CLI framework
- Command router for handling multiple entry points
- Basic command execution flow

## Current Task
Implement a logging and error handling framework that meets the specifications and provides consistent behavior across all components.

## Test Requirements
1. Test logging at different levels (info, warning, error, debug)
2. Test error creation with various categories
3. Test error formatting according to the specified format
4. Test error propagation between components
5. Test error wrapping and context addition

## Implementation Requirements
1. Create an error framework with the specified error categories:
   - Configuration Errors
   - Version Errors
   - Module Errors
   - Execution Errors
   - Type Errors
   - System Errors
   - User Input Errors
2. Implement error codes with component prefixes (PVM-, PVX-, PVI-, PSC-, CFG-, SYS-)
3. Create a structured error format with description, detail, location, and hint
4. Implement a logging system with appropriate levels and output formatting
5. Add error wrapping capabilities for inter-component error communication

## Integration Points
- Integrate with the CLI framework for displaying errors to users
- Set up for use across all components

## Documentation Requirements
1. Document error categories and codes
2. Create guidelines for error message writing
3. Document logging levels and when to use each
4. Add code comments explaining key functions and types

## Considerations
- Logging should be configurable (verbosity levels, output destination)
- Error handling should be consistent across all components
- Consider internationalization potential for error messages
- Performance impact of error creation and logging should be minimal
```

#### Prompt 5: Configuration System - Basic TOML Parsing ✅

```
# Implement Basic TOML Configuration Parsing

## Context
The PVM Ecosystem uses TOML for configuration, with a layered approach following XDG Base Directory Specification. This step implements the basic TOML parsing functionality.

## Previous Implementation
- Project structure, CLI framework, and command routing
- Logging and error handling framework

## Current Task
Implement basic TOML configuration parsing that can read and validate configuration files for the PVM Ecosystem.

## Test Requirements
1. Test parsing valid TOML configuration files
2. Test handling of invalid TOML syntax
3. Test validation of configuration against expected schema
4. Test extraction of configuration values with appropriate types

## Implementation Requirements
1. Select and integrate a TOML parsing library
2. Create configuration structs matching the specification
3. Implement parsing of TOML files into configuration structs
4. Add basic validation of configuration values
5. Create helper functions for accessing configuration values
6. Implement proper error handling for parsing and validation errors

## Integration Points
- Integrate with the error handling framework for reporting configuration errors
- Design for later integration with the XDG directory support

## Documentation Requirements
1. Document the configuration format with examples
2. Add code comments explaining the configuration structs
3. Document error messages related to configuration parsing

## Considerations
- The implementation should be extensible for future configuration needs
- Consider performance for configuration parsing
- Ensure thread safety if configuration might be accessed concurrently
```

#### Prompt 6: XDG Directory Support ✅

```
# Implement XDG Directory Support for Configuration

## Context
The PVM Ecosystem follows the XDG Base Directory Specification for storing configuration and data files. This step implements support for the XDG directory structure and locating configuration files.

## Previous Implementation
- Project structure, CLI framework, and command routing
- Logging and error handling framework
- Basic TOML configuration parsing

## Current Task
Implement XDG Base Directory Specification support to locate and manage configuration and data files across different platforms.

## Test Requirements
1. Test detection of XDG directories on different platforms
2. Test fallback to default locations when XDG variables are not set
3. Test location of configuration files in the correct directories
4. Test creation of directories when needed
5. Test platform-specific behavior (Windows, macOS, Linux)

## Implementation Requirements
1. Implement detection of XDG_CONFIG_HOME, XDG_DATA_HOME, and XDG_CACHE_HOME
2. Create fallback paths for when environment variables are not set
3. Implement platform-specific path handling
4. Add functions to locate configuration files in the correct directories
5. Create helper functions for accessing data and cache directories
6. Implement creation of necessary directories when they don't exist

## Integration Points
- Integrate with the configuration parser to load files from the correct locations
- Design for use by other components that need to access data files

## Documentation Requirements
1. Document the directory structure used by the application
2. Add code comments explaining the XDG specification implementation
3. Document platform-specific behaviors

## Considerations
- Paths should work correctly on Windows, macOS, and Linux
- Consider permission issues when creating directories
- Ensure thread safety for directory operations
```

#### Prompt 7: Configuration Layering and Merging ✅

```
# Implement Configuration Layering and Merging

## Context
The PVM Ecosystem uses a layered configuration approach with project, user, and system-level configuration files. This step implements loading and merging configurations from multiple sources.

## Previous Implementation
- TOML configuration parsing
- XDG directory support for locating configuration files

## Current Task
Implement configuration layering that can load and merge configuration from multiple sources according to the priority specified in the specification.

## Test Requirements
1. Test loading configuration from multiple files
2. Test correct priority order (project > user > system)
3. Test merging of configuration values from different sources
4. Test overriding behavior for specific settings
5. Test handling of missing configuration files

## Implementation Requirements
1. Create a configuration loader that can discover configuration files in all locations
2. Implement merging of configuration values from multiple sources
3. Respect the priority order: project > user > system
4. Handle nested configuration sections during merging
5. Implement override markers or mechanisms if specified
6. Create a unified configuration object representing the merged state

## Integration Points
- Integrate with the XDG directory support for file discovery
- Integrate with the TOML parser for reading individual files

## Documentation Requirements
1. Document the configuration precedence rules
2. Explain how configuration merging works, especially for nested values
3. Add code comments explaining merge logic
4. Update user documentation with configuration information

## Considerations
- The merging logic should be clear and predictable for users
- Consider the performance impact of loading multiple files
- Ensure thread safety if configuration might be reloaded during runtime
```

### Phase 2: Environment Management

#### Prompt 8: System Perl Detection ✅

```
# Implement System Perl Detection

## Context
The PVM Ecosystem needs to detect and use the system's installed Perl when no other version is specified. This step implements the detection of system Perl installations.

## Previous Implementation
- Configuration system with layering and merging
- Directory structure support

## Current Task
Implement detection of system Perl installations, including version information and installation path.

## Test Requirements
1. Test detection of system Perl presence
2. Test extraction of system Perl version
3. Test handling of systems without Perl installed
4. Test detection on different platforms (mock if needed)
5. Test correct path resolution for system Perl

## Implementation Requirements
1. Implement detection of system Perl using PATH lookup
2. Create version extraction from "perl -v" output
3. Implement path resolution for the system Perl binary
4. Add structure for storing detected Perl information
5. Implement error handling for systems without Perl
6. Add platform-specific detection logic if needed

## Integration Points
- Integrate with the configuration system for default version fallback
- Design for use by version resolution algorithm later

## Documentation Requirements
1. Document how system Perl detection works
2. Explain version extraction logic
3. Add code comments for platform-specific behavior

## Considerations
- The detection should work on all supported platforms
- Consider performance implications of running Perl processes
- Handle unusual Perl installations or custom builds
```

#### Prompt 9: Version String Parsing and Constraints ✅

```
# Implement Version String Parsing and Constraints

## Context
The PVM Ecosystem needs to parse and compare Perl version strings and handle version constraints to support version selection and dependency resolution.

## Previous Implementation
- System Perl detection
- Configuration and directory management

## Current Task
Implement parsing and comparison of Perl version strings, along with support for version constraints as used in the specification.

## Test Requirements
1. Test parsing of various Perl version formats (5.32.1, v5.32.1, etc.)
2. Test comparison between versions (>, <, ==, etc.)
3. Test parsing of version constraints (>=5.30.0, <5.38.0, etc.)
4. Test validation of versions against constraints
5. Test handling of version aliases as specified in the configuration

## Implementation Requirements
1. Create a version type that represents a Perl version
2. Implement parsing of version strings into structured data
3. Add comparison operators for versions
4. Implement parsing of version constraints
5. Create validation functions for checking versions against constraints
6. Add support for version aliases (latest, stable, etc.)
7. Implement proper error handling for invalid version strings or constraints

## Integration Points
- Design for use by version resolution algorithm later
- Integrate with configuration for version aliases

## Documentation Requirements
1. Document supported version formats and constraints
2. Explain version comparison rules
3. Document version alias functionality
4. Add code comments explaining parsing logic

## Considerations
- The implementation should handle all Perl version formats correctly
- Consider performance for version comparisons that may happen frequently
- Ensure compatibility with existing tools' version specifications
```

#### Prompt 10: Legacy Tool Integration (plenv/perlbrew) ✅

```
# Implement Legacy Tool Integration (plenv/perlbrew)

## Context
The PVM Ecosystem is designed to integrate seamlessly with existing Perl version managers like plenv and perlbrew. This step implements detection and integration with these tools.

## Previous Implementation
- System Perl detection
- Version string parsing and constraints

## Current Task
Implement detection and integration with plenv and perlbrew installations, including reading their configuration and version information.

## Test Requirements
1. Test detection of plenv installations and versions
2. Test detection of perlbrew installations and versions
3. Test reading of .perl-version files (plenv)
4. Test reading of perlbrew configuration and aliases
5. Test import functionality from both systems

## Implementation Requirements
1. Implement detection of plenv installations in ~/.plenv/versions/
2. Implement detection of perlbrew installations in ~/perl5/perlbrew/perls/
3. Create readers for .perl-version files
4. Implement parsing of perlbrew configuration and aliases
5. Add import commands (pvm import-from plenv|perlbrew)
6. Create a unified registry for tracking detected installations
7. Implement version mapping for consistent naming

## Integration Points
- Integrate with version resolution for using detected installations
- Connect with the command router for import commands

## Documentation Requirements
1. Document how legacy tool integration works
2. Explain the import process and what it does/doesn't do
3. Add command documentation for import commands
4. Provide guidance on migrating from existing tools

## Considerations
- The integration should be non-destructive to existing installations
- Consider handling of custom installation paths for these tools
- Ensure correct version mapping between different naming schemes
```

#### Prompt 11: Version Resolution Algorithm ✅

```
# Implement Version Resolution Algorithm

## Context
The PVM Ecosystem needs to resolve which Perl version to use based on various inputs like explicit specification, project settings, user configuration, and system defaults.

## Previous Implementation
- System Perl detection
- Version string parsing and constraints
- Legacy tool integration (plenv/perlbrew)

## Current Task
Implement the version resolution algorithm that determines which Perl version to use in any given context according to the precedence rules in the specification.

## Test Requirements
1. Test resolution with explicit version specification
2. Test resolution with project-local .perl-version file
3. Test resolution with project-local .pvm/pvm.toml
4. Test resolution with user-level configuration
5. Test resolution with environment variables (PLENV_VERSION, PERLBREW_PERL)
6. Test fallback to system Perl
7. Test resolution with version aliases

## Implementation Requirements
1. Create a version resolver that implements the precedence rules
2. Implement checking for project-local version files
3. Add support for reading version from configuration
4. Implement environment variable checks
5. Create fallback chain to system Perl
6. Add alias resolution from configuration
7. Implement proper error handling for missing or invalid versions

## Integration Points
- Integrate with all previous version detection mechanisms
- Design for use by execution components later

## Documentation Requirements
1. Document the version resolution algorithm and precedence
2. Explain how different methods of specifying versions interact
3. Add examples of common scenarios
4. Include code comments explaining resolution logic

## Considerations
- The resolution should be fast as it may be called frequently
- Consider caching resolved versions where appropriate
- Ensure clear error messages when resolution fails
```

#### Prompt 12: Perl Installation - Source Downloading ✅

```
# Implement Perl Source Downloading

## Context
The PVM Ecosystem needs to download Perl source code for installing specific versions. This step implements the mechanisms for downloading source archives.

## Previous Implementation
- Version resolution algorithm
- Configuration and directory management

## Current Task
Implement functionality to download Perl source code archives from appropriate mirrors for installation.

## Test Requirements
1. Test URL construction for different Perl versions
2. Test downloading from configurable mirrors
3. Test handling of download failures
4. Test caching of downloaded archives
5. Test validation of downloaded files (checksums if available)

## Implementation Requirements
1. Create URL generation for Perl source archives
2. Implement configurable mirrors from configuration
3. Add download functionality with progress reporting
4. Implement caching in XDG_CACHE_HOME/pvm/sources/
5. Add checksum validation if available
6. Create retry mechanism for failed downloads
7. Implement proper error handling and reporting

## Integration Points
- Integrate with the configuration system for mirror settings
- Design for use by installation process later

## Documentation Requirements
1. Document the download process and caching
2. Explain mirror configuration options
3. Add code comments explaining URL generation
4. Document error handling for network issues

## Considerations
- Handle various network conditions gracefully
- Consider proxy settings and authentication if needed
- Ensure security of downloads (TLS, checksums)
- Consider bandwidth usage and provide progress information
```

#### Prompt 13: Perl Installation - Build Process ✅

```
# Implement Perl Build Process

## Context
After downloading Perl source code, the PVM Ecosystem needs to compile and install it. This step implements the build process for Perl installations.

## Previous Implementation
- Source downloading mechanism
- Configuration and directory management

## Current Task
Implement the build process for compiling Perl from source and installing it to the appropriate location.

## Test Requirements
1. Test extraction of source archives
2. Test configuration of build with various options
3. Test compilation process with error handling
4. Test installation to target directory
5. Test registration of installed version
6. Test cleanup after installation

## Implementation Requirements
1. Implement extraction of source archives
2. Create build directory management in XDG_CACHE_HOME/pvm/build/
3. Add Configure script execution with appropriate options
4. Implement make execution with parallelism
5. Add make test execution (optional based on configuration)
6. Implement make install to target directory
7. Create cleanup procedures for build directories
8. Implement proper error handling and reporting
9. Add platform-specific build adjustments

## Integration Points
- Integrate with source downloading
- Connect with version registration

## Documentation Requirements
1. Document the build process and options
2. Explain configuration options for builds
3. Add platform-specific build information
4. Document troubleshooting for common build issues

## Considerations
- The build process is time-consuming and should provide good progress information
- Consider platform-specific build requirements
- Handle build failures gracefully with useful error messages
- Ensure parallel builds work correctly
```

#### Prompt 14: Perl Installation - Registration and Management

```
# Implement Perl Installation Registration and Management

## Context
After building and installing Perl, the PVM Ecosystem needs to register and manage installed versions. This step implements the registration and management of Perl installations.

## Previous Implementation
- Source downloading and build process
- Configuration and directory management

## Current Task
Implement registration of installed Perl versions and commands for managing them.

## Test Requirements
1. Test registration of newly installed versions
2. Test listing of installed versions
3. Test uninstallation process
4. Test commands for installation management (install, versions, uninstall)
5. Test handling of installation errors

## Implementation Requirements
1. Create a registry for tracking installed versions
2. Implement version registration after successful installation
3. Add commands for listing installed versions (pvm versions)
4. Implement uninstallation functionality
5. Create commands for installation management
6. Add version validation for installation requests
7. Implement proper error handling and reporting

## Integration Points
- Integrate with the build process
- Connect with command router for installation commands

## Documentation Requirements
1. Document the installation commands and options
2. Explain version registry functionality
3. Add command documentation for version management
4. Include examples of common usage patterns

## Considerations
- Consider storage format for version registry
- Ensure thread safety for registration operations
- Handle partial or failed installations gracefully
```

#### Prompt 15: Environment Setup - Shim Generation

```
# Implement Shim Generation for Perl Executables

## Context
The PVM Ecosystem uses shims to intercept Perl-related commands and direct them to the appropriate version. This step implements the generation and management of shim executables.

## Previous Implementation
- Perl installation registration and management
- Version resolution algorithm

## Current Task
Implement generation and management of shim executables for Perl and related tools.

## Test Requirements
1. Test shim template compilation
2. Test shim generation for perl and core tools
3. Test shim execution with version resolution
4. Test rehash functionality for updating shims
5. Test platform-specific shim behavior

## Implementation Requirements
1. Create a shim template that performs version resolution
2. Implement detection of commands to shim (perl, cpan, etc.)
3. Add generation of shim executables in XDG_DATA_HOME/pvm/shims/
4. Implement PATH management for shims
5. Create rehash command for rebuilding shims
6. Add platform-specific shim generation (shell scripts vs. executables)
7. Implement proper error handling for shim operations

## Integration Points
- Integrate with version resolution
- Connect with command router for rehash command

## Documentation Requirements
1. Document how shims work and their purpose
2. Explain the rehash command functionality
3. Add information about PATH management
4. Include platform-specific shim information

## Considerations
- Shims should have minimal performance overhead
- Consider security implications of executable generation
- Ensure platform compatibility for execution
```

#### Prompt 16: Environment Setup - Shell Integration

```
# Implement Shell Integration for PVM

## Context
The PVM Ecosystem needs to integrate with different shells for version switching and environment variables. This step implements shell integration for various shells.

## Previous Implementation
- Shim generation and management
- Version resolution algorithm

## Current Task
Implement shell integration for different shells (bash, zsh, fish, PowerShell) to support version switching and environment setup.

## Test Requirements
1. Test generation of shell initialization scripts
2. Test version switching for different shells
3. Test environment variable setup
4. Test global, local, and shell-specific version settings
5. Test platform-specific shell behavior

## Implementation Requirements
1. Create shell initialization scripts for different shells
2. Implement commands for version switching (pvm use, pvm global, pvm local)
3. Add environment variable management
4. Create .perl-version file generation for local version
5. Implement shell detection and appropriate script selection
6. Add platform-specific shell handling
7. Create eval-based integration for shells

## Integration Points
- Integrate with version resolution
- Connect with command router for shell commands

## Documentation Requirements
1. Document shell integration process and supported shells
2. Explain version switching commands
3. Add shell-specific setup instructions
4. Include examples for common usage patterns

## Considerations
- Consider portability across different shell implementations
- Ensure minimal impact on shell startup time
- Handle edge cases like non-interactive shells
```

#### Prompt 17: Basic Execution Capabilities - PVX Foundation

```
# Implement Basic Execution Capabilities for PVX

## Context
The PVM Ecosystem includes PVX (Perl Version eXecutor) for executing Perl code in isolated environments. This step implements the foundation for PVX functionality.

## Previous Implementation
- Shell integration and environment setup
- Version resolution algorithm

## Current Task
Implement the basic execution capabilities for PVX to run Perl scripts with specific versions.

## Test Requirements
1. Test basic script execution with specified version
2. Test command-line argument passing
3. Test environment variable handling
4. Test exit code propagation
5. Test error handling for execution failures

## Implementation Requirements
1. Create basic command structure for PVX
2. Implement script execution with specific Perl version
3. Add command-line argument passing
4. Create environment variable handling
5. Implement exit code propagation
6. Add error capturing and reporting
7. Create integration with version resolution

## Integration Points
- Integrate with version resolution algorithm
- Connect with command router for PVX commands

## Documentation Requirements
1. Document basic PVX usage
2. Explain command-line options
3. Add examples of simple execution patterns
4. Include error handling documentation

## Considerations
- Consider performance overhead of execution
- Ensure proper isolation between different executions
- Handle path resolution correctly across platforms
```

#### Prompt 18: Environment Isolation for PVX

```
# Implement Environment Isolation for PVX

## Context
PVX needs to execute Perl code in isolated environments to prevent interference between different scripts or modules. This step implements environment isolation for PVX execution.

## Previous Implementation
- Basic execution capabilities for PVX
- Version resolution and management

## Current Task
Implement environment isolation for PVX to provide controlled execution environments with configurable isolation levels.

## Test Requirements
1. Test different isolation levels (none, low, medium, high)
2. Test environment variable isolation
3. Test filesystem isolation where applicable
4. Test module path isolation
5. Test cleanup after execution

## Implementation Requirements
1. Create isolated environment structures
2. Implement different isolation levels
3. Add environment variable isolation
4. Create module path isolation
5. Implement filesystem isolation where needed
6. Add cleanup procedures for temporary environments
7. Create configuration options for isolation settings

## Integration Points
- Integrate with basic execution capabilities
- Connect with configuration system for isolation settings

## Documentation Requirements
1. Document isolation levels and their implications
2. Explain environment isolation mechanisms
3. Add configuration options for isolation
4. Include examples of use cases for different levels

## Considerations
- Balance isolation with performance
- Consider security implications of isolation
- Ensure consistent behavior across platforms
```

### Phase 3: Module Management

#### Prompt 19: CPAN Integration - Metadata Retrieval

```
# Implement CPAN Metadata Retrieval

## Context
The PVM Ecosystem includes PVI (Perl Version Installer) for managing CPAN modules. This step implements the foundation for CPAN integration with metadata retrieval.

## Previous Implementation
- Environment isolation for PVX
- Version management and resolution

## Current Task
Implement CPAN metadata retrieval for discovering and analyzing available modules.

## Test Requirements
1. Test CPAN mirror configuration and access
2. Test module metadata retrieval
3. Test search functionality
4. Test handling of metadata caching
5. Test error handling for network issues

## Implementation Requirements
1. Create CPAN mirror configuration and management
2. Implement metadata retrieval from CPAN
3. Add search capabilities for modules
4. Create metadata caching system
5. Implement module information display
6. Add error handling for network and data issues

## Integration Points
- Integrate with configuration system for mirror settings
- Design for use by module installation later

## Documentation Requirements
1. Document CPAN integration features
2. Explain mirror configuration options
3. Add search command documentation
4. Include information about metadata caching

## Considerations
- Handle network connectivity issues gracefully
- Consider caching strategies for performance
- Ensure compatibility with official CPAN API
```

#### Prompt 20: Dependency Resolution for Modules

```
# Implement Dependency Resolution for Modules

## Context
PVI needs to resolve module dependencies during installation. This step implements dependency resolution for CPAN modules.

## Previous Implementation
- CPAN metadata retrieval
- Module search capabilities

## Current Task
Implement dependency resolution for CPAN modules following the approach specified in the documentation (similar to Menlo and App::cpanminus).

## Test Requirements
1. Test recursive dependency resolution
2. Test handling of version constraints
3. Test conflict resolution
4. Test handling of circular dependencies
5. Test dependency caching

## Implementation Requirements
1. Create a dependency resolver following the specified approach
2. Implement recursive resolution of dependencies
3. Add version constraint checking
4. Create conflict resolution strategy
5. Implement dependency caching
6. Add visualization of dependency trees
7. Create proper error handling for resolution failures

## Integration Points
- Integrate with CPAN metadata retrieval
- Design for use by module installation process

## Documentation Requirements
1. Document the dependency resolution algorithm
2. Explain conflict resolution strategy
3. Add information about handling circular dependencies
4. Include examples of dependency resolution

## Considerations
- Balance thoroughness with performance
- Consider memory usage for large dependency trees
- Ensure reliable resolution in edge cases
```

#### Prompt 21: Module Installation Process

```
# Implement Module Installation Process

## Context
PVI needs to install CPAN modules for specific Perl versions. This step implements the module installation process.

## Previous Implementation
- CPAN metadata retrieval
- Dependency resolution for modules

## Current Task
Implement the module installation process for CPAN modules, including downloading, building, testing, and installation.

## Test Requirements
1. Test module download and extraction
2. Test module build process
3. Test module testing (optional)
4. Test module installation
5. Test handling of installation failures
6. Test installation with and without dependencies

## Implementation Requirements
1. Create module download and extraction functionality
2. Implement build process management
3. Add module testing capability (configurable)
4. Create installation process
5. Implement handling of install prerequisites
6. Add installation tracking
7. Create cleanup procedures
8. Implement proper error handling and reporting

## Integration Points
- Integrate with dependency resolution
- Connect with Perl version management

## Documentation Requirements
1. Document the installation process
2. Explain configuration options for installation
3. Add installation command documentation
4. Include troubleshooting information

## Considerations
- Balance installation thoroughness with speed
- Consider handling of failed dependencies
- Ensure consistent installation across platforms
```

#### Prompt 22: Module Management Commands

```
# Implement Module Management Commands

## Context
PVI needs a complete set of commands for managing modules. This step implements the full set of module management commands specified in the documentation.

## Previous Implementation
- Module installation process
- Dependency resolution
- CPAN integration

## Current Task
Implement the complete set of module management commands for PVI, including install, list, update, remove, search, deps, bundle, etc.

## Test Requirements
1. Test all module management commands
2. Test command output formatting
3. Test error handling for various scenarios
4. Test integration with installed Perl versions
5. Test bundle export and import

## Implementation Requirements
1. Implement pvi install command
2. Create pvi list command
3. Add pvi update command
4. Implement pvi remove command
5. Create pvi search command
6. Add pvi deps command
7. Implement pvi bundle export/import commands
8. Create pvi mirror command
9. Add pvi outdated command
10. Implement proper error handling for all commands

## Integration Points
- Integrate with module installation process
- Connect with command router for PVI commands

## Documentation Requirements
1. Document all module management commands
2. Explain command options and arguments
3. Add examples for common use cases
4. Include bundle format documentation

## Considerations
- Ensure consistent command behavior
- Consider performance for listing and searching
- Provide helpful output for all commands
```

#### Prompt 23: Type Definition Support - Basic Format

```
# Implement Basic Type Definition Support

## Context
The PVM Ecosystem includes PSC (Perl Script Compiler) for type checking. PVI needs to support type definitions for modules. This step implements the basic type definition format and support.

## Previous Implementation
- Module management commands
- CPAN integration

## Current Task
Implement the basic type definition support, including the format definition and management in PVI.

## Test Requirements
1. Test type definition file format parsing
2. Test type definition storage
3. Test type definition retrieval
4. Test basic type definition validation
5. Test the pvi type command

## Implementation Requirements
1. Define and implement the Perl Type Definition (ptd) file format
2. Create storage for type definitions in XDG_DATA_HOME/pvm/type_definitions/
3. Implement the pvi type command for managing definitions
4. Add type definition validation
5. Create retrieval functionality for type definitions
6. Implement proper error handling for definition management

## Integration Points
- Integrate with module management
- Design for use by PSC type checking later

## Documentation Requirements
1. Document the type definition file format
2. Explain type definition management
3. Add command documentation for pvi type
4. Include examples of type definitions

## Considerations
- Ensure format is extensible for future type system enhancements
- Consider version compatibility for type definitions
- Design storage structure for efficient retrieval
```

### Phase 4: Type System

#### Prompt 24: Parser Enhancement - Type Annotation Syntax

```
# Implement Parser Enhancement for Type Annotations

## Context
PSC (Perl Script Compiler) requires parser enhancements to support type annotations in Perl code. This step implements the necessary parser extensions.

## Previous Implementation
- Basic type definition support
- Module management system

## Current Task
Implement parser enhancements to support type annotations in Perl code, following the specification for grammar extensions.

## Test Requirements
1. Test parsing of type annotations in variable declarations
2. Test parsing of type annotations in subroutine declarations
3. Test parsing of type annotations in method declarations
4. Test parsing of type annotations in attribute declarations
5. Test parsing of complex type expressions
6. Test error handling for invalid syntax

## Implementation Requirements
1. Select and integrate a parser framework (possibly tree-sitter as mentioned)
2. Implement extensions for variable declaration annotations
3. Add extensions for subroutine and method annotations
4. Create extensions for attribute annotations
5. Implement parsing of type expressions (simple, parameterized, union, etc.)
6. Add error reporting for parsing issues
7. Create AST representation of typed code

## Integration Points
- Design for use by type checking system later
- Connect with type definition support

## Documentation Requirements
1. Document the type annotation syntax
2. Explain parser extension approach
3. Add examples of supported syntax
4. Include information about the AST representation

## Considerations
- Ensure backward compatibility with regular Perl syntax
- Consider performance for parsing large files
- Design for extensibility of the type system
```

#### Prompt 25: Type Checking System - Basic Implementation

```
# Implement Basic Type Checking System

## Context
PSC needs to implement a type checking system for analyzing Perl code with type annotations. This step implements the basic type checking system.

## Previous Implementation
- Parser enhancements for type annotations
- Type definition support

## Current Task
Implement the basic type checking system that can validate types according to the specification's type hierarchy.

## Test Requirements
1. Test type checking of variable declarations
2. Test type checking of function parameters and returns
3. Test type checking of assignments
4. Test validation against the type hierarchy
5. Test handling of type errors

## Implementation Requirements
1. Implement the core type hierarchy as specified
2. Create type checking logic for variable declarations
3. Add type checking for function parameters and returns
4. Implement assignment compatibility checking
5. Create type error reporting
6. Add integration with type definitions
7. Implement basic PSC commands for checking

## Integration Points
- Integrate with parser enhancements
- Connect with type definition support

## Documentation Requirements
1. Document the type checking system
2. Explain the type hierarchy
3. Add information about type compatibility rules
4. Include examples of type checking

## Considerations
- Balance type checking thoroughness with performance
- Consider incremental type checking for large files
- Design for extensibility of the type system
```

#### Prompt 26: Type Checking System - Flow-Sensitive Analysis

```
# Implement Flow-Sensitive Type Analysis

## Context
PSC's type system should include flow-sensitive analysis that recognizes validation patterns to refine types. This step implements flow-sensitive type analysis.

## Previous Implementation
- Basic type checking system
- Parser enhancements for type annotations

## Current Task
Implement flow-sensitive type analysis that can refine types based on control flow and validation patterns.

## Test Requirements
1. Test type refinement with simple conditions
2. Test type refinement with validation functions
3. Test handling of branches (if/else)
4. Test loop handling
5. Test handling of potentially undefined values

## Implementation Requirements
1. Implement control flow analysis
2. Create type refinement based on conditions
3. Add recognition of common validation patterns
4. Implement handling of branches with different type states
5. Create merge points for type states
6. Add support for common type guards
7. Implement proper error reporting for flow-sensitive issues

## Integration Points
- Integrate with basic type checking system
- Connect with AST traversal from parser

## Documentation Requirements
1. Document flow-sensitive analysis capabilities
2. Explain supported validation patterns
3. Add examples of type refinement
4. Include information about type guards

## Considerations
- Balance analysis thoroughness with performance
- Consider handling of complex control flow patterns
- Design for extensibility of recognized patterns
```

#### Prompt 27: PSC Commands Implementation

```
# Implement PSC Commands

## Context
PSC needs a complete set of commands for type checking and management. This step implements the full set of PSC commands specified in the documentation.

## Previous Implementation
- Type checking system with flow-sensitive analysis
- Parser enhancements for type annotations

## Current Task
Implement the complete set of PSC commands, including check, strip, run, watch, and def.

## Test Requirements
1. Test all PSC commands
2. Test command output formatting
3. Test error handling for various scenarios
4. Test integration with type definitions
5. Test watch mode functionality

## Implementation Requirements
1. Implement psc check command
2. Create psc strip command
3. Add psc run command
4. Implement psc watch command
5. Create psc def command
6. Add command-line options for all commands
7. Implement proper error handling and reporting
8. Create output formatting for type errors

## Integration Points
- Integrate with type checking system
- Connect with PVX for the run command
- Integrate with command router for PSC commands

## Documentation Requirements
1. Document all PSC commands
2. Explain command options and arguments
3. Add examples for common use cases
4. Include type error format documentation

## Considerations
- Ensure consistent command behavior
- Consider performance for large projects
- Provide helpful error messages for users
```

#### Prompt 28: Cross-Component Integration

```
# Implement Full Cross-Component Integration

## Context
The PVM Ecosystem consists of four main components (PVM, PVX, PVI, PSC) that need to work together seamlessly. This step implements the full cross-component integration.

## Previous Implementation
- All individual components with their functionality
- Basic integration between some components

## Current Task
Implement the full cross-component integration to ensure all parts of the system work together seamlessly as specified in the integration points section of the specification.

## Test Requirements
1. Test PVM and PVX integration
2. Test PVX and PVI integration
3. Test PSC and PVX integration
4. Test PSC and PVI integration
5. Test end-to-end workflows across all components

## Implementation Requirements
1. Finalize PVM and PVX integration for version resolution
2. Complete PVX and PVI integration for module installation
3. Implement PSC and PVX integration for type-checked execution
4. Finalize PSC and PVI integration for type definitions
5. Create unified error handling across components
6. Add cross-component communication mechanisms
7. Implement end-to-end workflows that use multiple components

## Integration Points
This task focuses on all integration points mentioned in the specification:
- PVM and PVX Integration
- PVX and PVI Integration
- PSC and PVX Integration
- PSC and PVI Integration

## Documentation Requirements
1. Document how components work together
2. Explain cross-component workflows
3. Add troubleshooting information for integration issues
4. Include examples of end-to-end usage

## Considerations
- Ensure consistent behavior across all components
- Consider performance implications of cross-component calls
- Design for future extensibility of the integration
```

#### Prompt 29: Editor Integration Support

```
# Implement Editor Integration Support

## Context
The PVM Ecosystem should support integration with text editors and IDEs for features like type checking, autocomplete, and error reporting. This step implements editor integration support.

## Previous Implementation
- Full cross-component integration
- Complete type checking system

## Current Task
Implement editor integration support for the PVM Ecosystem, focusing on PSC for type checking and error reporting.

## Test Requirements
1. Test language server protocol implementation
2. Test error reporting format
3. Test type information queries
4. Test autocompletion suggestions
5. Test integration with common editors (VSCode, etc.)

## Implementation Requirements
1. Implement a language server protocol (LSP) interface
2. Create structured output formats for errors and warnings
3. Add type information query functionality
4. Implement autocompletion suggestion generation
5. Create editor-specific configuration examples
6. Add documentation for editor integration

## Integration Points
- Integrate with PSC type checking system
- Connect with error reporting framework

## Documentation Requirements
1. Document the LSP implementation
2. Explain how to set up editor integration
3. Add editor-specific setup guides
4. Include troubleshooting information

## Considerations
- Ensure responsiveness for editor integration
- Consider incremental analysis for better performance
- Design for compatibility with multiple editors
```

#### Prompt 30: Final Polishing and Integration

```
# Implement Final Polishing and Integration

## Context
With all major components and integrations implemented, this final step focuses on polishing the entire system, improving performance, fixing edge cases, and ensuring a seamless user experience.

## Previous Implementation
- All major components and their integration
- Editor integration support

## Current Task
Implement final polishing and integration tasks to ensure the entire PVM Ecosystem works smoothly, efficiently, and provides a great user experience.

## Test Requirements
1. Test end-to-end workflows with real-world projects
2. Test performance with large codebases
3. Test edge cases identified during development
4. Test cross-platform compatibility
5. Test upgrade paths and migration

## Implementation Requirements
1. Optimize critical performance paths
2. Fix any remaining edge cases or bugs
3. Improve error messages and help text
4. Enhance documentation with real-world examples
5. Add migration guides for existing tool users
6. Create getting started documentation
7. Implement consistent styling and messaging
8. Add telemetry (if specified) with opt-out

## Integration Points
- Ensure all components work together seamlessly
- Verify external tool integration (editors, build systems)

## Documentation Requirements
1. Finalize all documentation
2. Create comprehensive user guide
3. Add complete command reference
4. Include real-world examples and tutorials
5. Create troubleshooting guide

## Considerations
- Focus on user experience as the primary goal
- Ensure consistency across all components
- Consider future maintenance and extensibility
```

## Implementation Notes

This implementation plan follows a logical progression from the foundation to more advanced features. Each step builds upon previous work, ensuring that the system is developed incrementally with proper testing at every stage.

Key aspects of this approach:

1. **Test-Driven Development**: Each step begins with test requirements before implementation
2. **Incremental Progress**: Components are built in small, manageable steps
3. **Integration Focus**: Clear integration points ensure components work together seamlessly
4. **Documentation Emphasis**: Documentation is treated as a first-class requirement

By following these prompts in sequence, a code-generation LLM can implement the PVM Ecosystem in a structured, testable manner that aligns with the specification requirements.
