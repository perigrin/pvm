# Friendly Fire Forks

Date: 2026-03-24

## Problem

Organizations and individuals maintain custom forks of perl.git -- patched
for specific needs, experimental branches, or compliance requirements. These
forks must use a different name per the Artistic License. There is no
standard way to build, track, distribute, and install these forks alongside
stock Perl.

PVM already builds Perl from source, manages versions, and supports custom
mirrors. This design extends PVM with a remote system (modeled after git
remotes) that lets users build and install Perl forks from any git repo.

## Phased Approach

- **Phase A** (this spec): git remotes, build from source, registry tracking
- **Phase B** (future): manifest-based binary distribution, patch file management
- **Phase C** (future): writable hosting, `pvm publish` for binary distribution

Each phase is independently useful. Phase A enables building forks. Phase B
enables distributing them as binaries. Phase C enables a full CI/CD pipeline
for fork maintenance.

## Prerequisites

Phase A requires the `git` CLI to be installed on the system. PVM uses git
for clone, fetch, and tag listing operations on fork remotes. When git is
not found, `pvm remote` and `pvm install <fork>` commands fail with a clear
error message: "git is required for fork operations but was not found in PATH."

## Core Concept: PVM Remotes

PVM remotes work like git remotes. Each remote is a source of Perl variants.

- `origin` is preconfigured to canonical perl5.git. Always present.
- Users add fork remotes pointing to their own perl.git derivatives.
- Forks are namespaced under their remote: `mycompany/myfork-5.40.2`.
- Stock Perl has no namespace prefix, preserving backwards compatibility.

`origin` is a special case: it uses MetaCPAN and binary mirrors for version
discovery (the existing `pvm available` behavior), not git tags. All other
remotes use git tags for discovery.

## Fork Identifier Grammar

A fork identifier has the form `<remote>/<forkname>-<base_version>`:

```
fork_identifier  = remote "/" fork_version
remote           = [a-z0-9][a-z0-9-]*     (no "/" allowed in remote names)
fork_version     = fork_name "-" base_version
fork_name        = [a-z][a-z0-9-]*         (not "perl")
base_version     = <standard perl version>  (parsed by ParseVersion)
```

Examples:
- `mycompany/myfork-5.40.2` → remote=mycompany, fork=myfork, base=5.40.2
- `otherteam/experiment-5.41.1` → remote=otherteam, fork=experiment, base=5.41.1

A bare version with no `/` is always stock Perl from origin (backwards
compatible). The remote name `origin` is reserved and cannot be used as
a fork remote prefix in identifiers.

### Parsing Strategy

A new `ParseForkIdentifier` function decomposes the identifier:

1. Split on first `/` → remote prefix and fork version
2. If no `/` → stock Perl, delegate to existing `ParseVersion`
3. Split fork version on last `-` followed by a digit → fork name and base version
4. Validate base version with existing `ParseVersion`

All registry and resolver callsites that accept user input go through
`ParseForkIdentifier` first. The base version is what `ParseVersion`
validates. The full identifier (including remote prefix) is never passed
to `ParseVersion` directly.

## UX

```
pvm remote add mycompany git@github.com:mycompany/perl-fork.git
pvm remote list
pvm remote remove mycompany
pvm available --remote mycompany
pvm install mycompany/myfork-5.40.2
pvm uninstall mycompany/myfork-5.40.2
pvm use mycompany/myfork-5.40.2
pvm versions
```

`pvm versions` output:

```
  5.40.2                          (origin)
  5.38.0                          (origin)
* mycompany/myfork-5.40.2        (mycompany)
```

### Error Cases

- `pvm remote add <existing-name>` → error: "remote '<name>' already exists"
- `pvm remote add origin` → error: "cannot add 'origin' -- it is reserved"
- `pvm remote remove origin` → error: "cannot remove 'origin'"
- `pvm remote add <name> <unreachable-url>` → succeeds (URL not validated
  until first use, same as git)
- `pvm install <remote>/<fork>` where remote is unknown → error:
  "remote '<name>' not configured. Run 'pvm remote add <name> <url>' first."

## Remote Configuration (pvm.toml)

Remotes nest under `[pvm]` to match the existing config structure:

```toml
[[pvm.remotes]]
name = "mycompany"
url = "git@github.com:mycompany/perl-fork.git"
type = "git"
```

`origin` is implicit -- present even without config. Its URL can be
overridden but not removed.

### Config Merging

Remote lists are merged additively across config levels (system + user +
project). When the same remote name appears at multiple levels, the
higher-precedence level wins (project overrides user overrides system).
This lets a project override a remote's URL without affecting the user's
global config.

### Remote Validation

Remote names must match `[a-z0-9][a-z0-9-]*`. URLs must be non-empty.
The `type` field defaults to `"git"` if omitted (the only type in phase A).

