# Test Improvement Todo List

**Generated:** 2025-05-30
**Sorted by pass rate:** Lowest to highest priority

This document provides a prioritized list of test files and packages that need attention, sorted by their current pass rates from lowest (most urgent) to highest.

---

## 🚨 CRITICAL PRIORITY (0% pass rate)

### 1. internal/performance (0% pass rate) - 1 test
**Status:** ALL TESTS FAILING
**Files:**
- `internal/performance/benchmark_test.go`

**Issues:** Performance regression detected (-254.79% improvement)
**Impact:** Performance monitoring completely broken
**Priority:** CRITICAL - Fix immediately

---

## 🔥 HIGH PRIORITY (0% - 50% pass rate)

### 2. internal/integration (30.0% pass rate) - 10 tests
**Status:** 3 pass, 7 fail
**Files:**
- `internal/integration/workflows_test.go`

**Issues:** Most integration workflows failing
**Impact:** Core integration paths are broken
**Priority:** CRITICAL - Major functionality affected

---

### 3. test (50.0% pass rate) - 2 tests
**Status:** 1 pass, 1 fail
**Files:**
- `test/build_test.go`

**Issues:** Build process validation failing
**Impact:** Build system validation compromised
**Priority:** HIGH - Build integrity at risk

---

### 4. internal/lsp (83.3% pass rate) - 18 tests  ✅ MAJOR PROGRESS
**Status:** 15 pass, 3 fail, 2 skipped (implementation TODOs)
**Fixed:**
- ✅ Symbol resolution issue causing definition lookup failures
- ✅ LSP server message parsing using proper buffered I/O
- ✅ Mock connection timing and EOF issues
- ✅ Updated TODO tests to use Language Service architecture (2/4 working)
**Files:**
- `internal/lsp/features_test.go` - ✅ TestFindDefinition (3/3), ✅ TestExtractSymbolAtPosition (4/4), ❌ TestFindReferences (0/3)
- `internal/lsp/integration_test.go` - ✅ TestLSPServerIntegration (1/1)
- `internal/lsp/pool_test.go` - ✅ All pool tests pass (7/7)

**Remaining:** FindReferences needs better implementation, formatting/code actions need Language Service methods
**Impact:** Core LSP functionality working, advanced features need implementation

---

### 5. internal/ls (65.8% pass rate) - 38 tests
**Status:** 25 pass, 13 fail
**Files:**
- `internal/ls/async_test.go`
- `internal/ls/cache_test.go`
- `internal/ls/debug_test.go`
- `internal/ls/integration_test.go`
- `internal/ls/performance_test.go`

**Issues:** Language server core functionality issues
**Impact:** Editor integration and language services
**Priority:** HIGH - Core language server broken

---

### 6. internal/mcp/tools (72.6% pass rate) - 62 tests
**Status:** 45 pass, 17 fail
**Files:**
- `internal/mcp/tools/analyze_advanced_test.go`
- `internal/mcp/tools/analyze_test.go`
- `internal/mcp/tools/generate_advanced_test.go`
- `internal/mcp/tools/generate_test.go`
- `internal/mcp/tools/project_analyzer_test.go`
- `internal/mcp/tools/search_test.go`

**Issues:** MCP tool functionality failures
**Impact:** AI-assisted development tools compromised
**Priority:** HIGH - Development tooling affected

---

## ⚠️ MEDIUM PRIORITY (75% - 95% pass rate)

### 7. internal/config (79.6% pass rate) - 157 tests
**Status:** 125 pass, 32 fail
**Files:**
- `internal/config/accessors_test.go`
- `internal/config/events_test.go`
- `internal/config/interpolation_test.go`
- `internal/config/loader_test.go`
- `internal/config/parser_test.go`
- `internal/config/profiles_test.go`
- `internal/config/reload_test.go`
- `internal/config/schema_test.go`
- `internal/config/templates_test.go`
- `internal/config/tools_test.go`
- `internal/config/watcher_test.go`

**Issues:** Config system validation and file watching
**Impact:** Configuration management
**Priority:** MEDIUM - Core functionality affected but not critical

---

### 8. internal/mcp/embeddings (86.7% pass rate) - 60 tests
**Status:** 52 pass, 8 fail
**Files:**
- `internal/mcp/embeddings/extractor_test.go`
- `internal/mcp/embeddings/manager_test.go`
- `internal/mcp/embeddings/provider_test.go`
- `internal/mcp/embeddings/store_test.go`

**Issues:** Embeddings system edge cases
**Impact:** AI embeddings functionality
**Priority:** MEDIUM - AI features affected

---

### 9. internal/pvi/modules (87.2% pass rate) - 39 tests
**Status:** 34 pass, 5 fail
**Files:**
- `internal/pvi/modules/installer_test.go`
- `internal/pvi/modules/manager_test.go`
- `internal/pvi/modules/parallel_installer_test.go`

