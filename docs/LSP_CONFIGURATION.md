# LSP Configuration for PSC

The PSC Language Server can be configured through `pvm.toml` configuration files, providing persistent settings for logging, performance, features, and project-specific behavior.

## Configuration Structure

LSP configuration is specified under the `[psc.lsp]` section in your `pvm.toml` file:

```toml
[psc.lsp]
# Logging configuration
log_file = "/tmp/psc-lsp.log"
log_level = "info"    # debug, info, warn, error
verbose = false

# Communication settings
default_mode = "stdio"  # stdio, tcp
tcp_port = 9999

# Feature toggles
enable_hover = true
enable_completion = true
enable_definition = true
enable_references = true
enable_formatting = true

# Performance settings
max_cache_size = 1000
request_timeout = "5s"
diagnostics_delay = "500ms"

# Project-specific settings
workspace_symbols = true
cross_file_analysis = true
exclude_patterns = ["**/test_data/**", "**/temp/**"]
include_directories = ["lib", "script"]
```

## Configuration Precedence

Configuration follows the standard PVM hierarchy:

1. **Project configuration** (`project/.pvm/pvm.toml`) - highest precedence
2. **User configuration** (`$XDG_CONFIG_HOME/pvm/pvm.toml`) - medium precedence
3. **System configuration** (`/etc/pvm/pvm.toml`) - lowest precedence

**Command-line flags always override configuration file settings.**

## Configuration Examples

### Global User Configuration

Set up default LSP behavior for all projects:

```toml
# ~/.config/pvm/pvm.toml
[psc.lsp]
# Enable verbose logging by default for development
verbose = true
log_file = "$XDG_CACHE_HOME/pvm/logs/psc-lsp.log"
log_level = "info"

# Enable all features
enable_hover = true
enable_completion = true
enable_definition = true
enable_references = true
enable_formatting = true

# Reasonable performance defaults
max_cache_size = 1000
request_timeout = "5s"
diagnostics_delay = "500ms"

# Default to stdio for editor integration
default_mode = "stdio"
```

### Project-Specific Configuration

Override global settings for a specific project:

```toml
# myproject/.pvm/pvm.toml
[psc.lsp]
# Debug logging for this project
log_file = "logs/psc-lsp.log"
log_level = "debug"
verbose = true

# Enable advanced features for this project
workspace_symbols = true
cross_file_analysis = true

# Project-specific includes/excludes
include_directories = ["lib", "script", "bin"]
exclude_patterns = ["**/test_data/**", "**/temp/**", "**/local/**"]

# Faster diagnostics for active development
diagnostics_delay = "200ms"
```

### Minimal Configuration

Disable optional features for basic type checking only:

```toml
[psc.lsp]
# Minimal feature set
enable_hover = false
enable_completion = false
enable_definition = false
enable_references = false

# Quiet logging
log_level = "error"
verbose = false

# Basic performance settings
max_cache_size = 500
```

### Development/Debugging Configuration

Enhanced logging and debugging features:

```toml
[psc.lsp]
# Comprehensive logging
log_file = "/tmp/psc-lsp-debug.log"
log_level = "debug"
verbose = true

# Use TCP mode for easier debugging
default_mode = "tcp"
tcp_port = 9999

# Enable all analysis features
workspace_symbols = true
cross_file_analysis = true
enable_hover = true
enable_completion = true
enable_definition = true
enable_references = true
enable_formatting = true

# Aggressive caching for performance analysis
max_cache_size = 2000
```

## Configuration Options Reference

### Logging Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `log_file` | string | "" | Path to log file (supports env vars like `$HOME`) |
| `log_level` | string | "info" | Logging level: debug, info, warn, error |
| `verbose` | boolean | false | Enable verbose logging output |

### Communication Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `default_mode` | string | "stdio" | Default communication mode: stdio, tcp |
| `tcp_port` | integer | 9999 | TCP port for server mode (1-65535) |

### Feature Toggles

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enable_hover` | boolean | true | Enable hover information |
| `enable_completion` | boolean | true | Enable auto-completion |
| `enable_definition` | boolean | true | Enable go-to-definition |
| `enable_references` | boolean | true | Enable find references |
| `enable_formatting` | boolean | true | Enable code formatting |

### Performance Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `max_cache_size` | integer | 1000 | Maximum number of cached files |
| `request_timeout` | duration | "5s" | Timeout for LSP requests |
| `diagnostics_delay` | duration | "500ms" | Delay before sending diagnostics |

### Project-Specific Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `workspace_symbols` | boolean | true | Enable workspace symbol search |
| `cross_file_analysis` | boolean | true | Enable cross-file type analysis |
| `exclude_patterns` | array | ["**/test_data/**", "**/temp/**", "**/.git/**"] | File patterns to exclude |
| `include_directories` | array | ["lib", "script"] | Directories to include in analysis |

## Command Usage

Once configured, simply run the LSP server without flags:

```bash
# Uses configuration from pvm.toml files
psc lsp

# Command-line flags still work and override config
psc lsp --verbose --port 8080
```

## Editor Integration

With configuration files, editor setup becomes much simpler:

### Vim/Neovim (vim-lsp)

```vim
if executable('psc')
    au User lsp_setup call lsp#register_server({
        \ 'name': 'pvm-lsp',
        \ 'cmd': ['psc', 'lsp'],  " No flags needed - uses config
        \ 'allowlist': ['perl'],
        \ })
endif
```

### VS Code

```json
{
    "settings": {
        "perl.perlPath": "psc",
        "perl.perlArgs": ["lsp"]
    }
}
```

### Emacs (lsp-mode)

```elisp
(lsp-register-client
 (make-lsp-client :new-connection (lsp-stdio-connection '("psc" "lsp"))
                  :major-modes '(perl-mode)
                  :server-id 'pvm-lsp))
```

## Environment Variable Support

Configuration files support environment variable expansion:

```toml
[psc.lsp]
log_file = "$HOME/logs/psc-lsp.log"
log_file = "$XDG_CACHE_HOME/pvm/lsp.log"
exclude_patterns = ["**/$USER/temp/**"]
```

## Validation

Configuration values are validated when loaded. Invalid values will cause the LSP server to fall back to defaults and emit warnings:

```bash
$ psc lsp
Warning: Failed to load configuration: LSP log_level must be one of: debug, info, warn, error
Starting PSC Language Server in stdio mode
```

## Troubleshooting

### Configuration Not Loading

1. Check file path: `psc lsp` looks for config files in standard locations
2. Verify TOML syntax: malformed TOML will prevent loading
3. Check file permissions: ensure the config file is readable

### Logging Issues

1. Verify log file directory exists and is writable
2. Use absolute paths for `log_file` when in doubt
3. Check that environment variables are set if used in paths

### Performance Issues

1. Reduce `max_cache_size` if memory usage is high
2. Increase `diagnostics_delay` to reduce CPU usage
3. Use `exclude_patterns` to avoid analyzing unnecessary files

### TCP Mode Not Working

1. Check that `tcp_port` is not in use by another process
2. Verify port is within valid range (1-65535)
3. Consider firewall restrictions if connecting remotely
