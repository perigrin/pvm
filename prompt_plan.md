# PVX Global Tool Execution Implementation Plan

## Overview

This plan implements the global tool execution functionality for PVX as specified in issue #29. The implementation enhances PVX to work like Python's `uvx` or Node.js's `npx`, allowing seamless execution of Perl tools from anywhere without project context.

## Architecture

The global tool execution system will extend PVX with the following components:

- **Tool Detection**: Identify when PVX is invoked in "tool mode" vs "script mode"
- **Tool Mapping**: Map common tool names to CPAN modules
- **Global Installation**: Install tools to system-level location with isolation
- **Shim Management**: Create and manage executable shims in PATH
- **Integration**: Seamlessly integrate with existing `pvm tool` commands

## Implementation Strategy

- **Test-Driven Development**: Write failing tests first, implement to pass
- **Incremental Build**: Each step builds on previous functionality
- **Backward Compatibility**: Maintain existing PVX script execution
- **Cross-Platform**: Support Windows, macOS, Linux from the start

---

## Step 1: Tool Detection and Execution Mode

**Goal**: Implement detection logic to determine when PVX should run in "tool mode" vs "script mode"

**Context**: Foundation for global tool execution - must accurately detect execution context and maintain backward compatibility.

```
Implement tool detection logic to distinguish between script execution and tool execution modes.

Create the foundational infrastructure for determining when PVX should operate in global tool mode versus traditional script execution mode.

**Requirements**:
1. Detect tool execution mode vs script execution mode
2. Maintain backward compatibility with existing PVX functionality
3. Add command-line argument parsing for tool mode
4. Create comprehensive test suite for mode detection

**Implementation Tasks**:

1. **Create tool package** in `internal/tool/`:
   - `detector.go`: Mode detection and argument parsing
   - `types.go`: Tool execution data structures
   - `errors.go`: Tool-specific error handling

2. **Mode Detection Logic**:
   - Implement `DetectExecutionMode()` function that analyzes command-line arguments
   - Check if first argument is a known tool name vs script path
   - Handle edge cases (tools named like files, files named like tools)
   - Add fallback logic for ambiguous cases

3. **Argument Parsing**:
   - Parse tool name and arguments from command line
   - Preserve existing script execution argument handling
   - Add validation for tool name format and characters
   - Support both `pvx tool-name args` and `pvx tool-name -- args` formats

4. **Backward Compatibility**:
   - Ensure existing PVX script execution continues to work
   - Add clear differentiation between tool and script execution
   - Maintain existing error messages for script execution
   - Add migration path for existing workflows

5. **Testing Requirements**:
   - Unit tests for mode detection with various argument patterns
   - Edge case testing for ambiguous tool/script names
   - Backward compatibility tests for existing script execution
   - Integration tests with current PVX functionality

**Success Criteria**:
- Mode detection works correctly for all argument patterns
- Existing PVX script execution remains unaffected
- Clear error messages for ambiguous cases
- Comprehensive test coverage for detection logic
- Foundation ready for tool mapping in Step 2

**Integration Points**:
- Extends existing PVX command-line processing
- Provides foundation for tool mapping in Step 2
- Maintains compatibility with current PVX architecture
```

## Step 2: Tool-to-Module Mapping System

**Goal**: Implement mapping system to translate tool names to CPAN modules

**Context**: Core functionality that enables seamless tool execution by mapping common tool names to their corresponding CPAN distributions.

