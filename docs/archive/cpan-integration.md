# CPAN Integration in PVM

This document describes the CPAN integration features in the PVM Ecosystem, including metadata retrieval, module search, and caching.

## Overview

The PVM Ecosystem integrates with CPAN (Comprehensive Perl Archive Network) to provide functionalities such as:

- Searching for CPAN modules
- Retrieving metadata about modules
- Managing module dependencies
- Installing modules for specific Perl versions

The integration is primarily provided through the PM (Perl Version Installer) component, which uses various metadata providers to interact with CPAN.

## Metadata Providers

PVM supports multiple metadata providers for accessing CPAN information:

- **MetaCPAN**: The default provider, using the [MetaCPAN API](https://github.com/metacpan/metacpan-api).
- **CPAN**: A provider that directly accesses CPAN index files.
- **Custom**: A provider that allows users to specify a custom API endpoint.

Each provider implements the same interface, ensuring consistent behavior regardless of the chosen provider.

## Configuration Options

CPAN integration can be configured through the PVM configuration system. The following options are available in the `pm` section of the configuration file:

```toml
[pm]
# Preferred installation method for modules (auto, cpanm, cpan, cpm)
preferred_installer = "auto"

# Default CPAN mirror to use for downloads
default_mirror = "https://cpan.metacpan.org"

# Additional mirrors to use as fallbacks if the default mirror fails
additional_mirrors = ["https://cpan.mirror.co.uk", "https://www.cpan.org"]

# Source for CPAN metadata (metacpan, cpan, custom)
metadata_source = "metacpan"

# URL for the metadata API (used when metadata_source is "custom")
metadata_url = "https://api.metacpan.org/v1"

# Directory to use for caching metadata
cache_dir = "$XDG_CACHE_HOME/pvm/cpan"

# Time-to-live for cached metadata in hours (0 to disable caching)
cache_ttl = 24

# Whether to run tests during module installation
test_during_install = false

# Whether to cache modules for faster installation
cache_modules = true

# Whether to force reinstallation of modules
force_reinstall = false

# Whether to check module signatures
check_signatures = true

# Whether to disable network access (for testing)
disable_network = false
```

### Mirror Configuration

Mirrors are CPAN repository mirrors that provide module packages for download. The configuration supports:

- A primary mirror via `default_mirror`
- Backup mirrors via `additional_mirrors`

PVM automatically falls back to additional mirrors if the default mirror fails. You can configure mirrors globally in the user configuration file or per-project in the project configuration file.

### Metadata Source Configuration

The `metadata_source` option determines which provider to use for retrieving CPAN metadata:

- `metacpan`: Uses the MetaCPAN API (default)
- `cpan`: Uses the CPAN index directly
- `custom`: Uses a custom API specified by `metadata_url`

The `metadata_url` option is required when using a custom metadata source.

### Caching Configuration

PVM caches CPAN metadata to improve performance and reduce network usage. The caching system can be configured with:

- `cache_dir`: The directory where cache files are stored
- `cache_ttl`: The time-to-live for cached metadata in hours

Cache files are stored in a directory structure following the XDG Base Directory Specification. The `cache_dir` can use the `$XDG_CACHE_HOME` variable, which will be expanded to the actual cache directory.

Setting `cache_ttl` to 0 effectively disables caching by making all cache entries expire immediately.

### Installation Configuration

Module installation behavior can be configured with:

- `preferred_installer`: The preferred tool for installing modules
- `test_during_install`: Whether to run tests during installation
- `cache_modules`: Whether to cache downloaded modules
- `force_reinstall`: Whether to force reinstallation of already installed modules
- `check_signatures`: Whether to check module signatures

## Command-Line Options

Command-line options can override the configuration settings for individual commands. The following options are available for the `pm search` command:

- `--source, -s`: Use a specific metadata source (metacpan, cpan, custom)
- `--limit, -l`: Limit the number of search results
- `--no-cache`: Disable caching for this search

Example:
```bash
pvm pm search --source metacpan --limit 50 --no-cache "Moose"
```

Other PM commands will have similar options for controlling metadata retrieval and caching.

## Directory Structure

The CPAN integration uses the following directory structure:

- `$XDG_CONFIG_HOME/pvm/pvm.toml`: User configuration file
- `$XDG_CACHE_HOME/pvm/cpan`: Metadata cache directory
- `$XDG_DATA_HOME/pvm/modules`: Module installation directory

## Error Handling

CPAN integration includes robust error handling for various scenarios:

- Network connectivity issues
- Invalid API responses
- Cache access failures
- Configuration errors

Errors are reported with detailed information about the failure, including:

- The source of the error (which provider)
- An error code
- A detailed error message
- Context-specific information (e.g., URL, HTTP status code)

## Performance Considerations

- Metadata caching significantly improves performance for repeated operations
- The default cache TTL of 24 hours provides a good balance between freshness and performance
- Module caching reduces download time for frequently used modules
