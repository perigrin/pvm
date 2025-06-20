# PVM Self-Updater Implementation Plan

## Overview

This plan implements the self-updater functionality for PVM as specified in issue #14. The implementation follows a test-driven, incremental approach that builds from basic version checking to full atomic binary replacement with rollback capabilities.

## Architecture

The self-updater will be implemented as a new `pvm update` command with the following components:

- **Version Detection**: Compare current version against GitHub releases
- **Platform Detection**: Identify OS/architecture for correct binary selection
- **Download Manager**: Secure binary download with validation
- **Atomic Replacement**: Safe binary replacement with backup/rollback
- **Integration Detection**: Handle Homebrew and other package managers

## Implementation Strategy

- **Test-Driven Development**: Write failing tests first, implement to pass
- **Incremental Build**: Each step builds on previous functionality
- **Safety First**: Comprehensive validation and rollback capabilities
- **Cross-Platform**: Support Windows, macOS, Linux from the start

---

## Step 1: Version Detection Infrastructure ✅ COMPLETED

**Goal**: Implement version checking against GitHub releases API

**Context**: Foundation for all update functionality - must accurately detect current version and compare against available releases.

```
Implement version detection and comparison system for PVM self-updater.

Create the foundational infrastructure for version checking that will be used by all subsequent update functionality.

**Requirements**:
1. Create version detection system that can identify current PVM version
2. Implement GitHub API client for release checking
3. Add semantic version comparison utilities
4. Create comprehensive test suite for version operations

**Implementation Tasks**:

1. **Create version package** in `internal/version/`:
   - `version.go`: Core version detection and comparison
   - `github.go`: GitHub API client for release checking
   - `types.go`: Version and release data structures

2. **Version Detection**:
   - Implement `GetCurrentVersion()` function that reads from build info
   - Add fallback to version flag parsing for development builds
   - Handle version string normalization (v1.0.0 vs 1.0.0)
   - Support pre-release version detection

3. **GitHub API Integration**:
   - Create `GitHubClient` struct with release fetching
   - Implement `GetLatestRelease()` and `GetReleaseByTag()` methods
   - Add proper error handling for network issues and API limits
   - Include authentication support for higher rate limits

4. **Version Comparison**:
   - Implement semantic version parsing and comparison
   - Support pre-release version handling (alpha, beta, rc)
   - Add version constraint matching for specific version updates
   - Create helper functions for version validation

5. **Testing Requirements**:
   - Unit tests for version parsing and comparison
   - Mock GitHub API tests for release fetching
   - Integration tests with real GitHub API (rate limited)
   - Edge case testing for malformed versions
   - Network failure simulation and recovery

**Success Criteria**:
- Current version detection works in all deployment scenarios
- GitHub API integration handles all response types correctly
- Version comparison follows semantic versioning rules
- Comprehensive error handling for network and API issues
- 100% test coverage for all version operations

**Integration Points**:
- Will be used by update command for version checking
- Provides foundation for download manager platform detection
- Enables update availability notifications
```

---

## Step 2: Platform Detection and Binary Selection ✅ COMPLETED

**Goal**: Implement cross-platform detection and binary selection logic

**Context**: Different platforms require different binaries. Must correctly identify platform and select appropriate download URLs.