```
Implement comprehensive tool-to-module mapping system with extensible configuration.

Build on Step 1's mode detection to add intelligent tool name resolution to CPAN modules.

**Requirements**:
1. Built-in mapping for common Perl tools
2. Extensible configuration system for custom mappings
3. CPAN search fallback for unknown tools
4. Validation and error handling for mappings

**Implementation Tasks**:

1. **Extend tool package** with mapping functionality:
   - `mapping.go`: Core tool-to-module mapping logic
   - `builtin.go`: Built-in mappings for common tools
   - `config.go`: Configuration file handling for custom mappings
   - `resolver.go`: CPAN search and resolution logic

2. **Built-in Tool Mappings**:
   - Implement hardcoded mappings for common tools:
     - `ack` -> `App::Ack`
     - `cpanm` -> `App::cpanminus`
     - `prove` -> `Test::Harness`
     - `perltidy` -> `Perl::Tidy`
     - `perlcritic` -> `Perl::Critic`
     - `fatpack` -> `App::FatPacker`
   - Add validation for built-in mappings
   - Support version-specific mappings where needed

3. **Configuration System**:
   - Create YAML/JSON configuration format for custom mappings
   - Support user-level and system-level configuration files
   - Add configuration validation and error reporting
   - Implement configuration merging (system + user)

4. **CPAN Resolution**:
   - Implement MetaCPAN API client for tool lookup
   - Add fuzzy matching for similar tool names
   - Support explicit module specification (`pvx App::Ack`)
   - Cache resolution results for performance

5. **Validation and Error Handling**:
   - Validate tool names against common patterns
   - Provide helpful error messages for unknown tools
   - Suggest alternatives for misspelled tool names
   - Handle API failures gracefully with fallbacks

**Testing Requirements**:
- Unit tests for all built-in tool mappings
- Configuration file parsing and validation tests
- MetaCPAN API integration tests (with mocking)
- Error handling tests for unknown tools
- Performance tests for resolution caching

**Success Criteria**:
- All common Perl tools resolve to correct CPAN modules
- Configuration system works reliably
- CPAN search provides helpful fallbacks
- Clear error messages for resolution failures
- Performance meets expectations for cached lookups

**Integration Points**:
- Uses mode detection from Step 1
- Provides tool resolution for installation in Step 3
- Enables extensibility for user customization
```

## Step 3: Global Tool Installation Infrastructure

**Goal**: Implement system-level tool installation with isolation and management

**Context**: Core installation system that manages global tool storage, isolation, and lifecycle management.

```
Implement comprehensive global tool installation system with isolation and lifecycle management.

Build on Steps 1-2 to add robust tool installation infrastructure that manages global tools separately from project dependencies.

**Requirements**:
1. System-level tool storage with isolation
2. Installation lifecycle management (install, update, remove)
3. Dependency resolution and conflict handling
4. Integration with existing PVM local-lib system

**Implementation Tasks**:

1. **Create installation package** in `internal/tool/install/`:
   - `installer.go`: Core installation logic
   - `storage.go`: Tool storage and isolation management
   - `lifecycle.go`: Install, update, remove operations
   - `deps.go`: Dependency resolution and conflict handling

2. **Storage System**:
   - Implement isolated tool storage in `~/.local/share/pvm/tools/`
   - Create separate local-lib for each tool to prevent conflicts
   - Add tool metadata storage (version, dependencies, install date)
   - Support storage cleanup and garbage collection

3. **Installation Logic**:
   - Implement tool installation using existing PVM local-lib system
   - Add progress reporting for installation operations
   - Support installation from CPAN and local files
   - Handle installation failures and rollback

4. **Lifecycle Management**:
   - Implement `InstallTool()` function for new tool installation
   - Add `UpdateTool()` for tool updates and version management
   - Create `RemoveTool()` for clean tool removal
   - Support `ListTools()` for installed tool inventory

5. **Dependency Resolution**:
   - Resolve tool dependencies during installation
   - Handle version conflicts between tools
   - Support dependency upgrade and downgrade
   - Add conflict detection and resolution strategies

**Testing Requirements**:
- Installation success and failure scenarios
- Tool isolation and storage verification
- Dependency resolution with various conflict scenarios
- Lifecycle operations (install, update, remove) testing
- Integration with existing PVM local-lib system

**Success Criteria**:
- Tools install successfully with proper isolation
- Lifecycle operations work reliably
- Dependency conflicts are handled gracefully
- Storage system is robust and maintainable
- Integration with PVM system is seamless

**Integration Points**:
- Uses tool mapping from Step 2
- Leverages existing PVM local-lib infrastructure
- Provides foundation for tool execution in Step 4
```

---
## Step 4: Tool Execution with Isolation

**Goal**: Implement isolated tool execution using installed global tools

**Context**: Core execution system that runs tools in their isolated environments while maintaining clean separation from project dependencies.

