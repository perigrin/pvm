# Step 1.1: Code Analysis and Extraction Planning Report

## Executive Summary

Analysis of `internal/pm/command.go` (2403 lines) reveals significant code duplication and opportunities for extraction into reusable packages. The file contains 12 commands with substantial repeated patterns, particularly in provider setup (40-60 lines per command), configuration handling, and progress tracking.

## Current State Analysis

### File Structure Overview

**Target File:** `internal/pm/command.go` (2403 lines)
- **Commands Analyzed:** 12 primary commands + 4 simplified prototypes
- **Code Duplication:** High (estimated 60-70% repetitive patterns)
- **Complexity:** Very High (multiple responsibilities per command)

### Command Breakdown

| Command | Lines | Primary Functionality | Duplication Score |
|---------|-------|----------------------|-------------------|
| `install` | ~340 | Module installation with cpanfile support | High (90%) |
| `list` | ~123 | Module listing with formatting | High (85%) |
| `update` | ~158 | Module updates with batch processing | High (90%) |
| `remove` | ~56 | Module removal | Medium (70%) |
| `search` | ~63 | Module searching | High (85%) |
| `deps` | ~114 | Dependency resolution and display | High (90%) |
| `bundle` | ~232 | Bundle export/import operations | High (85%) |
| `add` | ~136 | cpanfile addition with installation | High (90%) |
| `sync` | ~119 | Snapshot generation/installation | Medium (75%) |
| `outdated` | ~126 | Outdated module detection | High (85%) |
| `mirror` | ~92 | Mirror configuration | Low (40%) |
| `type` | ~11 | Type definition management | Low (30%) |

### Existing Infrastructure

**Already Extracted Components:**
- `ProviderBuilder` (181 lines) - Fluent builder for CPAN providers ✅
- `CpanfileManager` (597 lines) - cpanfile manipulation ✅
- `internal/pm/modules/` - Module operations (installer, manager, etc.) ✅
- `internal/pm/deps/` - Dependency resolution ✅

**Prototype Implementations:**
The file contains simplified command implementations (`newSimplified*Command`) that demonstrate the target refactored patterns using helper functions.

## Identified Patterns and Duplication

### 1. Provider Setup Pattern (HIGH DUPLICATION)

**Frequency:** Found in 10 out of 12 commands
**Lines per occurrence:** 40-60 lines
**Total duplication:** ~500 lines

```go
// Pattern repeated in every command:
cfg, err := config.LoadEffectiveConfig()
result, err := NewProviderBuilder().
    WithConfig(cfg).
    WithSource(source).
    WithNoCache(noCache).
    WithResolver().  // Sometimes included
    Build()
provider := result.Provider
resolver := result.Resolver
```

**Extraction Target:** Already extracted into `ProviderBuilder` ✅

### 2. Project Context Detection Pattern (MEDIUM DUPLICATION)

**Frequency:** Found in 4 commands (install, add, sync, resolveModuleNames helper)
**Lines per occurrence:** 15-25 lines
**Total duplication:** ~80 lines

```go
// Pattern for project-aware behavior:
projectCtx, err := project.GetCurrentProject()
if !projectCtx.IsProject {
    return fmt.Errorf("not in a project directory...")
}
cpanfilePath := filepath.Join(projectCtx.RootDir, "cpanfile")
```

**Extraction Target:** `internal/project/helpers.go`

### 3. Progress Tracking Pattern (HIGH DUPLICATION)

**Frequency:** Found in 8 commands
**Lines per occurrence:** 10-20 lines
**Total duplication:** ~150 lines

```go
// Pattern for progress callbacks:
ProgressCallback: func(stage InstallProgressStage, module string, details string, progress float64) {
    if verbose {
        ui.Debug("[%s] %s: %s (%.0f%%)", stage.String(), module, details, progress*100)
    } else if stage != StageFinished {
        ui.Info("[%s] %s", stage.String(), module)
    }
},
```

**Extraction Target:** `internal/cli/progress/tracker.go`

### 4. Result Formatting Pattern (HIGH DUPLICATION)

**Frequency:** Found in 6 commands
**Lines per occurrence:** 30-50 lines
**Total duplication:** ~250 lines

```go
// Pattern for displaying results:
switch format {
case "json":
    jsonData, _ := json.MarshalIndent(results, "", "  ")
    ui.Println(string(jsonData))
case "simple":
    // Simple format logic
default:
    // Table format logic
}
```

**Extraction Target:** `internal/cli/progress/formatting.go`

### 5. Installation Options Pattern (HIGH DUPLICATION)

**Frequency:** Found in 5 commands
**Lines per occurrence:** 20-30 lines
**Total duplication:** ~125 lines