```
Implement platform detection and binary selection for cross-platform updates.

Build on Step 1's version detection to add platform-aware binary selection from GitHub releases.

**Requirements**:
1. Detect current platform (OS, architecture) accurately
2. Map platform to GitHub release asset names
3. Handle special cases (Homebrew, development builds)
4. Validate binary compatibility before download

**Implementation Tasks**:

1. **Extend version package** with platform detection:
   - Add `platform.go`: Platform detection and binary mapping
   - Extend `github.go`: Asset filtering and selection
   - Update `types.go`: Platform and asset data structures

2. **Platform Detection**:
   - Implement `DetectPlatform()` using runtime.GOOS/GOARCH
   - Create platform normalization for GitHub asset naming
   - Add architecture mapping (amd64, arm64, etc.)
   - Handle Windows executable extension (.exe)

3. **Binary Selection Logic**:
   - Map platform to GitHub release asset patterns
   - Implement asset filtering by platform and architecture
   - Add checksum file detection and validation
   - Support multiple naming conventions for assets

4. **Special Case Handling**:
   - Detect Homebrew installation paths
   - Identify development builds vs release builds
   - Handle custom installation locations
   - Add warnings for unsupported platforms

5. **Binary Validation**:
   - Verify binary compatibility before download
   - Check architecture compatibility (arm64 vs amd64)
   - Validate file size and basic format
   - Support dry-run mode for testing

**Testing Requirements**:
- Platform detection accuracy across Windows, macOS, Linux
- Binary selection for all supported platform combinations
- Homebrew detection and handling
- Asset name pattern matching with real GitHub releases
- Edge case handling for unsupported platforms

**Success Criteria**:
- Accurate platform detection on all supported systems
- Correct binary selection from GitHub release assets
- Proper handling of Homebrew and package manager installations
- Clear error messages for unsupported platforms
- Dry-run mode works correctly for testing

**Integration Points**:
- Uses version detection from Step 1
- Provides platform info for download manager in Step 3
- Enables installation method detection for Step 4
```

---

## Step 3: Secure Download Manager ✅ COMPLETED

**Goal**: Implement secure binary download with integrity validation

**Context**: Downloads must be secure, validated, and handle network issues gracefully. Foundation for atomic replacement.

```
Implement secure download manager with integrity validation and error recovery.

Build on Steps 1-2 to add secure binary downloading with comprehensive validation and error handling.

**Requirements**:
1. Secure HTTPS download with progress tracking
2. Checksum validation and integrity verification
3. Robust error handling and retry logic
4. Temporary file management and cleanup

**Implementation Tasks**:

1. **Create download package** in `internal/download/`:
   - `downloader.go`: Core download logic with progress tracking
   - `validation.go`: Checksum and integrity verification
   - `retry.go`: Retry logic and error recovery
   - `temp.go`: Temporary file management

2. **Download Implementation**:
   - Create `Downloader` struct with progress callbacks
   - Implement streaming download with progress reporting
   - Add timeout handling and connection management
   - Support resume for interrupted downloads

3. **Integrity Validation**:
   - Download and verify SHA256 checksums
   - Implement file signature verification (if available)
   - Add basic binary format validation
   - Verify downloaded file size matches expected

4. **Error Handling and Retry**:
   - Implement exponential backoff for retries
   - Handle network timeouts and connection errors
   - Add user-friendly error messages
   - Support offline mode detection

5. **Temporary File Management**:
   - Create secure temporary files with proper permissions
   - Implement cleanup on success and failure
   - Add atomic file operations where possible
   - Handle disk space validation

**Testing Requirements**:
- Download success with various file sizes
- Checksum validation with corrupted files
- Network error simulation and retry logic
- Progress tracking accuracy
- Temporary file cleanup verification

**Success Criteria**:
- Downloads complete successfully with progress indication
- All integrity checks pass with valid files
- Network errors are handled gracefully with retries
- Temporary files are cleaned up properly
- Clear error messages for all failure scenarios

**Integration Points**:
- Uses platform detection from Step 2
- Provides validated binaries for atomic replacement in Step 4
- Enables progress reporting for user experience
```

---

## Step 4: Atomic Binary Replacement ✅ COMPLETED

**Goal**: Implement safe binary replacement with backup and rollback

**Context**: The critical operation that must be atomic and reversible. Cannot leave system in broken state.

