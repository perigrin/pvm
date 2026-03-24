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

## Core Concept: PVM Remotes

PVM remotes work like git remotes. Each remote is a source of Perl variants.

- `origin` is preconfigured to canonical perl5.git. Always present.
- Users add fork remotes pointing to their own perl.git derivatives.
- Forks are namespaced under their remote: `mycompany/myfork-5.40.2`.
- Stock Perl has no namespace prefix, preserving backwards compatibility.

## UX

```
pvm remote add mycompany git@github.com:mycompany/perl-fork.git
pvm remote list
pvm remote remove mycompany
pvm available --remote mycompany
pvm install mycompany/myfork-5.40.2
pvm use mycompany/myfork-5.40.2
pvm versions
```

`pvm versions` output:

```
  5.40.2                          (origin)
  5.38.0                          (origin)
* mycompany/myfork-5.40.2        (mycompany)
```

## Remote Configuration (pvm.toml)

Remotes live in pvm.toml alongside other config. The three-level config
hierarchy (system/user/project) applies: project-level remotes override
user-level, etc.

```toml
[[remotes]]
name = "origin"
url = "https://github.com/Perl/perl5.git"
type = "git"

[[remotes]]
name = "mycompany"
url = "git@github.com:mycompany/perl-fork.git"
type = "git"
```

`origin` is implicit -- present even without config. Its URL can be
overridden but not removed. All other remotes are user-configured.

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
configure_flags = ["-Dusethreads", "-Duse64bitall"]
```

### Manifest Fields (Phase A)

| Field | Required | Description |
|-------|----------|-------------|
| fork.name | yes | Fork product name (not "perl") |
| fork.description | no | Human-readable description |
| fork.base_version | yes | Upstream perl version this derives from |
| fork.license | no | License identifier |
| build.configure_flags | no | Additional Configure flags (appended to defaults) |

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
clone. The fork name defaults to the remote name, and the version derives
from the checked-out tag. Build uses default Configure flags. This supports
quick experimentation without manifest overhead.

## Version Discovery

`pvm available --remote mycompany`:

1. Shallow-clone or fetch the remote repo (cached in `$XDG_CACHE_HOME/pvm/remotes/`)
2. List tags matching version patterns (e.g., `v5.40.2-1`, `myfork-5.40.2`)
3. For each tag, read `.pvm-fork.toml` if present
4. Display as `mycompany/<forkname>-<base_version>`

For `origin`, this is equivalent to the existing `pvm available` behavior
(querying MetaCPAN and binary mirrors). Fork remotes use git tags.

## Installation Flow

`pvm install mycompany/myfork-5.40.2`:

1. Parse the identifier: remote=`mycompany`, name=`myfork-5.40.2`
2. Look up the remote URL in the merged config
3. (**Phase B/C**: check manifest for binaries -- skip in phase A)
4. Clone or fetch the repo, checkout the matching tag
5. Read `.pvm-fork.toml` for fork metadata and configure flags
6. Call the existing `BuildPerl()` infrastructure with:
   - Source directory = the cloned repo (not a CPAN tarball)
   - Configure flags = defaults + fork's custom flags
   - Install prefix = namespaced path under versions/
7. Install to `~/.local/share/pvm/versions/mycompany/myfork-5.40.2/`
8. Register in the version registry with full qualified name and fork metadata

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
  remotes/                           # cached remote clones
    mycompany/                       # shallow clone of fork repo
    otherteam/
```

### Backwards Compatibility

Stock Perl versions (from `origin`) remain flat: `versions/5.40.2/`.
Only fork remotes use the namespaced `versions/<remote>/<name>/` layout.
Existing installs and `.perl-version` files work unchanged.

## Version Registry

The existing registry JSON gains fork metadata fields:

```json
{
  "versions": {
    "5.40.2": {
      "version": "5.40.2",
      "install_path": "/home/user/.local/share/pvm/versions/5.40.2",
      "source": "pvm"
    },
    "mycompany/myfork-5.40.2": {
      "version": "myfork-5.40.2",
      "install_path": "/home/user/.local/share/pvm/versions/mycompany/myfork-5.40.2",
      "source": "pvm",
      "remote": "mycompany",
      "fork_name": "myfork",
      "base_version": "5.40.2"
    }
  }
}
```

Entries without a `remote` field are implicitly `origin`. No migration of
existing registry data is needed.

## Version Resolution

The existing resolver gains fork awareness:

- `.perl-version` can contain `mycompany/myfork-5.40.2`
- `pvm use mycompany/myfork-5.40.2` sets the active version
- Environment variable `PVM_PERL_VERSION=mycompany/myfork-5.40.2` works
- If no `/` is present, the version is stock (origin) -- backwards compatible
- Version aliases (`@latest`, `@stable`) only apply to origin

## Files to Create

- `internal/pvm/remote_command.go` -- `pvm remote add/list/remove` subcommands
- `internal/perl/fork.go` -- fork manifest parsing, git clone/fetch, version
  discovery from remote tags
- `internal/perl/remote.go` -- remote configuration types and lookup

## Files to Modify

- `internal/config/types.go` -- add `Remote` struct and `Remotes []Remote`
  field to PVMConfig
- `internal/perl/build.go` -- extend `BuildOptions` with `SourceDir` field,
  skip download when set
- `internal/perl/registry.go` -- extend `VersionInfo` with `Remote`,
  `ForkName`, `BaseVersion` fields
- `internal/perl/resolver.go` -- parse `remote/name` version identifiers,
  resolve from registry
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

- Remote config parsing and lookup (add/list/remove)
- Fork manifest parsing (.pvm-fork.toml)
- Version identifier parsing (`mycompany/myfork-5.40.2` → remote + name)
- Registry with fork metadata (store, lookup, list)
- Resolver handling of fork identifiers

### Integration tests

- Build from a local git repo (simulates clone step)
- Install to namespaced directory
- Coexistence of stock and fork versions
- `.perl-version` with fork identifier
- `pvm versions` showing both stock and fork

### E2E tests

- `pvm remote add` / `pvm remote list` / `pvm remote remove`
- `pvm install mycompany/myfork-5.40.2` from a test git repo
