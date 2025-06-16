# GitHub Actions Workflow Troubleshooting

This document helps debug and resolve common issues with the PVM release workflow.

## Common Workflow Failures

### 1. Tree-sitter Build Failures

**Symptom**: Build fails during tree-sitter compilation with CGO errors
**Platforms Affected**: Usually Linux ARM64, sometimes Windows
**Solution**: PSC building is disabled on problematic platforms via `build_psc: false`

**Debug Steps**:
```bash
# Check if tree-sitter builds locally
make tree-sitter
# Verify CGO is working
go env CGO_ENABLED
```

### 2. Cross-compilation Issues

**Symptom**: `error: C compiler "gcc" not found` or similar
**Platforms Affected**: Linux ARM64 cross-compilation
**Root Cause**: Missing cross-compilation toolchain

**Fixed By**:
- Installing `gcc-aarch64-linux-gnu` in workflow
- Setting `CC=aarch64-linux-gnu-gcc` for ARM64 builds
- Proper CGO_ENABLED environment handling

### 3. Windows Archive Creation Failures

**Symptom**: PowerShell archiving commands fail
**Original Issue**:
```bash
powershell -Command "Compress-Archive -Path ..."
```

**Fixed With**:
```bash
7z a "${ARCHIVE}" $FILES
```

**Why**: 7z is pre-installed on GitHub runners and more reliable than PowerShell for this use case.

### 4. Binary Testing Failures

**Symptom**: Cross-compiled binaries fail to run during testing
**Root Cause**: Trying to execute ARM64 binaries on AMD64 runners

**Solution**: Platform-specific testing logic that only tests native binaries:
```yaml
case "${{ matrix.goos }}-${{ matrix.goarch }}" in
  "linux-amd64")
    if [ "${{ runner.os }}" == "Linux" ]; then
      # Test only on matching platform
    fi
    ;;
esac
```

### 5. Environment Variable Issues

**Symptom**: CGO_ENABLED not set correctly
**Original Problem**:
```yaml
CGO_ENABLED: ${{ matrix.build_psc && '1' || '0' }}
```

**Fixed With**:
```bash
if [ "${{ matrix.build_psc }}" == "true" ]; then
  export CGO_ENABLED=1
else
  export CGO_ENABLED=0
fi
```

## Debugging Workflow Runs

### 1. Check Action Logs

1. Go to GitHub repository → **Actions** tab
2. Click on the failed workflow run
3. Click on the failed job (e.g., "Build linux-amd64")
4. Expand the failed step to see detailed logs

### 2. Download Build Artifacts

Even failed builds often upload artifacts:
1. Go to the workflow run page
2. Scroll to **Artifacts** section
3. Download available artifacts to inspect build results

### 3. Local Reproduction

Reproduce workflow steps locally:
```bash
# Set up same environment
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=1

# Run same commands as workflow
make tree-sitter
go build -ldflags="-s -w" -o build/pvm-linux-amd64 ./cmd/pvm
```

### 4. Matrix Strategy Debugging

The workflow uses a build matrix. To debug specific combinations:
```yaml
strategy:
  fail-fast: false  # Don't stop other builds if one fails
  matrix:
    include:
      - goos: linux
        goarch: amd64
        build_psc: true
        # Add debug: true for extra logging
```

## Platform-Specific Issues

### Linux ARM64
- **Issue**: Cross-compilation complexity
- **Solution**: Uses `gcc-aarch64-linux-gnu` cross-compiler
- **PSC**: Disabled due to tree-sitter CGO complexity
- **Testing**: Skipped (can't run ARM64 on AMD64 runners)

### macOS (Darwin)
- **Issue**: Xcode tools requirement
- **Solution**: `xcode-select --install` in workflow
- **PSC**: Enabled on both Intel and Apple Silicon
- **Testing**: Works on both architectures

### Windows
- **Issue**: CGO complexity, archiving format differences
- **Solution**:
  - PSC disabled (no tree-sitter)
  - 7z archiving instead of PowerShell
  - Proper .exe file handling
- **Testing**: Uses Windows-specific paths and commands

## Release Notes Generation Issues

### Git History Problems

**Symptom**: Release notes empty or missing commits
**Root Cause**: Shallow checkout or missing git history

**Solution**: Workflow uses `fetch-depth: 0` to get full history:
```yaml
- name: Checkout code
  uses: actions/checkout@v5
  with:
    fetch-depth: 0  # Required for changelog generation
```

### Tag Comparison Issues

**Symptom**: "No previous tag found" or comparison errors
**Debug**: Check git tag sorting:
```bash
git tag --sort=-version:refname | head -5
```

## Artifact Upload Issues

### Missing Artifacts

**Symptom**: No artifacts uploaded despite successful build
**Root Cause**: File path mismatches

**Solution**: Verify file paths match upload patterns:
```yaml
- name: Upload build artifacts
  uses: actions/upload-artifact@v5
  with:
    name: release-${{ matrix.goos }}-${{ matrix.goarch }}
    path: build/release/*.tar.gz  # Must match actual files
```

### Large Artifact Warnings

**Symptom**: Warnings about large artifacts
**Mitigation**:
- Use compressed archives
- Set appropriate retention days
- Clean up intermediate files

## Workflow Performance Issues

### Slow Builds

**Causes**:
1. Tree-sitter compilation time
2. Multiple matrix builds running sequentially
3. Large dependency downloads

**Solutions**:
1. Enable matrix parallelization (`fail-fast: false`)
2. Cache dependencies where possible
3. Use pre-built tree-sitter when available

### Resource Limits

**GitHub Actions Limits**:
- 6 hours per job maximum
- 2000 minutes/month for private repos
- Concurrent job limits

**Mitigation**:
- Optimize build steps
- Use matrix strategy efficiently
- Consider self-hosted runners for heavy workloads

## Testing Locally

To test the workflow logic locally:

```bash
# Simulate workflow environment
export GITHUB_REF=refs/tags/v1.0.0-rc4
export GITHUB_EVENT_NAME=push

# Test version extraction
TAG=${GITHUB_REF#refs/tags/}
VERSION=${TAG#v}
echo "Tag: $TAG, Version: $VERSION"

# Test build commands
mkdir -p build/release
go build -ldflags="-s -w" -o build/release/pvm-test ./cmd/pvm
```

## Getting Help

If issues persist:

1. **Check existing issues**: Look for similar problems in GitHub Issues
2. **Create detailed bug report**: Include:
   - Workflow run URL
   - Full error logs
   - Platform/architecture affected
   - Steps to reproduce
3. **Test locally**: Try reproducing the issue outside GitHub Actions
4. **Simplify**: Create minimal test case that demonstrates the problem

## Workflow Improvements

For future enhancements:
- Add more comprehensive testing
- Implement artifact signing
- Add performance benchmarking
- Create notification systems
- Implement rollback mechanisms
