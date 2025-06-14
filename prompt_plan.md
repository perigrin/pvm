# Dead Code Cleanup Plan for PVM

## Overview

This plan addresses the extensive dead code identified by deadcode analysis (~4,000-5,000 lines across multiple packages). The approach prioritizes safety, incremental progress, and maintaining system stability.

## Phase 1: Immediate Wins - Completely Unused Systems (Est. 2,000 lines)

### Step 1.1: Remove Cache System
**Target**: `internal/cache/` package - entirely unused advanced caching infrastructure

```
Remove the entire internal/cache/ package which contains unused caching infrastructure:
- compression.go: Compression strategies (NoOp, Gzip, Zstd, Adaptive)
- distributed.go: Redis-based distributed caching with connection pooling
- multilevel.go: Multi-tier caching with LRU/LFU/FIFO eviction policies

Verify no imports exist with: `ag "internal/cache" --go`
Test that removal doesn't break builds with: `make test`
```

### Step 1.2: Remove Unused CLI Framework
**Target**: `internal/cli/` unused components

```
Remove unused CLI framework components:
- internal/cli/error.go: Alternative error handling (unused)
- internal/cli/root.go: Alternative CLI root command (unused)

Keep only the parts actually used by checking imports:
`ag "internal/cli" --go`

Verify the main CLI implementation in cmd/ packages still works.
Test with: `make pvm && ./build/pvm --help`
```

### Step 1.3: Remove Legacy AST Adapters
**Target**: `internal/ast/adapters.go` - 20+ unused adapter functions

```
Remove the legacy parser compatibility layer in internal/ast/adapters.go:
- All NewParserAdapter, ParserCompatibilityAdapter methods
- ConvertLegacyAST, ConvertLegacyTypeAnnotation functions
- WrapLegacyParser and related wrapper functions

These were compatibility shims for old parser interface that are no longer needed.
Verify with: `ag "adapters\." --go` and `ag "NewParserAdapter\|ConvertLegacy" --go`
```

## Phase 2: Legacy Navigation System (Est. 1,500 lines)

### Step 2.1: Remove AST Visitor Pattern
**Target**: `internal/astnav/visitor.go` - unused visitor pattern implementation

```
Remove the comprehensive AST visitor pattern in internal/astnav/visitor.go:
- BaseVisitor with all Visit* methods
- CollectVisitor, TransformVisitor, PrintVisitor implementations
- WalkVisitor, WalkPrint utility functions

Check if any tests or other code depends on these:
`ag "BaseVisitor\|CollectVisitor\|WalkVisitor" --go`

Keep the Navigator in internal/astnav/navigator.go as it may be used.
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
