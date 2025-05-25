# Editor Integration for PVM/PSC

This document provides setup instructions for integrating the PSC (Perl Script Compiler) Language Server with various text editors and IDEs.

## Overview

The PSC Language Server provides:
- **Type checking** with real-time error reporting
- **Hover information** for types, functions, and variables
- **Auto-completion** for Perl keywords, types, and symbols
- **Go to Definition** to navigate to symbol definitions
- **Find References** to locate all uses of a symbol
- **Document Formatting** to clean up code style
- **Code Actions** for quick fixes and refactoring
- **Structured diagnostics** in multiple formats (LSP, JSON, SARIF, etc.)

## Quick Start

1. Ensure PSC is installed and accessible in your PATH
2. The language server can be started in two modes:
   - **stdio mode**: `psc lsp --stdio` (recommended for editors)
   - **TCP mode**: `psc lsp --tcp --port 9999` (for debugging)

## Editor-Specific Setup

### Visual Studio Code

Create or update `.vscode/settings.json`:

```json
{
  "perl.languageServer": {
    "enable": true,
    "command": "psc",
    "args": ["lsp", "--stdio"],
    "filetypes": ["perl", "pl", "pm"]
  },
  "perl.typecheck": {
    "enable": true,
    "onSave": true,
    "showDiagnostics": true
  },
  "files.associations": {
    "*.pl": "perl",
    "*.pm": "perl",
    "*.t": "perl"
  }
}
```

**Alternative: Using a generic LSP extension**

If using the `vscode-languageserver-client` extension:

```json
{
  "languageServerExample.enable": true,
  "languageServerExample.command": "psc",
  "languageServerExample.args": ["lsp", "--stdio"],
  "languageServerExample.filetypes": ["perl"],
  "languageServerExample.settings": {
    "psc": {
      "typecheck": true,
      "flowSensitive": true,
      "diagnostics": "lsp"
    }
  }
}
```

### Neovim (with nvim-lspconfig)

Add to your Neovim configuration (`init.lua` or `init.vim`):

```lua
-- Using nvim-lspconfig
local lspconfig = require('lspconfig')

-- Register PSC language server
local configs = require('lspconfig.configs')
if not configs.psc then
  configs.psc = {
    default_config = {
      cmd = {'psc', 'lsp', '--stdio'},
      filetypes = {'perl'},
      root_dir = function(fname)
        return lspconfig.util.find_git_ancestor(fname) or vim.fn.getcwd()
      end,
      settings = {
        psc = {
          typecheck = true,
          flowSensitive = true,
          diagnostics = "lsp"
        }
      }
    }
  }
end

-- Setup the server
lspconfig.psc.setup({
  on_attach = function(client, bufnr)
    -- Enable completion triggered by <c-x><c-o>
    vim.api.nvim_buf_set_option(bufnr, 'omnifunc', 'v:lua.vim.lsp.omnifunc')

    -- Mappings
    local bufopts = { noremap=true, silent=true, buffer=bufnr }
    vim.keymap.set('n', 'gD', vim.lsp.buf.declaration, bufopts)
    vim.keymap.set('n', 'gd', vim.lsp.buf.definition, bufopts)
    vim.keymap.set('n', 'K', vim.lsp.buf.hover, bufopts)
    vim.keymap.set('n', '<space>rn', vim.lsp.buf.rename, bufopts)
    vim.keymap.set('n', '<space>ca', vim.lsp.buf.code_action, bufopts)
  end,
  capabilities = require('cmp_nvim_lsp').default_capabilities()
})

-- Auto-detect Perl files with type annotations
vim.api.nvim_create_autocmd({"BufRead", "BufNewFile"}, {
  pattern = {"*.pl", "*.pm", "*.t"},
  callback = function()
    local content = table.concat(vim.api.nvim_buf_get_lines(0, 0, 50, false), "\n")
    if string.match(content, "my%s+%w+%s+%$") then  -- Basic type annotation detection
      vim.bo.filetype = "perl"
    end
  end,
})
```

### Vim (with vim-lsp)

Add to your `.vimrc`:

