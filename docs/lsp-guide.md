# PVM Language Server Protocol (LSP) Guide

## Introduction

PVM provides a comprehensive Language Server Protocol implementation that delivers modern IDE features for Perl development. With symbol-aware analysis and enhanced performance, PVM's LSP offers TypeScript-quality development experience for Perl projects.

## Features Overview

### Core LSP Features ✅
- **Goto Definition**: Jump to symbol declarations with accuracy
- **Find References**: Locate all uses of symbols across files
- **Hover Information**: Rich symbol details and type information
- **Code Completion**: Context-aware suggestions with symbol filtering
- **Document Symbols**: Hierarchical outline view
- **Workspace Symbols**: Cross-file symbol search
- **Rename Symbol**: Safe renaming with scope awareness
- **Diagnostics**: Real-time error detection with enhanced messages

### Enhanced Features ✅
- **Symbol-aware analysis**: Leverages binder phase for accuracy
- **Cross-module resolution**: Accurate navigation across file boundaries
- **Type-aware completion**: Completions based on symbol types and context
- **Performance optimization**: Multi-level caching and async processing
- **Enhanced diagnostics**: Context-aware error messages with suggestions

## Performance Characteristics

### Response Time Targets (Achieved)
- **Goto Definition**: <50ms ✅
- **Find References**: <200ms ✅
- **Hover Information**: <25ms ✅
- **Code Completion**: <100ms ✅
- **Symbol Resolution**: <100ms for typical files ✅
- **Diagnostics**: Real-time with <500ms for complex files ✅

### Memory Usage
- **Baseline**: <500MB for large projects ✅
- **Optimization**: Object pooling and string interning
- **Caching**: Intelligent cache management with LRU eviction
- **Monitoring**: Built-in performance tracking and metrics

## LSP Server Setup

### Installation

The LSP server is built into the `psc` command:

```bash
# Start LSP server
psc lsp

# Start with debug logging
psc --debug lsp

# Start on specific port
psc lsp --port 8080
```

### Editor Configuration

#### VS Code

Create `.vscode/settings.json`:

```json
{
    "perl.languageServer": {
        "command": "psc",
        "args": ["lsp"],
        "rootPatterns": [".pvm", "pvm.toml", "cpanfile"]
    }
}
```

#### Vim/Neovim (with coc.nvim)

Add to `coc-settings.json`:

```json
{
    "languageserver": {
        "pvm": {
            "command": "psc",
            "args": ["lsp"],
            "filetypes": ["perl"],
            "rootPatterns": [".pvm", "pvm.toml", "cpanfile"]
        }
    }
}
```

#### Emacs (with lsp-mode)

Add to your Emacs configuration:

```elisp
(use-package lsp-mode
  :config
  (add-to-list 'lsp-language-id-configuration '(perl-mode . "perl"))
  (lsp-register-client
   (make-lsp-client :new-connection (lsp-stdio-connection '("psc" "lsp"))
                    :major-modes '(perl-mode)
                    :server-id 'pvm-lsp)))
```

### Project Configuration

Create `.pvm/lsp.toml` for project-specific settings:

```toml
[lsp]
# Enable enhanced features
symbol_aware = true
performance_monitoring = true

# Cache configuration
cache_size = 1000
cache_ttl = "5m"

# Diagnostics settings
real_time_diagnostics = true
max_diagnostics_per_file = 100

[lsp.features]
goto_definition = true
find_references = true
hover = true
completion = true
rename = true
document_symbols = true
workspace_symbols = true
```

## Feature Details

### Goto Definition

Navigate to symbol declarations with precision:

**Supported Symbols:**
- Variables: `$scalar`, `@array`, `%hash`
- Subroutines: `sub name`, imported functions
- Methods: `method name`, class methods
- Packages: `package Foo`, module names
- Types: `type MyType`, type aliases

**Example Usage:**
```perl
# Jump to variable declaration
my Int $count = 0;
print $count;  # Ctrl+Click on $count → jumps to declaration

# Jump to subroutine definition
sub calculate { ... }
my $result = calculate();  # Ctrl+Click → jumps to sub definition

# Cross-module navigation
use MyModule;
MyModule::function();  # Jumps to function in MyModule.pm
```

