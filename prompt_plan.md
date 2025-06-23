# PVM Split Build/Install Commands Implementation Plan

## Project Overview

This plan implements Issue #36: Split build-perl into separate build and install commands, add relocatable build support, and implement custom binary mirror functionality.

**Key Deliverables:**
1. Split `build-perl` command into `build-perl` (build-only) and `install-perl` (install-only)
2. Add relocatable build configuration for binary portability
3. Implement custom binary mirror support with authentication
4. Create automated binary publishing workflow
5. Maintain backward compatibility

## Architecture Analysis

**Current State:**
- `build-perl` command builds AND installs Perl
- `install` command supports binary-first installation with source fallback
- Mirror system exists with health checking and failover
- Configuration system supports TOML-based settings
- Build system is well-separated from installation logic

**Target State:**
- `build-perl` builds Perl without installation (relocatable by default)
- `install-perl` installs from build directories, archives, or URLs
- `install` command maintains current behavior for backward compatibility
- Custom mirror configuration with authentication support
- GitHub Actions workflow for automated binary distribution

## Implementation Steps

### Phase 1: Core Infrastructure (Steps 1-3)

#### Step 1: Relocatable Build Foundation ✅ COMPLETED

```
Implement relocatable build support in the existing build system without changing command behavior. This establishes the foundation for portable binaries while maintaining current functionality.

Add relocatable Configure options to the Perl build process:
- Add `-Duserelocatableinc` flag to Configure options
- Update platform-specific configure options in `internal/perl/build.go`
- Create tests to verify relocatable builds work correctly
- Ensure backward compatibility with existing build behavior

Test that built Perl installations can be moved to different directories and still function correctly. This is critical for binary distribution.

Key files to modify:
- `internal/perl/build.go` - Add relocatable configure options
- Create test cases for relocatable functionality

Success criteria:
- Existing `build-perl` command continues to work unchanged
- Built Perl installations are relocatable (can be moved to different directories)
- All existing tests continue to pass
- New tests verify relocatable functionality

COMPLETED: Commit f7e663a - Added -Duserelocatableinc flag to Configure options
and comprehensive test to verify relocatable functionality. All tests pass.
```

#### Step 2: Build Output Directory Support ✅ COMPLETED

```
Add `--output-dir` flag to `build-perl` command to control where Perl is built without changing installation behavior. This prepares for the build/install split while maintaining backward compatibility.

Extend the existing `build-perl` command with:
- Add `--output-dir` flag to specify build destination
- Modify build logic to respect output directory
- Maintain current installation behavior when flag is not used
- Add validation for output directory permissions and space
- Create comprehensive tests for different output directory scenarios

The build should create a complete, relocatable Perl installation in the specified directory that could be packaged for distribution.

Key files to modify:
- `internal/pvm/command.go` - Add --output-dir flag to build-perl command
- `internal/perl/build.go` - Modify build logic to support custom output directory
- Add tests for output directory functionality

Success criteria:
- `--output-dir` flag works correctly with build-perl
- Built Perl is complete and functional in output directory
- Without flag, behavior is identical to current implementation
- All existing tests pass, new tests cover output directory functionality

COMPLETED: Commit b2cf982 - Added --output-dir flag to build-perl command with logic to
prioritize output directory over prefix. Added comprehensive test for custom output
directory functionality. All tests pass at 100% rate.
```

#### Step 3: Custom Mirror Configuration Structure ✅ COMPLETED

```
Implement the configuration structure for custom binary mirrors without changing command behavior. This establishes the data model for custom mirrors while maintaining current functionality.

Add to the existing configuration system:
- Extend `PVMBinaryConfig` to support custom mirrors
- Add authentication configuration structure
- Support for multiple mirror fallback priority
- Per-version custom build mappings
- Validation for mirror URLs and authentication settings

The configuration should integrate with the existing mirror system and be ready for use by commands in later steps.

Key files to modify:
- `internal/config/types.go` - Add custom mirror configuration structures
- `internal/config/config.go` - Add validation and loading logic
- Add tests for configuration parsing and validation

Success criteria:
- Configuration structure supports all custom mirror requirements
- Configuration validates correctly (URLs, authentication, etc.)
- Existing configuration behavior is unchanged
- Tests cover all configuration scenarios

COMPLETED: Commit c073fbd - Added PVMCustomMirrorConfig, PVMCustomMirrorAuth, and
PVMCustomMirrorOAuth2 types with comprehensive validation. Supports all authentication
methods (none, basic, bearer, api-key, oauth2), URL templating, version mapping, and
custom headers. Complete test coverage with 4305 tests passing.
```

### Phase 2: Command Implementation (Steps 4-6)

#### Step 4: Build-Only Mode Implementation ✅ COMPLETED