```
Implement isolated tool execution system that runs global tools in their dedicated environments.

Build on Steps 1-3 to add reliable tool execution with proper environment isolation and argument handling.

**Requirements**:
1. Isolated execution environment for each tool
2. Proper argument passing and environment setup
3. Integration with installation system
4. Error handling and debugging support

**Implementation Tasks**:

1. **Create execution package** in `internal/tool/exec/`:
   - `executor.go`: Core tool execution logic
   - `environment.go`: Environment setup and isolation
   - `args.go`: Argument parsing and passing
   - `process.go`: Process management and monitoring

2. **Execution Environment**:
   - Set up isolated environment for each tool execution
   - Configure PATH to include tool's local-lib binary directories
   - Set PERL5LIB to include tool's dependencies
   - Handle environment variable inheritance and isolation

3. **Argument Handling**:
   - Parse tool arguments from command line
   - Handle argument escaping and quoting
   - Support both simple and complex argument patterns
   - Pass arguments correctly to tool processes

4. **Process Management**:
   - Execute tools as child processes with proper I/O handling
   - Monitor tool execution and handle timeouts
   - Capture and forward stdout/stderr appropriately
   - Handle process termination and cleanup

5. **Error Handling**:
   - Detect and report tool execution failures
   - Provide debugging information for failed executions
   - Handle missing tools with helpful error messages
   - Support verbose mode for troubleshooting

**Testing Requirements**:
- Tool execution success with various argument patterns
- Environment isolation verification
- Process management and cleanup testing
- Error handling for missing and failed tools
- Integration with installation system

**Success Criteria**:
- Tools execute successfully in isolated environments
- Arguments are passed correctly to tools
- Error handling provides helpful debugging information
- Process management is robust and reliable
- Integration with installation system works seamlessly

**Integration Points**:
- Uses tool installation from Step 3
- Provides foundation for shim management in Step 5
- Enables complete tool execution workflow
```

## Step 4: Tool Execution with Isolation

**Goal**: Implement isolated tool execution using installed global tools

**Context**: Core execution system that runs tools in their isolated environments while maintaining clean separation from project dependencies.

```
Implement isolated tool execution system that runs global tools in their dedicated environments.

Build on Steps 1-3 to add reliable tool execution with proper environment isolation and argument handling.

**Requirements**:
1. Isolated execution environment for each tool
2. Proper argument passing and environment setup
3. Integration with installation system
4. Error handling and debugging support

**Implementation Tasks**:

1. **Create execution package** in `internal/tool/exec/`:
   - `executor.go`: Core tool execution logic
   - `environment.go`: Environment setup and isolation
   - `args.go`: Argument parsing and passing
   - `process.go`: Process management and monitoring

2. **Execution Environment**:
   - Set up isolated environment for each tool execution
   - Configure PATH to include tool's local-lib binary directories
   - Set PERL5LIB to include tool's dependencies
   - Handle environment variable inheritance and isolation

3. **Argument Handling**:
   - Parse tool arguments from command line
   - Handle argument escaping and quoting
   - Support both simple and complex argument patterns
   - Pass arguments correctly to tool processes

4. **Process Management**:
   - Execute tools as child processes with proper I/O handling
   - Monitor tool execution and handle timeouts
   - Capture and forward stdout/stderr appropriately
   - Handle process termination and cleanup

5. **Error Handling**:
   - Detect and report tool execution failures
   - Provide debugging information for failed executions
   - Handle missing tools with helpful error messages
   - Support verbose mode for troubleshooting

**Testing Requirements**:
- Tool execution success with various argument patterns
- Environment isolation verification
- Process management and cleanup testing
- Error handling for missing and failed tools
- Integration with installation system

**Success Criteria**:
- Tools execute successfully in isolated environments
- Arguments are passed correctly to tools
- Error handling provides helpful debugging information
- Process management is robust and reliable
- Integration with installation system works seamlessly

**Integration Points**:
- Uses tool installation from Step 3
- Provides foundation for shim management in Step 5
- Enables complete tool execution workflow
```

---

## Step 5: Shim Management and PATH Integration

**Goal**: Implement executable shim creation and PATH management for installed tools