```
Implement atomic binary replacement with backup and rollback capabilities.

Build on Steps 1-3 to add the core update functionality with comprehensive safety measures.

**Requirements**:
1. Atomic binary replacement without breaking running processes
2. Backup creation and rollback functionality
3. Permission and ownership preservation
4. Cross-platform compatibility for file operations

**Implementation Tasks**:

1. **Create updater package** in `internal/updater/`:
   - `replacer.go`: Atomic binary replacement logic
   - `backup.go`: Backup creation and management
   - `rollback.go`: Rollback functionality and validation
   - `permissions.go`: Cross-platform permission handling

2. **Atomic Replacement Logic**:
   - Implement atomic rename/move operations
   - Handle running process detection and warnings
   - Add file locking and exclusive access
   - Support cross-filesystem moves with copy+remove

3. **Backup Management**:
   - Create timestamped backups before replacement
   - Store backup metadata and validation info
   - Implement backup cleanup and retention policies
   - Add backup verification before replacement

4. **Rollback Implementation**:
   - Automatic rollback on replacement failure
   - Manual rollback command for user-initiated recovery
   - Rollback validation and integrity checking
   - Clear rollback status reporting

5. **Cross-Platform Considerations**:
   - Windows: Handle executable file locking and permissions
   - macOS: Preserve code signing and quarantine attributes
   - Linux: Handle file permissions and ownership
   - All: Atomic operations and proper error handling

**Testing Requirements**:
- Atomic replacement success and failure scenarios
- Backup creation and verification
- Rollback functionality with corrupted updates
- Permission preservation across platforms
- Running process detection and handling

**Success Criteria**:
- Binary replacement is atomic and never leaves broken state
- Backup and rollback work reliably
- Permissions and ownership are preserved
- Clear status reporting throughout operation
- Cross-platform compatibility verified

**Integration Points**:
- Uses validated binaries from Step 3
- Provides foundation for update command in Step 5
- Enables rollback functionality for Step 6
```

---

## Step 5: Update Command Implementation ✅ COMPLETED

**Goal**: Implement the `pvm update` command with full user interface

**Context**: User-facing command that orchestrates all previous components into a cohesive update experience.

```
Implement the complete `pvm update` command with comprehensive user interface and options.

Build on Steps 1-4 to create the complete user-facing update functionality with all command options.

**Requirements**:
1. Complete command interface with all specified options
2. Interactive and non-interactive modes
3. Comprehensive progress reporting and user feedback
4. Integration with all previous components

**Implementation Tasks**:

1. **Create update command** in `internal/pvm/commands/`:
   - `update.go`: Main update command implementation
   - `update_flags.go`: Command line flag definitions
   - `update_ui.go`: User interface and progress reporting
   - `update_validation.go`: Pre-update validation

2. **Command Interface**:
   - Implement `pvm update` (update to latest)
   - Add `pvm update --check` (check for updates only)
   - Support `pvm update --version v1.0.0` (specific version)
   - Include `pvm update --force` (skip checks)
   - Add `pvm update --dry-run` (show what would happen)

3. **User Experience**:
   - Interactive prompts for confirmation
   - Progress bars for download and installation
   - Clear status messages throughout process
   - Colorized output for better readability

4. **Pre-Update Validation**:
   - Check for sufficient disk space
   - Verify write permissions to installation directory
   - Detect and warn about running processes
   - Validate network connectivity

5. **Integration and Orchestration**:
   - Coordinate version checking, download, and replacement
   - Handle errors gracefully with appropriate user messaging
   - Implement proper cleanup on success and failure
   - Add comprehensive logging for troubleshooting

**Testing Requirements**:
- All command line options and combinations
- Interactive and non-interactive modes
- Progress reporting accuracy
- Error handling and user messaging
- Integration with all underlying components

**Success Criteria**:
- All command options work as specified
- User experience is clear and professional
- Error messages are helpful and actionable
- Progress reporting is accurate and responsive
- Integration with underlying components is seamless

**Integration Points**:
- Orchestrates all components from Steps 1-4
- Provides foundation for advanced features in Step 6
- Enables testing of complete update workflow
```

---

## Step 6: Advanced Features and Edge Cases

**Goal**: Implement advanced features and handle edge cases

**Context**: Handle special installation methods, add convenience features, and ensure robustness.