```vim
" Using vim-lsp plugin
if executable('psc')
  autocmd User lsp_setup call lsp#register_server({
      \ 'name': 'psc',
      \ 'cmd': {server_info->['psc', 'lsp', '--stdio']},
      \ 'allowlist': ['perl'],
      \ 'workspace_config': {
      \   'psc': {
      \     'typecheck': v:true,
      \     'flowSensitive': v:true,
      \     'diagnostics': 'lsp'
      \   }
      \ }
      \ })
endif

" Keybindings
function! s:on_lsp_buffer_enabled() abort
    setlocal omnifunc=lsp#complete
    setlocal signcolumn=yes
    nmap <buffer> gd <plug>(lsp-definition)
    nmap <buffer> <f2> <plug>(lsp-rename)
    nmap <buffer> K <plug>(lsp-hover)
endfunction

augroup lsp_install
    au!
    autocmd User lsp_buffer_enabled call s:on_lsp_buffer_enabled()
augroup END
```

### Emacs (with lsp-mode)

Add to your Emacs configuration:

```elisp
;; Using lsp-mode
(use-package lsp-mode
  :hook (perl-mode . lsp)
  :commands lsp
  :config
  (add-to-list 'lsp-language-id-configuration '(perl-mode . "perl"))
  (lsp-register-client
   (make-lsp-client
    :new-connection (lsp-stdio-connection '("psc" "lsp" "--stdio"))
    :major-modes '(perl-mode)
    :server-id 'psc
    :initialization-options
    '(:psc (:typecheck t :flowSensitive t :diagnostics "lsp")))))

;; Optional: Enhanced Perl mode
(use-package cperl-mode
  :mode "\\.\\(pl\\|pm\\|t\\)\\'"
  :config
  (defalias 'perl-mode 'cperl-mode)
  (setq cperl-highlight-variables-indiscriminately t))
```

### Sublime Text (with LSP package)

Create or update `LSP.sublime-settings`:

```json
{
  "clients": {
    "psc": {
      "enabled": true,
      "command": ["psc", "lsp", "--stdio"],
      "selector": "source.perl",
      "settings": {
        "psc.typecheck": true,
        "psc.flowSensitive": true,
        "psc.diagnostics": "lsp"
      }
    }
  }
}
```

Add to `Perl.sublime-syntax` or create a custom syntax file:

```yaml
%YAML 1.2
---
name: Perl with Types
file_extensions: [pl, pm, t]
scope: source.perl.typed

extends: Packages/Perl/Perl.sublime-syntax

contexts:
  main:
    - match: '\b(my|our|state)\s+([A-Z][A-Za-z0-9_]*(?:\[[^\]]*\])?)\s+(\$[a-zA-Z_][a-zA-Z0-9_]*)'
      captures:
        1: keyword.declaration.perl
        2: storage.type.perl
        3: variable.other.perl
```

### Atom (with atom-ide-ui)

Add to `~/.atom/config.cson`:

```cson
"*":
  "atom-ide-ui":
    "atom-ide-diagnostics-ui":
      showDiagnosticTraces: true
    "atom-ide-code-actions":
      showActionMenu: true
  "ide-perl":
    serverPath: "psc"
    serverArgs: ["lsp", "--stdio"]
    settings:
      psc:
        typecheck: true
        flowSensitive: true
        diagnostics: "lsp"
```

## Configuration Options

The PSC Language Server supports these configuration options:

### Server Settings

```json
{
  "psc": {
    "typecheck": true,              // Enable type checking
    "flowSensitive": true,          // Enable flow-sensitive analysis
    "diagnostics": "lsp",           // Diagnostic format: "lsp", "json", "text"
    "maxErrors": 100,               // Maximum errors to report
    "includeWarnings": true,        // Include warnings in diagnostics
    "autoSave": false,              // Auto-save before type checking
    "timeout": 5000,                // Type checking timeout (ms)
    "logLevel": "info"              // Logging level: "debug", "info", "warn", "error"
  }
}
```

### Trigger Characters

The language server provides completion on these trigger characters:
- `$` - Variable completion
- `@` - Array completion
- `%` - Hash completion
- `:` - Module method completion (after `::`)
- `.` - Object method completion
- `-` - Dereference completion (after `->`)

### File Type Detection

The server activates for files with:
- Extensions: `.pl`, `.pm`, `.t`
- Content containing type annotations (pattern: `my Type $var`)
- Shebang lines with `perl`

## Troubleshooting

### Language Server Not Starting

1. **Check PSC installation**: `psc --version`
2. **Verify LSP command**: `psc lsp --help`
3. **Test stdio mode**: `echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | psc lsp --stdio`

### No Type Checking

