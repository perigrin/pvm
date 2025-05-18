# PVM Ecosystem Specification

## 1. Executive Summary

The PVM Ecosystem provides a comprehensive suite of tools for Perl development environment management. Built as a single Go binary with multiple entry points, the system consists of four main components:

- **PVM (Perl Version Manager)** - Manages Perl installations and versions
- **PVX (Perl Version eXecutor)** - Executes modules/scripts in isolated environments
- **PVI (Perl Version Installer)** - Manages CPAN modules for installed Perl versions
- **PSC (Perl Script Compiler)** - Provides static type checking for Perl code

This project aims to modernize Perl development workflows by providing fast, reliable tooling with a consistent interface across platforms, while ensuring integration with existing Perl version managers and supporting zero-configuration startup.

## 2. Design Principles

### 2.1 Zero-Configuration Operation

A key design principle is that the tools work immediately without requiring any setup:

- **No Initial Configuration Required**: `pvm` and `pvx` work out-of-the-box with sensible defaults
- **System Perl Detection**: Automatically detects and uses the system's installed Perl when no other version is specified
- **Progressive Configuration**: Configuration is optional and can be added incrementally as needed
- **Environment Detection**: Automatically detects project structure and adapts behavior accordingly

### 2.2 Integration with Existing Tools

The system seamlessly integrates with existing Perl version managers:

- **plenv Integration**: 
  - Detects and reads `.perl-version` files
  - Recognizes Perl versions installed via plenv in `~/.plenv/versions/`
  - Uses the same version specification format for compatibility

- **perlbrew Integration**:
  - Detects Perl versions installed via perlbrew in `~/perl5/perlbrew/perls/`
  - Reads perlbrew configuration and aliases
  - Supports the same version naming conventions

- **Migration Path**:
  - Provides commands to register existing installations (`pvm import-from plenv|perlbrew`)
  - Non-destructive operation that doesn't modify existing tool configurations
  - Clear documentation on migration strategies

### 2.3 Unified System

All components are part of a single binary, providing:

- Consistent interface and behavior
- Shared configuration and resources
- Integrated workflows across components
- Minimal dependencies and installation complexity

## 3. System Architecture

### 3.1 System Overview

