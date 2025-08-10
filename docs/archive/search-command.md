# PM Search Command

This document provides detailed information about using the `pm search` command to find and explore CPAN modules.

## Overview

The `pm search` command allows you to search for CPAN modules by name, description, or other attributes. It provides a convenient way to discover modules that match your requirements directly from the command line.

## Basic Usage

```bash
pvm pm search [options] <query>
```

The search query can be a module name, a keyword, or a phrase. The command will search for modules that match the query and display the results.

## Examples

Search for modules related to "database":
```bash
pvm pm search database
```

Search for modules related to "HTTP client":
```bash
pvm pm search "HTTP client"
```

Search for modules with a specific name:
```bash
pvm pm search DBI
```

## Command Options

The `search` command supports the following options:

- `--limit, -l <number>`: Limit the number of results (default: 20)
- `--source, -s <source>`: Use a specific metadata source (metacpan, cpan, custom)
- `--no-cache`: Disable caching for this search

## Output Format

The search results are displayed in a formatted list, with each result including:

- Module name and version
- Author information
- Release date
- Module abstract (short description)

Example output:
```
Search results for 'database' (20 of 1234 results from metacpan):

[1] DBI (1.643)
    Database independent interface for Perl
    Author: TIMB | Released: 2020-01-15

[2] DBD::SQLite (1.70)
    Self-contained RDBMS in a DBI driver
    Author: ISHIGAKI | Released: 2021-02-28

[3] DBIx::Class (0.082842)
    Extensible and flexible object <-> relational mapper
    Author: RIBASUSHI | Released: 2020-05-12
```

## Metadata Sources

The `--source` option allows you to choose which metadata provider to use for the search:

- `metacpan`: Uses the MetaCPAN API (default)
- `cpan`: Uses the CPAN index directly
- `custom`: Uses a custom API specified in the configuration

Each source may provide slightly different results or performance characteristics.

## Configuration Integration

The search command integrates with the PVM configuration system and respects settings in the configuration files:

- The default metadata source is determined by the `pm.metadata_source` configuration option
- The default mirror is determined by the `pm.default_mirror` configuration option
- Caching behavior is controlled by the `pm.cache_modules` and `pm.cache_ttl` configuration options

## Caching Behavior

Search results are cached to improve performance and reduce network usage:

- By default, search results are cached based on the configuration
- The cache TTL (time-to-live) is specified in the configuration (default: 24 hours)
- Caching can be disabled for a specific search with the `--no-cache` option

## Performance Tips

- Use specific search terms to narrow down results
- For frequently used searches, allow caching to improve performance
- If you need the most up-to-date results, use the `--no-cache` option
- The `--limit` option can reduce the response time for large result sets

## Troubleshooting

If you encounter issues with the search command:

- Ensure you have network connectivity to access CPAN metadata
- Check that your mirror configuration is correct
- Try a different metadata source with the `--source` option
- Use the `--no-cache` option to ensure you're getting fresh results
- Increase the search verbosity by using the global `--verbose` flag

## Related Commands

- `pm install`: Install a module after finding it with search
- `pm deps`: Show dependencies for a module
- `pm info`: Get detailed information about a specific module
- `pm mirror`: Configure CPAN mirrors
