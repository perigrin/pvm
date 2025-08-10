# Built-in Markdown Rendering

PVM provides built-in styled markdown rendering using the [Glamour](https://github.com/charmbracelet/glamour) library. This ensures consistent, beautiful markdown display across all platforms without requiring external dependencies.

## Overview

Glamour is a Go library for rendering markdown in terminal applications with automatic theme detection and beautiful styling. PVM integrates Glamour directly, providing rich markdown formatting out-of-the-box.

## Features

### Automatic Styling
- **Dark/Light Theme Detection**: Automatically detects terminal background and applies appropriate styling
- **Syntax Highlighting**: Code blocks are highlighted with appropriate colors
- **Responsive Layout**: Content wraps appropriately for different terminal widths
- **Rich Typography**: Headers, emphasis, lists, and other elements are beautifully styled

### Commands with Enhanced Markdown Display

#### Release Notes and Changelog
- **`pvm self release-notes [version]`**: View specific version release notes
- **`pvm self changelog`**: View complete PVM changelog
- **Update commands**: `pvm self update` shows styled release notes

#### Help System
- **`pvm help docs <name>`**: Enhanced documentation viewing
- **Command help**: All help content uses improved markdown rendering

#### Version Information
- **`pvm version --detailed`**: Shows current version with styled release notes

## Configuration Options

### Raw Markdown Mode

For users who prefer plain text output or need compatibility with certain terminals:

```bash
pvm self changelog --raw-markdown
```

The `--raw-markdown` flag disables styled rendering and outputs plain markdown text.

### Color Mode Control

Markdown styling respects PVM's global color settings:

```bash
pvm self changelog --color=never    # Plain text, no styling
pvm self changelog --color=always   # Force styled output
pvm self changelog --color=auto     # Automatic detection (default)
```

## Technical Details

### Built-in Integration
- **No External Dependencies**: Glamour is compiled into PVM binaries
- **Cross-platform**: Works consistently on Linux, macOS, and Windows
- **Performance**: No subprocess overhead, faster rendering
- **Reliability**: No installation or PATH configuration required

### Fallback Behavior
If Glamour rendering fails for any reason, PVM automatically falls back to basic markdown display, ensuring commands always work.

## Migration from External Glow

Previous versions of PVM used external [glow](https://github.com/charmbracelet/glow) as an optional dependency. The new built-in approach provides several advantages:

### Benefits
- **Consistent Experience**: All users get enhanced markdown rendering
- **Simplified Installation**: No separate tool installation required
- **Better Reliability**: No dependency on external tool versions or availability
- **Performance**: Eliminates subprocess spawning overhead

### Compatibility
- All existing commands work exactly the same way
- Enhanced markdown display is now the default for all users
- Users who prefer plain text can use the `--raw-markdown` flag

## Examples

### Viewing Release Notes
```bash
# Styled release notes (default)
pvm self release-notes v1.0.0

# Plain text release notes
pvm self release-notes v1.0.0 --raw-markdown
```

### Viewing Changelog
```bash
# Full styled changelog
pvm self changelog

# Changelog without color styling
pvm self changelog --color=never
```

### Update with Release Notes
```bash
# Update with styled release notes display
pvm self update

# Update with plain text release notes
pvm self update --raw-markdown
```

The built-in markdown rendering provides a professional, consistent experience that enhances PVM's usability without adding complexity or dependencies.