```
┌─────────────────────────────────────────────────────────┐
│                     PVM Ecosystem                        │
│                                                         │
│  ┌─────────┐   ┌──────────┐   ┌────────────┐   ┌─────┐  │
│  │   PVM   │   │    PVX   │   │    PVI     │   │ PSC │  │
│  │ Command │   │ Command  │   │  Command   │   │ Cmd │  │
│  │ Router  │   │  Router  │   │   Router   │   │ Rtr │  │
│  └────┬────┘   └────┬─────┘   └─────┬──────┘   └──┬──┘  │
│       │             │               │             │     │
│       │             │               │             │     │
│       ▼             ▼               ▼             ▼     │
│  ┌─────────────────────────────────────────────────┐    │
│  │                  Core Components                 │    │
│  │                                                  │    │
│  │  ┌───────────┐  ┌────────────┐  ┌────────────┐  │    │
│  │  │  Version  │  │Environment │  │   Build    │  │    │
│  │  │  Manager  │  │  Manager   │  │  Manager   │  │    │
│  │  └───────────┘  └────────────┘  └────────────┘  │    │
│  │                                                  │    │
│  │  ┌───────────┐  ┌────────────┐  ┌────────────┐  │    │
│  │  │  Module   │  │   Config   │  │ Execution  │  │    │
│  │  │  Manager  │  │  Manager   │  │  Manager   │  │    │
│  │  └───────────┘  └────────────┘  └────────────┘  │    │
│  │                                                  │    │
│  │  ┌───────────┐  ┌────────────┐  ┌────────────┐  │    │
│  │  │   Type    │  │   Parser   │  │  Analysis  │  │    │
│  │  │  Manager  │  │  Manager   │  │  Manager   │  │    │
│  │  └───────────┘  └────────────┘  └────────────┘  │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### 3.2 Component Architecture

The PVM ecosystem follows a unified architecture with specialized entry points:

1. **Single Binary Design**:
   - One Go executable with four entry points based on filename/symlink (pvm, pvx, pvi, psc)
   - Shared core components and configuration
   - Platform-specific optimizations handled internally

2. **Component Integration**:
   - Core components shared across all entry points
   - Configuration system shared between all components
   - Type information shared between PSC and PVX for runtime execution

3. **Data Flow**:
   - Commands flow through the command router to appropriate managers
   - Version information is centrally managed
   - Module operations check version context
   - Execution environment isolation maintains separation
   - Type checking can be performed before execution

## 4. Configuration System

### 4.1 Configuration Locations

Following XDG Base Directory Specification with git-style layering:

- **Project configuration**: `.pvm/pvm.toml` in the project directory
- **User configuration**: `$XDG_CONFIG_HOME/pvm/pvm.toml` (defaults to `~/.config/pvm/pvm.toml`)
- **System configuration**: `/etc/pvm/pvm.toml`

All configuration files are **optional**. The system works without any configuration files present.

### 4.2 Data Directories

Runtime data is stored in XDG-compliant locations:

- **Installed Perl versions**: `$XDG_DATA_HOME/pvm/versions/`
- **Downloaded source archives**: `$XDG_CACHE_HOME/pvm/sources/`  
- **Shim executables**: `$XDG_DATA_HOME/pvm/shims/`
- **Type definitions**: `$XDG_DATA_HOME/pvm/type_definitions/`
- **Build cache**: `$XDG_CACHE_HOME/pvm/build/`

### 4.3 Configuration Format

All configuration uses TOML format for readability and ease of editing:

```toml
# Main PVM configuration section
[pvm]
default_perl = "5.38.0"
build_jobs = 4
download_mirror = "https://www.cpan.org/src/5.0"
version_aliases = { latest = "5.38.0", stable = "5.36.0" }

# Module management configuration
[pvi]
preferred_installer = "auto"
default_mirror = "https://cpan.metacpan.org"
test_during_install = false

# Execution environment configuration
[pvx]
cache_modules = true
cleanup_after = true
isolation_level = "medium"

# Type checking configuration
[psc]
type_definitions_path = "$XDG_DATA_HOME/pvm/type_definitions"
strict_mode = false
watch_exclude = ["**/node_modules/**", "**/.git/**"]
```

### 4.4 Default Behavior

Without configuration, the system follows these guidelines:

- Use system Perl if no other version is specified
- Look for existing plenv or perlbrew installations and use them if available
- Default to standard CPAN mirrors for modules
- Use reasonable isolation for script execution
- Store temporary files in appropriate system locations

## 5. Component Details

### 5.1 PVM (Perl Version Manager)

#### 5.1.1 Overview

PVM manages Perl installations with commands for installing, switching between, and organizing Perl versions.

#### 5.1.2 Key Features

- **Zero Configuration Start**: Works immediately with system Perl
- **Legacy Tool Detection**: Automatically detects plenv and perlbrew installations
- **Version Management**: Supports global, local, and shell-specific version selection
- **Multi-Platform Support**: Works on Linux, macOS, and Windows

#### 5.1.3 Existing Installation Detection

The system uses the following algorithm to detect existing Perl installations:

1. Check for plenv installations in `~/.plenv/versions/`
2. Check for perlbrew installations in `~/perl5/perlbrew/perls/`
3. Check for system Perl in PATH
4. Register found installations in an internal registry
5. Apply appropriate version name mapping for consistency

#### 5.1.4 Version Resolution

Version resolution follows a clear precedence:

1. Explicitly specified version
2. Project-local version from `.perl-version` or `.pvm/pvm.toml`
3. User-level version from plenv/perlbrew or pvm configuration
4. System Perl as fallback

#### 5.1.5 Commands

```
pvm install <version>         # Install a Perl version
pvm use <version>             # Use a specific version in the current shell
pvm global <version>          # Set the global Perl version
pvm local <version>           # Set the local version for a directory
pvm versions                  # List installed versions
pvm available                 # List available Perl versions
pvm exec <version> <command>  # Execute a command with a specific version
pvm uninstall <version>       # Remove a Perl version
pvm import-from plenv         # Register perls installed by plenv
pvm import-from perlbrew      # Register perls installed by perlbrew
pvm rehash                    # Rebuild shim executables
```

### 5.2 PVX (Perl Version eXecutor)

#### 5.2.1 Overview

PVX executes Perl code in isolated environments, handling dependencies and environment setup.

#### 5.2.2 Key Features

- **Immediate Usability**: Works without configuration using detected Perl versions
- **Environment Isolation**: Provides controlled execution environments
- **Dependency Management**: Automatically resolves and installs required modules
- **Type-Check Integration**: Optional integration with PSC for type checking

#### 5.2.3 Version Resolution

Uses the following precedence:
1. Explicitly specified version (`--perl=VERSION`)
2. Version from project-specific `.perl-version` file (plenv compatible)
3. Version from local pvm configuration
4. Version from user pvm configuration
5. Version from environment variables (PLENV_VERSION, PERLBREW_PERL)
6. System Perl

#### 5.2.4 Commands

```
pvx [options] <script|module> [args...]  # Run a script in an isolated environment