Phase B adds `mirror_url` and `auth` fields to the remote config.
Phase C adds `publish_url` for writable endpoints.

## Fork Manifest (.pvm-fork.toml)

Lives in the fork repo's root. Tells PVM what this fork provides and
how to build it.

```toml
[fork]
name = "myfork"
description = "MyCompany's patched Perl"
base_version = "5.40.2"
license = "Artistic-2.0"

[build]
configure_flags = ["-Duse64bitall"]
```

### Manifest Fields (Phase A)

| Field | Required | Description |
|-------|----------|-------------|
| fork.name | yes | Fork product name (not "perl") |
| fork.description | no | Human-readable description |
| fork.base_version | yes | Upstream perl version this derives from |
| fork.license | no | License identifier |
| build.configure_flags | no | Additional Configure flags (appended to PVM defaults; duplicates are harmless) |

### Future Manifest Fields

Phase B adds:

```toml
[[versions]]
tag = "v5.40.2-1"
base_version = "5.40.2"

[[versions.binaries]]
platform = "linux-amd64"
url = "https://mirrors.mycompany.com/myfork/myfork-5.40.2-linux-amd64.tar.gz"
sha256 = "abc123..."

[[versions.patches]]
file = "patches/fix-threading.patch"
description = "Fix thread safety in XS loading"
```

Phase C adds:

```toml
[publish]
url = "s3://mycompany-perl-builds/"
auth_type = "aws"
```

The manifest format is forward-compatible: phase A consumers ignore fields
they do not recognize.

### Missing Manifest

If `.pvm-fork.toml` is absent, PVM treats the repo as a straight perl5.git
clone. The fork name defaults to the remote name, the base version derives
from the checked-out tag (parsed as a perl version). Build uses default
Configure flags. This supports quick experimentation without manifest
overhead.

## Version Discovery

`pvm available --remote mycompany`:

1. Clone or fetch the remote repo (see "Clone Caching" below)
2. List git tags
3. For tags that match a perl version pattern (e.g., `v5.40.2`, `v5.40.2-1`),
   read `.pvm-fork.toml` from that tag
4. Display as `mycompany/<forkname>-<base_version>`

### Tag Convention

Fork repos should tag releases with a version pattern. PVM recognizes:
- `v<perl_version>` → e.g., `v5.40.2`
- `v<perl_version>-<release>` → e.g., `v5.40.2-1` (first release)
- `<forkname>-<perl_version>` → e.g., `myfork-5.40.2`

Tags that do not contain a recognizable perl version are ignored.

### Clone Caching

Cached remote clones live in `$XDG_CACHE_HOME/pvm/remotes/<remote_name>/`.

- First access: `git clone --bare --no-single-branch <url>` (bare clone,
  all branches/tags visible)
- Subsequent access: `git fetch --tags` on the cached clone
- Cache invalidation: `pvm remote remove` deletes the cached clone
- No automatic expiration (these are small -- just git metadata)

## Installation Flow

`pvm install mycompany/myfork-5.40.2`:

1. Parse with `ParseForkIdentifier`: remote=`mycompany`, fork=`myfork`,
   base=`5.40.2`
2. Look up the remote URL in the merged config
3. (**Phase B/C**: check manifest for binaries -- skip in phase A)
4. Clone or fetch the repo, checkout the matching tag
5. Read `.pvm-fork.toml` for fork metadata and configure flags
6. Call the existing `BuildPerl()` infrastructure with:
   - `SourceDir` = the cloned repo checkout (not a CPAN tarball)
   - Configure flags = PVM defaults + fork's custom flags
   - Install prefix = namespaced path under versions/
7. Install to `~/.local/share/pvm/versions/mycompany/myfork-5.40.2/`
8. Register in the version registry with fork metadata

### Building from Git (vs Tarball)

The existing `BuildPerl()` downloads a tarball from CPAN. For forks, the
source is a git checkout. The build flow is the same (Configure/make/install)
but the source comes from a different place. `BuildPerl()` needs a new
option to accept a local directory as the source instead of downloading.

The `BuildOptions` struct gains:

```go
SourceDir string  // Local directory containing perl source (from git clone)
```

When `SourceDir` is set, the download/extract steps are skipped and the
build proceeds directly from the specified directory.

### Uninstall

`pvm uninstall mycompany/myfork-5.40.2`:

1. Parse with `ParseForkIdentifier`
2. Look up in registry by full qualified name
3. Remove the install directory
4. Remove the registry entry
5. Do NOT remove the remote's cached clone (other versions may use it)

## Disk Layout