**Context**: User experience enhancement that makes global tools available directly in PATH without requiring pvx prefix.

```
Implement comprehensive shim management system for seamless tool execution from PATH.

Build on Steps 1-4 to add PATH integration that makes global tools available as direct commands.

**Requirements**:
1. Executable shim creation for installed tools
2. PATH directory management and integration
3. Shell integration and activation
4. Shim lifecycle management (create, update, remove)

**Implementation Tasks**:

1. **Create shim package** in `internal/tool/shim/`:
   - `manager.go`: Core shim management logic
   - `generator.go`: Shim script generation
   - `path.go`: PATH directory management
   - `shell.go`: Shell integration and activation

2. **Shim Generation**:
   - Create executable shim scripts for each installed tool
   - Generate platform-specific shims (shell scripts for Unix, batch files for Windows)
   - Include proper shebang and error handling in shims
   - Add version information and metadata to shims

3. **PATH Management**:
   - Create and manage shim directory (e.g., `~/.local/share/pvm/shims/`)
   - Add shim directory to user's PATH
   - Handle PATH conflicts with existing tools
   - Support system-wide and user-specific installations

4. **Shell Integration**:
   - Generate shell activation scripts for bash, zsh, fish
   - Add PATH updates to shell configuration files
   - Support manual and automatic shell integration
   - Handle shell-specific syntax and requirements

5. **Shim Lifecycle**:
   - Create shims automatically when tools are installed
   - Update shims when tools are updated or reconfigured
   - Remove shims when tools are uninstalled
   - Handle shim conflicts and resolution

**Testing Requirements**:
- Shim generation for all supported platforms
- PATH integration and shell activation testing
- Shim lifecycle operations (create, update, remove)
- Conflict detection and resolution
- Cross-platform compatibility verification

**Success Criteria**:
- Shims are created correctly for all installed tools
- PATH integration works on all supported shells
- Shim lifecycle operations are reliable
- Conflicts are detected and handled gracefully
- Cross-platform compatibility is maintained

**Integration Points**:
- Uses tool installation and execution from Steps 3-4
- Provides foundation for pvm tool integration in Step 6
- Enables complete seamless tool execution experience
```

## Step 6: Integration with Existing PVM Tool Commands

**Goal**: Integrate global tool execution with existing `pvm tool` command structure

**Context**: Seamless integration that unifies tool management under the existing PVM tool interface while maintaining backward compatibility.

```
Integrate global tool execution with existing pvm tool commands for unified tool management.

Build on Steps 1-5 to create seamless integration with existing PVM tool infrastructure.

**Requirements**:
1. Extend existing `pvm tool` commands to support global tools
2. Maintain backward compatibility with existing functionality
3. Provide unified interface for tool management
4. Add configuration and preference management

**Implementation Tasks**:

1. **Extend existing pvm tool commands**:
   - Update `pvm tool install` to support global tool installation
   - Extend `pvm tool list` to show global tools alongside project tools
   - Add `pvm tool update` support for global tools
   - Enhance `pvm tool remove` to handle global tool removal

2. **Command Integration**:
   - Add global tool detection to existing tool commands
   - Implement scope selection (project vs global)
   - Support mixed project/global tool operations
   - Add clear indicators for global vs project tools

3. **Configuration Management**:
   - Extend existing configuration system for global tool preferences
   - Add global tool configuration to `pvm config` command
   - Support tool-specific configuration and settings
   - Implement configuration validation and migration

4. **Unified Interface**:
   - Create consistent command interface for all tool operations
   - Add help and documentation for global tool features
   - Implement tab completion for global tools
   - Provide migration path from existing workflows

5. **Backward Compatibility**:
   - Ensure existing project tool functionality continues to work
   - Maintain existing command-line interfaces and behaviors
   - Add deprecation warnings for conflicting features
   - Provide clear upgrade path documentation

**Testing Requirements**:
- Integration with all existing pvm tool commands
- Backward compatibility with existing functionality
- Configuration system integration and validation
- Command-line interface consistency
- Migration path testing

**Success Criteria**:
- All existing pvm tool commands work with global tools
- Backward compatibility is maintained
- Configuration system is properly integrated
- Command interface is consistent and intuitive
- Migration from existing workflows is smooth

**Integration Points**:
- Uses all functionality from Steps 1-5
- Integrates with existing PVM tool infrastructure
- Provides foundation for comprehensive testing in Step 7
```

