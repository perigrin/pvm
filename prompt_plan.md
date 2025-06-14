# Dead Code Cleanup Plan for PVM

## Overview

This plan addresses the extensive dead code identified by deadcode analysis (~4,000-5,000 lines across multiple packages). The approach prioritizes safety, incremental progress, and maintaining system stability.

## Phase 1: Immediate Wins - Completely Unused Systems (Est. 2,000 lines) ✅ COMPLETED

### Step 1.1: Remove Cache System ✅ COMPLETED
**Target**: `internal/cache/` package - entirely unused advanced caching infrastructure

```
✅ COMPLETED: Removed the entire internal/cache/ package which contained unused caching infrastructure:
- compression.go: Compression strategies (NoOp, Gzip, Zstd, Adaptive)
- distributed.go: Redis-based distributed caching with connection pooling
- multilevel.go: Multi-tier caching with LRU/LFU/FIFO eviction policies

✅ Verified no imports existed with: `ag "internal/cache" --go`
✅ Tested that removal doesn't break builds with: `make test` - all 2851 tests pass
✅ Commit: 3cbef29 "Phase 1.1 complete: remove unused cache system"
```

### Step 1.2: Remove Unused CLI Framework ❌ SKIPPED
**Target**: `internal/cli/` unused components

```
❌ SKIPPED: Investigation revealed CLI framework components are actively used, not dead code:
- internal/cli/error.go: Used extensively for error handling and logging
- internal/cli/root.go: Used by all main commands for root command creation
- internal/cli/registry.go: GlobalRegistry used by all cmd/ main functions
- internal/cli/router.go: DetectComponent and CreateRootCommand used by all mains
- internal/cli/symlinks.go: CreateSymlinks and VerifySymlinks actively used

The original prompt plan was incorrect about these being unused.
All CLI framework components are integral to the system.
```

### Step 1.3: Remove Legacy AST Adapters ✅ COMPLETED
**Target**: `internal/ast/adapters.go` - 20+ unused adapter functions

```
✅ COMPLETED: Removed the legacy parser compatibility layer from internal/ast/adapters.go:
- All NewParserAdapter, ParserCompatibilityAdapter methods
- ConvertLegacyAST, ConvertLegacyTypeAnnotation functions
- WrapLegacyParser and related wrapper functions
- Removed ~330 lines of dead compatibility code
- Kept only CreateConcreteAST function which may still be useful

✅ Verified no usage with: `ag "NewParserAdapter\|ConvertLegacy" --go`
✅ Tested removal: all 2851 tests pass, build succeeds
✅ Commit: 2d757ba "Phase 1.3 complete: remove legacy AST adapters"
```

**Phase 1 Results**: Successfully removed ~500+ lines of dead code from cache system and AST adapters. CLI framework investigation prevented unnecessary removal of active code.

## Phase 2: Legacy Navigation System (Est. 1,500 lines)

### Step 2.1: Remove AST Visitor Pattern ✅ COMPLETED
**Target**: `internal/astnav/visitor.go` - unused visitor pattern implementation

```
✅ COMPLETED: Removed the comprehensive AST visitor pattern in internal/astnav/visitor.go:
- BaseVisitor with all Visit* methods for 23 different AST node types
- CollectVisitor, TransformVisitor, PrintVisitor implementations
- WalkVisitor, WalkPrint utility functions
- Removed ~304 lines of unused visitor pattern code

✅ Verified no usage with: `ag "BaseVisitor\|CollectVisitor\|WalkVisitor" --go`
✅ Tested removal: all 2851 tests pass, build succeeds
✅ Kept Navigator in internal/astnav/navigator.go as it is actively used
✅ Commit: 64b00e0 "Phase 2.1 complete: remove unused AST visitor pattern"
```

### Step 2.2: Remove Unused CPAN Integration
**Target**: `internal/cpan/integration.go` - if superseded by newer implementation

```
Audit internal/cpan/integration.go for actual usage:
- Check if Integration, CPANMinus, SystemPerl types are used
- Verify MetaCPANClient, RateLimiter, ModuleCache are needed
- Look for imports and actual instantiation

Only remove if confirmed that newer CPAN implementation exists and works.
Run integration tests: `make test-integration`
```

## Phase 3: Advanced Language Server Features (Est. 800 lines)

### Step 3.1: Move Enhanced LS Features to Experimental
**Target**: Advanced LS features not yet implemented

```
Move these files to internal/experimental/ rather than deleting:
- internal/ls/enhanced_completion.go (~520 lines)
- internal/ls/enhanced_diagnostics.go (~740 lines)
- internal/ls/enhanced_navigation.go (~590 lines)
- internal/ls/incremental.go (~400 lines)

Create internal/experimental/ directory and add README.md explaining these are future features.
Update any imports to point to experimental location.
Add TODO comments linking to roadmap items.
```

### Step 3.2: Clean Advanced Diagnostics
**Target**: `internal/diagnostics/` unused advanced features

```
Audit and clean internal/diagnostics/ package:
- enhanced.go: EnhancedDiagnosticEngine if unused
- integration.go: EnhancedTypeChecker integration layer
- usage_tracker.go: SymbolUsageTracker if not used

Check actual usage with: `ag "EnhancedDiagnosticEngine\|EnhancedTypeChecker" --go`
Keep only what's actually integrated into the main type checking pipeline.
```

## Phase 4: Configuration System Cleanup (Est. 700 lines)

### Step 4.1: Remove Advanced Config Features
**Target**: Unused configuration management features

```
Remove unused configuration features:
- internal/config/interpolation.go: Variable interpolation (~280 lines)
- internal/config/reload.go: Hot reloading system (~140 lines)
- internal/config/tools.go: Advanced config management tools (~430 lines)

Check usage: `ag "InterpolationEngine\|HotReloader\|ConfigManager" --go`
Verify basic config loading still works: `make test` and check config tests pass.
```

### Step 4.2: Remove Unused Config Components
**Target**: Additional config features if unused

```
Audit and potentially remove:
- internal/config/templates.go: Config template system
- internal/config/merger.go: Advanced config merging
- internal/config/watcher.go: File system watching

Only remove if confirmed unused by checking imports and testing.
```

## Safety Measures

### Before Each Step:
```
1. Run: `ag "package_or_function_name" --go` to check usage
2. Run: `make test` to ensure current tests pass
3. Create git commit before changes
4. After removal, run: `make test` to verify nothing breaks
```

### Verification Commands:
```
# Check for broken imports
go build ./...

# Run full test suite
make test

# Verify CLI tools still work
make pvm && ./build/pvm --help
./build/pvm version
```

## Implementation Order

1. **Start with Phase 1** - highest confidence, completely unused systems
2. **Move to Phase 2** - legacy systems likely safe to remove
3. **Be cautious with Phase 3** - move to experimental rather than delete
4. **Phase 4 last** - config system is foundational, test thoroughly

## Expected Impact

- **Lines removed**: ~4,000-5,000 lines of unused code
- **Maintenance reduction**: ~30-40% less code to maintain
- **Build time**: Faster compilation due to fewer files
- **Clarity**: Easier navigation and understanding of codebase
- **Risk**: Low risk due to systematic verification approach

## Rollback Strategy

Each phase should be committed separately, allowing easy rollback:
```bash
# If something breaks after a phase
git revert HEAD    # Revert last commit
make test         # Verify system works again
```