```
~/.local/share/pvm/
  versions/
    5.40.2/                          # stock perl (origin), flat
    5.38.0/                          # stock perl
    mycompany/                       # fork remote namespace
      myfork-5.40.2/                 # fork install
        bin/
        lib/
        ...
    otherteam/
      experiment-5.41.1/

~/.cache/pvm/
  remotes/                           # cached bare clones
    mycompany/                       # bare clone of fork repo
    otherteam/
```

### Backwards Compatibility

Stock Perl versions (from `origin`) remain flat: `versions/5.40.2/`.
Only fork remotes use the namespaced `versions/<remote>/<name>/` layout.
Existing installs and `.perl-version` files work unchanged.

## Version Registry

The existing registry uses UUID keys with `VersionInfo` values. Fork
metadata is added as fields on `VersionInfo`:

```go
type VersionInfo struct {
    // ... existing fields ...
    Remote      string  // Remote name ("" = origin)
    ForkName    string  // Fork product name ("" = stock perl)
    BaseVersion string  // Upstream perl version the fork derives from
}
```

The `DisplayName` for a fork is `<remote>/<forkname>-<base_version>`.
For stock Perl, it is the bare version string. Lookup functions gain
a `FindByDisplayName` variant for fork identifiers.

### RebuildRegistry

`RebuildRegistry` gains a two-level directory scan:

1. Scan `versions/` entries
2. If entry contains `bin/perl` directly → stock install (existing behavior)
3. If entry is a directory with no `bin/perl` → potential remote namespace
4. Scan its subdirectories for `bin/perl` → fork installs
5. Reconstruct fork metadata from the directory path
   (`versions/<remote>/<forkname-version>/`)

## Version Resolution

The existing resolver gains fork awareness:

- `.perl-version` can contain `mycompany/myfork-5.40.2`
- `pvm use mycompany/myfork-5.40.2` sets the active version
- Environment variable `PVM_PERL_VERSION=mycompany/myfork-5.40.2` works
- If no `/` is present, the version is stock (origin) -- backwards compatible
- Version aliases (`@latest`, `@stable`) only apply to origin

### Shell Integration

`pvm sh-use` and `pvm sh-env-activate` must handle the `/` in fork
identifiers when constructing the PATH to `versions/<path>/bin`. The
install path is looked up from the registry (not constructed from the
version string), so the namespaced directory structure is transparent
to the shell integration layer.

### plenv/perlbrew Compatibility

A `.perl-version` file containing a fork identifier (`mycompany/myfork-5.40.2`)
will not be understood by plenv or perlbrew. This is intentional: projects
that require a custom fork have already moved beyond those tools.

## Files to Create

- `internal/pvm/remote_command.go` -- `pvm remote add/list/remove` subcommands
- `internal/perl/fork.go` -- fork manifest parsing, git clone/fetch, version
  discovery from remote tags, `ParseForkIdentifier`
- `internal/perl/remote.go` -- remote configuration types and lookup

## Files to Modify

- `internal/config/types.go` -- add `Remote` struct and `Remotes []Remote`
  field to PVMConfig
- `internal/perl/build.go` -- extend `BuildOptions` with `SourceDir` field,
  skip download when set
- `internal/perl/registry.go` -- extend `VersionInfo` with fork fields,
  add `FindByDisplayName`, update `RebuildRegistry` for two-level scan
- `internal/perl/resolver.go` -- route fork identifiers through
  `ParseForkIdentifier`, resolve from registry by display name
- `internal/pvm/command.go` -- wire `remote` subcommand into the command tree

## What Phase A Does NOT Include

- Binary distribution from fork remotes (phase B)
- Patch file management via manifest (phase B)
- Publishing binaries to remote hosting (phase C)
- Writable hosting configuration (phase C)
- Fork-specific version aliases
- Automatic rebuild when fork repo updates

## Test Plan

### Unit tests

- `ParseForkIdentifier` parsing: valid identifiers, bare versions, edge cases
- Remote config parsing, validation, and lookup (add/list/remove)
- Remote config merging across levels (additive, same-name override)
- Fork manifest parsing (.pvm-fork.toml) with all field combinations
- Missing manifest fallback behavior
- Registry with fork metadata (store, lookup by display name, list)
- `RebuildRegistry` with mixed stock and fork installs
- Resolver handling of fork identifiers in `.perl-version` and env vars

### Integration tests

- Build from a local git repo (simulates clone step)
- Install to namespaced directory
- Coexistence of stock and fork versions
- `.perl-version` with fork identifier
- `pvm versions` showing both stock and fork
- Uninstall fork version

### E2E tests

- `pvm remote add` / `pvm remote list` / `pvm remote remove`
- `pvm remote add` error cases (duplicate, origin, invalid name)
- `pvm install mycompany/myfork-5.40.2` from a test git repo
- `pvm uninstall mycompany/myfork-5.40.2`
