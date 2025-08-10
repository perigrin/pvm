# Glow Integration

PVM integrates with [glow](https://github.com/charmbracelet/glow) from Charm Bracelet to provide enhanced markdown rendering for release notes, documentation, and help content.

## Overview

Glow is a terminal-based markdown reader that renders beautiful, styled markdown content. PVM automatically detects and uses glow when available, providing a significantly enhanced user experience for viewing markdown content.

## Features

### Enhanced Release Notes Display

- **Update command**: `pvm update` now shows release notes with rich formatting
- **Dedicated commands**: New commands for viewing release notes and changelogs
- **Version command**: Enhanced version display with release notes

### New Commands

#### `pvm release-notes [version]`

View release notes for specific PVM versions with enhanced formatting.

**Options:**
- `--latest`: Show latest release notes (default if no version specified)
- `--prerelease`: Include pre-release versions
- `--token`: GitHub token for higher API rate limits

**Examples:**
```bash
pvm release-notes                    # Show latest release notes
pvm release-notes --latest           # Show latest release notes
pvm release-notes v1.0.0-rc26       # Show specific version
pvm release-notes --prerelease       # Include pre-releases
```

#### `pvm changelog`

View formatted changelog with recent releases.

**Options:**
- `--limit`: Number of recent releases to show (default: 10)
- `--prerelease`: Include pre-release versions
- `--token`: GitHub token for higher API rate limits

**Examples:**
```bash
pvm changelog                        # Show last 10 releases
pvm changelog --limit 5              # Show last 5 releases
pvm changelog --prerelease           # Include pre-releases
```

#### `pvm version --detailed`

Show detailed version information including release notes.

**Example:**
```bash
pvm version --detailed               # Show version with release notes
```

### Enhanced Help System

Documentation viewing (`pvm help docs <name>`) now uses glow for improved readability of embedded documentation.

## Installation

### Installing Glow

Glow is an optional dependency. PVM will automatically fall back to basic markdown rendering if glow is not available.

**macOS:**
```bash
brew install glow
```

**Linux:**
Download from the [releases page](https://github.com/charmbracelet/glow/releases) or use your package manager.

**Go:**
```bash
go install github.com/charmbracelet/glow@latest
```

### Verification

To verify glow is available and working:

```bash
which glow
glow --version
```

## Usage

### Automatic Detection

PVM automatically detects glow availability at runtime. No configuration is required.

### Fallback Behavior

If glow is not available:
- Commands still work normally
- Basic markdown rendering is used
- No functionality is lost

### Color Support

Glow integration respects PVM's color mode settings:
- `--color auto`: Let glow auto-detect terminal capabilities
- `--color always`: Force color output
- `--color never`: Disable color output

## Examples

### Viewing Release Notes

```bash
# View latest release notes with enhanced formatting
pvm release-notes

# View specific version
pvm release-notes v1.0.0-rc26

# View with pre-releases included
pvm release-notes --prerelease
```

### Viewing Changelog

```bash
# View recent changelog
pvm changelog

# View more releases
pvm changelog --limit 20

# Include pre-releases
pvm changelog --prerelease
```

### Detailed Version Information

```bash
# Show version with release notes
pvm version --detailed

# Show detailed version with color disabled
pvm version --detailed --color never
```

## Configuration

### GitHub Token

For higher API rate limits, configure a GitHub token:

**Environment Variable:**
```bash
export GITHUB_TOKEN=your_token_here
```

**Configuration File:**
```toml
[pvm.update]
github_token = "${GITHUB_TOKEN}"
```

**Command Line:**
```bash
pvm release-notes --token your_token_here
```

## Benefits

- **Enhanced readability**: Rich formatting for markdown content
- **Better user experience**: Professional appearance matching modern CLI tools
- **Improved accessibility**: Better text formatting and syntax highlighting
- **Seamless integration**: Works with existing PVM workflows

## Technical Details

### Integration Points

1. **UI Framework**: Enhanced `GlowMarkdown()` method in UI output system
2. **Release Notes**: Automatic glow rendering in update commands
3. **Help System**: Enhanced documentation display
4. **Version Command**: Detailed version information with formatting

### Error Handling

- Graceful fallback to basic markdown rendering
- Proper error messages if glow fails
- No impact on core functionality

### Performance

- Glow detection is cached for performance
- Minimal overhead when glow is unavailable
- No impact on command execution speed

## Troubleshooting

### Glow Not Detected

If glow is installed but not detected:

1. Ensure glow is in your PATH:
   ```bash
   which glow
   ```

2. Test glow directly:
   ```bash
   echo "# Test" | glow
   ```

3. Check PVM's detection:
   ```bash
   pvm version --detailed  # Should use glow if available
   ```

### Formatting Issues

If glow output doesn't look right:

1. Check your terminal's color support
2. Try different color modes:
   ```bash
   pvm release-notes --color never
   pvm release-notes --color always
   ```

3. Verify glow version:
   ```bash
   glow --version
   ```

## Related Commands

- `pvm update`: Enhanced release notes display
- `pvm help docs`: Enhanced documentation viewing
- `pvm version`: Basic and detailed version information

## See Also

- [Glow Documentation](https://github.com/charmbracelet/glow)
- [PVM Update Documentation](update.md)
- [PVM Help System](help.md)