Options:
  --no-install                # Don't install missing modules
  --perl=VERSION              # Use a specific Perl version
  --root=DIR                  # Set environment root directory
  --type-check                # Enable type checking before execution
  --verbose                   # Show additional output
```

### 5.3 PVI (Perl Version Installer)

#### 5.3.1 Overview

PVI manages CPAN modules for installed Perl versions, handling installation, updates, and dependencies.

#### 5.3.2 Key Features

- **Compatibility Mode**: Can manage modules for Perl installations from plenv and perlbrew
- **Automatic Version Selection**: Uses the same version resolution algorithm as PVX
- **Bundle Support**: Export and import module collections
- **Type Definition Integration**: Support for PSC type definitions

#### 5.3.3 Commands

```
pvi install <module>          # Install a module
pvi list                      # List installed modules
pvi update [module...]        # Update modules
pvi remove <module>           # Remove a module
pvi search <query>            # Search available modules
pvi deps <module>             # Show module dependencies
pvi bundle export <file>      # Export a module bundle
pvi bundle import <file>      # Import a module bundle
pvi type <module>             # Manage type definitions for a module
pvi mirror [url]              # Set/get CPAN mirror
pvi outdated                  # Show outdated modules
```

#### 5.3.4 Dependency Resolution

The dependency resolution in PVI follows the pattern established in Menlo and App::cpanminus:

1. **Recursive Resolution**: Recursively resolves dependencies depth-first
2. **Conflict Resolution**: Uses a simple "latest wins" strategy for version conflicts
3. **Metadata Sources**:
   - Uses META.json/META.yml when available
   - Falls back to parsing POD/README for dependency information
   - Uses MYMETA files generated during configuration
4. **Caching Strategy**:
   - Caches metadata to avoid repeated requests
   - Uses a build cache to speed up repeated installations
5. **Version Constraints**:
   - Supports standard CPAN version constraints (>=, ==, !=, etc.)
   - Handles range specifications

The implementation will maintain compatibility with the cpanminus approach while improving performance through Go's concurrency capabilities.

### 5.4 PSC (Perl Script Compiler)

#### 5.4.1 Overview

PSC provides static type checking for Perl code, implementing a gradual typing system inspired by TypeScript.

#### 5.4.2 Key Features

- **Gradual Adoption**: Works with any Perl installation, including system Perl
- **Integration with Existing Toolchains**: Compatible with code managed by plenv/perlbrew
- **Zero Runtime Overhead**: Types are stripped after analysis
- **Flow-Sensitive Analysis**: Recognizes validation patterns to refine types

#### 5.4.3 Type System

The core type hierarchy is defined as follows:

```
Unknown  (unanalyzed expressions)
Any      (explicitly polymorphic)

