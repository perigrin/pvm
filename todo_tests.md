# Test Improvement Todo List - COMPLETE ANALYSIS

**Generated:** 2025-06-04 23:30:00
**Current Status:** 82 failing tests out of 3,872 total tests (97.9% pass rate)
**Goal:** Achieve 100% test pass rate

---

## 📊 CURRENT FAILURE BREAKDOWN

**Total Failing Tests: 82**

### By Package:
- **internal/ls**: 6 failures (Language Service)
- **internal/lsp**: 8 failures (LSP Protocol)
- **internal/mcp/embeddings**: 8 failures (MCP Embeddings)
- **internal/parser**: 22 failures (Parser Core)
- **internal/typechecker**: 4 failures (Type Checker)
- **internal/xdg**: 3 failures (XDG Directories)
- **test/e2e**: 31 failures (End-to-End Integration)

---

## 🚨 CRITICAL PRIORITY - LANGUAGE SERVICE (6 failures)

### internal/ls Package Failures:
1. **TestFullWorkflow** - Language service end-to-end workflow
2. **TestConcurrentAccess** - Concurrent access to language service
3. **TestPerformanceTargets** - Performance benchmarks failing
4. **TestCacheEffectiveness** - Caching not working properly
5. **TestMemoryUsage** - Memory usage exceeding limits
6. **TestCacheInvalidation** - Cache invalidation logic broken

**Impact**: Core language server functionality compromised
**Priority**: CRITICAL - Affects editor integration

---

## 🚨 CRITICAL PRIORITY - LSP PROTOCOL (8 failures)

### internal/lsp Package Failures:
1. **TestFindDefinition/Subroutine_definition** - Cannot find subroutine definitions
2. **TestFindDefinition** - Go-to-definition feature broken
3. **TestFindReferences/Variable_references_with_declaration** - Reference finding broken
4. **TestFindReferences/Variable_references_without_declaration** - Reference finding broken
5. **TestFindReferences/Subroutine_references** - Subroutine reference finding broken
6. **TestFindReferences** - All reference finding broken
7. **TestExtractSymbolAtPosition/Subroutine_name** - Symbol extraction broken
8. **TestExtractSymbolAtPosition** - Symbol extraction broken

**Impact**: Essential IDE features (go-to-definition, find-references) not working
**Priority**: CRITICAL - Core development experience broken

---

## 🔥 HIGH PRIORITY - MCP EMBEDDINGS (8 failures)

### internal/mcp/embeddings Package Failures:
1. **TestExtractorWithFixedNodeTypes** - Node type extraction failing
2. **TestExtractor_ExtractFromFile/simple_sub.pl** - Simple subroutine extraction failing
3. **TestExtractor_ExtractFromFile/typed_sub.pl** - Typed subroutine extraction failing
4. **TestExtractor_ExtractFromFile/class_example.pl** - Class extraction failing
5. **TestExtractor_ExtractFromFile/package_example.pl** - Package extraction failing
6. **TestExtractor_ExtractFromFile** - File extraction completely broken
7. **TestExtractor_ConvertToDocuments** - Document conversion failing
8. **TestExtractor_EdgeCases/nested_subs** - Nested subroutine handling broken

**Impact**: AI-assisted development features not working
**Priority**: HIGH - AI features compromised

---

## ⚠️ MEDIUM PRIORITY - PARSER CORE (22 failures)

### internal/parser Package Failures:

#### Constraint Parsing (5 failures):
1. **TestConstraintParsingErrorRecovery/Missing_Constraint_Expression**
2. **TestConstraintParsingErrorRecovery/Invalid_Constraint_Syntax**
3. **TestConstraintParsingErrorRecovery/Malformed_Value_Constraint**
4. **TestConstraintParsingErrorRecovery/Incomplete_Where_Clause**
5. **TestConstraintParsingErrorRecovery**