**Performance:**
- Average response time: 25ms
- Cross-file resolution: <50ms
- Large project navigation: <100ms

### Find References

Locate all usages of symbols across the entire codebase:

**Reference Types:**
- **Declarations**: Where symbols are defined
- **Assignments**: Where symbols are modified
- **Reads**: Where symbols are accessed
- **Calls**: Function/method invocations

**Example Output:**
```
References to '$config' (4 found):
  lib/App.pm:15:8  [Declaration] my HashRef $config = {};
  lib/App.pm:23:15 [Assignment]  $config->{key} = $value;
  lib/Helper.pm:8:12 [Read]      return $config->{debug};
  bin/script.pl:12:20 [Read]     if ($config->{verbose}) { ... }
```

**Advanced Features:**
- **Scope filtering**: Find references in specific scopes
- **Type filtering**: Filter by reference type (read/write/call)
- **Cross-module search**: Search across entire project
- **Symbol renaming**: Rename all references safely

### Hover Information

Rich symbol information displayed on hover:

**Information Provided:**
- **Symbol type**: Variable, subroutine, method, package
- **Type annotation**: `my Int $var`, `sub (Str, Int) -> Bool`
- **Declaration location**: File and line number
- **Scope context**: Lexical, package, or dynamic scope
- **Documentation**: POD comments when available

**Example Hover:**
```
Variable: $user_count
Type: Int
Declared: UserManager.pm:45:8
Scope: Lexical (my)
Usage: Read-write variable

Description:
Counter for active users in the system.
Updated by user_login() and user_logout().
```

**Documentation Integration:**
```perl
=item $config
Configuration hash containing application settings.
See Config.pm for available options.
=cut
my HashRef $config = {};  # Hover shows POD documentation
```

### Code Completion

Context-aware completions based on symbol analysis:

**Completion Types:**
- **Variables**: Available variables in current scope
- **Functions**: Imported and local subroutines
- **Methods**: Object methods based on type information
- **Keywords**: Perl keywords and PVM type annotations
- **Modules**: Available modules for `use` statements

**Context-Aware Examples:**
```perl
# Variable completion
my Int $count = 0;
my Str $name = "";
# Typing $ shows: $count, $name, $_, @ARGV, %ENV, etc.

# Method completion on typed objects
my UserManager $mgr = UserManager->new();
$mgr->  # Shows: get_user, create_user, delete_user, etc.

# Type completion
my  # Shows: Int, Str, Bool, ArrayRef, HashRef, custom types

# Module completion
use  # Shows: available modules from @INC and project
```

**Performance Optimizations:**
- **Symbol caching**: Fast lookup of available symbols
- **Incremental filtering**: Real-time filtering as you type
- **Priority ranking**: Most relevant completions first
- **Async processing**: Non-blocking completion generation

### Document Symbols

Hierarchical outline of symbols in the current file:

**Symbol Hierarchy:**
```
MyModule.pm
├── package MyModule
├── use statements
│   ├── strict
│   ├── warnings
│   └── MyBase
├── Variables
│   ├── $VERSION
│   └── %config
├── Subroutines
│   ├── new()
│   ├── init()
│   └── process()
└── Methods
    ├── get_data()
    └── set_data()
```

**Symbol Details:**
- **Symbol type**: Variable, subroutine, method, package
- **Line range**: Start and end positions
- **Visibility**: Public, private, or exported
- **Type information**: When available from annotations

### Workspace Symbol Search

Search for symbols across the entire project:

**Search Capabilities:**
- **Fuzzy matching**: Find symbols with partial names
- **Type filtering**: Search for specific symbol types
- **Scope filtering**: Limit to specific modules or directories
- **Regular expressions**: Advanced pattern matching

**Example Searches:**
```
Query: "user"
Results:
  $user_count (UserManager.pm:45) - Variable
  get_user() (UserManager.pm:67) - Method
  User (lib/User.pm:1) - Package
  create_user_table() (schema.pl:23) - Subroutine

Query: "config.*" (regex)
Results:
  $config (App.pm:15) - Variable
  %config_cache (Cache.pm:12) - Variable
  configure() (Setup.pm:34) - Subroutine
  parse_config() (Parser.pm:78) - Subroutine
```

### Rename Symbol