├── Scalar
│   ├── Str
│   │   └── Num
│   │       └── Int
│   │           └── Bool
│   ├── Undef
│   └── Ref
│       ├── ScalarRef
│       ├── ArrayRef
│       ├── HashRef
│       ├── CodeRef
│       └── ...
├── List
│   ├── Array
│   └── Hash
├── Code
└── Glob
```

#### 5.4.4 Commands

```
psc check <file|dir>          # Check a file or directory for type errors
psc strip <file> [output]     # Strip type annotations from a file
psc run <file> [args...]      # Type-check and execute a file
psc watch <file|dir>          # Watch files and report type errors on change
psc def <module>              # Generate type definitions for a module
```

#### 5.4.5 Type Definition Files

Type definitions for modules are stored in Perl Type Definition (ptd) files:

```perl
# DBI.ptd
package DBI {
    class DBI::db {
        method prepare(Str $query) -> DBI::st;
        method selectall_arrayref(Str $query) -> ArrayRef[ArrayRef[Scalar]];
    }
    
    class DBI::st {
        method execute(@params) -> Bool;
        method fetchrow_array() -> List;
    }
}
```

#### 5.4.6 Tree-sitter Grammar Extensions

PSC requires extending the existing Tree-sitter grammar for Perl to support type annotations. The following are the specific places where the grammar needs to be updated:

1. **Variable Declarations**:
   - Scalar variables: `my Type $name`
   - Array variables: `my Type @array`
   - Hash variables: `my Type %hash`
   - With assignments: `my Type $var = value`

2. **Subroutine Declarations**:
   - Parameter types: `sub name(Type $param, AnotherType @array)`
   - Return types: `sub name() -> ReturnType`
   - Combined: `sub name(Type $param) -> ReturnType`

3. **Method Declarations**:
   - In regular packages: `sub method(Type $self, Type $param) -> ReturnType`
   - In class syntax: `method name(Type $param) -> ReturnType`

4. **Attribute Declarations**:
   - In class syntax: `field Type $attribute`
   - With default values: `field Type $attribute = default_value`

5. **Type Expressions**:
   - Simple types: `Int`, `Str`, `Bool`
   - Parameterized types: `ArrayRef[Type]`, `HashRef[KeyType, ValueType]`
   - Union types: `Type1|Type2`
   - Intersection types: `Type1&Type2`
   - Negation types: `!Type`

6. **Package-level Type Declarations**:
   - Type aliases: `type TypeName = Type`
   - Class declarations: `class ClassName { ... }`

7. **Typecast Expressions**:
   - Type assertion: `$var as Type`

These grammar extensions should be implemented in a way that remains backward compatible with regular Perl syntax and doesn't interfere with parsing of non-typed code.

## 6. Integration Points

### 6.1 PVM and PVX Integration

- PVX uses PVM to determine the appropriate Perl version for execution
- Execution environments inherit PVM version settings
- Both use the same version resolution algorithm

### 6.2 PVX and PVI Integration 

- PVX uses PVI to install required modules in isolated environments
- Module installation respects version-specific contexts
- Shared caching mechanisms improve performance

### 6.3 PSC and PVX Integration

- PSC can pass type-checked code to PVX for execution
- Command `psc run` performs type checking and then executes via PVX
- Type information from PSC is made available to PVX for optimizations

### 6.4 PSC and PVI Integration

- PVI can install type definitions alongside modules
- PSC uses PVI to discover module information for type checking
- Type definitions are stored in a centralized location for reuse

## 7. Error Handling and Communication

### 7.1 Error Categories

The PVM ecosystem defines standardized error categories shared by all components:

1. **Configuration Errors**: Issues with configuration files or settings
2. **Version Errors**: Problems with Perl version detection, resolution, or installation
3. **Module Errors**: Issues with CPAN modules installation or dependencies
4. **Execution Errors**: Problems during script or command execution
5. **Type Errors**: Type checking failures or inconsistencies
6. **System Errors**: Issues with the underlying operating system or environment
7. **User Input Errors**: Problems with command-line arguments or inputs

### 7.2 Error Format

All errors follow a consistent structured format:

```
ERROR_CODE: Brief description
  Detail: More detailed explanation of the issue
  Location: File or context where the error occurred
  Hint: Suggested resolution or more information
