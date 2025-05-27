# Performance Profiling Summary

Generated at: 2025-05-27 09:43:57

## Component Performance

| Component | Duration | Memory Allocated | Goroutines |
|-----------|----------|------------------|------------|
| Parser | 93.429708ms | 91931096 bytes | 0 |
| Binder | 89.031375ms | 91949872 bytes | 0 |
| TypeChecker | 92.570416ms | 92720184 bytes | 0 |
| IntegratedPipeline | 776.277292ms | 913677384 bytes | 0 |

## Analysis Commands

To analyze the profiles in detail, use these commands:

### CPU Profiles
```bash
go tool pprof performance-profiles/parser_cpu.prof
```
```bash
go tool pprof performance-profiles/binder_cpu.prof
```
```bash
go tool pprof performance-profiles/typechecker_cpu.prof
```
```bash
go tool pprof performance-profiles/pipeline_cpu.prof
```

### Memory Profiles
```bash
go tool pprof performance-profiles/parser_mem.prof
```
```bash
go tool pprof performance-profiles/binder_mem.prof
```
```bash
go tool pprof performance-profiles/typechecker_mem.prof
```
```bash
go tool pprof performance-profiles/pipeline_mem.prof
```

## Optimization Recommendations

Based on the profiling results:

1. **Highest Duration Component**: Focus optimization efforts here
2. **Memory Usage**: Consider object pooling for high-allocation components
3. **Goroutine Count**: Monitor for potential goroutine leaks

Run this profiler regularly to track optimization progress.
