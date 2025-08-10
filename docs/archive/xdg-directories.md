# XDG Base Directory Support

The PVM Ecosystem follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) for storing configuration and data files. This ensures that all files are organized in a platform-appropriate manner and respects user preferences.

## Directory Structure

PVM uses the following directory structure:

### Configuration Files

- **Project-level**: `.pvm/pvm.toml` in the project directory
- **User-level**: `$XDG_CONFIG_HOME/pvm/pvm.toml` (defaults to `~/.config/pvm/pvm.toml`)
- **System-level**: `/etc/pvm/pvm.toml` (on Windows: `%ProgramData%\pvm\pvm.toml`)

Configuration files follow a layered approach with clear precedence:
1. Project configuration (highest priority)
2. User configuration
3. System configuration (lowest priority)

### Data Directories

PVM uses the following XDG directories for storing data:

- **Installed Perl versions**: `$XDG_DATA_HOME/pvm/versions/`
- **Downloaded source archives**: `$XDG_CACHE_HOME/pvm/sources/`
- **Shim executables**: `$XDG_DATA_HOME/pvm/shims/`
- **Type definitions**: `$XDG_DATA_HOME/pvm/type_definitions/`
- **Build cache**: `$XDG_CACHE_HOME/pvm/build/`
- **Global module libraries (PM)**: `$XDG_DATA_HOME/pvm/library/{version}/`

#### Module Installation Directories

PM organizes global module installations using version-specific directories to prevent conflicts between different Perl versions:

- **System Perl modules**: `$XDG_DATA_HOME/pvm/library/system/`
- **PVM-managed Perl modules**: `$XDG_DATA_HOME/pvm/library/{perl-version}/` (e.g., `library/5.38.0/`)
- **Project-local modules**: `{project-root}/local/` (when in project context)

## Platform-Specific Defaults

If XDG environment variables are not set, PVM falls back to platform-specific default locations:

### Linux/Unix
- `$XDG_CONFIG_HOME`: `~/.config`
- `$XDG_CACHE_HOME`: `~/.cache`
- `$XDG_DATA_HOME`: `~/.local/share`
- `$XDG_STATE_HOME`: `~/.local/state`

### macOS
- `$XDG_CONFIG_HOME`: `~/.config`
- `$XDG_CACHE_HOME`: `~/Library/Caches`
- `$XDG_DATA_HOME`: `~/Library/Application Support`
- `$XDG_STATE_HOME`: `~/.local/state`

### Windows
- `$XDG_CONFIG_HOME`: `%USERPROFILE%\AppData\Roaming`
- `$XDG_CACHE_HOME`: `%USERPROFILE%\AppData\Local\Cache`
- `$XDG_DATA_HOME`: `%USERPROFILE%\AppData\Local`
- `$XDG_STATE_HOME`: `%USERPROFILE%\AppData\Local\State`

## Configuration Management

PVM provides several commands for working with configuration files:

```
pvm config show               # Show effective (merged) configuration
pvm config get pvm.default_perl  # Get a specific configuration value
pvm config set pvm.default_perl 5.38.0  # Set a configuration value

# Initialize configuration files
pvm config init --user        # Initialize user configuration (default)
pvm config init --project     # Initialize project configuration
pvm config init --system      # Initialize system configuration (requires admin privileges)
```

## Project Detection

PVM automatically detects project roots by looking for a `.pvm` directory in the current directory or any parent directory. This allows project-specific configuration to be applied automatically when working within a project.

## Environment Variable Expansion

Configuration values that start with `$` are automatically expanded using environment variables. This is particularly useful for paths that need to reference user-specific locations:

```toml
[psc]
type_definitions_path = "$XDG_DATA_HOME/pvm/type_definitions"
```

## Implementation Details

The XDG Base Directory support is implemented in the `internal/xdg` package, which provides:

- Detection of XDG directories with proper platform-specific fallbacks
- Creation of required directories as needed
- Helper functions for accessing configuration and data paths
- Project detection and path resolution

This implementation ensures that PVM works correctly across all supported platforms while following platform-specific conventions and user preferences.