```

Error codes are prefixed with the component that generated them:
- `PVM-`: For version manager errors
- `PVX-`: For execution errors
- `PVI-`: For module installer errors
- `PSC-`: For type checking errors
- `CFG-`: For configuration errors
- `SYS-`: For system-level errors

### 7.3 Inter-Component Error Communication

When components interact, errors are passed between them in a structured way:

1. **Error Object Structure**:
   - Maintains the original error information
   - Adds context about the calling component
   - Preserves the full error chain

2. **Error Propagation Rules**:
   - Components can wrap errors from other components with additional context
   - The original error information is always preserved
   - Call stack information is maintained for debugging

3. **User Presentation**:
   - High-level errors are presented to users by default
   - Detailed information is available with verbose flags
   - Related errors are grouped logically

### 7.4 Error Handling Between PSC and PVX

For the specific case of PSC and PVX integration:

1. When PSC detects type errors during `psc run`:
   - Type errors are reported with file and line information
   - Execution is prevented if errors are present
   - A summary of errors is provided

2. When type checking succeeds but runtime errors occur:
   - PVX reports the runtime error
   - If possible, it correlates the runtime error with type information
   - The error message includes both type information and runtime context

## 8. Testing Strategy

### 8.1 Test-Driven Development Approach

The project follows a strict Test-Driven Development (TDD) approach:

1. **Write a failing test** that defines the expected behavior
2. **Verify the test fails** as expected, confirming it's testing the right thing
3. **Write the minimal code** necessary to make the test pass
4. **Refactor the code** while keeping tests passing
5. **Repeat** for each new feature or behavior

### 8.2 Test Categories

The testing strategy includes multiple test categories:

1. **Unit Tests**: Test individual functions and methods in isolation
   - Coverage target: >90% for all core functionality
   - Focus on boundary conditions and error cases
   - Use mocks and stubs for external dependencies

2. **Integration Tests**: Test interactions between components
   - Verify correct data flow between components
   - Test configuration sharing and resolution
   - Ensure proper error propagation

3. **End-to-End Tests**: Test complete workflows from user perspective
   - Simulate actual command execution
   - Test with real Perl versions and modules (in test environment)
   - Verify correct output and error handling

4. **Cross-Platform Tests**: Ensure consistent behavior across platforms
   - Test on Linux, macOS, and Windows
   - Verify platform-specific behavior (paths, shells, etc.)
   - Ensure consistent error handling across platforms

### 8.3 Test Infrastructure

The test infrastructure supports the TDD approach:

1. **Test Helpers**:
   - Mock filesystem interface for consistent testing
   - Fake CPAN server for testing module installation
   - Version manager simulation for testing without actual Perl installations
   - Customizable environment for testing different configurations

2. **Continuous Integration**:
   - Automated test runs on all pull requests
   - Cross-platform testing in CI
   - Performance regression testing
   - Integration with code coverage tools

3. **Test Organization**:
   - Tests mirroring the package structure
   - Clear naming convention for test cases
   - Separation of fast and slow tests
   - Comprehensive test documentation

### 8.4 Testing PSC Type Checking

For the PSC component specifically:

1. **Grammar Tests**:
   - Test parsing of type annotations in various contexts
   - Verify correct AST generation for typed code
   - Test error handling for incorrect type syntax

2. **Type Checking Tests**:
   - Test type inference with various code patterns
   - Verify correct handling of type constraints
   - Test flow-sensitive type refinement

3. **Type Definition Tests**:
   - Test parsing and validation of type definition files
   - Verify correct application of type definitions
   - Test generation of type definitions from modules

## 9. Implementation Strategy

### 9.1 Phase 1: Core Infrastructure

1. **Configuration System**
   - TOML configuration parser
   - XDG directory support
   - Legacy tool integration

2. **Version Management**
   - Perl version detection
   - plenv/perlbrew import
   - Version resolution

3. **Command Routing**
   - Entry point handling
   - Command registration
   - Help system

### 9.2 Phase 2: Version and Environment Management

1. **Perl Installation**
   - Source downloading
   - Build process
   - Installation registration

2. **Environment Setup**
   - Shim generation
   - PATH management
   - Environment variables

3. **Execution Preparation**
   - Isolation environment setup
   - Version switching
   - Command execution

### 9.3 Phase 3: Module Management

1. **CPAN Integration**
   - Module search and metadata
   - Dependency resolution
   - Installation strategies

2. **Module Maintenance**
   - Updates and removal
   - Bundle support
   - Version-specific management

3. **Type Definition Support**
   - Type definition format
   - Module type discovery
   - Definition management

### 9.4 Phase 4: Type System and Integration

1. **Parser Development**
   - Tree-sitter grammar enhancements
   - Type annotation parsing
   - Source analysis

2. **Type System Implementation**
   - Type hierarchy
   - Type inference
   - Flow-sensitive analysis

3. **Cross-Component Integration**
   - PSC and PVX execution flow
   - Type-checked script running
   - Editor integration

## 10. Migration Strategies

### 10.1 Coexistence Mode

- PVM can be used alongside plenv/perlbrew without conflicts
- Shared access to installed Perl versions
- No modification of existing tool configurations
- Clear documentation on interoperability

### 10.2 Gradual Migration

1. Start with `pvm import-from` to register existing installations
2. Begin using PVM commands while maintaining existing workflows
3. Gradually transition projects to use PVM configuration
4. Eventually consolidate on PVM ecosystem

### 10.3 Clean Start

For new projects or users:
1. Install PVM ecosystem
2. Use its native capabilities from the beginning
3. Follow best practices for project configuration
4. Leverage integrated type checking and module management

## 11. Appendix

### 11.1 Configuration Reference

Sample complete configuration:

```toml
# Main PVM configuration section
[pvm]
default_perl = "5.38.0"
build_jobs = 4
download_mirror = "https://www.cpan.org/src/5.0"
patches_dir = "$XDG_DATA_HOME/pvm/patches"
compiler = "gcc"
run_tests = true
version_aliases = { latest = "5.38.0", stable = "5.36.0", legacy = "5.32.1" }