```go
// Pattern for creating installation options:
installOptions := &pmModules.ModuleInstallOptions{
    ModuleName:         moduleName,
    VersionConstraint:  version,
    PerlPath:           perlPath,
    InstallDir:         installDir,
    RunTests:           !skipTests,
    Force:              force,
    // ... 10+ more fields
}
```

**Extraction Target:** `internal/modules/options.go`

## Dependency Analysis

### Internal Dependencies

```
command.go depends on:
├── internal/cli (UI and CLI utilities)
├── internal/config (Configuration management)
├── internal/cpan (CPAN provider interfaces)
├── internal/modules (Module management - needs extraction)
├── internal/perl (Perl interpreter utilities)
├── internal/project (Project context detection)
├── internal/pm/deps (Dependency resolution)
├── internal/pm/modules (PM-specific module operations)
└── github.com/spf13/cobra (CLI framework)
```

### Circular Dependency Risks

1. **Low Risk:** Most dependencies are unidirectional
2. **Medium Risk:** `internal/modules` and `internal/pm/modules` overlap
3. **Mitigation:** Extract to unified `internal/modules` package

### External Dependencies

- `github.com/spf13/cobra` - CLI framework (stable)
- Standard library packages (stable)
- No breaking dependency changes expected

## Extraction Opportunities by Priority

### Phase 1: Foundation (HIGH IMPACT, LOW RISK)

1. **Core Types and Interfaces** - Extract shared data structures
   - Target: `internal/modules/types.go`
   - Impact: Enables all other extractions
   - Risk: Low (mostly type definitions)

2. **Configuration Helpers** - Extract config resolution patterns
   - Target: `internal/config/resolution.go`
   - Impact: Reduces 20-30 lines per command
   - Risk: Low (existing config package)

### Phase 2: Module Management (HIGH IMPACT, MEDIUM RISK)

3. **Module Installation Logic** - Extract installation coordination
   - Target: `internal/modules/installer.go`
   - Impact: Reduces 100+ lines per command
   - Risk: Medium (complex logic, existing alternatives)

4. **Module Listing and Management** - Extract listing/searching
   - Target: `internal/modules/manager.go`
   - Impact: Reduces 50+ lines per command
   - Risk: Medium (formatting complexity)

5. **Parallel Installation** - Extract parallel coordination
   - Target: `internal/modules/parallel.go`
   - Impact: High reusability across components
   - Risk: Medium (concurrency complexity)

### Phase 3: Provider and Dependencies (MEDIUM IMPACT, LOW RISK)

6. **CPAN Provider Builder** - Already extracted ✅
   - Current: `internal/pm/provider_builder.go`
   - Target: Move to `internal/cpan/builder.go`
   - Impact: Provider setup already simplified
   - Risk: Low (existing implementation)

7. **Dependency Management** - Extract cpanfile operations
   - Current: `internal/pm/cpanfile.go` (597 lines)
   - Target: `internal/dependencies/cpanfile.go`
   - Impact: Reusable across components
   - Risk: Low (well-isolated)

### Phase 4: Progress and UI (MEDIUM IMPACT, LOW RISK)

8. **Progress Tracking** - Extract progress patterns
   - Target: `internal/cli/progress/tracker.go`
   - Impact: Consistent progress reporting
   - Risk: Low (UI concern)

9. **Result Formatting** - Extract formatting patterns
   - Target: `internal/cli/progress/formatting.go`
   - Impact: Consistent output formats
   - Risk: Low (presentation logic)

## Interface Design Specifications

### Core Module Management Interface

```go
package modules

type ModuleManager interface {
    List(ctx context.Context, filter ModuleFilter) ([]*Module, error)
    Install(ctx context.Context, modules []string, opts InstallOptions) ([]*InstallResult, error)
    Remove(ctx context.Context, modules []string) ([]*RemoveResult, error)
    Update(ctx context.Context, modules []string) ([]*InstallResult, error)
    Search(ctx context.Context, query string, opts SearchOptions) (*SearchResults, error)
}

type ModuleInstaller interface {
    InstallModule(ctx context.Context, module string, opts InstallOptions) (*InstallResult, error)
    InstallBatch(ctx context.Context, modules []string, opts InstallOptions) ([]*InstallResult, error)
    ValidateInstallation(module string) error
}

type ProgressTracker interface {
    Start(operation string, total int)
    Update(current int, message string)
    Finish(result *OperationResult)
}
```

### Configuration Resolution Interface

```go
package config

type ConfigResolver interface {
    ResolveStringValue(flag, config, default string) string
    ResolveBoolValue(flag, config, default bool) bool
    ResolveStringSlice(flag, config, default []string) []string
    GetEffectiveConfiguration(flagsChanged map[string]bool) (*Config, error)
}
```

## Risk Assessment and Mitigation

### High-Risk Areas

1. **Parallel Installation Logic**
   - Risk: Complex concurrency patterns
   - Mitigation: Comprehensive testing, gradual extraction
   - Timeline: Phase 2