#### Baseline Testing (7 failures):
6. **TestParser_Baselines/simple_variables**
7. **TestParser_Baselines/type_annotations**
8. **TestParser_Baselines**
9. **TestParser_SpecificBaselines/simple_variables**
10. **TestParser_SpecificBaselines/type_annotations**
11. **TestParser_SpecificBaselines/subroutines**
12. **TestParser_SpecificBaselines/control_structures**
13. **TestParser_SpecificBaselines/complex_expressions**
14. **TestParser_SpecificBaselines**

#### Class/Role Parsing (7 failures):
15. **TestClassDeclarationParsing/basic_class_declarations.json**
16. **TestClassDeclarationParsing/generic_class_declarations.json**
17. **TestClassDeclarationParsing/class_inheritance.json**
18. **TestClassDeclarationParsing/complex_inheritance_constraints.json**
19. **TestClassDeclarationParsing/access_modifiers_visibility.json**
20. **TestClassDeclarationParsing/constructor_destructor_methods.json**
21. **TestClassDeclarationParsing**
22. **TestRoleDeclarationParsing/basic_role_declarations.json**
23. **TestRoleDeclarationParsing/generic_role_declarations.json**
24. **TestRoleDeclarationParsing/role_composition_conflicts.json**
25. **TestRoleDeclarationParsing**
26. **TestComprehensiveClassRoleIntegration**

#### Advanced Parser Features (3 failures):
27. **TestComplexTypeErrorRecovery/Incomplete_union_type**
28. **TestComplexTypeErrorRecovery**
29. **TestEnhancedParserCommonErrorFixes/missing_bracket_fix**
30. **TestEnhancedParserCommonErrorFixes**
31. **TestEnhancedParserSegmentSplitting/simple_statements**
32. **TestEnhancedParserSegmentSplitting**

#### Specialized Features (3 failures):
33. **TestMooseAttributeExtraction**
34. **TestParseMethodTypeAnnotations**
35. **TestParserTestFramework_AST_Validation**
36. **TestTestFramework_Test_Execution**

**Impact**: Advanced Perl parsing features not working
**Priority**: MEDIUM - Core parsing mostly works, advanced features need fixes

---

## ⚠️ MEDIUM PRIORITY - TYPE CHECKER (4 failures)

### internal/typechecker Package Failures:
1. **TestTypeChecker_SpecificBaselines/complex_inference**
2. **TestTypeChecker_SpecificBaselines/function_types**
3. **TestTypeChecker_SpecificBaselines/method_chaining**
4. **TestTypeChecker_SpecificBaselines/conditional_types**
5. **TestTypeChecker_SpecificBaselines**

**Impact**: Advanced type checking features not working
**Priority**: MEDIUM - Basic type checking works, advanced inference broken

---

## ⚠️ LOW PRIORITY - XDG DIRECTORIES (3 failures)

### internal/xdg Package Failures:
1. **TestGetDirs/ExplicitEnvironmentVariables**
2. **TestGetDirs**
3. **TestGetConfigFilePath**

**Impact**: XDG directory resolution not working properly
**Priority**: LOW - Affects config file locations

---

## 🔥 HIGH PRIORITY - END-TO-END INTEGRATION (31 failures)

### test/e2e Package Failures:

#### Comprehensive Integration (3 failures):
1. **TestComprehensiveIntegration_LegacyMigration** (5.24s) - Long-running
2. **TestComprehensiveIntegration_ErrorHandling**
3. **TestComprehensiveIntegration_BackwardCompatibility**

#### Cross-Component Integration (1 failure):
4. **TestCrossComponentIntegration_EndToEnd**

#### LSP Integration (1 failure):
5. **TestLSPIntegration_ErrorHandling**

#### Migration Compatibility (4 failures):
6. **TestMigrationCompatibility_ExistingConfigs**
7. **TestMigrationCompatibility_EnvironmentVariables**
8. **TestMigrationCompatibility_ModuleHandling** (2.93s) - Long-running
9. **TestMigrationCompatibility_UpgradePath**

#### Pooling Integration (2 failures):
10. **TestPoolingIntegration_BackwardCompatibility/legacy_perl**
11. **TestPoolingIntegration_BackwardCompatibility**