```
Modify `build-perl` command to support build-only mode while maintaining backward compatibility. This implements the core split functionality without breaking existing workflows.

Implement build-only functionality:
- Add `--build-only` flag to disable automatic installation
- Modify build logic to skip installation when flag is set
- Ensure output directory contains complete, relocatable Perl installation
- Add progress reporting for build-only operations
- Create archive-ready directory structure for distribution

When `--build-only` is used, the command should build a complete Perl installation in a directory that can be archived or moved to other systems.

Key files to modify:
- `internal/pvm/command.go` - Add --build-only flag and logic
- `internal/perl/build.go` - Separate build and install phases
- Add tests for build-only functionality

Success criteria:
- `--build-only` flag prevents automatic installation
- Built directory contains complete, functional Perl installation
- Default behavior (with installation) is unchanged
- Progress reporting works correctly for build-only mode
- All tests pass including new build-only tests

COMPLETED: Added --build-only flag to build-perl command that skips installation
stage when enabled. Added BuildOnly field to BuildOptions struct and updated
BuildPerl function to conditionally skip installation and version registration.
Added comprehensive test coverage. All 4306 tests pass at 100% rate.
```

#### Step 5: Install-Perl Command Foundation ✅ COMPLETED

```
Create new `install-perl` command that can install Perl from build directories. This establishes the install-only command without adding complex features yet.

Implement `install-perl` command:
- Create new command structure in CLI system
- Support installation from build directory with `--from-build` flag
- Integrate with existing PVM version management
- Add progress reporting for installation operations
- Validate source directory contains complete Perl installation

The command should be able to take a directory created by `build-perl --build-only` and install it into PVM's version management system.

Key files to modify:
- `internal/pvm/command.go` - Add new install-perl command
- Create installation logic for build directories
- Add tests for install-perl functionality

Success criteria:
- `install-perl --from-build [directory]` works correctly
- Installed Perl is registered with PVM version management
- Progress reporting shows installation status
- Validation prevents installation of incomplete builds
- All tests pass including new install-perl tests

COMPLETED: Added newInstallPerlCommand() function with --from-build, --version, and --force flags.
Implemented installPerlFromBuild() function with complete installation logic including directory
validation, version detection, registry integration, and comprehensive directory copying.
Added test coverage for command flag parsing. All 4311 tests pass at 100% rate.
```

#### Step 6: Archive Installation Support

```
Extend `install-perl` command to support installation from archives (tar.gz files). This adds key functionality for binary distribution without adding complexity of URL downloads yet.

Implement archive installation:
- Add support for `.tar.gz` archive files as input to `install-perl`
- Add archive extraction logic with proper error handling
- Validate archive contents before installation
- Add progress reporting for extraction and installation
- Support both compressed and uncompressed archives

The command should be able to install from archive files created by packaging build-only output directories.

Key files to modify:
- `internal/pvm/command.go` - Extend install-perl for archive support
- Add archive extraction and validation logic
- Add tests for archive installation

Success criteria:
- `install-perl [archive.tar.gz]` extracts and installs correctly
- Archive validation prevents installation of corrupted files
- Progress reporting covers extraction and installation phases
- Error handling provides clear messages for archive issues
- All tests pass including archive installation tests
```

### Phase 3: Advanced Features (Steps 7-8)

#### Step 7: Custom Mirror Integration

```
Integrate custom mirror configuration with the existing mirror system. This connects the configuration structure from Step 3 with the actual mirror resolution and download logic.

Implement custom mirror functionality:
- Extend `MirrorManager` to use custom mirror configuration
- Add authentication support for custom mirrors
- Implement mirror priority and fallback logic
- Add URL templating for flexible mirror patterns
- Support per-version custom build mappings

The mirror system should seamlessly integrate custom mirrors with existing GitHub releases and CDN mirrors.

Key files to modify:
- `internal/perl/mirrors.go` - Extend MirrorManager for custom mirrors
- Add authentication handling for different auth types
- Add tests for custom mirror resolution and fallback

Success criteria:
- Custom mirrors work with existing mirror health checking
- Authentication works for private mirrors
- Mirror priority and fallback logic functions correctly
- URL templating supports various mirror patterns
- All tests pass including custom mirror functionality
```

#### Step 8: URL Installation Support

```
Extend `install-perl` command to support direct URL installation with custom mirror integration. This completes the installation command functionality.

Implement URL installation:
- Add support for direct URLs as input to `install-perl`
- Integrate with custom mirror system for URL resolution
- Add download progress reporting and retry logic
- Support mirror override with `--mirror` flag
- Add authentication support for URL downloads

The command should be able to install from any valid Perl binary URL, with custom mirror configuration providing defaults.

Key files to modify:
- `internal/pvm/command.go` - Add URL support to install-perl
- Integrate with mirror system for URL downloads
- Add tests for URL installation with various mirror types

Success criteria:
- `install-perl [URL]` downloads and installs correctly
- Custom mirror configuration provides URL defaults
- `--mirror` flag overrides default mirror for single operation
- Authentication works with URL downloads
- All tests pass including URL installation tests
```