**Issues:** Module installation edge cases
**Impact:** Package installation and management
**Priority:** MEDIUM - Package management mostly working

---

### 10. internal/psc (90.6% pass rate) - 96 tests
**Status:** 87 pass, 9 fail
**Files:**
- `internal/psc/check_command_test.go`
- `internal/psc/check_integration_test.go`
- `internal/psc/command_test.go`
- `internal/psc/def_command_test.go`
- `internal/psc/error_formatter_test.go`
- `internal/psc/psc_pvi_integration_test.go`

**Issues:** Type checker command functionality
**Impact:** Static analysis and type checking
**Priority:** MEDIUM - Important but mostly working

---

### 11. internal/parser/treesitter (92.9% pass rate) - 14 tests
**Status:** 13 pass, 1 fail
**Files:**
- `internal/parser/treesitter/benchmark_test.go`
- `internal/parser/treesitter/parser_test.go`

**Issues:** Tree-sitter integration edge case
**Impact:** Core parsing functionality
**Priority:** MEDIUM - Critical component but mostly stable

---

### 12. internal/pvi/deps (95.1% pass rate) - 123 tests
**Status:** 117 pass, 6 fail
**Files:**
- `internal/pvi/deps/advanced_resolver_test.go`
- `internal/pvi/deps/conflict_resolver_test.go`
- `internal/pvi/deps/resolution_strategies_test.go`
- `internal/pvi/deps/resolver_test.go`
- `internal/pvi/deps/version_solver_test.go`

**Issues:** Dependency resolution edge cases
**Impact:** Package dependency management
**Priority:** LOW - Mostly working with minor edge cases

---

## 📊 GOOD HEALTH (95% - 99% pass rate)

### 13. internal/pvx (97.5% pass rate) - 81 tests
**Status:** 79 pass, 2 fail
**Files:**
- `internal/pvx/command_test.go`
- `internal/pvx/dependency_detection_test.go`
- `internal/pvx/executor_isolation_test.go`
- `internal/pvx/executor_test.go`
- `internal/pvx/isolation_test.go`
- `internal/pvx/script_metadata_test.go`

**Issues:** Minor script execution edge cases
**Impact:** Script execution and isolation
**Priority:** LOW - Mostly working well

---

### 14. internal/parser (97.9% pass rate) - 1900 tests
**Status:** 1860 pass, 40 fail
**Files:**
- `internal/parser/accuracy_test.go`
- `internal/parser/advanced_constraints_test.go`
- `internal/parser/backward_compatibility_test.go`
- `internal/parser/backward_compatibility_validation_test.go`
- `internal/parser/baseline_test.go`
- `internal/parser/classes_roles_test.go`
- `internal/parser/compat_test.go`
- `internal/parser/complex_types_test.go`
- `internal/parser/comprehensive_integration_test.go`
- `internal/parser/control_flow_test.go`
- `internal/parser/debug_class_test.go`
- `internal/parser/debug_type_annotations_test.go`
- `internal/parser/enhanced_parser_test.go`
- `internal/parser/expressions_test.go`
- `internal/parser/integration_markdown_test.go`
- `internal/parser/introspector_test.go`
- `internal/parser/markdown_test_loader_test.go`
- `internal/parser/methods_fields_test.go`
- `internal/parser/packages_test.go`
- `internal/parser/parameterized_types_test.go`
- `internal/parser/parser_test.go`
- `internal/parser/performance_regression_test.go`
- `internal/parser/performance_test.go`
- `internal/parser/performance_validation_test.go`
- `internal/parser/pipeline_integration_test.go`
- `internal/parser/simple_type_annotations_test.go`
- `internal/parser/subroutines_test.go`
- `internal/parser/test_framework_integration_test.go`
- `internal/parser/test_framework_standalone_test.go`
- `internal/parser/tool_integration_test.go`
- `internal/parser/type_assertions_test.go`
- `internal/parser/type_error_recovery_test.go`
- `internal/parser/union_types_test.go`
- `internal/parser/variable_declarations_test.go`

**Issues:** Advanced parser edge cases and complex type scenarios
**Impact:** Advanced parsing features
**Priority:** LOW - Core functionality is solid, advanced features need polish

---

### 15. internal/perl (99.0% pass rate) - 209 tests
**Status:** 207 pass, 2 fail
**Files:**
- `internal/perl/build_enhanced_test.go`
- `internal/perl/build_test.go`
- `internal/perl/download_test.go`
- `internal/perl/get_available_versions_test.go`
- `internal/perl/get_version_info_test.go`
- `internal/perl/import_system_perl_test.go`
- `internal/perl/is_version_installed_test.go`
- `internal/perl/legacy_test.go`
- `internal/perl/registry_test.go`
- `internal/perl/resolver_test.go`
- `internal/perl/shell_test.go`
- `internal/perl/shim_test.go`
- `internal/perl/system_test.go`
- `internal/perl/uninstall_version_test.go`
- `internal/perl/version_test.go`