#### PSC Integration (3 failures):
12. **TestPSCBasicTypeChecking**
13. **TestPSCStripAnnotations**
14. **TestPSCCompleteWorkflow**

#### System Integration (1 failure):
15. **TestImportSystemPerl**

**Impact**: Complete system integration broken
**Priority**: HIGH - System-wide functionality not working

---

## 📋 ACTION PLAN

### Phase 1: Critical LSP/Language Service (14 failures)
**Goal**: Restore core development experience
1. Fix LSP find-definition functionality
2. Fix LSP find-references functionality
3. Fix language service caching and performance
4. Fix symbol extraction

### Phase 2: High Priority Integration (31 failures)
**Goal**: Restore system-wide functionality
1. Fix end-to-end integration tests
2. Fix PSC workflow integration
3. Fix migration compatibility

### Phase 3: MCP and Parser (30 failures)
**Goal**: Restore advanced features
1. Fix MCP embeddings extraction
2. Fix parser class/role declarations
3. Fix parser constraint parsing
4. Fix parser baseline tests

### Phase 4: Remaining Issues (7 failures)
**Goal**: Complete 100% pass rate
1. Fix type checker advanced features
2. Fix XDG directory resolution

---

---

## 🔴 CRITICAL INFRASTRUCTURE - TREE-SITTER GRAMMAR (181 failures)

### tree-sitter-typed-perl Grammar Test Failures:

**Status**: Critical infrastructure issue affecting all parsing
**Total Failures**: 181 out of 186 corpus tests (97% failure rate)

#### String/Literal Parsing (10 failures):
- `'' strings` - Single quote string parsing broken
- `q() strings` - q() quote operator broken
- `"" strings` - Double quote string parsing broken
- `qq() strings` - qq() quote operator broken
- `quotelike strings - tricky delimeters` - Complex quote delimiters broken
- `Interpolation in "" strings` - String interpolation broken
- `qw() lists` - Quote word lists broken
- `` ` strings` - Backtick command strings broken
- `qx() strings` - qx() command strings broken

#### Expression Parsing (23 failures):
- `do { STMT; }` - do blocks broken
- `eval STRING/BLOCK` - eval expressions broken
- `Blocks that look like hashes` - Hash/block disambiguation broken
- `Anonymous hash` - Anonymous hash creation broken
- `Anonymous Slices` - Array/hash slices broken
- `Slices` - Variable slices broken
- `Keyval Slices` - Key-value slices broken
- `local and dynamically` - local/state declarations broken
- `return` - return statements broken

#### Function Call Parsing (12 failures):
- `Function call (0/1/2 args)` - Function calls broken
- `Filetest operators` - File test operators (-f, -d, etc.) broken
- `ambiguous funcs` - Ambiguous function call parsing broken
- `ambiguous funcs with indirect objects` - Indirect object syntax broken
- `ambiguous funcs - indirect object fakeouts` - Complex indirect object cases broken
- `non-ambiguous indirob handling (for builtins)` - Builtin indirect objects broken

#### Heredoc Parsing (8 failures):
- `Non-quoted heredoc` - Basic heredocs broken
- `Quoted heredocs` - Quoted heredoc delimiters broken
- `Command heredocs` - Command heredocs broken
- `Indented heredocs` - Indented heredoc syntax broken
- `Insane heredocs` - Complex heredoc cases broken
- `heredocs after a statement` - Heredoc positioning broken
- `weird heredoc interpolation` - Heredoc interpolation broken
- `utf8 heredoc delims` - UTF-8 heredoc delimiters broken

#### String Interpolation (11 failures):
- `Fancy indirob interpolation` - Indirect object interpolation broken
- `Array element interpolation` - Array element interpolation broken
- `Hash element interpolation` - Hash element interpolation broken
- `Method calls don't interpolate` - Method call interpolation broken
- `Space skips interpolation` - Whitespace interpolation rules broken
- `Slice interpolation` - Slice interpolation broken
- `Postfix star interpolation` - Postfix operators in interpolation broken
- `Punctuation vars that interpolate` - Special variable interpolation broken
- `Punctuation vars that do not interpolate` - Special variable non-interpolation broken
- `braced vars do not subscript` - Braced variable subscripting broken
- `Nested quotes!` - Nested quote handling broken