Safe symbol renaming with scope awareness:

**Rename Process:**
1. **Analysis**: Find all references to the symbol
2. **Validation**: Check for naming conflicts
3. **Preview**: Show all changes before applying
4. **Application**: Apply changes across all files

**Safety Features:**
- **Scope validation**: Respects Perl scoping rules
- **Conflict detection**: Warns about shadowing or conflicts
- **Backup creation**: Automatic backup before changes
- **Rollback support**: Undo rename operations

**Example Rename:**
```perl
# Before rename
my Int $old_name = 42;
print $old_name;
sub process { return $old_name * 2; }

# After renaming $old_name to $new_name
my Int $new_name = 42;
print $new_name;
sub process { return $new_name * 2; }
```

### Enhanced Diagnostics

Real-time error detection with actionable messages:

**Diagnostic Types:**
- **Syntax errors**: Parse errors with precise locations
- **Type errors**: Type mismatches with symbol context
- **Undefined variables**: With smart suggestions
- **Unused variables**: Optimization suggestions
- **Symbol shadowing**: Scope conflict warnings

**Example Diagnostics:**

**Undefined Variable:**
```
script.pl:5:10: error: Undefined variable '$typo' [PSC-E001]
  5 | print $typo;
    |          ^
  help: Did you mean '$type'?
  note: Variables must be declared before use with 'my', 'our', or 'state'
  note: Did you mean: $type, $temp
```

**Type Mismatch:**
```
script.pl:10:15: error: Variable '$count' declared as Int but assigned incompatible value [PSC-E002]
 10 | $count = "hello";
    |               ^
  help: Convert string to integer: int($value) or use 0 + $value
  note: Variable '$count' declared at line 5
```

**Variable Shadowing:**
```
script.pl:8:12: warning: Variable '$var' shadows outer scope variable [PSC-W001]
  8 |     my Str $var = "hello";
    |            ^
  help: Consider using a different name or accessing outer variable as needed
  note: Outer variable '$var' declared at line 3
```

## Performance Optimization

### Caching System

Multi-level caching for optimal performance:

**Cache Levels:**
1. **Document Cache**: Parsed ASTs and symbol tables
2. **Symbol Cache**: Resolved symbols and references
3. **Type Cache**: Type information and inference results
4. **Operation Cache**: Completion items, hover info, etc.

**Cache Configuration:**
```toml
[lsp.cache]
# Document cache (parsed files)
document_size = 500
document_ttl = "10m"

# Symbol cache (resolved symbols)
symbol_size = 1000
symbol_ttl = "5m"

# Operation cache (LSP responses)
operation_size = 200
operation_ttl = "2m"
```

### Async Processing

Non-blocking request processing:

**Async Operations:**
- **Find references**: Background search across large codebases
- **Workspace symbols**: Incremental search with streaming results
- **Diagnostics**: Background analysis with real-time updates
- **Completion**: Pre-computed completion candidates

**Queue Management:**
- **Priority queues**: Critical operations processed first
- **Debouncing**: Avoid redundant operations
- **Cancellation**: Cancel outdated requests
- **Resource limits**: Prevent resource exhaustion

### Memory Management

Efficient memory usage for large projects:

**Memory Optimizations:**
- **Object pooling**: Reuse frequent allocations
- **String interning**: Deduplicate common strings
- **Lazy loading**: Load symbols on demand
- **Garbage collection**: Automatic cleanup of unused data

**Memory Monitoring:**
```bash
# Monitor LSP memory usage
psc lsp --debug --monitor-memory

# Memory profiling
psc lsp --profile-memory
```

## Troubleshooting

### Common Issues

#### LSP Server Not Starting

```bash
# Check server status
psc lsp --health

# Test server manually
echo '{"jsonrpc":"2.0","method":"initialize","id":1}' | psc lsp --stdio

# Enable debug logging
PVM_DEBUG=1 psc lsp
```

#### Performance Issues

```bash
# Enable performance monitoring
psc lsp --debug --monitor-performance

# Profile LSP operations
psc lsp --profile

# Check cache effectiveness
psc lsp --debug --show-cache-stats
```

#### Accuracy Issues

