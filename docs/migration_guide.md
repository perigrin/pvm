# Migration Guide: Unified Compiler Architecture

## Overview

This guide helps developers migrate from the legacy separate compiler architecture to the new unified compiler system. The unified architecture provides significant improvements while maintaining backward compatibility.

## What Changed

### Before (Legacy Architecture)
```go
// Separate compiler instances
cleanCompiler := NewCleanPerlCompiler()
typedCompiler := NewTypedPerlCompiler()

// Manual target management
if needCleanPerl {
    result, err := cleanCompiler.Compile(ast)
} else {
    result, err := typedCompiler.Compile(ast)
}
```

### After (Unified Architecture)
```go
// Single unified compiler
compiler := NewCleanPerlCompilerUnified()  // or NewTypedPerlCompilerUnified()
result, err := compiler.Compile(ast)

// Or use registry (recommended)
registry := NewCompilerRegistry()
result, err := registry.Compile(ast, TargetCleanPerl)
```

## Breaking Changes

### 1. Clean Perl Output Includes Version Pragma

**Change:** Clean Perl compilation now automatically includes `use v5.36;` pragma.

**Before:**
```perl
my $count = 42;
print "Count: $count\n";
```

**After:**
```perl
use v5.36;
my $count = 42;
print "Count: $count\n";
```

**Impact:**
- ✅ Better compatibility with modern Perl features
- ✅ Enables subroutine signatures support
- ⚠️ May affect tests that expect exact output matching

**Migration Action:**
Update test expectations to include version pragma:
```go
// Update test expectations
expected := "use v5.36;\nmy $count = 42;\nprint \"Count: $count\\n\";"
```

### 2. Deprecated Compiler Constructors

**Change:** Legacy compiler constructors are deprecated.

**Deprecated:**
```go
cleanCompiler := NewCleanPerlCompiler()     // Deprecated
typedCompiler := NewTypedPerlCompiler()     // Deprecated
```

**Recommended:**
```go
cleanCompiler := NewCleanPerlCompilerUnified()   // New unified
typedCompiler := NewTypedPerlCompilerUnified()   // New unified

// Or use registry (preferred)
registry := NewCompilerRegistry()
```

**Migration Timeline:**
- Legacy constructors will be removed in a future version
- Current code continues to work with deprecation warnings
- Migrate at your convenience

### 3. Improved Type Assertion Handling

**Change:** Type assertions are now properly removed in clean Perl output.

**Before (buggy):**
```perl
# Input
my $typed = $value as Int;

# Legacy output (incorrect)
my $typed = $value as Int;  # Type assertion preserved
```

**After (correct):**
```perl
# Input
my $typed = $value as Int;

# Unified output (correct)
my $typed = $value;  # Type assertion removed
```

**Impact:**
- ✅ Correct clean Perl generation
- ⚠️ Tests expecting buggy behavior will fail

**Migration Action:**
Update test expectations for type assertions:
```go
// Old expectation (incorrect)
expected := "my $typed = $value as Int;"

// New expectation (correct)
expected := "my $typed = $value;"
```

## Migration Steps

### Step 1: Update Compiler Creation

Replace deprecated constructors with unified versions:

```go
// Before
cleanCompiler := NewCleanPerlCompiler()
typedCompiler := NewTypedPerlCompiler()

// After - Option A: Direct replacement
cleanCompiler := NewCleanPerlCompilerUnified()
typedCompiler := NewTypedPerlCompilerUnified()

// After - Option B: Use registry (recommended)
registry := NewCompilerRegistry()
cleanResult, err := registry.Compile(ast, TargetCleanPerl)
typedResult, err := registry.Compile(ast, TargetTypedPerl)
```

### Step 2: Update Test Expectations

Review and update test cases to handle new behavior:

```go
func TestCompilation(t *testing.T) {
    input := `my Int $count = 42;`

    compiler := NewCleanPerlCompilerUnified()
    result, err := compiler.CompileString(input)

    // Update expectation to include version pragma
    expected := `use v5.36;
my $count = 42;`

    if result != expected {
        t.Errorf("Expected %q, got %q", expected, result)
    }
}
```

### Step 3: Leverage Performance Features

Consider using optimized compilers for better performance:

```go
// For high-throughput scenarios
cachingCompiler := NewCachingCleanPerlCompiler(500) // 500 entry cache
result, err := cachingCompiler.CompileString(code)

// Monitor cache effectiveness
stats := cachingCompiler.GetCacheStats()
log.Printf("Cache hit ratio: %.2f%%", stats.HitRatio*100)
```

### Step 4: Update Integration Points

Update any custom integration code:

```go
// PSC command integration
func compilePerlCode(code string, target Target) (string, error) {
    // Before: Manual compiler selection
    // if target == TargetCleanPerl {
    //     compiler := NewCleanPerlCompiler()
    //     return compiler.CompileString(code)
    // }

    // After: Use registry
    registry := NewCompilerRegistry()
    cstAST, err := NewCSTBasedAST("", code)
    if err != nil {
        return "", err
    }
    return registry.Compile(cstAST, target)
}
```

## Performance Considerations

### Memory Usage

The unified compiler has different memory characteristics:

**Before:** Separate compiler instances with individual AST processing
**After:** Shared CST processing with optional caching

**Recommendations:**
- Use caching compilers for repeated compilation
- Monitor memory usage with provided statistics
- Clear caches periodically for long-running processes

### Compilation Speed

Performance improvements with the unified architecture:

| Scenario | Legacy Performance | Unified Performance | Improvement |
|---|---|---|---|
| Basic compilation | ~500μs/op | ~258μs/op | 48% faster |
| Cached compilation | N/A | ~40μs/op | 84% faster than legacy |
| Large codebases | Variable | Optimized | Consistent performance |