#### Autoquote/Bareword Parsing (13 failures):
- `AUTOQUOTED => EXPR` - Hash key autoquoting broken
- `quotelike followed by =>` - Quote operators with fat comma broken
- `hash autoquoting` - Hash key autoquoting broken
- `hash autoquoting for quotelike` - Hash autoquoting with quotes broken
- `indirob autoquoting` - Indirect object autoquoting broken
- `autoquoting keywords` - Keyword autoquoting broken
- `autoquoting postfix` - Postfix autoquoting broken
- `autoquoting lowprec (list-expr)` - Low precedence autoquoting broken
- `autoquoting lowprec (ambiguous_func)` - Ambiguous function autoquoting broken
- `autoquoting else blocks` - else block autoquoting broken
- `autoquote edge cases` - Edge case autoquoting broken

#### Map/Grep/Sort Parsing (6 failures):
- `map - BLOCK form` - map with block form broken
- `map - EXPR form` - map with expression form broken
- `map - goshdarn parens` - map parentheses handling broken
- `map - different LISTs` - map with different list types broken
- `sort - with and without a BLOCK` - sort block/expression forms broken
- `sort SUBNAME` - sort with subroutine name broken

#### POD Documentation (3 failures):
- `POD` - Plain Old Documentation parsing broken
- `not confused by leading whitespace` - POD whitespace handling broken
- `POD can appear anywhere within an expression` - POD positioning broken

#### Regular Expression Parsing (6 failures):
- `qr() strings` - qr() regex quoting broken
- `modifiers whitespace` - Regex modifier whitespace broken
- `Regexp match` - Regex match operators broken
- `transliteration` - tr/y operators broken
- `transliteration - 3 part quotelike insanity` - Complex transliteration broken

#### Operator Parsing (5 failures):
- `EXPR eq EXPR - list/non assoc` - String comparison operator precedence broken
- `EXPR < EXPR - list/non assoc` - Numeric comparison operator precedence broken
- `range ops - nonassoc` - Range operator precedence broken
- `diamond operators` - Diamond operator broken
- `transliteration` - Transliteration operators broken

#### Variable Declaration Edge Cases (multiple failures):
- Variable declarations with complex attributes broken
- Special variable parsing broken ($^X, ${^_arbitrary_VAR}, etc.)
- Unicode variable names broken

**Root Cause**: The tree-sitter-typed-perl grammar has fundamental parsing issues that affect:
1. **String literal parsing** - No strings work correctly
2. **Quote operators** - All quote-like operators broken
3. **Function call disambiguation** - Cannot distinguish function calls from other constructs
4. **Operator precedence** - Many operator precedence rules broken
5. **Complex syntax** - Heredocs, POD, regex, etc. all broken

**Impact**:
- **CRITICAL**: All string literals fail to parse (affects typed variable with string values)
- **CRITICAL**: Basic Perl constructs like function calls and operators don't work
- **CRITICAL**: This explains why our Go parser integration shows missing type expressions
- **CRITICAL**: The grammar is not production-ready despite having typed extension rules

**Priority**: CRITICAL INFRASTRUCTURE - Must be fixed before proceeding with other tests

**Action Required**:
1. **IMMEDIATE**: Focus on Go parser integration issues as workaround
2. **SHORT TERM**: Document and report tree-sitter grammar issues to upstream
3. **LONG TERM**: Contribute fixes to tree-sitter-typed-perl grammar

---

## 🎯 SUCCESS METRICS

- **Current**: 3,790/3,872 tests passing (97.9%)
- **Target**: 3,872/3,872 tests passing (100%)
- **Remaining**: 82 tests to fix
- **Tree-sitter Infrastructure**: 181 corpus test failures (critical dependency issue)

**Estimated Work**: Each phase represents 1-2 days of focused work for experienced developer

**Note**: Tree-sitter grammar issues are documented but should be treated as separate infrastructure work
