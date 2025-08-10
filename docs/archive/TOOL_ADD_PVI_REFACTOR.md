# Tool Installation Refactor: From cpanm to PM

## Architectural Improvement

We've refactored `pvm tool add` to use PM (PVM's native module installer) instead of cpanm, eliminating architectural inconsistencies and the fork bomb risk.

## Before (cpanm-based)
```
pvm tool add ack
  ↓
Generate Perl script with system('cpanm', 'App::Ack')
  ↓
PVX executes script in isolated environment
  ↓
Script calls cpanm shim → FORK BOMB RISK
  ↓
Required cpanm shim exclusion workaround
```

## After (PM-based)
```
pvm tool add ack
  ↓
Call PM directly with InstallModulesForPVX()
  ↓
PM handles module installation natively
  ↓
Clean, consistent architecture
```

## Benefits

### ✅ **Architectural Consistency**
- Uses PVM's own module installer (PM)
- Eliminates external dependency on cpanm
- Consistent with rest of PVM ecosystem

### ✅ **Security & Stability**
- No more fork bomb risk from cpanm shims
- No need for cpanm shim exclusion workaround
- Proper error handling through PM's structured results

### ✅ **Better User Experience**
- PM's advanced progress tracking
- Better error messages with structured feedback
- Proper dependency resolution
- Parallel installation support

### ✅ **Feature Consistency**
- Same caching and mirror configuration as `pvm module`
- Same retry logic and timeout handling
- Same test execution policies

## Code Changes

### Removed
- ~80 lines of cpanm script generation
- Complex cpanm output parsing
- cpanm-specific error handling
- Dependency on external cpanm binary

### Added
- Direct PM integration
- Structured error handling with PMXIntegrationResult
- Version constraint support (@version syntax)
- Better progress feedback

## Next Steps

1. Update PVX auto-install to use PM (not cpanm)
2. Remove cpanm shim exclusion workaround
3. Test with various module installations

This brings PVM's tool installation in line with its architectural principles and eliminates the technical debt from the cpanm-based approach.