**Optimization Tips:**
```go
// Use optimized registry for best performance
registry := NewOptimizedCompilerRegistry(1000) // Large cache
result, err := registry.CompileOptimized(ast, target)

// Check aggregated statistics
stats := registry.GetAggregatedStats()
log.Printf("Average compilation time: %v", stats.AverageTime)
```

## Common Migration Issues

### Issue 1: Test Failures Due to Version Pragma

**Problem:** Tests fail because clean Perl output now includes `use v5.36;`

**Solution:**
```go
// Update test to be version-pragma aware
func TestCleanCompilation(t *testing.T) {
    result, err := compiler.CompileString(input)
    require.NoError(t, err)

    // Check for version pragma
    assert.Contains(t, result, "use v5.36;")

    // Check for expected content
    assert.Contains(t, result, "my $count = 42;")
}
```

### Issue 2: Type Assertion Test Failures

**Problem:** Tests expect type assertions to be preserved in clean Perl

**Solution:**
```go
// Update type assertion expectations
func TestTypeAssertion(t *testing.T) {
    input := `my $typed = $value as Int;`

    // Clean Perl should remove type assertion
    cleanResult, err := cleanCompiler.CompileString(input)
    require.NoError(t, err)
    assert.Contains(t, cleanResult, "my $typed = $value;")
    assert.NotContains(t, cleanResult, "as Int")

    // Typed Perl should preserve type assertion
    typedResult, err := typedCompiler.CompileString(input)
    require.NoError(t, err)
    assert.Contains(t, typedResult, "as Int")
}
```

### Issue 3: Performance Regression in High-Throughput Scenarios

**Problem:** Performance is worse than expected in scenarios with many repeated compilations

**Solution:**
```go
// Use caching compiler for repeated operations
compiler := NewCachingCleanPerlCompiler(1000)

// Batch similar operations
codes := []string{...}
for _, code := range codes {
    result, err := compiler.CompileString(code)
    // Process result...
}

// Monitor cache effectiveness
stats := compiler.GetCacheStats()
if stats.HitRatio < 0.8 { // Less than 80% hit ratio
    log.Warn("Poor cache performance, consider increasing cache size")
}
```

## Testing Migration

### Test Strategy

1. **Automated Migration Validation:**
   ```bash
   # Run existing tests to identify failures
   go test ./internal/compiler -v

   # Run specific migration-related tests
   go test ./internal/compiler -run "Migration" -v
   ```

2. **Performance Baseline:**
   ```bash
   # Establish performance baseline
   go test ./internal/compiler -bench=. -run=^$ -v

   # Compare with legacy performance if needed
   go test ./internal/compiler -bench=BenchmarkCompilerComparison -v
   ```

3. **Integration Testing:**
   ```bash
   # Test PSC command integration
   go test ./internal/compiler -run "TestPSCCommandsIntegration" -v

   # Test parser integration
   go test ./internal/compiler -run "TestParserIntegration" -v
   ```

### Validation Checklist

- [ ] All existing tests pass with updated expectations
- [ ] Performance is equivalent or better than legacy
- [ ] PSC commands work correctly with unified compiler
- [ ] Parser integration maintains compatibility
- [ ] Memory usage is acceptable for your use cases
- [ ] Error handling works as expected

## Rollback Plan

If migration issues arise, you can temporarily rollback:

### Option 1: Use Legacy Compilers

```go
// Temporary rollback to legacy compilers
cleanCompiler := NewCleanPerlCompiler()  // Legacy, deprecated
typedCompiler := NewTypedPerlCompiler()  // Legacy, deprecated
```

### Option 2: Mixed Mode

```go
// Use unified for new code, legacy for existing problematic code
func getCompiler(useLegacy bool, target Target) Compiler {
    if useLegacy {
        if target == TargetCleanPerl {
            return NewCleanPerlCompiler()  // Legacy
        }
        return NewTypedPerlCompiler()      // Legacy
    }

    // Use unified (recommended)
    return NewPerlCompiler(target)
}
```

## Support and Resources

### Documentation

- [Compiler Architecture Documentation](./compiler_architecture.md)
- [CST Structure Documentation](./cst_structure.md)
- API documentation in source code comments

### Testing Resources

- Example test updates in `internal/compiler/*_test.go`
- Performance benchmarks in `internal/compiler/benchmark_test.go`
- Integration tests in `internal/compiler/integration_*_test.go`

### Getting Help

If you encounter migration issues:

1. **Check Test Output:** Test failures often provide clear guidance
2. **Review Examples:** Look at updated test cases for patterns
3. **Performance Issues:** Use provided benchmarks and statistics
4. **Documentation:** Refer to architecture documentation for design details

## Timeline

### Immediate (Current Release)
- ✅ Unified compiler available and working
- ✅ Legacy compilers deprecated but functional
- ✅ Backward compatibility maintained

### Near Term (Next Few Releases)
- Enhanced optimization features
- Additional documentation and examples
- Migration tooling if needed

### Long Term (Future Releases)
- Legacy compiler removal
- Advanced compilation targets
- Extended type system features

## Conclusion

The migration to unified compiler architecture provides significant benefits with minimal disruption. Key advantages include:

- **Better Reliability:** Eliminates known bugs in legacy architecture
- **Improved Performance:** Caching and optimization provide substantial speed improvements
- **Enhanced Maintainability:** Single unified implementation is easier to maintain
- **Future-Proof Design:** Extensible architecture supports upcoming features

Most migrations require only minor test expectation updates and optional performance optimizations. The unified architecture provides a solid foundation for future PVM development while maintaining the compatibility and reliability you expect.