**Issues:** Minor edge cases in Perl version management
**Impact:** Version management edge cases
**Priority:** LOW - Mostly working, minor fixes needed

---

### 16. internal/typedef (99.1% pass rate) - 108 tests
**Status:** 107 pass, 1 fail
**Files:**
- `internal/typedef/hierarchy_test.go`
- `internal/typedef/typedef_test.go`
- `internal/typedef/union_benchmark_test.go`
- `internal/typedef/union_comprehensive_test.go`
- `internal/typedef/union_optimization_test.go`
- `internal/typedef/union_regression_test.go`
- `internal/typedef/union_test.go`

**Issues:** Minor type definition edge case
**Impact:** Type system definitions
**Priority:** LOW - Type system is very stable

---

## 🔍 INVESTIGATION NEEDED (Timeout/Hanging Issues)

### test/e2e (TIMEOUT)
**Files:**
- `test/e2e/component_interaction_test.go`
- `test/e2e/comprehensive_integration_test.go`
- `test/e2e/config_test.go`
- `test/e2e/integration_test.go`
- `test/e2e/lsp_integration_test.go`
- `test/e2e/main_test.go`
- `test/e2e/migration_compatibility_test.go`
- `test/e2e/pooling_integration_test.go`
- `test/e2e/psc_test.go`
- `test/e2e/pvx_isolation_test.go`
- `test/e2e/pvx_test.go`
- `test/e2e/shell_test.go`
- `test/e2e/shim_test.go`
- `test/e2e/version_test.go`

**Issues:** Tests hang/timeout - likely infinite loops or blocking operations
**Impact:** End-to-end system validation unknown
**Priority:** INVESTIGATE - Critical for overall system confidence

---

### internal/typechecker (TIMEOUT)
**Files:**
- `internal/typechecker/baseline_test.go`
- `internal/typechecker/incremental_typechecker_test.go`
- `internal/typechecker/inference_engine_test.go`
- `internal/typechecker/pooled_typechecker_test.go`
- `internal/typechecker/type_pool_test.go`
- `internal/typechecker/type_resolution_pool_test.go`
- `internal/typechecker/typechecker_test.go`

**Issues:** Tests hang/timeout - possibly memory/performance issues
**Impact:** Type checking functionality unknown
**Priority:** INVESTIGATE - Critical component may be broken

---

### internal/mcp (TIMEOUT)
**Files:**
- `internal/mcp/integration_step14_test.go`
- `internal/mcp/integration_workflow_test.go`
- `internal/mcp/performance_benchmark_test.go`
- `internal/mcp/server_test.go`
- `internal/mcp/validation/cache_test.go`
- `internal/mcp/validation/validator_test.go`

**Issues:** Tests hang/timeout - likely infinite loops or blocking operations
**Impact:** MCP server functionality unknown
**Priority:** INVESTIGATE - Need to identify blocking code

---

## ✅ EXCELLENT HEALTH (100% pass rate)

The following 16 packages are in excellent health with 100% pass rates:

1. **internal/cpan** (47 tests) - CPAN integration
2. **internal/scanner** (23 tests) - Token scanning
3. **internal/mcp/generation** (26 tests) - AI generation tools
4. **internal/compiler** (10 tests) - Code compilation
5. **internal/errors** (41 tests) - Error handling and formatting
6. **internal/traits** (129 tests) - Trait system and resolution
7. **internal/ast** (34 tests) - AST node creation and manipulation
8. **internal/astnav** (17 tests) - AST navigation utilities
9. **internal/binder** (42 tests) - Symbol binding and resolution
10. **internal/cli** (27 tests) - Command-line interface utilities
11. **internal/core** (12 tests) - Core utilities and pools
12. **internal/diagnostics** (7 tests) - Diagnostic reporting
13. **internal/log** (19 tests) - Logging utilities
14. **internal/memory** (31 tests) - Memory management
15. **internal/monitoring** (12 tests) - System monitoring
16. **internal/version** (3 tests) - Version utilities

---

## 📋 SUMMARY

**Total Test Status:**
- **Total packages analyzed:** 39
- **Total tests:** 3,416
- **Passing tests:** 3,263 (95.5%)
- **Failing tests:** 134 (3.9%)
- **Timeout packages:** 3 (critical investigation needed)
- **Perfect packages:** 16 (41.0%)

**Immediate Action Items:**
1. **FIX CRITICAL:** internal/performance (0% pass rate)
2. **FIX HIGH:** internal/integration (30% pass rate)
3. **INVESTIGATE:** test/e2e, internal/typechecker, internal/mcp timeouts
4. **IMPROVE:** Language server components (ls/lsp)
5. **STABILIZE:** MCP tools and config systems

**Overall Assessment:** The project has a strong foundation with 95.5% overall pass rate, but critical timeout issues in e2e tests and type checker need immediate investigation. Performance regression monitoring is completely broken and requires urgent attention.