1. **Verify file has type annotations**: Look for patterns like `my Str $var`
2. **Check file extension**: Ensure `.pl`, `.pm`, or `.t`
3. **Enable diagnostics**: Set `"psc.typecheck": true` in settings

### No Completions

1. **Check trigger characters**: Try typing `$`, `my `, or other triggers
2. **Verify server capabilities**: Check initialization response includes completion
3. **Enable verbose logging**: Set `"psc.logLevel": "debug"`

### Performance Issues

1. **Increase timeout**: Set higher `"psc.timeout"` value
2. **Limit errors**: Set lower `"psc.maxErrors"` value
3. **Disable flow-sensitive**: Set `"psc.flowSensitive": false`

## Using LSP Features

### Go to Definition
- **VSCode**: `F12` or `Ctrl+Click` on a symbol
- **Neovim**: `gd` in normal mode
- **Vim**: `gd` or `<plug>(lsp-definition)`
- **Emacs**: `M-.` or `xref-find-definitions`

### Find References
- **VSCode**: `Shift+F12` or right-click → "Find All References"
- **Neovim**: `gr` in normal mode
- **Vim**: `<plug>(lsp-references)`
- **Emacs**: `M-?` or `xref-find-references`

### Document Formatting
- **VSCode**: `Shift+Alt+F` or right-click → "Format Document"
- **Neovim**: `<space>f` or `:lua vim.lsp.buf.formatting()`
- **Vim**: `<plug>(lsp-document-format)`
- **Emacs**: `M-x lsp-format-buffer`

### Code Actions
- **VSCode**: `Ctrl+.` or lightbulb icon
- **Neovim**: `<space>ca` or `:lua vim.lsp.buf.code_action()`
- **Vim**: `<plug>(lsp-code-action)`
- **Emacs**: `M-x lsp-execute-code-action`

Available code actions include:
- **Quick fixes** for undefined variables
- **Type mismatch corrections**
- **Extract variable** refactoring
- **Add missing type annotations**

## Advanced Configuration

### Custom Type Definitions

Place `.ptd` (Perl Type Definition) files in your project to enhance completions:

```json
{
  "module": "My::Module",
  "types": [
    {
      "name": "CustomType",
      "kind": "class",
      "methods": [
        {
          "name": "new",
          "parameters": [{"name": "class", "type": "Str"}],
          "returnType": "CustomType"
        }
      ]
    }
  ]
}
```

### Project Configuration

Create `.pvm/pvm.toml` in your project root:

```toml
[psc]
typecheck = true
flow_sensitive = true
max_errors = 50

[psc.diagnostics]
format = "lsp"
include_warnings = true
include_context = true

[psc.completion]
include_builtins = true
include_modules = true
trigger_on_dot = true
```

### Integration with Build Systems

#### With Make

```makefile
typecheck:
	@find . -name "*.pl" -o -name "*.pm" | xargs -I {} psc check {}

typecheck-json:
	@find . -name "*.pl" -o -name "*.pm" | xargs -I {} psc check --format json {}
```

#### With GitHub Actions

```yaml
name: Type Check
on: [push, pull_request]
jobs:
  typecheck:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Install PVM
      run: |
        curl -sSL https://install.pvm.dev | sh
        echo "$HOME/.pvm/bin" >> $GITHUB_PATH
    - name: Type Check
      run: psc check --format sarif --output typecheck.sarif src/
    - name: Upload SARIF
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: typecheck.sarif
```

## Integration Examples

### VSCode Extension Package

Minimal `package.json` for a custom PSC extension:

```json
{
  "name": "psc-language-support",
  "version": "1.0.0",
  "engines": {
    "vscode": "^1.50.0"
  },
  "activationEvents": [
    "onLanguage:perl"
  ],
  "main": "./out/extension.js",
  "contributes": {
    "languages": [{
      "id": "perl",
      "extensions": [".pl", ".pm", ".t"],
      "aliases": ["Perl", "perl"]
    }],
    "grammars": [{
      "language": "perl",
      "scopeName": "source.perl",
      "path": "./syntaxes/perl.tmLanguage.json"
    }],
    "configuration": {
      "type": "object",
      "title": "PSC Configuration",
      "properties": {
        "psc.enable": {
          "type": "boolean",
          "default": true,
          "description": "Enable PSC language server"
        }
      }
    }
  }
}
```

This comprehensive guide should help users integrate PSC type checking into their preferred development environment.