```
Implement advanced update features and comprehensive edge case handling.

Build on Steps 1-5 to add advanced functionality and handle all edge cases for production deployment.

**Requirements**:
1. Homebrew and package manager integration
2. Auto-update checking and notifications
3. Configuration and preference management
4. Comprehensive error recovery

**Implementation Tasks**:

1. **Package Manager Integration**:
   - Detect Homebrew installations and delegate to `brew upgrade`
   - Handle apt/dnf/pacman package manager installations
   - Add warnings for unsupported installation methods
   - Provide migration paths from package managers

2. **Auto-Update Features**:
   - Implement background update checking
   - Add configurable update notifications
   - Support update channels (stable, beta, alpha)
   - Create update scheduling and preferences

3. **Configuration Management**:
   - Add update preferences and settings
   - Implement configuration file handling
   - Support user-specific update policies
   - Add system-wide update configuration

4. **Advanced Error Recovery**:
   - Implement comprehensive rollback scenarios
   - Add recovery from corrupted downloads
   - Handle partial update states
   - Provide diagnostic and repair tools

5. **Shell Integration Updates**:
   - Handle shell configuration updates after replacement
   - Detect and update PATH modifications
   - Refresh shell integration automatically
   - Add compatibility with existing installations

**Testing Requirements**:
- Homebrew detection and delegation
- Auto-update checking and notifications
- Configuration file handling
- Complex error recovery scenarios
- Shell integration updates

**Success Criteria**:
- Homebrew installations are handled correctly
- Auto-update features work reliably
- Configuration is persistent and respected
- Recovery from all error states is possible
- Shell integration continues working after updates

**Integration Points**:
- Extends update command from Step 5
- Uses all underlying infrastructure from Steps 1-4
- Provides complete production-ready update system
```

---

## Step 7: Comprehensive Testing and Documentation

**Goal**: Ensure production readiness with complete testing and documentation

**Context**: Final step to ensure reliability, performance, and usability for production deployment.

```
Complete comprehensive testing, performance validation, and user documentation.

Finalize the self-updater implementation with production-grade testing and complete user documentation.

**Requirements**:
1. Complete end-to-end testing across all platforms
2. Performance and reliability validation
3. User documentation and troubleshooting guides
4. Integration with existing PVM documentation

**Implementation Tasks**:

1. **Comprehensive Test Suite**:
   - End-to-end integration tests for complete update workflow
   - Cross-platform testing on Windows, macOS, Linux
   - Network failure and recovery testing
   - Homebrew and package manager integration testing
   - Performance testing with large binaries

2. **Security and Reliability Testing**:
   - Security validation of download and verification
   - Stress testing with network interruptions
   - Concurrent update attempt handling
   - File system permission edge cases
   - Rollback reliability under various failure modes

3. **User Documentation**:
   - Update command reference documentation
   - Troubleshooting guide for common issues
   - Security and verification explanation
   - Configuration options and preferences
   - Migration guide from manual updates

4. **Integration Documentation**:
   - Developer guide for update system maintenance
   - Architecture documentation for future enhancements
   - API documentation for programmatic access
   - Monitoring and logging documentation

**Testing Requirements**:
- 100% test coverage for all update functionality
- Cross-platform compatibility validation
- Performance benchmarks and regression testing
- Security audit of download and verification processes
- Real-world usage testing with various installation methods

**Success Criteria**:
- All tests pass on all supported platforms
- Performance meets acceptable benchmarks
- Documentation is complete and accurate
- Security review identifies no issues
- Real-world testing confirms reliability

**Integration Points**:
- Validates all functionality from Steps 1-6
- Provides foundation for maintenance and enhancement
- Ensures production readiness for deployment
```

---

## Implementation Summary

### Development Timeline
- **Step 1**: Version Detection (2-3 days)
- **Step 2**: Platform Detection (2-3 days)
- **Step 3**: Download Manager (3-4 days)
- **Step 4**: Atomic Replacement (4-5 days)
- **Step 5**: Update Command (3-4 days)
- **Step 6**: Advanced Features (3-4 days)
- **Step 7**: Testing & Documentation (2-3 days)

**Total Estimated Time**: 19-26 days

### Key Success Factors
1. **Test-Driven Development**: Write failing tests first for all functionality
2. **Incremental Integration**: Each step builds on and integrates with previous steps
3. **Cross-Platform Focus**: Support Windows, macOS, Linux from the beginning
4. **Safety First**: Comprehensive validation and rollback at every step
5. **User Experience**: Clear progress reporting and error messaging throughout

### Risk Mitigation
- **Atomic Operations**: All file operations are atomic to prevent corruption
- **Comprehensive Testing**: Edge cases and error conditions are thoroughly tested
- **Platform Compatibility**: Cross-platform testing ensures reliability
- **Rollback Capability**: All operations can be reversed if they fail

This plan provides a solid foundation for implementing the PVM self-updater with production-grade reliability, security, and user experience.
