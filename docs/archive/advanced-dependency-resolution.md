# Advanced Dependency Resolution in PM

The PVM Installer (PM) includes an advanced dependency resolver that provides sophisticated conflict resolution and optimization strategies for complex dependency scenarios.

## Overview

The advanced resolver extends the basic dependency resolution with:

- **Conflict Resolution Strategies**: Multiple approaches to handle version conflicts
- **Optimization Strategies**: Different algorithms to optimize the dependency tree
- **Parallel Resolution**: Concurrent processing for faster resolution
- **Version Preferences**: Support for locked and preferred versions
- **Comprehensive Metrics**: Detailed performance and resolution metrics

## Usage

Enable the advanced resolver using the `--advanced` flag with the `pm deps` command:

```bash
pm deps Module::Name --advanced
```

## Conflict Resolution Strategies

When multiple modules require different versions of the same dependency, the resolver can apply different strategies:

### Latest Compatible (Default)
```bash
pm deps Module::Name --advanced --conflict-strategy latest
```
Selects the latest version that satisfies all constraints. This is often the best choice for getting the most up-to-date features while maintaining compatibility.

### Minimal Version
```bash
pm deps Module::Name --advanced --conflict-strategy minimal
```
Selects the minimal version that satisfies all constraints. Useful for conservative deployments where stability is prioritized over features.

### Prefer Existing
```bash
pm deps Module::Name --advanced --conflict-strategy prefer-existing
```
Prefers already-resolved versions when possible. This helps maintain consistency across the dependency tree.

### Fail Fast
```bash
pm deps Module::Name --advanced --conflict-strategy fail-fast
```
Fails immediately on any conflict without attempting resolution. Useful for CI/CD pipelines where conflicts should be manually resolved.

## Optimization Strategies

The resolver can optimize the dependency tree using different approaches:

### Shared Dependencies (Default)
```bash
pm deps Module::Name --advanced --optimization shared
```
Maximizes sharing of common dependencies across the tree. This reduces duplication and installation size.

### Minimal Tree
```bash
pm deps Module::Name --advanced --optimization minimal-tree
```
Minimizes the total number of dependencies. Useful for lightweight deployments.

### Parallel Processing
```bash
pm deps Module::Name --advanced --optimization parallel --parallel-workers 8
```
Uses concurrent processing for faster resolution. The number of workers can be customized.

### No Optimization
```bash
pm deps Module::Name --advanced --optimization none
```
Performs basic resolution without any optimization. Useful for debugging or when optimization causes issues.

## Advanced Features

### Version Locking

You can lock specific module versions in your configuration:

```toml
[pm.locked_versions]
"DBI" = "1.643"
"DBD::SQLite" = "1.70"
```

### Preferred Versions

Specify preferred versions that will be used when multiple versions satisfy constraints:

```toml
[pm.preferred_versions]
"Test::More" = "1.302190"
"Moose" = "2.2206"
```

### Excluded Versions

Exclude specific versions that are known to have issues:

```toml
[pm.excluded_versions]
"Module::Build" = ["0.4229", "0.4230"]
```

## Performance Metrics

Enable metrics display to see resolution performance:

```bash
pm deps Module::Name --advanced --metrics
```

This shows:
- Total modules resolved
- Number of conflicts found and resolved
- Resolution time
- Cache hit/miss statistics
- Parallel processing metrics

## Examples

### Complex Module with Known Conflicts
```bash
pm deps Catalyst::Runtime --advanced \
  --conflict-strategy latest \
  --optimization shared \
  --include-test \
  --metrics
```

### Conservative Installation
```bash
pm deps App::cpanminus --advanced \
  --conflict-strategy minimal \
  --optimization minimal-tree \
  --max-depth 3
```

### Fast Parallel Resolution
```bash
pm deps Mojolicious --advanced \
  --optimization parallel \
  --parallel-workers 12 \
  --verbose
```

### Debugging Dependency Issues
```bash
pm deps problematic-module --advanced \
  --conflict-strategy fail-fast \
  --verbose \
  --metrics
```

## Integration with Installation

The advanced resolver is automatically used when installing modules with the `--advanced` flag:

```bash
pm install Module::Name --advanced \
  --conflict-strategy latest \
  --optimization shared
```

## Best Practices

1. **Start with defaults**: The default strategies work well for most cases
2. **Use fail-fast in CI**: Detect conflicts early in automated pipelines
3. **Profile with metrics**: Use `--metrics` to understand resolution performance
4. **Lock critical versions**: Use version locking for production dependencies
5. **Test optimization strategies**: Different strategies may work better for different module sets

## Troubleshooting

### Resolution Takes Too Long
- Try `--optimization parallel` with more workers
- Use `--max-depth` to limit traversal depth
- Enable caching if not already enabled

### Too Many Conflicts
- Try `--conflict-strategy minimal` for conservative resolution
- Review excluded versions configuration
- Consider locking known-good version combinations

### Memory Usage Too High
- Reduce `--parallel-workers` count
- Use `--optimization minimal-tree`
- Enable caching to avoid redundant resolution

### Unexpected Versions Selected
- Check locked and preferred versions configuration
- Use `--verbose` to see resolution decisions
- Try `--conflict-strategy prefer-existing`