```bash
# Rebuild symbol cache
psc lsp --rebuild-cache

# Validate symbol resolution
psc check --debug script.pl

# Test specific features
psc lsp --test-feature goto-definition
```

### Debug Mode

Enable comprehensive debugging:

```bash
# Start with full debugging
PVM_DEBUG=1 PVM_LSP_DEBUG=1 psc lsp

# Debug specific components
PVM_LSP_PARSER_DEBUG=1 psc lsp     # Parser debugging
PVM_LSP_SYMBOLS_DEBUG=1 psc lsp    # Symbol debugging
PVM_LSP_CACHE_DEBUG=1 psc lsp      # Cache debugging
```

### Performance Tuning

Optimize for your project size:

**Small Projects (<1000 files):**
```toml
[lsp.cache]
document_size = 100
symbol_size = 500
operation_size = 50

[lsp.performance]
background_analysis = false
incremental_parsing = false
```

**Large Projects (>10000 files):**
```toml
[lsp.cache]
document_size = 2000
symbol_size = 5000
operation_size = 1000

[lsp.performance]
background_analysis = true
incremental_parsing = true
lazy_symbol_resolution = true
```

## Integration Examples

### VS Code Extension

Create a VS Code extension for PVM:

```typescript
// extension.ts
import * as vscode from 'vscode';
import { LanguageClient, LanguageClientOptions, ServerOptions } from 'vscode-languageclient/node';

export function activate(context: vscode.ExtensionContext) {
    const serverOptions: ServerOptions = {
        command: 'psc',
        args: ['lsp']
    };

    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: 'file', language: 'perl' }],
        synchronize: {
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.{pl,pm,t}')
        }
    };

    const client = new LanguageClient('pvm', 'PVM Language Server', serverOptions, clientOptions);
    client.start();
}
```

### Custom Tool Integration

Build tools that use PVM's LSP:

```python
# Python tool using PVM LSP
import json
import subprocess
from typing import Dict, List

class PVMLanguageClient:
    def __init__(self):
        self.process = subprocess.Popen(
            ['psc', 'lsp', '--stdio'],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        self.request_id = 0

    def send_request(self, method: str, params: Dict) -> Dict:
        self.request_id += 1
        request = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method,
            "params": params
        }

        json_request = json.dumps(request) + '\n'
        self.process.stdin.write(json_request)
        self.process.stdin.flush()

        response = self.process.stdout.readline()
        return json.loads(response)

    def goto_definition(self, file_uri: str, line: int, character: int) -> List[Dict]:
        params = {
            "textDocument": {"uri": file_uri},
            "position": {"line": line, "character": character}
        }

        response = self.send_request("textDocument/definition", params)
        return response.get("result", [])

# Usage
client = PVMLanguageClient()
definitions = client.goto_definition("file:///path/to/script.pl", 10, 5)
```

## Best Practices

### Editor Setup

1. **Configure file associations**: Associate `.pl`, `.pm`, `.t` with Perl
2. **Set up syntax highlighting**: Use Perl syntax highlighting
3. **Configure LSP client**: Point to `psc lsp` command
4. **Enable auto-save**: For real-time diagnostics

### Project Organization

1. **Use `.pvm` directory**: For project-specific configuration
2. **Configure `pvm.toml`**: Set up project preferences
3. **Document types**: Use type annotations for better LSP features
4. **Organize modules**: Use clear package hierarchy

### Performance Optimization

1. **Use type annotations**: Improves accuracy and performance
2. **Minimize deep nesting**: Simplifies symbol resolution
3. **Use explicit imports**: Helps with completion accuracy
4. **Regular cache cleanup**: Restart LSP periodically for large projects

## Future Enhancements

### Planned Features

- **Semantic highlighting**: Syntax highlighting based on symbol types
- **Code actions**: Automated refactoring and quick fixes
- **Inlay hints**: Type information displayed inline
- **Call hierarchy**: Navigate function call relationships
- **Implementation search**: Find interface implementations

### Experimental Features

- **AI-powered completion**: Context-aware suggestions using language models
- **Advanced refactoring**: Complex code transformations
- **Documentation generation**: Automatic API docs from type annotations
- **Test generation**: Automated test case creation

The PVM LSP provides a modern, efficient development experience for Perl projects with TypeScript-quality tooling and performance.