## Step 7: Comprehensive Testing and Edge Cases

**Goal**: Ensure production readiness with comprehensive testing and edge case handling

**Context**: Validation step that ensures reliability, performance, and robustness across all supported platforms and usage patterns.

```
Implement comprehensive testing suite and handle all edge cases for production deployment.

Build on Steps 1-6 to ensure robust, reliable global tool execution across all supported platforms and usage scenarios.

**Requirements**:
1. Complete end-to-end testing across all platforms
2. Edge case handling and error recovery
3. Performance testing and optimization
4. Cross-platform compatibility validation

**Implementation Tasks**:

1. **Comprehensive Test Suite**:
   - End-to-end integration tests for complete tool execution workflow
   - Cross-platform testing on Windows, macOS, Linux
   - Performance testing with various tool sizes and dependency counts
   - Concurrent execution testing for multiple tools
   - Network failure and recovery testing for CPAN operations

2. **Edge Case Handling**:
   - Tool name conflicts with system commands
   - Corrupted tool installations and recovery
   - Disk space limitations during installation
   - Permission issues and privilege escalation
   - Network connectivity issues during tool resolution

3. **Error Recovery**:
   - Automatic recovery from failed installations
   - Rollback capabilities for corrupted tools
   - Graceful degradation when services are unavailable
   - Clear error messages for all failure scenarios
   - Diagnostic tools for troubleshooting

4. **Performance Optimization**:
   - Caching for tool resolution and metadata
   - Parallel operations for tool installation
   - Lazy loading for tool discovery
   - Memory usage optimization for large tool sets
   - Startup time optimization for tool execution

5. **Cross-Platform Compatibility**:
   - Windows-specific testing (batch files, PowerShell, permissions)
   - macOS-specific testing (Homebrew integration, quarantine)
   - Linux distribution testing (various shells, package managers)
   - Architecture testing (x86_64, ARM64)
   - Shell compatibility testing (bash, zsh, fish, PowerShell)

**Testing Requirements**:
- 100% test coverage for all global tool functionality
- Performance benchmarks and regression testing
- Cross-platform compatibility validation
- Edge case and error scenario testing
- Real-world usage testing with various tools

**Success Criteria**:
- All tests pass on all supported platforms
- Performance meets acceptable benchmarks
- Edge cases are handled gracefully
- Error recovery works reliably
- Cross-platform compatibility is verified

**Integration Points**:
- Validates all functionality from Steps 1-6
- Provides foundation for documentation in Step 8
- Ensures production readiness for deployment
```

---

## Step 8: Documentation and User Experience

**Goal**: Complete user documentation and polish the user experience

**Context**: Final step that ensures usability and provides comprehensive documentation for users and maintainers.

```
Complete comprehensive documentation and polish user experience for production deployment.

Finalize the global tool execution implementation with production-grade documentation and user experience.

**Requirements**:
1. Complete user documentation and guides
2. Developer documentation for maintenance
3. User experience polish and refinement
4. Migration and troubleshooting guides

**Implementation Tasks**:

1. **User Documentation**:
   - Complete command reference for all global tool operations
   - Tutorial for getting started with global tool execution
   - Common use cases and examples
   - Migration guide from existing workflows
   - Troubleshooting guide for common issues

2. **Developer Documentation**:
   - Architecture documentation for global tool system
   - API documentation for programmatic access
   - Maintenance guide for tool mappings and configuration
   - Extension guide for custom tool integrations
   - Testing guide for contributors

3. **User Experience Polish**:
   - Consistent error messages and help text
   - Progress indicators for long-running operations
   - Color-coded output for better readability
   - Interactive prompts for confirmation operations
   - Tab completion for tools and commands

4. **Migration Support**:
   - Automated migration from existing tool installations
   - Compatibility layer for existing workflows
   - Clear upgrade path documentation
   - Deprecation warnings for obsolete features
   - Rollback procedures for failed migrations

5. **Help and Support**:
   - Context-sensitive help for all commands
   - Examples and use cases in help text
   - Link to documentation and troubleshooting resources
   - Error code reference and resolution guide
   - Community support channel information

**Testing Requirements**:
- Documentation accuracy and completeness
- User experience testing with real users
- Migration path validation
- Help system functionality
- Error message clarity and usefulness

**Success Criteria**:
- Documentation is complete and accurate
- User experience is polished and intuitive
- Migration paths work reliably
- Help system is comprehensive and useful
- System is ready for production deployment

**Integration Points**:
- Documents all functionality from Steps 1-7
- Provides maintenance foundation for ongoing development
- Ensures production readiness and user adoption
```