# Module management configuration
[pvi]
preferred_installer = "auto"
default_mirror = "https://cpan.metacpan.org"
test_during_install = false
cache_modules = true
force_reinstall = false
check_signatures = true

# Execution environment configuration
[pvx]
cache_modules = true
cleanup_after = true
isolation_level = "medium"
always_install_deps = true
timeout = 300
max_memory = "512MB"

# Type checking configuration
[psc]
type_definitions_path = "$XDG_DATA_HOME/pvm/type_definitions"
strict_mode = false
watch_exclude = ["**/node_modules/**", "**/.git/**", "**/local/**"]
generate_missing_types = true
check_before_run = true
```

### 11.2 Supported Platforms

The PVM ecosystem is designed to work on:
- **Linux**: Major distributions including Ubuntu, Fedora, Debian, CentOS
- **macOS**: Both Intel and Apple Silicon architectures
- **Windows**: Windows 10 and 11, with integration for both Command Prompt and PowerShell

### 11.3 Type System Syntax Examples

Examples of PSC type annotations:

```perl
# Variable declarations
my Str $name = "John";
our Int @counts;
state HashRef[Str] $cache;

# Subroutine declarations
sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

sub process(Str $input, @options) -> ArrayRef[HashRef] {
    # Implementation
}

# Complex types
my Str|Undef $name;
sub validate(Any $value) -> Bool { ... }
my ArrayRef[HashRef[Str]] $complex_data;
```