2. **Provider Integration**
   - Risk: Breaking existing provider interfaces
   - Mitigation: Adapter patterns, backward compatibility
   - Timeline: Phase 3

### Medium-Risk Areas

1. **Progress Callback Integration**
   - Risk: UI coupling, callback patterns
   - Mitigation: Interface segregation, mock implementations
   - Timeline: Phase 4

2. **Configuration Precedence**
   - Risk: Configuration ordering changes
   - Mitigation: Explicit testing of precedence rules
   - Timeline: Phase 1

### Low-Risk Areas

1. **Type Definitions**
   - Risk: Minimal (mostly data structures)
   - Mitigation: Standard Go practices
   - Timeline: Phase 1

2. **Result Formatting**
   - Risk: UI changes only
   - Mitigation: Format preservation tests
   - Timeline: Phase 4

## Performance Impact Assessment

### Expected Improvements

1. **Code Reduction:** 60-70% reduction in command.go size (2403 → ~700 lines)
2. **Compilation:** Faster incremental builds due to smaller files
3. **Testing:** More focused, faster unit tests
4. **Memory:** No significant memory impact expected

### Potential Regressions

1. **Import Overhead:** Minimal impact from additional imports
2. **Indirection:** Negligible performance impact from interfaces
3. **Package Loading:** Slight increase in package loading time

### Mitigation Strategies

1. **Benchmark Critical Paths:** Before and after performance testing
2. **Interface Optimization:** Keep interfaces minimal and focused
3. **Package Structure:** Avoid deep package hierarchies

## Test Strategy

### Phase 1 Testing (Foundation)

1. **Interface Compliance Tests**
   - Verify all implementations satisfy interfaces
   - Use testify/suite for consistent patterns

2. **Configuration Resolution Tests**
   - Test precedence ordering
   - Test flag/config/default combinations

### Phase 2 Testing (Module Management)

1. **Installation Integration Tests**
   - End-to-end installation workflows
   - Parallel installation scenarios
   - Error handling and recovery

2. **Module Management Tests**
   - List, search, remove operations
   - Filter and query functionality

### Phase 3 Testing (Provider/Dependencies)

1. **Provider Builder Tests**
   - Builder pattern combinations
   - Configuration integration
   - Error scenarios

2. **Dependency Management Tests**
   - cpanfile operations
   - Snapshot generation/reading
   - Version constraint handling

### Phase 4 Testing (Progress/UI)

1. **Progress Tracking Tests**
   - Progress callback patterns
   - Parallel progress aggregation
   - UI integration tests

2. **Formatting Tests**
   - Output format consistency
   - JSON/table/simple formats
   - Error message formatting

## Success Criteria

### Quantitative Metrics

- **Line Reduction:** 60-70% reduction in command.go (2403 → 700-900 lines)
- **Code Duplication:** <10% duplicated patterns across commands
- **Test Coverage:** >95% for all extracted packages
- **Build Time:** No regression in build performance
- **Module Count:** Enable module management in all 4 PVM components

### Qualitative Metrics

- **Maintainability:** Each package has single, clear responsibility
- **Reusability:** Extracted packages usable across PVM components
- **Testability:** Focused, fast unit tests for each package
- **Documentation:** Clear interfaces with comprehensive examples
- **Backward Compatibility:** All existing functionality preserved

## Recommended Implementation Order

1. **Phase 1 (Weeks 1-2):** Foundation and interfaces
   - Step 1.2: Core interfaces and types
   - Step 1.3: Configuration helpers

2. **Phase 2 (Weeks 3-5):** Module management extraction
   - Step 2.1: Module installer
   - Step 2.2: Module manager and listing
   - Step 2.3: Parallel coordination

3. **Phase 3 (Weeks 6-7):** Provider and dependency management
   - Step 3.1: Provider builder relocation
   - Step 3.2: Configuration management
   - Step 3.3: Dependency management extraction

4. **Phase 4 (Weeks 8-9):** Progress and UI standardization
   - Step 5.1: Progress tracking framework
   - Step 5.2: Result formatting
   - Step 5.3: Integration and testing

5. **Phase 5 (Weeks 10-11):** Integration and optimization
   - Step 6.1: Command updates
   - Step 6.2: Cross-component integration
   - Step 6.3: Performance optimization

## Conclusion

The analysis confirms that `internal/pm/command.go` contains significant extractable functionality with high potential for reuse across PVM components. The existing prototype implementations demonstrate viable extraction patterns, and the risk assessment shows manageable complexity with proper phasing.

**Recommendation:** Proceed with extraction following the phased approach, starting with foundation interfaces and progressing through module management, provider configuration, and UI standardization.

**Key Success Factors:**
- Test-driven extraction with comprehensive coverage
- Incremental approach with frequent validation
- Interface-first design to minimize coupling
- Performance monitoring throughout the process