---

## Implementation Summary

### Development Timeline
- **Step 1**: Tool Detection (2-3 days)
- **Step 2**: Tool Mapping (3-4 days)
- **Step 3**: Global Installation (4-5 days)
- **Step 4**: Tool Execution (3-4 days)
- **Step 5**: Shim Management (3-4 days)
- **Step 6**: PVM Integration (2-3 days)
- **Step 7**: Comprehensive Testing (3-4 days)
- **Step 8**: Documentation (2-3 days)

**Total Estimated Time**: 22-30 days

### Key Success Factors
1. **Test-Driven Development**: Write failing tests first for all functionality
2. **Incremental Integration**: Each step builds on and integrates with previous steps
3. **Backward Compatibility**: Maintain existing PVX functionality throughout
4. **Cross-Platform Focus**: Support Windows, macOS, Linux from the beginning
5. **User Experience**: Focus on seamless, intuitive tool execution

### Risk Mitigation
- **Isolation**: Tools are isolated to prevent conflicts
- **Backward Compatibility**: Existing functionality remains unaffected
- **Error Recovery**: Comprehensive error handling and recovery mechanisms
- **Testing**: Thorough testing across all platforms and scenarios
- **Documentation**: Complete documentation for users and maintainers

### Integration with Existing Systems
- **PVX Architecture**: Extends existing PVX without breaking changes
- **PVM Tool Commands**: Seamlessly integrates with existing `pvm tool` interface
- **Local-lib System**: Leverages existing PVM local-lib infrastructure
- **Configuration**: Uses existing PVM configuration system

This plan provides a solid foundation for implementing global tool execution in PVX with production-grade reliability, performance, and user experience that matches the convenience of Python's `uvx` and Node.js's `npx`.
## Implementation Summary

### Development Timeline
- **Step 1**: Tool Detection (2-3 days)
- **Step 2**: Tool Mapping (3-4 days)
- **Step 3**: Global Installation (4-5 days)
- **Step 4**: Tool Execution (3-4 days)
- **Step 5**: Shim Management (3-4 days)
- **Step 6**: PVM Integration (2-3 days)
- **Step 7**: Comprehensive Testing (3-4 days)
- **Step 8**: Documentation (2-3 days)

**Total Estimated Time**: 22-30 days

### Key Success Factors
1. **Test-Driven Development**: Write failing tests first for all functionality
2. **Incremental Integration**: Each step builds on and integrates with previous steps
3. **Backward Compatibility**: Maintain existing PVX functionality throughout
4. **Cross-Platform Focus**: Support Windows, macOS, Linux from the beginning
5. **User Experience**: Focus on seamless, intuitive tool execution

### Risk Mitigation
- **Isolation**: Tools are isolated to prevent conflicts
- **Backward Compatibility**: Existing functionality remains unaffected
- **Error Recovery**: Comprehensive error handling and recovery mechanisms
- **Testing**: Thorough testing across all platforms and scenarios
- **Documentation**: Complete documentation for users and maintainers

### Integration with Existing Systems
- **PVX Architecture**: Extends existing PVX without breaking changes
- **PVM Tool Commands**: Seamlessly integrates with existing `pvm tool` interface
- **Local-lib System**: Leverages existing PVM local-lib infrastructure
- **Configuration**: Uses existing PVM configuration system

This plan provides a solid foundation for implementing global tool execution in PVX with production-grade reliability, performance, and user experience that matches the convenience of Python's `uvx` and Node.js's `npx`.