### Phase 4: Integration and Automation (Steps 9-11)

#### Step 9: Upload Command Implementation

```
Create `upload-binary` command for publishing built binaries to GitHub releases and custom mirrors. This enables the automated publishing workflow.

Implement upload functionality:
- Create new `upload-binary` command
- Support GitHub releases upload with authentication
- Support custom mirror upload endpoints
- Add archive creation from build directories
- Add validation and progress reporting for uploads

The command should be able to take build output and publish it to various mirror types for distribution.

Key files to modify:
- `internal/pvm/command.go` - Add upload-binary command
- Add GitHub API integration for releases
- Add custom mirror upload logic
- Add tests for upload functionality (with mocking)

Success criteria:
- `upload-binary` successfully uploads to GitHub releases
- Custom mirror uploads work with authentication
- Archive creation works correctly from build directories
- Progress reporting shows upload status
- All tests pass including upload functionality (mocked)
```

#### Step 10: Build-Upload Integration

```
Add `--upload` flag to `build-perl` command to enable direct build-and-upload workflow. This streamlines the binary publishing process.

Implement integrated build-upload:
- Add `--upload` flag to `build-perl` command
- Integrate upload logic with build process
- Support platform matrix building with `--platforms` flag
- Add configuration for default upload targets
- Maintain separation of build and upload logic for testing

The command should be able to build and immediately upload binaries in a single operation while maintaining the ability to build without uploading.

Key files to modify:
- `internal/pvm/command.go` - Add --upload flag to build-perl
- Integrate upload functionality with build process
- Add tests for integrated build-upload workflow

Success criteria:
- `build-perl --upload` builds and uploads in single operation
- Platform matrix building works correctly
- Upload can be disabled for testing (build-only)
- Configuration supports default upload targets
- All tests pass including integrated workflow tests
```

#### Step 11: GitHub Actions Workflow

```
Create GitHub Actions workflow for automated binary publishing with platform matrix and version management. This completes the automated distribution system.

Implement CI/CD workflow:
- Create `.github/workflows/build-perl-binaries.yml`
- Support platform matrix: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
- Version matrix for latest 2 major Perl versions
- Manual dispatch and scheduled triggers
- Release management with proper tagging

The workflow should automatically build and publish binaries for supported platforms and versions.

Key files to create:
- `.github/workflows/build-perl-binaries.yml` - Main workflow
- Add workflow testing and validation

Success criteria:
- Workflow builds binaries for all platform/version combinations
- Manual dispatch allows custom version building
- Scheduled builds maintain current version binaries
- Release tagging and asset management works correctly
- Workflow includes proper testing and validation
```

### Phase 5: Documentation and Polish (Step 12)

#### Step 12: Integration Testing and Documentation

```
Perform comprehensive integration testing and create user documentation. This ensures the complete system works together and users can adopt the new functionality.

Complete system integration:
- Create end-to-end tests covering complete workflows
- Add comprehensive documentation for new commands
- Update existing documentation for changed behavior
- Add troubleshooting guides for common issues
- Validate backward compatibility across all scenarios

The system should be fully functional with clear documentation and comprehensive test coverage.

Key tasks:
- Create integration tests for complete build-install workflows
- Add documentation for all new commands and flags
- Update configuration documentation for custom mirrors
- Add troubleshooting section for binary distribution

Success criteria:
- End-to-end tests validate complete workflows
- Documentation covers all new functionality
- Backward compatibility is fully maintained
- Troubleshooting guides address common issues
- All tests pass at 100% rate
```

## Prompt Structure

Each step above provides a complete prompt that:
1. Clearly states the objective and scope
2. Builds incrementally on previous steps
3. Maintains backward compatibility
4. Includes specific implementation guidance
5. Defines clear success criteria
6. Focuses on testing and validation

The prompts are designed to be executed in sequence, with each step building on the previous work while maintaining system stability and test coverage throughout the implementation process.

## Success Metrics

**Technical Success:**
- All existing functionality maintains backward compatibility
- 100% test pass rate throughout implementation
- Relocatable builds work across all supported platforms
- Custom mirrors integrate seamlessly with existing mirror system
- Automated workflow successfully publishes binaries

**User Experience Success:**
- Build-only workflow reduces installation overhead
- Install-from-archive workflow simplifies binary distribution
- Custom mirror configuration enables enterprise deployment
- Automated binaries reduce build times for end users
- Documentation clearly explains new functionality

**Project Impact:**
- Issue #36 requirements fully implemented
- Foundation established for future binary distribution enhancements
- Community benefits from automated binary availability
- Enterprise users can deploy custom builds effectively
